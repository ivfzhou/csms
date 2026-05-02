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
	"fmt"
	"net/http"
	"path"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	fp "gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
)

func httpFastlaneApplyBundleID(ctx context.Context, bundleID string) (*fp.ApplyInHouseBundleIDRsp, error) {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathBundleApplyInHouse))
	result, status, rspBody, err := util.HTTPPostJSONToJSON[util.Response[fp.ApplyInHouseBundleIDRsp]](ctx, reqURL,
		&fp.ApplyInHouseBundleIDReq{BundleID: bundleID})
	if err != nil {
		log.Error(ctx, "failed to request fastlane to register apple bundle id", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || result == nil {
		if result != nil {
			return nil, errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return nil, errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return result.Data, nil
}

func httpFastlaneModifyBundleIDCapabilities(ctx context.Context, bundleID string, m map[string]bool) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathBundleModifyInHouseCapabilities))
	result, status, rspBody, err := util.HTTPPostJSONToJSON[util.Response[any]](ctx, reqURL,
		&fp.ModifyInHouseBundleIDCapabilitiesReq{BundleID: bundleID, Service: m})
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to modify apple bundle capabilities", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		if result != nil {
			return errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return nil
}

func httpFastlaneDeleteInHouseBundleID(ctx context.Context, bundleID string) error {
	query := util.EncodeStructToURLQuery(&fp.RemoveInHouseBundleIDReq{BundleID: bundleID})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathBundleRemoveInHouse), query)
	result, status, rspBody, err := util.HTTPDeleteToJSON[util.Response[any]](ctx, reqURL)
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to delete apple bundle id", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		if result != nil {
			return errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return nil
}

func httpFastlaneApplyInHouseProfile(ctx context.Context, bundleID string) (*fp.ApplyInHouseProfileRsp, error) {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathProfileApplyInHouse))
	result, status, rspBody, err := util.HTTPPostJSONToJSON[util.Response[fp.ApplyInHouseProfileRsp]](ctx, reqURL,
		&fp.ApplyInHouseProfileReq{BundleID: bundleID})
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to apply apple profile", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || result == nil {
		if result != nil {
			return nil, errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return nil, errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return result.Data, nil
}

func httpFastlaneApplyPushCert(ctx context.Context, req *fp.ApplyPushCertificateReq) (
	*fp.ApplyPushCertificateRsp, error) {

	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathCertificateApplyPush))
	result, status, rspBody, err := util.HTTPPostJSONToJSON[util.Response[fp.ApplyPushCertificateRsp]](ctx, reqURL, req)
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to apply certificate", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || result == nil {
		if result != nil {
			return nil, errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return nil, errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return result.Data, nil
}

func httpFastlaneRemoveInHouseProfile(ctx context.Context, id string) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathProfileRemoveInHouse))
	result, status, rspBody, err := util.HTTPPostJSONToJSON[util.Response[any]](ctx, reqURL,
		&fp.RemoveInHouseProfileReq{ID: id})
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to delete apple profile", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || result == nil {
		if result != nil {
			return errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return nil
}

func httpFastlaneRemovePushCertificate(ctx context.Context, id, bundleID string, environment, typ int) error {
	query := util.EncodeStructToURLQuery(&fp.RemovePushCertificateReq{
		CertificateID: id,
		BundleID:      bundleID,
		Environment:   environment,
		Type:          typ,
	})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameFastlane, fp.HTTPPathCertificateRemovePush), query)
	result, status, rspBody, err := util.HTTPDeleteToJSON[util.Response[any]](ctx, reqURL)
	if err != nil {
		log.Error(ctx, "failed to http request fastlane to delete apple profile", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		if result != nil {
			return errs.NewWithError(consts.ErrSystem,
				fmt.Errorf("failed to request fastlane: %d %d %s", status, result.Code, result.Message))
		}
		return errs.NewWithError(consts.ErrSystem,
			fmt.Errorf("failed to request fastlane: %d %s", status, rspBody))
	}
	return nil
}
