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

package filter

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

const (
	// 请求间隔校验脚本。
	antiShakeRedisScript = `
-- 防抖使用的 Redis Hash 键。
local key = '` + consts.RedisKeyAntiShake + `';
local field = KEYS[1]..KEYS[2];
-- 获取当前时间。
local nowArr = redis.call('time');
local curAccessTime = tonumber(nowArr[1] * 1000);
-- 获取用户该请求的上次请求时间。
local lastAccessTime = tonumber(redis.call('hget', key, field) or 0);
-- 比较与本次请求时间。
local delta = curAccessTime - lastAccessTime;
local limit = tonumber(ARGV[1]);
if (delta >= 0 and delta < limit) or (delta < 0 and delta > - limit) then 
	return 0;
end
-- 通过校验更新时间值。
redis.call('hset', key, field, curAccessTime);
return 1;
`
)

// 请求间隔校验脚本的摘要值。
var antiShakeRedisCmdSha string

// InitAntiShakeScript 初始化防抖脚本。
func InitAntiShakeScript(ctx context.Context) {
	var err error
	antiShakeRedisCmdSha, err = conn.RedisClient(ctx).ScriptLoad(ctx, antiShakeRedisScript).Result()
	if err != nil {
		log.Fatal(ctx, cc.ExitCodeLoadRedisScriptError, "failed to load anti shake redis script", err)
	}
}

// AntiShakeFilter 请求防抖过滤器。
func AntiShakeFilter(c *gin.Context) {
	// 获取上下文信息。
	ctx := c.Request.Context()
	user := ctxs.User(ctx)
	mark := ctxs.RequestIP(ctx)
	if cc.SkipRateLimit() {
		c.Next()
		return
	}
	if util.IsLocalEnvironment() && cc.TestMode() {
		log.Warn(ctx, "test mode skip anti shake filter")
		c.Next()
		return
	}
	log.Info(ctx, "anti shake request")

	// 登录情况下，使用用户名为维度。
	if user != nil {
		mark = user.NameEn
	}

	// 执行请求间隔校验脚本。
	reqPath := ctxs.RequestURI(ctx)
	pass, err := conn.RedisClient(ctx).EvalSha(ctx, antiShakeRedisCmdSha,
		[]string{mark, reqPath}, cfg.Get().Backend().MinimumRequestInterval()/time.Millisecond).Bool()
	if err != nil {
		c.Abort()
		log.Error(ctx, "failed to eval anti shake redis script", err)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}

	// 未通过校验，丢掉请求。
	if !pass {
		c.Abort()
		util.ResponseCode(c, consts.ErrRateLimitReached)
		return
	}

	c.Next()
}
