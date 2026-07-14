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

package api_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestWindowsWebUploadCertificate(t *testing.T) {
	const reqPath = "/web/windows/uploadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificate := windowsCertificate
		password := windowsCertificatePassword
		// 模拟数据库中的 AES 加密密钥记录。
		mockAesKey := &model.AesKey{CreatedTime: time.Now(), Secret: util.RandomBytes(16)}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)                           // 查询数据库中加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 缓存证书 ID 到 Redis。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.CreateOnce(nil)             // 创建证书记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                       // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebUploadCertificateReq{
				Certificate: certificate,
				Password:    password,
			})),
			consts.AlertSuccess,
		)
	})

	for _, v := range []struct {
		Name        string
		Certificate string
		Password    string
		ErrCode     errs.Code
	}{
		{"证书缺失", "", windowsCertificatePassword, errs.ErrInvalidRequestParameters},
		{"证书非法", "汉", "123456", consts.ErrParameterInvalid},
		{"密码缺失", windowsCertificate, "", errs.ErrInvalidRequestParameters},
		{"密码过长", windowsCertificate, util.FastRandomAlphaNumberString(65), errs.ErrInvalidRequestParameters},
	} {
		validateErrorRequest := func(t *testing.T, certificate, password string, errCode errs.Code) {
			ctx := context.Background()

			defer mvt.Chain(AppInfo).
				Elem().
				FieldByName("Platform").
				Set(model.AppPlatformWindows).
				Reset()

			redisMocker := MockRedis(ctx)
			dbUserMocker := MockDBClient[model.User](ctx)
			dbAppMocker := MockDBClient[model.App](ctx)
			dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
			redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
			redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
			redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
			redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
			dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
			dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
			dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
			defer redisMocker.Reset()
			defer dbUserMocker.Reset()
			defer dbAppMocker.Reset()
			defer dbUserRoleMocker.Reset()

			CheckAndUnmarshalBody[any](
				t,
				ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebUploadCertificateReq{
					Certificate: certificate,
					Password:    password,
				})),
				errCode,
			)
		}

		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Certificate, v.Password, v.ErrCode)
		})
	}
}

func TestWindowsWebListCertificates(t *testing.T) {
	const reqPath = "/web/windows/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟数据库中的 Windows 证书列表数据（空结构体，仅占位）。
		mockCertList := []*model.WindowsCertificate{{}, {}}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                         // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                    // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                        // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                                                   // 校验应用管理员权限。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询已授权证书 IDs。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                                                                                 // 查询上传人英文名。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.FindOnce(mockCertList, nil)                                                     // 批量查询证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})
}

func TestWindowsWebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/windows/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("123"))
		// 模拟数据库中的 Windows 证书记录。
		mockCert := &model.WindowsCertificate{Content: encrypt}
		// 模拟数据库中的 AES 加密密钥记录。
		mockAesKey := &model.AesKey{Secret: secret}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)     // 查询数据库证书记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                       // 添加应用事件到数据库。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                           // 查询数据库解密密钥。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbEventMocker.Reset()
		defer dbAesKeyMocker.Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.WindowsWebDownloadCertificateReq{CertificateID: certificateID}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is %v", rsp.Code)
		}
		bs, _ := io.ReadAll(rsp.Body)
		if len(bs) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebDownloadCertificateReq{
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestWindowsWebAddEVCertificate(t *testing.T) {
	const reqPath = "/web/windows/addEVCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		sha1 := "5aa8907490b5210c828ce2f3089f794b9c7605d3"
		owner := "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN"
		publisher := "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US"
		signatureAlgorithm := "SHA256-RSA"
		publicKeyAlgorithm := "RSA"
		password := "123456"
		machineIP := "127.0.0.1"
		serialNumber := "123"
		version := 119
		notBefore := time.Now()
		notAfter := time.Now().AddDate(1, 0, 0)
		isMicrosoftVerifyCertificate := true
		typ := model.WindowsCertificateTypePersonalEV

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.CountOnce(0, nil)           // 检查相同指纹证书是否已存在。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 缓存 EV 证书 ID 到 Redis。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.CreateOnce(nil)             // 创建 EV 证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebAddEVCertificateReq{
				SHA1:                         sha1,
				Owner:                        owner,
				Publisher:                    publisher,
				SignatureAlgorithm:           signatureAlgorithm,
				PublicKeyAlgorithm:           publicKeyAlgorithm,
				Password:                     password,
				MachineIP:                    machineIP,
				SerialNumber:                 serialNumber,
				Version:                      version,
				NotBefore:                    protocol.Time(notBefore),
				NotAfter:                     protocol.Time(notAfter),
				IsMicrosoftVerifyCertificate: isMicrosoftVerifyCertificate,
				Type:                         typ,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, sha1, owner, publisher, signatureAlgorithm, publicKeyAlgorithm, password, machineIP, serialNumber string, version int, notBefore, notAfter time.Time, isMicrosoftVerifyCertificate bool, typ int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebAddEVCertificateReq{
				SHA1:                         sha1,
				Owner:                        owner,
				Publisher:                    publisher,
				SignatureAlgorithm:           signatureAlgorithm,
				PublicKeyAlgorithm:           publicKeyAlgorithm,
				Password:                     password,
				MachineIP:                    machineIP,
				SerialNumber:                 serialNumber,
				Version:                      version,
				NotBefore:                    protocol.Time(notBefore),
				NotAfter:                     protocol.Time(notAfter),
				IsMicrosoftVerifyCertificate: isMicrosoftVerifyCertificate,
				Type:                         typ,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                                                                              string
		SHA1, Owner, Publisher, SignatureAlgorithm, PublicKeyAlgorithm, Password, MachineIP, SerialNumber string
		Version                                                                                           int
		NotBefore, NotAfter                                                                               time.Time
		IsMicrosoftVerifyCertificate                                                                      bool
		Type                                                                                              int
	}{
		{"指纹缺失", "", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"指纹错误", util.FastRandomAlphaNumberString(39), "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"所有者缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"颁发者缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"签名算法缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "", "RSA", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"公钥算法缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "", "123456", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"密码缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"密码过短", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "1", "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"密码过长", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", util.FastRandomAlphaNumberString(65), "127.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"机器IP错误", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "256.0.0.1", "123", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"序列号缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "", 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"序列号过长", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", util.FastRandomAlphaNumberString(1025), 119, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"版本号缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 0, time.Now(), time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"生效时间缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 1, time.Time{}, time.Now().AddDate(1, 0, 0), true, model.WindowsCertificateTypePersonalEV},
		{"过期时间缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 1, time.Now(), time.Time{}, true, model.WindowsCertificateTypePersonalEV},
		{"过期时间小于生效时间", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 1, time.Now(), time.Now().AddDate(0, 0, -1), true, model.WindowsCertificateTypePersonalEV},
		{"类型缺失", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 1, time.Now(), time.Now().AddDate(0, 0, -1), true, 0},
		{"类型错误", "5aa8907490b5210c828ce2f3089f794b9c7605d3", "CN=ivfzhou,O=ivfzhou,L=Changsha,ST=Hunan,C=CN", "CN=DigiCert SHA2 Assured ID Code Signing CA,OU=www.digicert.com,O=DigiCert Inc,C=US", "SHA256-RSA", "RSA", "123456", "127.0.0.1", "123", 1, time.Now(), time.Now().AddDate(0, 0, -1), true, -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SHA1, v.Owner, v.Publisher, v.SignatureAlgorithm, v.PublicKeyAlgorithm, v.Password, v.MachineIP, v.SerialNumber, v.Version, v.NotBefore, v.NotAfter, v.IsMicrosoftVerifyCertificate, v.Type)
		})
	}
}

func TestWindowsWebUploadCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/uploadCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificate := windowsCertificate
		password := windowsCertificatePassword
		// 模拟数据库中的 AES 加密密钥记录。
		mockAesKey := &model.AesKey{CreatedTime: time.Now(), Secret: util.RandomBytes(16)}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)                           // 查询数据库中加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 缓存公司证书 ID 到 Redis。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.CreateOnce(nil)             // 创建公司证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebUploadCompanyCertificateReq{
				Certificate: certificate,
				Password:    password,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificate, password string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebUploadCompanyCertificateReq{
				Certificate: certificate,
				Password:    password,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		Certificate string
		Password    string
	}{
		{"证书缺失", "", windowsCertificatePassword},
		{"证书格式错误", "汉", windowsCertificatePassword},
		{"密码缺失", windowsCertificate, ""},
		{"密码过长", windowsCertificate, util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Certificate, v.Password)
		})
	}
}

