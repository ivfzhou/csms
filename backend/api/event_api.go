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

	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// EventWebList 获取应用事件列表。
//
//	@Summary	获取应用事件列表
//	@Tags		Event-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.EventWebListReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.EventWebListRsp]
//	@Router		/web/event/list [get]
func EventWebList(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.EventWebListReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing app events", &req)
	rsp, err := service.EventWebList(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list app events", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing app events", rsp)
	util.ResponseData(c, rsp)
}

// EventWebStatistic 获取应用事件统计数量。
//
//	@Summary	获取应用事件统计数量
//	@Tags		Event-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string							true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string							true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.EventWebStatisticReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.EventWebStatisticRsp]
//	@Router		/web/event/statistic [get]
func EventWebStatistic(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.EventWebStatisticReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for statistic app events", &req)
	rsp, err := service.EventWebStatistic(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to statistic app events", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for statistic app events", rsp)
	util.ResponseData(c, rsp)
}
