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

	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func TestNoticeWebLast(t *testing.T) {
	const reqPath = "/web/notice/last"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockNotice := &model.Notice{
			ID:            1,
			Content:       "测试公告内容",
			UserID:        1,
			CreatedTime:   time.Now(),
			ExpiredTime:   time.Now().Add(24 * time.Hour),
			ActivatedTime: time.Now().Add(-24 * time.Hour),
		}

		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		dbNoticeMocker = dbNoticeMocker.TakeOnce(mockNotice, nil)              // 查询数据库中活跃的公告。
		defer dbNoticeMocker.Reset()
		defer redisMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.NoticeWebLastRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.Message != mockNotice.Content {
			t.Errorf("expect message %q, but got %q", mockNotice.Content, rspBodyObj.Data.Message)
		}
	})

	t.Run("正常测试_无活跃通知", func(t *testing.T) {
		ctx := context.Background()

		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		dbNoticeMocker = dbNoticeMocker.TakeOnce(nil, gorm.ErrRecordNotFound)  // 查询数据库中活跃的公告（无数据）。
		defer dbNoticeMocker.Reset()
		defer redisMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.NoticeWebLastRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.Message != "" {
			t.Errorf("expect empty message, but got %q", rspBodyObj.Data.Message)
		}
	})

	t.Run("异常测试_数据库错误", func(t *testing.T) {
		ctx := context.Background()

		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		dbNoticeMocker = dbNoticeMocker.TakeOnce(nil, gorm.ErrInvalidData)     // 查询数据库中活跃的公告（数据库错误）。
		defer dbNoticeMocker.Reset()
		defer redisMocker.Reset()

		CheckAndUnmarshalBody[protocol.NoticeWebLastRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			consts.ErrSystem,
		)
	})
}

func TestNoticeWebAdd(t *testing.T) {
	const reqPath = "/web/notice/add"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		message := "~"
		activatedTime := time.Now()
		expiredTime := activatedTime.Add(24 * time.Hour)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取登录会话。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 获取数据库用户信息。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 判断用户权限。
		dbNoticeMocker = dbNoticeMocker.CreateOnce(nil)                        // 添加公告到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbNoticeMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.NoticeWebAddReq{
				Message:       message,
				ActivatedTime: protocol.Time(activatedTime),
				ExpiredTime:   protocol.Time(expiredTime),
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, message string, activatedTime, expiredTime time.Time) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取登录会话。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 获取数据库用户信息。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 判断用户权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.NoticeWebAddReq{
				Message:       message,
				ActivatedTime: protocol.Time(activatedTime),
				ExpiredTime:   protocol.Time(expiredTime),
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                       string
		Message                    string
		ActivatedTime, ExpiredTime time.Time
	}{
		{"通知内容缺失", "", time.Now(), time.Now().Add(24 * time.Hour)},
		{"通知内容过多", util.FastRandomAlphaNumberString(16777215 + 1), time.Now(), time.Now().Add(24 * time.Hour)},
		{"没有生效时间", "~", time.Time{}, time.Now().Add(24 * time.Hour)},
		{"没有失效时间", "~", time.Now(), time.Time{}},
		{"没有失效时间小于生效时间", "~", time.Now(), time.Now().Add(-time.Second)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Message, v.ActivatedTime, v.ExpiredTime)
		})
	}
}

func TestNoticeWebList(t *testing.T) {
	const reqPath = "/web/notice/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockNotices := []*model.Notice{
			{
				ID:            1,
				Content:       "~",
				UserID:        1,
				ExpiredTime:   time.Now(),
				ActivatedTime: time.Now(),
				CreatedTime:   time.Now(),
			},
		}
		mockUsers := []*model.User{
			{
				ID:     1,
				NameEn: "~",
			},
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取登录会话。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 获取数据库用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 判断用户权限。
		dbNoticeMocker = dbNoticeMocker.FindOnce(mockNotices, nil)             // 获取公告。
		dbUserMocker = dbUserMocker.FindOnce(mockUsers, nil)                   // 获取用户名。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbNoticeMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.NoticeWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) != len(mockNotices) {
			t.Errorf("expect notices len %v, but got %v", len(mockNotices), len(rspBodyObj.Data.List))
		}
	})
}

func TestNoticeWebRemove(t *testing.T) {
	const reqPath = "/web/notice/remove"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		id := 1

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)               // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)           // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                  // 获取登录会话。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                             // 获取数据库用户信息。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                 // 执行防抖过滤 Redis Lua 脚本。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                            // 判断用户权限。
		dbNoticeMocker = dbNoticeMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil) // 删除公告。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbNoticeMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.NoticeWebRemoveReq{ID: id})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbNoticeMocker := MockDBClient[model.Notice](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取登录会话。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 获取数据库用户信息。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 判断用户权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbNoticeMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.NoticeWebRemoveReq{ID: id})),
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
