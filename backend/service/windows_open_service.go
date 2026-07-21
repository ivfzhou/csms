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
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen/field"
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

// WindowsAPIDownloadCertificate 下载证书。
func WindowsAPIDownloadCertificate(ctx context.Context, req *protocol.WindowsAPIDownloadCertificateReq) (
	fileObj *FileInfo, err error) {

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
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.AesKeyID,
			windowsCertificateDo.Content,
			windowsCertificateDo.Type,
			windowsCertificateDo.CommonName,
			windowsCertificateDo.Sha1,
			windowsCertificateDo.Publisher,
			windowsCertificateDo.NotBefore,
			windowsCertificateDo.NotAfter,
		).Where(
			windowsCertificateDo.CertificateID.Eq(req.CertificateID),
			windowsCertificateDo.AppID.Eq(app.ID),
			windowsCertificateDo.Type.Eq(model.WindowsCertificateTypePersonalOV),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "windows certificate not found")
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var certificateBytes []byte
	{
		log.Info(ctx, "decrypt windows certificate data")
		var secret []byte
		secret, err = getAESSecret(ctx, windowsCertificate.AesKeyID)
		if err != nil {
			return
		}
		certificateBytes, err = util.AESCBCDecrypt(windowsCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt windows certificate data", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录应用事件到数据库中。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      apiAccount.ID,
			Type:        model.EventTypeDownloadWindowsCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceAPI,
		}, map[EventField]any{
			EventUser:                  apiAccount.AccountID,
			EventApp:                   app.Name,
			EventCertificateCommonName: windowsCertificate.CommonName,
			EventDetail: util.GetPrintJSON(map[string]any{
				"sha1":          windowsCertificate.Sha1,
				"certificateId": req.CertificateID,
				"type":          model.AllWindowsCertificateDescriptions[windowsCertificate.Type],
				"notAfter":      formatTime(&windowsCertificate.NotAfter),
				"notBefore":     formatTime(&windowsCertificate.NotBefore),
				"publisher":     windowsCertificate.Publisher,
			}),
		})
		if err != nil {
			return
		}
	}

	fileObj = &FileInfo{
		Name:   windowsCertificate.CommonName + ".pfx",
		Size:   int64(len(certificateBytes)),
		Reader: io.NopCloser(bytes.NewReader(certificateBytes)),
	}

	return
}

// WindowsAPIGetCertificatePassword 查看证书密码。
func WindowsAPIGetCertificatePassword(ctx context.Context, req *protocol.WindowsAPIGetCertificatePasswordReq) (
	rsp *protocol.WindowsAPIGetCertificatePasswordRsp, err error) {

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
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information from database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.ID,
			windowsCertificateDo.Type,
			windowsCertificateDo.Password,
		).Where(
			windowsCertificateDo.CertificateID.Eq(req.CertificateID),
			windowsCertificateDo.DeletedTime.IsNull(),
			windowsCertificateDo.WithContext(ctx).Where(
				windowsCertificateDo.AppID.Eq(app.ID),
				windowsCertificateDo.Type.Eq(model.WindowsCertificateTypePersonalOV),
			).Or(
				windowsCertificateDo.AppID.IsNull(),
				windowsCertificateDo.Type.Eq(model.WindowsCertificateTypePersonalEV),
			),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "windows certificate not found")
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验证书已授权给应用。
	{
		if windowsCertificate.Type == model.WindowsCertificateTypePersonalEV {
			log.Info(ctx, "check whether windows certificate is authorized")
			windowsCertificateAuthorizationDo := conn.MySQLClient(ctx).WindowsCertificateAuthorization
			var count int64
			count, err = windowsCertificateAuthorizationDo.WithContext(ctx).Where(
				windowsCertificateAuthorizationDo.AppID.Eq(app.ID),
				windowsCertificateAuthorizationDo.CertificateID.Eq(windowsCertificate.ID),
			).Count()
			if err != nil {
				log.Error(ctx, "failed to check whether windows certificate is authorized", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if count <= 0 {
				log.Warn(ctx, "windows certificate is not authorized for app")
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"certificate unauthorized use")
				return
			}
		}
	}

	rsp = &protocol.WindowsAPIGetCertificatePasswordRsp{Password: windowsCertificate.Password}

	return
}

