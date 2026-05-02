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

package protocol

import (
	"reflect"

	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/validator"
)

const (
	HTTPPathBundleApplyInHouse              = "/internal/bundle/applyInHouse"
	HTTPPathBundleRemoveInHouse             = "/internal/bundle/removeInHouse"
	HTTPPathBundleModifyInHouseCapabilities = "/internal/bundle/modifyInHouseCapabilities"
)

// ApplyInHouseBundleIDReq 申请企业内测 Bundle ID 请求体。
type ApplyInHouseBundleIDReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64"`
}

// ApplyInHouseBundleIDRsp 申请企业内测 Bundle ID 响应体。
type ApplyInHouseBundleIDRsp struct {
	// ID
	ID string `json:"id,omitempty"`
	// 平台
	Platform string `json:"platform,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 能力项
	Capabilities string `json:"capabilities,omitempty"`
	// Bundle ID
	BundleID string `json:"bundleId,omitempty"`
	// 团队 ID
	TeamID string `json:"teamId,omitempty"`
	// 是否通配符
	IsWildcard bool `json:"isWildcard,omitempty"`
}

// RemoveInHouseBundleIDReq 删除企业内测 Bundle ID 请求体。
type RemoveInHouseBundleIDReq struct {
	// Bundle ID
	BundleID string `form:"bundleId" binding:"required,max=64"`
}

// ModifyInHouseBundleIDCapabilitiesReq 修改企业内测 Bundle ID 能力项请求体。
type ModifyInHouseBundleIDCapabilitiesReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64"`
	// 能力增减信息
	Service map[string]bool `json:"service,omitempty" binding:"min=1,dive,keys,capability,endkeys"`
}

func initBundle() {
	validator.AddTranslationMessage(reflect.TypeFor[ApplyInHouseBundleIDReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[RemoveInHouseBundleIDReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[ModifyInHouseBundleIDCapabilitiesReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
		"Service": {
			"min": {
				i18n.LanguageEnglish: "request parameter lacks capability item information",
			},
			"capability": {
				i18n.LanguageEnglish: "wrong apple bundle id capability item",
			},
		},
	})
}
