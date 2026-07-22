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
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// SubmitWindowsSigningJob 提交 Windows 签名任务。
func SubmitWindowsSigningJob(cfg *Configuration, token, fileID string) (string, string, []string, error) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	signingType := 0
	switch cfg.SignConfig.Windows.Type {
	case WindowsSigningTypePE:
		signingType = model.WindowsSigningJobTypePE
	case WindowsSigningTypeAttestation:
		signingType = model.WindowsSigningJobTypeAttestation
	case WindowsSigningTypePEAndAttestation:
		signingType = model.WindowsSigningJobTypePEAndAttestation
	default:
		return token, "", nil, fmt.Errorf("任务类型非法，请检查：%v", cfg.SignConfig.Windows.Type)
	}
	reqBodyBytes, _ := json.Marshal(&bp.WindowsAPISubmitSigningJobReq{
		SigningType:   signingType,
		CertificateID: cfg.SignConfig.Windows.CertificateID,
		FileID:        fileID,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPISubmitSigningJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return token, "", nil, fmt.Errorf("创建 HTTP 请求体失败：%v", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return token, "", nil, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		return token, "", nil, fmt.Errorf("提交签名任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitSigningJobRsp]](response.Body)
		return token, result.Data.JobID, []string{fmt.Sprintf("任务 ID：%s", result.Data.JobID)}, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitWHQLJob 提交 WHQL 签名任务。
func SubmitWHQLJob(cfg *Configuration, token, fileID string) (string, string, []string, error) {
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
		return token, "", nil, fmt.Errorf("任务类型非法，请检查：%v", cfg.SignConfig.WHQL.Type)
	}
	var hlkConfig string
	if len(cfg.SignConfig.WHQL.TestConfigFilePath) > 0 {
		hlkConfigBytes, err := os.ReadFile(cfg.SignConfig.WHQL.TestConfigFilePath)
		if err != nil {
			return token, "", nil, fmt.Errorf("读取 HLK 测试配置文件失败，请检查文件路径：%v", err)
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
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsAPISubmitWHQLJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return token, "", nil, fmt.Errorf("创建 HTTP 请求体失败：%v", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return token, "", nil, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		return token, "", nil, fmt.Errorf("提交签名任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.WindowsAPISubmitWHQLJobRsp]](response.Body)
		return token, result.Data.JobID, []string{fmt.Sprintf("任务 ID：%s", result.Data.JobID)}, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitAndroidSigningJob 提交安卓签名任务。
func SubmitAndroidSigningJob(cfg *Configuration, token, fileID string) (string, string, []string, error) {
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
		return token, "", nil, fmt.Errorf("任务类型非法，请检查：%v", cfg.SignConfig.Android.Type)
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
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAPKSigningJob))
	case model.AndroidSigningJobTypeAAB:
		reqBodyBytes, _ = json.Marshal(&bp.AndroidAPISubmitAABSigningJobReq{
			CertificateID: cfg.SignConfig.Android.CertificateID,
			FileID:        fileID,
		})
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAABSigningJob))
	case model.AndroidSigningJobTypePatch:
		reqBodyBytes, _ = json.Marshal(&bp.AndroidAPISubmitAPKPatchSigningJobReq{
			SignatureSchema:   cfg.SignConfig.Android.SignatureSchema,
			CertificateID:     cfg.SignConfig.Android.CertificateID,
			FileID:            fileID,
			MinimumSDKVersion: cfg.SignConfig.Android.MinimumSDKVersion,
		})
		reqURL = fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
			path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidAPISubmitAPKPatchSigningJob))
	}

	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return token, "", nil, fmt.Errorf("创建 HTTP 请求体失败：%v", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return token, "", nil, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		return token, "", nil, fmt.Errorf("提交签名任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
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
		return token, jobID, []string{fmt.Sprintf("任务 ID：%s", jobID)}, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	}
}

// SubmitAppleSigningJob 提交苹果签名任务。
func SubmitAppleSigningJob(cfg *Configuration, token, fileID string) (string, string, []string, error) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	reqBodyBytes, _ := json.Marshal(&bp.AppleAPISubmitSigningJobReq{
		FileID:    fileID,
		ProfileID: cfg.SignConfig.Apple.ProvisionID,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleAPISubmitSigningJob))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return token, "", nil, fmt.Errorf("创建 HTTP 请求体失败：%v", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return token, "", nil, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		return token, "", nil, fmt.Errorf("提交签名任务失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.AppleAPISubmitSigningJobRsp]](response.Body)
		return token, result.Data.JobID, []string{fmt.Sprintf("任务 ID：%s", result.Data.JobID)}, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return token, "", nil, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	}
}
