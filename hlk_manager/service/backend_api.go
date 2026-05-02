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
	"io"
	"net/http"
	"path"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func httpBackendGetWHQLJobInformation(ctx context.Context, id int) (*model.WhqlJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetWHQLJobReq{ID: id})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetWHQLJob), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.WhqlJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get whql job information %w", err)
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get whql job information %d %s %v", status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendGetWHQLJobToInitialTestMachine(ctx context.Context, system string) (*model.WhqlJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetWHQLJobToInitialTestMachineReq{System: system})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetWHQLJobToInitialTestMachine), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.WhqlJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get whql job to initial test machine %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get whql job to initial test machine %d %s %v",
			status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendDownloadFile(ctx context.Context, fileId, fileDirectoryPath string) (
	filePath string, fileSize int64, err error) {

	query := util.EncodeStructToURLQuery(&bp.FileInternalDownloadReq{FileID: fileId})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileInternalDownload), query)
	status, filePath, fileSize, err := util.HTTPGetToDisk2(ctx, reqURL, fileDirectoryPath)
	if err != nil {
		err = fmt.Errorf("failed to http backend to download file %v %w", status, err)
		return
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		err = fmt.Errorf("failed to http backend to download file %d", status)
		return
	}
	return
}

func httpBackendUpdateWHQLJob(ctx context.Context, req *bp.WindowsInternalUpdateWHQLJobReq) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalUpdateWHQLJob))
	status, rspBody, err := util.HTTPPostJSON(ctx, reqURL, req)
	if err != nil {
		return fmt.Errorf("failed to http backend to update whql job %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("failed to http backend to update whql job %d %s", status, rspBody)
	}
	return nil
}

func httpBackendGetWHQLJobToStartTest(ctx context.Context, systems []string) (*model.WhqlJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetWHQLJobToStartTestReq{Systems: systems})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetWHQLJobToStartTest), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.WhqlJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get whql job to start test %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get whql job to start test %d %s %v",
			status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendGetTestingWHQLJobs(ctx context.Context, systems []string) ([]*model.WhqlJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetTestingWHQLJobsReq{Systems: systems})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetTestingWHQLJobs), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[[]*model.WhqlJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get testing whql jobs %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get testing whql jobs %d %s %v",
			status, rspBody, rsp)
	}

	if rsp.Data == nil {
		return nil, nil
	}
	return *rsp.Data, nil
}

func httpBackendUploadFile(ctx context.Context, fileType int, fileName string, appID int, body io.Reader,
	fileSize int64) (string, error) {

	query := util.EncodeStructToURLQuery(&bp.FileInternalUploadReq{
		Type:  fileType,
		Name:  fileName,
		AppID: appID,
	})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileInternalUpload), query)
	rsp, status, rspBody, err := util.HTTPPostToJSON[util.Response[string]](
		ctx, reqURL, body, fileSize)
	if err != nil {
		return "", fmt.Errorf("failed to http backend to upload file %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil || rsp.Data == nil ||
		len(*rsp.Data) <= 0 {
		return "", fmt.Errorf("failed to http backend to upload file %d %s %v",
			status, rspBody, rsp)
	}

	return *rsp.Data, nil
}
