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

func httpBackendGetMachineEVCertificates(ctx context.Context, ip string) ([]string, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetMachineEVCertificatesReq{IP: ip})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetMachineEVCertificates), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[[]string]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get machine ev certificates %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get machine ev certificates %d %s %v", status, rspBody, rsp)
	}

	if rsp.Data == nil {
		return nil, nil
	}

	return *rsp.Data, nil
}

func httpBackendGetWindowsSigningJob(ctx context.Context, jobID string) (*model.WindowsSigningJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetWindowsSigningJobReq{JobID: jobID})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetWindowsSigningJob), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.WindowsSigningJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get windows signing job %w", err)
	}
	if util.In(status, http.StatusBadRequest, http.StatusNotFound) {
		return nil, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get windows signing job %d %s %v", status, rspBody, rsp)
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

func httpBackendGetCertificate(ctx context.Context, id int) (*model.WindowsCertificate, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetCertificateReq{ID: id})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalGetCertificate), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.WindowsCertificate]](ctx, reqURL)
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get windows certificate %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get windows certificate %d %s %v", status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendUpdateWindowsSigningJob(ctx context.Context, req *bp.WindowsInternalUpdateSigningJobReq) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathWindowsInternalUpdateSigningJob))
	status, rspBody, err := util.HTTPPostJSON(ctx, reqURL, req)
	if err != nil {
		return fmt.Errorf("failed to http backend to update windows signing job %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("failed to http backend to update windows signing job %d %s", status, rspBody)
	}
	return nil
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

func httpBackendGetAndroidSigningJob(ctx context.Context, jobID string) (*model.AndroidSigningJob, error) {
	query := util.EncodeStructToURLQuery(&bp.AndroidInternalGetSigningJobReq{JobID: jobID})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidInternalGetSigningJob), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.AndroidSigningJob]](ctx, reqURL)
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get android signing job %w", err)
	}
	if util.In(status, http.StatusBadRequest, http.StatusNotFound) {
		return nil, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get android signing job %d %s %v", status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendGetAndroidCertificate(ctx context.Context, id int) (*model.AndroidCertificate, error) {
	query := util.EncodeStructToURLQuery(&bp.AndroidInternalGetCertificateReq{ID: id})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidInternalGetCertificate), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.AndroidCertificate]](ctx, reqURL)
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get android certificate %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get android certificate %d %s %v", status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendUpdateAndroidSigningJob(ctx context.Context, req *bp.AndroidInternalUpdateSigningJobReq) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAndroidInternalUpdateSigningJob))
	status, rspBody, err := util.HTTPPostJSON(ctx, reqURL, req)
	if err != nil {
		return fmt.Errorf("failed to http backend to update android signing job %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("failed to http backend to update android signing job %d %s", status, rspBody)
	}
	return nil
}

func httpBackendGetAppleSigningJob(ctx context.Context, jobID string) (*model.AppleSigningJob, error) {
	query := util.EncodeStructToURLQuery(&bp.WindowsInternalGetWindowsSigningJobReq{JobID: jobID})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleInternalGetSigningJob), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[model.AppleSigningJob]](ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get apple signing job %w", err)
	}
	if util.In(status, http.StatusBadRequest, http.StatusNotFound) {
		return nil, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get apple signing job %d %s %v", status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendGetAppleCertificateAndProfile(ctx context.Context, profileID int) (
	*bp.AppleInternalGetCertificateAndProfileRsp, error) {

	query := util.EncodeStructToURLQuery(&bp.AppleInternalGetCertificateAndProfileReq{ProfileID: profileID})
	reqURL := fmt.Sprintf("http://%s:%d/%s?%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleInternalGetCertificateAndProfile), query)
	rsp, status, rspBody, err := util.HTTPGetToJSON[util.Response[bp.AppleInternalGetCertificateAndProfileRsp]](
		ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to http backend to get apple certificate and profile %w", err)
	}
	if util.In(status, http.StatusBadRequest, http.StatusNotFound) {
		return nil, nil
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices || rsp == nil {
		return nil, fmt.Errorf("failed to http backend to get apple certificate and profile %d %s %v",
			status, rspBody, rsp)
	}

	return rsp.Data, nil
}

func httpBackendUpdateAppleSigningJob(ctx context.Context, req *bp.AppleInternalUpdateSigningJobReq) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathAppleInternalUpdateSigningJob))
	status, rspBody, err := util.HTTPPostJSON(ctx, reqURL, req)
	if err != nil {
		return fmt.Errorf("failed to http backend to update apple signing job %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("failed to http backend to update apple signing job %d %s", status, rspBody)
	}
	return nil
}
