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
	"strings"
	"testing"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestTodoAPI_WebCount(t *testing.T) {
	const reqPath = "/web/todo/count"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		todoMocker = todoMocker.CountOnce(1, nil)                                           // 统计待处理待办数。
		todoMocker = todoMocker.CountOnce(2, nil)                                           // 统计已处理待办数。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer todoMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.TodoWebCountRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.Done != 2 {
			t.Errorf("want 2, but got %v", rspBodyObj.Data.Done)
		}
		if rspBodyObj.Data.NeedToDeal != 1 {
			t.Errorf("want 1, but got %v", rspBodyObj.Data.Done)
		}
	})
}

func TestTodoAPI_WebList(t *testing.T) {
	const reqPath = "/web/todo/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		pageNumber, pageSize := 1, 10

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                    // 查询申请人英文名。
		todoMocker = todoMocker.CountOnce(1, nil)                                           // 统计待处理待办总数。
		todoMocker = todoMocker.FindOnce([]*model.Todo{{}, {}}, nil)                        // 分页查询待处理待办。
		dbAppMocker = dbAppMocker.FindOnce([]*model.App{{}}, nil)                           // 批量查询应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer todoMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageNumber, PageSize int
	}{
		{"每页条数错误", 1, 0},
		{"页数错误", 0, 10},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageNumber, v.PageSize)
		})
	}
}

func TestTodoAPI_WebListDealt(t *testing.T) {
	const reqPath = "/web/todo/listDealt"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                   // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}}, nil)                        // 查询申请人英文名。
		todoMocker = todoMocker.CountOnce(1, nil)                                           // 统计已处理待办总数。
		todoMocker = todoMocker.FindOnce([]*model.Todo{{}, {}}, nil)                        // 分页查询已处理待办。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = AppInfo.ID }, nil)     // 查询关联应用 IDs。
		dbAppMocker = dbAppMocker.FindOnce([]*model.App{{}}, nil)                           // 批量查询应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer todoMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListDealtRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListDealtReq{
				AppID:      AppInfo.AppID,
				Types:      []int{model.TodoTypeActivateApp},
				Status:     []int{model.TodoStatusApproved},
				PageNumber: 1,
				PageSize:   10,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, types, status []int, pageNumber, pageSize int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListDealtRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListDealtReq{
				AppID:      appID,
				Types:      types,
				Status:     status,
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		AppID                string
		Types, Status        []int
		PageNumber, PageSize int
	}{
		{"应用 ID 错误", "汉", []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 1, 10},
		{"待办类型错误", AppInfo.AppID, []int{-1}, []int{model.TodoStatusApproved}, 1, 10},
		{"待办状态错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{1}, 1, 10},
		{"每页条数错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 1, 0},
		{"页数错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 0, 10},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AppID, v.Types, v.Status, v.PageNumber, v.PageSize)
		})
	}
}

func TestTodoAPI_WebCreate(t *testing.T) {
	const reqPath = "/web/todo/create"

	t.Run("正常测试_加入应用", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                  // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                      // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.FindOnce([]*model.UserRole{{}, {}}, nil)              // 查询应用已有成员角色。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询应用管理员。
		todoMocker = todoMocker.CountOnce(0, nil)                                             // 检查是否已有待处理申请。
		todoMocker = todoMocker.CreateOnce(nil)                                               // 创建加入应用待办。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        model.TodoTypeJoinApp,
				AppID:       AppInfo.AppID,
				ApplyReason: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                  // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                      // 查询数据库应用信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                     // 校验是否为应用成员。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询应用管理员。
		todoMocker = todoMocker.CountOnce(0, nil)                                             // 检查是否已有待处理申请。
		todoMocker = todoMocker.CreateOnce(nil)                                               // 创建签名权限申请待办。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer userRoleMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        model.TodoTypeApplySigner,
				AppID:       AppInfo.AppID,
				ApplyReason: "~",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ int, appID, applyReason string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        typ,
				AppID:       appID,
				ApplyReason: applyReason,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name               string
		Type               int
		AppID, ApplyReason string
	}{
		{"类型错误", 0, AppInfo.AppID, "~"},
		{"应用 ID 缺失", model.TodoTypeJoinApp, "", "~"},
		{"应用 ID 非法", model.TodoTypeJoinApp, "汉" + util.FastRandomAlphaNumberString(31), "~"},
		{"应用 ID 错误", model.TodoTypeJoinApp, util.FastRandomAlphaNumberString(33), "~"},
		{"理由缺失", model.TodoTypeJoinApp, AppInfo.AppID, ""},
		{"理由过长", model.TodoTypeJoinApp, AppInfo.AppID, util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.AppID, v.ApplyReason)
		})
	}
}

