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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
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

// WindowsInternalGetWHQLJob 获取 WHQL 任务信息。
func WindowsInternalGetWHQLJob(ctx context.Context, req *protocol.WindowsInternalGetWHQLJobReq) (
	whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务信息。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "job not found")
				return
			}
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetWHQLJobToInitialTestMachine 给 HLK 测试虚拟机初始化任务。
func WindowsInternalGetWHQLJobToInitialTestMachine(ctx context.Context,
	req *protocol.WindowsInternalGetWHQLJobToInitialTestMachineReq) (whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusWaitingTest),
			whqlJobDo.TestSystem.Eq(req.System),
		).Order(whqlJobDo.ID.Asc()).Limit(1).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve whql job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if whqlJob == nil {
			// 没有任务可初始化。
			err = nil
			return
		}
	}

	// 更新任务状态。
	{
		log.Info(ctx, "update whql job status")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
			whqlJobDo.Status.Eq(model.WHQLJobStatusWaitingTest),
			whqlJobDo.TestSystem.Eq(req.System),
		).UpdateColumnSimple(
			whqlJobDo.Status.Value(model.WHQLJobStatusInitiallingTestMachine),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if sqlResult.RowsAffected <= 0 {
			// 没有获取到任务，被其它进程抢了。
			whqlJob = nil
			return
		}
		whqlJob.Status = model.WHQLJobStatusInitiallingTestMachine
	}

	return
}

// WindowsInternalUpdateWHQLJob 更新任务。
func WindowsInternalUpdateWHQLJob(ctx context.Context, req *protocol.WindowsInternalUpdateWHQLJobReq) (err error) {
	// 更新任务。
	{
		log.Info(ctx, "update whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, whqlJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(whqlJobDo.Log, logString+"\n"))
		}
		if len(req.TestMachineName) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.TestMachineName.Value(req.TestMachineName))
		}
		if len(req.ServiceName) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.ServiceName.Value(req.ServiceName))
		}
		if len(req.HLKLogFileID) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.HlkLogFileID.Value(req.HLKLogFileID))
		}
		finishedTestTime := time.Time(req.FinishedTestTime)
		if !finishedTestTime.IsZero() {
			assignExprs = append(assignExprs, whqlJobDo.FinishTestTime.Value(finishedTestTime))
		}
		if len(req.HLKXFileID) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.HlkxFileID.Value(req.HLKXFileID))
		}
		if len(assignExprs) <= 0 {
			log.Warn(ctx, "no fields need updating")
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusBadRequest)
			return
		}
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetWHQLJobToStartTest 获取任务，调度测试。
func WindowsInternalGetWHQLJobToStartTest(ctx context.Context, req *protocol.WindowsInternalGetWHQLJobToStartTestReq) (
	whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusFinishInitiallingTestMachine),
			whqlJobDo.TestSystem.In(req.Systems...),
			whqlJobDo.UpdatedTime.Lt(time.Now().Add(-cfg.Get().Backend().WaitingDelayTimeOfDispatchingTest())),
		).Order(whqlJobDo.UpdatedTime.Asc()).Limit(1).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve whql job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if whqlJob == nil {
			err = nil
			// 没有任务可初始化。
			return
		}
	}

	// 更新任务状态。
	{
		log.Info(ctx, "update whql job status")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
			whqlJobDo.Status.Eq(model.WHQLJobStatusFinishInitiallingTestMachine),
			whqlJobDo.TestSystem.In(req.Systems...),
		).UpdateColumnSimple(
			whqlJobDo.Status.Value(model.WHQLJobStatusDispatchHLKTest),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if sqlResult.RowsAffected <= 0 {
			// 没有获取到任务，被其它进程抢了。
			whqlJob = nil
			return
		}
		whqlJob.Status = model.WHQLJobStatusDispatchHLKTest
	}

	return
}

