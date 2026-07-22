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
	"encoding/base64"
	"io"
	"net/http"
	"testing"
	"time"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestAndroidWebAddOrganization(t *testing.T) {
	const reqPath = "/web/android/addOrganization"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		commonName := "company certificate"
		dName := "C=CN,ST=Hunan,L=Changsha,CN=company_android_cert"

		redisMocker := MockRedis(ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAndroidOrganizationMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                     // 校验系统管理员权限。
		dbAndroidOrganizationMocker = dbAndroidOrganizationMocker.CreateOnce(nil) // 保存安卓证书主体到数据库。
		defer redisMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAndroidOrganizationMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AndroidWebAddOrganizationReq{
				CommonName: commonName,
				DName:      dName,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, commonName, dName string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AndroidWebAddOrganizationReq{
				CommonName: commonName,
				DName:      dName,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name       string
		CommonName string
		DName      string
	}{
		{"通用名缺失", "", "C=CN,ST=Hunan,L=Changsha,CN=company_android_cert"},
		{"通用名过长", util.FastRandomAlphaNumberString(33), "C=CN,ST=Hunan,L=Changsha,CN=company_android_cert"},
		{"DName 缺失", "company certificate", ""},
		{"DName 过长", "company certificate", util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CommonName, v.DName)
		})
	}
}

func TestAndroidWebListOrganizations(t *testing.T) {
	const reqPath = "/web/android/listOrganizations"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockUserList := []*model.User{{ID: 1, NameEn: "zs"}, {ID: 2, NameEn: "ls"}}        // 模拟数据库中的用户列表数据。
		mockOrgList := []*model.AndroidOrganization{{UserID: 1}, {UserID: 1}, {UserID: 2}} // 模拟数据库中的安卓证书主体列表数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAndroidOrganizationMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                 // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                              // 查询数据库中用户英文名。
		dbAndroidOrganizationMocker = dbAndroidOrganizationMocker.FindOnce(mockOrgList, nil) // 查询数据库中安卓证书主体数据。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAndroidOrganizationMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListOrganizationsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("want list, got empty list")
		}
	})
}

func TestAndroidWebApplyCertificate(t *testing.T) {
	const reqPath = "/web/android/applyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certType := model.AndroidCertificateTypeDebug
		ownerID := 1
		alias := "alias"
		mockOrgDName := "CN=ivfzhou" // 模拟数据库中安卓证书主体的 DName 数据。
		mockAesKey := &model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		} // 模拟数据库中的 AES 加密密钥记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAndroidOrganizationMocker := MockDBClient[model.AndroidOrganization](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		keytoolPathReset := mvt.Chain(&consts.KeytoolBinaryPath).Elem().Set("keytool")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                 // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                                // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                     // 查询数据库应用信息。
		dbAndroidOrganizationMocker = dbAndroidOrganizationMocker.ScanOnce(func(v any) { *v.(*string) = mockOrgDName }, nil) // 查询数据库中安卓证书主体数据。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)                                                            // 查询数据库中证书 AES 加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                           // 添加安卓证书 ID 到 Redis。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.CreateOnce(nil)                                              // 保存安卓证书信息到数据库。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                                        // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer keytoolPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAndroidOrganizationMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebApplyCertificateReq{
					Type:    certType,
					OwnerID: ownerID,
					Alias:   alias,
				}),
			),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ, ownerID int, alias string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebApplyCertificateReq{
					Type:    typ,
					OwnerID: ownerID,
					Alias:   alias,
				}),
			),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name    string
		Type    int
		OwnerID int
		Alias   string
	}{
		{"类型缺失", 0, 1, "alias"},
		{"类型错误", 3, 1, "alias"},
		{"主体缺失", 1, 0, "alias"},
		{"证书别名缺失", 1, 1, ""},
		{"证书别名字符非法", 1, 1, "～"},
		{"证书别名字符过多", 1, 1, util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.OwnerID, v.Alias)
		})
	}
}