func TestTodoAPI_WebDetail(t *testing.T) {
	const reqPath = "/web/todo/getDetail"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		appleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}}, nil)                    // 查询候选人英文名。
		todoMocker = todoMocker.TakeOnce(&model.Todo{ApproverID: LoginUser.ID}, nil)        // 查询待办详情。
		dbAppMocker = dbAppMocker.TakeOnce(&model.App{}, nil)                               // 查询关联应用信息。
		appleDeviceMocker = appleDeviceMocker.TakeOnce(&model.AppleDevice{}, nil)           // 查询关联苹果设备信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer todoMocker.Reset()
		defer dbAppMocker.Reset()
		defer appleDeviceMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebGetDetailReq{ID: 1})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebGetDetailReq{
				ID: id,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID 错误", 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}

func TestTodoAPI_WebDeal(t *testing.T) {
	const reqPath = "/web/todo/deal"

	t.Run("正常测试_注册应用", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                     // 查询数据库登录用户信息。
		todoMocker = todoMocker.TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterApp}, nil) // 查询待办详情。
		todoMocker = todoMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                     // 更新待办为已处理。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                   // 更新应用状态为已注册。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer todoMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_加入应用", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                 // 查询数据库登录用户信息。
		todoMocker = todoMocker.TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeJoinApp}, nil) // 查询待办详情。
		userRoleMocker = userRoleMocker.CreateOnce(nil)                                                                                                      // 创建应用成员角色。
		todoMocker = todoMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                 // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                     // 查询数据库登录用户信息。
		todoMocker = todoMocker.TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeApplySigner}, nil) // 查询待办详情。
		userRoleMocker = userRoleMocker.CreateOnce(nil)                                                                                                          // 创建成员签名权限角色。
		todoMocker = todoMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                     // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer userRoleMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_启用应用", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                         // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                          // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                     // 查询数据库登录用户信息。
		todoMocker = todoMocker.TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeActivateApp}, nil) // 查询待办详情。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                   // 更新应用状态为已启用。
		todoMocker = todoMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)                                                                     // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_注册设备", func(t *testing.T) {
		ctx := context.Background()

		key, _, err := GenerateECDSAKeyPEM("P256")
		if err != nil {
			t.Fatal(err)
		}
		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("AppleAPI").
			FieldByName("Secret").
			Set(key).
			Reset()

		httpMocker := MockHTTPClient(ctx)
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		appleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		todoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                                                              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                                                 // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                                                                  // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                                                             // 查询数据库登录用户信息。
		todoMocker = todoMocker.TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterAppleDevice}, nil) // 查询待办详情。
		appleDeviceMocker = appleDeviceMocker.TakeOnce(&model.AppleDevice{}, nil)                                                                                        // 查询苹果设备信息。
		httpMocker = httpMocker.ResponseOnce(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader("{}")),
		}, nil) // 调用苹果设备注册 API。
		appleDeviceMocker = appleDeviceMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 更新设备状态。
		dbUserMocker = dbUserMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)           // 更新用户设备信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                         // 缓存设备 ID 到 Redis。
		todoMocker = todoMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)               // 更新待办为已处理。
		defer httpMocker.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer appleDeviceMocker.Reset()
		defer todoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             id,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID 错误", 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}
