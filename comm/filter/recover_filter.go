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
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// Recover 恐慌恢复过滤器。
func Recover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.Abort()
			log.Error(c.Request.Context(), "request panic", err, util.GetStackCallers())
			if util.IsLocalEnvironment() {
				debug.PrintStack()
			}
			util.ResponseStatusCode(c, http.StatusInternalServerError, errs.ErrServerPanic)
		}
	}()

	c.Next()
}
