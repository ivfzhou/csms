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

	"github.com/ivfzhou/cron/v3"

	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type cronLoggerImpl struct{}

// 新建实例。
func newCronLogger() cron.Logger {
	return &cronLoggerImpl{}
}

func (l *cronLoggerImpl) Info(msg string, keysAndValues ...any) {
	if internal.GetLevel() <= log.LevelInfo {
		arr := make([]any, 0, len(keysAndValues)+1)
		arr = append(arr, msg)
		arr = append(arr, keysAndValues...)
		internal.SendBuilderToWriter(
			internal.CreateBuilder(context.TODO(), arr...).SetLevel(log.LevelInfo).SetLibrary("cron"))
	}
}

func (l *cronLoggerImpl) Error(err error, msg string, keysAndValues ...any) {
	if internal.GetLevel() <= log.LevelError {
		arr := make([]any, 0, len(keysAndValues)+2)
		arr = append(arr, err)
		arr = append(arr, msg)
		arr = append(arr, keysAndValues...)
		internal.SendBuilderToWriter(
			internal.CreateBuilder(context.TODO(), arr...).SetLevel(log.LevelError).SetLibrary("cron"))
	}
}
