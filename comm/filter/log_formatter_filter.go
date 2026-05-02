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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

// LogFormatterFilter 日志过滤器。
func LogFormatterFilter(c *gin.Context) {
	// 组装上下信息。
	requestID := c.Request.Header.Get(consts.HTTPHeaderRequestID)
	if len(requestID) <= 0 {
		requestID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	ctx := ctxs.WithRequestID(c.Request.Context(), requestID)
	ctx = ctxs.WithRequestURI(ctx, c.FullPath())
	requestIP := c.Request.Header.Get(consts.HTTPHeaderIP)
	if len(requestIP) <= 0 {
		requestIP = c.ClientIP()
	}
	ctx = ctxs.WithRequestIP(ctx, requestIP)
	c.Request = c.Request.WithContext(ctx)

	// 打印请求信息。
	contentType := c.Request.Header.Get("Content-Type")
	contentLength := c.Request.Header.Get("Content-Length")
	log.Info(ctx, "START",
		c.Request.Method, contentType, contentLength, c.Request.URL.RawQuery, c.Request.URL.RawFragment)

	// 打印响应信息。
	now := time.Now()
	defer func() {
		cost := time.Since(now)
		ctx = c.Request.Context()
		err := ctxs.Error(ctx)
		contentLength = c.Writer.Header().Get("Content-Length")
		contentType = c.Writer.Header().Get("Content-Type")
		if err != nil {
			log.Warn(ctx, "END",
				cost, c.Writer.Status(), contentType, contentLength, err)
		} else {
			log.Info(ctx, "END",
				cost, c.Writer.Status(), contentType, contentLength)
		}
	}()

	c.Next()
}
