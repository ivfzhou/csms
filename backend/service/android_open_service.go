/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AndroidAPIDownloadCertificate 下载安卓证书。
func AndroidAPIDownloadCertificate(ctx context.Context, req *protocol.AndroidAPIDownloadCertificateReq) (
	fileObj *FileInfo, err error) {

	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中获取证书信息。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate information from database")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.Content,
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Category,
			androidCertificateDo.Alias_,
			androidCertificateDo.Sha1,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.Category.Eq(model.AndroidCertificateTypeDebug),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var jksData []byte
	{
		log.Info(ctx, "decrypt android certificate")
		var secret []byte
		secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
		if err != nil {
			return
		}
		jksData, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件保存到数据库。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      apiAccount.ID,
			Type:        model.EventTypeDownloadAndroidCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceAPI,
		}, map[EventField]any{
			EventUser:  apiAccount.AccountID,
			EventApp:   app.Name,
			EventAlias: androidCertificate.Alias_,
			EventDetail: util.GetPrintJSON(map[string]any{
				"certificateId": req.CertificateID,
				"sha1":          androidCertificate.Sha1,
				"owner":         androidCertificate.Owner,
				"publisher":     androidCertificate.Publisher,
				"notAfter":      formatTime(&androidCertificate.NotAfter),
				"notBefore":     formatTime(&androidCertificate.NotBefore),
				"type":          model.AllAndroidCertificateTypeDescriptions[androidCertificate.Category],
			}),
		})
		if err != nil {
			return
		}
	}

	fileObj = &FileInfo{
		Name:   androidCertificate.Alias_ + ".jks",
		Size:   int64(len(jksData)),
		Reader: io.NopCloser(bytes.NewReader(jksData)),
	}

	return
}

// AndroidAPISubmitAPKSigningJob 提交 APK 文件签名任务。
func AndroidAPISubmitAPKSigningJob(ctx context.Context, req *protocol.AndroidAPISubmitAPKSigningJobReq) (
	rsp *protocol.AndroidAPISubmitAPKSigningJobRsp, err error) {

	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
			return
		}
	}

	// 查询数据库，获取文件信息，并校验文件。
	{
		log.Info(ctx, "verify file to be signed")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		var file *model.File
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Type,
			fileDo.AppID,
			fileDo.Name,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if file == nil || file.Type != model.FileTypeAndroidSigning || file.AppID != app.ID ||
			strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionAPK {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file not found")
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.ID,
			androidCertificateDo.NotAfter,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificate.NotAfter) > 0 {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"certificate has expired")
			return
		}
	}

	// 将签名任务信息保存到数据库。
	var jobID string
	{
		log.Info(ctx, "save android signing job to database")
		jobID, err = generateID(ctx, IDAndroidJob)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAndroidJob, jobID), "failed to reclaim android signing job id", jobID)
			}
		}()
		err = createAndroidSigningJob(ctx, &model.AndroidSigningJob{
			JobID:            jobID,
			AppID:            app.ID,
			UserID:           apiAccount.ID,
			Type:             model.AndroidSigningJobTypeAPK,
			CertificateID:    androidCertificate.ID,
			FileID:           req.FileID,
			SignatureSchemas: req.SignatureSchema,
			Source:           model.SourceAPI,
			Status:           model.AndroidSigningJobStatusSigning,
			CreatedTime:      time.Now(),
		})
		if err != nil {
			return
		}
	}

	// 发送签名消息。
	{
		log.Info(ctx, "publish android signing job message")
		err = publishMessageToQueue(ctx, cfg.Get().RabbitMQ().AndroidSigningJobQueue(), []byte(jobID))
		if err != nil {
			return
		}
	}

	rsp = &protocol.AndroidAPISubmitAPKSigningJobRsp{JobID: jobID}

	return
}

