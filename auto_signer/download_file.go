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
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/util"
)

// DownloadFile 下载文件。
func DownloadFile(cfg *Configuration, token, fileID string, step *StepRunner) ([]string, error) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.FileAPIDownloadReq{FileID: fileID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIDownload), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败：%v", err)
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return nil, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		return nil, fmt.Errorf("下载文件失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusNotFound:
		CloseIO(response.Body)
		return nil, fmt.Errorf("下载文件失败，文件不存在")
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return nil, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		return copyDataToDestination(response.Body, cfg.Base.OutFile, response.ContentLength, step)
	default:
		retryTimes--
		if retryTimes < 0 {
			return nil, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		time.Sleep(time.Second)
		goto Do
	}
}

func copyDataToDestination(reader io.ReadCloser, filePath string, fileSize int64, step *StepRunner) ([]string, error) {
	defer CloseIO(reader)

	// 创建文件夹。
	err := os.MkdirAll(filepath.Dir(filePath), cc.DirectoryMode)
	if err != nil {
		return nil, fmt.Errorf("创建文件夹失败，请检查输出文件夹路径：%v", err)
	}

	// 打开文件流。
	fileStream, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, cc.FileMode)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败，请检查输出文件路径：%v", err)
	}
	defer CloseIO(fileStream)

	// 写入文件，使用进度回调。
	progressReader := NewProgressReader(reader, fileSize, 200*time.Millisecond, func(s string) {
		step.UpdateRunning(s)
	})
	written, err := io.Copy(fileStream, progressReader)
	if err != nil {
		return nil, fmt.Errorf("写入硬盘文件失败：%v", err)
	}
	progressReader.Finish()
	if written != fileSize {
		return nil, fmt.Errorf("写入硬盘文件字节数不符合预期：%d!=%d", written, fileSize)
	}

	return []string{fmt.Sprintf("下载平均速度 %s/s", FormatSize(progressReader.GetSpeed()))}, nil
}
