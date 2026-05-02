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

// CertificateApplyPush 申请 Push 证书。
func CertificateApplyPush(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.ApplyPushCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying apple push certificate", req.BundleID, req.Type, req.Environment)
	rsp, err := service.CertificateApplyPush(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to apply apple push certificate", err, req.BundleID, req.Type, req.Environment)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for applying apple push certificate", rsp.ID)
	util.ResponseData(c, rsp)
}

// CertificateRemovePush 删除 Push 证书。
func CertificateRemovePush(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.RemovePushCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing apple push certificate", &req)
	if err = service.CertificateRemovePush(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove apple push certificate", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusOK)
}
