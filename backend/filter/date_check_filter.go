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
	"time"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// DateCheckFilter HTTP 请求头时间校验。
func DateCheckFilter(c *gin.Context) {
	ctx := c.Request.Context()
	log.Info(ctx, "check http date")

	// 获取并校验请求时间。
	dateString := c.Request.Header.Get("Date")
	date, err := time.ParseInLocation("Mon, 02 Jan 2006 15:04:05 GMT", dateString, time.Local)
	if err != nil {
		c.Abort()
		log.Warn(ctx, "failed to parse Date", err, dateString)
		util.ResponseCode(c, consts.ErrRequestDateInvalid)
		return
	}

	// 请求时间超时。
	if period := time.Since(date); period > cfg.Get().Backend().MaximumSendInterval() || period < 0 {
		c.Abort()
		util.ResponseCode(c, consts.ErrRequestDateInvalid)
		return
	}

	c.Next()
}
