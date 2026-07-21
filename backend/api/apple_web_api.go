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

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AppleWebApplyBundleID 申请 Bundle ID。
//
//	@Summary	申请 Bundle ID
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		body		protocol.AppleWebApplyBundleIDReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyBundleID/{appId} [post]
func AppleWebApplyBundleID(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebApplyBundleIDReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying apple bundle id", &req)
	if err = service.AppleWebApplyBundleID(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply apple bundle id", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebModifyBundleID 修改 Bundle ID 能力项。
//
//	@Summary	修改 Bundle ID 能力项
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		body		protocol.AppleWebModifyBundleIDReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/modifyBundleID/{appId} [post]
func AppleWebModifyBundleID(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebModifyBundleIDReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for modifying apple bundle id", &req)
	b, err := service.AppleWebModifyBundleID(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to modify apple bundle id", err, &req)
		util.ResponseError(c, err)
		return
	}
	if b {
		util.ResponseCode(c, consts.AlertErrorInModifyingAppleBundleID)
	} else {
		util.ResponseCode(c, consts.AlertSuccess)
	}
}

// AppleWebApplyCertificate 申请苹果签名证书。
//
//	@Summary	申请苹果签名证书
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.AppleWebApplyCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyCertificate [post]
func AppleWebApplyCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebApplyCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying certificate", &req)
	if err = service.AppleWebApplyCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebListBundleIDs 获取 Bundle ID 列表。
//
//	@Summary	获取 Bundle ID 列表
//	@Tags		Apple-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"
//	@Response	200		{object}	util.Response[protocol.AppleWebListBundleIDsRsp]
//	@Router		/web/apple/listBundleIDs/{appId} [get]
func AppleWebListBundleIDs(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AppleWebListBundleIDs(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list apple bundle ids", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing apple bundle ids", rsp)
	util.ResponseData(c, rsp)
}

// AppleWebListCertificates 获取苹果证书信息列表。
//
//	@Summary	获取苹果证书信息列表
//	@Tags		Apple-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.AppleWebListCertificatesRsp]
//	@Router		/web/apple/listCertificates [get]
func AppleWebListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AppleWebListCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list apple certificates", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing apple certificates", rsp)
	util.ResponseData(c, rsp)
}

