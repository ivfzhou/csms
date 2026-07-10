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
	"testing"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func TestOpenAPI_WebApply(t *testing.T) {
	const reqPath = "/web/open/apply"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		apiAuthMocker := MockDBClient[model.APIAuthorization](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验应用管理员权限。
		apiAccountMocker = apiAccountMocker.TakeOnce(nil, nil)                              // 检查凭证 ID 是否已存在。
		apiAccountMocker = apiAccountMocker.CreateOnce(nil)                                 // 创建 API 凭证。
		apiAuthMocker = apiAuthMocker.CreateInBatchesOnce(nil)                              // 创建 API 授权记录。
		eventMocker = eventMocker.CreateOnce(nil)                                           // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer apiAuthMocker.Reset()
		defer eventMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebApplyReq{
				AccountID:   "a" + util.FastRandomAlphaNumberString(5),
				Authorities: []int{model.CapabilityGetWindowsOVCertificatePassword},
				RequestIP:   "127.0.0.1",
				Frequency:   100,
			})),
			0,
		)

		if len(rspBodyObj.Data.Password) <= 0 {
			t.Errorf("password is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, accountID string, authorities []int, requestIP string, frequency int) {
		ctx := context.Background()

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

		CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebApplyReq{
				AccountID:   accountID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		AccountID   string
		Authorities []int
		RequestIP   string
		Frequency   int
	}{
		{"凭证 ID 缺失", "", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证 ID 字符非法", "1a", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项重复", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword, model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项非法", "authId", []int{-1}, "127.0.0.1", 100},
		{"请求源缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "", 100},
		{"请求源过多", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, util.FastRandomAlphaNumberString(257), 100},
		{"请求频率缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 0},
		{"请求频率过大", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 99999},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID, v.Authorities, v.RequestIP, v.Frequency)
		})
	}
}

func TestOpenAPI_WebUpdate(t *testing.T) {
	const reqPath = "/web/open/update"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		apiAuthMocker := MockDBClient[model.APIAuthorization](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		apiAccountMocker = apiAccountMocker.TakeOnce(&model.APIAccount{}, nil)                                     // 查询 API 凭证。
		apiAccountMocker = apiAccountMocker.UpdateColumnSimpleOnce(gen.ResultInfo{}, nil)                          // 更新 API 凭证信息。
		apiAuthMocker = apiAuthMocker.DeleteOnce(gen.ResultInfo{}, nil)                                            // 删除旧授权记录。
		apiAuthMocker = apiAuthMocker.CreateInBatchesOnce(nil)                                                     // 创建新授权记录。
		eventMocker = eventMocker.CreateOnce(nil)                                                                  // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer apiAuthMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebUpdateReq{
				AccountID:   "a" + util.FastRandomAlphaNumberString(5),
				Authorities: []int{model.CapabilityGetWindowsOVCertificatePassword},
				RequestIP:   "127.0.0.1",
				Frequency:   100,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string, authorities []int, requestIP string, frequency int) {
		ctx := context.Background()

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

		CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebUpdateReq{
				AccountID:   accountID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		AccountID   string
		Authorities []int
		RequestIP   string
		Frequency   int
	}{
		{"凭证 ID 缺失", "", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证 ID 字符非法", "1a", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项重复", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword, model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项非法", "authId", []int{-1}, "127.0.0.1", 100},
		{"请求源缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "", 100},
		{"请求源过多", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, util.FastRandomAlphaNumberString(257), 100},
		{"请求频率缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 0},
		{"请求频率过大", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 99999},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID, v.Authorities, v.RequestIP, v.Frequency)
		})
	}
}

func TestOpenAPI_WebGetInformation(t *testing.T) {
	const reqPath = "/web/open/getInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		apiAuthMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                              // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil) // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.ScanOnce(func(v any) { *v.(*string) = "" }, nil)                                  // 查询创建人英文名。
		apiAccountMocker = apiAccountMocker.TakeOnce(&model.APIAccount{}, nil)                                        // 查询 API 凭证详情。
		apiAuthMocker = apiAuthMocker.FindOnce([]*model.APIAuthorization{{}}, nil)                                    // 查询 API 授权列表。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer apiAuthMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.OpenWebGetInformationReq{AccountID: "a" + util.FastRandomAlphaNumberString(5)})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                              // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil) // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebGetInformationReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证 ID 缺失", ""},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65)},
		{"凭证 ID 字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebList(t *testing.T) {
	const reqPath = "/web/open/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		apiAuthMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil) // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{}, nil)                                                    // 查询创建人英文名。
		apiAccountMocker = apiAccountMocker.CountOnce(1, nil)                                                         // 统计 API 凭证总数。
		apiAccountMocker = apiAccountMocker.FindOnce([]*model.APIAccount{{}}, nil)                                    // 分页查询 API 凭证。
		apiAuthMocker = apiAuthMocker.ScanOnce(func(v any) {}, nil)                                                   // 查询 API 授权信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer apiAuthMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebListReq{
				PageNumber: 1,
				PageSize:   10,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, pageSize, pageNumber int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                               // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                          // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                              // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil) // 校验系统管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageSize, PageNumber int
	}{
		{"页数非法", 10, 0},
		{"每页条数非法", 0, 1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageSize, v.PageNumber)
		})
	}
}

