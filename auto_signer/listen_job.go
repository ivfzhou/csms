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
	"net/http"
	"path"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// ListenWindowsJob 监听任务结果。
func ListenWindowsJob(cfg *Configuration, token, jobID string, step *StepRunner) (string, string, []string, error) {
	beginTime := time.Now()
	var info *bp.WindowsAPIGetSigningJobInformationRsp
	var err error
	for range time.Tick(3 * time.Second) {
		info, token, err = getWindowsSigningJob(cfg, token, jobID)
		if err != nil {
			return token, "", nil, err
		}

		if info != nil && info.Status == model.WindowsSigningJobStatusSuccess {
			return token, info.SignedFileID, []string{fmt.Sprintf("签名文件 ID：%s", info.SignedFileID)}, nil
		}

		if info != nil && info.Status == model.WindowsSigningJobStatusFailure {
			return token, "", nil, fmt.Errorf("签名失败：%s", FormatOneline(info.Log))
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
		step.UpdateRunning(fmt.Sprintf("监听中，耗时：%s", FormatDuration(time.Since(beginTime))))
	}

	return token, "", nil, fmt.Errorf("未知错误")
}

// ListenWHQLJob 监听任务结果。
func ListenWHQLJob(cfg *Configuration, token, jobID string, step *StepRunner) (string, string, []string, error) {
	beginTime := time.Now()
	var info *bp.WindowsAPIGetWHQLJobInformationRsp
	var err error
	for range time.Tick(10 * time.Second) {
		info, token, err = getWHQLJob(cfg, token, jobID)
		if err != nil {
			return token, "", nil, err
		}

		if info != nil && info.Status == model.WHQLJobStatusSuccess {
			return token, info.SignedFileID, []string{fmt.Sprintf("签名文件 ID：%s", info.SignedFileID)}, nil
		}

		if info != nil && info.Status == model.WHQLJobStatusFailure {
			return token, "", nil, fmt.Errorf("签名失败：%s", FormatOneline(info.Log))
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
		step.UpdateRunning(fmt.Sprintf("监听中，耗时：%s", FormatDuration(time.Since(beginTime))))
	}

	return token, "", nil, fmt.Errorf("未知错误")
}

// ListenAndroidJob 监听任务结果。
func ListenAndroidJob(cfg *Configuration, token, jobID string, step *StepRunner) (string, string, []string, error) {
	beginTime := time.Now()
	var info *bp.AndroidAPIGetSigningJobInformationRsp
	var err error
	for range time.Tick(3 * time.Second) {
		info, token, err = getAndroidSigningJob(cfg, token, jobID)
		if err != nil {
			return token, "", nil, err
		}

		if info != nil && info.Status == model.AppleSigningJobStatusSuccess {
			return token, info.SignedFileID, []string{fmt.Sprintf("签名文件 ID：%s", info.SignedFileID)}, nil
		}

		if info != nil && info.Status == model.AppleSigningJobStatusFailure {
			return token, "", nil, fmt.Errorf("签名失败：%s", FormatOneline(info.Log))
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
		step.UpdateRunning(fmt.Sprintf("监听中，耗时：%s", FormatDuration(time.Since(beginTime))))
	}

	return token, "", nil, fmt.Errorf("未知错误")
}

// ListenAppleJob 监听任务结果。
func ListenAppleJob(cfg *Configuration, token, jobID string, step *StepRunner) (string, string, []string, error) {
	beginTime := time.Now()
	var info *bp.AppleAPIGetSigningJobInformationRsp
	var err error
	for range time.Tick(3 * time.Second) {
		info, token, err = getAppleSigningJob(cfg, token, jobID)
		if err != nil {
			return token, "", nil, err
		}

		if info != nil && info.Status == model.AppleSigningJobStatusSuccess {
			return token, info.SignedFileID, []string{fmt.Sprintf("签名文件 ID：%s", info.SignedFileID)}, nil
		}

		if info != nil && info.Status == model.AppleSigningJobStatusFailure {
			return token, "", nil, fmt.Errorf("签名失败：%s", FormatOneline(info.Log))
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
		step.UpdateRunning(fmt.Sprintf("监听中，耗时：%s", FormatDuration(time.Since(beginTime))))
	}

	return token, "", nil, fmt.Errorf("未知错误")
}

func getWindowsSigningJob(cfg *Configuration, token, jobID string) (
	*bp.WindowsAPIGetSigningJobInformationRsp, string, error) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.WindowsAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, token, nil
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, token, nil
	}
	if response == nil {
		return nil, token, nil
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		return nil, token, fmt.Errorf("监听任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusNotFound:
		CloseIO(response.Body)
		return nil, token, fmt.Errorf("监听任务失败，任务不存在：%s", response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		_ = ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, nil
	default:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, fmt.Errorf("监听任务失败，响应信息：%s %s %s",
			response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
	}
}

func getWHQLJob(cfg *Configuration, token, jobID string) (*bp.WindowsAPIGetWHQLJobInformationRsp, string, error) {
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.WindowsAPIGetWHQLJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPIGetWHQLJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, token, nil
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, token, nil
	}
	if response == nil {
		return nil, token, nil
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		return nil, token, fmt.Errorf("监听任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusNotFound:
		CloseIO(response.Body)
		return nil, token, fmt.Errorf("监听任务失败，任务不存在：%s", response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		_ = ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		return result.Data, token, nil
	default:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, fmt.Errorf("监听任务失败，响应信息：%s %s %s",
			response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
	}
}

func getAndroidSigningJob(cfg *Configuration, token, jobID string) (
	*bp.AndroidAPIGetSigningJobInformationRsp, string, error) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.AndroidAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, token, nil
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, token, nil
	}
	if response == nil {
		return nil, token, nil
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		return nil, token, fmt.Errorf("监听任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusNotFound:
		CloseIO(response.Body)
		return nil, token, fmt.Errorf("监听任务失败，任务不存在：%s", response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		_ = ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, nil
	default:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, fmt.Errorf("监听任务失败，响应信息：%s %s %s",
			response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
	}
}

func getAppleSigningJob(cfg *Configuration, token, jobID string) (
	*bp.AppleAPIGetSigningJobInformationRsp, string, error) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.AppleAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, token, nil
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, token, nil
	}
	if response == nil {
		return nil, token, nil
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		return nil, token, fmt.Errorf("监听任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusNotFound:
		CloseIO(response.Body)
		return nil, token, fmt.Errorf("监听任务失败，任务不存在：%s", response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		_ = ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		time.Sleep(time.Second)
		return nil, token, nil
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, nil
	default:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		return nil, token, fmt.Errorf("监听任务失败，响应信息：%s %s %s",
			response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
	}
}
