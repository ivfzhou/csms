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

package impl

import (
	"context"

	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type redisLoggerImpl struct{}

// 新建实例。
func newRedisLogger() log.RedisLogger {
	return &redisLoggerImpl{}
}

func (l *redisLoggerImpl) Printf(ctx context.Context, format string, args ...any) {
	internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLibrary("redis"))
}
