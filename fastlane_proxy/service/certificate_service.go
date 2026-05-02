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
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"software.sslmate.com/src/go-pkcs12"

	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/fastlane_proxy/consts"
	"gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
)

// CertificateApplyPush 申请 Push 证书。
func CertificateApplyPush(ctx context.Context, req *protocol.ApplyPushCertificateReq) (
	rsp *protocol.ApplyPushCertificateRsp, err error) {

	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			return &protocol.ApplyPushCertificateRsp{
				Certificate: "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",
				ID:          util.FastRandomAlphaNumberString(10),
			}, nil
		}
	}

	// 运行 fastlane Push 申请证书。
	var fileName string
	var teamID string
	{
		log.Info(ctx, "run fastlane to apply apple push certificate")
		teamID = cfg.Get().Apple().AppStoreTeamID()
		if req.Type == model.AppleBundleIDTypeInHouse {
			teamID = cfg.Get().Apple().InHouseTeamID()
		}
		dev := "true"
		if req.Environment == model.AppleCertificateEnvironmentProduction {
			dev = "false"
		}
		fileName = "production_" + req.BundleID + "_" + time.Now().Format("20060102150405")
		if req.Environment == model.AppleCertificateEnvironmentProduction {
			fileName = "development_" + req.BundleID + "_" + time.Now().Format("20060102150405")
		}
		var output string
		output, err = executeFastlaneCommand(
			`run`,
			`get_push_certificate`,
			`p12_password:`+req.Password,
			`app_identifier:`+req.BundleID,
			`team_id:`+teamID,
			`output_path:`+os.TempDir(),
			`development:`+dev,
			`force:true`,
			`pem_name:`+fileName+".pem",
			`username:`+cfg.Get().Apple().AccountName(),
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to apply apple push certificate", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		defer func() {
			util.RemoveFile(ctx, filepath.Join(os.TempDir(), fileName+".p12"))
			util.RemoveFile(ctx, filepath.Join(os.TempDir(), fileName+".pem"))
			util.RemoveFile(ctx, filepath.Join(os.TempDir(), fileName+".pkey"))
		}()
	}

	// 读取证书。
	var certificateData []byte
	{
		log.Info(ctx, "read certificate file data")
		certificateData, err = os.ReadFile(filepath.Join(os.TempDir(), fileName+".p12"))
		if err != nil {
			log.Error(ctx, "failed to read certificate file data", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 解析证书。
	var expire string
	{
		log.Info(ctx, "parse certificate")
		var certificateInfo *x509.Certificate
		_, certificateInfo, err = pkcs12.Decode(certificateData, req.Password)
		if err != nil {
			log.Error(ctx, "failed to parse certificate", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		expire = certificateInfo.NotAfter.Format("20060102150405")
	}

	// 运行 fastlane 命令，获取证书 ID。
	var fastlaneResult []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		Created       string `json:"created"`
		Expires       string `json:"expires"`
		OwnerID       string `json:"ownerId"`
		TypeDisplayID string `json:"typeDisplayId"`
		CanDownload   string `json:"canDownload"`
		OwnerType     string `json:"ownerType"`
		OwnerName     string `json:"ownerName"`
	}
	{
		nameLike := "\"Apple Sandbox Push\""
		if req.Environment == model.AppleCertificateEnvironmentProduction {
			nameLike = "\"Apple Push\""
		}
		log.Info(ctx, "run fastlane to get apple push certificate id")
		var output string
		output, err = executeFastlaneCommand(
			`get_cert_info`,
			`bundle_id:`+req.BundleID,
			`team_id:`+teamID,
			`name_like:`+nameLike,
			`expire:`+expire,
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to get apple push certificate id", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		certificateInfoStr := extractOutput("get_cert_info", output)
		if len(certificateInfoStr) <= 0 {
			log.Error(ctx, "failed to to get certificate information", output)
			return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("cannot found certificate information"))
		}

		if err = json.Unmarshal([]byte(certificateInfoStr), &fastlaneResult); err != nil {
			log.Error(ctx, "failed to unmarshal profile information", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		if len(fastlaneResult) <= 0 {
			log.Error(ctx, "no certificate id found")
			return nil, errs.New(consts.ErrSystem)
		}
		if len(fastlaneResult) > 1 {
			log.Warn(ctx, "multiple certificate id found", fastlaneResult)
		}
	}

	return &protocol.ApplyPushCertificateRsp{
		Certificate: base64.StdEncoding.EncodeToString(certificateData),
		ID:          fastlaneResult[0].ID,
	}, nil
}

// CertificateRemovePush 删除 Push 证书。
func CertificateRemovePush(ctx context.Context, req *protocol.RemovePushCertificateReq) (err error) {
	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			return nil
		}
	}

	// 运行 fastlane 删除 Push 证书。
	{
		log.Info(ctx, "run fastlane to remove apple push certificate")
		teamID := cfg.Get().Apple().AppStoreTeamID()
		if req.Type == model.AppleBundleIDTypeInHouse {
			teamID = cfg.Get().Apple().InHouseTeamID()
		}
		nameLike := "\"Apple Sandbox Push\""
		if req.Environment == model.AppleCertificateEnvironmentProduction {
			nameLike = "\"Apple Push\""
		}
		var output string
		output, err = executeFastlaneCommand(
			`del_cert`,
			`bundle_id:`+req.BundleID,
			`team_id:`+teamID,
			`id:`+req.CertificateID,
			`name_like:`+nameLike,
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to remove apple push certificate", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return nil
}