// WindowsInternalGetTestingWHQLJobs 获取正在测试中的任务。
func WindowsInternalGetTestingWHQLJobs(ctx context.Context, req *protocol.WindowsInternalGetTestingWHQLJobsReq) (
	whqlJobs []*model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql jobs")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJobs, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusHLKTesting),
			whqlJobDo.TestSystem.In(req.Systems...),
			whqlJobDo.UpdatedTime.Lt(time.Now()),
		).Order(whqlJobDo.UpdatedTime.Asc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve whql jobs from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetMachineEVCertificates 获取签名机器上在用的 EV UKey。
func WindowsInternalGetMachineEVCertificates(ctx context.Context,
	req *protocol.WindowsInternalGetMachineEVCertificatesReq) (sha1s []string, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get windows certificates")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		var windowsCertificates []*model.WindowsCertificate
		windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.Sha1.Distinct(),
		).Where(
			windowsCertificateDo.IP.Eq(req.IP),
			windowsCertificateDo.Type.In(model.WindowsCertificateTypePersonalEV, model.WindowsCertificateTypeCompanyEV),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows certificates from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		sha1s = util.ListTo(windowsCertificates,
			func(e *model.WindowsCertificate) string { return strings.ToUpper(e.Sha1) })
		sha1s = util.CleanStrings(sha1s)
	}

	return
}

// WindowsInternalGetWindowsSigningJob 获取签名任务信息。
func WindowsInternalGetWindowsSigningJob(ctx context.Context, req *protocol.WindowsInternalGetWindowsSigningJobReq) (
	windowsSigningJob *model.WindowsSigningJob, err error) {

	// 查库，获取任务信息。
	{
		log.Info(ctx, "get windows signing job")
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(req.JobID))
		windowsSigningJob, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve windows signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	return
}

// WindowsInternalGetCertificate 获取证书信息。
func WindowsInternalGetCertificate(ctx context.Context, req *protocol.WindowsInternalGetCertificateReq) (
	windowsCertificate *model.WindowsCertificate, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get windows certificate")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Where(
			windowsCertificateDo.ID.Eq(req.ID),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if windowsCertificate == nil {
			err = nil
			return
		}
	}

	// 解密证书。
	{
		if windowsCertificate.AesKeyID > 0 {
			var secret []byte
			secret, err = getAESSecret(ctx, windowsCertificate.AesKeyID)
			if err != nil {
				return
			}
			windowsCertificate.Content, err = util.AESCBCDecrypt(windowsCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt windows certificate", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	return
}

// WindowsInternalUpdateSigningJob 更新签名任务。
func WindowsInternalUpdateSigningJob(ctx context.Context, req *protocol.WindowsInternalUpdateSigningJobReq) (
	err error) {

	// 更新任务。
	{
		log.Info(ctx, "update windows signing job")
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(req.JobID))
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, windowsSigningJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(windowsSigningJobDo.Log, logString+"\n"))
		}
		if len(req.SignedFileID) > 0 {
			assignExprs = append(assignExprs, windowsSigningJobDo.SignedFileID.Value(req.SignedFileID))
		}
		finishedTime := time.Time(req.FinishedTime)
		if !finishedTime.IsZero() {
			assignExprs = append(assignExprs, windowsSigningJobDo.FinishedTime.Value(finishedTime))
		}
		finishedPETime := time.Time(req.FinishedPESignTime)
		if !finishedPETime.IsZero() {
			assignExprs = append(assignExprs, windowsSigningJobDo.FinishPeTime.Value(finishedPETime))
		}
		if len(assignExprs) <= 0 {
			log.Error(ctx, "no updated values")
			return
		}
		assignExprs = append(assignExprs, windowsSigningJobDo.UpdatedTime.Value(time.Now()))
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.JobID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update windows signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// CronSubmitCabFileSigning 提交 cab 文件签名。
func CronSubmitCabFileSigning(ctx context.Context, _ string, _ time.Time) {
	// 获取需要提交签名的签名任务。
	log.Debug(ctx, "get windows signing jobs that waiting cab signing")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusWaitCabSign)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no cab signing job need to start")
		return
	}

	// 获取签名证书。
	log.Info(ctx, "find hlkx signing certificate from database")
	windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
	var windowsCertificate *model.WindowsCertificate
	windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
		windowsCertificateDo.Sha1,
	).Where(
		windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
		windowsCertificateDo.DeletedTime.IsNull(),
		windowsCertificateDo.NotAfter.Gte(time.Now()),
	).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "no windows certificate available")
		}
		return
	}

	// 签名队列。
	queue := cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + strings.ToUpper(windowsCertificate.Sha1)

	// 发送 mq 消息。
	for _, v := range windowsSignJobIDs {
		err = publishMessageToQueue(ctx, queue, []byte(v))
		if err != nil {
			continue
		}

		// 更新任务状态。
		windowsSigningJobDo2 := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(v))
		_, err = windowsSigningJobDo2.WithContext(ctx).Where(
			windowsSigningJobDo2.JobID.Eq(v),
			windowsSigningJobDo2.Status.Eq(model.WindowsSigningJobStatusWaitCabSign),
		).UpdateColumnSimple(
			windowsSigningJobDo2.Status.Value(model.WindowsSigningJobStatusCabSigning),
		)
		if err != nil {
			log.Error(ctx, "failed to update windows signing job", v)
			continue
		}
	}
}

