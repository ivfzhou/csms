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

// AppleInternalGetSigningJob 获取任务信息。
func AppleInternalGetSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleInternalGetSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting apple signing job information", &req)
	rsp, err := service.AppleInternalGetSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get apple signing job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AppleInternalGetCertificateAndProfile 获取证书和描述文件信息。
func AppleInternalGetCertificateAndProfile(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleInternalGetCertificateAndProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting apple certificate and profile information", &req)
	rsp, err := service.AppleInternalGetCertificateAndProfile(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get apple certificate and profile information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AppleInternalUpdateSigningJob 更新任务信息。
func AppleInternalUpdateSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleInternalUpdateSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating apple signing job information", &req)
	err = service.AppleInternalUpdateSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to update apple signing job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusNoContent)
}