// WindowsAPISubmitSigningJob 提交 Windows 签名任务。
func WindowsAPISubmitSigningJob(ctx context.Context, req *protocol.WindowsAPISubmitSigningJobReq) (
	rsp *protocol.WindowsAPISubmitSigningJobRsp, err error) {

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
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		if req.SigningType == model.WindowsSigningJobTypeAttestation {
			log.Info(ctx, "get windows certificate information")
			log.Info(ctx, "find cab signing certificate from database")
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.Type,
				windowsCertificateDo.AppID,
				windowsCertificateDo.ID,
				windowsCertificateDo.Sha1,
			).Where(
				windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
				windowsCertificateDo.DeletedTime.IsNull(),
				windowsCertificateDo.NotAfter.Gte(time.Now()),
			).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(ctx, "no windows certificate available")
					err = errs.New(consts.ErrSystem)
					return
				}
				log.Error(ctx, "failed to retrieve windows certificate information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		} else {
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.Type,
				windowsCertificateDo.AppID,
				windowsCertificateDo.ID,
				windowsCertificateDo.Sha1,
			).Where(
				windowsCertificateDo.CertificateID.Eq(req.CertificateID),
				windowsCertificateDo.WithContext(ctx).Where(windowsCertificateDo.AppID.Eq(app.ID)).
					Or(windowsCertificateDo.AppID.IsNull()),
				windowsCertificateDo.DeletedTime.IsNull(),
			).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"certificate not found")
					return
				}
				log.Error(ctx, "failed to retrieve windows certificate information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 校验证书授权给应用使用。
	{
		if req.SigningType != model.WindowsSigningJobTypeAttestation {
			log.Info(ctx, "verify windows certificate")
			switch windowsCertificate.Type {
			case model.WindowsCertificateTypePersonalEV:
				if windowsCertificate.AppID > 0 {
					log.Warn(ctx, "app can not use the windows certificate")
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"certificate cannot be use")
					return
				}
				log.Info(ctx, "check whether the windows certificate authorized for app")
				windowsCertificateAuthorizationDo := conn.MySQLClient(ctx).WindowsCertificateAuthorization
				var count int64
				count, err = windowsCertificateAuthorizationDo.WithContext(ctx).Where(
					windowsCertificateAuthorizationDo.CertificateID.Eq(windowsCertificate.ID),
					windowsCertificateAuthorizationDo.AppID.Eq(app.ID),
				).Count()
				if err != nil {
					log.Error(ctx, "failed to retrieve windows certificate authorized for app", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
				if count <= 0 {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"certificate unauthorized use")
					return
				}
			case model.WindowsCertificateTypePersonalOV:
				if windowsCertificate.AppID != app.ID {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"app does not have the certificate")
					return
				}
			case model.WindowsCertificateTypeCompanyEV, model.WindowsCertificateTypeCompanyOV:
				if windowsCertificate.AppID > 0 {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"certificate cannot be use")
					return
				}
			}
		}
	}

	// 从数据中，查询文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Type,
			fileDo.AppID,
			fileDo.Name,
			fileDo.TusdID,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file not found")
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验文件。
	{
		log.Info(ctx, "verify file and signing type")
		if file.Type != model.FileTypeWindowsSigning {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file type not supported")
			return
		}
		if file.AppID != app.ID {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"file does not belong to app and user")
			return
		}
		switch req.SigningType {
		case model.WindowsSigningJobTypeHLKX:
			if strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionHLKX {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "not a hlkx file")
				return
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"not a in hlkx format file")
				return
			}
		case model.WindowsSigningJobTypePE:
			var isPEFile bool
			isPEFile, err = isFileInPEFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isPEFile {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "not a in pe format file")
				return
			}
		case model.WindowsSigningJobTypePEAndAttestation, model.WindowsSigningJobTypeAttestation:
			switch strings.ToLower(filepath.Ext(file.Name)) {
			case cc.ExtensionSYS:
				var isPEFile bool
				isPEFile, err = isFileInPEFormat(ctx, file.TusdID)
				if err != nil {
					return
				}
				if !isPEFile {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"not a in pe format file")
					return
				}
			case cc.ExtensionCAB:
				var isCabFile bool
				isCabFile, err = isFileInCabFormat(ctx, file.TusdID)
				if err != nil {
					return nil, err
				}
				if !isCabFile {
					err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
						"not a in cab format file")
				}
			default:
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"file extension not supported")
			}
		}
	}

	// 将任务信息保存到数据库。
	var jobID string
	{
		log.Info(ctx, "save windows sign job information")
		now := time.Now()
		jobID, err = generateIDWithTime(ctx, IDWindowsJob, now)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDWindowsJob, jobID), "failed to reclaim windows job id", jobID)
			}
		}()
		err = createWindowsSignJob(ctx, &model.WindowsSigningJob{
			JobID:         jobID,
			Type:          req.SigningType,
			AppID:         app.ID,
			UserID:        apiAccount.ID,
			CertificateID: windowsCertificate.ID,
			FileID:        req.FileID,
			Source:        model.SourceAPI,
			Status:        model.WindowsSigningJobStatusSigning,
			CreatedTime:   now,
		})
		if err != nil {
			return
		}
	}

	// 发送任务 MQ 消息。
	{
		log.Info(ctx, "publish windows sign job message to rabbitmq")
		queue := cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
		if slices.Contains([]int{
			model.WindowsCertificateTypePersonalEV,
			model.WindowsCertificateTypeCompanyEV,
		}, windowsCertificate.Type) {
			queue = cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + windowsCertificate.Sha1
		}
		err = publishMessageToQueue(ctx, queue, []byte(jobID))
		if err != nil {
			return
		}
	}

	rsp = &protocol.WindowsAPISubmitSigningJobRsp{JobID: jobID}

	return
}

