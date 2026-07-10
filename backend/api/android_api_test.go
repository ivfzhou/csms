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

func TestAndroidAPI_WebAddOrganization(t *testing.T) {
	const reqPath = "/web/android/addOrganization"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		androidOrgMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		androidOrgMocker = androidOrgMocker.CreateOnce(nil)                                 // 保存安卓证书主体到数据库。
		defer redisMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbUserMocker.Reset()
		defer androidOrgMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AndroidWebAddOrganizationReq{
				CommonName: "company certificate",
				DName:      "C=CN,ST=Hunan,L=Changsha,CN=company_android_cert",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, commonName, dName string) {
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
		{"DName缺失", "company certificate", ""},
		{"DName过长", "company certificate", util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CommonName, v.DName)
		})
	}
}

func TestAndroidAPI_WebListOrganizations(t *testing.T) {
	const reqPath = "/web/android/listOrganizations"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		androidOrgMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                   // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                                      // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{ID: 1, NameEn: "zs"}, {ID: 2, NameEn: "ls"}}, nil)                 // 查询数据库中用户英文名。
		androidOrgMocker = androidOrgMocker.FindOnce([]*model.AndroidOrganization{{UserID: 1}, {UserID: 1}, {UserID: 2}}, nil) // 查询数据库中安卓证书主体数据。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer androidOrgMocker.Reset()

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

func TestAndroidAPI_WebApplyCertificate(t *testing.T) {
	const reqPath = "/web/android/applyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		androidOrgMocker := MockDBClient[model.AndroidOrganization](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)            // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)            // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                               // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                           // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                              // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                               // 查询数据库应用信息。
		androidOrgMocker = androidOrgMocker.ScanOnce(func(v any) { *v.(*string) = "CN=ivfzhou" }, nil) // 查询数据库中安卓证书主体数据。
		aesKeyMocker = aesKeyMocker.LastOnce(&model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		}, nil) // 查询数据库中证书 AES 加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)            // 添加安卓证书 ID 到 Redis。
		androidCertMocker = androidCertMocker.CreateOnce(nil) // 保存安卓证书信息到数据库。
		eventMocker = eventMocker.CreateOnce(nil)             // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer androidOrgMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebApplyCertificateReq{
					Type:    model.AndroidCertificateTypeDebug,
					OwnerID: 1,
					Alias:   "alias",
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebUploadCertificate(t *testing.T) {
	const reqPath = "/web/android/uploadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		aesKeyMocker = aesKeyMocker.LastOnce(&model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		}, nil) // 查询数据库中证书 AES 加密密钥。
		redisMocker = redisMocker.SAddOnce(1, nil)            // 添加安卓证书 ID 到 Redis。
		androidCertMocker = androidCertMocker.CreateOnce(nil) // 保存安卓证书信息到数据库。
		eventMocker = eventMocker.CreateOnce(nil)             // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer androidCertMocker.Reset()
		defer aesKeyMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebUploadCertificateReq{
					Type:        model.AndroidCertificateTypeDebug,
					Storepass:   androidKeystoreStorepass,
					Keypass:     androidKeystoreKeypass,
					Certificate: androidCertificate,
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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
		{"类型缺失", 0, "123456", "123456", androidCertificate},
		{"类型错误", -1, "123456", "123456", androidCertificate},
		{"Storepass 缺失", model.AndroidCertificateTypeDebug, "", "123456", androidCertificate},
		{"Storepass 过长", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(33), "123456", androidCertificate},
		{"Storepass 过短", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(5), "123456", androidCertificate},
		{"Keypass 缺失", model.AndroidCertificateTypeDebug, "123456", "", androidCertificate},
		{"Keypass 过长", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(33), androidCertificate},
		{"Keypass 过短", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(5), androidCertificate},
		{"证书缺失", model.AndroidCertificateTypeDebug, "123456", "123456", ""},
		{"证书格式错误", model.AndroidCertificateTypeDebug, "123456", "123456", "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.Storepass, v.Keypass, v.Certificate)
		})
	}
}

