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
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// UserWebRegister 注册
//
//	@Summary	注册
//	@Tags		User-WebAPI
//	@Accept		multipart/form-data
//	@Produce	application/json
//	@Param		Date					header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		file					formData	file	true	"用户头像"
//	@Param		nameZh					formData	string	true	"中文名"		example(张三)
//	@Param		NameEn					formData	string	true	"英文名"		example(zhangsan)
//	@Param		password				formData	string	true	"密码"		example(123456)
//	@Param		passwordConfirmation	formData	string	true	"二次确认密码"	example(123456)
//	@Param		department				formData	string	true	"部门"		example(/技术部/研发)
//	@Response	200						{object}	util.Response[any]
//	@Router		/web/user/register [post]
func UserWebRegister(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.UserWebRegisterReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for registering user",
		req.NameEn, req.NameZh, req.Department, req.Avatar.Filename, req.Avatar.Size)
	if err = service.UserWebRegister(ctx, &req); err != nil {
		log.Warn(ctx, "failed to register user", err,
			req.NameEn, req.NameZh, req.Department, req.Avatar.Filename, req.Avatar.Size)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertRegisterUser)
}

// UserWebLogin 登录。
//
//	@Summary	登录
//	@Tags		User-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		_		body		protocol.UserWebLoginReq	true	"请求体"
//	@Header		200		{string}	Set-Cookie					"会话"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/user/login [post]
func UserWebLogin(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.UserWebLoginReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for user login", req.NameEn)
	session, err := service.UserWebLogin(ctx, &req)
	if err != nil {
		log.Warn(ctx, "user login failed", err, req.NameEn)
		util.ResponseError(c, err)
		return
	}
	c.SetCookie(consts.HTTPHeaderSessionKey,
		session, int(cfg.Get().Backend().WebSessionExpiration().Seconds()), "/", "", false, true)
	c.SetCookie(consts.HTTPHeaderSessionUser,
		req.NameEn, int(cfg.Get().Backend().WebSessionExpiration().Seconds()), "/", "", false, true)
	util.ResponseCode(c, consts.AlertLogin)
}

// UserWebGetInformation 获取用户信息。
//
//	@Summary	获取用户信息
//	@Tags		User-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.UserWebGetInformationRsp]
//	@Router		/web/user/getInformation [get]
func UserWebGetInformation(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.UserWebGetInformation(ctx)
	if err != nil {
		log.Warn(ctx, "failed to get user information", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting user information", rsp)
	util.ResponseData(c, rsp)
}

// UserWebUpdate 更新个人信息。
//
//	@Summary	更新个人信息
//	@Tags		User-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.UserWebUpdateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/user/update [post]
func UserWebUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.UserWebUpdateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating user information", req.Department, req.NameZh, req.AvatarFileID)
	if err = service.UserWebUpdate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to update user information", err, req.Department, req.NameZh, req.AvatarFileID)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// UserWebSearch 搜索用户。
//
//	@Summary	搜索用户
//	@Tags		User-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.UserWebSearchReq	true	"请求体"
//	@Response	200		{object}	util.Response[protocol.UserWebSearchRsp]
//	@Router		/web/user/search [get]
func UserWebSearch(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.UserWebSearchReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for searching users", &req)
	rsp, err := service.UserWebSearch(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to search users", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for searching users", rsp)
	util.ResponseData(c, rsp)
}

// UserWebLogout 登出。
//
//	@Summary	登出
//	@Tags		User-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[any]
//	@Header		200		{string}	Set-Cookie	"会话"
//	@Router		/web/user/logout [delete]
func UserWebLogout(c *gin.Context) {
	ctx := c.Request.Context()
	err := service.UserWebLogout(ctx)
	if err != nil {
		log.Warn(ctx, "failed to logout", err)
		util.ResponseError(c, err)
		return
	}
	c.SetCookie(consts.HTTPHeaderSessionKey, "", 1, "/", "", false, true)
	c.SetCookie(consts.HTTPHeaderSessionUser, "", 1, "/", "", false, true)
	util.ResponseCode(c, consts.AlertLogout)
}
