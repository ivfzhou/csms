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

package util

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	httpClient            *http.Client
	transport             *http.Transport
	initialHttpClientOnce sync.Once
)

// GetHTTPClient 返回单例 HTTP 客户端对象。
func GetHTTPClient() *http.Client {
	initialHttpClientOnce.Do(func() {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.Get().TLSInsecureSkipVerify(),
			},
		}
		httpClient = &http.Client{Transport: transport}
	})
	transport.TLSClientConfig.InsecureSkipVerify = cfg.Get().TLSInsecureSkipVerify()
	return httpClient
}

// HTTPGet 发送 HTTP GET 请求。
func HTTPGet(ctx context.Context, reqURL string, headers ...any) (status int, rspBody []byte, err error) {
	// 创建请求体。
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, rspBody, nil
}

// HTTPPost 发送 HTTP POST 请求。
func HTTPPost(ctx context.Context, reqURL string, reqBody io.Reader, headers ...any) (
	status int, rspBody []byte, err error) {

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPost, reqURL, reqBody)
	if err != nil {
		return 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, rspBody, nil
}

// HTTPDelete 发送 HTTP DELETE 请求。
func HTTPDelete(ctx context.Context, reqURL string, headers ...any) (status int, rspBody []byte, err error) {
	// 创建请求体。
	request, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, rspBody, nil
}

// HTTPGetToJSON 发送 HTTP GET 请求。响应体是 JSON 格式。
func HTTPGetToJSON[T any](ctx context.Context, reqURL string, headers ...any) (
	rsp *T, status int, rspBody []byte, err error) {

	// 创建请求体。
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return nil, 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	// 反序列化结果。
	var result T
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Warn(ctx, "failed to unmarshal http body", err)
		return nil, response.StatusCode, rspBody, nil
	}

	return &result, response.StatusCode, rspBody, nil
}

// HTTPPostToJSON 发送 HTTP POST 请求。响应体是 JSON 格式。
func HTTPPostToJSON[T any](ctx context.Context, reqURL string, reqBody io.Reader, bodySize int64, headers ...any) (
	rsp *T, status int, rspBody []byte, err error) {

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPost, reqURL, reqBody)
	if err != nil {
		return nil, 0, nil, err
	}
	request.ContentLength = bodySize
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return nil, 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	// 反序列化结果。
	var result T
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Warn(ctx, "failed to unmarshal http body", err)
		return nil, response.StatusCode, rspBody, nil
	}

	return &result, response.StatusCode, rspBody, nil
}

// HTTPGetToDisk 发送 HTTP GET 请求。将响应数据写入外存。
func HTTPGetToDisk(ctx context.Context, reqURL, filePath string, headers ...any) (
	status int, fileName string, fileSize int64, fileMD5 string, err error) {

	// 创建文件句柄。
	err = os.MkdirAll(filepath.Dir(filePath), consts.DirectoryMode)
	if err != nil {
		return 0, "", 0, "", err
	}
	fileStream, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, consts.FileMode)
	if err != nil {
		return 0, "", 0, "", err
	}
	defer CloseIO(ctx, fileStream)

	// 创建请求体。
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, "", 0, "", err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, "", 0, "", err
	}
	defer CloseHTTPBody(ctx, response)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		bs, _ := io.ReadAll(response.Body)
		return response.StatusCode, "", 0, "", fmt.Errorf("http failed %d %s", response.StatusCode, bs)
	}

	// 读取结果，计算结果 MD5。
	md5Hash := md5.New()
	written, err := io.Copy(io.MultiWriter(fileStream, md5Hash), response.Body)
	if err != nil {
		return response.StatusCode, "", 0, "", err
	}
	if written != response.ContentLength {
		log.Error(ctx, "size of the http response byte is inconsistent with the size of the written file")
	}
	fileMD52 := md5Hash.Sum(nil)

	// 提取文件名。
	fileName = extractFileName(response.Header.Get("Content-Disposition"))

	return response.StatusCode, fileName, written, hex.EncodeToString(fileMD52[:]), nil
}

