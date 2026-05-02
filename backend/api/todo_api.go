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

// TodoWebCount 获取用户待办、已办数量。
//
//	@Summary	获取用户待办、已办数量
//	@Tags		Todo-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.TodoWebCountRsp]
//	@Router		/web/todo/count [get]
func TodoWebCount(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.TodoWebCount(ctx)
	if err != nil {
		log.Warn(ctx, "failed to get the number of user todo count", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for count todo", rsp)
	util.ResponseData(c, rsp)
}

// TodoWebList 获取需要处理的待办。
//
//	@Summary	获取需要处理的待办
//	@Tags		Todo-WebAPI
//	@Produce	application/x-www-form-urlencoded
//	@Accept		application/json
//	@Param		Date	header		string					true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string					true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.TodoWebListReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.TodoWebListRsp]
//	@Router		/web/todo/list [get]
func TodoWebList(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.TodoWebListReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing todos", &req)
	rsp, err := service.TodoWebList(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list todos", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing todos", rsp)
	util.ResponseData(c, rsp)
}

// TodoWebListDealt 获取已处理的待办列表。
//
//	@Summary	获取已处理的待办列表
//	@Tags		Todo-WebAPI
//	@Produce	application/x-www-form-urlencoded
//	@Accept		application/json
//	@Param		Date	header		string							true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string							true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.TodoWebListDealtReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.TodoWebListDealtRsp]
//	@Router		/web/todo/listDealt [get]
func TodoWebListDealt(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.TodoWebListDealtReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing todos", &req)
	rsp, err := service.TodoWebListDealt(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list todos", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing todos", rsp)
	util.ResponseData(c, rsp)
}

// TodoWebCreate 创建。
//
//	@Summary	创建
//	@Tags		Todo-WebAPI
//	@Produce	application/json
//	@Accept		application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.TodoWebCreateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/todo/create [post]
func TodoWebCreate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.TodoWebCreateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for creating todo", &req)
	if err = service.TodoWebCreate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to creating todo", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// TodoWebGetDetail 获取待办详情。
//
//	@Summary	获取待办详情
//	@Tags		Todo-WebAPI
//	@Produce	application/x-www-form-urlencoded
//	@Accept		application/json
//	@Param		Date	header		string							true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string							true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.TodoWebGetDetailReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.TodoWebGetDetailRsp]
//	@Router		/web/todo/getDetail [get]
func TodoWebGetDetail(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.TodoWebGetDetailReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting todo detail", &req)
	rsp, err := service.TodoWebGetDetail(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get todo detail", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting todo detail", rsp)
	util.ResponseData(c, rsp)
}

// TodoWebDeal 审批。
//
//	@Summary	审批
//	@Tags		Todo-WebAPI
//	@Produce	application/json
//	@Accept		application/json
//	@Param		Date	header		string					true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string					true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.TodoWebDealReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/todo/deal [post]
func TodoWebDeal(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.TodoWebDealReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for dealing todo", &req)
	if err = service.TodoWebDeal(ctx, &req); err != nil {
		log.Warn(ctx, "failed to dealing todo", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}
