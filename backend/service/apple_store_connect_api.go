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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

type appleAPIResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			//公共
			Name       string `json:"name"`
			Identifier string `json:"identifier"`
			Platform   string `json:"platform"`
			SeedID     string `json:"seedId"`

			// devices
			AddedDate   string `json:"addedDate"`
			DeviceClass string `json:"deviceClass"`
			Model       string `json:"model"`
			UdID        string `json:"udid"`
			Status      string `json:"status"`

			// capability
			CapabilityType string `json:"capabilityType"`

			// certificate
			CertificateContent string `json:"certificateContent"`
			DisplayName        string `json:"displayName"`
			ExpirationDate     string `json:"expirationDate"`
			SerialNumber       string `json:"serialNumber"`
			CertificateType    string `json:"certificateType"`

			// profile
			ProfileContent string `json:"profileContent"`
			ProfileState   string `json:"profileState"`
			ProfileType    string `json:"profileType"`
			Uuid           string `json:"uuid"`
		} `json:"attributes"`
		Relationships struct {
			BundleIDCapabilities struct {
				Meta struct {
					Paging struct {
						Total int   `json:"total"`
						Limit int64 `json:"limit"`
					} `json:"paging"`
				} `json:"meta"`
				Links struct {
					Self    string `json:"self"`
					Related string `json:"related"`
				} `json:"links"`
			} `json:"bundleIdCapabilities"`
			Profiles struct {
				Meta struct {
					Paging struct {
						Total int   `json:"total"`
						Limit int64 `json:"limit"`
					} `json:"paging"`
				} `json:"meta"`
				Links struct {
					Self    string `json:"self"`
					Related string `json:"related"`
				} `json:"links"`
			} `json:"profiles"`
		} `json:"relationships"`
	} `json:"data"`
	Errors []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func httpAppleAPIApplyCertificate(ctx context.Context, token, typ string) (*appleAPIResponse, error) {
	if util.IsLocalEnvironment() && cc.TestMode() {
		apiResult := &appleAPIResponse{}
		apiResult.Data.Attributes.CertificateContent = `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`
		apiResult.Data.ID = util.FastRandomAlphaNumberString(10)
		return apiResult, nil
	}

	reqURL := consts.AppleServerDomain + "/v1/certificates"
	req := map[string]any{
		"data": map[string]any{
			"type": "certificates",
			"attributes": map[string]any{
				"certificateType": typ,
				"csrContent":      cfg.Get().Apple().ApplyCertificateCSR(),
			},
		},
	}
	result, rspCode, _, err := util.HTTPPostJSONToJSON[appleAPIResponse](ctx, reqURL, req, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http request apple to apply certificate", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if rspCode != http.StatusCreated && len(result.Errors) > 0 {
		errMsg := result.Errors[0]
		if errMsg.Status == strconv.Itoa(http.StatusConflict) &&
			strings.Contains(errMsg.Detail, "You already have a current") &&
			strings.Contains(errMsg.Detail, "certificate or a pending certificate request") {
			return nil, errs.New(consts.ErrAppleCertificateExists)
		}
		return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return result, nil
}

func httpAppleAPIDeleteBundleID(ctx context.Context, token, bundleID string) error {
	if util.IsLocalEnvironment() && cc.TestMode() {
		return nil
	}

	reqURL := consts.AppleServerDomain + "/v1/bundleIDs/" + bundleID
	rspCode, rspBody, err := util.HTTPDelete(ctx, reqURL, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http request apple to delete apple bundle id", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if rspCode != http.StatusNoContent {
		var httpResult appleAPIResponse
		if err = json.Unmarshal(rspBody, &httpResult); err != nil {
			log.Error(ctx, "failed to unmarshal http reponse body", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		errMsg := httpResult.Errors[0]
		if errMsg.Status == strconv.Itoa(http.StatusNotFound) &&
			strings.Contains(errMsg.Detail, fmt.Sprintf("There is no App ID with ID '%s' on this team", bundleID)) {
			log.Warn(ctx, "deleting non exists apple bundle id", bundleID)
			return nil
		}
		log.Error(ctx, "apple api error", rspBody)
		return errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return nil
}

func httpAppleAPIApplyBundleID(ctx context.Context, token, bundleID string) (*appleAPIResponse, error) {
	if util.IsLocalEnvironment() && cc.TestMode() {
		apiResult := &appleAPIResponse{}
		apiResult.Data.ID = util.FastRandomAlphaNumberString(10)
		apiResult.Data.Attributes.Platform = model.ApplePlatformUniversalDescription
		return apiResult, nil
	}

	arr := strings.Split(bundleID, ".")
	name := arr[len(arr)-1]
	reqURL := consts.AppleServerDomain + "/v1/bundleIds"
	req := map[string]any{
		"data": map[string]any{
			"attributes": map[string]any{
				"name":       name,
				"identifier": bundleID,
			},
			"type": "bundleIds",
		},
	}
	httpResult, httpCode, _, err := util.HTTPPostJSONToJSON[appleAPIResponse](ctx, reqURL, req, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to register apple bundle id from apple", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated && len(httpResult.Errors) > 0 {
		errMsg := httpResult.Errors[0]
		if errMsg.Status == strconv.Itoa(http.StatusConflict) &&
			strings.Contains(errMsg.Detail, fmt.Sprintf("An App ID with Identifier '%s' is not available", bundleID)) {
			return nil, errs.New(consts.ErrAppleBundleIDExist)
		}
		log.Error(ctx, "apple api error", httpResult)
		return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return httpResult, nil
}

func httpAppleAPIRemoveBundleIDCapability(ctx context.Context, token, bundleID, capability string) error {
	if util.IsLocalEnvironment() && cc.TestMode() {
		return nil
	}

	reqURL := fmt.Sprintf("%s/v1/bundleIdCapabilities/%s_%s", consts.AppleServerDomain, bundleID,
		cc.AppleBundleIDCapabilities[capability][0])
	rspCode, rspBody, err := util.HTTPDelete(ctx, reqURL, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http apple server to remove apple bundle capability", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if rspCode != http.StatusNoContent {
		var result appleAPIResponse
		if err = json.Unmarshal(rspBody, &result); err != nil {
			log.Error(ctx, "failed to unmarshal data", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}

		// Bundle ID 没有该能力项。
		errMsg := result.Errors[0]
		if errMsg.Status == "404" {
			return nil
		}

		log.Error(ctx, "apple api error", rspBody)
		return errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v %v %v",
			errMsg.Code, errMsg.Status, errMsg.Title, errMsg.Detail))
	}
	return nil
}

func httpAppleAPIEnableBundleIDCapability(ctx context.Context, token, bundleID, capability string) error {
	if util.IsLocalEnvironment() && cc.TestMode() {
		return nil
	}

	reqURL := consts.AppleServerDomain + "/v1/bundleIdCapabilities"
	req := map[string]any{
		"data": map[string]any{
			"attributes": map[string]any{
				"capabilityType": cc.AppleBundleIDCapabilities[capability][0],
			},
			"relationships": map[string]any{
				"bundleID": map[string]any{
					"data": map[string]any{
						"id":   bundleID,
						"type": "bundleIds",
					},
				},
			},
			"type": "bundleIdCapabilities",
		},
	}
	result, rspCode, _, err := util.HTTPPostJSONToJSON[appleAPIResponse](ctx, reqURL, req, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http request apple to enable apple bundle capability", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if rspCode != http.StatusCreated && len(result.Errors) > 0 {
		errMsg := result.Errors[0]
		log.Error(ctx, "apple api error", result)
		return errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return nil
}

func httpAppleAPIApplyProfile(ctx context.Context, token, typ, bundleID, name, certID string, deviceIDs []string) (
	*appleAPIResponse, error) {

	if util.IsLocalEnvironment() && cc.TestMode() {
		apiResult := &appleAPIResponse{}
		apiResult.Data.ID = util.FastRandomAlphaNumberString(10)
		apiResult.Data.Attributes.ProfileContent = base64.StdEncoding.EncodeToString([]byte(`~<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>~</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>~</string>
	</array>
	<key>CreationDate</key>
	<date>2025-09-03T03:42:31Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
		<string>xrOS</string>
		<string>visionOS</string>
	</array>
	<key>IsXcodeManaged</key>
	<false/>
	<key>DeveloperCertificates</key>
	<array>
		<data>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</data>
	</array>
	<key>DER-Encoded-Profile</key>
	<data>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</data>
	<key>Entitlements</key>
	<dict>
		<key>beta-reports-active</key>
		<true/>
				<key>com.apple.developer.networking.networkextension</key>
		<array>
				<string>app-proxy-provider</string>
				<string>content-filter-provider</string>
				<string>packet-tunnel-provider</string>
				<string>dns-proxy</string>
				<string>dns-settings</string>
				<string>relay</string>
				<string>url-filter-provider</string>
				<string>hotspot-provider</string>
		</array>
				<key>aps-environment</key>
		<string>production</string>
				<key>application-identifier</key>
		<string>~</string>
				<key>keychain-access-groups</key>
		<array>
				<string>~</string>
				<string>com.apple.token</string>
		</array>
				<key>get-task-allow</key>
		<false/>
				<key>com.apple.developer.team-identifier</key>
		<string>~</string>
	</dict>
	<key>ExpirationDate</key>
	<date>2125-09-03T02:43:52Z</date>
	<key>Name</key>
	<string>~</string>
	<key>TeamIdentifier</key>
	<array>
		<string>~</string>
	</array>
	<key>TeamName</key>
	<string>~</string>
	<key>TimeToLive</key>
	<integer>364</integer>
	<key>UUID</key>
	<string>` + uuid.NewString() + `</string>
	<key>Version</key>
	<integer>1</integer>
</dict>
</plist>
~`))
		return apiResult, nil
	}

	devices := make([]map[string]any, len(deviceIDs))
	for i, v := range deviceIDs {
		devices[i] = map[string]any{
			"id":   v,
			"type": "devices",
		}
	}
	req := map[string]any{
		"data": map[string]any{
			"type": "profiles",
			"attributes": map[string]any{
				"profileType": typ,
				"name":        name,
			},
			"relationships": map[string]any{
				"bundleId": map[string]any{
					"data": map[string]any{
						"id":   bundleID,
						"type": "bundleIds",
					},
				},
				"certificates": map[string]any{
					"data": []map[string]any{
						{
							"id":   certID,
							"type": "certificates",
						},
					},
				},
			},
		},
	}
	if len(devices) > 0 {
		req["data"].(map[string]any)["relationships"].(map[string]any)["devices"] = map[string]any{
			"data": devices,
		}
	}
	reqURL := consts.AppleServerDomain + "/v1/profiles"
	httpResult, httpCode, _, err := util.HTTPPostJSONToJSON[appleAPIResponse](ctx, reqURL, req, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http request apple to apply apple profile", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated && len(httpResult.Errors) > 0 {
		errMsg := httpResult.Errors[0]
		if errMsg.Status == strconv.Itoa(http.StatusConflict) &&
			strings.Contains(errMsg.Detail, fmt.Sprintf("Multiple profiles found with the name '%s'", name)) {
			return nil, errs.New(consts.ErrAppleProfileNameExist)
		}
		log.Error(ctx, "apple api error", httpResult)
		return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return httpResult, nil
}

func httpAppleAPIRegisterDevice(ctx context.Context, token, name, udid, platform string) (*appleAPIResponse, error) {
	if util.IsLocalEnvironment() && cc.TestMode() {
		apiResult := &appleAPIResponse{}
		apiResult.Data.ID = util.FastRandomAlphaNumberString(10)
		apiResult.Data.Attributes.Platform = "IOS"
		apiResult.Data.Attributes.Model = "iPhone 15 Pro Max"
		return apiResult, nil
	}

	reqURL := consts.AppleServerDomain + "/v1/devices"
	req := map[string]any{
		"data": map[string]any{
			"attributes": map[string]any{
				"name":     name,
				"udid":     udid,
				"platform": platform,
			},
			"type": "devices",
		},
	}
	httpResult, httpCode, _, err := util.HTTPPostJSONToJSON[appleAPIResponse](ctx, reqURL, req, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http request apple to register device", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if httpCode != http.StatusCreated && len(httpResult.Errors) > 0 {
		errMsg := httpResult.Errors[0]
		if errMsg.Detail == fmt.Sprintf("A device with number '%s' already exists on this team.", udid) {
			return nil, errs.New(consts.ErrAppleDeviceRegistered)
		} else if errMsg.Detail == fmt.Sprintf("An invalid value '%s' was provided for the parameter 'udid'.", udid) {
			return nil, errs.New(consts.ErrInvalidAppleDeviceUDID)
		} else if strings.HasPrefix(errMsg.Detail, "Your development team has reached the maximum number of registered ") &&
			strings.HasSuffix(errMsg.Detail, " devices.") {
			return nil, errs.New(consts.ErrAppleDeviceRegisterReachLimit)
		}
		log.Error(ctx, "apple api error", httpResult)
		return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v", errMsg.Code, errMsg.Detail))
	}
	return httpResult, nil
}

func httpAppleAPIRemoveProfile(ctx context.Context, token, id string) error {
	if util.IsLocalEnvironment() && cc.TestMode() {
		return nil
	}

	reqURL := fmt.Sprintf("%s/v1/profiles/%s", consts.AppleServerDomain, id)
	rspCode, rspBody, err := util.HTTPDelete(ctx, reqURL, "Authorization", token)
	if err != nil {
		log.Error(ctx, "failed to http apple server to remove apple profile", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if rspCode != http.StatusNoContent {
		var result appleAPIResponse
		if err = json.Unmarshal(rspBody, &result); err != nil {
			log.Error(ctx, "failed to unmarshal data", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}

		// Bundle ID 没有该能力项。
		errMsg := result.Errors[0]
		if errMsg.Status == "404" {
			return nil
		}

		log.Error(ctx, "apple api error", rspBody)
		return errs.NewWithError(consts.ErrSystem, fmt.Errorf("%v %v %v %v",
			errMsg.Code, errMsg.Status, errMsg.Title, errMsg.Detail))
	}
	return nil
}
