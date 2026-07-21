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

// WindowsAPIDownloadCertificate 下载证书。
//
//	@Summary	下载证书
//	@Tags		Windows-OpenAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/octet-stream
//	@Param		Authorization	header		string										true	"请求凭据"
//	@Param		_				query		protocol.WindowsAPIDownloadCertificateReq	true	"请求参数"
//	@Response	200				{file}		Response
//	@Header		200				{string}	Content-Disposition
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/downloadCertificate [get]
func WindowsAPIDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPIDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading windows certificate", &req)
	fileObj, err := service.WindowsAPIDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download windows certificate", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading windows certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// WindowsAPIGetCertificatePassword 查看证书密码。
//
//	@Summary	查看证书密码
//	@Tags		Windows-OpenAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Authorization	header		string											true	"请求凭据"
//	@Param		_				query		protocol.WindowsAPIGetCertificatePasswordReq	true	"请求参数"
//	@Response	200				{object}	util.Response[protocol.WindowsAPIGetCertificatePasswordRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/getCertificatePassword [get]
func WindowsAPIGetCertificatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPIGetCertificatePasswordReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting windows certificate password", &req)
	rsp, err := service.WindowsAPIGetCertificatePassword(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get windows certificate password", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsAPISubmitSigningJob 提交签名任务。
//
//	@Summary	提交签名任务
//	@Tags		Windows-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string									true	"请求凭据"
//	@Param		_				body		protocol.WindowsAPISubmitSigningJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.WindowsAPISubmitSigningJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/submitSigningJob [post]
func WindowsAPISubmitSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPISubmitSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting windows signing job", &req)
	rsp, err := service.WindowsAPISubmitSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit windows signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsAPISubmitWHQLJob 提交 WHQL 任务。
//
//	@Summary	提交 WHQL 任务
//	@Tags		Windows-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string								true	"请求凭据"
//	@Param		_				body		protocol.WindowsAPISubmitWHQLJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.WindowsAPISubmitWHQLJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/submitWHQLJob [post]
func WindowsAPISubmitWHQLJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPISubmitWHQLJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting whql job", &req)
	rsp, err := service.WindowsAPISubmitWHQLJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit whql job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsAPIListCertificates 证书列表。
//
//	@Summary	证书列表
//	@Tags		Windows-OpenAPI
//	@Produce	application/json
//	@Param		Authorization	header		string	true	"请求凭据"
//	@Response	200				{object}	util.Response[protocol.WindowsAPIListCertificatesRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/listCertificates [get]
func WindowsAPIListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.WindowsAPIListCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list windows certificates", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for listing windows certificates", rsp)
	util.ResponseData(c, rsp)
}

// WindowsAPIGetSigningJobInformation 获取签名任务信息。
//
//	@Summary	获取签名任务信息
//	@Tags		Windows-OpenAPI
//	@Produce	application/json
//	@Param		Authorization	header		string											true	"请求凭据"
//	@Param		_				query		protocol.WindowsAPIGetSigningJobInformationReq	true	"请求参数"
//	@Response	200				{object}	util.Response[protocol.WindowsAPIGetSigningJobInformationRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/getSigningJobInformation [get]
func WindowsAPIGetSigningJobInformation(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPIGetSigningJobInformationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting signing job information", &req)
	rsp, err := service.WindowsAPIGetSigningJobInformation(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get signing job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsAPIGetWHQLJobInformation 获取 WHQL 任务信息。
//
//	@Summary	获取 WHQL 任务信息
//	@Tags		Windows-OpenAPI
//	@Produce	application/json
//	@Param		Authorization	header		string										true	"请求凭据"
//	@Param		_				query		protocol.WindowsAPIGetWHQLJobInformationReq	true	"请求参数"
//	@Response	200				{object}	util.Response[protocol.WindowsAPIGetWHQLJobInformationRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/windows/getWHQLJobInformation [get]
func WindowsAPIGetWHQLJobInformation(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsAPIGetWHQLJobInformationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting whql job information", &req)
	rsp, err := service.WindowsAPIGetWHQLJobInformation(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get whql job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}
