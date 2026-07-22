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
	"strings"
	"testing"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestAppleWebApplyBundleID(t *testing.T) {
	const reqPath = "/web/apple/applyBundleID"

	t.Run("正常测试_AppStore类型", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyBundleIDReq{
			BundleID: "cn.ivfzhou.test",
			Type:     model.AppleBundleIDTypeAppStore,
		}
		mockAppleAPIResponse := CreateAppleAPIResponse(`{"data":{"id":"bid","type":"bundleIds"}}`)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		httpMocker := MockHTTPClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.CountOnce(0, nil)        // 查询 Bundle ID 是否已被注册（不存在）。
		httpMocker = httpMocker.ResponseOnce(mockAppleAPIResponse, nil)        // HTTP 请求 Apple API 注册 Bundle ID。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.CreateOnce(nil)          // 插入 Apple Bundle ID 记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                          // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbEventMocker.Reset()
		defer httpMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_InHouse类型", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyBundleIDReq{
			BundleID: "cn.ivfzhou.test.inhouse",
			Type:     model.AppleBundleIDTypeInHouse,
		}
		mockAppleAPIResponse := CreateAppleAPIResponse(`{"code":0,"data":{"id":"bid","platform":"IOS"}}`)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		httpMocker := MockHTTPClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.CountOnce(0, nil)        // 查询 Bundle ID 是否已被注册（不存在）。
		httpMocker = httpMocker.ResponseOnce(mockAppleAPIResponse, nil)        // HTTP 请求 Fastlane 代理注册 Bundle ID。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.CreateOnce(nil)          // 插入 Apple Bundle ID 记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                          // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbEventMocker.Reset()
		defer httpMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID string, typ int) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyBundleIDReq{
			BundleID: bundleID,
			Type:     typ,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name     string
		bundleID string
		typ      int
	}{
		{"BundleID 为空", "", model.AppleBundleIDTypeAppStore},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65), model.AppleBundleIDTypeAppStore},
		{"Type 小于最小值", "cn.ivfzhou.test", 0},
		{"Type 大于最大值", "cn.ivfzhou.test", 3},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID, v.typ)
		})
	}
}

func TestAppleWebModifyBundleID(t *testing.T) {
	const reqPath = "/web/apple/modifyBundleID"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebModifyBundleIDReq{
			BundleID:     "cn.ivfzhou.test",
			Capabilities: []string{"in app purchase", "game center"},
		}
		mockBundle := &model.AppleBundleID{
			InAppleID:    "bid",
			Environment:  model.AppleBundleIDTypeAppStore,
			Capabilities: model.StringList{"in app purchase", "game center"},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                            // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                            // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                              // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                         // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.TakeOnce(mockBundle, nil)                                       // 查询 Bundle ID 信息。
		httpMocker = httpMocker.ResponseOnce(CreateAppleAPIResponse(`{"data":{"id":"bid","type":"bundleIds"}}`), nil) // HTTP 请求 Apple API 修改 Bundle ID 能力项。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                                 // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID string, capabilities []string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebModifyBundleIDReq{
			BundleID:     bundleID,
			Capabilities: capabilities,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name         string
		bundleID     string
		capabilities []string
	}{
		{"BundleID 为空", "", []string{"game center"}},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65), []string{"game center"}},
		{"Capabilities 包含重复值", "cn.ivfzhou.test", []string{"game center", "game center"}},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID, v.capabilities)
		})
	}
}

func TestAppleWebApplyCertificate(t *testing.T) {
	const reqPath = "/web/apple/applyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyCertificateReq{
			Type: "IOS_DEVELOPMENT",
		}
		mockAesKey := &model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		}
		mockAppleAPIResponse := CreateAppleAPIResponse(`{"data":{"id":"cert1","type":"certificates","attributes":{"certificateContent":"` + AppleCertificateCertDERBase64 + `"}}}`)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)               // 校验用户角色权限。
		httpMocker = httpMocker.ResponseOnce(mockAppleAPIResponse, nil)     // HTTP 请求 Apple API 申请签名证书。
		redisMocker = redisMocker.SAddOnce(1, nil)                          // Redis 生成唯一 ID。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)           // 查询数据库中证书 AES 加密密钥。
		dbAppleCertificateMocker = dbAppleCertificateMocker.CreateOnce(nil) // 插入 Apple 证书记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                       // 保存应用事件。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certType string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyCertificateReq{
			Type: certType,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name     string
		certType string
	}{
		{"证书类型非法", "INVALID_TYPE"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.certType)
		})
	}
}

