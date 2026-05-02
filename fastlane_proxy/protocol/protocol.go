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

package protocol

import (
	"context"
	"fmt"

	vt "github.com/go-playground/validator/v10"

	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/validator"
)

// Init 初始化。
func Init(ctx context.Context) {
	initBundle()
	initProfile()
	initCertificate()
	validator.RegisterValidation("capability", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := cc.AppleBundleIDCapabilities[val]
		return ok
	})
	validator.Init(ctx)
}
