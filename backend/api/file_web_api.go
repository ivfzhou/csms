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

// FileWebDownload 下载。
//
//	@Summary	下载
//	@Tags		File-WebAPI
//	@Accept		application/x-www-form-urlencoded
//	@Produce	application/octet-stream
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.FileWebDownloadReq	true	"请求参数"
//	@Response	200		{file}		Response
//	@Header		200		{string}	Content-Disposition
//	@Router		/web/file/download [get]
func FileWebDownload(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileWebDownloadReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading file", &req)
	fileObj, err := service.FileWebDownload(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download file", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading file", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// FileWebInitial 初始化上传。
//
//	@Summary	初始化上传
//	@Tags		File-WebAPI
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Date	header		string						true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string						true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		body		protocol.FileWebInitialReq	true	"请求体"
//	@Response	200		{object}	util.Response[protocol.FileWebInitialRsp]
//	@Router		/web/file/initial [post]
func FileWebInitial(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileWebInitialReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for initializing uploading file", &req)
	rsp, err := service.FileWebInitial(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to initialize uploading file", err, &req)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "response data for initializing uploading file", rsp)
	util.ResponseData(c, rsp)
}

// FileWebUploadPart 上传分片。
//
//	@Summary	上传分片
//	@Tags		File-WebAPI
//	@Accept		application/form-data
//	@Produce	application/json
//	@Param		Date		header		string	true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie		header		string	true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		chunk		formData	file	true	"文件分片"
//	@Param		fileId		formData	string	true	"文件 ID"		example(2026014ef83c03e2ce4f1f94c11168d1acd087)
//	@Param		chunkNumber	formData	integer	true	"文件分片序号"	example(1)
//	@Response	200			{object}	util.Response[any]
//	@Router		/web/file/uploadPart [post]
func FileWebUploadPart(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileWebUploadPartReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for uploading file part",
		req.FileID, req.ChunkNumber, req.Chunk.Filename, req.Chunk.Size)
	if err = service.FileWebUploadPart(ctx, &req); err != nil {
		log.Warn(ctx, "failed to upload file part", err,
			req.FileID, req.ChunkNumber, req.Chunk.Filename, req.Chunk.Size)
		util.ResponseError(c, err)
		return
	}
	util.ResponseData[any](c, nil)
}

// FileWebMergeParts 合并分片。
//
//	@Summary	合并分片
//	@Tags		File-WebAPI
//	@Accept		application/x-www-urlencoded
//	@Produce	application/json
//	@Param		Date	header		string							true	"请求日期"	example(Mon, 02 Jan 2006 15:04:05 GMT)
//	@Param		Cookie	header		string							true	"会话凭据"	example(csms_user=; csms_seesion=)
//	@Param		_		query		protocol.FileWebMergePartsReq	true	"请求参数"
//	@Response	200		{object}	util.Response[any]
//	@Router		/web/file/mergeParts [get]
func FileWebMergeParts(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileWebMergePartsReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseError(c, err)
		return
	}
	log.Info(ctx, "request parameters for merge file parts", &req)
	if err = service.FileWebMergeParts(ctx, &req); err != nil {
		log.Warn(ctx, "failed to merge file parts", err, &req)
		util.ResponseError(c, err)
		return
	}
	util.ResponseCode(c, consts.AlertFileUpload)
}
