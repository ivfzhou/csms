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
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"howett.net/plist"
	"software.sslmate.com/src/go-pkcs12"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
	fp "gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
)

// AppleWebApplyBundleID 申请 Bundle ID。
func AppleWebApplyBundleID(ctx context.Context, req *protocol.AppleWebApplyBundleIDReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 判断 Bundle ID 没有被注册。
	{
		log.Info(ctx, "check whether apple bundle id has not been registered")
		appleBundleIDDo := conn.MySQLTxClient(ctx).AppleBundleID
		var count int64
		count, err = appleBundleIDDo.WithContext(ctx).Clauses(query.ForUpdate()).Where(
			appleBundleIDDo.BundleID.Eq(req.BundleID),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple apple bundle id information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			err = errs.New(consts.ErrAppleBundleIDExist)
			return
		}
	}

	// 申请 Bundle ID。
	var now time.Time
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "register apple bundle id")
		now = time.Now()
		switch req.Type {
		case model.AppleBundleIDTypeAppStore:
			var token string
			token, err = generateAppleAPIToken(ctx)
			if err != nil {
				return
			}
			var apiResult *appleAPIResponse
			apiResult, err = httpAppleAPIApplyBundleID(ctx, token, req.BundleID)
			if err != nil {
				return
			}

			// 保存 Bundle ID 信息到数据库。
			log.Info(ctx, "save apple bundle id information")
			appleBundleIDTxDo := conn.MySQLTxClient(ctx).AppleBundleID
			appleBundleID = &model.AppleBundleID{
				AppID:       app.ID,
				UserID:      user.ID,
				InAppleID:   apiResult.Data.ID,
				BundleID:    req.BundleID,
				Environment: req.Type,
				CreatedTime: now,
				Platform:    model.AllApplePlatformDescriptionToNumber[apiResult.Data.Attributes.Platform],
			}
			if err = appleBundleIDTxDo.WithContext(ctx).Select(
				appleBundleIDTxDo.InAppleID,
				appleBundleIDTxDo.BundleID,
				appleBundleIDTxDo.AppID,
				appleBundleIDTxDo.Environment,
				appleBundleIDTxDo.Platform,
				appleBundleIDTxDo.CreatedTime,
				appleBundleIDTxDo.UserID,
			).Create(appleBundleID); err != nil {
				log.Error(ctx, "failed to save apple bundle id information in database", err)
				return errs.NewWithError(consts.ErrSystem, err)
			}
		case model.AppleBundleIDTypeInHouse:
			var apiResult *fp.ApplyInHouseBundleIDRsp
			apiResult, err = httpFastlaneApplyBundleID(ctx, req.BundleID)
			if err != nil {
				return
			}

			// 保存 Bundle ID 信息到数据库。
			log.Info(ctx, "save apple bundle id information")
			appleBundleIDTxDo := conn.MySQLTxClient(ctx).AppleBundleID
			appleBundleID = &model.AppleBundleID{
				AppID:       app.ID,
				UserID:      user.ID,
				BundleID:    req.BundleID,
				Environment: req.Type,
				CreatedTime: now,
				Platform:    model.AllApplePlatformDescriptionToNumber[apiResult.Platform],
				InAppleID:   apiResult.ID,
			}
			if err = appleBundleIDTxDo.WithContext(ctx).Select(
				appleBundleIDTxDo.BundleID,
				appleBundleIDTxDo.AppID,
				appleBundleIDTxDo.Environment,
				appleBundleIDTxDo.CreatedTime,
				appleBundleIDTxDo.Platform,
				appleBundleIDTxDo.InAppleID,
				appleBundleIDTxDo.UserID,
			).Create(appleBundleID); err != nil {
				log.Error(ctx, "failed to save apple bundle id information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		default:
			log.Error(ctx, "unknown apple bundle id type", req.BundleID, req.Type)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyAppleBundleID,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"id":          appleBundleID.InAppleID,
				"environment": model.AllAppleBundleIDDescriptions[appleBundleID.Environment],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebModifyBundleID 修改 Bundle ID 能力项。
func AppleWebModifyBundleID(ctx context.Context, req *protocol.AppleWebModifyBundleIDReq) (hasFailed bool, err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，确保 Bundle ID 存在，属于应用。
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle id information")
		appleBundleIDTxDo := conn.MySQLTxClient(ctx).AppleBundleID
		appleBundleID, err = appleBundleIDTxDo.WithContext(ctx).Select(
			appleBundleIDTxDo.ID,
			appleBundleIDTxDo.InAppleID,
			appleBundleIDTxDo.Environment,
			appleBundleIDTxDo.Capabilities,
		).Clauses(query.ForUpdate()).Where(
			appleBundleIDTxDo.BundleID.Eq(req.BundleID),
			appleBundleIDTxDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "apple bundle id not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 分拣出能力项。
	var bundleIDCapabilityMap map[string]bool
	var remainCapabilities []string
	{
		log.Info(ctx, "deal apple bundle id capabilities")
		bundleIDCapabilityMap, remainCapabilities = dealBundleIDCapabilities(
			req.Capabilities, appleBundleID.Capabilities)
		if len(bundleIDCapabilityMap) <= 0 {
			return
		}
	}

	// 更新 Bundle ID 能力。
	var capabilities []string
	{
		log.Info(ctx, "modify apple bundle id capabilities in apple")
		switch appleBundleID.Environment {
		case model.AppleBundleIDTypeInHouse:
			capabilities = req.Capabilities
			if err = httpFastlaneModifyBundleIDCapabilities(ctx, req.BundleID, bundleIDCapabilityMap); err != nil {
				return
			}
		case model.AppleBundleIDTypeAppStore:
			var token string
			token, err = generateAppleAPIToken(ctx)
			if err != nil {
				return
			}
			hasSuccess := false
			for v, ok := range bundleIDCapabilityMap {
				if ok {
					err = httpAppleAPIEnableBundleIDCapability(ctx, token, appleBundleID.InAppleID, v)
					if err != nil {
						hasFailed = true
						continue
					}
					hasSuccess = true
					remainCapabilities = append(remainCapabilities, v)
				} else {
					err = httpAppleAPIRemoveBundleIDCapability(ctx, token, appleBundleID.InAppleID, v)
					if err != nil {
						hasFailed = true
						remainCapabilities = append(remainCapabilities, v)
						continue
					}
					hasSuccess = true
				}
			}
			if hasFailed && !hasSuccess {
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			capabilities = remainCapabilities
		default:
			log.Error(ctx, "unknown apple bundle id environment", appleBundleID.Environment)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 入库，保存能力项。
	var now time.Time
	{
		log.Info(ctx, "update apple bundle id in database")
		now = time.Now()
		appleBundleIDTxDo := conn.MySQLTxClient(ctx).AppleBundleID
		_, err = appleBundleIDTxDo.WithContext(ctx).Where(appleBundleIDTxDo.ID.Eq(appleBundleID.ID)).
			UpdateColumnSimple(
				appleBundleIDTxDo.Capabilities.Value(model.StringList(capabilities)),
				appleBundleIDTxDo.UpdatedTime.Value(now),
			)
		if err != nil {
			log.Error(ctx, "failed to update apple bundle id in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存事件记录。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeModifyAppleBundleID,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"id":           appleBundleID.InAppleID,
				"capabilities": capabilities,
				"environment":  model.AllAppleBundleIDDescriptions[appleBundleID.Environment],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebApplyCertificate 申请苹果签名证书。
func AppleWebApplyCertificate(ctx context.Context, req *protocol.AppleWebApplyCertificateReq) (err error) {
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

	// 请求苹果服务器申请证书。
	var apiResult *appleAPIResponse
	{
		log.Info(ctx, "apply certificate from apple")
		var token string
		token, err = generateAppleAPIToken(ctx)
		if err != nil {
			return
		}
		apiResult, err = httpAppleAPIApplyCertificate(ctx, token, req.Type)
		if err != nil {
			return
		}
	}

	// 解析证书。
	var certificateData []byte
	var appleCertificate *model.AppleCertificate
	var certificate *x509.Certificate
	{
		log.Info(ctx, "parse certificate")
		certificateData, err = base64.StdEncoding.DecodeString(apiResult.Data.Attributes.CertificateContent)
		if err != nil {
			log.Error(ctx, "failed to base64 decode certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		certificate, err = x509.ParseCertificate(certificateData)
		if err != nil {
			log.Error(ctx, "failed to parse apple certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		var certificateID string
		certificateID, err = generateID(ctx, IDAppleCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAppleCertificate, certificateID),
					"failed to reclaim apple certificate id", certificateID)
			}
		}()
		appleCertificate = &model.AppleCertificate{
			Category:           model.AppleCertificateCategorySigning,
			CertificateID:      certificateID,
			UserID:             user.ID,
			InAppleID:          apiResult.Data.ID,
			Type:               req.Type,
			Password:           util.RandomPrintableASCIINoSpaceString(consts.AppleCertificatePasswordLength),
			Publisher:          certificate.Issuer.String(),
			Owner:              certificate.Subject.String(),
			SignatureAlgorithm: certificate.SignatureAlgorithm.String(),
			PublicKeyAlgorithm: certificate.PublicKeyAlgorithm.String(),
			SerialNumber:       certificate.SerialNumber.String(),
			NotBefore:          certificate.NotBefore,
			NotAfter:           certificate.NotAfter,
			CreatedTime:        time.Now(),
		}
		certificateData, err = util.CertificateDERToPEM(certificateData)
		if err != nil {
			log.Error(ctx, "failed to reformat certificate content", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 证书与私钥合并。
	{
		log.Info(ctx, "parse private key")
		keyBlock, _ := pem.Decode([]byte(cfg.Get().Apple().CertificatePrivateKey()))
		if keyBlock == nil {
			log.Error(ctx, "apple certificate private key is nil")
			err = errs.New(consts.ErrSystem)
			return
		}
		var privateKey any
		switch keyBlock.Type {
		case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		case "PRIVATE KEY":
			privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		default:
			log.Error(ctx, "not support private key type", keyBlock.Type)
			err = errs.New(consts.ErrSystem)
			return
		}
		if err != nil {
			log.Error(ctx, "failed to parse private key", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		certificateData, err = pkcs12.Modern.Encode(privateKey, certificate, nil, appleCertificate.Password)
		if err != nil {
			log.Error(ctx, "failed to generate pkcs12 content", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 加密证书。
	{
		log.Info(ctx, "encrypt certificate")
		var aesKey *model.AesKey
		aesKey, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		appleCertificate.AesKeyID = aesKey.ID
		appleCertificate.Content, err = util.AESCBCEncrypt(aesKey.Secret, certificateData)
		if err != nil {
			log.Error(ctx, "failed to encrypt certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存证书入库。
	{
		log.Info(ctx, "save apple certificate information")
		appleSignCertificateTxDo := conn.MySQLTxClient(ctx).AppleCertificate
		err = appleSignCertificateTxDo.WithContext(ctx).Select(
			appleSignCertificateTxDo.CertificateID,
			appleSignCertificateTxDo.UserID,
			appleSignCertificateTxDo.InAppleID,
			appleSignCertificateTxDo.Category,
			appleSignCertificateTxDo.Type,
			appleSignCertificateTxDo.Password,
			appleSignCertificateTxDo.Publisher,
			appleSignCertificateTxDo.Owner,
			appleSignCertificateTxDo.SignatureAlgorithm,
			appleSignCertificateTxDo.PublicKeyAlgorithm,
			appleSignCertificateTxDo.SerialNumber,
			appleSignCertificateTxDo.NotBefore,
			appleSignCertificateTxDo.NotAfter,
			appleSignCertificateTxDo.CreatedTime,
			appleSignCertificateTxDo.AesKeyID,
			appleSignCertificateTxDo.Content,
		).Create(appleCertificate)
		if err != nil {
			log.Error(ctx, "failed to save apple certificate information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// AppleWebListBundleIDs 获取 Bundle ID 列表。
func AppleWebListBundleIDs(ctx context.Context) (rsp *protocol.AppleWebListBundleIDsRsp, err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取 Bundle ID。
	var appleBundleIDs []*model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle ids")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		appleBundleIDs, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.BundleID,
			appleBundleIDDo.Environment,
			appleBundleIDDo.CreatedTime,
			appleBundleIDDo.UserID,
			appleBundleIDDo.Capabilities,
		).Where(
			appleBundleIDDo.AppID.Eq(app.ID),
		).Order(appleBundleIDDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple bundle ids", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(appleBundleIDs) <= 0 {
			return
		}
	}

	// 查库，获取用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get user information")
		var userIDs []int
		userIDs = util.ListTo(appleBundleIDs, func(e *model.AppleBundleID) int {
			return e.UserID
		})
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
	}

	// 组转数据。
	{
		list := make([]*protocol.AppleWebListBundleIDsItem, len(appleBundleIDs))
		for i, v := range appleBundleIDs {
			list[i] = &protocol.AppleWebListBundleIDsItem{
				ID:           v.BundleID,
				Env:          v.Environment,
				CreatedTime:  formatTime(&v.CreatedTime),
				Creator:      userIDToName[v.UserID],
				Capabilities: util.ListDropZero(v.Capabilities),
			}
		}
		rsp = &protocol.AppleWebListBundleIDsRsp{List: list}
	}

	return
}

// AppleWebListCertificates 获取苹果证书信息列表。
func AppleWebListCertificates(ctx context.Context) (rsp *protocol.AppleWebListCertificatesRsp, err error) {
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

	// 获取签名证书信息。
	var appleCertificates []*model.AppleCertificate
	{
		log.Info(ctx, "get apple certificates")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		appleCertificates, err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.UserID,
			appleCertificateDo.CertificateID,
			appleCertificateDo.Type,
			appleCertificateDo.Owner,
			appleCertificateDo.Publisher,
			appleCertificateDo.SignatureAlgorithm,
			appleCertificateDo.PublicKeyAlgorithm,
			appleCertificateDo.CreatedTime,
			appleCertificateDo.NotAfter,
		).Where(
			appleCertificateDo.Category.Eq(model.AppleCertificateCategorySigning),
			appleCertificateDo.DeletedTime.IsNull(),
		).Order(
			appleCertificateDo.CreatedTime.Desc(),
			appleCertificateDo.ID.Desc(),
		).Order(appleCertificateDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple certificates", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDs := util.CleanNumbers(util.ListTo(appleCertificates,
			func(e *model.AppleCertificate) int { return e.UserID }))
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AppleWebListCertificatesItem, len(appleCertificates))
		for i, v := range appleCertificates {
			list[i] = &protocol.AppleWebListCertificatesItem{
				ID:                 v.CertificateID,
				Type:               v.Type,
				Owner:              v.Owner,
				Publisher:          v.Publisher,
				SignatureAlgorithm: v.SignatureAlgorithm,
				PublicKeyAlgorithm: v.PublicKeyAlgorithm,
				CreatedTime:        formatTime(&v.CreatedTime),
				ExpirationTime:     formatTime(&v.NotAfter),
				Creator:            userIDToName[v.UserID],
			}
		}
		rsp = &protocol.AppleWebListCertificatesRsp{List: list}
	}

	return
}

// AppleWebRegisterDevice 注册测试设备。
func AppleWebRegisterDevice(ctx context.Context, req *protocol.AppleWebRegisterDeviceReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 入库，添加设备记录。
	var now time.Time
	var appleDeviceID int
	{
		log.Info(ctx, "create apple device to database")
		appleDeviceTxDo := conn.MySQLTxClient(ctx).AppleDevice
		now = time.Now()
		appleDevice := &model.AppleDevice{
			AppID:       app.ID,
			UserID:      user.ID,
			Udid:        req.UDID,
			Remark:      req.Remark,
			Platform:    getPlatformByAppleDevice(req.Device),
			Status:      model.AppleDeviceStatusApproving,
			CreatedTime: now,
		}
		err = appleDeviceTxDo.WithContext(ctx).Select(
			appleDeviceTxDo.AppID,
			appleDeviceTxDo.UserID,
			appleDeviceTxDo.Udid,
			appleDeviceTxDo.Remark,
			appleDeviceTxDo.Platform,
			appleDeviceTxDo.Status,
			appleDeviceTxDo.CreatedTime,
		).Create(appleDevice)
		if err != nil {
			log.Error(ctx, "failed to create apple device to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		appleDeviceID = appleDevice.ID
	}

	// 查询系统管理员。
	var systemAdmins []int
	{
		log.Info(ctx, "retrieve system administrators from database")
		userRoleDo := conn.MySQLClient(ctx).UserRole
		err = userRoleDo.WithContext(ctx).Select(
			userRoleDo.UserID,
		).Where(
			userRoleDo.Role.Eq(model.UserRoleSystemAdmin),
		).Scan(&systemAdmins)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to get system administrator from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(systemAdmins) <= 0 {
			// 没有系统管理员。
			log.Warn(ctx, "no system administrators were found")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 入库，创建待办。
	{
		log.Info(ctx, "create todo")
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		err = todoTxDo.WithContext(ctx).Select(
			todoTxDo.AppID,
			todoTxDo.Type,
			todoTxDo.ApplierID,
			todoTxDo.Candidates,
			todoTxDo.Information,
			todoTxDo.Status,
			todoTxDo.CreatedTime,
		).Create(&model.Todo{
			AppID:       app.ID,
			Type:        model.TodoTypeRegisterAppleDevice,
			ApplierID:   user.ID,
			Candidates:  systemAdmins,
			Information: strconv.Itoa(appleDeviceID),
			Status:      model.TodoStatusProcessing,
			CreatedTime: now,
		})
		if err != nil {
			log.Error(ctx, "failed to create todo to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRegisterAppleDevice,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:             user.NameEn,
			EventApp:              app.Name,
			EventAppleDeviceModel: req.UDID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"device": req.Device,
				"remark": req.Remark,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebListDevices 设备列表。
func AppleWebListDevices(ctx context.Context, req *protocol.AppleWebListDevicesReq) (
	rsp *protocol.AppleWebListDevicesRsp, err error) {

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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取设备记录。
	var appleDevices []*model.AppleDevice
	var count int64
	{
		log.Info(ctx, "get apple devices information")
		appleDeviceDo := conn.MySQLClient(ctx).AppleDevice
		count, err = appleDeviceDo.WithContext(ctx).Where(
			appleDeviceDo.AppID.Eq(app.ID),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count apple devices from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		appleDevices, err = appleDeviceDo.WithContext(ctx).Select(
			appleDeviceDo.Model,
			appleDeviceDo.Udid,
			appleDeviceDo.CreatedTime,
			appleDeviceDo.Remark,
			appleDeviceDo.Status,
			appleDeviceDo.UserID,
		).Where(
			appleDeviceDo.AppID.Eq(app.ID),
		).Limit(req.PageSize).Offset((req.PageNumber - 1) * req.PageSize).Order(appleDeviceDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple devices information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(appleDevices) <= 0 {
			rsp = &protocol.AppleWebListDevicesRsp{Count: count}
			return
		}
	}

	// 查库，获取用户信息。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDs := util.CleanNumbers(util.ListTo(appleDevices, func(e *model.AppleDevice) int { return e.UserID }))
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AppleWebListDevicesItem, len(appleDevices))
		for i, v := range appleDevices {
			list[i] = &protocol.AppleWebListDevicesItem{
				Model:       v.Model,
				UDID:        v.Udid,
				User:        userIDToName[v.UserID],
				CreatedTime: formatTime(&v.CreatedTime),
				Remark:      v.Remark,
				Status:      v.Status,
			}
		}
		rsp = &protocol.AppleWebListDevicesRsp{Count: count, List: list}
	}

	return
}

// AppleWebApplyProfile 申请描述文件。
func AppleWebApplyProfile(ctx context.Context, req *protocol.AppleWebApplyProfileReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Info(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取 Bundle ID 信息，判断 Bundle ID 属于应用，类型正确。
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle id information from database")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		appleBundleID, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.ID,
			appleBundleIDDo.InAppleID,
		).Where(
			appleBundleIDDo.BundleID.Eq(req.BundleID),
			appleBundleIDDo.AppID.Eq(app.ID),
			appleBundleIDDo.Environment.Eq(model.AppleBundleIDTypeAppStore),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "apple bundle id not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查库，获取签名证书。
	var appleCertificate *model.AppleCertificate
	var now time.Time
	var certificateType string
	{
		log.Info(ctx, "get apple certificate information from database")
		certificateType = getAppleCertificateType(req.Platform, req.Type)
		if len(certificateType) <= 0 {
			log.Warn(ctx, "no certificate type found")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		now = time.Now()
		appleCertificate, err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.InAppleID,
			appleCertificateDo.ID,
			appleCertificateDo.CertificateID,
		).Where(
			appleCertificateDo.Type.Eq(certificateType),
			appleCertificateDo.DeletedTime.IsNull(),
			appleCertificateDo.NotAfter.Gt(now),
			appleCertificateDo.NotBefore.Lt(now),
		).Order(appleCertificateDo.NotAfter.Desc()).Limit(1).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.New(consts.ErrNoAppleSigningCertificate)
				return
			}
			log.Error(ctx, "failed to retrieve apple certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取应用所有的测试设备。
	var deviceIDs []string
	var udids []string
	{
		if slices.Contains([]string{
			model.AppleProfileTypeIOSAppAdhoc,
			model.AppleProfileTypeIOSAppDevelopment,
			model.AppleProfileTypeMacAppDevelopment,
		}, req.Type) {
			log.Info(ctx, "get apple device information from database")
			platform := model.ApplePlatformIOSDescription
			if req.Type == model.AppleProfileTypeMacAppDevelopment {
				platform = model.ApplePlatformMacOSDescription
			}
			appleDeviceDo := conn.MySQLClient(ctx).AppleDevice
			var appleDevices []*model.AppleDevice
			appleDevices, err = appleDeviceDo.WithContext(ctx).Select(
				appleDeviceDo.InAppleID,
				appleDeviceDo.Udid,
			).Where(
				appleDeviceDo.AppID.Eq(app.ID),
				appleDeviceDo.Status.Eq(model.AppleDeviceStatusOK),
				appleDeviceDo.Platform.Eq(platform),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve apple device information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			deviceIDs, udids = util.ListTo2(appleDevices,
				func(e *model.AppleDevice) (string, string) { return e.InAppleID, e.Udid })
		}
	}

	// 请求苹果服务器申请描述文件。
	var apiResult *appleAPIResponse
	{
		log.Info(ctx, "apply apple profile from apple server")
		var token string
		token, err = generateAppleAPIToken(ctx)
		if err != nil {
			return
		}
		apiResult, err = httpAppleAPIApplyProfile(ctx, token, req.Type, appleBundleID.InAppleID,
			getAppleProfileName(req.Type, req.BundleID, req.Platform), appleCertificate.InAppleID, deviceIDs)
		if err != nil {
			return
		}
	}

	// 解析描述文件。
	var data, pdata []byte
	var p map[string]any
	{
		log.Info(ctx, "parse apple profile information")
		data, err = base64.StdEncoding.DecodeString(apiResult.Data.Attributes.ProfileContent)
		if err != nil {
			log.Error(ctx, "failed to base64 decode apple profile content", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		index := bytes.Index(data, []byte("<?xml"))
		if index == -1 {
			log.Error(ctx, "unknown provision", string(data))
			return errs.New(consts.ErrSystem)
		}
		pdata = data[index:]
		if index = bytes.Index(pdata, []byte("</plist>")); index == -1 {
			log.Error(ctx, "unknown provision", string(pdata))
			return errs.New(consts.ErrSystem)
		}
		pdata = pdata[:index+len("</plist>")]
		_, err = plist.Unmarshal(pdata, &p)
		if err != nil {
			log.Error(ctx, "failed to unmarshal plist data", err, string(pdata))
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 保存描述文件。
	var profileID string
	{
		log.Info(ctx, "save apple profile information to database")
		profileID = strings.ReplaceAll(p["UUID"].(string), "-", "")
		appleProfileTxDo := conn.MySQLTxClient(ctx).AppleProfile
		if err = appleProfileTxDo.WithContext(ctx).Select(
			appleProfileTxDo.ProfileID,
			appleProfileTxDo.AppID,
			appleProfileTxDo.UserID,
			appleProfileTxDo.CertificateID,
			appleProfileTxDo.BundleID,
			appleProfileTxDo.InAppleID,
			appleProfileTxDo.Text,
			appleProfileTxDo.Type,
			appleProfileTxDo.NotBefore,
			appleProfileTxDo.NotAfter,
			appleProfileTxDo.Content,
			appleProfileTxDo.CreatedTime,
		).Create(&model.AppleProfile{
			ProfileID:     profileID,
			AppID:         app.ID,
			UserID:        user.ID,
			CertificateID: appleCertificate.ID,
			InAppleID:     apiResult.Data.ID,
			BundleID:      appleBundleID.ID,
			Text:          string(pdata),
			Type:          req.Type,
			NotBefore:     p["CreationDate"].(time.Time),
			NotAfter:      p["ExpirationDate"].(time.Time),
			Content:       data,
			CreatedTime:   now,
		}); err != nil {
			log.Error(ctx, "failed to save apple profile information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存事件信息。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyProvision,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventProvisionType: req.Type,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"certificateId":   appleCertificate.CertificateID,
				"certificateType": certificateType,
				"profileId":       profileID,
				"udids":           udids,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebApplyInHouseProfile 申请企业内测描述文件。
func AppleWebApplyInHouseProfile(ctx context.Context, req *protocol.AppleWebApplyInHouseProfileReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取 Bundle ID 信息，判断 Bundle ID 属于应用，类型正确。
	var appleBundleID int
	{
		log.Info(ctx, "get apple bundle id from database")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.ID,
		).Where(
			appleBundleIDDo.BundleID.Eq(req.BundleID),
			appleBundleIDDo.AppID.Eq(app.ID),
			appleBundleIDDo.Environment.Eq(model.AppleBundleIDTypeInHouse),
		).Scan(&appleBundleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "apple bundle id not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 调用 fastlane 申请。
	var apiResult *fp.ApplyInHouseProfileRsp
	{
		log.Info(ctx, "request fastlane to apply apple profile")
		apiResult, err = httpFastlaneApplyInHouseProfile(ctx, req.BundleID)
		if err != nil {
			return
		}
	}

	// 解析描述文件。
	var data, pdata []byte
	var p map[string]any
	{
		log.Info(ctx, "parse apple profile")
		data, err = base64.StdEncoding.DecodeString(apiResult.Profile)
		if err != nil {
			log.Error(ctx, "failed to base64 decode apple profile content", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		index := bytes.Index(data, []byte("<?xml"))
		if index == -1 {
			log.Error(ctx, "unknown provision", string(data))
			err = errs.New(consts.ErrSystem)
			return
		}
		pdata = data[index:]
		if index = bytes.Index(pdata, []byte("</plist>")); index == -1 {
			log.Error(ctx, "unknown provision", string(pdata))
			err = errs.New(consts.ErrSystem)
			return
		}
		pdata = pdata[:index+len("</plist>")]
		_, err = plist.Unmarshal(pdata, &p)
		if err != nil {
			log.Error(ctx, "failed to unmarshal plist data", err, string(pdata))
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查找签名证书。
	var certificateID int
	{
		log.Info(ctx, "find apple certificate")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.ID,
		).Where(
			appleCertificateDo.InAppleID.Eq(apiResult.CertificateID),
			appleCertificateDo.DeletedTime.IsNull(),
			appleCertificateDo.AppID.IsNull(),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategorySigning),
		).Limit(1).Scan(&certificateID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve apple certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存描述文件。
	var now time.Time
	var profileID string
	{
		log.Info(ctx, "save apple profile to database")
		now = time.Now()
		profileID = strings.ReplaceAll(p["UUID"].(string), "-", "")
		appleProfileTxDo := conn.MySQLTxClient(ctx).AppleProfile
		if err = appleProfileTxDo.WithContext(ctx).Select(
			appleProfileTxDo.ProfileID,
			appleProfileTxDo.CertificateID,
			appleProfileTxDo.AppID,
			appleProfileTxDo.UserID,
			appleProfileTxDo.BundleID,
			appleProfileTxDo.Text,
			appleProfileTxDo.Type,
			appleProfileTxDo.NotBefore,
			appleProfileTxDo.NotAfter,
			appleProfileTxDo.Content,
			appleProfileTxDo.InAppleID,
			appleProfileTxDo.CreatedTime,
		).Create(&model.AppleProfile{
			ProfileID:     profileID,
			AppID:         app.ID,
			UserID:        user.ID,
			CertificateID: certificateID,
			InAppleID:     apiResult.ID,
			BundleID:      appleBundleID,
			Text:          string(pdata),
			Type:          apiResult.Type,
			NotBefore:     p["CreationDate"].(time.Time),
			NotAfter:      p["ExpirationDate"].(time.Time),
			Content:       data,
			CreatedTime:   now,
		}); err != nil {
			log.Error(ctx, "failed to save apple profile to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存事件信息。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyProvision,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventProvisionType: apiResult.Type,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"profileId": profileID,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebApplyCommonProfile 申请企业内测通配符描述文件。
func AppleWebApplyCommonProfile(ctx context.Context) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 解析描述文件。
	var data, pdata []byte
	var p map[string]any
	{
		log.Info(ctx, "parse apple profile")
		data, err = base64.StdEncoding.DecodeString(cfg.Get().Apple().CommonProfile())
		if err != nil {
			log.Error(ctx, "failed to debase64 apple profile", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		index := bytes.Index(data, []byte("<?xml"))
		if index == -1 {
			log.Error(ctx, "unknown provision", string(data))
			err = errs.New(consts.ErrSystem)
			return
		}
		pdata = data[index:]
		if index = bytes.Index(pdata, []byte("</plist>")); index == -1 {
			log.Error(ctx, "unknown provision:", string(pdata))
			err = errs.New(consts.ErrSystem)
			return
		}
		pdata = pdata[:index+len("</plist>")]
		_, err = plist.Unmarshal(pdata, &p)
		if err != nil {
			log.Error(ctx, err, string(pdata))
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查找签名证书。
	var certificateID int
	{
		log.Info(ctx, "find apple certificate")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.ID,
		).Where(
			appleCertificateDo.InAppleID.Eq(cfg.Get().Apple().CertificateIDOfCommonProfile()),
			appleCertificateDo.DeletedTime.IsNull(),
			appleCertificateDo.AppID.IsNull(),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategorySigning),
		).Limit(1).Scan(&certificateID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve apple certificate information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存描述文件。
	var now time.Time
	var profileID string
	{
		log.Info(ctx, "save apple profile to database")
		now = time.Now()
		profileID = strings.ReplaceAll(p["UUID"].(string), "-", "")
		appleProfileTxDo := conn.MySQLTxClient(ctx).AppleProfile
		if err = appleProfileTxDo.WithContext(ctx).Select(
			appleProfileTxDo.ProfileID,
			appleProfileTxDo.CertificateID,
			appleProfileTxDo.InAppleID,
			appleProfileTxDo.AppID,
			appleProfileTxDo.UserID,
			appleProfileTxDo.Type,
			appleProfileTxDo.Text,
			appleProfileTxDo.NotBefore,
			appleProfileTxDo.NotAfter,
			appleProfileTxDo.Content,
			appleProfileTxDo.CreatedTime,
		).Create(&model.AppleProfile{
			ProfileID:     profileID,
			AppID:         app.ID,
			UserID:        user.ID,
			Text:          string(pdata),
			Type:          model.AppleProfileTypeIOSAppInHouse,
			NotBefore:     p["CreationDate"].(time.Time),
			NotAfter:      p["ExpirationDate"].(time.Time),
			Content:       data,
			CreatedTime:   now,
			InAppleID:     cfg.Get().Apple().CommonProfileID(),
			CertificateID: certificateID,
		}); err != nil {
			if errs.IsMySQLError(err, mysql.ErrDupEntry) {
				log.Warn(ctx, "apple common enterprise profile already exists", profileID)
				// 描述文件已存在。返回成功。
				err = nil
				return
			}
			log.Error(ctx, "failed to save apple profile to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存事件信息。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyProvision,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventProvisionType: model.AppleProfileTypeIOSAppInHouse,
			EventAppleBundleID: consts.AppleWildcardBundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"profileId": profileID,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebApplyPushCertificate 申请 Push 证书。
func AppleWebApplyPushCertificate(ctx context.Context, req *protocol.AppleWebApplyPushCertificateReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取 Bundle ID 信息。
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle id information from database")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		appleBundleID, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.Environment,
			appleBundleIDDo.ID,
		).Where(
			appleBundleIDDo.BundleID.Eq(req.BundleID),
			appleBundleIDDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "app apple bundle id not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 调用 Fastlane 服务。
	var apiResult *fp.ApplyPushCertificateRsp
	var password string
	{
		log.Info(ctx, "request fastlane to apply certificate")
		password = util.RandomPrintableASCIINoSpaceString(consts.AppleCertificatePasswordLength)
		apiResult, err = httpFastlaneApplyPushCert(ctx, &fp.ApplyPushCertificateReq{
			BundleID:    req.BundleID,
			Environment: req.Environment,
			Type:        appleBundleID.Environment,
			Password:    password,
		})
		if err != nil {
			return
		}
	}

	// 解析证书信息。
	var certificateData []byte
	var appleCertificate *model.AppleCertificate
	var now time.Time
	{
		log.Info(ctx, "parse certificate")
		certificateData, err = base64.StdEncoding.DecodeString(apiResult.Certificate)
		if err != nil {
			log.Error(ctx, "failed to base64 decode certificate content", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		var certificateInfo *x509.Certificate
		certificateInfo, err = x509.ParseCertificate(certificateData)
		if err != nil {
			log.Error(ctx, "failed to parse certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		now = time.Now()
		var certificateID string
		certificateID, err = generateID(ctx, IDAppleCertificate)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDAppleCertificate, certificateID),
					"failed to reclaim apple push certificate id")
			}
		}()
		appleCertificate = &model.AppleCertificate{
			CertificateID:      certificateID,
			AppID:              app.ID,
			UserID:             user.ID,
			InAppleID:          apiResult.ID,
			Environment:        req.Environment,
			BundleID:           appleBundleID.ID,
			Category:           model.AppleCertificateCategoryPush,
			Password:           password,
			Publisher:          certificateInfo.Issuer.String(),
			Owner:              certificateInfo.Subject.String(),
			SignatureAlgorithm: certificateInfo.SignatureAlgorithm.String(),
			PublicKeyAlgorithm: certificateInfo.PublicKeyAlgorithm.String(),
			NotBefore:          certificateInfo.NotBefore,
			NotAfter:           certificateInfo.NotAfter,
			CreatedTime:        now,
		}
	}

	// 加密证书。
	{
		log.Info(ctx, "encrypt certificate")
		var aesKey *model.AesKey
		aesKey, err = getLastAESSecret(ctx)
		if err != nil {
			return
		}
		appleCertificate.AesKeyID = aesKey.ID
		appleCertificate.Content, err = util.AESCBCEncrypt(aesKey.Secret, certificateData)
		if err != nil {
			log.Error(ctx, "failed to encrypt certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 入库，保存证书。
	{
		log.Info(ctx, "save apple push certificate to database")
		appleCertificateTxDo := conn.MySQLTxClient(ctx).AppleCertificate
		err = appleCertificateTxDo.WithContext(ctx).Select(
			appleCertificateTxDo.CertificateID,
			appleCertificateTxDo.AppID,
			appleCertificateTxDo.InAppleID,
			appleCertificateTxDo.UserID,
			appleCertificateTxDo.Environment,
			appleCertificateTxDo.Category,
			appleCertificateTxDo.BundleID,
			appleCertificateTxDo.Password,
			appleCertificateTxDo.Publisher,
			appleCertificateTxDo.Owner,
			appleCertificateTxDo.SignatureAlgorithm,
			appleCertificateTxDo.PublicKeyAlgorithm,
			appleCertificateTxDo.NotBefore,
			appleCertificateTxDo.NotAfter,
			appleCertificateTxDo.CreatedTime,
			appleCertificateTxDo.AesKeyID,
			appleCertificateTxDo.Content,
		).Create(appleCertificate)
		if err != nil {
			log.Error(ctx, "failed to save apple push certificate to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存事件信息。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyPushCertificate,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"certificateId": appleCertificate.CertificateID,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebDeleteBundleID 删除 Bundle ID。
func AppleWebDeleteBundleID(ctx context.Context, req *protocol.AppleWebDeleteBundleIDReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，确保 Bundle ID 存在，属于应用。
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle id information from database")
		appleBundleIDDo := conn.MySQLTxClient(ctx).AppleBundleID
		appleBundleID, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.ID,
			appleBundleIDDo.BundleID,
			appleBundleIDDo.InAppleID,
			appleBundleIDDo.Environment,
		).Clauses(query.ForUpdate()).Where(
			appleBundleIDDo.BundleID.Eq(req.BundleID),
			appleBundleIDDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "apple bundle id not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve apple bundle id information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 确保没有证书和描述文件使用 Bundle ID。
	{
		log.Info(ctx, "check that apple bundle is not being used")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		var count int64
		count, err = appleProfileDo.WithContext(ctx).Where(
			appleProfileDo.BundleID.Eq(appleBundleID.ID),
			appleProfileDo.DeletedTime.IsNull(),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count apple profile in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			log.Warn(ctx, "apple profile exists")
			err = errs.New(consts.ErrAppleBundleIDIsUsing)
			return
		}
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		count, err = appleCertificateDo.WithContext(ctx).Where(
			appleCertificateDo.BundleID.Eq(appleBundleID.ID),
			appleCertificateDo.DeletedTime.IsNull(),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategoryPush),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count apple push certificate in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			log.Warn(ctx, "push certificate exists")
			err = errs.New(consts.ErrAppleBundleIDIsUsing)
			return
		}
	}

	// 根据不同类型，删除 Bundle ID。
	{
		log.Info(ctx, "delete apple bundle id in apple")
		switch appleBundleID.Environment {
		case model.AppleBundleIDTypeInHouse:
			if err = httpFastlaneDeleteInHouseBundleID(ctx, appleBundleID.BundleID); err != nil {
				return
			}
		case model.AppleBundleIDTypeAppStore:
			var token string
			token, err = generateAppleAPIToken(ctx)
			if err != nil {
				return
			}
			if err = httpAppleAPIDeleteBundleID(ctx, token, appleBundleID.InAppleID); err != nil {
				return
			}
		}
	}

	// 删除数据库中的 Bundle ID。
	{
		log.Info(ctx, "delete apple bundle id in database")
		appleBundleIDTxDo := conn.MySQLTxClient(ctx).AppleBundleID
		_, err = appleBundleIDTxDo.WithContext(ctx).Where(appleBundleIDTxDo.ID.Eq(appleBundleID.ID)).Delete()
		if err != nil {
			log.Error(ctx, "failed to delete apple bundle id in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemoveAppleBundleID,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventAppleBundleID: req.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"id":          appleBundleID.InAppleID,
				"environment": model.AllAppleBundleIDDescriptions[appleBundleID.Environment],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebRemoveCertificate 删除苹果证书。
func AppleWebRemoveCertificate(ctx context.Context, req *protocol.AppleWebRemoveCertificateReq) (err error) {
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

	// 软删除证书信息。
	{
		log.Info(ctx, "delete apple certificate information")
		appleCertificateTxDo := conn.MySQLTxClient(ctx).AppleCertificate
		var sqlResult gen.ResultInfo
		sqlResult, err = appleCertificateTxDo.WithContext(ctx).Where(
			appleCertificateTxDo.CertificateID.Eq(req.CertificateID),
			appleCertificateTxDo.Category.Eq(model.AppleCertificateCategorySigning),
			appleCertificateTxDo.DeletedTime.IsNull(),
		).UpdateColumnSimple(
			appleCertificateTxDo.DeletedTime.Value(time.Now()),
		)
		if err != nil {
			log.Error(ctx, "failed to update apple certificate information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "update apple certificate information without rows affected")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	return
}

// AppleWebListAppCertificates 获取应用证书和描述文件列表。
func AppleWebListAppCertificates(ctx context.Context) (rsp *protocol.AppleWebListAppCertificatesRsp, err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，获取应用描述文件。
	var appleProfiles []*model.AppleProfile
	var userIDs, bundleIDs []int
	{
		log.Info(ctx, "get apple profiles")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		appleProfiles, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.BundleID,
			appleProfileDo.UserID,
			appleProfileDo.CertificateID,
			appleProfileDo.ProfileID,
			appleProfileDo.InAppleID,
			appleProfileDo.Type,
			appleProfileDo.Content,
			appleProfileDo.CreatedTime,
			appleProfileDo.NotAfter,
		).Where(
			appleProfileDo.AppID.Eq(app.ID),
			appleProfileDo.DeletedTime.IsNull(),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple profiles from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userIDs = util.ListTo(appleProfiles, func(e *model.AppleProfile) int {
			return e.UserID
		})
		bundleIDs = util.ListTo(appleProfiles, func(e *model.AppleProfile) int { return e.BundleID })
	}

	// 查库，获取应用的 Push 证书。
	var applePushCertificates []*model.AppleCertificate
	{
		log.Info(ctx, "get apple push certificates")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		applePushCertificates, err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.BundleID,
			appleCertificateDo.UserID,
			appleCertificateDo.CertificateID,
			appleCertificateDo.InAppleID,
			appleCertificateDo.Owner,
			appleCertificateDo.Environment,
			appleCertificateDo.NotAfter,
			appleCertificateDo.CreatedTime,
		).Where(
			appleCertificateDo.AppID.Eq(app.ID),
			appleCertificateDo.DeletedTime.IsNull(),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategoryPush),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple push certificates from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(appleProfiles) <= 0 && len(applePushCertificates) <= 0 {
			return
		}
		userIDs = append(userIDs, util.ListTo(applePushCertificates,
			func(e *model.AppleCertificate) int { return e.UserID })...)
		bundleIDs = append(bundleIDs, util.ListTo(applePushCertificates,
			func(e *model.AppleCertificate) int { return e.BundleID })...)
	}

	// 查库，获取描述文件关联的证书信息。
	var appleSigningCertificates []*model.AppleCertificate
	{
		certificateIDs := util.CleanNumbers(util.ListTo(appleProfiles,
			func(e *model.AppleProfile) int { return e.CertificateID }))
		if len(certificateIDs) > 0 {
			log.Info(ctx, "get apple signing certificates")
			appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
			appleSigningCertificates, err = appleCertificateDo.WithContext(ctx).Select(
				appleCertificateDo.CertificateID,
				appleCertificateDo.InAppleID,
				appleCertificateDo.UserID,
				appleCertificateDo.Environment,
				appleCertificateDo.Owner,
				appleCertificateDo.Type,
				appleCertificateDo.NotAfter,
				appleCertificateDo.CreatedTime,
			).Where(
				appleCertificateDo.ID.In(certificateIDs...),
				appleCertificateDo.AppID.IsNull(),
				appleCertificateDo.Category.Eq(model.AppleCertificateCategorySigning),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve apple signing certificates from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			userIDs = append(userIDs, util.ListTo(appleSigningCertificates,
				func(e *model.AppleCertificate) int { return e.UserID })...)
		}
	}

	// 查库，获取 Bundle ID 信息。
	var bundleIDToBundle map[int]string
	{
		log.Info(ctx, "get apple bundle information")
		bundleIDs = util.CleanNumbers(bundleIDs)
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		var appleBundleIDs []*model.AppleBundleID
		appleBundleIDs, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.BundleID,
			appleBundleIDDo.ID,
		).Where(
			appleBundleIDDo.ID.In(bundleIDs...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apple bundles from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		bundleIDToBundle = util.ListToMap(appleBundleIDs,
			func(e *model.AppleBundleID) (int, string) { return e.ID, e.BundleID })
	}

	// 查询用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDs = util.CleanNumbers(userIDs)
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AppleWebListAppCertificatesItem, 0, len(appleProfiles)+len(applePushCertificates))
		for _, v := range appleProfiles {
			list = append(list, &protocol.AppleWebListAppCertificatesItem{
				ID:             v.ProfileID,
				AppleID:        v.InAppleID,
				Category:       protocol.AppleFileTypeProfile,
				Type:           v.Type,
				BundleID:       bundleIDToBundle[v.BundleID],
				User:           userIDToName[v.UserID],
				CreatedTime:    formatTime(&v.CreatedTime),
				ExpirationTime: formatTime(&v.NotAfter),
				ProfileContent: string(v.Content),
			})
		}
		for _, v := range applePushCertificates {
			list = append(list, &protocol.AppleWebListAppCertificatesItem{
				ID:               v.CertificateID,
				AppleID:          v.InAppleID,
				Environment:      v.Environment,
				Category:         protocol.AppleFileTypePushCertificate,
				BundleID:         bundleIDToBundle[v.BundleID],
				CertificateOwner: v.Owner,
				User:             userIDToName[v.UserID],
				CreatedTime:      formatTime(&v.CreatedTime),
				ExpirationTime:   formatTime(&v.NotAfter),
			})
		}
		for _, v := range appleSigningCertificates {
			list = append(list, &protocol.AppleWebListAppCertificatesItem{
				ID:               v.CertificateID,
				AppleID:          v.InAppleID,
				Environment:      v.Environment,
				Category:         protocol.AppleFileTypeSigningCertificate,
				Type:             v.Type,
				CertificateOwner: v.Owner,
				User:             userIDToName[v.UserID],
				CreatedTime:      formatTime(&v.CreatedTime),
				ExpirationTime:   formatTime(&v.NotAfter),
			})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedTime > list[j].CreatedTime })
		rsp = &protocol.AppleWebListAppCertificatesRsp{List: list}
	}

	return
}

// AppleWebSubmitSigningJob 提交签名任务。
func AppleWebSubmitSigningJob(ctx context.Context, req *protocol.AppleWebSubmitSigningJobReq) (err error) {
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

	// 校验应用状态。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
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
				log.Warn(ctx, "apple profile not found", req.ProfileID)
				err = errs.New(consts.ErrAppleProfileNotFound)
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
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
				log.Error(ctx, "failed to retrieve apple bundle information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 获取文件信息。
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		var fileInfo *model.File
		fileInfo, err = fileDo.WithContext(ctx).Select(
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
		if fileInfo.AppID != app.ID || fileInfo.UserID != user.ID || fileInfo.Type != model.FileTypeAppleSigning {
			log.Warn(ctx, "file does not belong to apple signing job", fileInfo)
			err = errs.New(consts.ErrParameterInvalid)
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
		appleSigningJobTxDo := conn.MySQLTxClient(ctx).AppleSigningJob.Table(model.GetAppleSigningJobByID(jobID))
		err = appleSigningJobTxDo.WithContext(ctx).Select(
			appleSigningJobTxDo.JobID,
			appleSigningJobTxDo.AppID,
			appleSigningJobTxDo.UserID,
			appleSigningJobTxDo.ProfileID,
			appleSigningJobTxDo.FileID,
			appleSigningJobTxDo.Source,
			appleSigningJobTxDo.Status,
			appleSigningJobTxDo.CreatedTime,
		).Create(&model.AppleSigningJob{
			JobID:       jobID,
			AppID:       app.ID,
			UserID:      user.ID,
			ProfileID:   appleProfile.ID,
			FileID:      req.FileID,
			Source:      model.SourceWeb,
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

	return
}

// AppleWebListSigningJobs 获取签名任务信息列表。
func AppleWebListSigningJobs(ctx context.Context, req *protocol.AppleWebListSigningJobsReq) (
	rsp *protocol.AppleWebListSigningJobsRsp, err error) {

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

	// 校验应用状态。
	{
		log.Info(ctx, "verify app")
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 获取签名任务信息。
	var appleSigningJobs []*model.AppleSigningJob
	var count int
	{
		log.Info(ctx, "get apple signing jobs")
		var tableNames []string
		tableNames, err = getAllAppleSignJobTableNames(ctx)
		if err != nil {
			return
		}
		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob
		count, err = appleSigningJobDo.WithContext(ctx).Count2(tableNames, app.ID)
		if err != nil {
			log.Error(ctx, "failed to count apple signing jobs", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count <= 0 {
			return
		}
		appleSigningJobs, err = appleSigningJobDo.WithContext(ctx).List(
			tableNames, app.ID, req.PageSize, (req.PageNumber-1)*req.PageSize)
		if err != nil {
			log.Error(ctx, "failed to list apple signing jobs", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(appleSigningJobs) <= 0 {
			rsp = &protocol.AppleWebListSigningJobsRsp{Count: count}
			return
		}
	}

	// 获取描述文件信息。
	var bundleIDs, userIDs []int
	var fileIDs []string
	var profileIDtoBundleID map[int]int
	var profileIDtoID map[int]string
	{
		log.Info(ctx, "get apple profile information")
		profileIDs := make([]int, len(appleSigningJobs))
		userIDs = make([]int, len(appleSigningJobs))
		fileIDs = make([]string, 0, len(appleSigningJobs)*2)
		for i, v := range appleSigningJobs {
			profileIDs[i] = v.ProfileID
			userIDs[i] = v.UserID
			fileIDs = append(fileIDs, v.FileID, v.SignedFileID)
		}
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		var appleProfiles []*model.AppleProfile
		appleProfiles, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.ID,
			appleProfileDo.ProfileID,
			appleProfileDo.BundleID,
		).Where(
			appleProfileDo.ID.In(util.CleanNumbers(profileIDs)...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to list apple profiles information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		profileIDtoID = make(map[int]string, len(appleProfiles))
		profileIDtoBundleID = make(map[int]int, len(appleProfiles))
		bundleIDs = make([]int, 0, len(appleProfiles))
		for _, v := range appleProfiles {
			profileIDtoID[v.ID] = v.ProfileID
			profileIDtoBundleID[v.ID] = v.BundleID
			bundleIDs = append(bundleIDs, v.BundleID)
		}
	}

	// 获取 Bundle ID。
	var bundleIDToID map[int]string
	{
		log.Info(ctx, "get apple bundle id information")
		bundleIDs = util.CleanNumbers(bundleIDs)
		if len(bundleIDs) > 0 {
			appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
			var appleBundleIDInfos []*model.AppleBundleID
			appleBundleIDInfos, err = appleBundleIDDo.WithContext(ctx).Select(
				appleBundleIDDo.ID,
				appleBundleIDDo.BundleID,
			).Where(
				appleBundleIDDo.ID.In(bundleIDs...),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to get apple bundle id information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			bundleIDToID = util.ListToMap(appleBundleIDInfos,
				func(e *model.AppleBundleID) (int, string) { return e.ID, e.BundleID })
		} else {
			bundleIDToID = make(map[int]string)
		}
	}

	// 获取文件信息。
	var fileIDToName map[string]string
	{
		log.Info(ctx, "get file information")
		fileIDs = util.CleanStrings(fileIDs)
		fileIDToName, err = GetFileNamesByIDs(ctx, fileIDs)
		if err != nil {
			return
		}
	}

	// 获取用户信息。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get user information")
		userIDs = util.CleanNumbers(userIDs)
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AppleWebListSigningJobsItem, len(appleSigningJobs))
		for i, v := range appleSigningJobs {
			list[i] = &protocol.AppleWebListSigningJobsItem{
				JobID:          v.JobID,
				BundleID:       bundleIDToID[profileIDtoBundleID[v.ProfileID]],
				ProfileID:      profileIDtoID[v.ProfileID],
				FileID:         v.FileID,
				FileName:       fileIDToName[v.FileID],
				UserName:       userIDToName[v.UserID],
				CreatedTime:    formatTime(&v.CreatedTime),
				FinishedTime:   formatTime(&v.FinishedTime),
				Status:         v.Status,
				Log:            v.Log,
				SignedFileID:   v.SignedFileID,
				SignedFileName: fileIDToName[v.SignedFileID],
			}
		}
		rsp = &protocol.AppleWebListSigningJobsRsp{Count: count, List: list}
	}

	return
}

// AppleWebRemoveProfile 删除描述文件。
func AppleWebRemoveProfile(ctx context.Context, req *protocol.AppleWebRemoveProfileReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，判断应用是否具有该描述文件。
	var appleProfile *model.AppleProfile
	{
		log.Info(ctx, "get apple profile information")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		appleProfile, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.BundleID,
			appleProfileDo.InAppleID,
			appleProfileDo.Type,
		).Where(
			appleProfileDo.ProfileID.Eq(req.ProfileID),
			appleProfileDo.AppID.Eq(app.ID),
			appleProfileDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			log.Error(ctx, "failed to get apple profile from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if appleProfile == nil {
			log.Warn(ctx, "apple profile not found")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 删除描述文件。
	var now time.Time
	{
		log.Info(ctx, "remove apple profile")
		now = time.Now()
		appleProfileTxDo := conn.MySQLTxClient(ctx).AppleProfile
		var sqlResult gen.ResultInfo
		sqlResult, err = appleProfileTxDo.WithContext(ctx).Where(
			appleProfileTxDo.ProfileID.Eq(req.ProfileID),
			appleProfileTxDo.AppID.Eq(app.ID),
			appleProfileTxDo.DeletedTime.IsNull(),
		).UpdateColumnSimple(
			appleProfileTxDo.DeletedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update apple profile from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "delete apple profile no effect")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	// 获取 Bundle ID。
	var bundleID string
	{
		log.Info(ctx, "get apple bundle id")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.BundleID,
		).Where(
			appleBundleIDDo.ID.Eq(appleProfile.BundleID),
		).Scan(&bundleID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if appleProfile.Type == model.AppleProfileTypeIOSAppInHouse && len(bundleID) <= 0 {
			bundleID = consts.AppleWildcardBundleID
		}
	}

	// 删除描述文件。
	{
		if bundleID != consts.AppleWildcardBundleID {
			log.Info(ctx, "remove apple profile in remote")
			if appleProfile.Type == model.AppleProfileTypeIOSAppInHouse {
				err = httpFastlaneRemoveInHouseProfile(ctx, appleProfile.InAppleID)
			} else {
				var token string
				token, err = generateAppleAPIToken(ctx)
				if err != nil {
					return
				}
				err = httpAppleAPIRemoveProfile(ctx, token, appleProfile.InAppleID)
			}
			if err != nil {
				return
			}
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemoveProvision,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventProvisionType: appleProfile.Type,
			EventAppleBundleID: bundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"id": req.ProfileID,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebRemovePushCertificate 删除 Push 证书文件。
func AppleWebRemovePushCertificate(ctx context.Context, req *protocol.AppleWebRemovePushCertificateReq) (err error) {
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 查库，判断应用是否具有该 Push 证书。
	var appleCertificate *model.AppleCertificate
	{
		log.Info(ctx, "get apple push certificate information")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		appleCertificate, err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.InAppleID,
			appleCertificateDo.BundleID,
			appleCertificateDo.Environment,
		).Where(
			appleCertificateDo.CertificateID.Eq(req.CertificateID),
			appleCertificateDo.AppID.Eq(app.ID),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategoryPush),
			appleCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			log.Error(ctx, "failed to get apple push certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if appleCertificate == nil {
			log.Warn(ctx, "push certificate not found")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 删除 Push 证书。
	var now time.Time
	{
		log.Info(ctx, "remove apple push certificate")
		now = time.Now()
		appleCertificateDo := conn.MySQLTxClient(ctx).AppleCertificate
		var sqlResult gen.ResultInfo
		sqlResult, err = appleCertificateDo.WithContext(ctx).Where(
			appleCertificateDo.CertificateID.Eq(req.CertificateID),
			appleCertificateDo.AppID.Eq(app.ID),
			appleCertificateDo.Category.Eq(model.AppleCertificateCategoryPush),
			appleCertificateDo.DeletedTime.IsNull(),
		).UpdateColumnSimple(
			appleCertificateDo.DeletedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update apple push certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "delete apple push certificate no effect")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	// 获取 Bundle ID。
	var appleBundleID *model.AppleBundleID
	{
		log.Info(ctx, "get apple bundle id")
		appleBundleIDDo := conn.MySQLClient(ctx).AppleBundleID
		appleBundleID, err = appleBundleIDDo.WithContext(ctx).Select(
			appleBundleIDDo.BundleID,
			appleBundleIDDo.Environment,
		).Where(
			appleBundleIDDo.ID.Eq(appleCertificate.BundleID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve apple bundle id information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if appleBundleID == nil {
			appleBundleID = &model.AppleBundleID{}
		}
	}

	// 删除 Push 证书。
	{
		log.Info(ctx, "delete apple push certificate in remote")
		err = httpFastlaneRemovePushCertificate(ctx, appleCertificate.InAppleID, appleBundleID.BundleID,
			appleCertificate.Environment, appleBundleID.Environment)
		if err != nil {
			return
		}
	}

	// 记录操作事件。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemovePushCertificate,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser:          user.NameEn,
			EventApp:           app.Name,
			EventAppleBundleID: appleBundleID.BundleID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"id":          req.CertificateID,
				"environment": model.AllPushCertificateEnvironments[appleCertificate.Environment],
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppleWebDownloadCertificate 下载证书和描述文件。
func AppleWebDownloadCertificate(ctx context.Context, req *protocol.AppleWebDownloadCertificateReq) (
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
		if app.Platform != model.AppPlatformApple {
			log.Warn(ctx, "app platform is not supported")
			err = errs.New(consts.ErrParameterInvalid)
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
					log.Warn(ctx, "apple profile not found")
					err = errs.New(consts.ErrParameterInvalid)
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
					log.Warn(ctx, "apple push certificate not found")
					err = errs.New(consts.ErrParameterInvalid)
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
					log.Warn(ctx, "apple signing certificate not found")
					err = errs.New(consts.ErrParameterInvalid)
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

// AppleWebStatisticSigningTimes 获取应用的 Apple 类型签名次数统计信息。
func AppleWebStatisticSigningTimes(ctx context.Context, req *protocol.AppleWebStatisticSigningTimesReq) (
	rsp *protocol.AppleWebStatisticSigningTimesRsp, err error) {

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
	var sqlResult []map[string]any
	{
		log.Info(ctx, "get times of apple job")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		// 获取表名。
		var tableNames []string
		tableNames, err = filterAppleSigningJobTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) <= 0 {
			rsp = &protocol.AppleWebStatisticSigningTimesRsp{}
			return
		}

		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob
		switch req.TimeStep {
		case protocol.TimeStepDay:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CountWithDay(tableNames, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CountWithWeek(tableNames, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CountWithMonth(tableNames, appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to count apple job from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 转换数据。
	var items []*protocol.AppleWebStatisticSigningTimesItem
	{
		log.Info(ctx, "deal sql data")
		items = make([]*protocol.AppleWebStatisticSigningTimesItem, 0, len(sqlResult)/2)
		item := &protocol.AppleWebStatisticSigningTimesItem{}
		for _, v := range sqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
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
				items = append(items, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.AppleWebStatisticSigningTimesItem{BeginTime: beginTime}
				items = append(items, item)
			}
			item.SigningTimes = count
		}
	}

	rsp = &protocol.AppleWebStatisticSigningTimesRsp{List: items}

	return
}

// AppleWebStatisticSigningCost 获取应用的 Apple 类型签名耗时统计信息。
func AppleWebStatisticSigningCost(ctx context.Context, req *protocol.AppleWebStatisticSigningCostReq) (
	rsp *protocol.AppleWebStatisticSigningCostRsp, err error) {

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
	var sqlResult []map[string]any
	{
		log.Info(ctx, "get cost of apple job")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		var tableNames []string
		tableNames, err = filterAppleSigningJobTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) <= 0 {
			return &protocol.AppleWebStatisticSigningCostRsp{}, nil
		}
		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob
		switch req.TimeStep {
		case protocol.TimeStepDay:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CostWithDay(tableNames, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CostWithWeek(tableNames, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			sqlResult, err = appleSigningJobDo.WithContext(ctx).CostWithMonth(tableNames, appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to count apple job from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 转换数据。
	var items []*protocol.AppleWebStatisticSigningCostItem
	{
		log.Info(ctx, "deal sql data")
		items = make([]*protocol.AppleWebStatisticSigningCostItem, 0, len(sqlResult)/2)
		item := &protocol.AppleWebStatisticSigningCostItem{}
		for _, v := range sqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
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
				items = append(items, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.AppleWebStatisticSigningCostItem{BeginTime: beginTime}
				items = append(items, item)
			}
			item.SigningCost = cost
		}
	}

	rsp = &protocol.AppleWebStatisticSigningCostRsp{List: items}

	return
}

func generateAppleAPIToken(ctx context.Context) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.MapClaims{
		"iss": cfg.Get().AppleAPI().IssuerID(),
		"exp": time.Now().Unix() + int64(19*time.Minute.Seconds()),
		"aud": "appstoreconnect-v1",
		"iat": time.Now().Unix(),
	})
	token.Header["alg"] = "ES256"
	token.Header["kid"] = cfg.Get().AppleAPI().KeyID()
	token.Header["typ"] = "JWT"

	block, _ := pem.Decode([]byte(cfg.Get().AppleAPI().Secret()))
	if block == nil {
		log.Error(ctx, "block is nil")
		return "", errs.New(consts.ErrSystem)
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Error(ctx, "failed to parse pkcs8 key", err)
		return "", errs.NewWithError(consts.ErrSystem, err)
	}

	signedString, err := token.SignedString(privateKey)
	if err != nil {
		log.Error(ctx, "failed to sign token", err)
		return "", errs.NewWithError(consts.ErrSystem, err)
	}

	return "Bearer " + signedString, nil
}

func getAppleProfileName(typ, bundleID, platform string) string {
	return fmt.Sprintf("%s_%s_%s_%s", bundleID, platform, typ, util.FastRandomAlphaNumberString(5))
}

func getAppleCertificateType(platform, typ string) string {
	switch platform {
	case model.ApplePlatformMacOSDescription:
		switch typ {
		case model.AppleProfileTypeMacAppDevelopment:
			return model.AppleCertificateTypeMacAppDevelopment
		case model.AppleProfileTypeMacAppStore:
			return model.AppleCertificateTypeMacAppDistribution
		}
	case model.ApplePlatformUniversalDescription:
		switch typ {
		case model.AppleProfileTypeIOSAppDevelopment, model.AppleProfileTypeMacAppDevelopment:
			return model.AppleCertificateTypeDevelopment
		case model.AppleProfileTypeIOSAppAdhoc, model.AppleProfileTypeIOSAppStore, model.AppleProfileTypeMacAppStore:
			return model.AppleCertificateTypeDistribution
		}
	case model.ApplePlatformIOSDescription:
		switch typ {
		case model.AppleProfileTypeIOSAppDevelopment:
			return model.AppleCertificateTypeIOSDevelopment
		case model.AppleProfileTypeIOSAppStore, model.AppleProfileTypeIOSAppAdhoc:
			return model.AppleCertificateTypeIOSDistribution
		}
	}
	return ""
}

func dealBundleIDCapabilities(newCapabilities, oldCapabilities []string) (map[string]bool, []string) {
	var hasCapabilitySet map[string]struct{}
	if len(oldCapabilities) > 0 {
		hasCapabilitySet = util.ListToMap(oldCapabilities, func(e string) (string, struct{}) { return e, struct{}{} })
	}
	bundleIDCapabilityMap := make(map[string]bool, len(oldCapabilities))
	remainCapabilities := make([]string, 0, len(oldCapabilities))
	for _, v := range newCapabilities {
		_, ok := hasCapabilitySet[v]
		if !ok {
			bundleIDCapabilityMap[v] = true
			continue
		}
		remainCapabilities = append(remainCapabilities, v)
		delete(hasCapabilitySet, v)
	}
	for v := range hasCapabilitySet {
		bundleIDCapabilityMap[v] = false
	}

	return bundleIDCapabilityMap, remainCapabilities
}

func getAllAppleSignJobTableNames(ctx context.Context) ([]string, error) {
	appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob
	tableNames, err := appleSigningJobDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to query database tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return tableNames, nil
}

func filterAppleSigningJobTables(ctx context.Context, begin, end time.Time) ([]string, error) {
	// 从数据库中查处所有任务表。
	appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob
	allTables, err := appleSigningJobDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to retrieve database tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	slices.Sort(allTables)
	if begin.IsZero() && end.IsZero() {
		return allTables, nil
	}

	// 过滤任务表。
	endTable := model.GetAppleSigningJobTableName(end)
	beginTable := model.GetAppleSigningJobTableName(begin)
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

	log.Info(ctx, "gather apple tables", allTables, begin, end)
	return allTables, nil
}

func getPlatformByAppleDevice(device string) string {
	switch device {
	case model.AppleDeviceTypeMac:
		return model.ApplePlatformMacOSDescription
	case model.AppleDeviceTypeIpad,
		model.AppleDeviceTypeIpod,
		model.AppleDeviceTypeIphone,
		model.AppleDeviceTypeAppleTV,
		model.AppleDeviceTypeAppleWatch:
		return model.ApplePlatformIOSDescription
	default:
		return model.ApplePlatformUniversalDescription
	}
}