func TestOpenAPI_WebRenewal(t *testing.T) {
	const reqPath = "/web/open/renewal"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                            // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                       // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil) // 校验应用管理员权限。
		apiAccountMocker = apiAccountMocker.TakeOnce(&model.APIAccount{}, nil)                                     // 查询 API 凭证。
		apiAccountMocker = apiAccountMocker.UpdateColumnSimpleOnce(gen.ResultInfo{}, nil)                          // 更新凭证有效期。
		eventMocker = eventMocker.CreateOnce(nil)                                                                  // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.OpenWebRenewalReq{AccountID: "a" + util.FastRandomAlphaNumberString(5)})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
		ctx := context.Background()

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
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRenewalReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证 ID 缺失", ""},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65)},
		{"凭证 ID 字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebReset(t *testing.T) {
	const reqPath = "/web/open/reset"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                 // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                  // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                             // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                 // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                // 校验应用管理员权限。
		apiAccountMocker = apiAccountMocker.TakeOnce(&model.APIAccount{}, nil)                           // 查询 API 凭证。
		apiAccountMocker = apiAccountMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 重置凭证密钥。
		eventMocker = eventMocker.CreateOnce(nil)                                                        // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer eventMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.OpenWebResetRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.OpenWebResetReq{AccountID: "a" + util.FastRandomAlphaNumberString(5)})),
			0,
		)

		if len(rspBodyObj.Data.Password) <= 0 {
			t.Errorf("password is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebResetReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证 ID 缺失", ""},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65)},
		{"凭证 ID 字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebRemove(t *testing.T) {
	const reqPath = "/web/open/remove"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		apiAccountMocker := MockDBClient[model.APIAccount](ctx)
		apiAuthMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                 // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                     // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                    // 校验应用管理员权限。
		apiAccountMocker = apiAccountMocker.TakeOnce(&model.APIAccount{}, nil)               // 查询 API 凭证。
		apiAuthMocker = apiAuthMocker.ScanOnce(func(v any) {}, nil)                          // 查询关联授权记录。
		apiAuthMocker = apiAuthMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil)       // 删除关联授权记录。
		apiAccountMocker = apiAccountMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 删除 API 凭证。
		eventMocker = eventMocker.CreateOnce(nil)                                            // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer apiAccountMocker.Reset()
		defer apiAuthMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer eventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID,
				protocol.OpenWebRemoveReq{AccountID: "a" + util.FastRandomAlphaNumberString(5)})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRemoveReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证 ID 缺失", ""},
		{"凭证 ID 过长", util.FastRandomAlphaNumberString(65)},
		{"凭证 ID 字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}
