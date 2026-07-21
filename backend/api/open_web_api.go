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

// OpenWebApply 申请。
//
//	@Summary	申请
//	@Tags		Open-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		body		protocol.OpenWebApplyReq	true	"请求体"
//	@Response	200		{object}	util.Response[protocol.OpenWebApplyRsp]
//	@Router		/web/open/apply/{appId} [post]
func OpenWebApply(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebApplyReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying api account", &req)
	rsp, err := service.OpenWebApply(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to apply api account", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// OpenWebUpdate 修改。
//
//	@Summary	修改
//	@Tags		Open-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		body		protocol.OpenWebUpdateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/open/update/{appId} [post]
func OpenWebUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebUpdateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating api account", &req)
	if err = service.OpenWebUpdate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to update api account", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// OpenWebGetInformation 获取请求凭证信息。
//
//	@Summary	获取请求凭证信息
//	@Tags		Open-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		query		protocol.OpenWebGetInformationReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.OpenWebGetInformationRsp]
//	@Router		/web/open/getInformation/{appId} [get]
func OpenWebGetInformation(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebGetInformationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting api account information", &req)
	rsp, err := service.OpenWebGetInformation(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get api account information", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting api account information", rsp)
	util.ResponseData(c, rsp)
}

// OpenWebList 获取请求凭证信息列表。
//
//	@Summary	获取请求凭证信息列表
//	@Tags		Open-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string					true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string					true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string					true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		query		protocol.OpenWebListReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.OpenWebListRsp]
//	@Router		/web/open/list/{appId} [get]
func OpenWebList(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebListReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing api accounts", &req)
	rsp, err := service.OpenWebList(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list api accounts", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing api accounts", rsp)
	util.ResponseData(c, rsp)
}

// OpenWebRenewal 续期。
//
//	@Summary	续期
//	@Tags		Open-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		query		protocol.OpenWebRenewalReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/open/renewal/{appId} [get]
func OpenWebRenewal(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebRenewalReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for renewal api account", &req)
	if err = service.OpenWebRenewal(ctx, &req); err != nil {
		log.Warn(ctx, "failed to renewal api account", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// OpenWebReset 重置密钥。
//
//	@Summary	重置密钥
//	@Tags		Open-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		query		protocol.OpenWebResetReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.OpenWebResetRsp]
//	@Router		/web/open/reset/{appId} [get]
func OpenWebReset(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebResetReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for resetting api account", &req)
	rsp, err := service.OpenWebReset(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to reset api account", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for resetting api account", &rsp)
	util.ResponseData(c, rsp)
}

// OpenWebRemove 删除。
//
//	@Summary	删除
//	@Tags		Open-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		query		protocol.OpenWebRemoveReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/open/remove/{appId} [delete]
func OpenWebRemove(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.OpenWebRemoveReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing api account", &req)
	if err = service.OpenWebRemove(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove api account", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}