// CronStartAttestationJobs 提交 Windows 签名任务中的 Attestation 签名。
func CronStartAttestationJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要提交 Attestation 签名的 Windows 签名任务。
	log.Debug(ctx, "get all windows sign job table names")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	log.Debug(ctx, "get windows sign jobs that require attestation signing from database")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusAttestationWaiting)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no attestation job need to start")
		return
	}

	// 处理任务。
	for _, v := range windowsSignJobIDs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock windows sign job", v)
			lockKey := fmt.Sprintf(consts.RedisKeyAttestationStartLockFmt, v)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire attestation lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire attestation lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release attestation lock")
				if !success {
					log.Error(ctx, "releasing attestation lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start attestation signing job", v)
			startAttestationJob(childCtx, v)
		}()
	}
}

// CronCheckAttestationJobsResult 检查 Attestation 签名结果。
func CronCheckAttestationJobsResult(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 Attestation 签名结果的任务。
	log.Debug(ctx, "get all windows sign job table names")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	log.Debug(ctx, "get windows sign jobs that waiting attestation signing result from database")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusAttestationSigning)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no attestation job need to check")
		return
	}

	// 处理任务。
	for _, v := range windowsSignJobIDs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock windows sign job", v)
			lockKey := fmt.Sprintf(consts.RedisKeyAttestationCheckLockFmt, v)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire attestation lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire attestation lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release attestation lock")
				if !success {
					log.Error(ctx, "releasing attestation lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start attestation signing job", v)
			checkAttestationJobResult(childCtx, v)
		}()
	}
}

// CronSubmitHLKXFileSigningJobs 将 whql_job 任务，通过 hlk 测试的 hlkx 文件提交签名。
func CronSubmitHLKXFileSigningJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 WHQL 签名结果的任务。
	log.Debug(ctx, "get windows whql jobs that waiting hlk signing")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.UserID,
		whqlJobDo.AppID,
		whqlJobDo.HlkxFileID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusFinishTest),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to check")
		return
	}

	// 查找签名证书。
	log.Info(ctx, "find hlkx signing certificate from database")
	windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
	var windowsCertificate *model.WindowsCertificate
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
		}
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			innerCtx := ctxs.New()
			log.Info(innerCtx, "create hlkx signing job", v.JobID)

			// 生成任务 ID。
			var jobID string
			jobID, err = generateID(innerCtx, IDWindowsJob)
			if err != nil {
				return
			}
			defer func() {
				if err != nil {
					log.ErrorIf(innerCtx, reclaimID(innerCtx, IDWindowsJob, jobID),
						"failed to reclaim windows sign job id", jobID)
				}
			}()

			// 开启事务。
			conn.BeginMySQLTx(innerCtx)
			defer func() {
				if err == nil {
					log.ErrorIf(innerCtx, conn.CommitMySQLTx(innerCtx), "commit tx failed")
				} else {
					_ = conn.RollbackMySQLTx(innerCtx)
				}
			}()

			// 创建任务。
			windowsSigningJob := &model.WindowsSigningJob{
				AppID:         v.AppID,
				JobID:         jobID,
				Type:          model.WindowsSigningJobTypeHLKX,
				UserID:        v.UserID,
				CertificateID: windowsCertificate.ID,
				FileID:        v.HlkxFileID,
				Source:        model.SourceInternal,
				Status:        model.WindowsSigningJobStatusSigning,
				CreatedTime:   time.Now(),
			}
			if err = createWindowsSignJob(innerCtx, windowsSigningJob); err != nil {
				log.Error(ctx, "failed to create windows sign job to database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			// 更新 whql 任务。
			whqlJobDo = conn.MySQLTxClient(innerCtx).WhqlJob
			_, err = whqlJobDo.WithContext(innerCtx).Where(
				whqlJobDo.ID.Eq(v.ID),
			).UpdateColumnSimple(
				whqlJobDo.HlkxSignJobID.Value(jobID),
				whqlJobDo.Status.Value(model.WHQLJobStatusHLKXFileSinging),
			)
			if err != nil {
				log.Error(ctx, "failed to update whql job in database", err)
				return
			}

			// 发送任务消息。
			queue := cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
			if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalEV,
				model.WindowsCertificateTypeCompanyEV) {
				queue = cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + windowsCertificate.Sha1
			}
			err = publishMessageToQueue(ctx, queue, []byte(jobID))
			if err != nil {
				return
			}
		}()
	}
}

