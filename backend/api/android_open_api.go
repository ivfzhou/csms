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

// AndroidAPIDownloadCertificate 下载安卓证书。
//
//	@Summary	下载安卓证书
//	@Tags		Android-OpenAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/octet-stream
//	@Param		Authorization	header		string										true	"请求凭据"
//	@Param		_				query		protocol.AndroidAPIDownloadCertificateReq	true	"请求参数"
//	@Response	200				{file}		Response
//	@Header		200				{string}	Content-Disposition
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/downloadCertificate [get]
func AndroidAPIDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidAPIDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading android certificate", &req)
	fileObj, err := service.AndroidAPIDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download android certificate", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading android certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AndroidAPISubmitAPKSigningJob 提交 APK 文件签名任务。
//
//	@Summary	提交 APK 文件签名任务
//	@Tags		Android-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string										true	"请求凭据"
//	@Param		_				body		protocol.AndroidAPISubmitAPKSigningJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.AndroidAPISubmitAPKSigningJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/submitAPKSigningJob [post]
func AndroidAPISubmitAPKSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidAPISubmitAPKSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apk signing job", &req)
	rsp, err := service.AndroidAPISubmitAPKSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit apk signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidAPISubmitAABSigningJob 提交 AAB 文件签名任务。
//
//	@Summary	提交 AAB 文件签名任务
//	@Tags		Android-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string										true	"请求凭据"
//	@Param		_				body		protocol.AndroidAPISubmitAABSigningJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.AndroidAPISubmitAABSigningJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/submitAABSigningJob [post]
func AndroidAPISubmitAABSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidAPISubmitAABSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting aab signing job", &req)
	rsp, err := service.AndroidAPISubmitAABSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit aab signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidAPISubmitAPKPatchSigningJob 提交 APK 补丁包文件签名任务。
//
//	@Summary	提交 APK 补丁包文件签名任务
//	@Tags		Android-OpenAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Authorization	header		string											true	"请求凭据"
//	@Param		_				body		protocol.AndroidAPISubmitAPKPatchSigningJobReq	true	"请求体"
//	@Response	200				{object}	util.Response[protocol.AndroidAPISubmitAPKPatchSigningJobRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/submitAPKPatchSigningJob [post]
func AndroidAPISubmitAPKPatchSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidAPISubmitAPKPatchSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apk patch signing job", &req)
	rsp, err := service.AndroidAPISubmitAPKPatchSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit apk patch signing job", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidAPIGetSigningJobInformation 获取任务信息。
//
//	@Summary	获取任务信息
//	@Tags		Android-OpenAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Authorization	header		string											true	"请求凭据"
//	@Param		_				query		protocol.AndroidAPIGetSigningJobInformationReq	true	"请求参数"
//	@Response	200				{object}	util.Response[protocol.AndroidAPIGetSigningJobInformationRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	404				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/getSigningJobInformation [get]
func AndroidAPIGetSigningJobInformation(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidAPIGetSigningJobInformationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting signing job information", &req)
	rsp, err := service.AndroidAPIGetJobInformation(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get signing job information", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidAPIListCertificates 获取安卓证书列表。
//
//	@Summary	获取安卓证书列表
//	@Tags		Android-OpenAPI
//	@Produce	application/json
//	@Param		Authorization	header		string	true	"请求凭据"
//	@Response	200				{object}	util.Response[protocol.AndroidAPIListCertificatesRsp]
//	@Response	400				{object}	util.Response[any]
//	@Response	401				{object}	util.Response[any]
//	@Response	403				{object}	util.Response[any]
//	@Response	429				{object}	util.Response[any]
//	@Response	500				{object}	util.Response[any]
//	@Router		/api/android/listCertificates [get]
func AndroidAPIListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AndroidAPIListCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list android certificates", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for listing android certificates", rsp)
	util.ResponseData(c, rsp)
}
