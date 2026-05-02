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
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	cl "gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// ListenWindowsJob 监听任务结果。
func ListenWindowsJob(cfg *Configuration, token, jobID string) (string, string, bool) {
	beginTime := time.Now()
	var info *bp.WindowsAPIGetSigningJobInformationRsp
	var ok bool
	for range time.Tick(3 * time.Second) {
		info, token, ok = getWindowsSigningJob(cfg, token, jobID)
		if !ok {
			return "", token, false
		}

		if info != nil && len(info.SignedFileID) > 0 {
			return info.SignedFileID, token, true
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
	}

	return "", token, false
}

// ListenWHQLJob 监听任务结果。
func ListenWHQLJob(cfg *Configuration, token, jobID string) (string, string, bool) {
	beginTime := time.Now()
	var info *bp.WindowsAPIGetWHQLJobInformationRsp
	var ok bool
	for range time.Tick(10 * time.Second) {
		info, token, ok = getWHQLJob(cfg, token, jobID)
		if !ok {
			return "", token, false
		}

		if info != nil && len(info.SignedFileID) > 0 {
			return info.SignedFileID, token, true
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
	}

	return "", token, false
}

// ListenAndroidJob 监听任务结果。
func ListenAndroidJob(cfg *Configuration, token, jobID string) (string, string, bool) {
	beginTime := time.Now()
	var info *bp.AndroidAPIGetSigningJobInformationRsp
	var ok bool
	for range time.Tick(3 * time.Second) {
		info, token, ok = getAndroidSigningJob(cfg, token, jobID)
		if !ok {
			return "", token, false
		}

		if info != nil && len(info.SignedFileID) > 0 {
			return info.SignedFileID, token, true
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
	}

	return "", token, false
}

// ListenAppleJob 监听任务结果。
func ListenAppleJob(cfg *Configuration, token, jobID string) (string, string, bool) {
	beginTime := time.Now()
	var info *bp.AppleAPIGetSigningJobInformationRsp
	var ok bool
	for range time.Tick(3 * time.Second) {
		info, token, ok = getAppleSigningJob(cfg, token, jobID)
		if !ok {
			return "", token, false
		}

		if info != nil && len(info.SignedFileID) > 0 {
			return info.SignedFileID, token, true
		}

		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			token, _ = CreateAuthorization(cfg)
			beginTime = time.Now()
		}
	}

	return "", token, false
}

func getWindowsSigningJob(cfg *Configuration, token, jobID string) (
	*bp.WindowsAPIGetSigningJobInformationRsp, string, bool) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.WindowsAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return nil, token, true
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return nil, token, true
	}
	if response == nil {
		return nil, token, true
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelError, "failed to get windows signing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return nil, token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try getting windows signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusNotFound:
		CloseIO(response.Body)
		log.Println(cl.LevelError, "windows signing job not found")
		return nil, token, false
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		log.Println(cl.LevelWarn, "try getting windows signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		log.Println(cl.LevelWarn, "try getting windows signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	}
}

func getWHQLJob(cfg *Configuration, token, jobID string) (*bp.WindowsAPIGetWHQLJobInformationRsp, string, bool) {
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.WindowsAPIGetWHQLJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPIGetWHQLJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return nil, token, true
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return nil, token, true
	}
	if response == nil {
		return nil, token, true
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		log.Println(cl.LevelError, "failed to get whql job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return nil, token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try getting whql job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusNotFound:
		CloseIO(response.Body)
		log.Println(cl.LevelError, "whql job not found")
		return nil, token, false
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		log.Println(cl.LevelWarn, "try getting whql job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPIGetWHQLJobInformationRsp]](response.Body)
		return result.Data, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		log.Println(cl.LevelWarn, "try getting whql job again")
		time.Sleep(time.Second)
		return nil, token, true
	}
}

func getAndroidSigningJob(cfg *Configuration, token, jobID string) (
	*bp.AndroidAPIGetSigningJobInformationRsp, string, bool) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.AndroidAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return nil, token, true
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return nil, token, true
	}
	if response == nil {
		return nil, token, true
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelError, "failed to get android signing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return nil, token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try getting android signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusNotFound:
		CloseIO(response.Body)
		log.Println(cl.LevelError, "android signing job not found")
		return nil, token, false
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		log.Println(cl.LevelWarn, "try getting android signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AndroidAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		log.Println(cl.LevelWarn, "try getting android signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	}
}

func getAppleSigningJob(cfg *Configuration, token, jobID string) (
	*bp.AppleAPIGetSigningJobInformationRsp, string, bool) {

	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.AppleAPIGetSigningJobInformationReq{JobID: jobID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleAPIGetSigningJobInformation), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return nil, token, true
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return nil, token, true
	}
	if response == nil {
		return nil, token, true
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelError, "failed to get apple signing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return nil, token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try getting apple signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusNotFound:
		CloseIO(response.Body)
		log.Println(cl.LevelError, "apple signing job not found")
		return nil, token, false
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		log.Println(cl.LevelWarn, "try getting apple signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPIGetSigningJobInformationRsp]](response.Body)
		return result.Data, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		log.Println(cl.LevelWarn, "try getting apple signing job again")
		time.Sleep(time.Second)
		return nil, token, true
	}
}