func TestWindowsWebListCompanyCertificates(t *testing.T) {
	const reqPath = "/web/windows/listCompanyCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟数据库中的公司证书列表数据（空结构体，仅占位）。
		mockCertList := []*model.WindowsCertificate{{}, {}}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询上传人英文名。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.FindOnce(mockCertList, nil) // 批量查询公司证书。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})
}

func TestWindowsWebGrantAppEVCertificate(t *testing.T) {
	const reqPath = "/web/windows/grantAppEVCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		certificateID := util.FastRandomAlphaNumberString(32)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                 // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = AppInfo.ID }, nil)                      // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                // 校验系统管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil) // 查询证书记录。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.CreateOnce(nil)    // 创建证书授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebGrantAppEVCertificateReq{
				AppID:         appID,
				CertificateID: certificateID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, appID, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.WindowsWebGrantAppEVCertificateReq{
				AppID:         appID,
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
		AppID         string
	}{
		{"应用ID缺失", "", util.FastRandomAlphaNumberString(32)},
		{"应用ID错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(32)},
		{"应用ID非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(32)},
		{"证书ID缺失", util.FastRandomAlphaNumberString(32), ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AppID, v.CertificateID)
		})
	}
}

func TestWindowsWebGetCertificatePassword(t *testing.T) {
	const reqPath = "/web/windows/getCertificatePassword"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		// 模拟数据库中的 EV 证书记录。
		mockCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV, Password: util.FastRandomAlphaNumberString(6)}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                 // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                 // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                               // 校验应用管理员权限。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.CountOnce(1, nil) // 校验应用是否已获授权。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)                     // 查询 EV 证书密码。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebGetCertificatePasswordRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebGetCertificatePasswordReq{
				CertificateID: certificateID,
			})),
			0,
		)

		if len(rspBodyObj.Data.Password) <= 0 {
			t.Errorf("password is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebGetCertificatePasswordReq{
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestWindowsWebDownloadCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/downloadCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("abc"))
		// 模拟数据库中的公司证书记录。
		mockCert := &model.WindowsCertificate{Content: encrypt}
		// 模拟数据库中的 AES 加密密钥记录。
		mockAesKey := &model.AesKey{Secret: secret}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                        // 查询上传人英文名。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)     // 查询公司证书记录。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                           // 查询数据库解密密钥。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()

		rsp := ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.WindowsWebDownloadCompanyCertificateReq{
			CertificateID: certificateID,
		}))

		if rsp.Code != http.StatusOK {
			t.Errorf("http code is not 200")
		}
		bs, _ := io.ReadAll(rsp.Body)
		if len(bs) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.WindowsWebDownloadCompanyCertificateReq{
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestWindowsWebListGrantCertificateApps(t *testing.T) {
	const reqPath = "/web/windows/listGrantCertificateApps"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		appID := util.FastRandomAlphaNumberString(32)
		pageSize := 1
		pageNumber := 1
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}}
		// 模拟数据库中的应用列表数据（空结构体，仅占位）。
		mockAppList := []*model.App{{}}
		// 模拟数据库中的证书列表数据（空结构体，仅占位）。
		mockCertList := []*model.WindowsCertificate{{}}
		// 模拟数据库中的授权记录列表（空结构体，仅占位）。
		mockAuthList := []*model.WindowsCertificateAuthorization{{}}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                         // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                                                       // 查询上传人英文名。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = AppInfo.ID }, nil)                               // 查询指定应用信息。
		dbAppMocker = dbAppMocker.FindOnce(mockAppList, nil)                                                          // 批量查询已授权应用信息。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)          // 查询证书记录。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.FindOnce(mockCertList, nil)                           // 批量查询证书记录。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.CountOnce(1, nil)           // 统计授权记录总数。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.FindOnce(mockAuthList, nil) // 分页查询授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebListGrantCertificateAppsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.WindowsWebListGrantCertificateAppsReq{
				AppID:         appID,
				CertificateID: certificateID,
				PageSize:      pageSize,
				PageNumber:    pageNumber,
			})),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID, appID string, pageSize, pageNumber int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.WindowsWebListGrantCertificateAppsReq{
				AppID:         appID,
				CertificateID: certificateID,
				PageSize:      pageSize,
				PageNumber:    pageNumber,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
		AppID         string
		PageSize      int
		PageNumber    int
	}{
		{"应用ID错误", util.FastRandomAlphaNumberString(31), "", 1, 1},
		{"应用ID非法", util.FastRandomAlphaNumberString(31) + "汉", "", 1, 1},
		{"证书ID错误", "", util.FastRandomAlphaNumberString(31), 1, 1},
		{"证书ID非法", "", util.FastRandomAlphaNumberString(31) + "汉", 1, 1},
		{"页码非法", "", "", 0, 1},
		{"页条数非法", "", "", 1, 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID, v.AppID, v.PageSize, v.PageNumber)
		})
	}
}