func TestAndroidAPI_WebListCertificates(t *testing.T) {
	const reqPath = "/web/android/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                     // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                         // 查询数据库中用户英文名。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                         // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                        // 校验系统管理员权限。
		androidCertMocker = androidCertMocker.FindOnce([]*model.AndroidCertificate{{}, {}}, nil) // 查询数据库中安卓证书数据。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer androidCertMocker.Reset()

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

func TestAndroidAPI_WebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/android/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。

		secret := util.RandomBytes(16)
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中证书加密密钥。
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("certificate content"))
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt}, nil) // 查询数据库中安卓证书数据。
		eventMocker = eventMocker.CreateOnce(nil)                                                        // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebGetGooglePlayCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。

		secret := util.RandomBytes(16)
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中证书加密密钥。
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil) // 查询数据库中安卓证书数据。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                                                                                            // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
			protocol.AndroidWebGetGooglePlayCertificateReq{CertificateID: util.FastRandomAlphaNumberString(32)}))

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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebGetGooglePlayDeployCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayDeployCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar").
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set("java").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。

		secret := util.RandomBytes(16)
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中证书加密密钥。
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil) // 查询数据库中安卓证书数据。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                                                                                            // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
			&protocol.AndroidWebGetGooglePlayDeployCertificateReq{
				CertificateID: util.FastRandomAlphaNumberString(32),
				PublicKey:     androidRSAPublicKey,
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebGetGooglePlayDeployCertificateReq{
					CertificateID: certificateID,
					PublicKey:     androidRSAPublicKey,
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

func TestAndroidAPI_WebGetGooglePlayUpgradeCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayUpgradeCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar").
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set("java").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。

		secret := util.RandomBytes(16)
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中 AesKey。
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中 AesKey（第二证书可能不同密钥）。
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil) // 查询数据库中部署证书数据。
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil) // 查询数据库中上传证书数据。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                                                                                            // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
			&protocol.AndroidWebGetGooglePlayUpgradeCertificateReq{
				DeployCertificateID: util.FastRandomAlphaNumberString(32),
				UploadCertificateID: util.FastRandomAlphaNumberString(32),
				PublicKey:           androidRSAPublicKey,
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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
		{"部署证书 ID 缺失", "", util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"部署证书 ID 错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"部署证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"升级证书 ID 缺失", util.FastRandomAlphaNumberString(32), "", androidRSAPublicKey},
		{"升级证书 ID 错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31), androidRSAPublicKey},
		{"升级证书 ID 非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31) + "汉", androidRSAPublicKey},
		{"公钥缺失", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(32), ""},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.DeployCertificateID, v.UploadCertificateID, v.PublicKey)
		})
	}
}

