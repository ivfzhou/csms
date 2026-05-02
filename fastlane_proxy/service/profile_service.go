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
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/fastlane_proxy/consts"
	"gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
)

// ProfileApplyInHouse 申请企业内测描述文件。
func ProfileApplyInHouse(ctx context.Context, req *protocol.ApplyInHouseProfileReq) (
	rsp *protocol.ApplyInHouseProfileRsp, err error) {

	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			uuidStr := uuid.NewString()
			return &protocol.ApplyInHouseProfileRsp{
				Profile: base64.StdEncoding.EncodeToString([]byte(`~<?xml version="1.0" encoding="UTF-8"?>
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
	<string>` + uuidStr + `</string>
	<key>Version</key>
	<integer>1</integer>
</dict>
</plist>
~`)),
				ID:            util.FastRandomAlphaNumberString(10),
				CertificateID: util.FastRandomAlphaNumberString(10),
				FileName:      "",
				ExpiredTime:   time.Now().AddDate(100, 0, 0),
				Status:        "",
				Platform:      "",
				UUID:          uuidStr,
				Type:          "",
				TeamID:        cfg.Get().Apple().InHouseTeamID(),
			}, nil
		}
	}

	// 运行 fastlane 生成描述文件。
	var profileFilePath string
	var profileName string
	{
		log.Info(ctx, "run fastlane to apply profile")
		profileName = fmt.Sprintf(
			"%s_%s_%s_%s.mobileprovision",
			req.BundleID,
			cfg.Get().Apple().InHouseTeamID(),
			time.Now().Format("20060102150405"),
			util.FastRandomAlphaNumberString(4),
		)
		var output string
		output, err = executeFastlaneCommand(
			`run`,
			`get_provisioning_profile`,
			`force:true`,
			`filename:`+profileName,
			`provisioning_name:`+profileName,
			`skip_install:true`,
			`app_identifier:`+req.BundleID,
			`team_id:`+cfg.Get().Apple().InHouseTeamID(),
			`output_path:`+os.TempDir(),
			`username:`+cfg.Get().Apple().AccountName(),
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to apply profile", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		profileFilePath = filepath.Join(os.TempDir(), profileName)
		defer util.RemoveFile(ctx, profileFilePath)
	}

	// 读取描述文件。
	var fileData []byte
	{
		log.Info(ctx, "read profile file data")
		fileData, err = os.ReadFile(profileFilePath)
		if err != nil {
			log.Error(ctx, "failed to read profile file data", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 获取描述文件信息。
	var fastlaneResult struct {
		ID            string `json:"id"`
		Uuid          string `json:"uuid"`
		Expires       string `json:"expires"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		Type          string `json:"type"`
		Platform      string `json:"platform"`
		CertificateID string `json:"certificateId"`
		TeamID        string `json:"teamId"`
		expiredTime   time.Time
	}
	{
		log.Info(ctx, "run fastlane to get profile information")
		var output string
		output, err = executeFastlaneCommand(
			`get_profile_info`,
			`bundle_id:"`+req.BundleID,
			`team_id:`+cfg.Get().Apple().InHouseTeamID(),
			`filename:`+profileName,
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to get profile information", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		profileInfoStr := extractOutput("get_profile_info", output)
		if len(profileInfoStr) <= 0 {
			log.Error(ctx, "failed to to get profile information", output)
			return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("cannot found profile information"))
		}
		if err = json.Unmarshal([]byte(profileInfoStr), &fastlaneResult); err != nil {
			log.Error(ctx, "failed to unmarshal profile information", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		fastlaneResult.expiredTime, err = time.Parse("2006-01-02T15:04:05Z07:00", fastlaneResult.Expires)
		if err != nil {
			log.Error(ctx, "failed to parse time", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return &protocol.ApplyInHouseProfileRsp{
		Profile:       base64.StdEncoding.EncodeToString(fileData),
		ID:            fastlaneResult.ID,
		CertificateID: fastlaneResult.CertificateID,
		FileName:      fastlaneResult.Name,
		ExpiredTime:   fastlaneResult.expiredTime,
		Status:        fastlaneResult.Status,
		Platform:      fastlaneResult.Platform,
		UUID:          fastlaneResult.Uuid,
		Type:          fastlaneResult.Type,
	}, nil
}

// ProfileRemoveInHouse 删除企业内测描述文件。
func ProfileRemoveInHouse(ctx context.Context, req *protocol.RemoveInHouseProfileReq) (err error) {
	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			return nil
		}
	}

	// 运行 fastlane 命令删除描述文件。
	{
		log.Info(ctx, "run fastlane to delete profile")
		var output string
		output, err = executeFastlaneCommand(
			`del_profile`,
			`id:`+req.ID,
			`team_id:`+cfg.Get().Apple().InHouseTeamID(),
		)
		log.Debug(ctx, "output of running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to delete profile", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return nil
}