func TestAppleWebListBundleIDs(t *testing.T) {
	const reqPath = "/web/apple/listBundleIDs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockBundleIDs := []*model.AppleBundleID{
			{InAppleID: "bundle_app_id_1", Environment: model.AppleBundleIDTypeAppStore},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                           // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                      // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.FindOnce(mockBundleIDs, nil) // 查询 Bundle ID 列表。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListBundleIDsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)
	})

	t.Run("异常测试_数据库错误", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)               // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                 // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                  // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                             // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                 // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                            // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.FindOnce(nil, gorm.ErrInvalidData) // 查询 Bundle ID 列表（数据库错误）。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListBundleIDsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.ErrSystem,
		)
	})
}

func TestAppleWebListCertificates(t *testing.T) {
	const reqPath = "/web/apple/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockCertificates := []*model.AppleCertificate{
			{CertificateID: "cert_id_1", Type: "IOS_DEVELOPMENT", Owner: LoginUser.NameEn},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.FindOnce(mockCertificates, nil) // 查询证书列表。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleCertificateMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)
	})

	t.Run("异常测试_数据库错误", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                   // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                  // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.FindOnce(nil, gorm.ErrInvalidData) // 查询证书列表（数据库错误）。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleCertificateMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			consts.ErrSystem,
		)
	})
}

func TestAppleWebRegisterDevice(t *testing.T) {
	const reqPath = "/web/apple/registerDevice"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebRegisterDeviceReq{
			UDID:   "00000000-0000000000000000000000000000000000000000000000000000000000000000",
			Device: "MAC",
			Remark: "测试设备",
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbAppleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                      // 校验用户角色权限。
		httpMocker = httpMocker.ResponseOnce(CreateAppleAPIResponse(`{"data":{"id":"d1","type":"devices"}}`), nil) // HTTP 请求 Apple API 注册测试设备。
		dbAppleDeviceMocker = dbAppleDeviceMocker.CreateOnce(nil)                                                  // 插入 Apple 设备记录。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)
		dbTodoMocker = dbTodoMocker.CreateOnce(nil)   // 创建待办。
		dbEventMocker = dbEventMocker.CreateOnce(nil) // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer httpMocker.Reset()
		defer dbAppleDeviceMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertRegisterAppleDevice,
		)
	})

	validateErrorRequest := func(t *testing.T, udid, device, remark string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebRegisterDeviceReq{
			UDID:   udid,
			Device: device,
			Remark: remark,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name   string
		udid   string
		device string
		remark string
	}{
		{"UDID 为空", "", "MAC", "测试设备"},
		{"UDID 过长", util.FastRandomAlphaNumberString(129), "MAC", "测试设备"},
		{"备注过长", "00000000-0000000000000000000000000000000000000000000000000000000000000000", "MAC", util.FastRandomAlphaNumberString(1025)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.udid, v.device, v.remark)
		})
	}
}

