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

// BundleApplyInHouse 申请企业内测 Bundle ID。
func BundleApplyInHouse(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.ApplyInHouseBundleIDReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying apple in house bundle id", &req)
	rsp, err := service.BundleApplyInHouse(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to apply apple in house bundle id", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for applying apple in house bundle id", rsp)
	util.ResponseData(c, rsp)
}

// BundleRemoveInHouse 删除企业内测 Bundle ID。
func BundleRemoveInHouse(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.RemoveInHouseBundleIDReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing apple in house bundle id", &req)
	if err = service.BundleRemoveInHouse(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove apple in house bundle id", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusOK)
}

// BundleModifyInHouseCapabilities 修改企业内测 Bundle ID 能力项。
func BundleModifyInHouseCapabilities(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.ModifyInHouseBundleIDCapabilitiesReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for modifying capabilities of apple in house bundle id", &req)
	if err = service.BundleModifyInHouseCapabilities(ctx, &req); err != nil {
		log.Warn(ctx, "failed to modify capabilities of apple in house bundle id", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusOK)
}
