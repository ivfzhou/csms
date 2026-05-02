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
	"time"

	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/validator"
)

const (
	HTTPPathProfileApplyInHouse  = "/internal/profile/applyInHouse"
	HTTPPathProfileRemoveInHouse = "/internal/profile/removeInHouse"
)

// ApplyInHouseProfileReq 申请企业内测描述文件请求体。
type ApplyInHouseProfileReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64"`
}

// ApplyInHouseProfileRsp 申请企业内测描述文件响应体。
type ApplyInHouseProfileRsp struct {
	// 描述文件 Base64 编码内容
	Profile string `json:"profile,omitempty"`
	// 描述文件 ID
	ID string `json:"id,omitempty"`
	// 关联的证书 ID
	CertificateID string `json:"certificateId,omitempty"`
	// 描述文件名
	FileName string `json:"fileName,omitempty"`
	// 描述文件过期时间
	ExpiredTime time.Time `json:"expireTime"`
	// 状态
	Status string `json:"status,omitempty"`
	// 平台
	Platform string `json:"platform,omitempty"`
	// uuid
	UUID string `json:"uuid,omitempty"`
	// 类型
	Type string `json:"type,omitempty"`
	// 团队 ID
	TeamID string `json:"teamId,omitempty"`
}

// RemoveInHouseProfileReq 删除企业内测描述文件请求体。
type RemoveInHouseProfileReq struct {
	// 描述文件 ID
	ID string `form:"id" binding:"len=10"`
}

func initProfile() {
	validator.AddTranslationMessage(reflect.TypeFor[ApplyInHouseProfileReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[RemoveInHouseProfileReq](), validator.TranslationMessage{
		"ID": {
			"len": {
				i18n.LanguageEnglish: "profile file id error",
			},
		},
	})
}
