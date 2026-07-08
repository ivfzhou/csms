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

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/comm/validator"
)

var pathToAuthorities map[string][]int

// AddPathAuthorities 添加请求权限。
func AddPathAuthorities(path string, auths ...int) {
	if pathToAuthorities == nil {
		pathToAuthorities = make(map[string][]int)
	}
	pathToAuthorities[path] = util.CleanNumbers(append(pathToAuthorities[path], auths...))
}

// PermissionWebAuthenticateFilter 鉴权函数。
func PermissionWebAuthenticateFilter(c *gin.Context) {
	// 获取需要的权限项。
	ctx := c.Request.Context()
	log.Info(ctx, "authenticate web request")
	reqPath := c.FullPath()
	authorities := pathToAuthorities[reqPath]

	// 获取应用信息。
	var err error
	var app *model.App
	appID := c.Param(consts.HTTPPathAppID)
	if len(appID) > 0 {
		err = validator.Validator.Var(appID, "len=32,alphanum")
		if err != nil {
			c.Abort()
			log.Warn(ctx, "app not found in database", appID)
			util.ResponseCode(c, errs.ErrInvalidRequestParameters)
			return
		}
		app, err = service.GetAppInfoByID(ctx, appID)
		if err != nil {
			c.Abort()
			util.ResponseError(c, err)
			return
		}
		if app == nil || app.ID <= 0 {
			c.Abort()
			log.Warn(ctx, "app not found in database", appID)
			util.ResponseCode(c, errs.ErrInvalidRequestParameters)
			return
		}
	}

	// 应用信息保存到上下文。
	ctx = ctxs.WithApp(ctx, app)
	c.Request = c.Request.WithContext(ctx)

	// 不需要权限。
	if len(authorities) <= 0 {
		c.Next()
		return
	}

	// 检索数据库，判断权限。
	user := ctxs.User(ctx)
	appIDInteger := 0
	if app != nil {
		appIDInteger = app.ID
	}
	hasRight, err := service.UserHasAnyRight(ctx, user.ID, appIDInteger, authorities[0], authorities[1:]...)
	if err != nil {
		c.Abort()
		util.ResponseError(c, err)
		return
	}

	// 无权，限制请求。
	if !hasRight {
		c.Abort()
		util.ResponseCode(c, consts.ErrPermissionDenied)
		return
	}

	c.Next()
}

// PermissionAPIAuthenticateFilter 鉴权函数。
func PermissionAPIAuthenticateFilter(c *gin.Context) {
	// 获取需要的权限项。
	ctx := c.Request.Context()
	log.Info(ctx, "authenticate api request")
	reqPath := c.FullPath()
	authorities := pathToAuthorities[reqPath]

	// 不需要权限。
	if len(authorities) <= 0 {
		c.Next()
		return
	}

	// 检索数据库判断用户是否有权限。
	apiAccount := ctxs.APIAccount(ctx)
	hasRight, err := service.APIAccountHasAnyRight(ctx, apiAccount.ID, authorities[0], authorities[1:]...)
	if err != nil {
		c.Abort()
		util.ResponseAPIError(c, err)
		return
	}

	// 无权，限制请求。
	if !hasRight {
		c.Abort()
		util.ResponseStatus(c, http.StatusForbidden)
		return
	}

	c.Next()
}
