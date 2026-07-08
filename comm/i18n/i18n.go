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

package i18n

import (
	"context"
	"sync/atomic"

	"gitee.com/ivfzhou/csms/comm/errs"
)

// 支持的提示语语言环境。
const (
	LanguageEnglish Language = "en"
	LanguageChinese Language = "zh"
)

var (
	initializedFlag atomic.Int32
	getFunc         = func(errs.Code, Language) (string, bool) { return "", false }
	closeFunc       = func(context.Context) {}
)

// Language 语言码。
type Language string

// RegisterImplement 设置实现者。
func RegisterImplement(get func(errs.Code, Language) (string, bool), close func(context.Context)) {
	if getFunc == nil || close == nil {
		panic("nil value is not allowed")
	}

	if !initializedFlag.CompareAndSwap(0, 1) {
		panic("function has already been called")
	}

	getFunc = get
	closeFunc = close
}

// Get 获取消息提示语。
func Get(code errs.Code, language Language) (string, bool) {
	return getFunc(code, language)
}

// Close 关闭对消息提示语文件的改动监听。
func Close(ctx context.Context) {
	closeFunc(ctx)
}
