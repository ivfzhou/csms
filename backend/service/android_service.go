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
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen"
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
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

type keystoreInfo struct {
	StoreType                 string
	StoreProvider             string
	AliasName                 string
	Owner                     string
	Issuer                    string
	SerialNumber              string
	SHA1                      string
	SHA256                    string
	SignatureAlgorithmName    string
	SubjectPublicKeyAlgorithm string
	Version                   string
	CreationDate              time.Time
	ValidFrom                 time.Time
	ValidUntil                time.Time
}

// AndroidWebAddOrganization 添加证书主体。
func AndroidWebAddOrganization(ctx context.Context, req *protocol.AndroidWebAddOrganizationReq) (err error) {
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

	// 将安卓主体信息添加到数据库。
	{
		log.Info(ctx, "add android organization")
		androidOrganizationDo := conn.MySQLClient(ctx).AndroidOrganization
		if err = androidOrganizationDo.WithContext(ctx).Create(&model.AndroidOrganization{
			Name:        req.CommonName,
			UserID:      user.ID,
			Owner:       req.DName,
			CreatedTime: time.Now(),
		}); err != nil {
			log.Error(ctx, "failed to add android organization to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// AndroidWebListOrganizations 获取安卓证书主体信息列表。
func AndroidWebListOrganizations(ctx context.Context) (rsp *protocol.AndroidWebListOrganizationsRsp, err error) {
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

	// 获取安卓证书主体信息。
	var androidOrganizations []*model.AndroidOrganization
	{
		log.Info(ctx, "get android organizations")
		androidOrganizationDo := conn.MySQLClient(ctx).AndroidOrganization
		androidOrganizations, err = androidOrganizationDo.WithContext(ctx).Order(
			androidOrganizationDo.CreatedTime.Desc(),
			androidOrganizationDo.ID.Desc(),
		).Order(androidOrganizationDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve android organization from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(androidOrganizations) <= 0 {
			return
		}
	}

	// 获取用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDs := util.ListTo(androidOrganizations, func(e *model.AndroidOrganization) int { return e.UserID })
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AndroidWebListOrganizationsItem, len(androidOrganizations))
		for i, v := range androidOrganizations {
			list[i] = &protocol.AndroidWebListOrganizationsItem{
				ID:          v.ID,
				CommonName:  v.Name,
				DName:       v.Owner,
				User:        userIDToName[v.UserID],
				CreatedTime: formatTime(&v.CreatedTime),
			}
		}
		rsp = &protocol.AndroidWebListOrganizationsRsp{List: list}
	}

	return
}

// AndroidWebApplyCertificate 申请安卓证书。
func AndroidWebApplyCertificate(ctx context.Context, req *protocol.AndroidWebApplyCertificateReq) (err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中获取证书主体信息。
	var owner string
	{
		log.Info(ctx, "get android organization information")
		androidOrganizationDo := conn.MySQLClient(ctx).AndroidOrganization
		err = androidOrganizationDo.WithContext(ctx).Select(
			androidOrganizationDo.Owner,
		).Where(
			androidOrganizationDo.ID.Eq(req.OwnerID),
		).Scan(&owner)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android organization not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to get android organization information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 运行 keytool，生成密钥库。
	var jksBytes []byte
	var jksFilePath string
	var keypass string
	var storepass string
	{
		log.Info(ctx, "run the keytool ro generate android keystore")
		jksFilePath, err = util.GenerateTemporaryFile(cc.ServiceNameBackend, "*.jks")
		if err != nil {
			log.Error(ctx, "failed to get a file path", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		keypass = util.RandomPrintableASCIINoSpaceString(consts.KeystorePasswordLength)
		storepass = util.RandomPrintableASCIINoSpaceString(consts.KeystorePasswordLength)
		var outputBytes []byte
		outputBytes, err = exec.Command(
			consts.KeytoolBinaryPath,
			"-J-Duser.language=en",
			"-J-Duser.country=US",
			"-genkeypair",
			"-alias", req.Alias,
			"-keypass", keypass,
			"-storepass", storepass,
			"-keyalg", "RSA",
			"-keysize", "3072",
			"-validity", "40000",
			"-keystore", jksFilePath,
			"-dname", owner,
			"-storetype", "jks",
		).CombinedOutput()
		log.Debug(ctx, "output of running the keytool", outputBytes)
		if err != nil {
			log.Error(ctx, "failed to run the keytool command", err, outputBytes)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, jksFilePath)
		jksBytes, err = os.ReadFile(jksFilePath)
		if err != nil {
			log.Error(ctx, "failed to read keystore file data", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取证书信息。
	var jksInfo *keystoreInfo
	{
		log.Info(ctx, "parse keystore file")
		jksInfo, err = parseKeystore(ctx, jksFilePath, storepass)
		if err != nil {
			log.Error(ctx, "failed to parse keystore file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 加密证书。
	var aesKeyID int
	{
		log.Info(ctx, "encrypt android certificate")
		var aesKeyInfo *model.AesKey
		aesKeyInfo, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		jksBytes, err = util.AESCBCEncrypt(aesKeyInfo.Secret, jksBytes)
		if err != nil {
			log.Error(ctx, "failed to encrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		aesKeyID = aesKeyInfo.ID
	}

	// 将证书信息保存到数据库。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "save android certificate")
		var certificateID string
		certificateID, err = generateID(ctx, IDAndroidCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAndroidCertificate, certificateID),
					"failed to reclaim android certificate id", certificateID)
			}
		}()
		now := time.Now()
		androidCertificate = &model.AndroidCertificate{
			CertificateID:      certificateID,
			AppID:              app.ID,
			UserID:             user.ID,
			Alias_:             req.Alias,
			Category:           req.Type,
			Publisher:          jksInfo.Issuer,
			Owner:              jksInfo.Owner,
			SignatureAlgorithm: jksInfo.SignatureAlgorithmName,
			PublicKeyAlgorithm: jksInfo.SubjectPublicKeyAlgorithm,
			Version:            jksInfo.Version,
			SerialNumber:       jksInfo.SerialNumber,
			Sha1:               jksInfo.SHA1,
			CreationDate:       jksInfo.CreationDate,
			StoreType:          jksInfo.StoreType,
			StoreProvider:      jksInfo.StoreProvider,
			Sha256:             jksInfo.SHA256,
			AesKeyID:           aesKeyID,
			Storepass:          storepass,
			Keypass:            keypass,
			NotBefore:          jksInfo.ValidFrom,
			NotAfter:           jksInfo.ValidUntil,
			Content:            jksBytes,
			CreatedTime:        now,
		}
		androidCertificateTxDo := conn.MySQLTxClient(ctx).AndroidCertificate
		err = androidCertificateTxDo.WithContext(ctx).Select(
			androidCertificateTxDo.CertificateID,
			androidCertificateTxDo.CreationDate,
			androidCertificateTxDo.AppID,
			androidCertificateTxDo.UserID,
			androidCertificateTxDo.Alias_,
			androidCertificateTxDo.Category,
			androidCertificateTxDo.Publisher,
			androidCertificateTxDo.Owner,
			androidCertificateTxDo.SignatureAlgorithm,
			androidCertificateTxDo.PublicKeyAlgorithm,
			androidCertificateTxDo.Version,
			androidCertificateTxDo.SerialNumber,
			androidCertificateTxDo.Sha1,
			androidCertificateTxDo.CreationDate,
			androidCertificateTxDo.StoreType,
			androidCertificateTxDo.StoreProvider,
			androidCertificateTxDo.Sha256,
			androidCertificateTxDo.AesKeyID,
			androidCertificateTxDo.Keypass,
			androidCertificateTxDo.Storepass,
			androidCertificateTxDo.Keypass,
			androidCertificateTxDo.NotBefore,
			androidCertificateTxDo.NotAfter,
			androidCertificateTxDo.Content,
			androidCertificateTxDo.CreatedTime,
		).Create(androidCertificate)
		if err != nil {
			log.Error(ctx, "failed to create android certificate to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件信息保存到数据库。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyAndroidCertificate,
			CreatedTime: androidCertificate.CreatedTime,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:  user.NameEn,
			EventApp:   app.Name,
			EventAlias: req.Alias,
			EventDetail: util.GetPrintJSON(map[string]any{
				"certificateId": androidCertificate.CertificateID,
				"sha1":          androidCertificate.Sha1,
				"owner":         androidCertificate.Owner,
				"publisher":     androidCertificate.Publisher,
				"notAfter":      formatTime(&androidCertificate.NotAfter),
				"notBefore":     formatTime(&androidCertificate.NotBefore),
				"type":          model.AllAndroidCertificateTypeDescriptions[req.Type],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AndroidWebUploadCertificate 上传安卓证书。
func AndroidWebUploadCertificate(ctx context.Context, req *protocol.AndroidWebUploadCertificateReq) (err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 解析证书。
	var jksData []byte
	var jksInfo *keystoreInfo
	var jksPath string
	{
		log.Info(ctx, "parse android certificate")
		jksData, err = base64.StdEncoding.DecodeString(req.Certificate)
		if err != nil {
			log.Error(ctx, "failed to base64 decode certificate", err)
			err = errs.NewWithError(consts.ErrParameterInvalid, err)
			return
		}
		jksPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks", jksData)
		if err != nil {
			log.Error(ctx, "failed to make a temporary jks file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, jksPath)
		jksInfo, err = parseKeystore(ctx, jksPath, req.Storepass)
		if err != nil {
			return
		}
	}

	// 校验 keypass 是否正确。
	{
		log.Info(ctx, "verify keypass of keystore")
		var outputBytes []byte
		outputBytes, err = exec.Command(
			"keytool",
			"-J-Duser.language=en",
			"-J-Duser.country=US",
			"-keypasswd",
			"-keystore", jksPath,
			"-storepass", req.Storepass,
			"-alias", jksInfo.AliasName,
			"-keypass", req.Keypass,
			"-new", req.Keypass,
		).CombinedOutput()
		log.Debug(ctx, "output of running keytool", outputBytes)
		if err != nil {
			log.Warn(ctx, "failed to run keytool", err)
			err = errs.New(consts.ErrKeystoreKeyPassInvalid)
			return
		}
	}

	// 加密证书。
	var aesKeyID int
	{
		log.Info(ctx, "encrypt android certificate")
		var aesKeyInfo *model.AesKey
		aesKeyInfo, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		if jksData, err = util.AESCBCEncrypt(aesKeyInfo.Secret, jksData); err != nil {
			log.Error(ctx, "failed to encrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		aesKeyID = aesKeyInfo.ID
	}

	// 将证书信息保存到数据库。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "save android certificate")
		var certificateID string
		certificateID, err = generateID(ctx, IDAndroidCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAndroidCertificate, certificateID),
					"failed to reclaim android certificate id", certificateID)
			}
		}()
		now := time.Now()
		androidCertificate = &model.AndroidCertificate{
			CertificateID:      certificateID,
			AppID:              app.ID,
			UserID:             user.ID,
			Alias_:             jksInfo.AliasName,
			Category:           req.Type,
			Publisher:          jksInfo.Issuer,
			Owner:              jksInfo.Owner,
			SignatureAlgorithm: jksInfo.SignatureAlgorithmName,
			PublicKeyAlgorithm: jksInfo.SubjectPublicKeyAlgorithm,
			Version:            jksInfo.Version,
			SerialNumber:       jksInfo.SerialNumber,
			Sha1:               jksInfo.SHA1,
			CreationDate:       jksInfo.CreationDate,
			StoreType:          jksInfo.StoreType,
			StoreProvider:      jksInfo.StoreProvider,
			Sha256:             jksInfo.SHA256,
			AesKeyID:           aesKeyID,
			Storepass:          req.Storepass,
			Keypass:            req.Keypass,
			NotBefore:          jksInfo.ValidFrom,
			NotAfter:           jksInfo.ValidUntil,
			Content:            jksData,
			CreatedTime:        now,
		}
		androidCertificateTxDo := conn.MySQLTxClient(ctx).AndroidCertificate
		err = androidCertificateTxDo.WithContext(ctx).Select(
			androidCertificateTxDo.CertificateID,
			androidCertificateTxDo.CreationDate,
			androidCertificateTxDo.AppID,
			androidCertificateTxDo.UserID,
			androidCertificateTxDo.Alias_,
			androidCertificateTxDo.Category,
			androidCertificateTxDo.Publisher,
			androidCertificateTxDo.Owner,
			androidCertificateTxDo.SignatureAlgorithm,
			androidCertificateTxDo.PublicKeyAlgorithm,
			androidCertificateTxDo.Version,
			androidCertificateTxDo.SerialNumber,
			androidCertificateTxDo.Sha1,
			androidCertificateTxDo.CreationDate,
			androidCertificateTxDo.StoreType,
			androidCertificateTxDo.StoreProvider,
			androidCertificateTxDo.Sha256,
			androidCertificateTxDo.AesKeyID,
			androidCertificateTxDo.Keypass,
			androidCertificateTxDo.Storepass,
			androidCertificateTxDo.Keypass,
			androidCertificateTxDo.NotBefore,
			androidCertificateTxDo.NotAfter,
			androidCertificateTxDo.Content,
			androidCertificateTxDo.CreatedTime,
		).Create(androidCertificate)
		if err != nil {
			log.Error(ctx, "failed to save android certificate to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录事件。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeUploadAndroidCertificate,
			CreatedTime: androidCertificate.CreatedTime,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:  user.NameEn,
			EventApp:   app.Name,
			EventAlias: jksInfo.AliasName,
			EventDetail: util.GetPrintJSON(map[string]any{
				"certificateId": androidCertificate.CertificateID,
				"sha1":          androidCertificate.Sha1,
				"owner":         androidCertificate.Owner,
				"publisher":     androidCertificate.Publisher,
				"notBefore":     formatTime(&androidCertificate.NotBefore),
				"notAfter":      formatTime(&androidCertificate.NotAfter),
				"type":          model.AllAndroidCertificateTypeDescriptions[req.Type],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AndroidWebListCertificates 获取安卓证书列表。
func AndroidWebListCertificates(ctx context.Context) (rsp *protocol.AndroidWebListCertificatesRsp, err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中获取证书信息。
	var androidCertificates []*model.AndroidCertificate
	{
		log.Info(ctx, "get android certificates from database")
		androidCertificateQuery := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificates, err = androidCertificateQuery.WithContext(ctx).Select(
			androidCertificateQuery.UserID,
			androidCertificateQuery.CertificateID,
			androidCertificateQuery.Alias_,
			androidCertificateQuery.Owner,
			androidCertificateQuery.SignatureAlgorithm,
			androidCertificateQuery.PublicKeyAlgorithm,
			androidCertificateQuery.Sha1,
			androidCertificateQuery.Sha256,
			androidCertificateQuery.Category,
			androidCertificateQuery.CreatedTime,
			androidCertificateQuery.NotBefore,
			androidCertificateQuery.NotAfter,
		).Where(
			androidCertificateQuery.AppID.Eq(app.ID),
			androidCertificateQuery.DeletedTime.IsNull(),
		).Order(androidCertificateQuery.ID.Desc()).Find()
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
		list := make([]*protocol.AndroidWebListCertificatesItem, len(androidCertificates))
		for i, v := range androidCertificates {
			list[i] = &protocol.AndroidWebListCertificatesItem{
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
		rsp = &protocol.AndroidWebListCertificatesRsp{List: list}
	}

	return
}

// AndroidWebDownloadCertificate 下载安卓证书。
func AndroidWebDownloadCertificate(ctx context.Context, req *protocol.AndroidWebDownloadCertificateReq) (
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

	// 校验应用平台。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 从数据库中获取证书信息。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate information from database")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.AppID,
			androidCertificateDo.Content,
			androidCertificateDo.Category,
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Alias_,
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
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var certificateData []byte
	{
		log.Info(ctx, "decrypt android certificate")
		var secret []byte
		secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
		if err != nil {
			return
		}
		certificateData, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
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
			UserID:      user.ID,
			Type:        model.EventTypeDownloadAndroidCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:  user.NameEn,
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
		Size:   int64(len(certificateData)),
		Reader: io.NopCloser(bytes.NewReader(certificateData)),
	}

	return
}

// AndroidWebGetGooglePlayCertificate 获取谷歌 Play 上传证书。
func AndroidWebGetGooglePlayCertificate(ctx context.Context, req *protocol.AndroidWebGetGooglePlayCertificateReq) (
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
		if app.Platform != model.AppPlatformAndroid {
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库获取证书。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate information from database")
		androidCertificateDo := conn.MySQLTxClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Content,
			androidCertificateDo.Alias_,
			androidCertificateDo.Storepass,
			androidCertificateDo.Category,
			androidCertificateDo.Sha1,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.Category.Eq(model.AndroidCertificateTypeRelease),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var certificateData []byte
	{
		log.Info(ctx, "decrypt android certificate data")
		var secret []byte
		secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
		if err != nil {
			return
		}
		certificateData, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取上传证书。
	var jksData []byte
	{
		log.Info(ctx, "run keytool to get certificate")
		var jksPath string
		jksPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks", certificateData)
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		var outPath string
		outPath, err = util.GenerateTemporaryFile(cc.ServiceNameBackend, "*.cer")
		if err != nil {
			log.Error(ctx, "failed to create file path", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, jksPath)
		var output []byte
		output, err = exec.Command(
			consts.KeytoolBinaryPath,
			"-J-Duser.language=en",
			"-J-Duser.country=US",
			"-export",
			"-rfc",
			"-v",
			"-keystore", jksPath,
			"-alias", androidCertificate.Alias_,
			"-file", outPath,
			"-storepass", androidCertificate.Storepass,
		).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to run command", err, output)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, outPath)
		jksData, err = os.ReadFile(outPath)
		if err != nil {
			log.Error(ctx, "failed to read file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录事件。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeDownloadGooglePlayCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:   app.Name,
			EventUser:  user.NameEn,
			EventAlias: androidCertificate.Alias_,
			EventDetail: util.GetPrintJSON(map[string]any{
				"category":      "上传证书",
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
		Name:   androidCertificate.Alias_ + ".cer",
		Size:   int64(len(jksData)),
		Reader: io.NopCloser(bytes.NewReader(jksData)),
	}

	return
}

// AndroidWebGetGooglePlayDeployCertificate 获取谷歌 Play 部署证书。
func AndroidWebGetGooglePlayDeployCertificate(ctx context.Context,
	req *protocol.AndroidWebGetGooglePlayDeployCertificateReq) (fileObj *FileInfo, err error) {

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
		if app.Platform != model.AppPlatformAndroid {
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库获取证书。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate information from database")
		androidCertificateDo := conn.MySQLTxClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.Category,
			androidCertificateDo.Content,
			androidCertificateDo.Alias_,
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Storepass,
			androidCertificateDo.Sha1,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
			androidCertificateDo.Keypass,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.Category.Eq(model.AndroidCertificateTypeRelease),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var certificateData []byte
	{
		log.Info(ctx, "decrypt android certificate data")
		var secret []byte
		secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
		if err != nil {
			return
		}
		certificateData, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取部署证书。
	var jksData []byte
	{
		log.Info(ctx, "run pepk to get deploy certificate")
		var jksPath string
		jksPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks", certificateData)
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, jksPath)
		var outPath string
		outPath, err = util.GenerateTemporaryFile(cc.ServiceNameBackend, "*.zip")
		if err != nil {
			log.Error(ctx, "failed to create file path", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		var pkeyPath string
		pkeyPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.pem", []byte(req.PublicKey))
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, pkeyPath)
		var output []byte
		output, err = exec.Command(
			consts.JavaBinaryPathForPepk, "-jar", consts.PepkJarPath,
			"--keystore="+jksPath,
			"--alias="+androidCertificate.Alias_,
			"--output="+outPath,
			"--encryption-key-path="+pkeyPath,
			"--include-cert",
			"--keystore-pass="+androidCertificate.Storepass,
			"--key-pass="+androidCertificate.Keypass,
			"--rsa-aes-encryption",
		).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to run command", err, output)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, outPath)
		jksData, err = os.ReadFile(outPath)
		if err != nil {
			log.Error(ctx, "failed to read file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录事件。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeDownloadGooglePlayCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:   app.Name,
			EventUser:  user.NameEn,
			EventAlias: androidCertificate.Alias_,
			EventDetail: util.GetPrintJSON(map[string]any{
				"category":      "部署证书",
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
		Name:   androidCertificate.Alias_ + ".zip",
		Size:   int64(len(jksData)),
		Reader: io.NopCloser(bytes.NewReader(jksData)),
	}

	return
}

// AndroidWebGetGooglePlayUpgradeCertificate 获取谷歌 Play 升级签名密钥。
func AndroidWebGetGooglePlayUpgradeCertificate(ctx context.Context,
	req *protocol.AndroidWebGetGooglePlayUpgradeCertificateReq) (fileObj *FileInfo, err error) {

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
		if app.Platform != model.AppPlatformAndroid {
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取部署证书。
	var deployAndroidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get deploy android certificate")
		androidCertificateDo := conn.MySQLTxClient(ctx).AndroidCertificate
		deployAndroidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.Category,
			androidCertificateDo.Content,
			androidCertificateDo.Alias_,
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Storepass,
			androidCertificateDo.Sha1,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
			androidCertificateDo.Keypass,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.DeployCertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.Category.Eq(model.AndroidCertificateTypeRelease),
			androidCertificateDo.CertificateID.Eq(req.DeployCertificateID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var deployCertificateData []byte
	var secret []byte
	{
		log.Info(ctx, "decrypt deploy android certificate")
		secret, err = getAESSecret(ctx, deployAndroidCertificate.AesKeyID)
		if err != nil {
			return
		}
		deployCertificateData, err = util.AESCBCDecrypt(deployAndroidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 部署证书写入文件系统。
	var deployCertificatePath string
	{
		log.Info(ctx, "write deploy certificate to disk")
		deployCertificatePath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks",
			deployCertificateData)
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, deployCertificatePath)
	}

	// 查库，获取上传证书。
	var uploadAndroidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get upload android certificate")
		androidCertificateDo := conn.MySQLTxClient(ctx).AndroidCertificate
		uploadAndroidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.Category,
			androidCertificateDo.Content,
			androidCertificateDo.Alias_,
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Storepass,
			androidCertificateDo.Sha1,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
			androidCertificateDo.Keypass,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.DeployCertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.Category.Eq(model.AndroidCertificateTypeRelease),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var uploadCertificateData []byte
	{
		log.Info(ctx, "decrypt upload android certificate")
		if deployAndroidCertificate.AesKeyID != uploadAndroidCertificate.AesKeyID {
			secret, err = getAESSecret(ctx, deployAndroidCertificate.AesKeyID)
			if err != nil {
				return
			}
		}
		uploadCertificateData, err = util.AESCBCDecrypt(uploadAndroidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 上传证书写入文件系统。
	var uploadCertificatePath string
	{
		log.Info(ctx, "write upload certificate to disk")
		uploadCertificatePath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks", uploadCertificateData)
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, uploadCertificatePath)
	}

	// 加密公钥写入文件系统。
	var pkeyPath string
	{
		log.Info(ctx, "write pkey to disk")
		pkeyPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.pem", []byte(req.PublicKey))
		if err != nil {
			log.Error(ctx, "failed to create file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, pkeyPath)
	}

	// 执行命令生成。
	var fileData []byte
	{
		log.Info(ctx, "run pepk to get certificate")
		var outPath string
		outPath, err = util.GenerateTemporaryFile(cc.ServiceNameBackend, "*.zip")
		if err != nil {
			log.Error(ctx, "failed to create file path", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		var output []byte
		output, err = exec.Command(
			consts.JavaBinaryPathForPepk, "-jar", consts.PepkJarPath,
			"--keystore", deployCertificatePath,
			"--alias", deployAndroidCertificate.Alias_,
			"--output", outPath,
			"--signing-keystore", uploadCertificatePath,
			"--signing-key-alias", uploadAndroidCertificate.Alias_,
			"--encryption-key-path", pkeyPath,
			"--keystore-pass", deployAndroidCertificate.Storepass,
			"--signing-store-pass", uploadAndroidCertificate.Storepass,
			"--key-pass", deployAndroidCertificate.Keypass,
			"--signing-key-pass", uploadAndroidCertificate.Keypass,
			"--rsa-aes-encryption",
		).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to run command", err, output)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, outPath)
		fileData, err = os.ReadFile(outPath)
		if err != nil {
			log.Error(ctx, "failed to read file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录事件。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeDownloadGooglePlayCertificate,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:   app.Name,
			EventUser:  user.NameEn,
			EventAlias: deployAndroidCertificate.Alias_,
			EventDetail: util.GetPrintJSON(map[string]any{
				"category":            "升级签名密钥",
				"deployCertificateId": req.DeployCertificateID,
				"deploySha1":          deployAndroidCertificate.Sha1,
				"deployOwner":         deployAndroidCertificate.Owner,
				"deployPublisher":     deployAndroidCertificate.Publisher,
				"deployNotAfter":      formatTime(&deployAndroidCertificate.NotAfter),
				"deployNotBefore":     formatTime(&deployAndroidCertificate.NotBefore),
				"deployType":          model.AllAndroidCertificateTypeDescriptions[deployAndroidCertificate.Category],
				"uploadCertificateId": req.UploadCertificateID,
				"uploadSha1":          uploadAndroidCertificate.Sha1,
				"uploadOwner":         uploadAndroidCertificate.Owner,
				"uploadPublisher":     uploadAndroidCertificate.Publisher,
				"uploadNotAfter":      formatTime(&uploadAndroidCertificate.NotAfter),
				"uploadNotBefore":     formatTime(&uploadAndroidCertificate.NotBefore),
				"uploadType":          model.AllAndroidCertificateTypeDescriptions[uploadAndroidCertificate.Category],
			}),
		})
		if err != nil {
			return
		}
	}

	fileObj = &FileInfo{
		Name:   deployAndroidCertificate.Alias_ + ".zip",
		Size:   int64(len(fileData)),
		Reader: io.NopCloser(bytes.NewReader(fileData)),
	}

	return
}

// AndroidWebGetCertificateFacebookDigest 获取证书的脸书摘要。
func AndroidWebGetCertificateFacebookDigest(ctx context.Context,
	req *protocol.AndroidWebGetCertificateFacebookDigestReq) (
	rsp *protocol.AndroidWebGetCertificateFacebookDigestRsp, err error) {

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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取部署证书。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate information from database")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.AesKeyID,
			androidCertificateDo.Alias_,
			androidCertificateDo.Storepass,
			androidCertificateDo.Content,
			androidCertificateDo.Category,
			androidCertificateDo.Sha1,
			androidCertificateDo.NotBefore,
			androidCertificateDo.NotAfter,
			androidCertificateDo.Owner,
			androidCertificateDo.Publisher,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	var jksBytes []byte
	{
		log.Info(ctx, "decrypt android certificate data")
		var secret []byte
		secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
		if err != nil {
			return
		}
		jksBytes, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt android certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 密钥库写入文件系统。
	var jksPath string
	{
		log.Info(ctx, "write certificate to disk")
		jksPath, err = util.CreateTemporaryFile(ctx, cc.ServiceNameBackend, "*.jks", jksBytes)
		if err != nil {
			log.Error(ctx, "failed to make a temporary jks file", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.RemoveFile(ctx, jksPath)
	}

	// 导出证书。
	var output []byte
	{
		log.Info(ctx, "run keytool to get certificate")
		command := exec.Command(
			consts.KeytoolBinaryPath,
			"-J-Duser.language=en",
			"-J-Duser.country=US",
			"-exportcert",
			"-alias", androidCertificate.Alias_,
			"-keystore", jksPath,
			"-storepass", androidCertificate.Storepass,
			"-storetype", "pkcs12",
		)
		errBuf := &bytes.Buffer{}
		command.Stderr = errBuf
		output, err = command.Output()
		if err != nil {
			log.Error(ctx, "failed to run command", err, errBuf.String(), output)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 计算脸书摘要。
	var digest string
	{
		log.Info(ctx, "get facebook digest")
		// 计算 SHA1 值。
		sha1Sum := sha1.Sum(output)
		// 编码 Base64。
		digest = base64.StdEncoding.EncodeToString(sha1Sum[:])
	}

	// 记录事件。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeGetFacebookCertificateDigest,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:  user.NameEn,
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

	rsp = &protocol.AndroidWebGetCertificateFacebookDigestRsp{Digest: digest}

	return
}

// AndroidWebSubmitAPKSigningJob 提交 APK 文件签名任务。
func AndroidWebSubmitAPKSigningJob(ctx context.Context, req *protocol.AndroidWebSubmitAPKSigningJobReq) (err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is invalid")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
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
			fileDo.UserID,
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
		if file == nil || file.Type != model.FileTypeAndroidSigning || file.UserID != user.ID || file.AppID != app.ID ||
			strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionAPK {
			log.Warn(ctx, "file is invalid", file)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.NotAfter,
			androidCertificateDo.ID,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificate.NotAfter) > 0 {
			log.Warn(ctx, "android certificate has expired")
			err = errs.New(consts.ErrParameterInvalid)
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
			UserID:           user.ID,
			Type:             model.AndroidSigningJobTypeAPK,
			CertificateID:    androidCertificate.ID,
			FileID:           req.FileID,
			SignatureSchemas: req.SignatureSchema,
			Source:           model.SourceWeb,
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

	return
}

// AndroidWebSubmitAABSigningJob 提交 AAB 文件签名任务。
func AndroidWebSubmitAABSigningJob(ctx context.Context, req *protocol.AndroidWebSubmitAABSigningJobReq) (err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is invalid")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
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
			fileDo.UserID,
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
		if file == nil || file.Type != model.FileTypeAndroidSigning || file.UserID != user.ID || file.AppID != app.ID ||
			strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionAAB {
			log.Warn(ctx, "file is invalid", file)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.NotAfter,
			androidCertificateDo.ID,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificate.NotAfter) > 0 {
			log.Warn(ctx, "android certificate has expired")
			err = errs.New(consts.ErrParameterInvalid)
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
			JobID:         jobID,
			AppID:         app.ID,
			UserID:        user.ID,
			Type:          model.AndroidSigningJobTypeAAB,
			CertificateID: androidCertificate.ID,
			FileID:        req.FileID,
			Source:        model.SourceWeb,
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

	return
}

// AndroidWebSubmitAPKPatchSigningJob 提交 APK 补丁包文件签名任务。
func AndroidWebSubmitAPKPatchSigningJob(ctx context.Context, req *protocol.AndroidWebSubmitAPKPatchSigningJobReq) (
	err error) {

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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is invalid")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
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
			fileDo.UserID,
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
		if file == nil || file.Type != model.FileTypeAndroidSigning || file.UserID != user.ID || file.AppID != app.ID ||
			strings.ToLower(filepath.Ext(file.Name)) != cc.ExtensionAPK {
			log.Warn(ctx, "file is invalid", file)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查询数据库，获取证书信息，并校验过期时间。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "verify android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Select(
			androidCertificateDo.NotAfter,
			androidCertificateDo.ID,
		).Where(
			androidCertificateDo.CertificateID.Eq(req.CertificateID),
			androidCertificateDo.AppID.Eq(app.ID),
			androidCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if time.Since(androidCertificate.NotAfter) > 0 {
			log.Warn(ctx, "android certificate has expired")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 将签名任务信息保存到数据库。
	var jobID string
	{
		log.Info(ctx, "save android signing job")
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
			UserID:           user.ID,
			Type:             model.AndroidSigningJobTypePatch,
			CertificateID:    androidCertificate.ID,
			FileID:           req.FileID,
			Source:           model.SourceWeb,
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

	return
}

// AndroidWebListSigningJobs 获取签名任务列表信息。
func AndroidWebListSigningJobs(ctx context.Context, req *protocol.AndroidWebListSigningJobsReq) (
	rsp *protocol.AndroidWebListSigningJobsRsp, err error) {

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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app is invalid")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查询安卓证书 IDs。
	var androidCertificateIDs []int
	{
		if len(req.CertificateAlias) > 0 {
			log.Info(ctx, "get android certificate ids from database")
			androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
			err = androidCertificateDo.WithContext(ctx).Select(
				androidCertificateDo.ID,
			).Where(
				androidCertificateDo.AppID.Eq(app.ID),
				androidCertificateDo.Alias_.Like("%"+req.CertificateAlias+"%"),
			).Scan(&androidCertificateIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve android certificate ids from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 查询用户 IDs。
	var userIDs []int
	{
		if len(req.KeyWord) > 0 {
			log.Info(ctx, "get user ids from database")
			userDo := conn.MySQLClient(ctx).User
			err = userDo.WithContext(ctx).Select(
				userDo.ID,
			).Where(
				userDo.NameEn.Like("%" + req.KeyWord + "%"),
			).Or(
				userDo.NameZh.Like("%" + req.KeyWord + "%"),
			).Scan(&userIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user ids from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			var apiAccountIDs []int
			apiAccountDo := conn.MySQLClient(ctx).APIAccount
			err = apiAccountDo.WithContext(ctx).Select(
				apiAccountDo.ID,
			).Where(
				apiAccountDo.AppID.Eq(app.ID),
				apiAccountDo.AccountID.Like("%"+req.KeyWord+"%"),
			).Scan(&apiAccountIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve api account ids from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			userIDs = append(userIDs, apiAccountIDs...)
		}
	}

	// 查询任务。
	var androidSigningJobs []*model.AndroidSigningJob
	var count int
	{
		log.Info(ctx, "get android signing jobs from database")
		var tableNames []string
		tableNames, err = getAllAndroidSignIngJobTableNames(ctx)
		if err != nil {
			return
		}
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob
		count, err = androidSigningJobDo.WithContext(ctx).Count2(
			tableNames, app.ID, req.KeyWord, req.Status, androidCertificateIDs, userIDs)
		if err != nil {
			log.Error(ctx, "failed to count android signing jobs", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		androidSigningJobs, err = androidSigningJobDo.WithContext(ctx).List(
			tableNames,
			app.ID,
			req.KeyWord,
			req.Status,
			androidCertificateIDs,
			userIDs,
			req.PageSize,
			(req.PageNumber-1)*req.PageSize,
		)
		if err != nil {
			log.Error(ctx, "failed to retrieve android signing jobs from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(androidSigningJobs) <= 0 {
			rsp = &protocol.AndroidWebListSigningJobsRsp{Count: count}
			return
		}
	}

	// 查询用户信息。
	var userIDToName map[int]string
	var apiAccountIDToName map[int]string
	var fileIDs []string
	var certificateIDs2 []int
	{
		log.Info(ctx, "get user information from database")
		userIDs2 := make([]int, 0, len(androidSigningJobs)/2)
		apiAccountIDs := make([]int, 0, len(androidSigningJobs)/2)
		certificateIDs2 = make([]int, 0, len(androidSigningJobs))
		fileIDs = make([]string, 0, len(androidSigningJobs))
		for _, v := range androidSigningJobs {
			switch v.Source {
			case model.SourceWeb:
				userIDs2 = append(userIDs2, v.UserID)
			case model.SourceAPI:
				apiAccountIDs = append(apiAccountIDs, v.UserID)
			default:
				log.Warn(ctx, "unknown source type", v.Source)
			}
			certificateIDs2 = append(certificateIDs2, v.CertificateID)
			fileIDs = append(fileIDs, v.FileID)
		}
		userIDs2 = util.CleanNumbers(userIDs2)
		apiAccountIDs = util.CleanNumbers(apiAccountIDs)
		certificateIDs2 = util.CleanNumbers(certificateIDs2)
		fileIDs = util.CleanStrings(fileIDs)
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs2)
		if err != nil {
			return
		}
		apiAccountIDToName, err = GetAPIAccountNamesByIDs(ctx, apiAccountIDs)
		if err != nil {
			return
		}
	}

	// 查询文件信息。
	var fileIDToName map[string]string
	{
		fileIDToName, err = GetFileNamesByIDs(ctx, fileIDs)
		if err != nil {
			return
		}
	}

	// 查询证书别名。
	var androidCertificateIDToAlias map[int]string
	{
		if len(certificateIDs2) > 0 {
			log.Info(ctx, "get android certificate information from database")
			var androidCertificates []*model.AndroidCertificate
			androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
			androidCertificates, err = androidCertificateDo.WithContext(ctx).Select(
				androidCertificateDo.ID,
				androidCertificateDo.Alias_,
			).Where(
				androidCertificateDo.ID.In(certificateIDs2...),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve android certificate information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			androidCertificateIDToAlias = util.ListToMap(
				androidCertificates, func(e *model.AndroidCertificate) (int, string) { return e.ID, e.Alias_ })
		} else {
			androidCertificateIDToAlias = make(map[int]string)
		}
	}

	// 组装数据。
	var list []*protocol.AndroidWebListSigningJobsItem
	{
		log.Info(ctx, "assembly data")
		list = make([]*protocol.AndroidWebListSigningJobsItem, len(androidSigningJobs))
		for i, v := range androidSigningJobs {
			userName := ""
			switch v.Source {
			case model.SourceWeb:
				userName = userIDToName[v.UserID]
			case model.SourceAPI:
				userName = apiAccountIDToName[v.UserID]
			default:
				log.Warn(ctx, "unknown source type", v.Source)
			}
			signingConfig := ""
			switch v.Type {
			case model.AndroidSigningJobTypeAPK:
				signingConfig = "signatureSchema=" + strings.Join(
					util.ListTo(v.SignatureSchemas, func(e int) string { return fmt.Sprintf("v%d", e) }), ",")
			case model.AndroidSigningJobTypePatch:
				signingConfig = fmt.Sprintf("minimumSdkVersion=%v", v.MinimumSdkLevel)
			default:
			}
			list[i] = &protocol.AndroidWebListSigningJobsItem{
				JobID:            v.JobID,
				Type:             v.Type,
				Source:           v.Source,
				CertificateAlias: androidCertificateIDToAlias[v.CertificateID],
				SigningConfig:    signingConfig,
				FileName:         fileIDToName[v.FileID],
				FileID:           v.FileID,
				User:             userName,
				CreatedTime:      formatTime(&v.CreatedTime),
				FinishedTime:     formatTime(&v.FinishedTime),
				Log:              v.Log,
				SignedFileID:     v.SignedFileID,
			}
		}
	}

	rsp = &protocol.AndroidWebListSigningJobsRsp{Count: count, List: list}

	return
}

// AndroidWebRemoveOrganization 删除证书主体。
func AndroidWebRemoveOrganization(ctx context.Context, req *protocol.AndroidWebRemoveOrganizationReq) (err error) {
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

	// 删除安卓证书主体。
	{
		log.Info(ctx, "remove android organization from database")
		androidOrganizationDo := conn.MySQLClient(ctx).AndroidOrganization
		var sqlResult gen.ResultInfo
		sqlResult, err = androidOrganizationDo.WithContext(ctx).Where(
			androidOrganizationDo.ID.Eq(req.ID),
		).Delete()
		if err != nil {
			log.Error(ctx, "failed to remove android organization from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "remove android organization from database without affected rows")
		}
	}

	return
}

// AndroidWebDeleteCertificate 删除安卓证书。
func AndroidWebDeleteCertificate(ctx context.Context, req *protocol.AndroidWebDeleteCertificateReq) (err error) {
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
		if app.Platform != model.AppPlatformAndroid {
			log.Warn(ctx, "app platform is invalid")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 查库，获取证书。
	var androidCertificate *model.AndroidCertificate
	{
		log.Info(ctx, "get android certificate")
		androidCertificateTxDo := conn.MySQLTxClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateTxDo.WithContext(ctx).Select(
			androidCertificateTxDo.ID,
			androidCertificateTxDo.Alias_,
			androidCertificateTxDo.Sha1,
			androidCertificateTxDo.Owner,
			androidCertificateTxDo.Publisher,
			androidCertificateTxDo.NotBefore,
			androidCertificateTxDo.NotAfter,
			androidCertificateTxDo.Category,
		).Where(
			androidCertificateTxDo.CertificateID.Eq(req.CertificateID),
			androidCertificateTxDo.AppID.Eq(app.ID),
			androidCertificateTxDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "android certificate not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 删除证书。
	var now time.Time
	{
		log.Info(ctx, "delete android certificate")
		now = time.Now()
		androidCertificateTxDo := conn.MySQLTxClient(ctx).AndroidCertificate
		_, err = androidCertificateTxDo.WithContext(ctx).Where(
			androidCertificateTxDo.ID.Eq(androidCertificate.ID),
		).UpdateColumnSimple(
			androidCertificateTxDo.DeletedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to delete android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemoveAndroidCertificate,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:  user.NameEn,
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

	return
}

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

// AndroidInternalGetSigningJob 获取签名任务信息。
func AndroidInternalGetSigningJob(ctx context.Context, req *protocol.AndroidInternalGetSigningJobReq) (
	androidSigningJob *model.AndroidSigningJob, err error) {

	// 查库，获取信息。
	{
		log.Info(ctx, "get android job")
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob.Table(model.GetAndroidSigningJobByID(req.JobID))
		androidSigningJob, err = androidSigningJobDo.WithContext(ctx).Where(
			androidSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android job information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	return
}

// AndroidInternalGetCertificate 获取安卓证书信息。
func AndroidInternalGetCertificate(ctx context.Context, req *protocol.AndroidInternalGetCertificateReq) (
	androidCertificate *model.AndroidCertificate, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Where(
			androidCertificateDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	// 解密证书内容。
	{
		if androidCertificate != nil && len(androidCertificate.Content) > 0 {
			var secret []byte
			secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
			if err != nil {
				return
			}
			androidCertificate.Content, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt android certificate content", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	return
}

// AndroidInternalUpdateSigningJob 更新任务信息。
func AndroidInternalUpdateSigningJob(ctx context.Context, req *protocol.AndroidInternalUpdateSigningJobReq) (
	err error) {

	// 更新任务。
	{
		log.Info(ctx, "update android signing job")
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob.Table(model.GetAndroidSigningJobByID(req.JobID))
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, androidSigningJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(androidSigningJobDo.Log, logString+"\n"))
		}
		if len(req.SignedFileID) > 0 {
			assignExprs = append(assignExprs, androidSigningJobDo.SignedFileID.Value(req.SignedFileID))
		}
		finishedTime := time.Time(req.FinishedTime)
		if !finishedTime.IsZero() {
			assignExprs = append(assignExprs, androidSigningJobDo.FinishedTime.Value(finishedTime))
		}
		if len(assignExprs) <= 0 {
			log.Error(ctx, "no updated values")
			return
		}
		_, err = androidSigningJobDo.WithContext(ctx).Where(
			androidSigningJobDo.JobID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update android signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

func createAndroidSigningJob(ctx context.Context, job *model.AndroidSigningJob) error {
	androidSignJobTxDo := conn.MySQLTxClient(ctx).AndroidSigningJob.Table(model.GetAndroidSigningJobByID(job.JobID))
	selectedFields := make([]field.Expr, 0, 12)
	selectedFields = append(selectedFields,
		androidSignJobTxDo.JobID,
		androidSignJobTxDo.AppID,
		androidSignJobTxDo.UserID,
		androidSignJobTxDo.Type,
		androidSignJobTxDo.CertificateID,
		androidSignJobTxDo.FileID,
		androidSignJobTxDo.CertificateID,
		androidSignJobTxDo.Source,
		androidSignJobTxDo.Status,
		androidSignJobTxDo.CreatedTime,
	)
	if len(job.SignatureSchemas) > 0 {
		selectedFields = append(selectedFields, androidSignJobTxDo.SignatureSchemas)
	}
	if job.MinimumSdkLevel > 0 {
		selectedFields = append(selectedFields, androidSignJobTxDo.MinimumSdkLevel)
	}
	err := androidSignJobTxDo.WithContext(ctx).Select(selectedFields...).Create(job)
	if err != nil {
		log.Error(ctx, "failed to create android signing job in database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}

func getAllAndroidSignIngJobTableNames(ctx context.Context) ([]string, error) {
	androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob
	tableNames, err := androidSigningJobDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to query android signing job tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return tableNames, nil
}

func parseKeystore(ctx context.Context, jksPath, storepass string) (*keystoreInfo, error) {
	outputBytes, err := exec.Command(
		consts.KeytoolBinaryPath,
		"-J-Duser.language=en",
		"-J-Duser.country=US",
		"-list",
		"-v",
		"-storepass", storepass,
		"-keystore", jksPath,
	).CombinedOutput()
	log.Debug(ctx, "output of running the keytool", outputBytes)
	output := string(outputBytes)
	if err != nil {
		if strings.Contains(output, "password was incorrect") {
			return nil, errs.New(consts.ErrKeystoreStorePassInvalid)
		}
		log.Error(ctx, "failed to parse keystore file", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	lines := strings.Split(string(outputBytes), "\n")
	jksInfo := keystoreInfo{}
	for i := len(lines) - 1; i >= 0; i-- {
		v := util.TrimBlank(lines[i])
		if len(v) <= 0 || !strings.Contains(v, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(v, "Keystore type: "):
			jksInfo.StoreType = v[len("Keystore type: "):]
		case strings.HasPrefix(v, "Keystore provider: "):
			jksInfo.StoreProvider = v[len("Keystore provider: "):]
		case strings.HasPrefix(v, "Alias name: "):
			jksInfo.AliasName = util.TrimBlank(v[len("Alias name: "):])
		case strings.HasPrefix(v, "Creation date: "):
			var t time.Time
			t, err = time.Parse("Jan 2, 2006", v[len("Creation date: "):])
			if err != nil {
				log.Error(ctx, "failed to parse time", err)
			} else {
				jksInfo.CreationDate = t
			}
		case strings.HasPrefix(v, "Owner: "):
			jksInfo.Owner = v[len("Owner: "):]
		case strings.HasPrefix(v, "Issuer: "):
			jksInfo.Issuer = v[len("Issuer: "):]
		case strings.HasPrefix(v, "Serial number: "):
			jksInfo.SerialNumber = v[len("Serial number: "):]
		case strings.HasPrefix(v, "Valid from: "):
			pair := strings.Split(v[len("Valid from: "):], " until: ")
			var t time.Time
			t, err = time.Parse("Mon Jan 02 15:04:05 CST 2006", pair[0])
			if err != nil {
				log.Error(ctx, "unknown time", err, pair[0])
			} else {
				jksInfo.ValidFrom = t
			}
			t, err = time.Parse("Mon Jan 02 15:04:05 CST 2006", pair[1])
			if err != nil {
				log.Error(ctx, "unknown time", err, pair[1])
			} else {
				jksInfo.ValidUntil = t
			}
		case strings.HasPrefix(v, "SHA1: "):
			jksInfo.SHA1 = strings.ReplaceAll(v[len("SHA1: "):], ":", "")
		case strings.HasPrefix(v, "SHA256: "):
			jksInfo.SHA256 = strings.ReplaceAll(v[len("SHA256: "):], ":", "")
		case strings.HasPrefix(v, "Signature algorithm name: "):
			jksInfo.SignatureAlgorithmName = v[len("Signature algorithm name: "):]
		case strings.HasPrefix(v, "Subject Public Key Algorithm: "):
			jksInfo.SubjectPublicKeyAlgorithm = v[len("Subject Public Key Algorithm: "):]
		case strings.HasPrefix(v, "Version: "):
			jksInfo.Version = v[len("Version: "):]
		}
	}
	return &jksInfo, nil
}