// CronStartWHQLJobs 提交 WHQL 签名。
func CronStartWHQLJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要提交 WHQL 签名的任务。
	log.Debug(ctx, "get whql job ids from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.JobID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusHLKXFileSinging),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to start")
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock whql job", v.JobID)
			lockKey := fmt.Sprintf(consts.RedisKeyWHQLStartLockFmt, v.JobID)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire whql lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire whql lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release whql lock")
				if !success {
					log.Error(ctx, "releasing whql lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start whql signing job", v.JobID, ctxs.RequestID(childCtx))
			startWHQLJob(childCtx, v.JobID)
		}()
	}
}

// CronCheckWHQLJobsResult 检查 WHQL 签名结果。
func CronCheckWHQLJobsResult(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 WHQL 签名结果的任务。
	log.Debug(ctx, "get windows whql jobs that waiting microsoft signing result from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.JobID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusWHQLSigning),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to check")
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock whql sign job", v.JobID)
			lockKey := fmt.Sprintf(consts.RedisKeyWHQLCheckLockFmt, v.JobID)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire whql lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire whql lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release whql lock")
				if !success {
					log.Error(ctx, "releasing whql lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start checking whql result", v, ctxs.RequestID(childCtx))
			checkWHQLJobResult(childCtx, v.JobID)
		}()
	}
}

