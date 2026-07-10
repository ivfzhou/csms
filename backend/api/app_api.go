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

// AppWebRegister 注册。
//
//	@Summary	注册
//	@Tags		App-WebAPI
//	@Accept		multipart/form-data
//	@Produce	application/json
//	@Param		Date		header		string		true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie		header		string		true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		name		formData	string		true	"应用名"	example(MyApp)
//	@Param		logo		formData	file		true	"应用图标"
//	@Param		platform	formData	integer		true	"平台"	example(1)
//	@Param		admins		formData	[]string	true	"管理员"	example([]{"zhangsan"})
//	@Param		members		formData	[]string	true	"成员"	example([]{"zhangsan"})
//	@Response	200			{object}	util.Response[any]
//	@Router		/web/app/register [post]
func AppWebRegister(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppWebRegisterReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for registering app",
		req.Name, req.Members, req.Admins, req.Platform, req.Logo.Filename, req.Logo.Size)
	if err = service.AppWebRegister(ctx, &req); err != nil {
		log.Warn(ctx, "failed to register app",
			err, req.Name, req.Members, req.Admins, req.Platform, req.Logo.Filename, req.Logo.Size)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertRegisterApp)
}

// AppWebSearch 查询。
//
//	@Summary	查询
//	@Tags		App-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.AppWebSearchReq	true	"请求体"
//	@Response	200		{object}	util.Response[protocol.AppWebSearchRsp]
//	@Router		/web/app/search [post]
func AppWebSearch(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppWebSearchReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for searching apps", &req)
	rsp, err := service.AppWebSearch(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to search apps", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for searching apps", rsp)
	util.ResponseData(c, rsp)
}

// AppWebUpdate 更新。
//
//	@Summary	更新
//	@Tags		App-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string						true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		_		body		protocol.AppWebUpdateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/app/update/{appId} [post]
func AppWebUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppWebUpdateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating app information", &req)
	if err = service.AppWebUpdate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to update app information", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppWebGetInformation 获取应用信息。
//
//	@Summary	获取应用信息
//	@Tags		App-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Response	200		{object}	util.Response[protocol.AppWebGetInformationRsp]
//	@Router		/web/app/getInformation/{appId} [get]
func AppWebGetInformation(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AppWebGetInformation(ctx)
	if err != nil {
		log.Warn(ctx, "failed to get app information", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting app information", rsp)
	util.ResponseData(c, rsp)
}

// AppWebInvalidate 无效化。
//
//	@Summary	无效化
//	@Tags		App-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/app/invalidate/{appId} [get]
func AppWebInvalidate(c *gin.Context) {
	ctx := c.Request.Context()
	err := service.AppWebInvalidate(ctx)
	if err != nil {
		log.Warn(ctx, "failed to invalidate app", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppWebEnable 启用。
//
//	@Summary	启用
//	@Tags		App-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"	example(4ef83c03e2ce4f1f94c11168d1acd087)
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/app/enable/{appId} [get]
func AppWebEnable(c *gin.Context) {
	ctx := c.Request.Context()
	err := service.AppWebEnable(ctx)
	if err != nil {
		log.Warn(ctx, "failed to enable app", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertEnableApp)
}

// AppWebCount 获取用户具有权限的应用个数。
//
//	@Summary	获取用户具有权限的应用个数
//	@Tags		App-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.AppWebCountRsp]
//	@Router		/web/app/count [get]
func AppWebCount(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AppWebCount(ctx)
	if err != nil {
		log.Warn(ctx, "failed to counting number of user apps", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for counting number of apps", &rsp)
	util.ResponseData(c, rsp)
}
