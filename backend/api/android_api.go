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

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AndroidWebAddOrganization 添加证书主体。
//
//	@Summary	添加证书主体
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.AndroidWebAddOrganizationReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/addOrganization [post]
func AndroidWebAddOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebAddOrganizationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for adding android organization", &req)
	if err = service.AndroidWebAddOrganization(ctx, &req); err != nil {
		log.Warn(ctx, "failed to add android organization", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebListOrganizations 获取安卓主体信息列表。
//
//	@Summary	获取安卓主体信息列表
//	@Tags		Android-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.AndroidWebListOrganizationsRsp]
//	@Router		/web/android/listOrganizations [get]
func AndroidWebListOrganizations(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AndroidWebListOrganizations(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list android organizations", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing android organizations", &rsp)
	util.ResponseData(c, rsp)
}

// AndroidWebApplyCertificate 申请安卓证书。
//
//	@Summary	申请安卓证书
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebApplyCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/applyCertificate/{appId} [post]
func AndroidWebApplyCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebApplyCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying android certificate", &req)
	if err = service.AndroidWebApplyCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply android certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebUploadCertificate 上传安卓证书。
//
//	@Summary	上传安卓证书
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebUploadCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/uploadCertificate/{appId} [post]
func AndroidWebUploadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebUploadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for uploading android certificate", req.Type)
	if err = service.AndroidWebUploadCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to upload android certificate", err, req.Type)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebListCertificates 获取安卓证书列表。
//
//	@Summary	获取安卓证书列表
//	@Tags		Android-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"
//	@Response	200		{object}	util.Response[protocol.AndroidWebListCertificatesRsp]
//	@Router		/web/android/listCertificates/{appId} [get]
func AndroidWebListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AndroidWebListCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list android certificates", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing android certificates", rsp)
	util.ResponseData(c, rsp)
}

// AndroidWebDownloadCertificate 下载安卓证书。
//
//	@Summary	下载安卓证书
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/octet-stream
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		query		protocol.AndroidWebDownloadCertificateReq	true	"请求参数"
//	@Response	200		{file}		Response
//	@Header		200		{string}	Content-Disposition
//	@Router		/web/android/downloadCertificate/{appId} [get]
func AndroidWebDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading android certificate", &req)
	fileObj, err := service.AndroidWebDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download android certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading android certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AndroidWebGetGooglePlayCertificate 获取谷歌 Play 上传证书。
//
//	@Summary	获取谷歌 Play 上传证书
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/octet-stream
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string											true	"应用 ID"
//	@Param		_		query		protocol.AndroidWebGetGooglePlayCertificateReq	true	"请求参数"
//	@Response	200		{file}		Response
//	@Header		200		{string}	Content-Disposition
//	@Router		/web/android/getGooglePlayCertificate/{appId} [get]
func AndroidWebGetGooglePlayCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebGetGooglePlayCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting google play certificate", &req)
	fileObj, err := service.AndroidWebGetGooglePlayCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get google play certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting google play certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AndroidWebGetGooglePlayDeployCertificate 获取谷歌 Play 部署证书。
//
//	@Summary	获取谷歌 Play 部署证书
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/octet-stream
//	@Param		Date	header		string													true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string													true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string													true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebGetGooglePlayDeployCertificateReq	true	"请求体"
//	@Header		200		{string}	Content-Disposition
//	@Response	200		{file}		Response
//	@Router		/web/android/getGooglePlayDeployCertificate/{appId} [post]
func AndroidWebGetGooglePlayDeployCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebGetGooglePlayDeployCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting google play deploy certificate", &req)
	fileObj, err := service.AndroidWebGetGooglePlayDeployCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get google play deploy certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting google play deploy certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AndroidWebGetGooglePlayUpgradeCertificate 获取谷歌 Play 升级签名密钥。
//
//	@Summary	获取谷歌 Play 升级签名密钥
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/octet-stream
//	@Param		Date	header		string													true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string													true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string													true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebGetGooglePlayUpgradeCertificateReq	true	"请求体"
//	@Header		200		{string}	Content-Disposition
//	@Response	200		{file}		Response
//	@Router		/web/android/getGooglePlayUpgradeCertificate/{appId} [post]
func AndroidWebGetGooglePlayUpgradeCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebGetGooglePlayUpgradeCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting google play upgrade certificate", &req)
	fileObj, err := service.AndroidWebGetGooglePlayUpgradeCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get google play upgrade certificate", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting google play upgrade certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AndroidWebGetCertificateFacebookDigest 获取证书的脸书摘要。
//
//	@Summary	获取证书的脸书摘要
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string												true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string												true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string												true	"应用 ID"
//	@Param		_		query		protocol.AndroidWebGetCertificateFacebookDigestReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AndroidWebGetCertificateFacebookDigestRsp]
//	@Router		/web/android/getCertificateFacebookDigest/{appId} [get]
func AndroidWebGetCertificateFacebookDigest(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebGetCertificateFacebookDigestReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting certificate facebook digest", &req)
	rsp, err := service.AndroidWebGetCertificateFacebookDigest(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get certificate facebook digest", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for getting certificate facebook digest", rsp)
	util.ResponseData(c, rsp)
}

// AndroidWebSubmitAPKSigningJob 提交 APK 文件签名任务。
//
//	@Summary	提交 APK 文件签名任务
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebSubmitAPKSigningJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/submitAPKSigningJob/{appId} [post]
func AndroidWebSubmitAPKSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebSubmitAPKSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apk signing job", &req)
	err = service.AndroidWebSubmitAPKSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit apk signing job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebSubmitAABSigningJob 提交 AAB 文件签名任务。
//
//	@Summary	提交 AAB 文件签名任务
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebSubmitAABSigningJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/submitAABSigningJob/{appId} [post]
func AndroidWebSubmitAABSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebSubmitAABSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting aab signing job", &req)
	err = service.AndroidWebSubmitAABSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit aab signing job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebSubmitAPKPatchSigningJob 提交 APK 补丁包文件签名任务。
//
//	@Summary	提交 APK 补丁包文件签名任务
//	@Tags		Android-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string											true	"应用 ID"
//	@Param		_		body		protocol.AndroidWebSubmitAPKPatchSigningJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/submitAPKPatchSigningJob/{appId} [post]
func AndroidWebSubmitAPKPatchSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebSubmitAPKPatchSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apk patch signing job", &req)
	err = service.AndroidWebSubmitAPKPatchSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit apk patch signing job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebListSigningJobs 获取签名任务列表信息。
//
//	@Summary	获取签名任务列表信息
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		query		protocol.AndroidWebListSigningJobsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AndroidWebListSigningJobsRsp]
//	@Router		/web/android/listSigningJobs/{appId} [get]
func AndroidWebListSigningJobs(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebListSigningJobsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing android signing jobs", &req)
	rsp, err := service.AndroidWebListSigningJobs(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to listing android signing jobs", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AndroidWebRemoveOrganization 删除证书主体。
//
//	@Summary	删除证书主体
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.AndroidWebRemoveOrganizationReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/removeOrganization [delete]
func AndroidWebRemoveOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebRemoveOrganizationReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing android organization", &req)
	if err = service.AndroidWebRemoveOrganization(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove android organization", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AndroidWebDeleteCertificate 删除安卓证书。
//
//	@Summary	删除安卓证书
//	@Tags		Android-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		query		protocol.AndroidWebDeleteCertificateReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/android/deleteCertificate/{appId} [delete]
func AndroidWebDeleteCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AndroidWebDeleteCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for deleting certificate", &req)
	if err = service.AndroidWebDeleteCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to delete android certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

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
