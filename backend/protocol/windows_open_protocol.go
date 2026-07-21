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

// WindowsAPIDownloadCertificateReq 下载证书请求体。
type WindowsAPIDownloadCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPIGetCertificatePasswordReq 查看证书密码请求体。
type WindowsAPIGetCertificatePasswordReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPIGetCertificatePasswordRsp 查看证书密码响应体。
type WindowsAPIGetCertificatePasswordRsp struct {
	// 密码
	Password string `json:"password,omitempty" example:"123456"`
}

// WindowsAPISubmitSigningJobReq 提交 Windows 签名请求体。
type WindowsAPISubmitSigningJobReq struct {
	// 签名类型
	SigningType int `json:"signingType,omitempty" binding:"oneof=1 2 3" example:"1"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"omitempty,len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPISubmitSigningJobRsp 提交 Windows 签名响应体。
type WindowsAPISubmitSigningJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPISubmitWHQLJobReq 提交 WHQL 任务请求体。
type WindowsAPISubmitWHQLJobReq struct {
	// 类型
	SigningType int `json:"signingType,omitempty" binding:"oneof=1 2" example:"1"`
	// 测试系统
	TestSystem string `json:"testSystem,omitempty" binding:"hlktestsystem" example:"Windows 10 22H2_64"`
	// 服务名称
	ServiceName string `json:"serviceName,omitempty" binding:"max=256"`
	// HLK Studio 测试对象
	TestTarget string `json:"testTarget,omitempty" binding:"max=256"`
	// HLK 测试配置
	TestConfig string `json:"testConfig,omitempty" binding:"omitempty,hlktestconfig,max=65535"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPISubmitWHQLJobRsp 提交 WHQL 任务响应体。
type WindowsAPISubmitWHQLJobRsp struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPIListCertificatesRsp 证书列表响应体。
type WindowsAPIListCertificatesRsp struct {
	// 证书列表
	List []*WindowsAPIListCertificatesItem `json:"list,omitempty"`
}

// WindowsAPIListCertificatesItem Windows 证书信息。
type WindowsAPIListCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 证书类型
	Type int `json:"type,omitempty" example:"1"`
	// 指纹
	Fingerprint string `json:"fingerprint,omitempty" example:"e4fd2280d075253896580e0c061730032e2b9c7a"`
	// 所有者
	Owner string `json:"owner,omitempty" example:"zhangsan"`
	// 颁发者
	Publisher string `json:"publisher,omitempty" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
	// 签名算法
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty" example:"SHA1WithRSA"`
	// 公钥算法
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty" example:"SHA1WithRSA"`
	// 生效时间
	NotBefore string `json:"notBefore,omitempty" example:"2026-01-20 09:48:00"`
	// 失效时间
	NotAfter string `json:"notAfter,omitempty" example:"2026-01-20 09:48:00"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2026-01-20 09:48:00"`
	// 创建人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
}

// WindowsAPIGetSigningJobInformationReq 获取签名任务信息请求体。
type WindowsAPIGetSigningJobInformationReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPIGetSigningJobInformationRsp 获取签名任务信息响应体。
type WindowsAPIGetSigningJobInformationRsp struct {
	// 任务签名类型
	SigningType int `json:"signingType,omitempty" example:"1"`
	// 任务来源
	Source int `json:"source,omitempty" example:"1"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 证书主体
	CertificateCommonName string `json:"certificateCommonName,omitempty" example:"MyCert"`
	// 源文件 ID
	FileID string `json:"fileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 文件名
	FileName string `json:"fileName,omitempty"`
	//提交人
	UserName string `json:"userName,omitempty" example:"zhangsan"`
	// 提交时间
	CreatedTime string `json:"createdTime,omitempty" example:"2026-01-20 09:48:00"`
	// 结束时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2026-01-20 09:48:00"`
	// 任务日志
	Log string `json:"log,omitempty"`
	// 状态
	Status int `json:"status,omitempty" example:"1"`
	// 结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 微软产品 ID
	ProductID string `json:"productId,omitempty" example:"15324896572190743"`
	// 微软提交 ID
	SubmissionID string `json:"submissionId,omitempty" example:"1713904832489657219"`
}

// WindowsAPIGetWHQLJobInformationReq 获取 WHQL 任务信息请求体。
type WindowsAPIGetWHQLJobInformationReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsAPIGetWHQLJobInformationRsp 获取 WHQL 任务信息响应体。
type WindowsAPIGetWHQLJobInformationRsp struct {
	// 类型
	Type int `json:"type,omitempty" example:"1"`
	// 来源
	Source int `json:"source,omitempty" example:"1"`
	// 测试系统
	TestSystem string `json:"testSystem,omitempty" example:"Windows 10 22H2_64"`
	// 文件名
	FileName string `json:"fileName,omitempty"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// HLKX 包文件 ID
	HLKXFileID string `json:"hlkxFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// HLKX 包文件名
	HLKXFileName string `json:"hlkxFileName,omitempty"`
	// HLK 日志文件 ID
	HLKLogFileID string `json:"hlkLogFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// HLKX 日志文件名
	HLKLogFileName string `json:"hlkLogFileName,omitempty"`
	// 签名结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 签名结果文件名
	SignedFileName string `json:"signedFileName,omitempty"`
	// 提交人
	UserName string `json:"userName,omitempty" example:"zhangsan"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2026-01-20 09:48:00"`
	// 结束时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2026-01-20 09:48:00"`
	// 状态
	Status int `json:"status,omitempty" example:"1"`
	// 日志
	Log string `json:"log,omitempty"`
	// 微软产品 ID
	ProductID string `json:"productId,omitempty" example:"15324896572190743"`
	// 微软提交 ID
	SubmissionID string `json:"submissionId,omitempty" example:"1713904832489657219"`
}

func initWindowsOpenProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPIDownloadCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPIGetCertificatePasswordReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPISubmitSigningJobReq](), validator.TranslationMessage{
		"SigningType": {
			"oneof": {
				i18n.LanguageEnglish: "unknown singing type",
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPISubmitWHQLJobReq](), validator.TranslationMessage{
		"SigningType": {
			"oneof": {
				i18n.LanguageEnglish: "unknown singing type",
			},
		},
		"TestSystem": {
			"hlktestsystem": {
				i18n.LanguageEnglish: "unknown testing system",
			},
		},
		"ServiceName": {
			"max": {
				i18n.LanguageEnglish: "service name is too long",
			},
		},
		"TestTarget": {
			"max": {
				i18n.LanguageEnglish: "test target is too long",
			},
		},
		"TestConfig": {
			"hlktestconfig": {
				i18n.LanguageEnglish: "unknown test config",
			},
			"max": {
				i18n.LanguageEnglish: "content is too large",
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPIGetSigningJobInformationReq](), validator.TranslationMessage{
		"JobID": {
			"alphanum": {
				i18n.LanguageEnglish: "job does not exist",
			},
			"len": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsAPIGetWHQLJobInformationReq](), validator.TranslationMessage{
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