func TestAndroidAPI_WebGetCertificateFacebookDigest(t *testing.T) {
	const reqPath = "/web/android/getCertificateFacebookDigest"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		aesKeyMocker := MockDBClient[model.AesKey](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。

		secret := util.RandomBytes(16)
		aesKeyMocker = aesKeyMocker.TakeOnce(&model.AesKey{Secret: secret}, nil) // 查询数据库中证书加密密钥。
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil) // 查询数据库中安卓证书数据。
		eventMocker = eventMocker.CreateOnce(nil)                                                                                                                                                            // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer aesKeyMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebGetCertificateFacebookDigestRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebGetCertificateFacebookDigestReq{CertificateID: util.FastRandomAlphaNumberString(32)})),
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebSubmitAPKSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		androidJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		}, nil) // 查询数据库中文件信息。
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                          // 添加安卓签名任务 ID 到 Redis。
		androidJobMocker = androidJobMocker.CreateOnce(nil)                                                                 // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                                                         // 发送安卓签名任务消息到消息队列。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer androidCertMocker.Reset()
		defer androidJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKSigningJobReq{
					SignatureSchema: []int{1},
					CertificateID:   util.FastRandomAlphaNumberString(32),
					FileID:          util.FastRandomAlphaNumberString(38),
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebSubmitAABSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAABSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		androidJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAAB,
			Type:   model.FileTypeAndroidSigning,
		}, nil) // 查询数据库中文件信息。
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                          // 添加安卓签名任务 ID 到 Redis。
		androidJobMocker = androidJobMocker.CreateOnce(nil)                                                                 // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                                                         // 发送安卓签名任务消息到消息队列。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer androidCertMocker.Reset()
		defer androidJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAABSigningJobReq{
					CertificateID: util.FastRandomAlphaNumberString(32),
					FileID:        util.FastRandomAlphaNumberString(38),
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebSubmitAPKPatchSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKPatchSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		androidJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		}, nil) // 查询数据库中文件信息。
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                          // 添加安卓签名任务 ID 到 Redis。
		androidJobMocker = androidJobMocker.CreateOnce(nil)                                                                 // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                                                         // 发送安卓签名任务消息到消息队列。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer androidCertMocker.Reset()
		defer androidJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID,
				&protocol.AndroidWebSubmitAPKPatchSigningJobReq{
					SignatureSchema:   []int{1},
					CertificateID:     util.FastRandomAlphaNumberString(32),
					FileID:            util.FastRandomAlphaNumberString(38),
					MinimumSDKVersion: 1,
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
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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

func TestAndroidAPI_WebListSigningJobs(t *testing.T) {
	const reqPath = "/web/android/listSigningJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		androidJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                       // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                      // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                      // 查询数据库中用户 IDs。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                                       // 查询数据库中用户英文名。
		apiAccountMocker = apiAccountMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)              // 查询数据库中请求凭证账号 IDs。
		apiAccountMocker = apiAccountMocker.FindOnce([]*model.APIAccount{{}, {}}, nil)                         // 查询数据库中请求凭证账号名称。
		androidCertMocker = androidCertMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)            // 查询数据库中安卓证书 IDs。
		androidCertMocker = androidCertMocker.FindOnce([]*model.AndroidCertificate{{}, {}}, nil)               // 查询数据库中证书别名。
		androidJobMocker = androidJobMocker.AndroidSigningJobGetTablesOnce([]string{"~"}, nil)                 // 获取安卓签名任务表名。
		androidJobMocker = androidJobMocker.AndroidSigningJobCount2Once(1, nil)                                // 统计安卓签名任务总数。
		androidJobMocker = androidJobMocker.AndroidSigningJobListOnce([]*model.AndroidSigningJob{{}, {}}, nil) // 分页查询安卓签名任务记录。
		dbFileMocker = dbFileMocker.FindOnce([]*model.File{{}, {}}, nil)                                       // 查询数据库中文件信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer apiAccountMocker.Reset()
		defer androidCertMocker.Reset()
		defer androidJobMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebListSigningJobsReq{
					KeyWord:          "~",
					Status:           1,
					CertificateAlias: "~",
					PageNumber:       1,
					PageSize:         1,
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

func TestAndroidAPI_WebRemoveOrganization(t *testing.T) {
	const reqPath = "/web/android/removeOrganization"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		androidOrgMocker := MockDBClient[model.AndroidOrganization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                 // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                    // 校验系统管理员权限。
		androidOrgMocker = androidOrgMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 删除数据库中安卓证书主体。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer androidOrgMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: 1})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
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

func TestAndroidAPI_WebDeleteCertificate(t *testing.T) {
	const reqPath = "/web/android/deleteCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		androidCertMocker := MockDBClient[model.AndroidCertificate](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                   // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                    // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                               // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                   // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                  // 校验系统管理员权限。
		androidCertMocker = androidCertMocker.TakeOnce(&model.AndroidCertificate{}, nil)                   // 查询数据库证书数据。
		androidCertMocker = androidCertMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 设置数据库安卓证书记录为软删除状态。
		eventMocker = eventMocker.CreateOnce(nil)                                                          // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer androidCertMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDeleteCertificateReq{CertificateID: util.FastRandomAlphaNumberString(32)})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
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