func TestAndroidWebUploadCertificate(t *testing.T) {
	const reqPath = "/web/android/uploadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockAesKey := &model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		} // 模拟数据库中的 AES 加密密钥记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformRest := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		keytoolPathReset := mvt.Chain(&consts.KeytoolBinaryPath).Elem().Set("keytool")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                        // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                         // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                    // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                        // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                   // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)               // 查询数据库中证书 AES 加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)                              // 添加安卓证书 ID 到 Redis。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.CreateOnce(nil) // 保存安卓证书信息到数据库。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                           // 添加应用事件到数据库。
		defer appPlatformRest.Reset()
		defer keytoolPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebUploadCertificateReq{
					Type:        model.AndroidCertificateTypeDebug,
					Storepass:   AndroidKeystoreStorepass,
					Keypass:     AndroidKeystoreKeypass,
					Certificate: AndroidCertificate,
				}),
			),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ int, storepass, keypass, certificate string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebUploadCertificateReq{
					Type:        typ,
					Storepass:   storepass,
					Keypass:     keypass,
					Certificate: certificate,
				}),
			),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		Type        int
		Storepass   string
		Keypass     string
		Certificate string
	}{
		{"类型缺失", 0, "123456", "123456", AndroidCertificate},
		{"类型错误", -1, "123456", "123456", AndroidCertificate},
		{"Storepass 缺失", model.AndroidCertificateTypeDebug, "", "123456", AndroidCertificate},
		{"Storepass 过长", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(33), "123456", AndroidCertificate},
		{"Storepass 过短", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(5), "123456", AndroidCertificate},
		{"Keypass 缺失", model.AndroidCertificateTypeDebug, "123456", "", AndroidCertificate},
		{"Keypass 过长", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(33), AndroidCertificate},
		{"Keypass 过短", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(5), AndroidCertificate},
		{"证书缺失", model.AndroidCertificateTypeDebug, "123456", "123456", ""},
		{"证书格式错误", model.AndroidCertificateTypeDebug, "123456", "123456", "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.Storepass, v.Keypass, v.Certificate)
		})
	}
}

func TestAndroidWebListCertificates(t *testing.T) {
	const reqPath = "/web/android/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockUserList := []*model.User{{}, {}}               // 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockCertList := []*model.AndroidCertificate{{}, {}} // 模拟数据库中的安卓证书列表数据（空结构体，仅占位）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询数据库中用户英文名。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.FindOnce(mockCertList, nil) // 查询数据库中安卓证书数据。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})
}

func TestAndroidWebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/android/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("certificate content"))
		mockAesKey := &model.AesKey{Secret: secret}             // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{Content: encrypt} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书加密密钥。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书数据。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
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
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDownloadCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidWebGetGooglePlayCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		secret := util.RandomBytes(16)
		bs, _ := base64.StdEncoding.DecodeString(AndroidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		certificateID := util.FastRandomAlphaNumberString(32)
		mockAesKey := &model.AesKey{Secret: secret} // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{
			Content:   encrypt,
			Storepass: AndroidKeystoreStorepass,
			Keypass:   AndroidKeystoreKeypass,
			Alias_:    AndroidKeystoreAlias,
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		keytoolPathReset := mvt.Chain(&consts.KeytoolBinaryPath).Elem().Set("keytool")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书加密密钥。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书数据。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer keytoolPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
			protocol.AndroidWebGetGooglePlayCertificateReq{CertificateID: certificateID}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebGetGooglePlayCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidWebGetGooglePlayDeployCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayDeployCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		secret := util.RandomBytes(16)
		bs, _ := base64.StdEncoding.DecodeString(AndroidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		certificateID := util.FastRandomAlphaNumberString(32)
		mockAesKey := &model.AesKey{Secret: secret} // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{
			Content:   encrypt,
			Storepass: AndroidKeystoreStorepass,
			Keypass:   AndroidKeystoreKeypass,
			Alias_:    AndroidKeystoreAlias,
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		pepkPathReset := mvt.Chain(&consts.PepkJarPath).Elem().Set("../../pepk.jar")
		javaPathReset := mvt.Chain(&consts.JavaBinaryPathForPepk).Elem().Set("java")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书加密密钥。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书数据。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer pepkPathReset.Reset()
		defer javaPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
			&protocol.AndroidWebGetGooglePlayDeployCertificateReq{
				CertificateID: certificateID,
				PublicKey:     AndroidRSAPublicKey,
			}),
		)

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebGetGooglePlayDeployCertificateReq{
					CertificateID: certificateID,
					PublicKey:     AndroidRSAPublicKey,
				}),
			),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidWebGetGooglePlayUpgradeCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayUpgradeCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		secret := util.RandomBytes(16)
		bs, _ := base64.StdEncoding.DecodeString(AndroidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		deployCertificateID := util.FastRandomAlphaNumberString(32)
		uploadCertificateID := util.FastRandomAlphaNumberString(32)
		mockAesKey := &model.AesKey{Secret: secret} // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{
			Content:   encrypt,
			Storepass: AndroidKeystoreStorepass,
			Keypass:   AndroidKeystoreKeypass, Alias_: AndroidKeystoreAlias,
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		pepkPathReset := mvt.Chain(&consts.PepkJarPath).Elem().Set("../../pepk.jar")
		javaPathReset := mvt.Chain(&consts.JavaBinaryPathForPepk).Elem().Set("java")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中 AesKey。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中 AesKey（第二证书可能不同密钥）。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中部署证书数据。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中上传证书数据。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer pepkPathReset.Reset()
		defer javaPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
			&protocol.AndroidWebGetGooglePlayUpgradeCertificateReq{
				DeployCertificateID: deployCertificateID,
				UploadCertificateID: uploadCertificateID,
				PublicKey:           AndroidRSAPublicKey,
			}),
		)

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, deployCertificateID, uploadCertificateID, publicKey string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebGetGooglePlayUpgradeCertificateReq{
					DeployCertificateID: deployCertificateID,
					UploadCertificateID: uploadCertificateID,
					PublicKey:           publicKey,
				}),
			),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                                string
		DeployCertificateID, UploadCertificateID, PublicKey string
	}{
		{"部署证书 ID 缺失", "", util.FastRandomAlphaNumberString(32), AndroidRSAPublicKey},
		{"部署证书 ID 错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(32), AndroidRSAPublicKey},
		{"部署证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(32), AndroidRSAPublicKey},
		{"升级证书 ID 缺失", util.FastRandomAlphaNumberString(32), "", AndroidRSAPublicKey},
		{"升级证书 ID 错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31), AndroidRSAPublicKey},
		{"升级证书 ID 非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31) + "汉", AndroidRSAPublicKey},
		{"公钥缺失", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(32), ""},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.DeployCertificateID, v.UploadCertificateID, v.PublicKey)
		})
	}
}