// AndroidAPISubmitAABSigningJob 提交 AAB 文件签名任务。
func AndroidAPISubmitAABSigningJob(ctx context.Context, req *protocol.AndroidAPISubmitAABSigningJobReq) (
	rsp *protocol.AndroidAPISubmitAABSigningJobRsp, err error) {

	var apiAccountInfo *model.APIAccount
	var appInfo *model.App
	{
		log.Info(ctx, "get context information")
		apiAccountInfo = ctxs.APIAccount(ctx)
		appInfo = ctxs.App(ctx)
		if apiAccountInfo == nil || appInfo == nil {
			log.Warn(ctx, "unknown context", appInfo, apiAccountInfo)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用平台和状态。
	{
		log.Info(ctx, "verify app")
		if appInfo.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
		if appInfo.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
			return
		}
	}

	// 查询数据库，获取文件信息，并校验文件。
	{
		log.Info(ctx, "verify file to be signed")
		fileQuery := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		var fileInfo *model.File
		fileInfo, err = fileQuery.WithContext(ctx).Select(
			fileQuery.Type,
			fileQuery.AppID,
			fileQuery.Name,
		).Where(
			fileQuery.FileID.Eq(req.FileID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to get file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if fileInfo == nil || fileInfo.Type != model.FileTypeAndroidSigning || fileInfo.AppID != appInfo.ID ||
			strings.ToLower(filepath.Ext(fileInfo.Name)) != cc.ExtensionAAB {
			log.Warn(ctx, "file is invalid", fileInfo)
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file not found")
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificateInfo *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateQuery := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificateInfo, err = androidCertificateQuery.WithContext(ctx).Select(
			androidCertificateQuery.ID,
			androidCertificateQuery.NotAfter,
		).Where(
			androidCertificateQuery.CertificateID.Eq(req.CertificateID),
			androidCertificateQuery.AppID.Eq(appInfo.ID),
			androidCertificateQuery.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificateInfo.NotAfter) > 0 {
			log.Warn(ctx, "android certificate has expired")
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"certificate has expired")
			return
		}
	}

	// 将签名任务信息保存到数据库。
	var jobID string
	{
		log.Info(ctx, "save android signing job to database")
		jobID, err = generateID(ctx, IDAndroidJob)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAndroidJob, jobID), "failed to reclaim android signing job id")
			}
		}()
		err = createAndroidSigningJob(ctx, &model.AndroidSigningJob{
			JobID:         jobID,
			AppID:         appInfo.ID,
			UserID:        apiAccountInfo.ID,
			Type:          model.AndroidSigningJobTypeAAB,
			CertificateID: androidCertificateInfo.ID,
			FileID:        req.FileID,
			Source:        model.SourceAPI,
			Status:        model.AndroidSigningJobStatusSigning,
			CreatedTime:   time.Now(),
		})
		if err != nil {
			return
		}
	}

	// 发送签名消息。
	{
		log.Info(ctx, "publish android signing job message")
		err = publishMessageToQueue(ctx, cfg.Get().RabbitMQ().AndroidSigningJobQueue(), []byte(jobID))
		if err != nil {
			return
		}
	}

	rsp = &protocol.AndroidAPISubmitAABSigningJobRsp{JobID: jobID}

	return
}

// AndroidAPISubmitAPKPatchSigningJob 提交 APK 补丁包文件签名任务。
func AndroidAPISubmitAPKPatchSigningJob(ctx context.Context, req *protocol.AndroidAPISubmitAPKPatchSigningJobReq) (
	rsp *protocol.AndroidAPISubmitAPKPatchSigningJobRsp, err error) {

	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
			return
		}
	}

	// 查询数据库，获取文件信息，并校验文件。
	{
		log.Info(ctx, "verify file to be signed")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		var file *model.File
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Type,
			fileDo.AppID,
			fileDo.Name,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if file == nil || file.Type != model.FileTypeAndroidSigning || file.AppID != app.ID ||
			strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionAPK {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file not found")
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.ID,
			androidCertificateDo.NotAfter,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificate.NotAfter) > 0 {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"certificate has expired")
			return
		}
	}

	// 将签名任务信息保存到数据库。
	var jobID string
	{
		log.Info(ctx, "save android signing job to database")
		jobID, err = generateID(ctx, IDAndroidJob)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAndroidJob, jobID), "failed to reclaim android signing job id", jobID)
			}
		}()
		err = createAndroidSigningJob(ctx, &model.AndroidSigningJob{
			JobID:            jobID,
			MinimumSdkLevel:  req.MinimumSDKVersion,
			SignatureSchemas: req.SignatureSchema,
			AppID:            app.ID,
			UserID:           apiAccount.ID,
			Type:             model.AndroidSigningJobTypePatch,
			CertificateID:    androidCertificate.ID,
			FileID:           req.FileID,
			Source:           model.SourceAPI,
			Status:           model.AndroidSigningJobStatusSigning,
			CreatedTime:      time.Now(),
		})
		if err != nil {
			return
		}
	}

	// 发送签名消息。
	{
		log.Info(ctx, "publish android signing job message")
		err = publishMessageToQueue(ctx, cfg.Get().RabbitMQ().AndroidSigningJobQueue(), []byte(jobID))
		if err != nil {
			return
		}
	}

	rsp = &protocol.AndroidAPISubmitAPKPatchSigningJobRsp{JobID: jobID}

	return
}

