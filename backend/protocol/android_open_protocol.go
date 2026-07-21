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

// AndroidAPIDownloadCertificateReq 下载安卓证书请求体。
type AndroidAPIDownloadCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AndroidAPISubmitAPKSigningJobReq 提交 APK 签名任务请求体。
type AndroidAPISubmitAPKSigningJobReq struct {
	// 签名方案
	SignatureSchema []int `json:"signatureSchema,omitempty" binding:"required,unique,dive,min=1,max=4" example:"1,2,3,4"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPISubmitAPKSigningJobRsp 提交 APK 签名任务响应体。
type AndroidAPISubmitAPKSigningJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPISubmitAABSigningJobReq 提交 AAB 签名任务请求体。
type AndroidAPISubmitAABSigningJobReq struct {
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPISubmitAABSigningJobRsp 提交 AAB 签名任务响应体。
type AndroidAPISubmitAABSigningJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPISubmitAPKPatchSigningJobReq 提交 APK 补丁包签名任务请求体。
type AndroidAPISubmitAPKPatchSigningJobReq struct {
	// 签名方案
	SignatureSchema []int `json:"signatureSchema,omitempty" binding:"required,unique,dive,min=1,max=4" example:"1,2,3,4"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 最低安卓 API 版本号
	MinimumSDKVersion int `json:"minimumSdkVersion,omitempty" binding:"required,gt=0" example:"19"`
}

// AndroidAPISubmitAPKPatchSigningJobRsp 提交 APK 补丁包签名任务响应体。
type AndroidAPISubmitAPKPatchSigningJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPIGetSigningJobInformationReq 获取任务信息请求体。
type AndroidAPIGetSigningJobInformationReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidAPIGetSigningJobInformationRsp 获取任务信息响应体。
type AndroidAPIGetSigningJobInformationRsp struct {
	// 类型
	Type int `json:"type,omitempty" example:"1"`
	//来源
	Source int `json:"source,omitempty" example:"1"`
	// 证书别名
	CertificateAlias string `json:"certificateAlias,omitempty" example:"my_certificate"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 签名配置
	SigningConfig string `json:"signingConfig,omitempty"`
	// 文件名
	FileName string `json:"fileName,omitempty"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 操作人
	User string `json:"user,omitempty" example:"zhangsan"`
	// 提交时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 结束时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2020-01-01 01:01:01"`
	// 日志
	Log string `json:"log,omitempty"`
}

// AndroidAPIListCertificatesRsp 获取安卓证书列表响应体。
type AndroidAPIListCertificatesRsp struct {
	// 证书信息
	List []*AndroidAPIListCertificatesItem `json:"list,omitempty"`
}

// AndroidAPIListCertificatesItem 安卓证书信息。
type AndroidAPIListCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 证书别名
	Alias string `json:"alias,omitempty" example:"my_certificate"`
	// 所有者
	Owner string `json:"owner,omitempty" example:"zhangsan"`
	// 签名算法
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty" example:"RSA_SHA256"`
	// 公钥算法
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty" example:"RSA_SHA256"`
	// 摘要
	SHA1 string `json:"sha1,omitempty" example:"3050762623C5FCADB879258983E758A2F9559AD2"`
	// 摘要
	SHA256 string `json:"sha256,omitempty" example:"50ad41624c25e493aa1dc7f4ab32bdc5a3b0b78ecc35b539936e3fea7c565af7"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 过期时间
	ExpiredTime string `json:"expiredTime,omitempty" example:"2020-01-01 01:01:01"`
	// 创建人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 生效时间
	EffectedTime string `json:"effectedTime,omitempty" example:"2020-01-01 01:01:01"`
	// 证书类型
	Type int `json:"type,omitempty" example:"1"`
	// 来源
	Source int `json:"source,omitempty" example:"1"`
}

func initAndroidOpenProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[AndroidAPIDownloadCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidAPISubmitAPKSigningJobReq](), validator.TranslationMessage{
		"SignatureSchema": {
			"required": {
				i18n.LanguageEnglish: "signature scheme must be selected",
			},
			"unique": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"len": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"max": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
		},
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "file does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidAPISubmitAABSigningJobReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "file does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidAPISubmitAPKPatchSigningJobReq](), validator.TranslationMessage{
		"SignatureSchema": {
			"required": {
				i18n.LanguageEnglish: "signature scheme must be selected",
			},
			"unique": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"len": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"max": {
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
		},
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "file does not exist",
			},
		},
		"MinimumSDKVersion": {
			"required": {
				i18n.LanguageEnglish: "please enter the minimum android api version number",
			},
			"gt": {
				i18n.LanguageEnglish: "illegal the minimum android api version number",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidAPIGetSigningJobInformationReq](), validator.TranslationMessage{
		"JobID": {
			"alphanum": {
				i18n.LanguageEnglish: "job does not exist",
			},
			"len": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
	})
}