func TestWindowsWebSubmitSigningJob(t *testing.T) {
	const reqPath = "/web/windows/submitSigningJob"

	t.Run("正常测试_提交PE签名任务", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		certificateID := util.FastRandomAlphaNumberString(32)
		// 模拟数据库中的 OV 证书记录。
		mockCert := &model.WindowsCertificate{
			Type:  model.WindowsCertificateTypePersonalOV,
			AppID: AppInfo.ID,
		}
		// 模拟数据库中的文件记录。
		mockFile := &model.File{
			Type:   model.FileTypeWindowsSigning,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   "test.dll",
			TusdID: util.FastRandomAlphaNumberString(38),
		}
		// PE 文件二进制数据。
		peBinary := GenerateMinimalPE()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbWindowsSigningJobMocker := MockDBClient[model.WindowsSigningJob](ctx)
		tusdMocker := MockTusdClient(ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)     // 查询证书记录。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                                 // 查询文件信息。
		tusdMocker = tusdMocker.DownloadToFileOnce(peBinary, nil)                           // 下载 PE 文件校验格式。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 生成任务 ID。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.CreateOnce(nil)               // 保存签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                         // 发送签名任务消息到消息队列。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbWindowsSigningJobMocker.Reset()
		defer tusdMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebSubmitSigningJobReq{
				SigningType:   model.WindowsSigningJobTypePE,
				CertificateID: certificateID,
				FileID:        fileID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signingType int, certificateID, fileID string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebSubmitSigningJobReq{
				SigningType:   signingType,
				CertificateID: certificateID,
				FileID:        fileID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		SigningType   int
		CertificateID string
		FileID        string
	}{
		{"签名类型缺失", 0, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"签名类型错误", 4, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"文件 ID 缺失", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 过长", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(39)},
		{"文件 ID 非法", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
		{"证书 ID 错误", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SigningType, v.CertificateID, v.FileID)
		})
	}
}

func TestWindowsWebListSigningJobs(t *testing.T) {
	const reqPath = "/web/windows/listSigningJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		pageNumber := 1
		pageSize := 10
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟签名任务表名列表。
		mockTableNames := []string{"~"}
		// 模拟数据库中的签名任务列表数据（空结构体，仅占位）。
		mockSigningJobList := []*model.WindowsSigningJob{
			{Source: model.SourceWeb, UserID: LoginUser.ID, CertificateID: 1, FileID: util.FastRandomAlphaNumberString(38)},
			{Source: model.SourceWeb, UserID: LoginUser.ID, CertificateID: 2, FileID: util.FastRandomAlphaNumberString(38)},
		}
		// 模拟数据库中的证书列表数据（空结构体，仅占位）。
		mockCertList := []*model.WindowsCertificate{{}, {}}
		// 模拟数据库中的文件列表数据（空结构体，仅占位）。
		mockFileList := []*model.File{{}, {}}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsSigningJobMocker := MockDBClient[model.WindowsSigningJob](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                       // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                      // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                          // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                     // 校验应用管理员权限。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.WindowsSigningJobGetTablesOnce(mockTableNames, nil) // 获取签名任务表名。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.WindowsSigningJobCount2Once(2, nil)                 // 统计签名任务总数。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.WindowsSigningJobListOnce(mockSigningJobList, nil)  // 分页查询签名任务记录。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                                                   // 查询用户英文名。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.FindOnce(mockCertList, nil)                       // 查询证书信息。
		dbFileMocker = dbFileMocker.FindOnce(mockFileList, nil)                                                   // 查询文件信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsSigningJobMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.WindowsWebListSigningJobsReq{
					PageNumber: pageNumber,
					PageSize:   pageSize,
				},
			)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
		if rspBodyObj.Data.Count <= 0 {
			t.Errorf("count is zero")
		}
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize, signingType, status int) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.WindowsWebListSigningJobsReq{
					PageNumber:  pageNumber,
					PageSize:    pageSize,
					SigningType: signingType,
					Status:      status,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageNumber, PageSize int
		SigningType, Status  int
	}{
		{"页数缺失", 0, 10, 0, 0},
		{"每页条数缺失", 1, 0, 0, 0},
		{"每页条数过大", 1, 101, 0, 0},
		{"签名类型非法", 1, 10, 5, 0},
		{"任务状态非法", 1, 10, 0, 8},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageNumber, v.PageSize, v.SigningType, v.Status)
		})
	}
}

