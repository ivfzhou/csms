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

// FileInternalDownload 下载文件。
func FileInternalDownload(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileInternalDownloadReq
	err := c.ShouldBind(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for downloading file", &req)
	fileObj, err := service.FileInternalDownload(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to download file", err, &req)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for downloading file", fileObj.Size, fileObj.Name)
	util.ResponseStream(c, fileObj.Size, fileObj.Name, fileObj.Reader)
}

// FileInternalUpload 文件上传。
func FileInternalUpload(c *gin.Context) {
	ctx := c.Request.Context()
	var req protocol.FileInternalUploadReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		log.Warn(ctx, "failed to parse request parameters", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "request parameters for uploading file", &req)
	req.Body = c.Request.Body
	req.Size = c.Request.ContentLength
	rsp, err := service.FileInternalUpload(ctx, &req)
	if err != nil {
		log.Warn(ctx, "failed to upload file", err)
		util.ResponseAPIError(c, err)
		return
	}
	log.Info(ctx, "response data for uploading file", rsp)
	util.ResponseData(c, &rsp)
}