func TestAndroidWebGetCertificateFacebookDigest(t *testing.T) {
	const reqPath = "/web/android/getCertificateFacebookDigest"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		secret := util.RandomBytes(16)
		bs, _ := base64.StdEncoding.DecodeString(AndroidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		certificateID := util.FastRandomAlphaNumberString(32)
		mockAesKey := &model.AesKey{Secret: secret} // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{
			Content:   encrypt,
			Storepass: AndroidKeystoreStorepass,
			Keypass:   AndroidKeystoreKeypass,
			Alias_:    AndroidKeystoreAlias,
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		keytoolPathReset := mvt.Chain(&consts.KeytoolBinaryPath).Elem().Set("keytool")
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书加密密钥。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书数据。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer keytoolPathReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebGetCertificateFacebookDigestRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebGetCertificateFacebookDigestReq{CertificateID: certificateID})),
			0,
		)

		if len(rspBodyObj.Data.Digest) <= 0 {
			t.Errorf("digest is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidWebSubmitAPKSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		signatureSchema := []int{1}
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（APK 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKSigningJobReq{
					SignatureSchema: signatureSchema,
					CertificateID:   certificateID,
					FileID:          fileID,
				},
			)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKSigningJobReq{
					SignatureSchema: signatureSchema,
					CertificateID:   certificateID,
					FileID:          fileID,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name            string
		SignatureSchema []int
		CertificateID   string
		FileID          string
	}{
		{"签名方案缺失", nil, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"签名方案重复", []int{1, 1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 缺失", []int{1}, "", util.FastRandomAlphaNumberString(38)},
		{"证书 ID 错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidWebSubmitAABSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAABSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAAB,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（AAB 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAABSigningJobReq{
					CertificateID: certificateID,
					FileID:        fileID,
				},
			)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAABSigningJobReq{
					CertificateID: certificateID,
					FileID:        fileID,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
		FileID        string
	}{
		{"证书 ID 缺失", "", util.FastRandomAlphaNumberString(38)},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidWebSubmitAPKPatchSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKPatchSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		signatureSchema := []int{1}
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		minimumSDKVersion := 1
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（APK 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKPatchSigningJobReq{
					SignatureSchema:   signatureSchema,
					CertificateID:     certificateID,
					FileID:            fileID,
					MinimumSDKVersion: minimumSDKVersion,
				},
			)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string, minimumSDKVersion int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKPatchSigningJobReq{
					SignatureSchema:   signatureSchema,
					CertificateID:     certificateID,
					FileID:            fileID,
					MinimumSDKVersion: minimumSDKVersion,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name              string
		SignatureSchema   []int
		CertificateID     string
		FileID            string
		MinimumSDKVersion int
	}{
		{"签名方案缺失", nil, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 1},
		{"签名方案重复", []int{1, 1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 缺失", []int{1}, "", util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38), 1},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), "", 1},
		{"文件 ID 错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37), 1},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉", 1},
		{"SDK 版本缺失", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 0},
		{"SDK 版本非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID, v.MinimumSDKVersion)
		})
	}
}

func TestAndroidWebListSigningJobs(t *testing.T) {
	const reqPath = "/web/android/listSigningJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		keyWord := "~"
		status := 1
		certificateAlias := "~"
		pageNumber := 1
		pageSize := 1
		mockUserList := []*model.User{{}, {}}                      // 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockAPIAccountList := []*model.APIAccount{{}, {}}          // 模拟数据库中的 API 凭证列表数据（空结构体，仅占位）。
		mockAndroidCertList := []*model.AndroidCertificate{{}, {}} // 模拟数据库中的安卓证书列表数据（空结构体，仅占位）。
		mockTableNames := []string{"~"}                            // 模拟签名任务表名列表。
		mockSigningJobList := []*model.AndroidSigningJob{{}, {}}   // 模拟数据库中的签名任务列表数据（空结构体，仅占位）。
		mockFileList := []*model.File{{}, {}}                      // 模拟数据库中的文件列表数据（空结构体，仅占位）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                            // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                            // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                         // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                             // 查询数据库中用户 IDs。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                                                       // 查询数据库中用户英文名。
		dbAPIAccountMocker = dbAPIAccountMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                 // 查询数据库中请求凭证账号 IDs。
		dbAPIAccountMocker = dbAPIAccountMocker.FindOnce(mockAPIAccountList, nil)                                     // 查询数据库中请求凭证账号名称。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中安卓证书 IDs。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.FindOnce(mockAndroidCertList, nil)                    // 查询数据库中证书别名。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobGetTablesOnce(mockTableNames, nil)     // 获取安卓签名任务表名。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobCount2Once(1, nil)                     // 统计安卓签名任务总数。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobListOnce(mockSigningJobList, nil)      // 分页查询安卓签名任务记录。
		dbFileMocker = dbFileMocker.FindOnce(mockFileList, nil)                                                       // 查询数据库中文件信息。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebListSigningJobsReq{
					KeyWord:          keyWord,
					Status:           status,
					CertificateAlias: certificateAlias,
					PageNumber:       pageNumber,
					PageSize:         pageSize,
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

	validateErrorRequest := func(t *testing.T, keyWord string, status int, certificateAlias string, pageNumber, pageSize int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebListSigningJobsReq{
				KeyWord:          keyWord,
				Status:           status,
				CertificateAlias: certificateAlias,
				PageNumber:       pageNumber,
				PageSize:         pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name             string
		Status           int
		CertificateAlias string
		PageNumber       int
		PageSize         int
	}{
		{"状态非法", -1, "", 1, 1},
		{"别名过长", 0, util.FastRandomAlphaNumberString(65), 1, 1},
		{"页码非法", 0, "", 0, 1},
		{"页条数非法", 0, "", 1, 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, "", v.Status, v.CertificateAlias, v.PageNumber, v.PageSize)
		})
	}
}

func TestAndroidWebRemoveOrganization(t *testing.T) {
	const reqPath = "/web/android/removeOrganization"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		id := 1
		mockResult := gen.ResultInfo{RowsAffected: 1} // 模拟数据库删除操作返回结果（影响一行）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAndroidOrganizationMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                  // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                 // 校验系统管理员权限。
		dbAndroidOrganizationMocker = dbAndroidOrganizationMocker.DeleteOnce(mockResult, nil) // 删除数据库中安卓证书主体。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAndroidOrganizationMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: id})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: id})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID 缺失", 0},
		{"ID 非法", -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}

func TestAndroidWebDeleteCertificate(t *testing.T) {
	const reqPath = "/web/android/deleteCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		mockCert := &model.AndroidCertificate{}       // 模拟数据库中的安卓证书记录（空结构体，表示已存在的证书）。
		mockResult := gen.ResultInfo{RowsAffected: 1} // 模拟数据库更新操作返回结果（影响一行）。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                           // 校验系统管理员权限。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil)                 // 查询数据库证书数据。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.UpdateColumnSimpleOnce(mockResult, nil) // 设置数据库安卓证书记录为软删除状态。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidWebStatisticSigningTimes(t *testing.T) {
	const reqPath = "/web/android/statisticSigningTimes"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_android_signing_job"}                               // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"count": 1, "type": 1, "day": "20260710"}} // 模拟按天统计的签名任务数量数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                            // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                           // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                          // 查询数据库应用 ID。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobGetTablesOnce(mockTableNames, nil)       // 获取安卓签名任务表名。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobCountWithDayOnce(mockStatisticData, nil) // 按天统计各类签名任务数量。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningTimesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningTimesReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, timeStep int, beginTime, endTime time.Time) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningTimesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningTimesReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name               string
		appID              string
		timeStep           int
		beginTime, endTime time.Time
	}{
		{"应用 ID 非法", util.FastRandomAlphaNumberString(31), protocol.TimeStepWeek, time.Now(), time.Now().Add(time.Hour)},
		{"结束时间小于开始时间", util.FastRandomAlphaNumberString(32), protocol.TimeStepWeek, time.Now(), time.Now().Add(-time.Hour)},
		{"时间步长非法", util.FastRandomAlphaNumberString(32), 0, time.Now(), time.Now().Add(time.Hour)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.appID, v.timeStep, v.beginTime, v.endTime)
		})
	}
}

