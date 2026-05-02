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

// WindowsWebUploadCertificate 上传个人 OV 证书。
//
//	@Summary	上传个人 OV 证书
//	@Tags		Windows-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.WindowsWebUploadCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/uploadCertificate/{appId} [post]
func WindowsWebUploadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebUploadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	if err = service.WindowsWebUploadCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to upload certificate", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebListCertificates 证书列表。
//
//	@Summary	证书列表
//	@Tags		Windows-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"
//	@Response	200		{object}	util.Response[protocol.WindowsWebListCertificatesRsp]
//	@Router		/web/windows/listCertificates/{appId} [get]
func WindowsWebListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.WindowsWebListCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list windows certificates", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing windows certificates", rsp)
	util.ResponseData(c, rsp)
}

// WindowsWebDownloadCertificate 下载证书。
//
//	@Summary	下载证书
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/octet-stream
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		query		protocol.WindowsWebDownloadCertificateReq	true	"请求参数"
//	@Header		200		{string}	Content-Disposition
//	@Response	200		{file}		Response
//	@Router		/web/windows/downloadCertificate/{appId} [get]
func WindowsWebDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading windows certificate", &req)
	fileObj, err := service.WindowsWebDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download windows certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading windows certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// WindowsWebAddEVCertificate 添加 EV 证书。
//
//	@Summary	添加 EV 证书
//	@Tags		Windows-WebAPI
//	@Produce	application/json
//	@Accept		application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.WindowsWebAddEVCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/addEVCertificate [post]
func WindowsWebAddEVCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebAddEVCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for adding ev certificate", &req)
	if err = service.WindowsWebAddEVCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to add ev certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebUploadCompanyCertificate 上传公司 OV 证书。
//
//	@Summary	上传公司 OV 证书
//	@Tags		Windows-WebAPI
//	@Produce	application/json
//	@Accept		application/json
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.WindowsWebUploadCompanyCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/uploadCompanyCertificate [post]
func WindowsWebUploadCompanyCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebUploadCompanyCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	if err = service.WindowsWebUploadCompanyCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to upload company certificate", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebListCompanyCertificates 获取后台管理中的 Windows 证书。
//
//	@Summary	获取后台管理中的 Windows 证书
//	@Tags		Windows-WebAPI
//	@Accept		application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.WindowsWebListCompanyCertificatesRsp]
//	@Router		/web/windows/listCompanyCertificates [get]
func WindowsWebListCompanyCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.WindowsWebListCompanyCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing company certificates", rsp)
	util.ResponseData(c, rsp)
}

// WindowsWebGrantAppEVCertificate 授权应用使用个人 EV 证书。
//
//	@Summary	授权应用使用个人 EV 证书
//	@Tags		Windows-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.WindowsWebGrantAppEVCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/grantAppEVCertificate [post]
func WindowsWebGrantAppEVCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebGrantAppEVCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for granting app ev certificate", &req)
	err = service.WindowsWebGrantAppEVCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to grant app ev certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebGetCertificatePassword 查看证书密码。
//
//	@Summary	查看证书密码
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string											true	"应用 ID"
//	@Param		_		query		protocol.WindowsWebGetCertificatePasswordReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.WindowsWebGetCertificatePasswordRsp]
//	@Router		/web/windows/getCertificatePassword/{appId} [get]
func WindowsWebGetCertificatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebGetCertificatePasswordReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for getting windows certificate password", &req)
	rsp, err := service.WindowsWebGetCertificatePassword(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to get windows certificate password", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// WindowsWebDownloadCompanyCertificate 下载公司证书。
//
//	@Summary	下载公司证书
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/octet-stream
//	@Param		Date	header		string												true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string												true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.WindowsWebDownloadCompanyCertificateReq	true	"请求参数"
//	@Header		200		{string}	Content-Disposition
//	@Response	200		{file}		Response
//	@Router		/web/windows/downloadCompanyCertificate [get]
func WindowsWebDownloadCompanyCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebDownloadCompanyCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading company certificate", &req)
	fileObj, err := service.WindowsWebDownloadCompanyCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download company certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading company certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// WindowsWebListGrantCertificateApps 获取授权 Windows EV 证书应用列表。
//
//	@Summary	获取授权 Windows EV 证书应用列表
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.WindowsWebListGrantCertificateAppsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.WindowsWebListGrantCertificateAppsRsp]
//	@Router		/web/windows/listGrantCertificateApps [get]
func WindowsWebListGrantCertificateApps(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebListGrantCertificateAppsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing grant certificate apps", &req)
	rsp, err := service.WindowsWebListGrantCertificateApps(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list grant certificate apps", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing grant certificate apps")
	util.ResponseData(c, rsp)
}

// WindowsWebSubmitSigningJob 提交签名任务。
//
//	@Summary	提交签名任务
//	@Tags		Windows-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.WindowsWebSubmitSigningJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/submitSigningJob/{appId} [post]
func WindowsWebSubmitSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebSubmitSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting windows signing job", &req)
	err = service.WindowsWebSubmitSigningJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit windows signing job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebListSigningJobs 获取签名任务列表。
//
//	@Summary	获取签名任务列表
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		query		protocol.WindowsWebListSigningJobsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.WindowsWebListSigningJobsRsp]
//	@Router		/web/windows/listSigningJobs/{appId} [get]
func WindowsWebListSigningJobs(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebListSigningJobsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing windows signing jobs", &req)
	rsp, err := service.WindowsWebListSigningJobs(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list windows signing jobs", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing windows signing jobs", rsp)
	util.ResponseData(c, rsp)
}

// WindowsWebSubmitWHQLJob 提交 WHQL 任务。
//
//	@Summary	提交 WHQL 任务
//	@Tags		Windows-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		body		protocol.WindowsWebSubmitWHQLJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/submitWHQLJob/{appId} [post]
func WindowsWebSubmitWHQLJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebSubmitWHQLJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting whql job", &req)
	err = service.WindowsWebSubmitWHQLJob(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to submit whql job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebListWHQLJobs 获取 WHQL 任务列表。
//
//	@Summary	获取 WHQL 任务列表
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		path		protocol.WindowsWebListWHQLJobsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.WindowsWebListWHQLJobsRsp]
//	@Router		/web/windows/listWHQLJobs/{appId} [get]
func WindowsWebListWHQLJobs(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebListWHQLJobsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing whql jobs", &req)
	rsp, err := service.WindowsWebListWHQLJobs(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list whql jobs", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing whql jobs", rsp)
	util.ResponseData(c, rsp)
}

// WindowsWebRemoveCompanyCertificate 删除公司证书。
//
//	@Summary	删除公司证书
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string											true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string											true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		path		protocol.WindowsWebRemoveCompanyCertificateReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/removeCompanyCertificate [delete]
func WindowsWebRemoveCompanyCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebRemoveCompanyCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing company certificate", &req)
	if err = service.WindowsWebRemoveCompanyCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove company certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// WindowsWebDeleteCertificate 删除证书。
//
//	@Summary	删除证书
//	@Tags		Windows-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		query		protocol.WindowsWebDeleteCertificateReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/windows/deleteCertificate/{appId} [delete]
func WindowsWebDeleteCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.WindowsWebDeleteCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for deleting windows certificate", &req)
	if err = service.WindowsWebDeleteCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to delete windows certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

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
