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
	"net/http"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
	"gitee.com/ivfzhou/csms/fastlane_proxy/service"
)

// ProfileApplyInHouse 申请企业内测描述文件。
func ProfileApplyInHouse(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.ApplyInHouseProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying in house profile", &req)
	rsp, err := service.ProfileApplyInHouse(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to apply in house profile", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for applying in house profile", rsp)
	util.ResponseData(c, rsp)
}

// ProfileRemoveInHouse 删除企业内测描述文件。
func ProfileRemoveInHouse(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.RemoveInHouseProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing in house profile", &req)
	if err = service.ProfileRemoveInHouse(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove in house profile", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusOK)
}
