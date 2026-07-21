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
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/x509"
	"debug/pe"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"software.sslmate.com/src/go-pkcs12"

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

// WindowsWebUploadCertificate 上传个人 OV 证书。
func WindowsWebUploadCertificate(ctx context.Context, req *protocol.WindowsWebUploadCertificateReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 证书解码。
	var certificateBytes []byte
	{
		log.Info(ctx, "debase64 certificate")
		certificateBytes, err = base64.StdEncoding.DecodeString(req.Certificate)
		if err != nil {
			log.Error(ctx, "failed to debase64 certificate", err)
			err = errs.NewWithError(consts.ErrParameterInvalid, err)
			return
		}
	}

	// 解析证书。
	var certificateInfo *x509.Certificate
	{
		log.Info(ctx, "parse windows certificate data")
		_, certificateInfo, err = pkcs12.Decode(certificateBytes, req.Password)
		if err != nil {
			if errors.Is(err, pkcs12.ErrIncorrectPassword) {
				err = errs.New(consts.ErrCertPasswordInvalid)
				return
			}
			if errors.Is(err, pkcs12.ErrDecryption) {
				err = errs.New(consts.ErrCertFormatInvalid)
				return
			}
			log.Error(ctx, "failed to parse windows certificate data", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// AES 加密证书内容。
	var certificateFingerprint [20]byte
	var aesKeyID int
	{
		log.Info(ctx, "encrypt windows certificate data")
		var aesKeyInfo *model.AesKey
		aesKeyInfo, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		certificateFingerprint = sha1.Sum(certificateBytes)
		if certificateBytes, err = util.AESCBCEncrypt(aesKeyInfo.Secret, certificateBytes); err != nil {
			log.Error(ctx, "encrypting windows certificate data failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		aesKeyID = aesKeyInfo.ID
	}

	// 将证书信息保存入库。
	var now time.Time
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "save windows certificate information")
		now = time.Now()
		var certificateID string
		certificateID, err = generateID(ctx, IDWindowsCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDWindowsCertificate, certificateID),
					"reclaim windows certificate id", certificateID)
			}
		}()
		windowsCertificate = &model.WindowsCertificate{
			AesKeyID:           aesKeyID,
			CertificateID:      certificateID,
			AppID:              app.ID,
			UserID:             user.ID,
			Sha1:               hex.EncodeToString(certificateFingerprint[:]),
			Type:               model.WindowsCertificateTypePersonalOV,
			Password:           req.Password,
			Version:            certificateInfo.Version,
			CommonName:         certificateInfo.Subject.CommonName,
			Publisher:          certificateInfo.Issuer.String(),
			Owner:              certificateInfo.Subject.String(),
			SignatureAlgorithm: certificateInfo.SignatureAlgorithm.String(),
			PublicKeyAlgorithm: certificateInfo.PublicKeyAlgorithm.String(),
			SerialNumber:       certificateInfo.SerialNumber.String(),
			NotBefore:          certificateInfo.NotBefore,
			NotAfter:           certificateInfo.NotAfter,
			Content:            certificateBytes,
			CreatedTime:        now,
		}
		windowsCertificateDo := conn.MySQLTxClient(ctx).WindowsCertificate
		err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.AesKeyID,
			windowsCertificateDo.CertificateID,
			windowsCertificateDo.AppID,
			windowsCertificateDo.UserID,
			windowsCertificateDo.Type,
			windowsCertificateDo.Sha1,
			windowsCertificateDo.Password,
			windowsCertificateDo.Version,
			windowsCertificateDo.Publisher,
			windowsCertificateDo.Owner,
			windowsCertificateDo.CommonName,
			windowsCertificateDo.SignatureAlgorithm,
			windowsCertificateDo.PublicKeyAlgorithm,
			windowsCertificateDo.SerialNumber,
			windowsCertificateDo.NotBefore,
			windowsCertificateDo.NotAfter,
			windowsCertificateDo.Content,
			windowsCertificateDo.CreatedTime,
		).Create(windowsCertificate)
		if err != nil {
			log.Error(ctx, "failed to save Windows certificate information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录应用事件到数据库中。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeUploadWindowsCertificate,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:                  user.NameEn,
			EventApp:                   app.Name,
			EventCertificateCommonName: windowsCertificate.CommonName,
			EventDetail: util.GetPrintJSON(map[string]any{
				"sha1":          windowsCertificate.Sha1,
				"certificateId": windowsCertificate.CertificateID,
				"publisher":     windowsCertificate.Publisher,
				"owner":         windowsCertificate.Owner,
				"notBefore":     windowsCertificate.NotBefore,
				"notAfter":      windowsCertificate.NotAfter,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// WindowsWebListCertificates 证书列表。
func WindowsWebListCertificates(ctx context.Context) (rsp *protocol.WindowsWebListCertificatesRsp, err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
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
		log.Info(ctx, "get app windows certificates information from database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
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
			windowsCertificateDo.UserID,
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
		log.Info(ctx, "get user's english name from database")
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
		list := make([]*protocol.WindowsWebListCertificatesItem, len(windowsCertificates))
		for i, v := range windowsCertificates {
			list[i] = &protocol.WindowsWebListCertificatesItem{
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
		rsp = &protocol.WindowsWebListCertificatesRsp{List: list}
	}

	return
}

// WindowsWebDownloadCertificate 下载证书。
func WindowsWebDownloadCertificate(ctx context.Context, req *protocol.WindowsWebDownloadCertificateReq) (
	fileObj *FileInfo, err error) {

	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformWindows {
			log.Error(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information from database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.Content,
			windowsCertificateDo.AesKeyID,
			windowsCertificateDo.Type,
			windowsCertificateDo.CommonName,
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
				err = errs.New(consts.ErrParameterInvalid)
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
			UserID:      user.ID,
			Type:        model.EventTypeDownloadWindowsCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:                  user.NameEn,
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

// WindowsWebAddEVCertificate 添加 EV 证书。
func WindowsWebAddEVCertificate(ctx context.Context, req *protocol.WindowsWebAddEVCertificateReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询数据库，判断不存在相同指纹的 EV 证书。
	{
		log.Info(ctx, "check whether there is a windows certificate with the same fingerprint in database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		var count int64
		count, err = windowsCertificateDo.WithContext(ctx).Where(
			windowsCertificateDo.Sha1.Eq(req.SHA1),
			windowsCertificateDo.Type.Eq(req.Type),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows certificate information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			err = errs.New(consts.ErrWindowsEVCertificateExist)
			return
		}
	}

	// 将证书信息保存到数据库。
	{
		log.Info(ctx, "save windows certificate information")
		var certificateID string
		certificateID, err = generateID(ctx, IDWindowsCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDWindowsCertificate, certificateID),
					"failed to reclaim windows certificate id", certificateID)
			}
		}()
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		selectedFields := make([]field.Expr, 0, 17)
		selectedFields = append(selectedFields,
			windowsCertificateDo.CertificateID,
			windowsCertificateDo.UserID,
			windowsCertificateDo.Sha1,
			windowsCertificateDo.Type,
			windowsCertificateDo.IP,
			windowsCertificateDo.Password,
			windowsCertificateDo.Publisher,
			windowsCertificateDo.Owner,
			windowsCertificateDo.CommonName,
			windowsCertificateDo.SignatureAlgorithm,
			windowsCertificateDo.PublicKeyAlgorithm,
			windowsCertificateDo.NotBefore,
			windowsCertificateDo.NotAfter,
			windowsCertificateDo.CreatedTime,
			windowsCertificateDo.SerialNumber,
			windowsCertificateDo.Version,
		)
		if req.IsMicrosoftVerifyCertificate {
			selectedFields = append(selectedFields, windowsCertificateDo.IsMicrosoftVerifyCertificate)
		}
		err = windowsCertificateDo.WithContext(ctx).Select(selectedFields...).Create(&model.WindowsCertificate{
			CertificateID:                certificateID,
			UserID:                       user.ID,
			Sha1:                         req.SHA1,
			Type:                         req.Type,
			IP:                           req.MachineIP,
			IsMicrosoftVerifyCertificate: model.Bool(req.IsMicrosoftVerifyCertificate),
			Password:                     req.Password,
			Publisher:                    req.Publisher,
			Owner:                        req.Owner,
			CommonName:                   getCertificateCommonName(req.Owner),
			SignatureAlgorithm:           req.SignatureAlgorithm,
			PublicKeyAlgorithm:           req.PublicKeyAlgorithm,
			NotBefore:                    time.Time(req.NotBefore),
			NotAfter:                     time.Time(req.NotAfter),
			CreatedTime:                  time.Now(),
			SerialNumber:                 req.SerialNumber,
			Version:                      req.Version,
		})
		if err != nil {
			log.Error(ctx, "failed to save windows certificate information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsWebUploadCompanyCertificate 上传公司 OV 证书。
func WindowsWebUploadCompanyCertificate(ctx context.Context, req *protocol.WindowsWebUploadCompanyCertificateReq) (
	err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 证书解码。
	var certificateBytes []byte
	{
		log.Info(ctx, "debase64 certificate")
		certificateBytes, err = base64.StdEncoding.DecodeString(req.Certificate)
		if err != nil {
			log.Error(ctx, "failed to debase64 certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解析证书。
	var certificateInfo *x509.Certificate
	{
		log.Info(ctx, "parse windows certificate information")
		_, certificateInfo, err = pkcs12.Decode(certificateBytes, req.Password)
		if err != nil {
			if errors.Is(err, pkcs12.ErrIncorrectPassword) {
				err = errs.New(consts.ErrCertPasswordInvalid)
				return
			}
			if errors.Is(err, pkcs12.ErrDecryption) {
				err = errs.New(consts.ErrCertFormatInvalid)
				return
			}
			log.Error(ctx, "failed to parse windows certificate information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 加密证书内容。
	var certificateSha1Summary [20]byte
	var aesKeyID int
	{
		log.Info(ctx, "encrypt windows certificate data")
		var aesKeyInfo *model.AesKey
		aesKeyInfo, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		certificateSha1Summary = sha1.Sum(certificateBytes)
		if certificateBytes, err = util.AESCBCEncrypt(aesKeyInfo.Secret, certificateBytes); err != nil {
			log.Error(ctx, "failed to encrypt windows certificate data", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		aesKeyID = aesKeyInfo.ID
	}

	// 将证书信息保存到数据库中。
	{
		log.Info(ctx, "save windows certificate information")
		var certificateID string
		certificateID, err = generateID(ctx, IDWindowsCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDWindowsCertificate, certificateID),
					"failed to reclaim windows certificate id", certificateID)
			}
		}()
		windowsCertificate := model.WindowsCertificate{
			AesKeyID:           aesKeyID,
			CertificateID:      certificateID,
			UserID:             user.ID,
			Sha1:               hex.EncodeToString(certificateSha1Summary[:]),
			Type:               model.WindowsCertificateTypeCompanyOV,
			Password:           req.Password,
			Version:            certificateInfo.Version,
			CommonName:         certificateInfo.Subject.CommonName,
			Publisher:          certificateInfo.Issuer.String(),
			Owner:              certificateInfo.Subject.String(),
			SignatureAlgorithm: certificateInfo.SignatureAlgorithm.String(),
			PublicKeyAlgorithm: certificateInfo.PublicKeyAlgorithm.String(),
			SerialNumber:       certificateInfo.SerialNumber.String(),
			NotBefore:          certificateInfo.NotBefore,
			NotAfter:           certificateInfo.NotAfter,
			Content:            certificateBytes,
			CreatedTime:        time.Now(),
		}
		windowsCertificateTxDo := conn.MySQLTxClient(ctx).WindowsCertificate
		err = windowsCertificateTxDo.WithContext(ctx).Select(
			windowsCertificateTxDo.AesKeyID,
			windowsCertificateTxDo.CertificateID,
			windowsCertificateTxDo.UserID,
			windowsCertificateTxDo.Type,
			windowsCertificateTxDo.Sha1,
			windowsCertificateTxDo.Password,
			windowsCertificateTxDo.Version,
			windowsCertificateTxDo.Publisher,
			windowsCertificateTxDo.Owner,
			windowsCertificateTxDo.CommonName,
			windowsCertificateTxDo.SignatureAlgorithm,
			windowsCertificateTxDo.PublicKeyAlgorithm,
			windowsCertificateTxDo.SerialNumber,
			windowsCertificateTxDo.NotBefore,
			windowsCertificateTxDo.NotAfter,
			windowsCertificateTxDo.Content,
			windowsCertificateTxDo.CreatedTime,
		).Create(&windowsCertificate)
		if err != nil {
			log.Error(ctx, "failed to save windows certificate information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsWebListCompanyCertificates 获取后台管理中的 Windows 证书。
func WindowsWebListCompanyCertificates(ctx context.Context) (
	rsp *protocol.WindowsWebListCompanyCertificatesRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取证书信息。
	var windowsCertificates []*model.WindowsCertificate
	{
		log.Info(ctx, "get windows company certificates from database")
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
			windowsCertificateDo.Password,
			windowsCertificateDo.IsMicrosoftVerifyCertificate,
			windowsCertificateDo.IP,
		).Where(
			windowsCertificateDo.AppID.IsNull(),
			windowsCertificateDo.DeletedTime.IsNull(),
			windowsCertificateDo.Type.In(model.WindowsCertificateTypeCompanyEV,
				model.WindowsCertificateTypeCompanyOV, model.WindowsCertificateTypePersonalEV),
		).Order(windowsCertificateDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows company certificates from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查询用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDs := util.ListToUnique(windowsCertificates, func(e *model.WindowsCertificate) int { return e.UserID })
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.WindowsWebListCompanyCertificatesItem, len(windowsCertificates))
		for i, v := range windowsCertificates {
			list[i] = &protocol.WindowsWebListCompanyCertificatesItem{
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
				Password:           v.Password,
				IsMSVerification:   bool(v.IsMicrosoftVerifyCertificate),
				MachineIP:          v.IP,
			}
		}
		rsp = &protocol.WindowsWebListCompanyCertificatesRsp{List: list}
	}

	return
}

// WindowsWebGrantAppEVCertificate 授权应用使用个人 EV 证书。
func WindowsWebGrantAppEVCertificate(ctx context.Context, req *protocol.WindowsWebGrantAppEVCertificateReq) (
	err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取应用信息。
	var appID int
	{
		log.Info(ctx, "get app information")
		appDo := conn.MySQLClient(ctx).App
		err = appDo.WithContext(ctx).Select(
			appDo.ID,
		).Where(
			appDo.AppID.Eq(req.AppID),
			appDo.Platform.Eq(model.AppPlatformWindows),
		).Scan(&appID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "app does not exist")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve app information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取证书信息。
	var certificateID int
	{
		log.Info(ctx, "get certificate information")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.ID,
		).Where(
			windowsCertificateDo.CertificateID.Eq(req.CertificateID),
			windowsCertificateDo.DeletedTime.IsNull(),
			windowsCertificateDo.AppID.IsNull(),
			windowsCertificateDo.Type.Eq(model.WindowsCertificateTypePersonalEV),
		).Scan(&certificateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "certificate does not exist")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve certificate information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 添加应用证书授权记录。
	{
		log.Info(ctx, "add app certificate authorization record")
		windowsCertificateAuthorizationDo := conn.MySQLClient(ctx).WindowsCertificateAuthorization
		err = windowsCertificateAuthorizationDo.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
			Create(&model.WindowsCertificateAuthorization{
				AppID:         appID,
				UserID:        user.ID,
				CertificateID: certificateID,
				CreatedTime:   time.Now(),
			})
		if err != nil {
			log.Error(ctx, "failed to create certificate authorization information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsWebGetCertificatePassword 查看证书密码。
func WindowsWebGetCertificatePassword(ctx context.Context, req *protocol.WindowsWebGetCertificatePasswordReq) (
	rsp *protocol.WindowsWebGetCertificatePasswordRsp, err error) {

	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information")
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
				err = errs.New(consts.ErrParameterInvalid)
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
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		}
	}

	rsp = &protocol.WindowsWebGetCertificatePasswordRsp{Password: windowsCertificate.Password}

	return
}

// WindowsWebDownloadCompanyCertificate 下载公司证书。
func WindowsWebDownloadCompanyCertificate(ctx context.Context, req *protocol.WindowsWebDownloadCompanyCertificateReq) (
	fileObj *FileInfo, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库获取证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information from database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.AesKeyID,
			windowsCertificateDo.Content,
			windowsCertificateDo.CommonName,
		).Where(
			windowsCertificateDo.CertificateID.Eq(req.CertificateID),
			windowsCertificateDo.Type.Eq(model.WindowsCertificateTypeCompanyOV),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "windows company certificate does not exist")
				err = errs.New(consts.ErrParameterInvalid)
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
			log.Error(ctx, "failed to retrieve aes key information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		certificateBytes, err = util.AESCBCDecrypt(windowsCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt windows company certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
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

// WindowsWebListGrantCertificateApps 获取授权 Windows EV 证书应用列表。
func WindowsWebListGrantCertificateApps(ctx context.Context, req *protocol.WindowsWebListGrantCertificateAppsReq) (
	rsp *protocol.WindowsWebListGrantCertificateAppsRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取授权记录。
	var windowsCertificateAuthorizations []*model.WindowsCertificateAuthorization
	var count int64
	{
		log.Info(ctx, "get windows certificate authorizations information")
		windowsCertificateAuthorizationDo := conn.MySQLClient(ctx).WindowsCertificateAuthorization
		conditions := make([]gen.Condition, 0, 2)
		if len(req.AppID) > 0 {
			// 获取应用信息。
			log.Info(ctx, "get app information")
			appDo := conn.MySQLClient(ctx).App
			var appID int
			err = appDo.WithContext(ctx).Select(
				appDo.ID,
			).Where(
				appDo.AppID.Eq(req.AppID),
				appDo.Platform.Eq(model.AppPlatformWindows),
			).Scan(&appID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "app does not exist")
					err = nil
					return
				}
				log.Error(ctx, "failed to retrieve app information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			conditions = append(conditions, windowsCertificateAuthorizationDo.AppID.Eq(appID))
		}
		if len(req.CertificateID) > 0 {
			// 获取证书信息。
			log.Info(ctx, "get certificate information")
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			var windowsCertificateID int
			err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.ID,
			).Where(
				windowsCertificateDo.CertificateID.Eq(req.CertificateID),
				windowsCertificateDo.DeletedTime.IsNull(),
				windowsCertificateDo.AppID.IsNull(),
				windowsCertificateDo.Type.Eq(model.WindowsCertificateTypePersonalEV),
			).Scan(&windowsCertificateID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "certificate does not exist")
					err = nil
					return
				}
				log.Error(ctx, "failed to retrieve certificate information", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			conditions = append(conditions, windowsCertificateAuthorizationDo.CertificateID.Eq(windowsCertificateID))
		}
		count, err = windowsCertificateAuthorizationDo.WithContext(ctx).Where(conditions...).Order(
			windowsCertificateAuthorizationDo.CreatedTime.Desc(),
			windowsCertificateAuthorizationDo.ID.Desc(),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count windows certificate authorization information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		windowsCertificateAuthorizations, err = windowsCertificateAuthorizationDo.WithContext(ctx).
			Select(
				windowsCertificateAuthorizationDo.CertificateID,
				windowsCertificateAuthorizationDo.AppID,
				windowsCertificateAuthorizationDo.UserID,
				windowsCertificateAuthorizationDo.CreatedTime,
			).Where(conditions...).Order(
			windowsCertificateAuthorizationDo.CreatedTime.Desc(),
			windowsCertificateAuthorizationDo.ID.Desc(),
		).Limit(req.PageSize).Offset((req.PageNumber - 1) * req.PageSize).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows certificate authorization information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(windowsCertificateAuthorizations) <= 0 {
			rsp = &protocol.WindowsWebListGrantCertificateAppsRsp{Count: count}
			return
		}
	}

	// 获取用户名。
	var userIDToName map[int]string
	var appIDs, certificateIDs []int
	{
		log.Info(ctx, "get user information")
		userIDs := make([]int, len(windowsCertificateAuthorizations))
		appIDs = make([]int, len(windowsCertificateAuthorizations))
		certificateIDs = make([]int, len(windowsCertificateAuthorizations))
		for i, v := range windowsCertificateAuthorizations {
			userIDs[i] = v.UserID
			appIDs[i] = v.AppID
			certificateIDs[i] = v.CertificateID
		}
		userIDs = util.CleanNumbers(userIDs)
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 查询应用信息。
	var appIDToInfo map[int]*model.App
	{
		log.Info(ctx, "get app information")
		appDo := conn.MySQLClient(ctx).App
		var apps []*model.App
		apps, err = appDo.WithContext(ctx).Select(
			appDo.ID,
			appDo.Name,
			appDo.AppID,
		).Where(
			appDo.ID.In(appIDs...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve app information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		appIDToInfo = util.ListAssociateBy(apps, func(e *model.App) int { return e.ID })
	}

	// 查询证书信息。
	var windowsCertificateIDToInfo map[int]*model.WindowsCertificate
	{
		log.Info(ctx, "get certificate information")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		var windowsCertificates []*model.WindowsCertificate
		windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.ID,
			windowsCertificateDo.CertificateID,
			windowsCertificateDo.CommonName,
		).Where(
			windowsCertificateDo.ID.In(certificateIDs...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows certificate information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		windowsCertificateIDToInfo = util.ListAssociateBy(windowsCertificates,
			func(e *model.WindowsCertificate) int { return e.ID })
	}

	// 组装数据。
	{
		list := make([]*protocol.WindowsWebListGrantCertificateAppsItem, len(windowsCertificateAuthorizations))
		for i, v := range windowsCertificateAuthorizations {
			appID := ""
			appName := ""
			if info, ok := appIDToInfo[v.AppID]; ok {
				appID = info.AppID
				appName = info.Name
			}
			certificateID := ""
			certificateOrganization := ""
			if info, ok := windowsCertificateIDToInfo[v.CertificateID]; ok {
				certificateID = info.CertificateID
				certificateOrganization = info.CommonName
			}
			list[i] = &protocol.WindowsWebListGrantCertificateAppsItem{
				AppID:                   appID,
				CertificateID:           certificateID,
				AppName:                 appName,
				CertificateOrganization: certificateOrganization,
				GrantTime:               formatTime(&v.CreatedTime),
				User:                    userIDToName[v.UserID],
			}
		}
		rsp = &protocol.WindowsWebListGrantCertificateAppsRsp{Count: count, List: list}
	}

	return
}

// WindowsWebSubmitSigningJob 提交 Windows 签名任务。
func WindowsWebSubmitSigningJob(ctx context.Context, req *protocol.WindowsWebSubmitSigningJobReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中查询证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information")
		if req.SigningType == model.WindowsSigningJobTypeAttestation {
			log.Info(ctx, "find cab signing certificate from database")
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.Type,
				windowsCertificateDo.ID,
				windowsCertificateDo.Sha1,
				windowsCertificateDo.AppID,
			).Where(
				windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
				windowsCertificateDo.DeletedTime.IsNull(),
				windowsCertificateDo.NotAfter.Gte(time.Now()),
			).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(ctx, "no windows certificate available")
					err = errs.New(consts.ErrParameterInvalid)
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
				windowsCertificateDo.ID,
				windowsCertificateDo.Sha1,
				windowsCertificateDo.AppID,
			).Where(
				windowsCertificateDo.CertificateID.Eq(req.CertificateID),
				windowsCertificateDo.WithContext(ctx).Where(windowsCertificateDo.AppID.Eq(app.ID)).
					Or(windowsCertificateDo.AppID.IsNull()),
				windowsCertificateDo.DeletedTime.IsNull(),
			).Take()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "windows company certificate does not exist")
					err = errs.New(consts.ErrParameterInvalid)
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
					err = errs.New(consts.ErrParameterInvalid)
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
					log.Warn(ctx, "windows company certificate is not authorized for app")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
			case model.WindowsCertificateTypePersonalOV:
				if windowsCertificate.AppID != app.ID {
					log.Warn(ctx, "app does not have the windows certificate")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
			case model.WindowsCertificateTypeCompanyEV, model.WindowsCertificateTypeCompanyOV:
				if windowsCertificate.AppID > 0 {
					log.Warn(ctx, "app can not use the windows certificate")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
			}
		}
	}

	// 从数据中，查询文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information from database")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Type,
			fileDo.UserID,
			fileDo.AppID,
			fileDo.Name,
			fileDo.TusdID,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Warn(ctx, "file does not exist")
				err = errs.New(consts.ErrParameterInvalid)
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
			log.Warn(ctx, "file type is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if file.AppID != app.ID || file.UserID != user.ID {
			log.Warn(ctx, "file does not belong to app and user")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		switch req.SigningType {
		case model.WindowsSigningJobTypeHLKX:
			if strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionHLKX {
				log.Warn(ctx, "file is not a hlkx")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				log.Warn(ctx, "file is not a valid hlkx")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case model.WindowsSigningJobTypePE:
			var isPEFile bool
			isPEFile, err = isFileInPEFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isPEFile {
				log.Warn(ctx, "file is not a pe format")
				err = errs.New(consts.ErrParameterInvalid)
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
					log.Warn(ctx, "file is not a pe format")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
			case cc.ExtensionCAB:
				var isCabFile bool
				isCabFile, err = isFileInCabFormat(ctx, file.TusdID)
				if err != nil {
					return
				}
				if !isCabFile {
					log.Warn(ctx, "file is not a valid cab")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
			default:
				log.Warn(ctx, "file is neither a sys format nor a cab format")
				err = errs.New(consts.ErrParameterInvalid)
				return
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
			UserID:        user.ID,
			CertificateID: windowsCertificate.ID,
			FileID:        req.FileID,
			Source:        model.SourceWeb,
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

	return
}

// WindowsWebListSigningJobs 获取签名任务列表。
func WindowsWebListSigningJobs(ctx context.Context, req *protocol.WindowsWebListSigningJobsReq) (
	rsp *protocol.WindowsWebListSigningJobsRsp, err error) {

	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查询证书 IDs。
	var certificateIDs []int
	{
		if len(req.KeyWord) > 0 {
			log.Info(ctx, "get windows certificate ids from database")
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.ID,
			).Where(
				windowsCertificateDo.CommonName.Like("%" + req.KeyWord + "%"),
			).Scan(&certificateIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "getting windows certificate ids from database failed", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 查询用户 IDs
	var userIDs []int
	{
		if len(req.KeyWord) > 0 {
			log.Info(ctx, "get user ids")
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.ID,
			).Where(
				userDo.NameEn.Like("%" + req.KeyWord + "%"),
			).Or(
				userDo.NameZh.Like("%" + req.KeyWord + "%"),
			).Scan(&userIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "getting user ids from database failed", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			log.Info(ctx, "get api account ids")
			var apiAccountIDs []int
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.ID,
			).Where(
				apiAccountDo.AccountID.Like("%" + req.KeyWord + "%"),
			).Scan(&apiAccountIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "getting api account ids from database failed", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			userIDs = append(userIDs, apiAccountIDs...)
		}
	}

	// 查询数据库，获取任务信息。
	var count int
	var windowsSigningJobs []*model.WindowsSigningJob
	{
		log.Info(ctx, "get windows sign jobs information")
		var tableNames []string
		tableNames, err = getAllWindowsSignJobTableNames(ctx)
		if err != nil {
			return
		}
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
		count, err = windowsSigningJobDo.WithContext(ctx).Count2(
			tableNames,
			app.ID,
			req.KeyWord,
			req.SigningType,
			req.Status,
			certificateIDs,
			userIDs,
		)
		if err != nil {
			log.Error(ctx, "failed to count windows sign jobs numbers from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		windowsSigningJobs, err = windowsSigningJobDo.WithContext(ctx).List(
			tableNames,
			app.ID,
			req.KeyWord,
			req.SigningType,
			req.Status,
			certificateIDs,
			userIDs,
			req.PageSize,
			(req.PageNumber-1)*req.PageSize,
		)
		if err != nil {
			log.Error(ctx, "failed to retrieve windows sign jobs information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(windowsSigningJobs) <= 0 {
			rsp = &protocol.WindowsWebListSigningJobsRsp{Count: count}
			return
		}
	}

	// 查询任务的提交人。
	var certificateIDs2 []int
	var fileIDs []string
	var userIDToName map[int]string
	var apiAccountIDToName map[int]string
	{
		log.Info(ctx, "get windows sign jobs user name from database")
		userIDs2 := make([]int, 0, len(windowsSigningJobs)/2)
		apiAccountIDs := make([]int, 0, len(windowsSigningJobs)/2)
		certificateIDs2 = make([]int, 0, len(windowsSigningJobs))
		fileIDs = make([]string, 0, len(windowsSigningJobs))
		for _, v := range windowsSigningJobs {
			switch v.Source {
			case model.SourceWeb:
				userIDs2 = append(userIDs2, v.UserID)
			case model.SourceAPI:
				apiAccountIDs = append(apiAccountIDs, v.UserID)
			default:
				log.Warn(ctx, "unsupported source type", v.Source)
			}
			certificateIDs2 = append(certificateIDs2, v.CertificateID)
			fileIDs = append(fileIDs, v.FileID)
		}
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs2))
		if err != nil {
			return
		}
		apiAccountIDToName, err = GetAPIAccountNamesByIDs(ctx, util.CleanNumbers(apiAccountIDs))
		if err != nil {
			return
		}
	}

	// 查询证书信息。
	var certificateIDToInfo map[int]*model.WindowsCertificate
	{
		log.Info(ctx, "get windows certificate information from database")
		certificateIDs2 = util.CleanNumbers(certificateIDs2)
		certificateIDToInfo = make(map[int]*model.WindowsCertificate, len(certificateIDs2))
		if len(certificateIDs2) > 0 {
			var windowsCertificates []*model.WindowsCertificate
			windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
			windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
				windowsCertificateDo.ID,
				windowsCertificateDo.CertificateID,
				windowsCertificateDo.CommonName,
			).Where(
				windowsCertificateDo.ID.In(certificateIDs2...),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve windows certificate information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			for _, v := range windowsCertificates {
				certificateIDToInfo[v.ID] = v
			}
		}
	}

	// 查询文件信息。
	var fileIDToName map[string]string
	{
		log.Info(ctx, "get file information from database")
		fileIDToName, err = GetFileNamesByIDs(ctx, util.CleanStrings(fileIDs))
		if err != nil {
			return
		}
	}

	// 整理数据。
	{
		list := make([]*protocol.WindowsWebListSigningJobsItem, len(windowsSigningJobs))
		for i, v := range windowsSigningJobs {
			userName := ""
			switch v.Source {
			case model.SourceWeb:
				userName = userIDToName[v.UserID]
			case model.SourceAPI:
				userName = apiAccountIDToName[v.UserID]
			default:
				log.Warn(ctx, "unsupported source type", v.Source)
			}
			certificateInfo := certificateIDToInfo[v.CertificateID]
			if certificateInfo == nil {
				certificateInfo = &model.WindowsCertificate{}
			}
			list[i] = &protocol.WindowsWebListSigningJobsItem{
				JobID:                 v.JobID,
				SigningType:           v.Type,
				Source:                v.Source,
				CertificateID:         certificateInfo.CertificateID,
				CertificateCommonName: certificateInfo.CommonName,
				FileID:                v.FileID,
				FileName:              fileIDToName[v.FileID],
				UserName:              userName,
				CreatedTime:           formatTime(&v.CreatedTime),
				FinishedTime:          formatTime(&v.FinishedTime),
				Log:                   v.Log,
				Status:                v.Status,
				SignedFileID:          v.SignedFileID,
			}
		}
		rsp = &protocol.WindowsWebListSigningJobsRsp{List: list, Count: count}
	}

	return
}

// WindowsWebSubmitWHQLJob 提交 WHQL 任务。
func WindowsWebSubmitWHQLJob(ctx context.Context, req *protocol.WindowsWebSubmitWHQLJobReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据中，查询文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information from database")
		fileQuery := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileQuery.WithContext(ctx).Select(
			fileQuery.Name,
			fileQuery.Type,
			fileQuery.UserID,
			fileQuery.TusdID,
			fileQuery.AppID,
		).Where(
			fileQuery.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "file does not exist")
				err = errs.New(consts.ErrParameterInvalid)
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
			log.Warn(ctx, "file type is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if file.AppID != app.ID || file.UserID != user.ID {
			log.Warn(ctx, "file does not belong to app and user")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		switch strings.ToLower(filepath.Ext(file.Name)) {
		case cc.ExtensionSYS:
			if req.SigningType != model.WHQLJobTypeHLKAndWHQL {
				log.Warn(ctx, "sys file should be hlk")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			var isPEFile bool
			isPEFile, err = isFileInPEFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isPEFile {
				log.Warn(ctx, "file is not a pe format")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case cc.ExtensionZIP:
			if req.SigningType != model.WHQLJobTypeHLKAndWHQL {
				log.Warn(ctx, "zip file should be hlk")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				log.Warn(ctx, "file is not a zip format")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 是 zip 包时，系统服务名须提供。
			if len(req.ServiceName) <= 0 {
				log.Warn(ctx, "service name not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case cc.ExtensionHLKX:
			if req.SigningType != model.WHQLJobTypeOnlyWHQL {
				log.Warn(ctx, "hlkx file should be whql")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			var isZipFile bool
			isZipFile, err = isFileInZipFormat(ctx, file.TusdID)
			if err != nil {
				return
			}
			if !isZipFile {
				log.Warn(ctx, "file is not a zip format")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		default:
			log.Warn(ctx, "file type is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 将任务信息保存到数据库。
	var jobID string
	var status int
	var windowsSigningJob *model.WindowsSigningJob
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "save whql job information to database")
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
				log.Error(ctx, "failed to retrieve windows certificate information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
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
					log.ErrorIf(ctx, reclaimID(ctx, IDWindowsJob, jobID2),
						"failed to reclaim windows sign job id", jobID2)
				}
			}()
			windowsSigningJob = &model.WindowsSigningJob{
				AppID:         app.ID,
				JobID:         jobID2,
				Type:          model.WindowsSigningJobTypeHLKX,
				UserID:        user.ID,
				CertificateID: windowsCertificate.ID,
				FileID:        req.FileID,
				Source:        model.SourceInternal,
				Status:        model.WindowsSigningJobStatusSigning,
				CreatedTime:   time.Now(),
			}
			if err = createWindowsSignJob(ctx, windowsSigningJob); err != nil {
				log.Error(ctx, "failed to create windows sign job to database", err)
				return errs.NewWithError(consts.ErrSystem, err)
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
			UserID:        user.ID,
			FileID:        req.FileID,
			Type:          req.SigningType,
			Source:        model.SourceWeb,
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
			if util.In(windowsCertificate.Type,
				model.WindowsCertificateTypePersonalEV, model.WindowsCertificateTypeCompanyEV) {
				queue = cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + windowsCertificate.Sha1
			}
			err = publishMessageToQueue(ctx, queue, []byte(windowsSigningJob.JobID))
			if err != nil {
				return
			}
		}
	}

	return
}

// WindowsWebListWHQLJobs 获取 WHQL 任务列表。
func WindowsWebListWHQLJobs(ctx context.Context, req *protocol.WindowsWebListWHQLJobsReq) (
	rsp *protocol.WindowsWebListWHQLJobsRsp, err error) {

	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库获取任务信息。
	var count int64
	var whqlJobs []*model.WhqlJob
	{
		log.Info(ctx, "get whql jobs information from database")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		count, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.AppID.Eq(app.ID),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count whql jobs in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		whqlJobs, err = whqlJobDo.WithContext(ctx).Select(
			whqlJobDo.Source,
			whqlJobDo.UserID,
			whqlJobDo.FileID,
			whqlJobDo.SignedFileID,
			whqlJobDo.HlkxSignJobID,
			whqlJobDo.HlkxFileID,
			whqlJobDo.HlkLogFileID,
			whqlJobDo.JobID,
			whqlJobDo.Type,
			whqlJobDo.Source,
			whqlJobDo.TestSystem,
			whqlJobDo.Log,
			whqlJobDo.Status,
			whqlJobDo.FinishedTime,
			whqlJobDo.CreatedTime,
		).Where(
			whqlJobDo.AppID.Eq(app.ID),
		).Order(whqlJobDo.CreatedTime.Desc(), whqlJobDo.ID.Desc()).
			Limit(req.PageSize).Offset((req.PageNumber - 1) * req.PageSize).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve whql jobs in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(whqlJobs) <= 0 {
			rsp = &protocol.WindowsWebListWHQLJobsRsp{Count: count}
			return
		}
	}

	// 查库，获取文件信息。
	var userIDs, apiAccountIDs []int
	var fileIDToInfo map[string]string
	{
		log.Info(ctx, "get file information from database")
		fileIDs := make([]string, 0, len(whqlJobs))
		userIDs = make([]int, 0, len(whqlJobs)/2)
		apiAccountIDs = make([]int, 0, len(whqlJobs)/2)
		for _, v := range whqlJobs {
			switch v.Source {
			case model.SourceWeb:
				userIDs = append(userIDs, v.UserID)
			case model.SourceAPI:
				apiAccountIDs = append(apiAccountIDs, v.UserID)
			default:
				log.Warn(ctx, "unknown source", v.Source)
			}
			fileIDs = append(fileIDs, v.FileID, v.SignedFileID, v.HlkLogFileID, v.HlkxFileID)
		}
		fileIDToInfo, err = GetFileNamesByIDs(ctx, util.CleanStrings(fileIDs))
		if err != nil {
			return
		}
	}

	// 查询任务的提交人。
	var userIDToName map[int]string
	var apiAccountIDToName map[int]string
	{
		log.Info(ctx, "get windows sign jobs user name from database")
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
		apiAccountIDToName, err = GetAPIAccountNamesByIDs(ctx, util.CleanNumbers(apiAccountIDs))
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.WindowsWebListWHQLJobsItem, len(whqlJobs))
		for i, v := range whqlJobs {
			userName := ""
			switch v.Source {
			case model.SourceWeb:
				userName = userIDToName[v.UserID]
			case model.SourceAPI:
				userName = apiAccountIDToName[v.UserID]
			default:
				log.Warn(ctx, "unknown source", v.Source)
			}
			list[i] = &protocol.WindowsWebListWHQLJobsItem{
				JobID:          v.JobID,
				Type:           v.Type,
				Source:         v.Source,
				TestSystem:     v.TestSystem,
				FileName:       fileIDToInfo[v.FileID],
				FileID:         v.FileID,
				HLKXFileID:     v.HlkxFileID,
				HLKXFileName:   fileIDToInfo[v.HlkxFileID],
				HLKLogFileID:   v.HlkLogFileID,
				HLKLogFileName: fileIDToInfo[v.HlkLogFileID],
				SignedFileID:   v.SignedFileID,
				SignedFileName: fileIDToInfo[v.SignedFileID],
				UserName:       userName,
				CreatedTime:    formatTime(&v.CreatedTime),
				FinishedTime:   formatTime(&v.FinishedTime),
				Status:         v.Status,
				Log:            v.Log,
			}
		}
		rsp = &protocol.WindowsWebListWHQLJobsRsp{Count: count, List: list}
	}

	return
}

// WindowsWebRemoveCompanyCertificate 删除公司证书。
func WindowsWebRemoveCompanyCertificate(ctx context.Context, req *protocol.WindowsWebRemoveCompanyCertificateReq) (
	err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取证书信息。
	var windowsCertificate *model.WindowsCertificate
	{
		log.Info(ctx, "get windows information from database")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.Type,
			windowsCertificateDo.ID,
		).Where(
			windowsCertificateDo.CertificateID.Eq(req.CertificateID),
			windowsCertificateDo.Type.In(model.WindowsCertificateTypeCompanyOV,
				model.WindowsCertificateTypePersonalEV, model.WindowsCertificateTypeCompanyEV),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Info(ctx, "windows company certificate does not exist")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 如果是个人 EV 证书，同时删除授权信息。
	{
		if windowsCertificate.Type == model.WindowsCertificateTypePersonalEV {
			log.Info(ctx, "remove windows certificate authorization information")
			windowsCertificateAuthorizationTxDo := conn.MySQLTxClient(ctx).WindowsCertificateAuthorization
			_, err = windowsCertificateAuthorizationTxDo.WithContext(ctx).Where(
				windowsCertificateAuthorizationTxDo.CertificateID.Eq(windowsCertificate.ID),
			).Delete()
			if err != nil {
				log.Error(ctx, "failed to remove windows certificate authorization information", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 删除证书。
	{
		log.Info(ctx, "update windows certificate information in database")
		windowsCertificateTxDo := conn.MySQLTxClient(ctx).WindowsCertificate
		var sqlResult gen.ResultInfo
		sqlResult, err = windowsCertificateTxDo.WithContext(ctx).Where(
			windowsCertificateTxDo.ID.Eq(windowsCertificate.ID),
		).UpdateColumnSimple(
			windowsCertificateTxDo.DeletedTime.Value(time.Now()),
		)
		if err != nil {
			log.Error(ctx, "failed to update windows certificate information in database", err)
			err = errs.New(consts.ErrSystem)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "delete windows certificate no effect")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	return
}

// WindowsWebDeleteCertificate 删除证书。
func WindowsWebDeleteCertificate(ctx context.Context, req *protocol.WindowsWebDeleteCertificateReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	var app *model.App
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		if app.Platform != model.AppPlatformWindows {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
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
			windowsCertificateDo.Owner,
			windowsCertificateDo.CommonName,
			windowsCertificateDo.Sha1,
			windowsCertificateDo.Publisher,
			windowsCertificateDo.NotBefore,
			windowsCertificateDo.NotAfter,
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
				log.Warn(ctx, "windows certificate does not exist")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 删除证书。
	var now time.Time
	{
		log.Info(ctx, "delete windows certificate")
		now = time.Now()
		switch windowsCertificate.Type {
		case model.WindowsCertificateTypePersonalOV:
			windowsCertificateTxDo := conn.MySQLTxClient(ctx).WindowsCertificate
			var sqlResult gen.ResultInfo
			windowsCertificateTxDo = conn.MySQLTxClient(ctx).WindowsCertificate
			sqlResult, err = windowsCertificateTxDo.WithContext(ctx).Where(
				windowsCertificateTxDo.ID.Eq(windowsCertificate.ID),
			).UpdateColumnSimple(
				windowsCertificateTxDo.DeletedTime.Value(now),
			)
			if err != nil {
				log.Error(ctx, "failed to update windows certificate information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if sqlResult.RowsAffected <= 0 {
				log.Warn(ctx, "no windows certificate information has been updated")
				err = errs.New(consts.ErrCommonFailure)
				return
			}
		case model.WindowsCertificateTypePersonalEV:
			windowsCertificateAuthorizationTxDo := conn.MySQLTxClient(ctx).WindowsCertificateAuthorization
			var sqlResult gen.ResultInfo
			sqlResult, err = windowsCertificateAuthorizationTxDo.WithContext(ctx).Where(
				windowsCertificateAuthorizationTxDo.AppID.Eq(app.ID),
				windowsCertificateAuthorizationTxDo.CertificateID.Eq(windowsCertificate.ID),
			).Delete()
			if err != nil {
				log.Error(ctx, "failed to delete windows certificate authorization information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if sqlResult.RowsAffected <= 0 {
				log.Warn(ctx, "no windows certificate authorization information has been deleted")
				err = errs.New(consts.ErrCommonFailure)
				return
			}
		default:
			log.Error(ctx, "unhandled windows certificate type", windowsCertificate.Type)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 记录应用事件到数据库中。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemoveWindowsCertificate,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:                  user.NameEn,
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

	return
}

// WindowsWebStatisticSigningTimes 获取应用的 Windows 类型签名次数统计信息。
func WindowsWebStatisticSigningTimes(ctx context.Context, req *protocol.WindowsWebStatisticSigningTimesReq) (
	rsp *protocol.WindowsWebStatisticSigningTimesRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询应用信息。
	var appID int
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app information")
			appQuery := conn.MySQLClient(ctx).App
			err = appQuery.WithContext(ctx).Select(
				appQuery.ID,
			).Where(
				appQuery.AppID.Eq(req.AppID),
			).Scan(&appID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "app not found")
					return nil, errs.New(consts.ErrParameterInvalid)
				}
				log.Error(ctx, "failed to retrieve app information from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 查询数据库，获取任务数量。
	var windowsSigningJobSqlResult []map[string]any
	{
		log.Info(ctx, "get times of windows job")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		var tableNames []string
		tableNames, err = filterWindowsSigningJobTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) > 0 {
			windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
			switch req.TimeStep {
			case protocol.TimeStepDay:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CountWithDay(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepWeek:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CountWithWeek(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepMonth:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CountWithMonth(tableNames, appID, req.BeginTime, req.EndTime)
			default:
				log.Warn(ctx, "unknown time step", req.TimeStep)
				return nil, errs.New(consts.ErrParameterInvalid)
			}
			if err != nil {
				log.Error(ctx, "failed to count windows job from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 转换数据。
	var windowsSigningJobItems []*protocol.WindowsWebStatisticSigningTimesItem
	{
		log.Info(ctx, "deal sql data")
		windowsSigningJobItems = make([]*protocol.WindowsWebStatisticSigningTimesItem, 0, len(windowsSigningJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningTimesItem{}
		for _, v := range windowsSigningJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var count int
			count, err = strconv.Atoi(fmt.Sprintf("%v", v["count"]))
			if err != nil {
				log.Error(ctx, "failed to convert count to int", err, v["count"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningTimesItem{BeginTime: beginTime}
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			switch typ {
			case model.WindowsSigningJobTypePE:
				item.PESigningTimes = count
			case model.WindowsSigningJobTypeAttestation:
				item.AttestationSigningTimes = count
			case model.WindowsSigningJobTypePEAndAttestation:
				item.PEAndAttestationSigningTimes = count
			default: // noop
			}
		}
	}

	// 查询数据库，获取 WHQL 任务数量。
	var whqlJobSqlResult []map[string]any
	{
		log.Info(ctx, "get times of whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		switch req.TimeStep {
		case protocol.TimeStepDay:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CountWithDay(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CountWithWeek(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CountWithMonth(appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to count windows job from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 转换数据。
	var whqlJobItems []*protocol.WindowsWebStatisticSigningTimesItem
	{
		log.Info(ctx, "deal sql data")
		whqlJobItems = make([]*protocol.WindowsWebStatisticSigningTimesItem, 0, len(whqlJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningTimesItem{}
		for _, v := range whqlJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var count int
			count, err = strconv.Atoi(fmt.Sprintf("%v", v["count"]))
			if err != nil {
				log.Error(ctx, "failed to convert count to int", err, v["count"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				whqlJobItems = append(whqlJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningTimesItem{BeginTime: beginTime}
				whqlJobItems = append(whqlJobItems, item)
			}
			switch typ {
			case model.WHQLJobTypeOnlyWHQL:
				item.WHQLTimes = count
			case model.WHQLJobTypeHLKAndWHQL:
				item.HLKAndWHQLTimes = count
			default: // noop
			}
		}
	}

	// 合并数据。两个切片均按 BeginTime 有序，使用双指针归并，同一时间合并为一项。
	var items []*protocol.WindowsWebStatisticSigningTimesItem
	{
		log.Info(ctx, "merge data")
		items = make([]*protocol.WindowsWebStatisticSigningTimesItem, 0, len(windowsSigningJobItems)+len(whqlJobItems))
		i, j := 0, 0
		for i < len(windowsSigningJobItems) || j < len(whqlJobItems) {
			var item *protocol.WindowsWebStatisticSigningTimesItem
			switch {
			case i < len(windowsSigningJobItems) && j < len(whqlJobItems) && windowsSigningJobItems[i].BeginTime == whqlJobItems[j].BeginTime:
				windowsSigningItem := windowsSigningJobItems[i]
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningTimesItem{
					BeginTime:                    windowsSigningItem.BeginTime,
					PESigningTimes:               windowsSigningItem.PESigningTimes,
					AttestationSigningTimes:      windowsSigningItem.AttestationSigningTimes,
					PEAndAttestationSigningTimes: windowsSigningItem.PEAndAttestationSigningTimes,
					HLKAndWHQLTimes:              whqlItem.HLKAndWHQLTimes,
					WHQLTimes:                    whqlItem.WHQLTimes,
				}
				i++
				j++
			case j >= len(whqlJobItems) || (i < len(windowsSigningJobItems) && windowsSigningJobItems[i].BeginTime < whqlJobItems[j].BeginTime):
				windowsSigningItem := windowsSigningJobItems[i]
				item = &protocol.WindowsWebStatisticSigningTimesItem{
					BeginTime:                    windowsSigningItem.BeginTime,
					PESigningTimes:               windowsSigningItem.PESigningTimes,
					AttestationSigningTimes:      windowsSigningItem.AttestationSigningTimes,
					PEAndAttestationSigningTimes: windowsSigningItem.PEAndAttestationSigningTimes,
				}
				i++
			default:
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningTimesItem{
					BeginTime:       whqlItem.BeginTime,
					HLKAndWHQLTimes: whqlItem.HLKAndWHQLTimes,
					WHQLTimes:       whqlItem.WHQLTimes,
				}
				j++
			}
			items = append(items, item)
		}
	}

	rsp = &protocol.WindowsWebStatisticSigningTimesRsp{List: items}

	return
}

// WindowsWebStatisticSigningCost 获取应用的 Windows 类型签名耗时统计信息。
func WindowsWebStatisticSigningCost(ctx context.Context, req *protocol.WindowsWebStatisticSigningCostReq) (
	rsp *protocol.WindowsWebStatisticSigningCostRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询应用信息。
	var appID int
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app information")
			appQuery := conn.MySQLClient(ctx).App
			err = appQuery.WithContext(ctx).Select(
				appQuery.ID,
			).Where(
				appQuery.AppID.Eq(req.AppID),
			).Scan(&appID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "app not found")
					return nil, errs.New(consts.ErrParameterInvalid)
				}
				log.Error(ctx, "failed to retrieve app information from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 查询数据库，获取任务数量。
	var windowsSigningJobSqlResult []map[string]any
	{
		log.Info(ctx, "get cost of windows job")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		var tableNames []string
		tableNames, err = filterWindowsSigningJobTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) > 0 {
			windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
			switch req.TimeStep {
			case protocol.TimeStepDay:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CostWithDay(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepWeek:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CostWithWeek(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepMonth:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).CostWithMonth(tableNames, appID, req.BeginTime, req.EndTime)
			default:
				log.Warn(ctx, "unknown time step", req.TimeStep)
				return nil, errs.New(consts.ErrParameterInvalid)
			}
			if err != nil {
				log.Error(ctx, "failed to count windows job from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 转换数据。
	var windowsSigningJobItems []*protocol.WindowsWebStatisticSigningCostItem
	{
		log.Info(ctx, "deal sql data")
		windowsSigningJobItems = make([]*protocol.WindowsWebStatisticSigningCostItem, 0, len(windowsSigningJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningCostItem{}
		for _, v := range windowsSigningJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var cost int
			cost, err = strconv.Atoi(fmt.Sprintf("%v", v["cost"]))
			if err != nil {
				log.Error(ctx, "failed to convert cost to int", err, v["cost"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningCostItem{BeginTime: beginTime}
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			switch typ {
			case model.WindowsSigningJobTypePE:
				item.PESigningCost = cost
			case model.WindowsSigningJobTypeAttestation:
				item.AttestationSigningCost = cost
			case model.WindowsSigningJobTypePEAndAttestation:
				item.PEAndAttestationSigningCost = cost
			default: // noop
			}
		}
	}

	// 查询数据库，获取 WHQL 任务数量。
	var whqlJobSqlResult []map[string]any
	{
		log.Info(ctx, "get cost of whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		switch req.TimeStep {
		case protocol.TimeStepDay:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CostWithDay(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CostWithWeek(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).CostWithMonth(appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to count windows job from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 转换数据。
	var whqlJobItems []*protocol.WindowsWebStatisticSigningCostItem
	{
		log.Info(ctx, "deal sql data")
		whqlJobItems = make([]*protocol.WindowsWebStatisticSigningCostItem, 0, len(whqlJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningCostItem{}
		for _, v := range whqlJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var cost int
			cost, err = strconv.Atoi(fmt.Sprintf("%v", v["cost"]))
			if err != nil {
				log.Error(ctx, "failed to convert cost to int", err, v["cost"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				whqlJobItems = append(whqlJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningCostItem{BeginTime: beginTime}
				whqlJobItems = append(whqlJobItems, item)
			}
			switch typ {
			case model.WHQLJobTypeOnlyWHQL:
				item.WHQLCost = cost
			case model.WHQLJobTypeHLKAndWHQL:
				item.HLKAndWHQLCost = cost
			default: // noop
			}
		}
	}

	// 合并数据。两个切片均按 BeginTime 有序，使用双指针归并，同一时间合并为一项。
	var items []*protocol.WindowsWebStatisticSigningCostItem
	{
		log.Info(ctx, "merge data")
		items = make([]*protocol.WindowsWebStatisticSigningCostItem, 0, len(windowsSigningJobItems)+len(whqlJobItems))
		i, j := 0, 0
		for i < len(windowsSigningJobItems) || j < len(whqlJobItems) {
			var item *protocol.WindowsWebStatisticSigningCostItem
			switch {
			case i < len(windowsSigningJobItems) && j < len(whqlJobItems) && windowsSigningJobItems[i].BeginTime == whqlJobItems[j].BeginTime:
				windowsSigningItem := windowsSigningJobItems[i]
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningCostItem{
					BeginTime:                   windowsSigningItem.BeginTime,
					PESigningCost:               windowsSigningItem.PESigningCost,
					AttestationSigningCost:      windowsSigningItem.AttestationSigningCost,
					PEAndAttestationSigningCost: windowsSigningItem.PEAndAttestationSigningCost,
					HLKAndWHQLCost:              whqlItem.HLKAndWHQLCost,
					WHQLCost:                    whqlItem.WHQLCost,
				}
				i++
				j++
			case j >= len(whqlJobItems) || (i < len(windowsSigningJobItems) && windowsSigningJobItems[i].BeginTime < whqlJobItems[j].BeginTime):
				windowsSigningItem := windowsSigningJobItems[i]
				item = &protocol.WindowsWebStatisticSigningCostItem{
					BeginTime:                   windowsSigningItem.BeginTime,
					PESigningCost:               windowsSigningItem.PESigningCost,
					AttestationSigningCost:      windowsSigningItem.AttestationSigningCost,
					PEAndAttestationSigningCost: windowsSigningItem.PEAndAttestationSigningCost,
				}
				i++
			default:
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningCostItem{
					BeginTime:      whqlItem.BeginTime,
					HLKAndWHQLCost: whqlItem.HLKAndWHQLCost,
					WHQLCost:       whqlItem.WHQLCost,
				}
				j++
			}
			items = append(items, item)
		}
	}

	rsp = &protocol.WindowsWebStatisticSigningCostRsp{List: items}

	return
}

// WindowsWebStatisticSigningPassRate 获取应用的 Windows 类型签名通过率统计信息。
func WindowsWebStatisticSigningPassRate(ctx context.Context, req *protocol.WindowsWebStatisticSigningPassRateReq) (
	rsp *protocol.WindowsWebStatisticSigningPassRateRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询应用信息。
	var appID int
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app information")
			appQuery := conn.MySQLClient(ctx).App
			err = appQuery.WithContext(ctx).Select(
				appQuery.ID,
			).Where(
				appQuery.AppID.Eq(req.AppID),
			).Scan(&appID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "app not found")
					return nil, errs.New(consts.ErrParameterInvalid)
				}
				log.Error(ctx, "failed to retrieve app information from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 查询数据库，获取任务数量。
	var windowsSigningJobSqlResult []map[string]any
	{
		log.Info(ctx, "get pass rate of windows job")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		var tableNames []string
		tableNames, err = filterWindowsSigningJobTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) > 0 {
			windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
			switch req.TimeStep {
			case protocol.TimeStepDay:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).PassRateWithDay(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepWeek:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).PassRateWithWeek(tableNames, appID, req.BeginTime, req.EndTime)
			case protocol.TimeStepMonth:
				windowsSigningJobSqlResult, err = windowsSigningJobDo.WithContext(ctx).PassRateWithMonth(tableNames, appID, req.BeginTime, req.EndTime)
			default:
				log.Warn(ctx, "unknown time step", req.TimeStep)
				return nil, errs.New(consts.ErrParameterInvalid)
			}
			if err != nil {
				log.Error(ctx, "failed to query windows job from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 转换数据。
	var windowsSigningJobItems []*protocol.WindowsWebStatisticSigningPassRateItem
	{
		log.Info(ctx, "deal sql data")
		windowsSigningJobItems = make([]*protocol.WindowsWebStatisticSigningPassRateItem, 0, len(windowsSigningJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningPassRateItem{}
		for _, v := range windowsSigningJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var rate int
			rate, err = strconv.Atoi(fmt.Sprintf("%v", v["rate"]))
			if err != nil {
				log.Error(ctx, "failed to convert rate to int", err, v["rate"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningPassRateItem{BeginTime: beginTime}
				windowsSigningJobItems = append(windowsSigningJobItems, item)
			}
			switch typ {
			case model.WindowsSigningJobTypePE:
				item.PESigningPassRate = rate
			case model.WindowsSigningJobTypeAttestation:
				item.AttestationSigningPassRate = rate
			case model.WindowsSigningJobTypePEAndAttestation:
				item.PEAndAttestationSigningPassRate = rate
			default: // noop
			}
		}
	}

	// 查询数据库，获取 WHQL 任务数量。
	var whqlJobSqlResult []map[string]any
	{
		log.Info(ctx, "get pass rate of whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		switch req.TimeStep {
		case protocol.TimeStepDay:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).PassRateWithDay(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).PassRateWithWeek(appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			whqlJobSqlResult, err = whqlJobDo.WithContext(ctx).PassRateWithMonth(appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to query windows job from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 转换数据。
	var whqlJobItems []*protocol.WindowsWebStatisticSigningPassRateItem
	{
		log.Info(ctx, "deal sql data")
		whqlJobItems = make([]*protocol.WindowsWebStatisticSigningPassRateItem, 0, len(whqlJobSqlResult)/2)
		item := &protocol.WindowsWebStatisticSigningPassRateItem{}
		for _, v := range whqlJobSqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert type to int", err, v["type"])
			}
			var rate int
			rate, err = strconv.Atoi(fmt.Sprintf("%v", v["rate"]))
			if err != nil {
				log.Error(ctx, "failed to convert rate to int", err, v["rate"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				whqlJobItems = append(whqlJobItems, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.WindowsWebStatisticSigningPassRateItem{BeginTime: beginTime}
				whqlJobItems = append(whqlJobItems, item)
			}
			switch typ {
			case model.WHQLJobTypeOnlyWHQL:
				item.WHQLPassRate = rate
			case model.WHQLJobTypeHLKAndWHQL:
				item.HLKAndWHQLPassRate = rate
			default: // noop
			}
		}
	}

	// 合并数据。两个切片均按 BeginTime 有序，使用双指针归并，同一时间合并为一项。
	var items []*protocol.WindowsWebStatisticSigningPassRateItem
	{
		log.Info(ctx, "merge data")
		items = make([]*protocol.WindowsWebStatisticSigningPassRateItem, 0, len(windowsSigningJobItems)+len(whqlJobItems))
		i, j := 0, 0
		for i < len(windowsSigningJobItems) || j < len(whqlJobItems) {
			var item *protocol.WindowsWebStatisticSigningPassRateItem
			switch {
			case i < len(windowsSigningJobItems) && j < len(whqlJobItems) && windowsSigningJobItems[i].BeginTime == whqlJobItems[j].BeginTime:
				windowsSigningItem := windowsSigningJobItems[i]
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningPassRateItem{
					BeginTime:                       windowsSigningItem.BeginTime,
					PESigningPassRate:               windowsSigningItem.PESigningPassRate,
					AttestationSigningPassRate:      windowsSigningItem.AttestationSigningPassRate,
					PEAndAttestationSigningPassRate: windowsSigningItem.PEAndAttestationSigningPassRate,
					HLKAndWHQLPassRate:              whqlItem.HLKAndWHQLPassRate,
					WHQLPassRate:                    whqlItem.WHQLPassRate,
				}
				i++
				j++
			case j >= len(whqlJobItems) || (i < len(windowsSigningJobItems) && windowsSigningJobItems[i].BeginTime < whqlJobItems[j].BeginTime):
				windowsSigningItem := windowsSigningJobItems[i]
				item = &protocol.WindowsWebStatisticSigningPassRateItem{
					BeginTime:                       windowsSigningItem.BeginTime,
					PESigningPassRate:               windowsSigningItem.PESigningPassRate,
					AttestationSigningPassRate:      windowsSigningItem.AttestationSigningPassRate,
					PEAndAttestationSigningPassRate: windowsSigningItem.PEAndAttestationSigningPassRate,
				}
				i++
			default:
				whqlItem := whqlJobItems[j]
				item = &protocol.WindowsWebStatisticSigningPassRateItem{
					BeginTime:          whqlItem.BeginTime,
					HLKAndWHQLPassRate: whqlItem.HLKAndWHQLPassRate,
					WHQLPassRate:       whqlItem.WHQLPassRate,
				}
				j++
			}
			items = append(items, item)
		}
	}

	rsp = &protocol.WindowsWebStatisticSigningPassRateRsp{List: items}

	return
}

func createWindowsSignJob(ctx context.Context, jobInfo *model.WindowsSigningJob) error {
	windowsSigningJobDo := conn.MySQLTxClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(jobInfo.JobID))
	fields := make([]field.Expr, 0, 16)
	fields = append(fields,
		windowsSigningJobDo.JobID,
		windowsSigningJobDo.AppID,
		windowsSigningJobDo.UserID,
		windowsSigningJobDo.CertificateID,
		windowsSigningJobDo.FileID,
		windowsSigningJobDo.Source,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.CreatedTime,
		windowsSigningJobDo.Type,
	)
	if len(jobInfo.SignedFileID) > 0 {
		fields = append(fields, windowsSigningJobDo.SignedFileID)
	}
	if len(jobInfo.Log) > 0 {
		fields = append(fields, windowsSigningJobDo.Log)
	}
	if len(jobInfo.ProductID) > 0 {
		fields = append(fields, windowsSigningJobDo.ProductID)
	}
	if len(jobInfo.SubmissionID) > 0 {
		fields = append(fields, windowsSigningJobDo.SubmissionID)
	}
	if !jobInfo.FinishedTime.IsZero() {
		fields = append(fields, windowsSigningJobDo.FinishedTime)
	}
	if !jobInfo.FinishPeTime.IsZero() {
		fields = append(fields, windowsSigningJobDo.FinishPeTime)
	}
	if !jobInfo.UpdatedTime.IsZero() {
		fields = append(fields, windowsSigningJobDo.UpdatedTime)
	}
	err := windowsSigningJobDo.WithContext(ctx).Select(fields...).Create(jobInfo)
	if err != nil {
		log.Error(ctx, "failed to save windows sign job information to database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}

func getAllWindowsSignJobTableNames(ctx context.Context) ([]string, error) {
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	tables, err := windowsSigningJobDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to query database tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return tables, nil
}

func filterWindowsSigningJobTables(ctx context.Context, begin, end time.Time) ([]string, error) {
	// 从数据库中查处所有任务表。
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	allTables, err := windowsSigningJobDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to retrieve database tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	slices.Sort(allTables)
	if begin.IsZero() && end.IsZero() {
		return allTables, nil
	}

	// 过滤任务表。
	endTable := model.GetWindowsSigningJobTableName(end)
	beginTable := model.GetWindowsSigningJobTableName(begin)
	if !begin.IsZero() {
		for i := range allTables {
			if allTables[i] >= beginTable {
				allTables = allTables[i:]
				break
			}
		}
	}
	if !end.IsZero() {
		for i := len(allTables) - 1; i >= 0; i-- {
			if allTables[i] <= endTable {
				allTables = allTables[:i+1]
				break
			}
		}
	}

	log.Info(ctx, "gather windows tables", allTables, begin, end)
	return allTables, nil
}

func getCertificateCommonName(s string) string {
	arr := strings.Split(s, ",")
	if len(arr) == 1 {
		return strings.TrimSpace(arr[0])
	}
	for _, v := range arr {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "CN=") {
			return v[len("CN="):]
		}
	}
	return strings.TrimSpace(s)
}

func isFileInPEFormat(ctx context.Context, tusdID string) (bool, error) {
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "pe_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, tusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download pe file", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	defer util.RemoveFile(ctx, filePath)
	fileObj, err := pe.Open(filePath)
	if err == nil {
		log.ErrorIf(ctx, fileObj.Close(), "failed to close pe file")
		return true, nil
	}
	return false, nil
}

func isFileInZipFormat(ctx context.Context, tusdID string) (bool, error) {
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "zip_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, tusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download zip file", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	defer util.RemoveFile(ctx, filePath)
	fileObj, err := zip.OpenReader(filePath)
	if err == nil {
		log.ErrorIf(ctx, fileObj.Close(), "failed to close zip file")
		return true, nil
	}
	return false, nil
}

func isFileInCabFormat(ctx context.Context, tusdID string) (bool, error) {
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "cab_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, tusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download cab file", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	defer util.RemoveFile(ctx, filePath)
	ok, err := util.IsFileInCabFormat(ctx, filePath)
	if err != nil {
		log.Error(ctx, "check cab file failed", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	return ok, nil
}
