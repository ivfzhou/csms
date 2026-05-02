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

// AndroidWebAddOrganizationReq 添加证书主体请求体。
type AndroidWebAddOrganizationReq struct {
	// 通用名
	CommonName string `json:"commonName,omitempty" binding:"required,max=32" example:"company certificate"`
	// 所有者
	DName string `json:"dName,omitempty" binding:"required,max=256" example:"C=CN,ST=Hunan,L=Changsha,CN=company_android_cert"`
}

// AndroidWebListOrganizationsRsp 获取安卓证书主体信息列表响应体。
type AndroidWebListOrganizationsRsp struct {
	// 安卓证书主体信息
	List []*AndroidWebListOrganizationsItem `json:"list,omitempty"`
}

// AndroidWebListOrganizationsItem 安卓证书主体信息。
type AndroidWebListOrganizationsItem struct {
	// ID
	ID int `json:"id,omitempty" example:"1"`
	// 组织名
	CommonName string `json:"commonName,omitempty" example:"company certificate"`
	// 证书信息
	DName string `json:"dname,omitempty" example:"C=CN,ST=Hunan,L=Changsha,CN=company_android_cert"`
	// 添加人
	User string `json:"user,omitempty" example:"zhangsan"`
	// 添加时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
}

// AndroidWebApplyCertificateReq 申请安卓证书请求体。
type AndroidWebApplyCertificateReq struct {
	// 证书类型
	Type int `json:"type,omitempty" binding:"min=1,max=2" example:"1"`
	// 证书主体
	OwnerID int `json:"ownerId,omitempty" binding:"gt=0" example:"1"`
	// 证书别名
	Alias string `json:"alias,omitempty" binding:"required,varname,max=64" example:"my_certificate"`
}

// AndroidWebUploadCertificateReq 上传安卓证书请求体。
type AndroidWebUploadCertificateReq struct {
	// 证书类型
	Type int `json:"type,omitempty" binding:"min=1,max=2" example:"1"`
	// 密钥库密码
	Storepass string `json:"storepass,omitempty" binding:"required,min=6,max=32" example:"123456"`
	// 私钥密码
	Keypass string `json:"keypass,omitempty" binding:"required,min=6,max=32" example:"123456"`
	// 证书，Base64 编码
	Certificate string `json:"certificate,omitempty" binding:"base64"`
}

// AndroidWebListCertificatesRsp 获取安卓证书列表响应体。
type AndroidWebListCertificatesRsp struct {
	// 证书信息
	List []*AndroidWebListCertificatesItem `json:"list,omitempty"`
}

// AndroidWebListCertificatesItem 安卓证书信息。
type AndroidWebListCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"1"`
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
}

// AndroidWebDownloadCertificateReq 下载安卓证书请求体。
type AndroidWebDownloadCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AndroidWebGetGooglePlayCertificateReq 获取谷歌 Play 上传证书请求体。
type AndroidWebGetGooglePlayCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AndroidWebGetGooglePlayDeployCertificateReq 获取谷歌 Play 部署证书请求体。
type AndroidWebGetGooglePlayDeployCertificateReq struct {
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 公钥 PEM 格式
	PublicKey string `json:"publicKey,omitempty" binding:"required"`
}

// AndroidWebGetGooglePlayUpgradeCertificateReq 获取谷歌 Play 升级签名密钥请求体。
type AndroidWebGetGooglePlayUpgradeCertificateReq struct {
	// 部署证书 ID
	DeployCertificateID string `json:"deployCertificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 上传证书 ID
	UploadCertificateID string `json:"uploadCertificateId,omitempty" binding:"len=32,alphanum,nefield=DeployCertificateID" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 公钥 PEM 格式
	PublicKey string `json:"publicKey,omitempty" binding:"required"`
}

// AndroidWebGetCertificateFacebookDigestReq 获取证书的脸书摘要请求体。
type AndroidWebGetCertificateFacebookDigestReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AndroidWebGetCertificateFacebookDigestRsp 获取证书的脸书摘要响应体。
type AndroidWebGetCertificateFacebookDigestRsp struct {
	// 摘要值
	Digest string `json:"digest,omitempty"`
}

// AndroidWebSubmitAPKSigningJobReq 提交 APK 签名任务请求体。
type AndroidWebSubmitAPKSigningJobReq struct {
	// 签名方案
	SignatureSchema []int `json:"signatureSchema,omitempty" binding:"required,unique,dive,min=1,max=4" example:"1,2,3,4"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidWebSubmitAABSigningJobReq 提交 AAB 签名任务请求体。
