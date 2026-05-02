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
	"net/http"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// ExitFilter 程序正在退出，拒绝服务。
func ExitFilter(ctx context.Context) func(*gin.Context) {
	return func(c *gin.Context) {
		select {
		case <-ctx.Done():
			c.Abort()
			log.Warn(c.Request.Context(), "server is exiting")
			util.ResponseStatusCode(c, http.StatusServiceUnavailable, errs.ErrServerExiting)
		default:
			c.Next()
		}
	}
}
