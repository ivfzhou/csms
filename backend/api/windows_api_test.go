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

func TestWindowsAPI_WebUploadCertificate(t *testing.T) {
	const reqPath = "/web/windows/uploadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificate := windowsCertificate
		password := windowsCertificatePassword

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil)      // 校验应用管理员权限。
		aesKeyMocker = aesKeyMocker.LastOnce(&model.AesKey{CreatedTime: time.Now(), Secret: util.RandomBytes(16)}, nil) // 查询数据库中加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                      // 缓存证书 ID 到 Redis。
		windowsCertMocker = windowsCertMocker.CreateOnce(nil)                                                           // 创建证书记录。
		eventMocker = eventMocker.CreateOnce(nil)                                                                       // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer windowsCertMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebUploadCertificateReq{
				Certificate: certificate,
				Password:    password,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificate, password string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.WindowsWebUploadCertificateReq{
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
		{"证书非法", "汉", "123456"},
		{"密码缺失", windowsCertificate, ""},
		{"密码过长", windowsCertificate, util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Certificate, v.Password)
		})
	}
}

func TestWindowsAPI_WebListCertificates(t *testing.T) {
	const reqPath = "/web/windows/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		windowsCertAuthMocker = windowsCertAuthMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)        // 查询已授权证书 IDs。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                                           // 查询上传人英文名。
		windowsCertMocker = windowsCertMocker.FindOnce([]*model.WindowsCertificate{{}, {}}, nil)                   // 批量查询证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertAuthMocker.Reset()
		defer windowsCertMocker.Reset()

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

func TestWindowsAPI_WebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/windows/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。

		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("123"))
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Content: encrypt}, nil) // 查询数据库证书记录。
		eventMocker = eventMocker.CreateOnce(nil)                                                        // 添加应用事件到数据库。
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil)                         // 查询数据库解密密钥。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer eventMocker.Reset()
		defer aesKeyMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebAddEVCertificate(t *testing.T) {
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		windowsCertMocker = windowsCertMocker.CountOnce(0, nil)                             // 检查相同指纹证书是否已存在。
		redisMocker = redisMocker.SAddOnce(1, nil)                                          // 缓存 EV 证书 ID 到 Redis。
		windowsCertMocker = windowsCertMocker.CreateOnce(nil)                               // 创建 EV 证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebUploadCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/uploadCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificate := windowsCertificate
		password := windowsCertificatePassword

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                            // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                               // 校验系统管理员权限。
		aesKeyMocker = aesKeyMocker.LastOnce(&model.AesKey{CreatedTime: time.Now(), Secret: util.RandomBytes(16)}, nil) // 查询数据库中加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                      // 缓存公司证书 ID 到 Redis。
		windowsCertMocker = windowsCertMocker.CreateOnce(nil)                                                           // 创建公司证书记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer aesKeyMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebListCompanyCertificates(t *testing.T) {
	const reqPath = "/web/windows/listCompanyCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                     // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                        // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                         // 查询上传人英文名。
		windowsCertMocker = windowsCertMocker.FindOnce([]*model.WindowsCertificate{{}, {}}, nil) // 批量查询公司证书。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()

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

func TestWindowsAPI_WebGrantAppEVCertificate(t *testing.T) {
	const reqPath = "/web/windows/grantAppEVCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		certificateID := util.FastRandomAlphaNumberString(32)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{}, nil)    // 查询证书记录。
		windowsCertAuthMocker = windowsCertAuthMocker.CreateOnce(nil)                       // 创建证书授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer windowsCertAuthMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebGetCertificatePassword(t *testing.T) {
	const reqPath = "/web/windows/getPassword"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                                         // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                                             // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                                        // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                                                            // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil)                                                                  // 校验应用管理员权限。
		windowsCertAuthMocker = windowsCertAuthMocker.CountOnce(1, nil)                                                                                                             // 校验应用是否已获授权。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV, Password: util.FastRandomAlphaNumberString(6)}, nil) // 查询 EV 证书密码。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer windowsCertAuthMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebDownloadCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/downloadCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                        // 查询上传人英文名。

		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("abc"))
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Content: encrypt}, nil) // 查询公司证书记录。
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil)                         // 查询数据库解密密钥。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer aesKeyMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebListGrantCertificateApps(t *testing.T) {
	const reqPath = "/web/windows/listGrantCertificateApps"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		appID := util.FastRandomAlphaNumberString(32)
		pageSize := 1
		pageNumber := 1

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                       // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                      // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                         // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                                              // 查询上传人英文名。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                          // 查询指定应用信息。
		dbAppMocker = dbAppMocker.FindOnce([]*model.App{{}}, nil)                                                 // 批量查询已授权应用信息。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{}, nil)                          // 查询证书记录。
		windowsCertMocker = windowsCertMocker.FindOnce([]*model.WindowsCertificate{{}}, nil)                      // 批量查询证书记录。
		windowsCertAuthMocker = windowsCertAuthMocker.CountOnce(1, nil)                                           // 统计授权记录总数。
		windowsCertAuthMocker = windowsCertAuthMocker.FindOnce([]*model.WindowsCertificateAuthorization{{}}, nil) // 分页查询授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer windowsCertAuthMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebRemoveCompanyCertificate(t *testing.T) {
	const reqPath = "/web/windows/removeCompanyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                         // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                                            // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                                                                 // 查询上传人英文名。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV}, nil) // 查询证书记录。
		windowsCertMocker = windowsCertMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                           // 设置证书为软删除状态。
		windowsCertAuthMocker = windowsCertAuthMocker.DeleteOnce(gen.ResultInfo{}, nil)                                              // 删除关联授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer windowsCertAuthMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()

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

func TestWindowsAPI_WebDeleteCertificate(t *testing.T) {
	const reqPath = "/web/windows/deleteCertificate"

	t.Run("正常测试_删除OV证书", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                         // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                             // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil)                   // 校验应用管理员权限。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalOV}, nil) // 查询 OV 证书记录。
		windowsCertMocker = windowsCertMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                           // 设置证书为软删除状态。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                    // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.WindowsWebDeleteCertificateReq{CertificateID: certificateID})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_删除EV证书", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformWindows).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		windowsCertMocker := MockDBClient[model.WindowsCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		windowsCertAuthMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                         // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                             // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil)                   // 校验应用管理员权限。
		windowsCertMocker = windowsCertMocker.TakeOnce(&model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV}, nil) // 查询 EV 证书记录。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                    // 添加应用事件到数据库。
		windowsCertAuthMocker = windowsCertAuthMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil)                               // 删除 EV 证书授权记录。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer windowsCertMocker.Reset()
		defer eventMocker.Reset()
		defer windowsCertAuthMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

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
