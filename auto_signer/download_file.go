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

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	cl "gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// DownloadFile 下载文件。
func DownloadFile(cfg *Configuration, token, fileID string) bool {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.FileAPIDownloadReq{FileID: fileID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIDownload), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return false
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return false
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return false
		}
		log.Println(cl.LevelWarn, "try downloading file again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelError, "failed to download file", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try downloading file again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusNotFound:
		CloseIO(response.Body)
		log.Println(cl.LevelError, "file not found")
		return false
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return false
		}
		log.Println(cl.LevelWarn, "try downloading file again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		return copyDataToDestination(response.Body, cfg.Base.OutFile, response.ContentLength)
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return false
		}
		log.Println(cl.LevelWarn, "try downloading file again")
		time.Sleep(time.Second)
		goto Do
	}
}

func copyDataToDestination(reader io.ReadCloser, filePath string, fileSize int64) bool {
	defer CloseIO(reader)

	// 创建文件夹。
	err := os.MkdirAll(filepath.Dir(filePath), cc.DirectoryMode)
	if err != nil {
		log.Println(cl.LevelError, "failed to create directory", err)
		return false
	}

	// 打开文件流。
	fileStream, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, cc.FileMode)
	if err != nil {
		log.Println(cl.LevelError, "failed to open file", err)
		return false
	}
	defer CloseIO(fileStream)

	// 写入文件。
	written, err := io.Copy(fileStream, reader)
	if err != nil {
		log.Println(cl.LevelError, "failed to write file", err)
		return false
	}
	if written != fileSize {
		log.Println(cl.LevelError, "written file bytes is unexpected", written, fileSize)
		return false
	}

	return true
}