func TestAppleWebListDevices(t *testing.T) {
	const reqPath = "/web/apple/listDevices"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockDeviceList := []*model.AppleDevice{
			{Model: "MAC", Udid: "device_udid_1", Remark: "测试设备"},
		}
		count := int64(1)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                        // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                         // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                    // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                        // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                   // 校验用户角色权限。
		dbAppleDeviceMocker = dbAppleDeviceMocker.CountOnce(count, nil)         // 查询设备总数。
		dbAppleDeviceMocker = dbAppleDeviceMocker.FindOnce(mockDeviceList, nil) // 查询设备列表。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{LoginUser}, nil)     // 查询用户信息。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleDeviceMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppleWebListDevicesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppleWebListDevicesReq{
				PageNumber: 1,
				PageSize:   10,
			})),
			0,
		)
		if rspBodyObj.Data == nil {
			t.Fatalf("expect data not nil")
		}
		if rspBodyObj.Data.Count != count {
			t.Errorf("expect count %d, but got %d", count, rspBodyObj.Data.Count)
		}
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListDevicesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppleWebListDevicesReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name       string
		pageNumber int
		pageSize   int
	}{
		{"页号小于最小值", 0, 10},
		{"每页条数小于最小值", 1, 0},
		{"每页条数大于最大值", 1, 101},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.pageNumber, v.pageSize)
		})
	}
}

func TestAppleWebApplyProfile(t *testing.T) {
	const reqPath = "/web/apple/applyProfile"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyProfileReq{
			BundleID: "cn.ivfzhou.test",
			Type:     "IOS_APP_DEVELOPMENT",
			Platform: "IOS",
		}
		plistContent := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>UUID</key><string>p1</string><key>CreationDate</key><date>2025-01-01T00:00:00Z</date><key>ExpirationDate</key><date>2025-12-31T00:00:00Z</date></dict></plist>`
		profileB64 := base64.StdEncoding.EncodeToString([]byte(plistContent))
		mockBundle := &model.AppleBundleID{
			ID:          1,
			InAppleID:   "bid",
			Environment: model.AppleBundleIDTypeAppStore,
		}
		mockAppleAPIResponse := CreateAppleAPIResponse(`{"data":{"id":"p1","type":"profiles","attributes":{"profileContent":"` + profileB64 + `"}}}`)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAppleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                         // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                             // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                        // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.TakeOnce(mockBundle, nil)                      // 查询 Bundle ID 信息。
		dbAppleCertificateMocker = dbAppleCertificateMocker.TakeOnce(&model.AppleCertificate{}, nil) // 查询签名证书信息。
		dbAppleDeviceMocker = dbAppleDeviceMocker.FindOnce([]*model.AppleDevice{}, nil)              // 查询测试设备列表。
		httpMocker = httpMocker.ResponseOnce(mockAppleAPIResponse, nil)                              // HTTP 请求 Apple API 申请描述文件。
		dbAppleProfileMocker = dbAppleProfileMocker.CreateOnce(nil)                                  // 插入 Apple 描述文件记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer httpMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAppleDeviceMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID, profileType, platform string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyProfileReq{
			BundleID: bundleID,
			Type:     profileType,
			Platform: platform,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		bundleID    string
		profileType string
		platform    string
	}{
		{"BundleID 为空", "", "IOS_APP_DEVELOPMENT", "IOS"},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65), "IOS_APP_DEVELOPMENT", "IOS"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID, v.profileType, v.platform)
		})
	}
}

func TestAppleWebApplyInHouseProfile(t *testing.T) {
	const reqPath = "/web/apple/applyInHouseProfile"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyInHouseProfileReq{
			BundleID: "cn.ivfzhou.test",
		}
		plistContent := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>UUID</key><string>p1</string><key>CreationDate</key><date>2025-01-01T00:00:00Z</date><key>ExpirationDate</key><date>2025-12-31T00:00:00Z</date></dict></plist>`
		profileB64 := base64.StdEncoding.EncodeToString([]byte(plistContent))
		mockAppleAPIResponse := CreateAppleAPIResponse(`{"code":0,"data":{"id":"p1","profile":"` + profileB64 + `","type":"ios_app_inhouse","certificateId":"cert1"}}`)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                           // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                      // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil) // 查询 Bundle ID。
		httpMocker = httpMocker.ResponseOnce(mockAppleAPIResponse, nil)                            // HTTP 请求 Fastlane 代理申请企业内测描述文件。
		dbAppleCertificateMocker = dbAppleCertificateMocker.ScanOnce(func(v any) {}, nil)
		dbAppleProfileMocker = dbAppleProfileMocker.CreateOnce(nil) // 插入 Apple 描述文件记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)               // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer httpMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyInHouseProfileReq{
			BundleID: bundleID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name     string
		bundleID string
	}{
		{"BundleID 为空", ""},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID)
		})
	}
}