func startAttestationJob(ctx context.Context, jobID string) {
	// 从数据中，获取任务信息。
	log.Info(ctx, "get windows sign job information from database", jobID)
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(jobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.ID,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.AppID,
		windowsSigningJobDo.JobID,
		windowsSigningJobDo.FileID,
		windowsSigningJobDo.Type,
		windowsSigningJobDo.SignedFileID,
		windowsSigningJobDo.CreatedTime,
	).Where(
		windowsSigningJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || windowsSigningJob == nil {
		log.Error(ctx, "windows sign job not found", err)
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify attestation job state")
	if windowsSigningJob.Status != model.WindowsSigningJobStatusAttestationWaiting {
		log.Warn(ctx, "windows sign job attestation status is not waiting")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	now := time.Now()
	if now.Sub(windowsSigningJob.CreatedTime) > cfg.Get().Backend().MaximumAttestationJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 查询数据库，获取文件信息。
	log.Info(ctx, "get file information from database")
	fileID := windowsSigningJob.SignedFileID
	fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(fileID))
	fileInfo, err := fileDo.WithContext(ctx).Select(
		fileDo.Name,
		fileDo.TusdID,
	).Where(
		fileDo.FileID.Eq(fileID),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "file not found", fileID)
			_, err = windowsSigningJobDo.WithContext(ctx).Where(
				windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
			).UpdateColumnSimple(
				windowsSigningJobDo.FinishedTime.Value(now),
				windowsSigningJobDo.UpdatedTime.Value(now),
				windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
				query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "待处理文件未找到")),
			)
			if err != nil {
				log.Error(ctx, "failed to update attestation job state", err)
			}
			return
		}
		return
	}

	// 下载文件到外存。
	log.Info(ctx, "download file to disk")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend,
		"attestation_signing_*"+filepath.Ext(fileInfo.Name))
	if err != nil {
		log.Error(ctx, "create a file path failed", err)
		return
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, fileInfo.TusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 从 cab 中提取 sys 文件。
	log.Info(ctx, "extract cab file", filePath)
	destinationDirectoryPath := filepath.Join(filepath.Dir(filePath),
		strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)))
	err = extractCab(filePath, destinationDirectoryPath)
	if err != nil {
		log.Error(ctx, "failed to extract cab file", err)
		return
	}
	defer util.RemoveDirectory(ctx, destinationDirectoryPath)

	// 校验 PE 文件格式。
	log.Info(ctx, "parse pe file machine type")
	peTyp := PETypeAmd64
	entries, err := os.ReadDir(destinationDirectoryPath)
	if err != nil {
		log.Error(ctx, "failed to read directory", err, destinationDirectoryPath)
		return
	}
	for _, v := range entries {
		if !strings.HasSuffix(strings.ToLower(v.Name()), cc.ExtensionSYS) {
			continue
		}
		func(filePath string) {
			var fileObj *pe.File
			fileObj, err = pe.Open(filePath)
			if err != nil {
				log.Error(ctx, "failed to open pe file", err)
				return
			}
			defer util.CloseIO(ctx, fileObj)
			switch fileObj.Machine {
			case pe.IMAGE_FILE_MACHINE_AMD64, pe.IMAGE_FILE_MACHINE_IA64:
				peTyp = PETypeAmd64
			case pe.IMAGE_FILE_MACHINE_ARM64, 0xA64E, 0xA641:
				peTyp = PETypeArm64
			case pe.IMAGE_FILE_MACHINE_I386, pe.IMAGE_FILE_MACHINE_LOONGARCH32, pe.IMAGE_FILE_MACHINE_RISCV32,
				0x160, 0x162, 0x184:
				peTyp = PETypeAmd32
			case pe.IMAGE_FILE_MACHINE_ARM:
				peTyp = PETypeArm32
			default:
				log.Warn(ctx, "unhandled pe machine", strconv.FormatUint(uint64(fileObj.Machine), 16))
			}
		}(filepath.Join(destinationDirectoryPath, v.Name()))
		if err != nil {
			log.Error(ctx, "parse sys file error", err)
			continue
		}
		break
	}
	log.Info(ctx, "machine code of sys file is", peTyp)

	// 创建微软产品。
	log.Info(ctx, "create microsoft product")
	productID, err := createAttestationProduct(ctx, strconv.Itoa(windowsSigningJob.AppID), peTyp)
	if err != nil {
		return
	}

	// 创建微软提交。
	log.Info(ctx, "create microsoft product's submission")
	submissionID, uploadURL, err := createMicrosoftSubmission(ctx, productID, windowsSigningJob.JobID)
	if err != nil {
		return
	}

	// 上传文件。
	log.Info(ctx, "upload file to microsoft server")
	err = uploadFileToMicrosoft(ctx, filePath, uploadURL)
	if err != nil {
		return
	}

	// 提交微软提交。
	log.Info(ctx, "commit a microsoft submission")
	err = commitMicrosoftSubmission(ctx, productID, submissionID)
	if err != nil {
		return
	}

	// 更新数据库中的任务进展信息。
	log.Info(ctx, "update windows sign job information in database")
	now = time.Now()
	_, err = windowsSigningJobDo.WithContext(ctx).Where(
		windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
	).UpdateColumnSimple(
		windowsSigningJobDo.UpdatedTime.Value(now),
		windowsSigningJobDo.ProductID.Value(productID),
		windowsSigningJobDo.SubmissionID.Value(submissionID),
		windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusAttestationSigning),
		query.Concat(windowsSigningJobDo.Log,
			formatJobLog(log.LevelInfo, "提交至微软方签名，PID：%s，SID：%s", productID, submissionID)),
	)
	if err != nil {
		log.Error(ctx, "failed to update windows sign job in database", err)
		return
	}
}