// WindowsAPISubmitWHQLJob 提交 WHQL 任务。
func WindowsAPISubmitWHQLJob(ctx context.Context, req *protocol.WindowsAPISubmitWHQLJobReq) (
	rsp *protocol.WindowsAPISubmitWHQLJobRsp, err error) {

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
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotValid, http.StatusBadRequest)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据中，查询文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Type,
			fileDo.AppID,
			fileDo.Name,
			fileDo.TusdID,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file not found")
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验文件和服务名。
	{
		log.Info(ctx, "verify file and signing type")
		if file.Type != model.FileTypeWindowsSigning {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file type not supported")
			return
		}
		if file.AppID != app.ID {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file cannot be used")
			return
		}
		switch strings.ToLower(filepath.Ext(file.Name)) {
		case cc.ExtensionSYS:
			if req.SigningType != model.WHQLJobTypeHLKAndWHQL {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"file extension not supported, support sys or zip")
				return
			}
			var isPEFile bool
			isPEFile, err = isFileInPEFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isPEFile {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"not a in pe format file")
				return
			}
		case cc.ExtensionZIP:
			if req.SigningType != model.WHQLJobTypeHLKAndWHQL {
				return nil, errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"file extension not supported, support sys or zip")
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"not a in zip format file")
				return
			}

			// 是 zip 包时，系统服务名须提供。
			if len(req.ServiceName) <= 0 {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"serviceName is required, if file format is zip")
				return
			}
		case cc.ExtensionHLKX:
			if req.SigningType != model.WHQLJobTypeOnlyWHQL {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"file extension not supported, only support hlkx")
				return
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"not a in hlkx format file")
				return
			}
		default:
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file extension not supported")
			return
		}
	}

	// 将任务信息保存到数据库。
	var status int
	var windowsSigningJob *model.WindowsSigningJob
	var windowsCertificate *model.WindowsCertificate
	var jobID string
	{
		log.Info(ctx, "save whql job information")
		status = model.WHQLJobStatusWaitingTest
		if req.SigningType == model.WHQLJobTypeOnlyWHQL {
			status = model.WHQLJobStatusHLKXFileSinging
		}
		if status == model.WHQLJobStatusHLKXFileSinging {
			// 查找签名证书。
			log.Info(ctx, "find hlkx signing certificate from database")
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.ID,
				windowsCertificateDo.Type,
				windowsCertificateDo.Sha1,
			).Where(
				windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
				windowsCertificateDo.DeletedTime.IsNull(),
				windowsCertificateDo.NotAfter.Gte(time.Now()),
			).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(ctx, "no windows certificate available")
					err = errs.New(consts.ErrSystem)
					return
				}
				return
			}

			log.Info(ctx, "create a windows sign job to database")
			var jobID2 string
			jobID2, err = generateID(ctx, IDWindowsJob)
			if err != nil {
				return
			}
			defer func() {
				if err != nil {
					log.ErrorIf(ctx, reclaimID(ctx, IDWindowsJob, jobID2), "failed to reclaim windows sign job id", jobID2)
				}
			}()
			windowsSigningJob = &model.WindowsSigningJob{
				AppID:         app.ID,
				JobID:         jobID2,
				Type:          model.WindowsSigningJobTypeHLKX,
				UserID:        apiAccount.ID,
				CertificateID: windowsCertificate.ID,
				FileID:        req.FileID,
				Source:        model.SourceInternal,
				Status:        model.WindowsSigningJobStatusSigning,
				CreatedTime:   time.Now(),
			}
			if err = createWindowsSignJob(ctx, windowsSigningJob); err != nil {
				log.Error(ctx, "failed to create windows sign job to database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
		jobID, err = generateID(ctx, IDWHQLJob)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDWHQLJob, jobID), "failed to reclaim job id", jobID)
			}
		}()
		whqlJobDo := conn.MySQLTxClient(ctx).WhqlJob
		selectFields := make([]field.Expr, 0, 13)
		selectFields = append(selectFields,
			whqlJobDo.JobID,
			whqlJobDo.AppID,
			whqlJobDo.UserID,
			whqlJobDo.FileID,
			whqlJobDo.Type,
			whqlJobDo.Source,
			whqlJobDo.TestSystem,
			whqlJobDo.Status,
			whqlJobDo.CreatedTime,
		)
		if len(req.ServiceName) > 0 {
			selectFields = append(selectFields, whqlJobDo.ServiceName)
		}
		if len(req.TestTarget) > 0 {
			selectFields = append(selectFields, whqlJobDo.TestTarget)
		}
		if len(req.TestConfig) > 0 {
			selectFields = append(selectFields, whqlJobDo.TestConfig)
		}
		hlkxSignJobID := ""
		if status == model.WHQLJobStatusHLKXFileSinging {
			selectFields = append(selectFields, whqlJobDo.HlkxSignJobID)
			hlkxSignJobID = windowsSigningJob.JobID
		}
		err = whqlJobDo.WithContext(ctx).Select(selectFields...).Create(&model.WhqlJob{
			JobID:         jobID,
			AppID:         app.ID,
			UserID:        apiAccount.ID,
			FileID:        req.FileID,
			Type:          req.SigningType,
			Source:        model.SourceAPI,
			TestSystem:    req.TestSystem,
			TestConfig:    req.TestConfig,
			Status:        status,
			CreatedTime:   time.Now(),
			ServiceName:   req.ServiceName,
			TestTarget:    req.TestTarget,
			HlkxSignJobID: hlkxSignJobID,
		})
		if err != nil {
			log.Error(ctx, "failed to save whql job information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 发送 HLKX 文件签名消息。
	{
		if status == model.WHQLJobStatusHLKXFileSinging {
			log.Info(ctx, "send hlkx file signing message to rabbitmq")
			queue := cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
			if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalEV,
				model.WindowsCertificateTypeCompanyEV) {
				queue = cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + windowsCertificate.Sha1
			}
			err = publishMessageToQueue(ctx, queue, []byte(windowsSigningJob.JobID))
			if err != nil {
				return
			}
		}
	}

	rsp = &protocol.WindowsAPISubmitWHQLJobRsp{JobID: jobID}

	return
}