// AppleWebRegisterDevice 注册测试设备。
//
//	@Summary	注册测试设备
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		body		protocol.AppleWebRegisterDeviceReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/registerDevice/{appId} [post]
func AppleWebRegisterDevice(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebRegisterDeviceReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for registering apple device", &req)
	if err = service.AppleWebRegisterDevice(ctx, &req); err != nil {
		log.Warn(ctx, "failed to register apple device", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertRegisterAppleDevice)
}

// AppleWebListDevices 设备列表。
//
//	@Summary	设备列表
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string							true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string							true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string							true	"应用 ID"
//	@Param		_		query		protocol.AppleWebListDevicesReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AppleWebListDevicesRsp]
//	@Router		/web/apple/listDevices/{appId} [get]
func AppleWebListDevices(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebListDevicesReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing apple profile", &req)
	rsp, err := service.AppleWebListDevices(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list apple profile", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData(c, rsp)
}

// AppleWebApplyProfile 申请描述文件。
//
//	@Summary	申请描述文件
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		body		protocol.AppleWebApplyProfileReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyProfile/{appId} [post]
func AppleWebApplyProfile(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebApplyProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying apple profile", &req)
	if err = service.AppleWebApplyProfile(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply apple profile", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebApplyInHouseProfile 申请企业内测描述文件。
//
//	@Summary	申请企业内测描述文件
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.AppleWebApplyInHouseProfileReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyInHouseProfile/{appId} [post]
func AppleWebApplyInHouseProfile(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebApplyInHouseProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying apple in house profile", &req)
	if err = service.AppleWebApplyInHouseProfile(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply apple in house profile", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebApplyCommonProfile 申请企业内测通配符描述文件。
//
//	@Summary	申请企业内测通配符描述文件
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string	true	"应用 ID"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyCommonProfile/{appId} [post]
func AppleWebApplyCommonProfile(c *gin.Context) {
	ctx := c.Request.Context()
	err := service.AppleWebApplyCommonProfile(ctx)
	if err != nil {
		log.Warn(ctx, "failed to apply apple common profile", err)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebApplyPushCertificate 申请 Push 证书。
//
//	@Summary	申请 Push 证书
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		body		protocol.AppleWebApplyPushCertificateReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/applyPushCertificate/{appId} [post]
func AppleWebApplyPushCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebApplyPushCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for applying push certificate", &req)
	if err = service.AppleWebApplyPushCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to apply apple push certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebDeleteBundleID 删除 Bundle ID。
//
//	@Summary	删除 Bundle ID
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		query		protocol.AppleWebDeleteBundleIDReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/deleteBundleID/{appId} [delete]
func AppleWebDeleteBundleID(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebDeleteBundleIDReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for deleting apple bundle id", &req)
	if err = service.AppleWebDeleteBundleID(ctx, &req); err != nil {
		log.Warn(ctx, "failed to delete apple bundle id", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebRemoveCertificate 删除苹果证书。
//
//	@Summary	删除苹果证书
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.AppleWebRemoveCertificateReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/removeCertificate/{appId} [delete]
func AppleWebRemoveCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebRemoveCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing certificate", &req)
	if err = service.AppleWebRemoveCertificate(ctx, &req); err != nil {
		log.Warn(ctx, "failed to remove certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebListAppCertificates 获取应用证书和描述文件列表。
//
//	@Summary	获取应用证书和描述文件列表
//	@Tags		Apple-WebAPI
//	@Produce	application/json
//	@Param		Date	header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Response	200		{object}	util.Response[protocol.AppleWebListAppCertificatesRsp]
//	@Router		/web/apple/listAppCertificates/{appId} [get]
func AppleWebListAppCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	rsp, err := service.AppleWebListAppCertificates(ctx)
	if err != nil {
		log.Warn(ctx, "failed to list app certificates", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing app certificates", rsp)
	util.ResponseData(c, rsp)
}

// AppleWebSubmitSigningJob 提交签名任务。
//
//	@Summary	提交签名任务
//	@Tags		Apple-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		body		protocol.AppleWebSubmitSigningJobReq	true	"请求体"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/submitSigningJob/{appId} [post]
func AppleWebSubmitSigningJob(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebSubmitSigningJobReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for submitting apple signing job", &req)
	if err = service.AppleWebSubmitSigningJob(ctx, &req); err != nil {
		log.Warn(ctx, "failed to submit apple signing job", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebListSigningJobs 获取签名任务信息列表。
//
//	@Summary	获取签名任务信息列表
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		query		protocol.AppleWebListSigningJobsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AppleWebListSigningJobsRsp]
//	@Router		/web/apple/listSigningJobs/{appId} [get]
func AppleWebListSigningJobs(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebListSigningJobsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for listing apple signing jobs", &req)
	rsp, err := service.AppleWebListSigningJobs(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to list apple signing jobs", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for listing apple signing jobs", rsp)
	util.ResponseData(c, rsp)
}

// AppleWebRemoveProfile 删除描述文件。
//
//	@Summary	删除描述文件
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string								true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string								true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string								true	"应用 ID"
//	@Param		_		query		protocol.AppleWebRemoveProfileReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/removeProfile/{appId} [delete]
func AppleWebRemoveProfile(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebRemoveProfileReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing apple profile", &req)
	err = service.AppleWebRemoveProfile(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to remove apple profile", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebRemovePushCertificate 删除 Push 证书。
//
//	@Summary	删除 Push 证书
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string										true	"应用 ID"
//	@Param		_		query		protocol.AppleWebRemovePushCertificateReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/apple/removePushCertificate/{appId} [delete]
func AppleWebRemovePushCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebRemovePushCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for removing apple push certificate", &req)
	err = service.AppleWebRemovePushCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to remove apple push certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertSuccess)
}

// AppleWebDownloadCertificate 下载证书和描述文件。
//
//	@Summary	下载证书和描述文件
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/otect-stream
//	@Param		Date	header		string									true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string									true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		appId	path		string									true	"应用 ID"
//	@Param		_		query		protocol.AppleWebDownloadCertificateReq	true	"请求参数"
//	@Response	200		{file}		Response
//	@Header		200		{string}	Content-Disposition
//	@Router		/web/apple/downloadCertificate/{appId} [get]
func AppleWebDownloadCertificate(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebDownloadCertificateReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading apple certificate", &req)
	fileObj, err := service.AppleWebDownloadCertificate(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download apple certificate", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading apple certificate", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// AppleWebStatisticSigningTimes 获取应用的 Apple 类型签名次数统计信息。
//
//	@Summary	获取应用的 Apple 类型签名次数统计信息
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.AppleWebStatisticSigningTimesReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AppleWebStatisticSigningTimesRsp]
//	@Router		/web/apple/statisticSigningTimes [get]
func AppleWebStatisticSigningTimes(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebStatisticSigningTimesReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for statistic job signing times", &req)
	rsp, err := service.AppleWebStatisticSigningTimes(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to statistic job signing times", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for statistic job signing times", rsp)
	util.ResponseData(c, rsp)
}

// AppleWebStatisticSigningCost 获取应用的 Apple 类型签名耗时统计信息。
//
//	@Summary	获取应用的 Apple 类型签名次数耗时统计信息
//	@Tags		Apple-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string										true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string										true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.AppleWebStatisticSigningCostReq	true	"请求参数"
//	@Response	200		{object}	util.Response[protocol.AppleWebStatisticSigningCostRsp]
//	@Router		/web/apple/statisticSigningCost [get]
func AppleWebStatisticSigningCost(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.AppleWebStatisticSigningCostReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for statistic job signing cost", &req)
	rsp, err := service.AppleWebStatisticSigningCost(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to statistic job signing cost", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for statistic job signing cost", rsp)
	util.ResponseData(c, rsp)
}