func checkAttestationJobResult(ctx context.Context, jobID string) {
	// 查询数据库，获取任务信息。
	log.Info(ctx, "get windows sign job information from database", jobID)
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(jobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.ID,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.UpdatedTime,
		windowsSigningJobDo.CreatedTime,
		windowsSigningJobDo.ProductID,
		windowsSigningJobDo.SubmissionID,
		windowsSigningJobDo.AppID,
	).Where(
		windowsSigningJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || windowsSigningJob == nil {
		log.Error(ctx, "windows sign job not found")
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify attestation job state")
	if windowsSigningJob.Status != model.WindowsSigningJobStatusAttestationSigning {
		log.Warn(ctx, "windows sign job attestation status is not signing")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	beginTime := windowsSigningJob.UpdatedTime
	if beginTime.IsZero() {
		beginTime = windowsSigningJob.CreatedTime
	}
	now := time.Now()
	if now.Sub(beginTime) > cfg.Get().Backend().MaximumAttestationJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 查询微软方签名结果。
	log.Info(ctx, "query submission state from microsoft")
	finished, downloadURL, err := queryMicrosoftSubmission(
		ctx, windowsSigningJob.ProductID, windowsSigningJob.SubmissionID)
	if !finished {
		log.Warn(ctx, "windows sign job attestation status is not finished")
		return
	}

	// 处理签名结果。
	log.Info(ctx, "deal with signing result")
	now = time.Now()
	if err != nil {
		// 更新数据库任务状态。
		log.Warn(ctx, "microsoft attestation job failed", err)
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log,
				formatJobLog(log.LevelError, "微软 Attestation 签名失败：%v", errs.String(err))),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 下载签名结果文件。
	log.Info(ctx, "download signing result file")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "attestation_result_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return
	}
	httpCode, fileName, fileSize, fileMD5, err := util.HTTPGetToDisk(ctx, downloadURL, filePath)
	if err != nil || !(httpCode >= http.StatusOK && httpCode < http.StatusMultipleChoices) {
		log.Error(ctx, "failed to download signing result file", err, httpCode)
		return
	}

	// 将文件上传到 Tusd。
	log.Info(ctx, "upload signing result file to tusd")
	tusdID, err := conn.TusdClient(ctx).MultipleUploadFromFile(ctx, filePath)
	if err != nil {
		log.Error(ctx, "failed to upload signing result file", err)
		return
	}
	defer func() {
		// 若失败，删除文件。
		if err != nil {
			log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, tusdID), "failed to delete tusd file", tusdID)
		}
	}()

	// 保存文件信息到数据库中。
	fileID, err := generateID(ctx, IDFile)
	if err != nil {
		log.Error(ctx, "failed to generate file id", err)
		return
	}
	defer func() {
		// 失败回收 ID。
		if err != nil {
			log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id", fileID)
		}
	}()
	now = time.Now()
	err = createFile(ctx, &model.File{
		FileID:      fileID,
		TusdID:      tusdID,
		AppID:       windowsSigningJob.AppID,
		Name:        fileName,
		Md5:         fileMD5,
		Size:        int(fileSize),
		Type:        model.FileTypeMicrosoftSigning,
		CreatedTime: now,
	})
	if err != nil {
		return
	}

	// 更新数据库中任务状态。
	log.Info(ctx, "update windows sign job information in database")
	_, err = windowsSigningJobDo.WithContext(ctx).Where(
		windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
	).UpdateColumnSimple(
		windowsSigningJobDo.UpdatedTime.Value(now),
		windowsSigningJobDo.FinishedTime.Value(now),
		windowsSigningJobDo.SignedFileID.Value(fileID),
		windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusSuccess),
		query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelInfo, "微软方签名成功")),
	)
	if err != nil {
		log.Error(ctx, "failed to update windows sign job in database", err)
		return
	}
}