// AndroidAPIGetJobInformation 获取任务信息。
func AndroidAPIGetJobInformation(ctx context.Context, req *protocol.AndroidAPIGetSigningJobInformationReq) (
	rsp *protocol.AndroidAPIGetSigningJobInformationRsp, err error) {

	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 查库，获取任务信息。
	var androidSigningJob *model.AndroidSigningJob
	{
		log.Info(ctx, "get job information")
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob.
			Table(model.GetAndroidSigningJobByID(req.JobID))
		androidSigningJob, err = androidSigningJobDo.WithContext(ctx).Select(
			androidSigningJobDo.UserID,
			androidSigningJobDo.Source,
			androidSigningJobDo.CertificateID,
			androidSigningJobDo.Type,
			androidSigningJobDo.FileID,
			androidSigningJobDo.SignedFileID,
			androidSigningJobDo.Status,
			androidSigningJobDo.MinimumSdkLevel,
			androidSigningJobDo.SignatureSchemas,
			androidSigningJobDo.FinishedTime,
			androidSigningJobDo.CreatedTime,
			androidSigningJobDo.Log,
		).Where(
			androidSigningJobDo.JobID.Eq(req.JobID),
			androidSigningJobDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查询用户名。
	var userName string
	{
		log.Info(ctx, "get user name")
		switch androidSigningJob.Source {
		case model.SourceAPI:
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.AccountID,
			).Where(
				apiAccountDo.ID.Eq(androidSigningJob.UserID),
			).Scan(&userName)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "api account not found", androidSigningJob.UserID)
				} else {
					log.Error(ctx, "failed to retrieve api account name from database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
			}
		case model.SourceWeb:
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.NameEn,
			).Where(
				userDo.ID.Eq(androidSigningJob.UserID),
			).Scan(&userName)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "user not found", androidSigningJob.UserID)
				} else {
					log.Error(ctx, "failed to retrieve user name from database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
			}
		default:
			log.Warn(ctx, "unknown source", androidSigningJob.Source)
		}
	}

	// 查询证书信息。
	var certificateAlias string
	var certificateID string
	{
		log.Info(ctx, "get certificate information")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		var androidCertificate *model.AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.CertificateID,
			androidCertificateDo.Alias_,
		).Where(
			androidCertificateDo.ID.Eq(androidSigningJob.CertificateID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found", androidCertificate.CertificateID)
			} else {
				log.Error(ctx, "failed to retrieve android certificate from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
		if androidCertificate != nil {
			certificateID = androidCertificate.CertificateID
			certificateAlias = androidCertificate.Alias_
		}
	}

	// 查询文件名。
	var fileName string
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(androidSigningJob.FileID))
		err = fileDo.WithContext(ctx).Select(
			fileDo.Name,
		).Where(
			fileDo.FileID.Eq(androidSigningJob.FileID),
		).Scan(&fileName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Warn(ctx, "file not found", androidSigningJob.FileID)
			} else {
				log.Error(ctx, "failed to retrieve file name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 返回数据。
	{
		signingConfig := ""
		switch androidSigningJob.Type {
		case model.AndroidSigningJobTypeAPK:
			signingConfig = "signatureSchema=" + strings.Join(util.ListTo(androidSigningJob.SignatureSchemas,
				func(e int) string { return fmt.Sprintf("v%d", e) }), ",")
		case model.AndroidSigningJobTypePatch:
			signingConfig = fmt.Sprintf("minimumSdkVersion=%v", androidSigningJob.MinimumSdkLevel)
		default:
		}
		rsp = &protocol.AndroidAPIGetSigningJobInformationRsp{
			Type:             androidSigningJob.Type,
			Source:           androidSigningJob.Source,
			CertificateAlias: certificateAlias,
			SigningConfig:    signingConfig,
			FileName:         fileName,
			FileID:           androidSigningJob.FileID,
			User:             userName,
			CreatedTime:      formatTime(&androidSigningJob.CreatedTime),
			FinishedTime:     formatTime(&androidSigningJob.FinishedTime),
			Log:              androidSigningJob.Log,
			SignedFileID:     androidSigningJob.SignedFileID,
			CertificateID:    certificateID,
			Status:           androidSigningJob.Status,
		}
	}

	return
}

// AndroidAPIListCertificates 获取安卓证书列表。
func AndroidAPIListCertificates(ctx context.Context) (rsp *protocol.AndroidAPIListCertificatesRsp, err error) {
	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中获取证书信息。
	var androidCertificates []*model.AndroidCertificate
	{
		log.Info(ctx, "get android certificates from database")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificates, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.UserID,
			androidCertificateDo.CertificateID,
			androidCertificateDo.Alias_,
			androidCertificateDo.Owner,
			androidCertificateDo.SignatureAlgorithm,
			androidCertificateDo.PublicKeyAlgorithm,
			androidCertificateDo.Sha1,
			androidCertificateDo.Sha256,
			androidCertificateDo.Category,
			androidCertificateDo.CreatedTime,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
		).Where(
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Order(androidCertificateDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve android certificates information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(androidCertificates) <= 0 {
			return
		}
	}

	// 从数据中查询用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users names")
		userIDs := util.CleanNumbers(util.ListTo(androidCertificates, func(e *model.AndroidCertificate) int {
			return e.UserID
		}))
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AndroidAPIListCertificatesItem, len(androidCertificates))
		for i, v := range androidCertificates {
			list[i] = &protocol.AndroidAPIListCertificatesItem{
				ID:                 v.CertificateID,
				Alias:              v.Alias_,
				Owner:              v.Owner,
				SignatureAlgorithm: v.SignatureAlgorithm,
				PublicKeyAlgorithm: v.PublicKeyAlgorithm,
				SHA1:               v.Sha1,
				SHA256:             v.Sha256,
				CreatedTime:        formatTime(&v.CreatedTime),
				ExpiredTime:        formatTime(&v.NotAfter),
				Creator:            userIDToName[v.UserID],
				EffectedTime:       formatTime(&v.NotBefore),
				Type:               v.Category,
			}
		}
		rsp = &protocol.AndroidAPIListCertificatesRsp{List: list}
	}

	return
}