// WindowsAPIListCertificates 证书列表。
func WindowsAPIListCertificates(ctx context.Context) (rsp *protocol.WindowsAPIListCertificatesRsp, err error) {
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
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 从数据库查询应用授权使用的 EV 证书 IDs。
	var certificateIDs []int
	{
		log.Info(ctx, "get windows certificates information authorized for use by app from database")
		windowsCertificateAuthorizationDo := conn.MySQLClient(ctx).WindowsCertificateAuthorization
		err = windowsCertificateAuthorizationDo.WithContext(ctx).Select(
			windowsCertificateAuthorizationDo.CertificateID,
		).Where(
			windowsCertificateAuthorizationDo.AppID.Eq(app.ID),
		).Scan(&certificateIDs)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to get windows certificates information authorized for use by app", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 从数据库中查询应用证书信息。
	var windowsCertificates []*model.WindowsCertificate
	{
		log.Info(ctx, "get app windows certificates information")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.UserID,
			windowsCertificateDo.CertificateID,
			windowsCertificateDo.Type,
			windowsCertificateDo.Sha1,
			windowsCertificateDo.Owner,
			windowsCertificateDo.Publisher,
			windowsCertificateDo.SignatureAlgorithm,
			windowsCertificateDo.PublicKeyAlgorithm,
			windowsCertificateDo.NotBefore,
			windowsCertificateDo.NotAfter,
			windowsCertificateDo.CreatedTime,
		).Where(
			windowsCertificateDo.DeletedTime.IsNull(),
			windowsCertificateDo.WithContext(ctx).Where(
				windowsCertificateDo.AppID.Eq(app.ID),
			).Or(
				windowsCertificateDo.ID.In(certificateIDs...),
			),
		).Order(windowsCertificateDo.ID.Desc()).Find()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve app windows certificates information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 从数据库中查询用户英文名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get user's english name")
		userIDs := util.ListToUnique(windowsCertificates, func(e *model.WindowsCertificate) int {
			return e.UserID
		})
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.WindowsAPIListCertificatesItem, len(windowsCertificates))
		for i, v := range windowsCertificates {
			list[i] = &protocol.WindowsAPIListCertificatesItem{
				ID:                 v.CertificateID,
				Type:               v.Type,
				Fingerprint:        v.Sha1,
				Owner:              v.Owner,
				Publisher:          v.Publisher,
				SignatureAlgorithm: v.SignatureAlgorithm,
				PublicKeyAlgorithm: v.PublicKeyAlgorithm,
				NotBefore:          formatTime(&v.NotBefore),
				NotAfter:           formatTime(&v.NotAfter),
				CreatedTime:        formatTime(&v.CreatedTime),
				Creator:            userIDToName[v.UserID],
			}
		}
		rsp = &protocol.WindowsAPIListCertificatesRsp{List: list}
	}

	return
}

