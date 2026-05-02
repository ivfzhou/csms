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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	cl "gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// SubmitWindowsSigningJob 提交 Windows 签名任务。
func SubmitWindowsSigningJob(cfg *Configuration, token, fileID string) (string, string, bool) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	signingType := 0
	switch cfg.SignConfig.Windows.SigningType {
	case WindowsSigningTypePE:
		signingType = model.WindowsSigningJobTypePE
	case WindowsSigningTypeAttestation:
		signingType = model.WindowsSigningJobTypeAttestation
	case WindowsSigningTypePEAndAttestation:
		signingType = model.WindowsSigningJobTypePEAndAttestation
	default:
		log.Println(cl.LevelError, "invalid signing type", cfg.SignConfig.Windows.SigningType)
		return "", token, false
	}
	reqBodyBytes, _ := json.Marshal(&bp.WindowsAPISubmitSigningJobReq{
		SigningType:   signingType,
		CertificateID: cfg.SignConfig.Windows.CertificateID,
		FileID:        fileID,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPISubmitSigningJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return "", token, false
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return "", token, false
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting windows singing job again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		log.Println(cl.LevelError, "failed to submit windows singing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return "", token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try submitting windows singing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting windows singing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		return result.Data.JobID, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting windows singing job again")
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitWHQLJob 提交 WHQL 签名任务。
func SubmitWHQLJob(cfg *Configuration, token, fileID string) (string, string, bool) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	typ := 0
	switch cfg.SignConfig.WHQL.Type {
	case WHQLTypeHLK:
		typ = model.WHQLJobTypeOnlyWHQL
	case WHQLTypeTestAndHLK:
		typ = model.WHQLJobTypeHLKAndWHQL
	default:
		log.Println(cl.LevelError, "invalid type", cfg.SignConfig.WHQL.Type)
		return "", token, false
	}
	var hlkConfig string
	if len(cfg.SignConfig.WHQL.TestConfigFilePath) > 0 {
		hlkConfigBytes, err := os.ReadFile(cfg.SignConfig.WHQL.TestConfigFilePath)
		if err != nil {
			log.Println(cl.LevelError, "failed to read test config file", err)
			return "", token, false
		}
		hlkConfig = string(hlkConfigBytes)
	}
	reqBodyBytes, _ := json.Marshal(&bp.WindowsAPISubmitWHQLJobReq{
		SigningType: typ,
		TestSystem:  cfg.SignConfig.WHQL.TestSystem,
		ServiceName: cfg.SignConfig.WHQL.ServiceName,
		TestTarget:  cfg.SignConfig.WHQL.TestTarget,
		TestConfig:  hlkConfig,
		FileID:      fileID,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPISubmitWHQLJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return "", token, false
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return "", token, false
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting whql job again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		log.Println(cl.LevelError, "failed to submit whql job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return "", token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try submitting whql job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting whql job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		return result.Data.JobID, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting whql job again")
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitAndroidSigningJob 提交安卓签名任务。
func SubmitAndroidSigningJob(cfg *Configuration, token, fileID string) (string, string, bool) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	typ := 0
	switch cfg.SignConfig.Android.Type {
	case AndroidTypeAPK:
		typ = model.AndroidSigningJobTypeAPK
	case AndroidTypeAAB:
		typ = model.AndroidSigningJobTypeAAB
	case AndroidTypePatch:
		typ = model.AndroidSigningJobTypePatch
	default:
		log.Println(cl.LevelError, "invalid type", cfg.SignConfig.Android.Type)
		return "", token, false
	}
	var reqBodyBytes []byte
	var reqURL string
	switch typ {
	case model.AndroidSigningJobTypeAPK:
		reqBodyBytes, _ = json.Marshal(&bp.AndroidAPISubmitAPKSigningJobReq{
			SignatureSchema: cfg.SignConfig.Android.SignatureSchema,
			CertificateID:   cfg.SignConfig.Android.CertificateID,
			FileID:          fileID,
		})
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAPKSigningJob))
	case model.AndroidSigningJobTypeAAB:
		reqBodyBytes, _ = json.Marshal(&bp.AndroidAPISubmitAABSigningJobReq{
			CertificateID: cfg.SignConfig.Android.CertificateID,
			FileID:        fileID,
		})
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAABSigningJob))
	case model.AndroidSigningJobTypePatch:
		reqBodyBytes, _ = json.Marshal(&bp.AndroidAPISubmitAPKPatchSigningJobReq{
			SignatureSchema:   cfg.SignConfig.Android.SignatureSchema,
			CertificateID:     cfg.SignConfig.Android.CertificateID,
			FileID:            fileID,
			MinimumSDKVersion: cfg.SignConfig.Android.MinimumSDKVersion,
		})
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAPKPatchSigningJob))
	}

	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return "", token, false
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return "", token, false
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelError, "failed to submit android signing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return "", token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		var jobID string
		switch typ {
		case model.AndroidSigningJobTypeAPK:
			result := ReadAndUnmarshal[util.Response[bp.AndroidAPISubmitAPKSigningJobRsp]](response.Body)
			jobID = result.Data.JobID
		case model.AndroidSigningJobTypeAAB:
			result := ReadAndUnmarshal[util.Response[bp.AndroidAPISubmitAABSigningJobRsp]](response.Body)
			jobID = result.Data.JobID
		case model.AndroidSigningJobTypePatch:
			result := ReadAndUnmarshal[util.Response[bp.AndroidAPISubmitAPKPatchSigningJobRsp]](response.Body)
			jobID = result.Data.JobID
		}
		return jobID, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitAppleSigningJob 提交苹果签名任务。
func SubmitAppleSigningJob(cfg *Configuration, token, fileID string) (string, string, bool) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	reqBodyBytes, _ := json.Marshal(&bp.AppleAPISubmitSigningJobReq{
		FileID:    fileID,
		ProfileID: cfg.SignConfig.Apple.ProvisionID,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleAPISubmitSigningJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return "", token, false
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to send http", err)
		return "", token, false
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelError, "failed to submit android signing job", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return "", token, false
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPISubmitSigningJobRsp]](response.Body)
		return result.Data.JobID, token, true
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return "", token, false
		}
		log.Println(cl.LevelWarn, "try submitting android signing job again")
		time.Sleep(time.Second)
		goto Do
	}
}
