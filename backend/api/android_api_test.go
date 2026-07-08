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

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行 Redis 防抖脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			CreateOnce(nil). // 添加安卓证书主体表记录。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行 Redis 防抖脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).                                                   // 获取数据库登录用户数据。
			FindOnce([]*model.User{{ID: 1, NameEn: "zs"}, {ID: 2, NameEn: "ls"}}, nil). // 获取数据库中用户的英文名。
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			FindOnce([]*model.AndroidOrganization{{UserID: 1}, {UserID: 1}, {UserID: 2}}, nil). // 获取数据库中安卓证书主体数据。
			Reset()

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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool"). // 设置 keytool 文件路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			SAddOnce(1, nil).                                          // 添加安卓证书 ID 到 Redis。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			ScanOnce(func(v any) { *v.(*string) = "CN=ivfzhou" }, nil). // 获取数据库中安卓证书主体数据。
			Reset()
		defer MockDBClient[model.AesKey](ctx).
			LastOnce(&model.AesKey{
				ID:          1,
				Secret:      util.RandomBytes(16),
				CreatedTime: time.Now(),
			}, nil). // 获取数据库中证书 AES 加密密钥。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			CreateOnce(nil). // 添加安卓证书信息到数据库。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool"). // 设置 keytool 文件路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			SAddOnce(1, nil).                                          // 添加安卓证书 ID 到 Redis。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			CreateOnce(nil). // 添加安卓证书信息到数据库。
			Reset()
		defer MockDBClient[model.AesKey](ctx).
			LastOnce(&model.AesKey{
				ID:          1,
				Secret:      util.RandomBytes(16),
				CreatedTime: time.Now(),
			}, nil). // 获取数据库中证书 AES 加密密钥。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()

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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).             // 获取数据库登录用户数据。
			FindOnce([]*model.User{{}, {}}, nil). // 获取数据库中用户英文名。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			FindOnce([]*model.AndroidCertificate{{}, {}}, nil). // 获取数据库中安卓证书数据。
			Reset()

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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥。
			Reset()
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("certificate content"))
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt}, nil). // 获取数据库中安卓证书数据。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool"). // 设置 keytool 文件路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥数据。
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil). // 获取数据库中安卓证书数据。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar"). // 设置 pepk jar 文件路径。
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set("java"). // 设置 java 执行器路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥数据。
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil). // 获取数据库中安卓证书数据。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar"). // 设置 pepk jar 文件路径。
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set("java"). // 设置 java 执行器路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥数据。
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥数据。
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil). // 获取数据库中安卓证书数据。
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil). // 获取数据库中安卓证书数据。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool"). // 设置 keytool 文件路径。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil). // 获取数据库中证书加密密钥数据。
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil). // 获取数据库中安卓证书数据。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加应用事件数据到数据库。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			SAddOnce(1, nil).                                          // 添加安卓签名任务 ID。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAPK,
				Type:   model.FileTypeAndroidSigning,
			}, nil). // 获取数据库中文件信息数据。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil). // 获取数据库中安卓证书信息。
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil). // 添加数据库中安卓签名任务信息。
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil). // 发送安卓签名任务消息到消息队列。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			SAddOnce(1, nil).                                          // 添加安卓签名任务 ID。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAAB,
				Type:   model.FileTypeAndroidSigning,
			}, nil). // 获取数据库中文件信息数据。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil). // 获取数据库中安卓证书信息。
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil). // 添加数据库中安卓签名任务信息。
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil). // 发送安卓签名任务消息到消息队列。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			SAddOnce(1, nil).                                          // 添加安卓签名任务 ID。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAPK,
				Type:   model.FileTypeAndroidSigning,
			}, nil). // 获取数据库中文件信息数据。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil). // 获取数据库中安卓证书信息。
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil). // 添加数据库中安卓签名任务信息。
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil). // 发送安卓签名任务消息到消息队列。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).                              // 获取数据库登录用户数据。
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil). // 查询数据库中用户 IDs。
			FindOnce([]*model.User{{}, {}}, nil).                  // 获取数据库中用户的英文名。
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil). // 查询数据库中请求凭证账号 IDs。
			FindOnce([]*model.APIAccount{{}, {}}, nil).            // 获取数据库中请求凭证账号名称。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil). // 查找数据库中安卓证书 IDs。
			FindOnce([]*model.AndroidCertificate{{}, {}}, nil).    // 查找数据库中证书的别名。
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			AndroidSigningJobGetTablesOnce([]string{"~"}, nil).                 // 获取安卓签名任务表名。
			AndroidSigningJobCount2Once(1, nil).                                // 获取数据库中安卓签名任务记录总数。
			AndroidSigningJobListOnce([]*model.AndroidSigningJob{{}, {}}, nil). // 获取数据库中安卓签名任务记录。
			Reset()
		defer MockDBClient[model.File](ctx).
			FindOnce([]*model.File{{}, {}}, nil). // 获取数据库中文件信息。
			Reset()

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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil). // 删除数据库中的安卓证书主体。
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: 1})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
			Set(model.AppPlatformAndroid). // 设置应用平台。
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{}, nil).                   // 获取数据库证书数据。
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil). // 设置数据库安卓证书记录为软删除状态。
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil). // 添加数据库应用事件记录。
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.AndroidWebDeleteCertificateReq{CertificateID: util.FastRandomAlphaNumberString(32)})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 防抖脚本。
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil). // 加载 Redis 限流脚本。
			EvalshaOnce(true, nil).                                    // 执行防抖过滤 Redis Lua 脚本。
			GetOnce(Session, nil).                                     // 获取 Redis 用户会话数据。
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil). // 获取数据库登录用户数据。
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil). // 获取数据库应用数据。
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil). // 判断登录用户角色。
			Reset()
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
