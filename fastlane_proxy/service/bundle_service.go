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
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/fastlane_proxy/consts"
	"gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
)

// BundleApplyInHouse 申请企业内测环境 Bundle ID。
func BundleApplyInHouse(ctx context.Context, req *protocol.ApplyInHouseBundleIDReq) (
	rsp *protocol.ApplyInHouseBundleIDRsp, err error) {

	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			platform := model.ApplePlatformIOSDescription
			switch rand.Intn(3) {
			case 0:
				platform = model.ApplePlatformMacOSDescription
			case 1:
				platform = model.ApplePlatformUniversalDescription
			}
			return &protocol.ApplyInHouseBundleIDRsp{
				ID:           util.FastRandomAlphaNumberString(10),
				Platform:     platform,
				Name:         "apple bundle id random name",
				Capabilities: "",
				BundleID:     req.BundleID,
				TeamID:       cfg.Get().Apple().InHouseTeamID(),
				IsWildcard:   false,
			}, nil
		}
	}

	// 运行 fastlane 命令申请 Bundle ID。
	{
		log.Info(ctx, "run fastlane to apply apple bundle id")
		arr := strings.Split(req.BundleID, ".")
		name := arr[len(arr)-1]
		output, err := executeFastlaneCommand(
			"run",
			"create_app_online",
			`skip_itc:true`,
			`app_identifier:`+req.BundleID,
			`app_name:`+name,
			`team_id:`+cfg.Get().Apple().InHouseTeamID(),
		)
		log.Debug(ctx, "output of running fastlane command", output)
		if err != nil {
			// Bundle ID 被占用了。
			if strings.Contains(output, fmt.Sprintf("'%s' already exists", req.BundleID)) {
				log.Warn(ctx, "already applied apple bundle id")
				return nil, errs.New(consts.ErrAppleBundleIDExists)
			}
			log.Error(ctx, "failed to run fastlane command of applying apple bundle id", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 运行 fastlane 命令，获取 Bundle ID 信息。
	var output string
	{
		log.Info(ctx, "run fastlane to get apple bundle id information")
		output, err = executeFastlaneCommand(
			"get_bundle_info",
			"bundle_id:"+req.BundleID,
			"team_id:"+cfg.Get().Apple().InHouseTeamID(),
		)
		log.Debug(ctx, "output of running fastlane command", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane command of getting apple bundle id information", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 解析和序列化 fastlane 命令输出。
	{
		log.Info(ctx, "parse bundle id information")
		bundleIDInfo := extractOutput("get_bundle_info", output)
		if len(bundleIDInfo) <= 0 {
			log.Error(ctx, "failed to to get apple bundle id information", output)
			return nil, errs.NewWithError(consts.ErrSystem, fmt.Errorf("cannot found apple bundle id information"))
		}
		rsp = &protocol.ApplyInHouseBundleIDRsp{}
		if err = json.Unmarshal([]byte(bundleIDInfo), rsp); err != nil {
			log.Error(ctx, "failed to unmarshal apple bundle id information", err, bundleIDInfo)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return
}

// BundleModifyInHouseCapabilities 修改企业内测 Bundle ID 能力项。
func BundleModifyInHouseCapabilities(ctx context.Context, req *protocol.ModifyInHouseBundleIDCapabilitiesReq) (
	err error) {

	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			return
		}
	}

	// 整理能力项命令参数。
	var capabilities string
	{
		capabilitiesMap := make(map[string]bool, len(req.Service))
		for k, v := range req.Service {
			str := cc.AppleBundleIDCapabilities[k][1]
			if len(str) > 0 {
				capabilitiesMap[str] = v
			}
		}
		capabilityBytes, _ := json.Marshal(capabilitiesMap)
		capabilities = string(capabilityBytes)
	}

	// 运行 fastlane 命令修改 Bundle ID 能力项。
	{
		log.Info(ctx, "run fastlane to modify apple bundle id capabilities")
		var output string
		output, err = executeFastlaneCommand(
			"run",
			"modify_services",
			"app_identifier:"+req.BundleID,
			"team_id:"+cfg.Get().Apple().InHouseTeamID(),
			fmt.Sprintf("services:'%s'", capabilities),
		)
		log.Debug(ctx, "output fo running fastlane", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to modify apple bundle id capabilities", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return
}

// BundleRemoveInHouse 删除企业内测 Bundle ID。
func BundleRemoveInHouse(ctx context.Context, req *protocol.RemoveInHouseBundleIDReq) (err error) {
	// 检测本地测试模式。
	{
		if util.IsLocalEnvironment() && cc.TestMode() {
			return nil
		}
	}

	// 运行 fastlane 命令删除 Bundle ID。
	{
		log.Info(ctx, "run fastlane to remove apple bundle id")
		var output string
		output, err = executeFastlaneCommand(
			`del_bundle`,
			`bundle_id:`+req.BundleID,
			`team_id:"`+cfg.Get().Apple().InHouseTeamID(),
		)
		log.Debug(ctx, "output of running fastlane command", output)
		if err != nil {
			log.Error(ctx, "failed to run fastlane to remove apple bundle id", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}

	return nil
}
