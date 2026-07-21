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
	HTTPPathAndroidInternalGetSigningJob       = "/internal/android/getSigningJob"
	HTTPPathAndroidInternalGetCertificate      = "/internal/android/getCertificate"
	HTTPPathAndroidInternalUpdateSigningJob    = "/internal/android/updateSigningJob"
	HTTPPathAndroidAPISubmitAPKSigningJob      = "/api/android/submitAPKSigningJob"
	HTTPPathAndroidAPISubmitAABSigningJob      = "/api/android/submitAABSigningJob"
	HTTPPathAndroidAPISubmitAPKPatchSigningJob = "/api/android/submitAPKPatchSigningJob"
	HTTPPathAndroidAPIGetSigningJobInformation = "/api/android/getSigningJobInformation"
)

// AndroidInternalGetSigningJobReq 获取签名任务信息请求体。
type AndroidInternalGetSigningJobReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum"`
}

// AndroidInternalGetCertificateReq 获取安卓证书信息请求体。
type AndroidInternalGetCertificateReq struct {
	// 证书 ID
	ID int `form:"id" binding:"gt=0"`
}

// AndroidInternalUpdateSigningJobReq 更新任务信息请求体。
type AndroidInternalUpdateSigningJobReq struct {
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

func initAndroidInternalProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[AndroidInternalGetSigningJobReq](), validator.TranslationMessage{
		"JobID": {
			"alphanum": {
				i18n.LanguageEnglish: "job does not exist",
			},
			"len": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidInternalGetCertificateReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidInternalUpdateSigningJobReq](), validator.TranslationMessage{
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
}