// WindowsAPIGetSigningJobInformation 获取签名任务信息。
func WindowsAPIGetSigningJobInformation(ctx context.Context, req *protocol.WindowsAPIGetSigningJobInformationReq) (
	rsp *protocol.WindowsAPIGetSigningJobInformationRsp, err error) {

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
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 查询签名任务信息。
	var windowsSigningJob *model.WindowsSigningJob
	{
		log.Info(ctx, "get windows signing job information")
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.
			Table(model.GetWindowsSigningJobByID(req.JobID))
		windowsSigningJob, err = windowsSigningJobDo.WithContext(ctx).Select(
			windowsSigningJobDo.CertificateID,
			windowsSigningJobDo.Source,
			windowsSigningJobDo.UserID,
			windowsSigningJobDo.FileID,
			windowsSigningJobDo.CreatedTime,
			windowsSigningJobDo.FinishedTime,
			windowsSigningJobDo.Log,
			windowsSigningJobDo.Status,
			windowsSigningJobDo.SignedFileID,
			windowsSigningJobDo.SubmissionID,
			windowsSigningJobDo.ProductID,
		).Where(
			windowsSigningJobDo.AppID.Eq(app.ID),
			windowsSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "job not found")
				return
			}
			log.Error(ctx, "failed to retrieve windows signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查询签名证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.CertificateID,
			windowsCertificateDo.CommonName,
		).Where(
			windowsCertificateDo.ID.Eq(windowsSigningJob.CertificateID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve windows certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if windowsCertificate == nil {
			log.Warn(ctx, "windows certificate not found", windowsSigningJob.CertificateID)
			windowsCertificate = &model.WindowsCertificate{}
		}
	}

	// 查询用户名。
	var userName string
	{
		switch windowsSigningJob.Source {
		case model.SourceAPI:
			log.Info(ctx, "get api account name")
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.AccountID,
			).Where(
				apiAccountDo.ID.Eq(windowsSigningJob.UserID),
			).Scan(&userName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve api account name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		case model.SourceWeb:
			log.Info(ctx, "get user name")
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.NameEn,
			).Where(
				userDo.ID.Eq(windowsSigningJob.UserID),
			).Scan(&userName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		default:
			log.Warn(ctx, "unknown job source", windowsSigningJob.Source)
		}
	}

	// 查询文件名。
	var fileName string
	{
		log.Info(ctx, "get file name")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(windowsSigningJob.FileID))
		err = fileDo.WithContext(ctx).Select(
			fileDo.Name,
		).Where(
			fileDo.FileID.Eq(windowsSigningJob.FileID),
		).Scan(&fileName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Warn(ctx, "file information not found in database", windowsSigningJob.FileID)
			} else {
				log.Error(ctx, "failed to retrieve file name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	rsp = &protocol.WindowsAPIGetSigningJobInformationRsp{
		SigningType:           windowsSigningJob.Type,
		Source:                windowsSigningJob.Source,
		CertificateID:         windowsCertificate.CertificateID,
		CertificateCommonName: windowsCertificate.CommonName,
		FileID:                windowsSigningJob.FileID,
		FileName:              fileName,
		UserName:              userName,
		CreatedTime:           formatTime(&windowsSigningJob.CreatedTime),
		FinishedTime:          formatTime(&windowsSigningJob.FinishedTime),
		Log:                   windowsSigningJob.Log,
		Status:                windowsSigningJob.Status,
		SignedFileID:          windowsSigningJob.SignedFileID,
		ProductID:             windowsSigningJob.ProductID,
		SubmissionID:          windowsSigningJob.SubmissionID,
	}

	return
}

// WindowsAPIGetWHQLJobInformation 获取 WHQL 任务信息。
func WindowsAPIGetWHQLJobInformation(ctx context.Context, req *protocol.WindowsAPIGetWHQLJobInformationReq) (
	rsp *protocol.WindowsAPIGetWHQLJobInformationRsp, err error) {

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
		if app.Platform != model.AppPlatformWindows {
			err = errs.NewWithStatus(consts.ErrAppPlatformNotSupported, http.StatusBadRequest)
			return
		}
	}

	// 查询 WHQL 任务信息。
	var whqlJob *model.WhqlJob
	{
		log.Info(ctx, "get whql job information")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Select(
			whqlJobDo.Type,
			whqlJobDo.Source,
			whqlJobDo.TestSystem,
			whqlJobDo.FileID,
			whqlJobDo.HlkxFileID,
			whqlJobDo.HlkLogFileID,
			whqlJobDo.SignedFileID,
			whqlJobDo.CreatedTime,
			whqlJobDo.FinishedTime,
			whqlJobDo.Status,
			whqlJobDo.Log,
			whqlJobDo.ProductID,
			whqlJobDo.SubmissionID,
			whqlJobDo.UserID,
		).Where(
			whqlJobDo.AppID.Eq(app.ID),
			whqlJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			log.Error(ctx, "failed to retrieve whql job information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查询用户名。
	var userName string
	{
		switch whqlJob.Source {
		case model.SourceAPI:
			log.Info(ctx, "get api account name")
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.AccountID,
			).Where(
				apiAccountDo.ID.Eq(whqlJob.UserID),
			).Scan(&userName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve api account name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		case model.SourceWeb:
			log.Info(ctx, "get user name")
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.NameEn,
			).Where(
				userDo.ID.Eq(whqlJob.UserID),
			).Scan(&userName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		default:
			log.Warn(ctx, "unknown job source", whqlJob.Source)
		}
	}

	// 查询文件名。
	var fileName string
	var hlkxFileName string
	var hlkLogFileName string
	var signedFileName string
	{
		log.Info(ctx, "get file names")
		fn := func(fileID string, name *string) error {
			fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(fileID))
			err2 := fileDo.WithContext(ctx).Select(
				fileDo.Name,
			).Where(
				fileDo.FileID.Eq(fileID),
			).Scan(name)
			if err2 != nil {
				if errors.Is(err2, gorm.ErrRecordNotFound) || errs.IsMySQLError(err2, mysql.ErrNoSuchTable) {
					log.Warn(ctx, "file information not found in database", fileID)
				} else {
					log.Error(ctx, "failed to retrieve file name from database", err2)
					return errs.NewWithError(consts.ErrSystem, err2)
				}
			}
			return nil
		}
		if err = fn(whqlJob.FileID, &fileName); err != nil {
			return
		}
		if err = fn(whqlJob.HlkxFileID, &hlkxFileName); err != nil {
			return
		}
		if err = fn(whqlJob.HlkLogFileID, &hlkLogFileName); err != nil {
			return
		}
		if err = fn(whqlJob.SignedFileID, &signedFileName); err != nil {
			return
		}
	}

	rsp = &protocol.WindowsAPIGetWHQLJobInformationRsp{
		Type:           whqlJob.Type,
		Source:         whqlJob.Source,
		TestSystem:     whqlJob.TestSystem,
		FileName:       fileName,
		FileID:         whqlJob.FileID,
		HLKXFileID:     whqlJob.HlkxFileID,
		HLKXFileName:   hlkxFileName,
		HLKLogFileID:   whqlJob.HlkLogFileID,
		HLKLogFileName: hlkLogFileName,
		SignedFileID:   whqlJob.SignedFileID,
		SignedFileName: signedFileName,
		UserName:       userName,
		CreatedTime:    formatTime(&whqlJob.CreatedTime),
		FinishedTime:   formatTime(&whqlJob.FinishedTime),
		Status:         whqlJob.Status,
		Log:            whqlJob.Log,
		ProductID:      whqlJob.ProductID,
		SubmissionID:   whqlJob.SubmissionID,
	}

	return
}