func TestAppleWebApplyCommonProfile(t *testing.T) {
	const reqPath = "/web/apple/applyCommonProfile"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		commonProfile := base64.StdEncoding.EncodeToString([]byte(`<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>AppIDName</key><string>test</string><key>UUID</key><string>p1</string><key>CreationDate</key><date>2025-01-01T00:00:00Z</date><key>ExpirationDate</key><date>2025-12-31T00:00:00Z</date></dict></plist>`))

		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		commonProfileReset := mvt.Chain(cfg.Get()).Elem().FieldByName("AppleConfiguration").FieldByName("CommonProfileValue").Set(commonProfile)
		certificateIDOfCommonProfileValueReset := mvt.Chain(cfg.Get()).Elem().FieldByName("AppleConfiguration").FieldByName("CertificateIDOfCommonProfileValue").Set("cert1")
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                             // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                 // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                            // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil) // 查询签名证书。
		dbAppleProfileMocker = dbAppleProfileMocker.CreateOnce(nil)                                      // 插入 Apple 描述文件记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                    // 保存应用事件。
		defer commonProfileReset.Reset()
		defer certificateIDOfCommonProfileValueReset.Reset()
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp[any](ctx, reqPath, AppInfo.AppID, nil)),
			consts.AlertSuccess,
		)
	})

	t.Run("异常测试_AppInfo 平台非 Apple", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp[any](ctx, reqPath, AppInfo.AppID, nil)),
			consts.ErrParameterInvalid,
		)
	})
}

func TestAppleWebApplyPushCertificate(t *testing.T) {
	const reqPath = "/web/apple/applyPushCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyPushCertificateReq{
			BundleID:    "cn.ivfzhou.test",
			Environment: 1,
		}
		mockAesKey := &model.AesKey{
			ID:          1,
			Secret:      util.RandomBytes(16),
			CreatedTime: time.Now(),
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                  // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                                      // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                                                                 // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.TakeOnce(&model.AppleBundleID{}, nil)                                                                   // 查询 Bundle ID 信息。
		httpMocker = httpMocker.ResponseOnce(CreateAppleAPIResponse(`{"code":0,"data":{"id":"pc1","certificate":"`+AppleCertificateCertDERBase64+`"}}`), nil) // HTTP 请求 Fastlane 代理申请 Push 证书。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                                                                            // Redis 生成唯一 ID。
		dbAesKeyMocker = dbAesKeyMocker.LastOnce(mockAesKey, nil)                                                                                             // 查询数据库中证书 AES 加密密钥。
		dbAppleCertificateMocker = dbAppleCertificateMocker.CreateOnce(nil)                                                                                   // 插入 Apple Push 证书记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                                                                         // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer httpMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID string, environment int) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebApplyPushCertificateReq{
			BundleID:    bundleID,
			Environment: environment,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		bundleID    string
		environment int
	}{
		{"BundleID 为空", "", 1},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65), 1},
		{"Environment 小于最小值", "cn.ivfzhou.test", 0},
		{"Environment 大于最大值", "cn.ivfzhou.test", 4},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID, v.environment)
		})
	}
}

func TestAppleWebDeleteBundleID(t *testing.T) {
	const reqPath = "/web/apple/deleteBundleID"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebDeleteBundleIDReq{
			BundleID: "cn.ivfzhou.test",
		}
		mockBundle := &model.AppleBundleID{
			ID:          1,
			BundleID:    "cn.ivfzhou.test",
			InAppleID:   "bid",
			Environment: model.AppleBundleIDTypeAppStore,
		}
		response := &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                           // 校验用户角色权限。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.TakeOnce(mockBundle, nil)         // 查询 Bundle ID 是否存在（存在）。
		dbAppleProfileMocker = dbAppleProfileMocker.CountOnce(0, nil)                   // 查询是否有描述文件引用（无）。
		dbAppleCertificateMocker = dbAppleCertificateMocker.CountOnce(0, nil)           // 查询是否有证书引用（无）。
		httpMocker = httpMocker.ResponseOnce(response, nil)                             // HTTP 请求 Apple API 删除 Bundle ID。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 保存应用事件。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.DeleteOnce(gen.ResultInfo{}, nil) // 删除 Bundle ID 记录。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, bundleID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebDeleteBundleIDReq{
			BundleID: bundleID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name     string
		bundleID string
	}{
		{"BundleID 为空", ""},
		{"BundleID 过长", util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.bundleID)
		})
	}
}

