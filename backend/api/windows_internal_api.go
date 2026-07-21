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

// WindowsInternalGetWHQLJob 获取 WHQL 任务信息。
func WindowsInternalGetWHQLJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetWHQLJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting whql job information", &req)
	rsp, err := service.WindowsInternalGetWHQLJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsInternalGetWHQLJobToInitialTestMachine 给 HLK 测试虚拟机初始化任务。
func WindowsInternalGetWHQLJobToInitialTestMachine(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetWHQLJobToInitialTestMachineReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting whql job to initial test machine", &req)
	rsp, err := service.WindowsInternalGetWHQLJobToInitialTestMachine(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to initial test machine", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsInternalUpdateWHQLJob 更新任务。
func WindowsInternalUpdateWHQLJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalUpdateWHQLJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating whql job", &req)
	err = service.WindowsInternalUpdateWHQLJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to update whql job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusNoContent)
}

// WindowsInternalGetWHQLJobToStartTest 获取任务，调度测试。
func WindowsInternalGetWHQLJobToStartTest(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetWHQLJobToStartTestReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting whql job to start test", &req)
	rsp, err := service.WindowsInternalGetWHQLJobToStartTest(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to start test", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsInternalGetTestingWHQLJobs 获取正在测试中的任务。
func WindowsInternalGetTestingWHQLJobs(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetTestingWHQLJobsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting testing whql jobs", &req)
	rsp, err := service.WindowsInternalGetTestingWHQLJobs(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to get testing whql jobs", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, &rsp)
}

// WindowsInternalGetMachineEVCertificates 获取签名机器上在用的 EV UKey。
func WindowsInternalGetMachineEVCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetMachineEVCertificatesReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting machine ev certificates", &req)
	rsp, err := service.WindowsInternalGetMachineEVCertificates(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to get machine ev certificates", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, &rsp)
}

// WindowsInternalGetWindowsSigningJob 获取签名任务信息。
func WindowsInternalGetWindowsSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetWindowsSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting windows signing job", &req)
	rsp, err := service.WindowsInternalGetWindowsSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to get windows signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsInternalGetCertificate 获取证书信息。
func WindowsInternalGetCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalGetCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting certificate", &req)
	rsp, err := service.WindowsInternalGetCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to get certificate", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsInternalUpdateSigningJob 更新签名任务。
func WindowsInternalUpdateSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsInternalUpdateSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for updating windows signing job", &req)
	err = service.WindowsInternalUpdateSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job to update windows signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseCode(c, http.StatusNoContent)
}
