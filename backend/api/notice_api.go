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

package api

import (
	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// NoticeWebLast 获取通知。
//
//	@Summary	获取通知
//	@Tags		Notice-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Response	200		{object}	util.Response[protocol.NoticeWebLastRsp]
//	@Router		/web/notice/last [get]
func NoticeWebLast(c *gin.Context) {
	ctx := c.Request.Context()
	log.Info(ctx, "request parameters for get last notice")
	rsp, err := service.NoticeWebLast(ctx)
	if err != nil {
		log.Warn(ctx, "failed to get last notice", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}
