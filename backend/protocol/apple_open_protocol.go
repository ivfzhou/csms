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

// AppleAPIDownloadCertificateReq 下载证书和描述文件请求体。
type AppleAPIDownloadCertificateReq struct {
	// ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 类型
	Type int `form:"type" binding:"min=1,max=3" example:"1"`
}

// AppleAPISubmitSigningJobReq 提交签名任务请求体。
type AppleAPISubmitSigningJobReq struct {
	// 描述文件 ID
	ProfileID string `json:"profileId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 待文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AppleAPISubmitSigningJobRsp 提交签名任务响应体。
type AppleAPISubmitSigningJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AppleAPIGetSigningJobInformationReq 获取签名任务信息请求体。
type AppleAPIGetSigningJobInformationReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AppleAPIGetSigningJobInformationRsp 获取签名任务信息响应体。
type AppleAPIGetSigningJobInformationRsp struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" example:"cn.ivfzhou.test"`
	// 描述文件 ID
	ProfileID string `json:"profileId,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 待签名文件 ID
	FileID string `json:"fileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 待签名文件名
	FileName string `json:"fileName,omitempty"`
	// 提交人
	UserName string `json:"userName,omitempty" example:"zhangsan"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 结束时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2020-01-01 01:01:01"`
	// 状态
	Status int `json:"status,omitempty" example:"1"`
	// 日志
	Log string `json:"log,omitempty"`
	// 结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 结果文件名
	SignedFileName string `json:"signedFileName,omitempty"`
	// 来源
	Source int `json:"source,omitempty" example:"1"`
}

func initAppleOpenProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[AppleAPIDownloadCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书文件不存在",
				i18n.LanguageEnglish: "certificate file does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书文件不存在",
				i18n.LanguageEnglish: "certificate file does not exist",
			},
		},
		"Type": {
			"min": {
				i18n.LanguageChinese: "类型错误",
				i18n.LanguageEnglish: "invalid type",
			},
			"max": {
				i18n.LanguageChinese: "类型错误",
				i18n.LanguageEnglish: "invalid type",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleAPISubmitSigningJobReq](), validator.TranslationMessage{
		"ProfileID": {
			"len": {
				i18n.LanguageChinese: "描述文件不存在",
				i18n.LanguageEnglish: "apple profile does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "描述文件不存在",
				i18n.LanguageEnglish: "apple profile does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageChinese: "清上传待签名文件",
				i18n.LanguageEnglish: "please upload file to be signed",
			},
			"alphanum": {
				i18n.LanguageChinese: "待签名文件不存在",
				i18n.LanguageEnglish: "file to be signed not found",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleAPIGetSigningJobInformationReq](), validator.TranslationMessage{
		"JobID": {
			"len": {
				i18n.LanguageChinese: "签名任务不存在",
				i18n.LanguageEnglish: "apple signing job does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "签名任务不存在",
				i18n.LanguageEnglish: "apple signing job does not exist",
			},
		},
	})
}