func TestAndroidWebStatisticSigningCost(t *testing.T) {
	const reqPath = "/web/android/statisticSigningCost"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_android_signing_job"}                              // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"cost": 1, "type": 1, "day": "20260710"}} // 模拟按天统计的签名任务耗时数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                             // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                         // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                               // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                           // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                          // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                         // 查询数据库应用 ID。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobGetTablesOnce(mockTableNames, nil)      // 获取安卓签名任务表名。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobCostWithDayOnce(mockStatisticData, nil) // 按天统计各类签名任务耗时。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningCostRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningCostReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, timeStep int, beginTime, endTime time.Time) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningCostRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningCostReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name               string
		appID              string
		timeStep           int
		beginTime, endTime time.Time
	}{
		{"应用 ID 非法", util.FastRandomAlphaNumberString(31), protocol.TimeStepWeek, time.Now(), time.Now().Add(time.Hour)},
		{"结束时间小于开始时间", util.FastRandomAlphaNumberString(32), protocol.TimeStepWeek, time.Now(), time.Now().Add(-time.Hour)},
		{"时间步长非法", util.FastRandomAlphaNumberString(32), 0, time.Now(), time.Now().Add(time.Hour)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.appID, v.timeStep, v.beginTime, v.endTime)
		})
	}
}