func startWHQLJob(ctx context.Context, jobID string) {
	// 从数据中，获取任务信息。
	log.Info(ctx, "get whql job information from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJob, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.Status,
		whqlJobDo.CreatedTime,
		whqlJobDo.HlkxSignJobID,
	).Where(
		whqlJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil {
		log.Error(ctx, "whql job not found")
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify whql job state")
	if whqlJob.Status != model.WHQLJobStatusHLKXFileSinging {
		log.Warn(ctx, "whql status is not in waiting")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	if time.Since(whqlJob.CreatedTime) > cfg.Get().Backend().MaximumWHQLJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 查询 HLKX 文件签名结果。
	log.Info(ctx, "get hlkx signing job")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.
		Table(model.GetWindowsSigningJobByID(whqlJob.HlkxSignJobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.Log,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.JobID,
		windowsSigningJobDo.SignedFileID,
		windowsSigningJobDo.AppID,
	).Where(
		windowsSigningJobDo.JobID.Eq(whqlJob.HlkxSignJobID),
	).Take()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名信息未找到")),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 判断文件签名结果。
	log.Info(ctx, "check hlkx signing job result")
	switch windowsSigningJob.Status {
	case model.WindowsSigningJobStatusSuccess:
		log.Info(ctx, "hlkx signing job pass")
	case model.WindowsSigningJobStatusFailure:
		log.Error(ctx, "hlkx file signing failed")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名失败：%s", windowsSigningJob.Log)),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	default:
		log.Warn(ctx, "hlkx signing job running", jobID)
		return
	}

	// 查询数据库，获取文件信息。
	log.Info(ctx, "get file information from database")
	fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(windowsSigningJob.SignedFileID))
	file, err := fileDo.WithContext(ctx).Select(
		fileDo.Name,
		fileDo.TusdID,
	).Where(
		fileDo.FileID.Eq(windowsSigningJob.SignedFileID),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "file not found", windowsSigningJob.SignedFileID)
			_, err = whqlJobDo.WithContext(ctx).Where(
				whqlJobDo.ID.Eq(whqlJob.ID),
			).UpdateColumnSimple(
				whqlJobDo.FinishedTime.Value(time.Now()),
				whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
				query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名结果文件未找到")),
			)
			if err != nil {
				log.Error(ctx, "failed to update whql job state", err)
			}
			return
		}
		return
	}

	// 下载文件到外存。
	log.Info(ctx, "download file to disk")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "whql_signing_*."+filepath.Ext(file.Name))
	if err != nil {
		log.Error(ctx, "create a file path failed", err)
		return
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, file.TusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 创建微软产品。
	log.Info(ctx, "create microsoft product")
	productID, err := createWHQLProduct(ctx, strconv.Itoa(windowsSigningJob.AppID))
	if err != nil {
		return
	}

	// 创建微软提交。
	log.Info(ctx, "create microsoft product's submission")
	submissionID, uploadURL, err := createMicrosoftSubmission(ctx, productID, windowsSigningJob.JobID)
	if err != nil {
		return
	}

	// 上传文件。
	log.Info(ctx, "upload file to microsoft server")
	err = uploadFileToMicrosoft(ctx, filePath, uploadURL)
	if err != nil {
		return
	}

	// 提交微软提交。
	log.Info(ctx, "commit a microsoft submission")
	err = commitMicrosoftSubmission(ctx, productID, submissionID)
	if err != nil {
		return
	}

	// 更新数据库中的任务进展信息。
	log.Info(ctx, "update whql job information in database")
	_, err = whqlJobDo.WithContext(ctx).Where(
		whqlJobDo.ID.Eq(whqlJob.ID),
	).UpdateColumnSimple(
		whqlJobDo.ProductID.Value(productID),
		whqlJobDo.SubmissionID.Value(submissionID),
		whqlJobDo.Status.Value(model.WHQLJobStatusWHQLSigning),
		query.Concat(whqlJobDo.Log,
			formatJobLog(log.LevelInfo, "提交至微软方签名，PID：%s，SID：%s", productID, submissionID)),
	)
	if err != nil {
		log.Error(ctx, "failed to update whql job in database", err)
		return
	}
}

func checkWHQLJobResult(ctx context.Context, jobID string) {
	// 查询数据库，获取任务信息。
	log.Info(ctx, "get whql job information from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJob, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.Status,
		whqlJobDo.CreatedTime,
		whqlJobDo.ProductID,
		whqlJobDo.SubmissionID,
		whqlJobDo.AppID,
	).Where(
		whqlJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || whqlJob == nil {
		log.Error(ctx, "failed to retrieve whql job information from database", err)
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify whql job state")
	if whqlJob.Status != model.WHQLJobStatusWHQLSigning {
		log.Warn(ctx, "whql job status is not in microsoft signing")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify whql job time consuming")
	if time.Since(whqlJob.CreatedTime) > cfg.Get().Backend().MaximumWHQLJobInterval() {
		log.Warn(ctx, "whql job has exceeded deadline")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError,
				"任务超时终止，最大时限为：%s", util.FormatDuration(cfg.Get().Backend().MaximumWHQLJobInterval()))),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 请求微软服务器查询结果。
	log.Info(ctx, "request microsoft server to get result")
	finished, downloadURL, err := queryMicrosoftSubmission(ctx, whqlJob.ProductID, whqlJob.SubmissionID)
	if !finished {
		log.Info(ctx, "whql job is not over yet")
		return
	}
	if err != nil {
		log.Warn(ctx, "whql job failed", err)
		// 更新数据库任务信息。
		log.Info(ctx, "update whql job in database")
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "微软审核不通过：%v", errs.String(err))),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job information in database", err)
			return
		}
		if sqlResult.RowsAffected != 1 {
			log.Error(ctx, "updating whql job information has no effect in database", sqlResult.RowsAffected)
		}
		return
	}

	// 下载微软签名结果文件。
	log.Info(ctx, "download microsoft signed file")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "whql_signed_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return
	}
	httpCode, fileName, fileSize, fileMD5, err := util.HTTPGetToDisk(ctx, downloadURL, filePath)
	if err != nil {
		log.Error(ctx, "failed to download microsoft audit result file", err, downloadURL)
		return
	}
	if httpCode != http.StatusOK {
		log.Error(ctx, "failed to download microsoft audit result file", err, downloadURL, httpCode)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 将文件上传到 Tusd。
	log.Info(ctx, "upload file to tusd", fileName, fileSize, fileMD5)
	tusdID, err := conn.TusdClient(ctx).MultipleUploadFromFile(ctx, filePath)
	if err != nil {
		log.Error(ctx, "failed to upload microsoft audit result file to tusd", err)
		return
	}
	defer func() {
		if err != nil {
			_, err2 := conn.TusdClient(ctx).Delete(ctx, &tus.DeleteRequest{Location: tusdID})
			log.ErrorIf(ctx, err2, "failed to delete tusd file", tusdID)
		}
	}()

	// 保存文件信息到数据库。
	log.Info(ctx, "save file information to database")
	fileID, err := generateID(ctx, IDFile)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id")
		}
	}()
	ctxTx := conn.BeginMySQLTx(ctx)
	defer func() {
		if err == nil {
			log.ErrorIf(ctxTx, conn.CommitMySQLTx(ctxTx), "failed to commit transaction")
		} else {
			log.ErrorIf(ctx, conn.RollbackMySQLTx(ctxTx), "rollback transaction failed")
		}
	}()
	err = createFile(ctxTx, &model.File{
		FileID:      fileID,
		TusdID:      tusdID,
		AppID:       whqlJob.AppID,
		Name:        fileName,
		Md5:         fileMD5,
		Size:        int(fileSize),
		Type:        model.FileTypeMicrosoftSigning,
		CreatedTime: time.Now(),
	})
	if err != nil {
		return
	}

	// 更新数据库中的任务信息。
	log.Info(ctx, "update whql job in database")
	var sqlResult gen.ResultInfo
	whqlJobTxDo := conn.MySQLTxClient(ctxTx).WhqlJob
	sqlResult, err = whqlJobTxDo.WithContext(ctx).Where(
		whqlJobDo.ID.Eq(whqlJob.ID),
	).UpdateColumnSimple(
		whqlJobDo.FinishedTime.Value(time.Now()),
		whqlJobDo.SignedFileID.Value(fileID),
		whqlJobDo.Status.Value(model.WHQLJobStatusSuccess),
		query.Concat(whqlJobDo.Log, formatJobLog(log.LevelInfo, "微软审核通过")),
	)
	if err != nil {
		log.Error(ctx, "failed to update whql job information in database", err)
		return
	}
	if sqlResult.RowsAffected <= 0 {
		log.Error(ctx, "updating whql job information result is not as expected in database", sqlResult.RowsAffected)
	}
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

// 解析 CAB 文件。使用 expand 和 cabextract 命令提取。
func extractCab(cabFilePath string, destinationDirectoryPath string) (err error) {
	err = os.MkdirAll(destinationDirectoryPath, cc.DirectoryMode)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		_, err = exec.Command("expand", cabFilePath, "-F:*", destinationDirectoryPath).CombinedOutput()
	} else {
		_, err = exec.Command(consts.CabextractFilePath, "-d", destinationDirectoryPath, cabFilePath).CombinedOutput()
	}
	return
}
