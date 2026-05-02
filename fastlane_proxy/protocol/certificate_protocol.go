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
	HTTPPathCertificateApplyPush  = "/internal/certificate/applyPush"
	HTTPPathCertificateRemovePush = "/internal/certificate/removePush"
)

// ApplyPushCertificateReq 申请 Push 证书请求体。
type ApplyPushCertificateReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64"`
	// 环境
	Environment int `json:"environment,omitempty" binding:"min=1,max=2"`
	// Bundle ID 类型
	Type int `json:"type,omitempty" binding:"min=1,max=2"`
	// 密码
	Password string `json:"password,omitempty" binding:"min=6,max=16"`
}

// ApplyPushCertificateRsp 申请 Push 证书响应体。
type ApplyPushCertificateRsp struct {
	// Base64 编码的证书内容
	Certificate string `json:"certificate,omitempty"`
	// 证书 ID
	ID string `json:"id,omitempty"`
}

// RemovePushCertificateReq 删除 Push 证书请求体。
type RemovePushCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=10"`
	// Bundle ID
	BundleID string `form:"bundleId" binding:"required,max=64"`
	// 环境
	Environment int `form:"environment" binding:"min=1,max=2"`
	// Bundle ID 类型
	Type int `form:"type" binding:"min=1,max=2"`
}

func initCertificate() {
	validator.AddTranslationMessage(reflect.TypeFor[ApplyPushCertificateReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
		"Environment": {
			"min": {
				i18n.LanguageEnglish: "wrong environment type",
			},
			"max": {
				i18n.LanguageEnglish: "wrong environment type",
			},
		},
		"Type": {
			"min": {
				i18n.LanguageEnglish: "error in type",
			},
			"max": {
				i18n.LanguageEnglish: "error in type",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageEnglish: "certificate password length is too short",
			},
			"max": {
				i18n.LanguageEnglish: "certificate password length is too long",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[RemovePushCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
		"BundleID": {
			"required": {
				i18n.LanguageEnglish: "request parameter is missing apple bundle id",
			},
			"max": {
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
		"Environment": {
			"min": {
				i18n.LanguageEnglish: "wrong environment type",
			},
			"max": {
				i18n.LanguageEnglish: "wrong environment type",
			},
		},
		"Type": {
			"min": {
				i18n.LanguageEnglish: "error in type",
			},
			"max": {
				i18n.LanguageEnglish: "error in type",
			},
		},
	})
}