func TestAndroidWebStatisticSigningPassRate(t *testing.T) {
	const reqPath = "/web/android/statisticSigningPassRate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_android_signing_job"}                              // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"rate": 1, "type": 1, "day": "20260710"}} // 模拟按天统计的签名任务通过率数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                 // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                             // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                   // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                    // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                               // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                              // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                             // 查询数据库应用 ID。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobGetTablesOnce(mockTableNames, nil)          // 获取安卓签名任务表名。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.AndroidSigningJobPassRateWithDayOnce(mockStatisticData, nil) // 按天统计各类签名任务通过率。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningPassRateRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningPassRateReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, timeStep int, beginTime, endTime time.Time) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AndroidWebStatisticSigningPassRateRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AndroidWebStatisticSigningPassRateReq{
				AppID:     appID,
				BeginTime: beginTime,
				EndTime:   endTime,
				TimeStep:  timeStep,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name               string
		appID              string
		timeStep           int
		beginTime, endTime time.Time
	}{
		{"应用 ID 非法", util.FastRandomAlphaNumberString(31), protocol.TimeStepWeek, time.Now(), time.Now().Add(time.Hour)},
		{"结束时间小于开始时间", util.FastRandomAlphaNumberString(32), protocol.TimeStepWeek, time.Now(), time.Now().Add(-time.Hour)},
		{"时间步长非法", util.FastRandomAlphaNumberString(32), 0, time.Now(), time.Now().Add(time.Hour)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.appID, v.timeStep, v.beginTime, v.endTime)
		})
	}
}
