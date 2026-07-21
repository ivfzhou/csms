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
	"io"
	"net/http"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AppleAPIDownloadCertificate 下载证书和描述文件。
func AppleAPIDownloadCertificate(ctx context.Context, req *protocol.AppleAPIDownloadCertificateReq) (
	fileObj *FileInfo, err error) {

	// 获取上下文信息。
	var app *model.App
	var apiAccount *model.APIAccount
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		apiAccount = ctxs.APIAccount(ctx)
		if app == nil || apiAccount == nil {
			log.Warn(ctx, "unknown context", apiAccount, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformApple {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 查询证书信息。
	{
		switch req.Type {
		case protocol.AppleFileTypeProfile:
			log.Info(ctx, "get app profile information")
			appleProfileDo := conn.MySQLClient(ctx).AppleProfile
			var appleProfile *model.AppleProfile
			appleProfile, err = appleProfileDo.WithContext(ctx).Select(
				appleProfileDo.Content,
			).Where(
				appleProfileDo.ProfileID.Eq(req.CertificateID),
				appleProfileDo.DeletedTime.IsNull(),
				appleProfileDo.AppID.Eq(app.ID),
			).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
					return
				}
				log.Error(ctx, "failed to get apple profile from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			fileObj = &FileInfo{
				Name:   req.CertificateID + ".profile",
				Size:   int64(len(appleProfile.Content)),
				Reader: io.NopCloser(bytes.NewReader(appleProfile.Content)),
			}
			return
		case protocol.AppleFileTypePushCertificate:
			log.Info(ctx, "get apple push certificate information")
			appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
			var appleCertificate *model.AppleCertificate
			appleCertificate, err = appleCertificateDo.WithContext(ctx).Select(
				appleCertificateDo.Content,
				appleCertificateDo.AesKeyID,
			).Where(
				appleCertificateDo.CertificateID.Eq(req.CertificateID),
				appleCertificateDo.AppID.Eq(app.ID),
				appleCertificateDo.Category.Eq(model.AppleCertificateCategoryPush),
				appleCertificateDo.DeletedTime.IsNull(),
			).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
					return
				}
				log.Error(ctx, "failed to get apple push certificate from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			// 解密证书。
			var secret []byte
			secret, err = getAESSecret(ctx, appleCertificate.AesKeyID)
			if err != nil {
				return
			}
			appleCertificate.Content, err = util.AESCBCDecrypt(appleCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt apple push certificate content", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			fileObj = &FileInfo{
				Name:   req.CertificateID + ".p12",
				Size:   int64(len(appleCertificate.Content)),
				Reader: io.NopCloser(bytes.NewReader(appleCertificate.Content)),
			}

			return
		case protocol.AppleFileTypeSigningCertificate:
			log.Info(ctx, "get apple signing certificate information")
			appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
			var appleCertificate *model.AppleCertificate
			appleCertificate, err = appleCertificateDo.WithContext(ctx).Select(
				appleCertificateDo.Content,
				appleCertificateDo.AesKeyID,
			).Where(
				appleCertificateDo.CertificateID.Eq(req.CertificateID),
				appleCertificateDo.AppID.IsNull(),
				appleCertificateDo.Category.Eq(model.AppleCertificateCategorySigning),
				appleCertificateDo.DeletedTime.IsNull(),
			).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
					return
				}
				log.Error(ctx, "failed to get apple signing certificate from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			// 解密证书。
			var secret []byte
			secret, err = getAESSecret(ctx, appleCertificate.AesKeyID)
			if err != nil {
				return
			}
			appleCertificate.Content, err = util.AESCBCDecrypt(appleCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt apple signing certificate content", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			fileObj = &FileInfo{
				Name:   req.CertificateID + ".p12",
				Size:   int64(len(appleCertificate.Content)),
				Reader: io.NopCloser(bytes.NewReader(appleCertificate.Content)),
			}
			return
		default:
			log.Warn(ctx, "unknown type", req.Type)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}
}

// AppleAPISubmitSigningJob 提交签名任务。
func AppleAPISubmitSigningJob(ctx context.Context, req *protocol.AppleAPISubmitSigningJobReq) (
	rsp *protocol.AppleAPISubmitSigningJobRsp, err error) {

	// 获取上下文信息。
	var app *model.App
	var apiAccount *model.APIAccount
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		apiAccount = ctxs.APIAccount(ctx)
		if app == nil || apiAccount == nil {
			log.Warn(ctx, "unknown context", apiAccount, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
		}
		if app.Platform != model.AppPlatformApple {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 获取描述文件信息。
	var appleProfile *model.AppleProfile
	{
		log.Info(ctx, "get apple profile information")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		appleProfile, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.Type,
			appleProfileDo.ID,
			appleProfileDo.BundleID,
		).Where(
			appleProfileDo.ProfileID.Eq(req.ProfileID),
			appleProfileDo.AppID.Eq(app.ID),
			appleProfileDo.DeletedTime.IsNull(),
		).Last()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "apple profile not found")
				err = errs.NewWithStatusMsg(consts.ErrAppleProfileNotFound, http.StatusBadRequest, "apple profile not found")
				return
			}
			log.Error(ctx, "failed to retrieve apple profile information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取 Bundle ID 信息。
	{
		if appleProfile.Type != model.AppleProfileTypeIOSAppInHouse {
			log.Info(ctx, "get apple bundle information")
			appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
			_, err = appleBundleIDDo.WithContext(ctx).Select(
				appleBundleIDDo.ID,
			).Where(
				appleBundleIDDo.AppID.Eq(app.ID),
				appleBundleIDDo.ID.Eq(appleProfile.BundleID),
			).Last()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "apple bundle not found", appleProfile.BundleID)
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"apple bundle not found")
					return
				}
				log.Error(ctx, "failed to retrieve apple bundle information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 校验文件信息。
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		var file *model.File
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.AppID,
			fileDo.UserID,
			fileDo.Type,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Warn(ctx, "file not found", req.FileID)
				err = errs.New(consts.ErrFileNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if file.AppID != app.ID || file.Type != model.FileTypeAppleSigning {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"file does not belong to apple signing job")
			return
		}
	}

	// 保存签名任务信息。
	var jobID string
	{
		log.Info(ctx, "save apple signing job to database")
		jobID, err = generateID(ctx, IDAppleJob)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAppleJob, jobID), "reclaim job id failed", jobID)
			}
		}()
		now := time.Now()
		appleSigningJobDo := conn.MySQLTxClient(ctx).AppleSigningJob.Table(model.GetAppleSigningJobByID(jobID))
		err = appleSigningJobDo.WithContext(ctx).Select(
			appleSigningJobDo.JobID,
			appleSigningJobDo.AppID,
			appleSigningJobDo.UserID,
			appleSigningJobDo.ProfileID,
			appleSigningJobDo.FileID,
			appleSigningJobDo.Source,
			appleSigningJobDo.Status,
			appleSigningJobDo.CreatedTime,
		).Create(&model.AppleSigningJob{
			JobID:       jobID,
			AppID:       app.ID,
			UserID:      apiAccount.ID,
			ProfileID:   appleProfile.ID,
			FileID:      req.FileID,
			Source:      model.SourceAPI,
			Status:      model.AppleSigningJobStatusRunning,
			CreatedTime: now,
		})
		if err != nil {
			log.Error(ctx, "failed to create apple signing job to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 发送消息到 MQ。
	{
		log.Info(ctx, "send mq")
		err = publishMessageToQueue(ctx, cfg.Get().RabbitMQ().AppleSigningJobQueue(), []byte(jobID))
		if err != nil {
			return
		}
	}

	rsp = &protocol.AppleAPISubmitSigningJobRsp{JobID: jobID}

	return
}

// AppleAPIGetSigningJobInformation 获取签名任务信息。
func AppleAPIGetSigningJobInformation(ctx context.Context, req *protocol.AppleAPIGetSigningJobInformationReq) (
	rsp *protocol.AppleAPIGetSigningJobInformationRsp, err error) {

	// 获取上下文信息。
	var app *model.App
	var apiAccount *model.APIAccount
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		apiAccount = ctxs.APIAccount(ctx)
		if app == nil || apiAccount == nil {
			log.Warn(ctx, "unknown context", apiAccount, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformApple {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中获取签名任务信息。
	var appleSigningJob *model.AppleSigningJob
	{
		log.Info(ctx, "get apple signing job information")
		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob.Table(model.GetAppleSigningJobByID(req.JobID))
		appleSigningJob, err = appleSigningJobDo.WithContext(ctx).Select(
			appleSigningJobDo.SignedFileID,
			appleSigningJobDo.FileID,
			appleSigningJobDo.CreatedTime,
			appleSigningJobDo.FinishedTime,
			appleSigningJobDo.Log,
			appleSigningJobDo.Source,
			appleSigningJobDo.Status,
			appleSigningJobDo.ProfileID,
		).Where(
			appleSigningJobDo.JobID.Eq(req.JobID),
			appleSigningJobDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve apple signing job information from database", err)
			return
		}
	}

	// 从数据库中查询描述文件信息。
	var appleProfile *model.AppleProfile
	{
		log.Info(ctx, "get apple profile information")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		appleProfile, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.ProfileID,
			appleProfileDo.BundleID,
		).Where(
			appleProfileDo.ID.Eq(appleSigningJob.ProfileID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve apple profile information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if appleProfile == nil {
			appleProfile = &model.AppleProfile{}
		}
	}

	// 从数据库中查询 Bundle ID。
	var bundleID string
	{
		if appleProfile.BundleID > 0 {
			log.Info(ctx, "get apple profile information")
			appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
			err = appleBundleIDDo.WithContext(ctx).Select(
				appleBundleIDDo.BundleID,
			).Where(
				appleBundleIDDo.AppID.Eq(app.ID),
				appleBundleIDDo.ID.Eq(appleSigningJob.ProfileID),
			).Scan(&bundleID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 从数据库中查询文件名。
	var fileIDToName map[string]string
	{
		log.Info(ctx, "get file names")
		fileIDToName, err = GetFileNamesByIDs(ctx,
			[]string{appleSigningJob.FileID, appleSigningJob.SignedFileID})
		if err != nil {
			return
		}
	}

	// 从数据库中查询提交人。
	var user string
	{
		log.Info(ctx, "get user name")
		switch appleSigningJob.Source {
		case model.SourceWeb:
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.NameEn,
			).Where(
				userDo.ID.Eq(appleSigningJob.UserID),
			).Scan(&user)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		case model.SourceAPI:
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.AccountID,
			).Where(
				apiAccountDo.ID.Eq(appleSigningJob.UserID),
			).Scan(&user)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve api account name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		default:
			log.Warn(ctx, "unknown source", appleSigningJob.Source)
		}
	}

	rsp = &protocol.AppleAPIGetSigningJobInformationRsp{
		BundleID:       bundleID,
		ProfileID:      appleProfile.ProfileID,
		FileID:         appleSigningJob.FileID,
		FileName:       fileIDToName[appleSigningJob.FileID],
		UserName:       user,
		CreatedTime:    formatTime(&appleSigningJob.CreatedTime),
		FinishedTime:   formatTime(&appleSigningJob.FinishedTime),
		Status:         appleSigningJob.Status,
		Log:            appleSigningJob.Log,
		SignedFileID:   appleSigningJob.SignedFileID,
		SignedFileName: fileIDToName[appleSigningJob.SignedFileID],
		Source:         appleSigningJob.Source,
	}

	return
}
