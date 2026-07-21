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

	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AndroidInternalGetSigningJob 获取签名任务信息。
func AndroidInternalGetSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidInternalGetSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting android job information", &req)
	rsp, err := service.AndroidInternalGetSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get android job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidInternalGetCertificate 获取安卓证书信息。
func AndroidInternalGetCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidInternalGetCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting android certificate information", &req)
	rsp, err := service.AndroidInternalGetCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get android certificate information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidInternalUpdateSigningJob 更新任务信息。
func AndroidInternalUpdateSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidInternalUpdateSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating android job information", &req)
	err = service.AndroidInternalUpdateSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to update android job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusNoContent)
}
