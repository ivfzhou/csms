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
	HTTPPathAppleInternalGetSigningJob            = "/internal/apple/getSigningJob"
	HTTPPathAppleInternalUpdateSigningJob         = "/internal/apple/updateSigningJob"
	HTTPPathAppleInternalGetCertificateAndProfile = "/internal/apple/getCertificateAndProfile"
	HTTPPathAppleAPISubmitSigningJob              = "/api/apple/submitSigningJob"
	HTTPPathAppleAPIGetSigningJobInformation      = "/api/apple/getSigningJobInformation"
)

// AppleInternalGetSigningJobReq 获取任务信息请求体。
type AppleInternalGetSigningJobReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum"`
}

// AppleInternalUpdateSigningJobReq 更新任务信息请求体。
type AppleInternalUpdateSigningJobReq struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" binding:"len=38,alphanum"`
	// 状态
	Status int `json:"status,omitempty" binding:"omitempty,gt=0"`
	// 日志
	AppendLog string `json:"appendLog,omitempty"`
	// 结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" binding:"omitempty,len=38,alphanum"`
	// 结束时间
	FinishedTime Time `json:"finishedTime"`
}

// AppleInternalGetCertificateAndProfileReq 获取证书和描述文件请求体。
type AppleInternalGetCertificateAndProfileReq struct {
	// 描述文件 ID
	ProfileID int `form:"profileId" binding:"gt=0"`
}

// AppleInternalGetCertificateAndProfileRsp 获取证书和描述文件响应体。
type AppleInternalGetCertificateAndProfileRsp struct {
	// 证书，Base64编码
	Certificate string `json:"certificate,omitempty"`
	// 描述文件，Base64编码
	Profile string `json:"profile,omitempty"`
	// 证书密码
	Password string `json:"password,omitempty"`
}

func initAppleInternalProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[AppleInternalGetSigningJobReq](), validator.TranslationMessage{
		"JobID": {
			"len": {
				i18n.LanguageEnglish: "apple signing job does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "apple signing job does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleInternalUpdateSigningJobReq](), validator.TranslationMessage{
		"JobID": {
			"alphanum": {
				i18n.LanguageEnglish: "job does not exist",
			},
			"len": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
		"SignedFileID": {
			"alphanum": {
				i18n.LanguageEnglish: "file id is invalid",
			},
			"len": {
				i18n.LanguageEnglish: "file id is invalid",
			},
		},
		"Status": {
			"gt": {
				i18n.LanguageEnglish: "status is invalid",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleInternalGetCertificateAndProfileReq](), validator.TranslationMessage{
		"ProfileID": {
			"gt": {
				i18n.LanguageEnglish: "profile does not exist",
			},
		},
	})
}