func TestWindowsWebSubmitWHQLJob(t *testing.T) {
	const reqPath = "/web/windows/submitWHQLJob"

	t.Run("正常测试_提交HLK测试任务", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		// 模拟数据库中的文件记录（.sys 类型）。
		mockFile := &model.File{
			Type:   model.FileTypeWindowsSigning,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   "test.sys",
			TusdID: util.FastRandomAlphaNumberString(38),
		}
		// PE 文件二进制数据。
		peBinary := GenerateMinimalPE()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbWhqlJobMocker := MockDBClient[model.WhqlJob](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                                 // 查询文件信息。
		tusdMocker = tusdMocker.DownloadToFileOnce(peBinary, nil)                           // 下载 .sys 文件校验 PE 格式。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 生成 WHQL 任务 ID。
		dbWhqlJobMocker = dbWhqlJobMocker.CreateOnce(nil)                                   // 保存 WHQL 任务到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbWhqlJobMocker.Reset()
		defer tusdMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebSubmitWHQLJobReq{
				SigningType: model.WHQLJobTypeHLKAndWHQL,
				TestSystem:  model.WHQLJobTestSystemWindows10_22H2_64,
				FileID:      fileID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signingType int, testSystem, fileID, serviceName, testTarget string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebSubmitWHQLJobReq{
				SigningType: signingType,
				TestSystem:  testSystem,
				FileID:      fileID,
				ServiceName: serviceName,
				TestTarget:  testTarget,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                        string
		SigningType                                 int
		TestSystem, FileID, ServiceName, TestTarget string
	}{
		{"签名类型缺失", 0, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38), "", ""},
		{"签名类型非法", 3, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38), "", ""},
		{"测试系统缺失", model.WHQLJobTypeHLKAndWHQL, "", util.FastRandomAlphaNumberString(38), "", ""},
		{"测试系统非法", model.WHQLJobTypeHLKAndWHQL, "InvalidSystem", util.FastRandomAlphaNumberString(38), "", ""},
		{"文件ID缺失", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, "", "", ""},
		{"文件ID错误", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(37), "", ""},
		{"文件ID过长", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(39), "", ""},
		{"文件ID非法", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(37) + "汉", "", ""},
		{"服务名过长", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38), util.FastRandomAlphaNumberString(257), ""},
		{"测试对象过长", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38), "", util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SigningType, v.TestSystem, v.FileID, v.ServiceName, v.TestTarget)
		})
	}
}