// HTTPGetToDisk2 将文件下载到指定的文件夹下，
func HTTPGetToDisk2(ctx context.Context, reqURL, fileDirectoryPath string, headers ...any) (
	status int, filePath string, fileSize int64, err error) {

	// 创建文件夹。
	err = os.MkdirAll(fileDirectoryPath, consts.DirectoryMode)
	if err != nil {
		return 0, "", 0, err
	}

	// 创建请求体。
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, "", 0, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, "", 0, err
	}
	defer CloseHTTPBody(ctx, response)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		bs, _ := io.ReadAll(response.Body)
		return response.StatusCode, "", 0, fmt.Errorf("http failed %d %s", response.StatusCode, bs)
	}

	// 提取文件名。
	fileName := extractFileName(response.Header.Get("Content-Disposition"))

	// 写入文件。
	filePath = filepath.Join(fileDirectoryPath, fileName)
	fileStream, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, consts.FileMode)
	if err != nil {
		return 0, "", 0, err
	}
	defer CloseIO(ctx, fileStream)
	written, err := io.Copy(fileStream, response.Body)
	if err != nil {
		return response.StatusCode, "", 0, err
	}
	if written != response.ContentLength {
		log.Error(ctx, "size of the http response byte is inconsistent with the size of the written file")
	}

	return response.StatusCode, filePath, response.ContentLength, nil
}

// HTTPDeleteToJSON 发送 HTTP DELETE 请求。
func HTTPDeleteToJSON[T any](ctx context.Context, reqURL string, headers ...any) (
	rsp *T, httpCode int, rspBody []byte, err error) {

	// 创建请求体。
	request, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return nil, 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	//读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	// 反序列化结果。
	var result T
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Warn(ctx, "failed to unmarshal http body", err)
		return nil, response.StatusCode, rspBody, nil
	}

	return &result, response.StatusCode, rspBody, nil
}

// HTTPPostJSONToJSON 发送 HTTP POST 请求。请求体和响应体都是 JSON 格式。
func HTTPPostJSONToJSON[T1, T2 any](ctx context.Context, reqURL string, req T2, headers ...any) (
	rsp *T1, status int, rspBody []byte, err error) {

	// 序列化请求体。
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, 0, nil, err
	}

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return nil, 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	// 反序列化结果。
	var result T1
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Warn(ctx, "failed to unmarshal http body", err)
		return nil, response.StatusCode, rspBody, nil
	}

	return &result, response.StatusCode, rspBody, nil
}

// HTTPPostJSON 发送 HTTP POST 请求。请求体是 JSON 格式。
func HTTPPostJSON[T any](ctx context.Context, reqURL string, req T, headers ...any) (
	status int, rspBody []byte, err error) {

	// 序列化请求体。
	reqBody, err := json.Marshal(req)
	if err != nil {
		return 0, nil, err
	}

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}
	request = request.WithContext(ctx)

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return 0, nil, err
	}

	// 读取结果。
	defer CloseHTTPBody(ctx, response)
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, rspBody, nil
}

// HTTPPostFormToJSON 发送 HTTP POST 请求。请求体是 Form 表单，响应体是 JSON 格式。
func HTTPPostFormToJSON[T any](ctx context.Context, reqURL string, req url.Values, headers ...any) (
	rsp *T, status int, rspBody []byte, err error) {

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(req.Encode()))
	if err != nil {
		return nil, 0, nil, err
	}
	requestID := ctxs.RequestID(ctx)
	request.Header.Set(consts.HTTPHeaderRequestID, requestID)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for i := 0; i < len(headers)-1; i += 2 {
		request.Header.Set(fmt.Sprint(headers[i]), fmt.Sprint(headers[i+1]))
	}

	// 发送请求。
	response, err := GetHTTPClient().Do(request)
	if err != nil {
		return nil, 0, nil, err
	}
	defer CloseHTTPBody(ctx, response)

	// 读取结果。
	rspBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	// 反序列化结果。
	var result T
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Warn(ctx, "failed to unmarshal http body", err)
		return nil, response.StatusCode, rspBody, nil
	}

	return &result, response.StatusCode, rspBody, nil
}

func extractFileName(contentDisposition string) string {
	for v := range strings.SplitSeq(contentDisposition, ";") {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "filename") {
			pair := strings.Split(v, "=")
			if len(pair) >= 2 {
				return strings.Trim(strings.TrimSpace(pair[1]), `"`)
			}
		}
	}
	return contentDisposition
}
