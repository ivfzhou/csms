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

// AppleAPIDownloadCertificate 下载证书和描述文件。
//
//	@Summary	下载证书和描述文件
//	@Tags		Apple-OpenAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/otect-stream
//	@Param		Authorization	header		string									true	"请求凭据"
//	@Param		_				query		protocol.AppleAPIDownloadCertificateReq	true	"请求参数"
//	@Response	200				{file}		Response
//	@Header		200				{string}	Content-Disposition
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/apple/downloadCertificate [get]
func AppleAPIDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleAPIDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading apple certificate", &req)
	fileObj, err := service.AppleAPIDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download apple certificate", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading apple certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AppleAPISubmitSigningJob 提交签名任务。
//
//	@Summary	提交签名任务
//	@Tags		Apple-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string									true	"请求凭据"
//	@Param		_				body		protocol.AppleAPISubmitSigningJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.AppleAPISubmitSigningJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/apple/submitSigningJob [post]
func AppleAPISubmitSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleAPISubmitSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apple signing job", &req)
	rsp, err := service.AppleAPISubmitSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit apple signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AppleAPIGetSigningJobInformation 获取签名任务信息。
//
//	@Summary	获取签名任务信息
//	@Tags		Apple-OpenAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Authorization	header		string											true	"请求凭据"
//	@Param		_				query		protocol.AppleAPIGetSigningJobInformationReq	true	"请求参数"
//	@Response	200				{object}	util.Response[protocol.AppleAPIGetSigningJobInformationRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/apple/getSigningJobInformation [get]
func AppleAPIGetSigningJobInformation(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleAPIGetSigningJobInformationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting apple signing job information", &req)
	rsp, err := service.AppleAPIGetSigningJobInformation(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get apple signing job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}