func TestWindowsWebListWHQLJobs(t *testing.T) {
	const reqPath = "/web/windows/listWHQLJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		pageNumber := 1
		pageSize := 10
		// 模拟数据库中的 WHQL 任务列表数据（空结构体，仅占位）。
		mockWhqlJobList := []*model.WhqlJob{
			{Source: model.SourceWeb, UserID: LoginUser.ID},
			{Source: model.SourceWeb, UserID: LoginUser.ID},
		}
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟数据库中的文件列表数据（空结构体，仅占位）。
		mockFileList := []*model.File{{}, {}}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWhqlJobMocker := MockDBClient[model.WhqlJob](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		dbWhqlJobMocker = dbWhqlJobMocker.CountOnce(2, nil)                                 // 统计 WHQL 任务总数。
		dbWhqlJobMocker = dbWhqlJobMocker.FindOnce(mockWhqlJobList, nil)                    // 分页查询 WHQL 任务记录。
		dbFileMocker = dbFileMocker.FindOnce(mockFileList, nil)                             // 查询文件信息。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询用户英文名。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWhqlJobMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsWebListWHQLJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.WindowsWebListWHQLJobsReq{
					PageNumber: pageNumber,
					PageSize:   pageSize,
				},
			)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
		if rspBodyObj.Data.Count <= 0 {
			t.Errorf("count is zero")
		}
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize int) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.WindowsWebListWHQLJobsReq{
					PageNumber: pageNumber,
					PageSize:   pageSize,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageNumber, PageSize int
	}{
		{"页数缺失", 0, 10},
		{"每页条数缺失", 1, 0},
		{"每页条数过大", 1, 101},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageNumber, v.PageSize)
		})
	}
}

func TestWindowsWebRemoveCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/removeCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		// 模拟数据库中的证书记录。
		mockCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}
		// 模拟数据库操作结果（空影响行数）。
		mockEmptyResult := gen.ResultInfo{}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                   // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                    // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                               // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                              // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                                                       // 查询上传人英文名。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)                                    // 查询证书记录。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)              // 设置证书为软删除状态。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.DeleteOnce(mockEmptyResult, nil) // 删除关联授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, &protocol.WindowsWebRemoveCompanyCertificateReq{
				CertificateID: certificateID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, &protocol.WindowsWebRemoveCompanyCertificateReq{
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestWindowsWebDeleteCertificate(t *testing.T) {
	const reqPath = "/web/windows/deleteCertificate"

	t.Run("正常测试_删除OV证书", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		// 模拟数据库中的 OV 证书记录。
		mockOVCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalOV}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                  // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                      // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                 // 校验应用管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockOVCert, nil)                     // 查询 OV 证书记录。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil) // 设置证书为软删除状态。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                         // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.WindowsWebDeleteCertificateReq{CertificateID: certificateID})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_删除EV证书", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		// 模拟数据库中的 EV 证书记录。
		mockEVCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                 // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                 // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                               // 校验应用管理员权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockEVCert, nil)                                   // 查询 EV 证书记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                                       // 添加应用事件到数据库。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.DeleteOnce(mockRowsAffected, nil) // 删除 EV 证书授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbEventMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.WindowsWebDeleteCertificateReq{CertificateID: certificateID})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebDeleteCertificateReq{
				CertificateID: certificateID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}
