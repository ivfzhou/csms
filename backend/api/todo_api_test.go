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

func TestTodoWebCount(t *testing.T) {
	const reqPath = "/web/todo/count"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.CountOnce(1, nil)                                       // 统计待处理待办数。
		dbTodoMocker = dbTodoMocker.CountOnce(2, nil)                                       // 统计已处理待办数。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbTodoMocker.Reset()

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

func TestTodoWebList(t *testing.T) {
	const reqPath = "/web/todo/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		pageNumber, pageSize := 1, 10
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟数据库中的待办列表数据（空结构体，仅占位）。
		mockTodoList := []*model.Todo{{}, {}}
		// 模拟数据库中的应用列表数据（空结构体，仅占位）。
		mockAppList := []*model.App{{}}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询申请人英文名。
		dbTodoMocker = dbTodoMocker.CountOnce(1, nil)                                       // 统计待处理待办总数。
		dbTodoMocker = dbTodoMocker.FindOnce(mockTodoList, nil)                             // 分页查询待处理待办。
		dbAppMocker = dbAppMocker.FindOnce(mockAppList, nil)                                // 批量查询应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbTodoMocker.Reset()
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

func TestTodoWebListDealt(t *testing.T) {
	const reqPath = "/web/todo/listDealt"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		// 查询已处理待办列表请求参数。
		dealAppID := AppInfo.AppID
		dealTypes := []int{model.TodoTypeActivateApp}
		dealStatus := []int{model.TodoStatusApproved}
		dealPageNumber := 1
		dealPageSize := 10
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}}
		// 模拟数据库中的待办列表数据（空结构体，仅占位）。
		mockTodoList := []*model.Todo{{}, {}}
		// 模拟数据库中的应用列表数据（空结构体，仅占位）。
		mockAppList := []*model.App{{}}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                               // 校验系统管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询申请人英文名。
		dbTodoMocker = dbTodoMocker.CountOnce(1, nil)                                       // 统计已处理待办总数。
		dbTodoMocker = dbTodoMocker.FindOnce(mockTodoList, nil)                             // 分页查询已处理待办。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = AppInfo.ID }, nil)     // 查询关联应用 IDs。
		dbAppMocker = dbAppMocker.FindOnce(mockAppList, nil)                                // 批量查询应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListDealtRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListDealtReq{
				AppID:      dealAppID,
				Types:      dealTypes,
				Status:     dealStatus,
				PageNumber: dealPageNumber,
				PageSize:   dealPageSize,
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

func TestTodoWebCreate(t *testing.T) {
	const reqPath = "/web/todo/create"

	t.Run("正常测试_加入应用", func(t *testing.T) {
		ctx := context.Background()
		// 创建待办请求参数。
		createType := model.TodoTypeJoinApp
		createAppID := AppInfo.AppID
		createReason := "~"

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                 // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                     // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{} }, nil)             // 查询应用已有成员角色。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{LoginUser.ID} }, nil) // 查询应用管理员。
		dbTodoMocker = dbTodoMocker.CountOnce(0, nil)                                                        // 检查是否已有待处理申请。
		dbTodoMocker = dbTodoMocker.CreateOnce(nil)                                                          // 创建加入应用待办。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        createType,
				AppID:       createAppID,
				ApplyReason: createReason,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
		ctx := context.Background()
		// 创建待办请求参数。
		createType := model.TodoTypeApplySigner
		createAppID := AppInfo.AppID
		createReason := "~"

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                     // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                 // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                     // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                // 校验是否为应用成员。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{LoginUser.ID} }, nil) // 查询应用管理员。
		dbTodoMocker = dbTodoMocker.CountOnce(0, nil)                                                        // 检查是否已有待处理申请。
		dbTodoMocker = dbTodoMocker.CreateOnce(nil)                                                          // 创建签名权限申请待办。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        createType,
				AppID:       createAppID,
				ApplyReason: createReason,
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

func TestTodoWebGetDetail(t *testing.T) {
	const reqPath = "/web/todo/getDetail"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		// 查询待办详情请求参数。
		detailID := 1
		// 模拟数据库中的用户列表数据（空结构体，仅占位）。
		mockUserList := []*model.User{{}, {}}
		// 模拟数据库中的待办记录。
		mockTodo := &model.Todo{ApproverID: LoginUser.ID}
		// 模拟数据库中的应用记录（空结构体，仅占位）。
		mockApp := &model.App{}
		// 模拟数据库中的苹果设备记录（空结构体，仅占位）。
		mockDevice := &model.AppleDevice{}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAppleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询候选人英文名。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbAppMocker = dbAppMocker.TakeOnce(mockApp, nil)                                    // 查询关联应用信息。
		dbAppleDeviceMocker = dbAppleDeviceMocker.TakeOnce(mockDevice, nil)                 // 查询关联苹果设备信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAppleDeviceMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebGetDetailReq{ID: detailID})),
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

func TestTodoWebDeal(t *testing.T) {
	const reqPath = "/web/todo/deal"

	t.Run("正常测试_注册应用", func(t *testing.T) {
		ctx := context.Background()
		// 处理待办请求参数。
		dealID := 1
		dealIsPass := true
		dealMessage := "~"
		// 模拟数据库中的待办记录（注册应用类型）。
		mockTodo := &model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterApp}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)           // 更新待办为已处理。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)             // 更新应用状态为已注册。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             dealID,
				IsPass:         dealIsPass,
				ApproveMessage: dealMessage,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_加入应用", func(t *testing.T) {
		ctx := context.Background()
		// 处理待办请求参数。
		dealID := 1
		dealIsPass := true
		dealMessage := "~"
		// 模拟数据库中的待办记录（加入应用类型）。
		mockTodo := &model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeJoinApp}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbUserRoleMocker = dbUserRoleMocker.CreateOnce(nil)                                 // 创建应用成员角色。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)           // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             dealID,
				IsPass:         dealIsPass,
				ApproveMessage: dealMessage,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
		ctx := context.Background()
		// 处理待办请求参数。
		dealID := 1
		dealIsPass := true
		dealMessage := "~"
		// 模拟数据库中的待办记录（申请签名权限类型）。
		mockTodo := &model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeApplySigner}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbUserRoleMocker = dbUserRoleMocker.CreateOnce(nil)                                 // 创建成员签名权限角色。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)           // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             dealID,
				IsPass:         dealIsPass,
				ApproveMessage: dealMessage,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_启用应用", func(t *testing.T) {
		ctx := context.Background()
		// 处理待办请求参数。
		dealID := 1
		dealIsPass := true
		dealMessage := "~"
		// 模拟数据库中的待办记录（启用应用类型）。
		mockTodo := &model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeActivateApp}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)             // 更新应用状态为已启用。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)           // 更新待办为已处理。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             dealID,
				IsPass:         dealIsPass,
				ApproveMessage: dealMessage,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_注册设备", func(t *testing.T) {
		ctx := context.Background()
		// 处理待办请求参数。
		dealID := 1
		dealIsPass := true
		dealMessage := "~"
		// 模拟数据库中的待办记录（注册苹果设备类型）。
		mockTodo := &model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterAppleDevice}
		// 模拟数据库中的苹果设备记录（空结构体，仅占位）。
		mockDevice := &model.AppleDevice{}
		// 模拟数据库操作影响行数。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1}

		key, _, err := GenerateECDSAKeyPEM("P256")
		if err != nil {
			t.Fatal(err)
		}
		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("AppleAPIConfiguration").
			FieldByName("SecretValue").
			Set(key).
			Reset()
		httpMocker := MockHTTPClient(ctx)
		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppleDeviceMocker := MockDBClient[model.AppleDevice](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		dbTodoMocker = dbTodoMocker.TakeOnce(mockTodo, nil)                                 // 查询待办详情。
		dbAppleDeviceMocker = dbAppleDeviceMocker.TakeOnce(mockDevice, nil)                 // 查询苹果设备信息。
		httpMocker = httpMocker.ResponseOnce(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader("{}")),
		}, nil) // 调用苹果设备注册 API。
		dbAppleDeviceMocker = dbAppleDeviceMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil) // 更新设备状态。
		dbUserMocker = dbUserMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)               // 更新用户设备信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                              // 缓存设备 ID 到 Redis。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)               // 更新待办为已处理。
		defer httpMocker.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppleDeviceMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             dealID,
				IsPass:         dealIsPass,
				ApproveMessage: dealMessage,
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
