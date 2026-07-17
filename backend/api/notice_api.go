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

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
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

// NoticeWebAdd 添加通知。
//
//	@Summary	添加通知
//	@Tags		Notice-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.NoticeWebAddReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/notice/add [post]
func NoticeWebAdd(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.NoticeWebAddReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for adding notice", &req)
	err = service.NoticeWebAdd(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to add notice", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// NoticeWebList 通知列表。
//
//	@Summary	通知列表
//	@Tags		Notice-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.NoticeWebListRsp]
//	@Router		/web/notice/list [get]
func NoticeWebList(c *gin.Context) {
	ctx := c.Request.Context()
	log.Info(ctx, "request parameters for get last notice")
	rsp, err := service.NoticeWebList(ctx)
	if err != nil {
		log.Warn(ctx, "failed to get list notices", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// NoticeWebList 删除通知。
//
//	@Summary	删除通知
//	@Tags		Notice-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.NoticeWebRemoveReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/notice/remove [delete]
func NoticeWebRemove(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.NoticeWebRemoveReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing notice", &req)
	err = service.NoticeWebRemove(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to remove notice", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}
