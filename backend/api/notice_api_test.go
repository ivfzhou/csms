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

	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/model"
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
