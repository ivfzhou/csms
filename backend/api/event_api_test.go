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
	"time"

	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func TestEventAPI_WebList(t *testing.T) {
	const reqPath = "/web/event/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                             // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                            // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(0, nil)                                                               // 校验是否为系统管理员（非管理员则走成员逻辑）。
		userRoleMocker = userRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                           // 查询用户有权限的应用 IDs。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                                 // 查询对应应用 IDs。
		dbAppMocker = dbAppMocker.FindOnce([]*model.App{{}, {}, {}}, nil)                                               // 批量查询应用基本信息。
		dbUserMocker = dbUserMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil)                               // 查询事件关联的操作人 IDs。
		dbUserMocker = dbUserMocker.FindOnce([]*model.User{{}, {}, {}}, nil)                                            // 批量查询操作人信息。
		eventMocker = eventMocker.EventGetTablesOnce([]string{"t_event"}, nil)                                          // 获取事件分表名。
		eventMocker = eventMocker.EventCount2Once(1, nil)                                                               // 统计事件总数。
		eventMocker = eventMocker.EventListOnce([]*model.Event{{Content: "{}"}, {Content: "{}"}, {Content: "{}"}}, nil) // 分页查询事件列表。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer eventMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.EventWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebListReq{
				App:        util.FastRandomAlphaNumberString(32),
				Platform:   model.AppPlatformAndroid,
				Source:     model.SourceWeb,
				Type:       model.EventTypeApplyOpenAPI,
				User:       "zhangsan",
				BeginTime:  time.Now(),
				EndTime:    time.Now().Add(time.Hour),
				PageSize:   10,
				PageNumber: 1,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, app string, platform, source, typ int, user string,
		beginTime, endTime time.Time, pageSize, pageNumber int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                     // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[protocol.EventWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebListReq{
				App:        app,
				Platform:   platform,
				Source:     source,
				Type:       typ,
				User:       user,
				BeginTime:  beginTime,
				EndTime:    endTime,
				PageSize:   pageSize,
				PageNumber: pageNumber,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                  string
		app                   string
		platform, source, typ int
		user                  string
		beginTime, endTime    time.Time
		pageSize, pageNumber  int
	}{
		{"应用名过长", util.FastRandomAlphaNumberString(65), model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"应用平台非法", "应用名", -1, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件来源非法", "应用名", model.AppPlatformAndroid, -1, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件类型非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, -1, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"操作人名过长", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, util.FastRandomAlphaNumberString(33), time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"操作人名字符非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "a" + string([]byte{128}), time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件开始时间缺失", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Time{}, time.Now().Add(time.Hour), 10, 1},
		{"事件结束时间缺失", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Time{}, 10, 1},
		{"事件结束时间早于开始时间", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(-time.Hour), 10, 1},
		{"每页条数非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 0, 1},
		{"页数非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(
				t, v.app, v.platform, v.source, v.typ, v.user, v.beginTime, v.endTime, v.pageSize, v.pageNumber)
		})
	}
}

func TestEventAPI_EventWebStatistic(t *testing.T) {
	const reqPath = "/web/event/statistic"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		eventMocker := MockDBClient[model.Event](ctx)
		userRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                        // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                                         // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                                    // 查询数据库登录用户信息。
		userRoleMocker = userRoleMocker.CountOnce(1, nil)                                                                       // 校验系统管理员权限。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*int) = 1 }, nil)                                                  // 查询用户有权限的应用数量。
		eventMocker = eventMocker.EventGetTablesOnce([]string{"t_event"}, nil)                                                  // 获取事件分表名。
		eventMocker = eventMocker.EventCountTypesWithDayOnce([]map[string]any{{"count": 1, "type": 1, "day": "20260710"}}, nil) // 按天统计各类事件数量。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer eventMocker.Reset()
		defer userRoleMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.EventWebStatisticRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebStatisticReq{
				AppID:     util.FastRandomAlphaNumberString(32),
				BeginTime: time.Now(),
				EndTime:   time.Now().AddDate(0, 10, 0),
				TimeStep:  protocol.TimeStepDay,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, timeStep int, beginTime, endTime time.Time) {
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

		CheckAndUnmarshalBody[protocol.EventWebStatisticRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebStatisticReq{
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
		{"结束时间大于开始时间", util.FastRandomAlphaNumberString(32), protocol.TimeStepWeek, time.Now(), time.Now().Add(-time.Hour)},
		{"时间步长非法", util.FastRandomAlphaNumberString(32), 0, time.Now(), time.Now().Add(time.Hour)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.appID, v.timeStep, v.beginTime, v.endTime)
		})
	}
}