type AndroidWebSubmitAABSigningJobReq struct {
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AndroidWebSubmitAPKPatchSigningJobReq 提交 APK 补丁包签名任务请求体。
type AndroidWebSubmitAPKPatchSigningJobReq struct {
	// 签名方案
	SignatureSchema []int `json:"signatureSchema,omitempty" binding:"required,unique,dive,min=1,max=4" example:"1,2,3,4"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 最低安卓 API 版本号
	MinimumSDKVersion int `json:"minimumSdkVersion,omitempty" binding:"required,gt=0" example:"19"`
}

// AndroidWebListSigningJobsReq 获取签名任务列表信息请求体。
type AndroidWebListSigningJobsReq struct {
	// 关键字
	KeyWord string `form:"keyWord"`
	// 状态
	Status int `form:"status" binding:"omitempty,oneof=1 2 3" example:"1"`
	// 证书别名
	CertificateAlias string `form:"certificateAlias" binding:"max=64" example:"my_certificate"`
	// 页数
	PageNumber int `form:"pageNumber" binding:"min=1" example:"1"`
	// 一页任务条数
	PageSize int `form:"pageSize" binding:"min=1,max=100" example:"10"`
}

// AndroidWebListSigningJobsRsp 获取签名任务列表信息响应体。
type AndroidWebListSigningJobsRsp struct {
	// 总数。
	Count int `json:"count,omitempty" example:"10"`
	// 任务信息。
	List []*AndroidWebListSigningJobsItem `json:"list,omitempty"`
}

// AndroidWebListSigningJobsItem 签名任务信息。
type AndroidWebListSigningJobsItem struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 类型
	Type int `json:"type,omitempty" example:"1"`
	//来源
	Source int `json:"source,omitempty" example:"1"`
	// 证书别名
	CertificateAlias string `json:"certificateAlias,omitempty" example:"my_certificate"`
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

// AndroidWebRemoveOrganizationReq 删除证书主体请求体。
type AndroidWebRemoveOrganizationReq struct {
	// 主体 ID
	ID int `form:"id" binding:"gt=0" example:"1"`
}

// AndroidWebDeleteCertificateReq 删除安卓证书请求体。
type AndroidWebDeleteCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

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

func initAndroid() {
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebAddOrganizationReq](), validator.TranslationMessage{
		"CommonName": {
			"required": {
				i18n.LanguageChinese: "请输入安卓证书组织名",
				i18n.LanguageEnglish: "please enter the android certificate organization name",
			},
			"max": {
				i18n.LanguageChinese: "安卓证书组织名过长",
				i18n.LanguageEnglish: "android certificate organization name is too long",
			},
		},
		"DName": {
			"required": {
				i18n.LanguageChinese: "请输入证书信息",
				i18n.LanguageEnglish: "please enter certificate information",
			},
			"max": {
				i18n.LanguageChinese: "证书信息过长",
				i18n.LanguageEnglish: "certificate information is too long",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebApplyCertificateReq](), validator.TranslationMessage{
		"Type": {
			"min": {
				i18n.LanguageChinese: "证书类型非法",
				i18n.LanguageEnglish: "certificate type is illegal",
			},
			"max": {
				i18n.LanguageChinese: "证书类型非法",
				i18n.LanguageEnglish: "certificate type is illegal",
			},
		},
		"OwnerID": {
			"gt": {
				i18n.LanguageChinese: "错误的证书主体",
				i18n.LanguageEnglish: "wrong certificate subject",
			},
		},
		"Alias": {
			"required": {
				i18n.LanguageChinese: "请输入证书别名",
				i18n.LanguageEnglish: "please enter certificate alias",
			},
			"varname": {
				i18n.LanguageChinese: "证书别名包含非法字符",
				i18n.LanguageEnglish: "certificate alias contains illegal characters",
			},
			"max": {
				i18n.LanguageChinese: "证书别名字符数过多",
				i18n.LanguageEnglish: "certificate alias has too many characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebUploadCertificateReq](), validator.TranslationMessage{
		"Type": {
			"min": {
				i18n.LanguageChinese: "证书类型非法",
				i18n.LanguageEnglish: "certificate type is illegal",
			},
			"max": {
				i18n.LanguageChinese: "证书类型非法",
				i18n.LanguageEnglish: "certificate type is illegal",
			},
		},
		"Storepass": {
			"required": {
				i18n.LanguageChinese: "请输入 storepass",
				i18n.LanguageEnglish: "please enter storepass",
			},
			"min": {
				i18n.LanguageChinese: "storepass 至少六位字符",
				i18n.LanguageEnglish: "storepass must be at least six characters long",
			},
			"max": {
				i18n.LanguageChinese: "storepass 字符数过多",
				i18n.LanguageEnglish: "too many characters in storepass",
			},
		},
		"Keypass": {
			"required": {
				i18n.LanguageChinese: "请输入 keypass",
				i18n.LanguageEnglish: "please enter keypass",
			},
			"min": {
				i18n.LanguageChinese: "keypass 至少六位字符",
				i18n.LanguageEnglish: "keypass must be at least six characters long",
			},
			"max": {
				i18n.LanguageChinese: "keypass 字符数过多",
				i18n.LanguageEnglish: "too many characters in keypass",
			},
		},
		"Certificate": {
			"base64": {
				i18n.LanguageChinese: "证书格式错误",
				i18n.LanguageEnglish: "certificate format error",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebDownloadCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebGetGooglePlayCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebGetGooglePlayDeployCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"PublicKey": {
			"required": {
				i18n.LanguageChinese: "请输入加密公钥",
				i18n.LanguageEnglish: "please enter the encryption public key",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebGetGooglePlayUpgradeCertificateReq](), validator.TranslationMessage{
		"DeployCertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"UploadCertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"nefield": {
				i18n.LanguageChinese: "不能选择相同的证书",
				i18n.LanguageEnglish: "cannot select the same certificate",
			},
		},
		"PublicKey": {
			"required": {
				i18n.LanguageChinese: "请输入加密公钥",
				i18n.LanguageEnglish: "please enter the encryption public key",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebGetCertificateFacebookDigestReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebSubmitAPKSigningJobReq](), validator.TranslationMessage{
		"SignatureSchema": {
			"required": {
				i18n.LanguageChinese: "须选择签名方案",
				i18n.LanguageEnglish: "signature scheme must be selected",
			},
			"unique": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"len": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"max": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
		},
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebSubmitAABSigningJobReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebSubmitAPKPatchSigningJobReq](), validator.TranslationMessage{
		"SignatureSchema": {
			"required": {
				i18n.LanguageChinese: "须选择签名方案",
				i18n.LanguageEnglish: "signature scheme must be selected",
			},
			"unique": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"len": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
			"max": {
				i18n.LanguageChinese: "选择了错误的签名方案",
				i18n.LanguageEnglish: "choosing the wrong signature scheme",
			},
		},
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
		"FileID": {
			"len": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "文件不存在",
				i18n.LanguageEnglish: "file does not exist",
			},
		},
		"MinimumSDKVersion": {
			"required": {
				i18n.LanguageChinese: "请填写最低安卓 API 版本号",
				i18n.LanguageEnglish: "please enter the minimum android api version number",
			},
			"gt": {
				i18n.LanguageChinese: "最低安卓 API 版本号非法",
				i18n.LanguageEnglish: "illegal the minimum android api version number",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebListSigningJobsReq](), validator.TranslationMessage{
		"Status": {
			"oneof": {
				i18n.LanguageChinese: "状态非法",
				i18n.LanguageEnglish: "invalid status",
			},
		},
		"CertificateAlias": {
			"max": {
				i18n.LanguageChinese: "证书别名太长",
				i18n.LanguageEnglish: "certificate alias is too long",
			},
		},
		"PageNumber": {
			"min": {
				i18n.LanguageChinese: "页码非法",
				i18n.LanguageEnglish: "illegal page numbers",
			},
		},
		"PageSize": {
			"min": {
				i18n.LanguageChinese: "页条数非法",
				i18n.LanguageEnglish: "illegal page count",
			},
			"max": {
				i18n.LanguageChinese: "每页条数过大",
				i18n.LanguageEnglish: "page size is too large",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebRemoveOrganizationReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageChinese: "安卓主体不存在",
				i18n.LanguageEnglish: "android organization does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AndroidWebDeleteCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "证书不存在",
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
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
