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

package cfg

import (
	"context"
	"sync/atomic"
)

var (
	initializedFlag int32
	closeFunc       = func(context.Context) {}
	addNotifierFunc = func(func(Configurer)) {}
	getFunc         = func() Configurer { return defaultValue }
)

// RegisterImplement 设置实现者。
func RegisterImplement(get func() Configurer, closer func(context.Context), addNotifier func(func(Configurer))) {
	if closer == nil || addNotifier == nil {
		panic("nil is not allowed")
	}
	if !atomic.CompareAndSwapInt32(&initializedFlag, 0, 1) {
		panic("function has already been called")
	}

	closeFunc = closer
	addNotifierFunc = addNotifier
	getFunc = get
}

// Get 获取配置。
func Get() Configurer {
	return getFunc()
}

// RegisterNotifier 注册配置更新监听者。
func RegisterNotifier(notifier func(Configurer)) {
	addNotifierFunc(notifier)
}

// Close 关闭配置刷新。
func Close(ctx context.Context) {
	closeFunc(ctx)
}