func TestAppleWebRemoveCertificate(t *testing.T) {
	const reqPath = "/web/apple/removeCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		reqObj := &protocol.AppleWebRemoveCertificateReq{
			CertificateID: certificateID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                               // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                               // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                 // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                  // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                             // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                            // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 软删除证书记录。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleCertificateMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebRemoveCertificateReq{
			CertificateID: certificateID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		certificateID string
	}{
		{"证书 ID 长度不正确", util.FastRandomAlphaNumberString(31)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.certificateID)
		})
	}
}

func TestAppleWebListAppCertificates(t *testing.T) {
	const reqPath = "/web/apple/listAppCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockAppProfiles := []*model.AppleProfile{
			{ProfileID: util.FastRandomAlphaNumberString(32), CertificateID: 1, BundleID: 1, UserID: LoginUser.ID},
		}
		mockAppCertificates := []*model.AppleCertificate{
			{CertificateID: util.FastRandomAlphaNumberString(32), Category: model.AppleCertificateCategoryPush, Owner: LoginUser.NameEn, BundleID: 1, UserID: LoginUser.ID},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                  // 校验用户角色权限。
		dbAppleProfileMocker = dbAppleProfileMocker.FindOnce(mockAppProfiles, nil)             // 查询应用描述文件列表。
		dbAppleCertificateMocker = dbAppleCertificateMocker.FindOnce(mockAppCertificates, nil) // 查询应用 Push 证书列表。
		dbAppleCertificateMocker = dbAppleCertificateMocker.FindOnce(nil, nil)                 // 查询签名证书列表。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.FindOnce(nil, nil)                       // 查询 Bundle ID 信息。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{LoginUser}, nil)                    // 查询用户信息。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListAppCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)
	})

	t.Run("异常测试_数据库错误", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)             // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)         // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                               // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                           // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                               // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                          // 校验用户角色权限。
		dbAppleProfileMocker = dbAppleProfileMocker.FindOnce(nil, gorm.ErrInvalidData) // 查询应用描述文件列表（数据库错误）。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleProfileMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListAppCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.ErrSystem,
		)
	})
}

func TestAppleWebSubmitSigningJob(t *testing.T) {
	const reqPath = "/web/apple/submitSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebSubmitSigningJobReq{
			ProfileID: util.FastRandomAlphaNumberString(32),
			FileID:    util.FastRandomAlphaNumberString(38),
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                  // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                   // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                              // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                  // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                                             // 校验用户角色权限。
		dbAppleProfileMocker = dbAppleProfileMocker.LastOnce(&model.AppleProfile{Type: model.AppleProfileTypeIOSAppInHouse}, nil)         // 查询描述文件信息（InHouse 类型跳过 Bundle ID 查询）。
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{AppID: AppInfo.ID, UserID: LoginUser.ID, Type: model.FileTypeAppleSigning}, nil) // 查询文件信息。
		rabbitMQMocker := MockRabbitMQClient(ctx)
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)       // 发送签名任务消息到消息队列。
		redisMocker = redisMocker.SAddOnce(1, nil)                        // Redis 生成唯一 ID。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.CreateOnce(nil) // 插入 Apple 签名任务记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                     // 保存应用事件。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbFileMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()
		defer rabbitMQMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, profileID, fileID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebSubmitSigningJobReq{
			ProfileID: profileID,
			FileID:    fileID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		profileID string
		fileID    string
	}{
		{"ProfileID 长度不正确", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"FileID 长度不正确", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.profileID, v.fileID)
		})
	}
}

