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

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// TODO: 补全优化枚举。
const (
	PETypeAmd64 peType = 1 + iota
	PETypeAmd32
	PETypeArm64
	PETypeArm32
)

type peType int

func getMicrosoftAccessToken(ctx context.Context) (string, error) {
	// 读取缓存，有缓存就返回缓存值。
	token, err := conn.RedisClient(ctx).Get(ctx, consts.RedisKeyMicrosoftAccessToken).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to get microsoft access token from redis", err)
	}
	if len(token) > 0 {
		log.Debug(ctx, "return microsoft access token from redis")
		return token, nil
	}

	// 创建请求对象。
	reqBodyObj := url.Values{}
	reqBodyObj.Set("grant_type", cfg.Get().MicrosoftAPI().GrantType())
	reqBodyObj.Set("client_id", cfg.Get().MicrosoftAPI().ClientID())
	reqBodyObj.Set("client_secret", cfg.Get().MicrosoftAPI().ClientSecret())
	reqBodyObj.Set("resource", cfg.Get().MicrosoftAPI().Resource())

	// 发送 HTTP 请求。
	rspBodyObj, httpCode, rspBody, err := util.HTTPPostFormToJSON[struct {
		AccessToken      string `json:"access_token"`
		ErrorDescription string `json:"error_description"`
		TokenType        string `json:"token_type"`
		ExpiresOn        string `json:"expires_on"`
	}](ctx, fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", cfg.Get().MicrosoftAPI().TenantID()), reqBodyObj)
	if err != nil {
		log.Error(ctx, "failed to get microsoft access token", err, httpCode, rspBody)
		return "", errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusOK {
		log.Error(ctx, "response code is not ok when getting microsoft access token", httpCode, rspBodyObj)
		return "", errs.New(consts.ErrSystem)
	}
	if len(rspBodyObj.AccessToken) <= 0 {
		log.Error(ctx, "no access token when getting microsoft access token", rspBodyObj.ErrorDescription)
		return "", errs.New(consts.ErrSystem)
	}

	// 保存到缓存中。
	log.Info(ctx, "cache microsoft access token")
	token = fmt.Sprintf("%s %s", rspBodyObj.TokenType, rspBodyObj.AccessToken)
	expirationTime := time.Unix(int64(util.Atoi(rspBodyObj.ExpiresOn)), 0)
	duration := time.Until(expirationTime)
	if duration > time.Minute {
		err = conn.RedisClient(ctx).
			Set(ctx, consts.RedisKeyMicrosoftAccessToken, token, duration-time.Minute).Err()
		if err != nil {
			log.Error(ctx, "failed to cache microsoft access token", err)
		}
	}

	return token, nil
}

func createWHQLProduct(ctx context.Context, productName string) (string, error) {
	// 获取 AccessToken。
	token, err := getMicrosoftAccessToken(ctx)
	if err != nil {
		return "", err
	}

	// 构建请求体。
	reqBodyObj := map[string]any{
		"productName":       productName,
		"testHarness":       "hlk",
		"announcementDate":  "2024-09-11T14:26:49",
		"deviceMetaDataIds": []string{},
		"firmwareVersion":   "980",
		"deviceType":        "internalExternal",
		"isTestSign":        false,
		"isFlightSign":      false,
		"markettingNames":   []string{},
		"selectedProductTypes": map[string]string{
			"Windows_v100_RS3": "Unclassified",
		},
		"requestedSignatures": []string{
			"WINDOWS_2008_SERVER",
			"WINDOWS_2008_SERVER_X64",
			"WINDOWS_VISTA",
			"WINDOWS_VISTA_X64",
			"WINDOWS_XP",
			"WINDOWS_XP_X64",
			"WINDOWS_2000",
		},
		"additionalAttributes": map[string]string{},
	}

	// 请求微软服务器。
	rspBodyObj, httpCode, rspBody, err := util.HTTPPostJSONToJSON[struct {
		ID uint64 `json:"id"`
	}](ctx, "https://manage.devcenter.microsoft.com/v2.0/my/hardware/products", reqBodyObj, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to create microsoft product", err, httpCode, rspBody)
		return "", errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated {
		log.Error(ctx, "microsoft product create failed", httpCode, rspBody)
		return "", errs.New(consts.ErrSystem)
	}
	if rspBodyObj == nil || rspBodyObj.ID <= 0 {
		log.Error(ctx, "failed to create microsoft product", httpCode, rspBody)
		return "", errs.New(consts.ErrSystem)
	}

	return strconv.FormatUint(rspBodyObj.ID, 10), nil
}

func createAttestationProduct(ctx context.Context, productName string, typ peType) (string, error) {
	// 获取 AccessToken。
	token, err := getMicrosoftAccessToken(ctx)
	if err != nil {
		return "", err
	}

	// 构建请求体。TODO: 补全优化请求参数。
	selectedProductTypes := map[string]string{}
	if typ != PETypeArm64 {
		selectedProductTypes["Windows_v100_RS1"] = "Network Media Device"
		selectedProductTypes["Windows_v100Server_RS1"] = "Network Media Device"
	}
	var requestedSignatures []string
	switch typ {
	case PETypeAmd64:
		requestedSignatures = []string{
			"WINDOWS_v100_X64_TH2_FULL",
			"WINDOWS_v100_X64_RS1_FULL",
			"WINDOWS_v100_X64_RS2_FULL",
			"WINDOWS_v100_X64_RS3_FULL",
			"WINDOWS_v100_X64_RS4_FULL",
			"WINDOWS_v100_X64_RS5_FULL",
			"WINDOWS_v100_X64_19H1_FULL",
			"WINDOWS_v100_X64_VB_FULL",
			"WINDOWS_v100_X64_CO_FULL",
		}
	case PETypeAmd32, PETypeArm32:
		requestedSignatures = []string{
			"WINDOWS_v100_TH2_FULL",
			"WINDOWS_v100_RS1_FULL",
			"WINDOWS_v100_RS2_FULL",
			"WINDOWS_v100_RS3_FULL",
			"WINDOWS_v100_RS4_FULL",
			"WINDOWS_v100_RS5_FULL",
			"WINDOWS_v100_19H1_FULL",
			"WINDOWS_v100_VB_FULL",
			"WINDOWS_v100_X64_CO_FULL",
			"WINDOWS_v100_X64_NI_FULL",
		}
	case PETypeArm64:
		requestedSignatures = []string{
			"WINDOWS_v100_ARM64_RS3_FULL",
			"WINDOWS_v100_ARM64_RS4_FULL",
			"WINDOWS_v100_ARM64_19H1_FULL",
			"WINDOWS_v100_ARM64_VB_FULL",
			"WINDOWS_v100_ARM64_NI_FULL",
			"WINDOWS_v100_ARM64_CO_FULL",
		}
	}
	reqBodyObj := map[string]any{
		"productName":            productName,
		"testHarness":            "Attestation",
		"announcementDate":       time.Now().UTC().Format(time.RFC3339),
		"DeviceMetadataCategory": "Network.NIC.Ethernet",
		"firmwareVersion":        "3.2.0.1022",
		"deviceType":             "external",
		"isTestSign":             false,
		"marketingNames":         []string{"Contoso Rapid Transfer Card Enterprise"},
		"selectedProductTypes":   selectedProductTypes,
		"requestedSignatures":    requestedSignatures,
	}

	// 请求微软服务器。
	rspBodyObj, httpCode, rspBody, err := util.HTTPPostJSONToJSON[struct {
		ID uint64 `json:"id"`
	}](ctx, "https://manage.devcenter.microsoft.com/v2.0/my/hardware/products", reqBodyObj, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to create microsoft product", err, httpCode, rspBody)
		return "", errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated {
		log.Error(ctx, "microsoft product create failed", httpCode, rspBody)
		return "", errs.New(consts.ErrSystem)
	}
	if rspBodyObj == nil || rspBodyObj.ID <= 0 {
		log.Error(ctx, "failed to create microsoft product", httpCode, rspBody)
		return "", errs.New(consts.ErrSystem)
	}

	return strconv.FormatUint(rspBodyObj.ID, 10), nil
}

func createMicrosoftSubmission(ctx context.Context, productID, submissionName string) (string, string, error) {
	// 获取 AccessToken。
	token, err := getMicrosoftAccessToken(ctx)
	if err != nil {
		return "", "", err
	}

	// 构建请求参数。
	reqBodyObj := map[string]string{
		"name": submissionName,
		"type": "initial",
	}

	// 请求微软服务器。
	rspBodyObj, httpCode, rspBody, err := util.HTTPPostJSONToJSON[struct {
		ID       uint64 `json:"id"`
		Download struct {
			Items []struct {
				URL string `json:"url"`
			} `json:"items"`
		} `json:"downloads"`
	}](ctx, fmt.Sprintf("https://manage.devcenter.microsoft.com/v2.0/my/hardware/products/%s/submissions", productID),
		reqBodyObj, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to create microsoft submission", err, httpCode, rspBody)
		return "", "", errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated {
		log.Error(ctx, "failed to create microsoft submission", httpCode, rspBody)
		return "", "", errs.New(consts.ErrSystem)
	}
	if rspBodyObj == nil || rspBodyObj.ID <= 0 || len(rspBodyObj.Download.Items) <= 0 ||
		len(rspBodyObj.Download.Items[0].URL) <= 0 {
		log.Error(ctx, "failed to create microsoft submission", httpCode, rspBody)
		return "", "", errs.New(consts.ErrSystem)
	}

	submissionID := strconv.FormatUint(rspBodyObj.ID, 10)
	uploadURL := rspBodyObj.Download.Items[0].URL
	return submissionID, uploadURL, nil
}

func uploadFileToMicrosoft(ctx context.Context, filePath, uploadURL string) error {
	// 读取文件。
	fileObj, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	fileInfo, err := fileObj.Stat()
	if err != nil {
		log.Error(ctx, "failed to state file", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}

	// 创建请求体。
	request, err := http.NewRequest(http.MethodPut, uploadURL, nil)
	if err != nil {
		log.Error(ctx, "failed to create http request", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	request.Body = fileObj
	request.ContentLength = fileInfo.Size()
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("x-ms-blob-type", "BlockBlob")

	// 上传文件。
	response, err := util.GetHTTPClient().Do(request)
	if err != nil {
		log.Error(ctx, "failed to upload file to microsoft cloud", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	defer util.CloseHTTPBody(ctx, response)
	if response.StatusCode != http.StatusCreated {
		log.Error(ctx, "failed to upload file to microsoft cloud", response.StatusCode)
		return errs.New(consts.ErrSystem)
	}

	return nil
}

func commitMicrosoftSubmission(ctx context.Context, productID, submissionID string) error {
	// 获取 AccessToken。
	token, err := getMicrosoftAccessToken(ctx)
	if err != nil {
		return err
	}

	// 请求微软服务器。
	httpCode, rspBody, err := util.HTTPPost(ctx,
		fmt.Sprintf("https://manage.devcenter.microsoft.com/v2.0/my/hardware/products/%s/submissions/%s/commit",
			productID, submissionID), nil, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to create microsoft submission", err, httpCode, rspBody)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusAccepted {
		log.Error(ctx, "failed to create microsoft submission", httpCode, rspBody)
		return errs.New(consts.ErrSystem)
	}

	return nil
}

func queryMicrosoftSubmission(ctx context.Context, productID, submissionID string) (bool, string, error) {
	// 获取 AccessToken。
	token, err := getMicrosoftAccessToken(ctx)
	if err != nil {
		return false, "", err
	}

	// 请求微软服务器。
	rspBodyObj, httpCode, rspBody, err := util.HTTPGetToJSON[struct {
		CommitStatus   string `json:"commitStatus"`
		WorkflowStatus struct {
			State       string `json:"state"`
			CurrentStep string `json:"currentStep"`
			ErrorReport string `json:"errorReport"`
		} `json:"workflowStatus"`
		Download struct {
			Items []struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"items"`
		} `json:"downloads"`
	}](ctx, fmt.Sprintf("https://manage.devcenter.microsoft.com/v2.0/my/hardware/products/%s/submissions/%s",
		productID, submissionID), "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to query microsoft submission state", err, httpCode, rspBody)
		return false, "", errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusOK || rspBodyObj == nil || len(rspBodyObj.CommitStatus) <= 0 {
		log.Error(ctx, "failed to query microsoft submission state", httpCode, rspBody)
		return false, "", errs.New(consts.ErrSystem)
	}

	// 处理结果。
	if rspBodyObj.WorkflowStatus.State == "failed" {
		_, errorMessage, _ := util.HTTPGet(ctx, rspBodyObj.WorkflowStatus.ErrorReport)
		errorMessage = bytes.ReplaceAll(errorMessage, []byte("\r\n"), []byte(`\r\n`))
		errorMessage = bytes.ReplaceAll(errorMessage, []byte("\n"), []byte(`\n`))
		log.Warn(ctx, "microsoft submission failed", rspBody, errorMessage)
		return true, "", errs.NewWithMsg(consts.ErrSystem, string(errorMessage))
	}
	if rspBodyObj.WorkflowStatus.State == "completed" && rspBodyObj.WorkflowStatus.CurrentStep == "finalizeIngestion" {
		for _, v := range rspBodyObj.Download.Items {
			if v.Type == "signedPackage" {
				return true, v.URL, nil
			}
		}
		log.Error(ctx, "microsoft result file url not found", rspBody)
		return true, "", errs.New(consts.ErrSystem)
	}

	log.Info(ctx, "microsoft result file information", rspBody)
	return false, "", nil
}