func TestAppleWebListSigningJobs(t *testing.T) {
	const reqPath = "/web/apple/listSigningJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockTableNames := []string{"t_apple_signing_job"}
		mockSigningJobs := []*model.AppleSigningJob{
			{JobID: util.FastRandomAlphaNumberString(38), Status: 1, ProfileID: 1, UserID: LoginUser.ID, FileID: util.FastRandomAlphaNumberString(38)},
		}
		const count int = 1

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                          // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                            // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                             // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                        // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                            // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                                                       // 校验用户角色权限。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobGetTablesOnce(mockTableNames, nil)                                         // 获取签名任务分表名。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobCount2Once(count, nil)                                                     // 统计签名任务总数。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobListOnce(mockSigningJobs, nil)                                             // 分页查询签名任务列表。
		dbAppleProfileMocker = dbAppleProfileMocker.FindOnce([]*model.AppleProfile{{ID: 1, ProfileID: mockSigningJobs[0].JobID, BundleID: 1}}, nil) // 查询描述文件信息。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.FindOnce(nil, nil)                                                                            // 查询 Bundle ID 信息。
		dbFileMocker = dbFileMocker.FindOnce(nil, nil)                                                                                              // 查询文件信息。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{LoginUser}, nil)                                                                         // 查询用户信息。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppleWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppleWebListSigningJobsReq{
				PageNumber: 1,
				PageSize:   10,
			})),
			0,
		)
		if rspBodyObj.Data.Count != count {
			t.Errorf("expect count %d, but got %d", count, rspBodyObj.Data.Count)
		}
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppleWebListSigningJobsReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name       string
		pageNumber int
		pageSize   int
	}{
		{"页号小于最小值", 0, 10},
		{"每页条数小于最小值", 1, 0},
		{"每页条数大于最大值", 1, 101},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.pageNumber, v.pageSize)
		})
	}
}

func TestAppleWebRemoveProfile(t *testing.T) {
	const reqPath = "/web/apple/removeProfile"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		profileID := util.FastRandomAlphaNumberString(32)
		reqObj := &protocol.AppleWebRemoveProfileReq{
			ProfileID: profileID,
		}
		response := &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}
		mockProfile := &model.AppleProfile{
			InAppleID: "p1",
			Type:      model.AppleProfileTypeIOSAppStore,
			BundleID:  1,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                       // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                     // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                         // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                    // 校验用户角色权限。
		dbAppleProfileMocker = dbAppleProfileMocker.TakeOnce(mockProfile, nil)                                   // 查询描述文件信息。
		httpMocker = httpMocker.ResponseOnce(response, nil)                                                      // HTTP 请求 Apple API 删除描述文件。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                            // 保存应用事件。
		dbAppleProfileMocker = dbAppleProfileMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 软删除描述文件记录。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.ScanOnce(func(v any) {                                     // 查询 Bundle ID。
			vv := v.(*string)
			*vv = "cn.ivfzhou.test"
		}, nil)
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, profileID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebRemoveProfileReq{
			ProfileID: profileID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		profileID string
	}{
		{"ProfileID 长度不正确", util.FastRandomAlphaNumberString(31)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.profileID)
		})
	}
}

func TestAppleWebRemovePushCertificate(t *testing.T) {
	const reqPath = "/web/apple/removePushCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		reqObj := &protocol.AppleWebRemovePushCertificateReq{
			CertificateID: certificateID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAppleBundleIDMocker := MockDBClient[model.AppleBundleID](ctx)
		httpMocker := MockHTTPClient(ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                               // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                               // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                 // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                  // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                             // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                 // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                            // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.TakeOnce(&model.AppleCertificate{}, nil)                     // 查询 Push 证书信息。
		httpMocker = httpMocker.ResponseOnce(CreateAppleAPIResponse(`{"code":0,"data":null}`), nil)                      // HTTP 请求 Fastlane 代理删除 Push 证书。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                                    // 保存应用事件。
		dbAppleCertificateMocker = dbAppleCertificateMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 软删除 Push 证书记录。
		dbAppleBundleIDMocker = dbAppleBundleIDMocker.TakeOnce(&model.AppleBundleID{}, nil)                              // 查询 Bundle ID 信息。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAppleBundleIDMocker.Reset()
		defer httpMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebRemovePushCertificateReq{
			CertificateID: certificateID,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		certificateID string
	}{
		{"证书 ID 长度不正确", util.FastRandomAlphaNumberString(31)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.certificateID)
		})
	}
}

func TestAppleWebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/apple/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		reqObj := &protocol.AppleWebDownloadCertificateReq{
			CertificateID: certificateID,
			Type:          protocol.AppleFileTypeSigningCertificate,
		}
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("test file content"))
		mockAesKey := &model.AesKey{Secret: secret}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                         // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                             // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                        // 校验用户角色权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.TakeOnce(&model.AppleCertificate{Content: encrypt}, nil) // 查询证书信息。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                                                    // 查询数据库解密密钥。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()

		_, _ = CheckAndReadBody(
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string, fileType int) {
		ctx := context.Background()
		reqObj := &protocol.AppleWebDownloadCertificateReq{
			CertificateID: certificateID,
			Type:          fileType,
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验用户角色权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, reqObj)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		certificateID string
		fileType      int
	}{
		{"证书 ID 长度不正确", util.FastRandomAlphaNumberString(31), protocol.AppleFileTypeSigningCertificate},
		{"文件类型小于最小值", util.FastRandomAlphaNumberString(32), 0},
		{"文件类型大于最大值", util.FastRandomAlphaNumberString(32), 4},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.certificateID, v.fileType)
		})
	}
}

func TestAppleWebStatisticSigningTimes(t *testing.T) {
	const reqPath = "/web/apple/statisticSigningTimes"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_apple_signing_job"}                      // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"count": 1, "day": "20260710"}} // 模拟按天统计的签名任务数量数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                     // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                    // 查询数据库应用 ID。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobGetTablesOnce(mockTableNames, nil)       // 获取苹果签名任务表名。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobCountWithDayOnce(mockStatisticData, nil) // 按天统计签名任务数量。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningTimesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningTimesReq{
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

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningTimesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningTimesReq{
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

func TestAppleWebStatisticSigningCost(t *testing.T) {
	const reqPath = "/web/apple/statisticSigningCost"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_apple_signing_job"}                     // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"cost": 1, "day": "20260710"}} // 模拟按天统计的签名任务耗时数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                       // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                     // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                    // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                   // 查询数据库应用 ID。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobGetTablesOnce(mockTableNames, nil)      // 获取苹果签名任务表名。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobCostWithDayOnce(mockStatisticData, nil) // 按天统计签名任务耗时。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningCostRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningCostReq{
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

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningCostRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningCostReq{
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

func TestAppleWebStatisticSigningPassRate(t *testing.T) {
	const reqPath = "/web/apple/statisticSigningPassRate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		appID := util.FastRandomAlphaNumberString(32)
		beginTime := time.Now()
		endTime := time.Now().AddDate(0, 10, 0)
		timeStep := protocol.TimeStepDay
		mockTableNames := []string{"t_apple_signing_job"}                     // 模拟签名任务分表名列表。
		mockStatisticData := []map[string]any{{"rate": 1, "day": "20260710"}} // 模拟按天统计的签名任务通过率数据。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                             // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                              // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                         // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                        // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                       // 查询数据库应用 ID。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobGetTablesOnce(mockTableNames, nil)          // 获取苹果签名任务表名。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.AppleSigningJobPassRateWithDayOnce(mockStatisticData, nil) // 按天统计签名任务通过率。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningPassRateRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningPassRateReq{
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

		CheckAndUnmarshalBody[protocol.AppleWebStatisticSigningPassRateRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.AppleWebStatisticSigningPassRateReq{
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
