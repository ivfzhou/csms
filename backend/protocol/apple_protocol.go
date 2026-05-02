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

const (
	AppleFileTypeSigningCertificate = iota + 1
	AppleFileTypeProfile
	AppleFileTypePushCertificate
)

// AppleWebApplyBundleIDReq 申请 Bundle ID 请求体。
type AppleWebApplyBundleIDReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64" example:"cn.ivfzhou.test"`
	// 类型
	Type int `json:"type,omitempty" binding:"min=1,max=2" example:"1"`
}

// AppleWebModifyBundleIDReq 修改 Bundle ID 能力项请求体。
type AppleWebModifyBundleIDReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64" example:"cn.ivfzhou.test"`
	// 能力项。
	Capabilities []string `json:"capabilities,omitempty" binding:"unique,dive,capability" example:"in app purchase"`
}

// AppleWebApplyCertificateReq 申请苹果签名证书请求体。
type AppleWebApplyCertificateReq struct {
	// 类型
	Type string `form:"type" binding:"certificatetype" example:"IOS_DEVELOPMENT"`
}

// AppleWebListBundleIDsRsp 获取 Bundle ID 列表响应。
type AppleWebListBundleIDsRsp struct {
	// Bundle ID 信息
	List []*AppleWebListBundleIDsItem `json:"list,omitempty"`
}

// AppleWebListBundleIDsItem Bundle ID 信息。
type AppleWebListBundleIDsItem struct {
	// Bundle ID
	ID string `json:"id,omitempty"`
	// 环境
	Env int `json:"env,omitempty"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty"`
	// 创建人
	Creator string `json:"creator,omitempty"`
	// 能力项
	Capabilities []string `json:"capabilities,omitempty"`
}

// AppleWebListCertificatesRsp 获取苹果证书信息列表响应体。
type AppleWebListCertificatesRsp struct {
	// 证书信息
	List []*AppleWebListCertificatesItem `json:"list,omitempty"`
}

// AppleWebListCertificatesItem 证书信息。
type AppleWebListCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty"`
	// 证书类型
	Type string `json:"type,omitempty"`
	// 主体
	Owner string `json:"owner,omitempty"`
	// 签发人
	Publisher string `json:"publisher,omitempty"`
	// 签名算法
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty"`
	// 公钥算法
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty"`
	// 过期时间
	ExpirationTime string `json:"expirationTime,omitempty"`
	// 创建人
	Creator string `json:"creator,omitempty"`
}

// AppleWebRegisterDeviceReq 注册测试设备请求体。
type AppleWebRegisterDeviceReq struct {
	// 设备 ID
	UDID string `json:"udid,omitempty" binding:"required,max=128"`
	// 设备类型
	Device string `json:"device,omitempty" binding:"appledevice" example:"MAC"`
	// 备注
	Remark string `json:"remark,omitempty" binding:"max=1024"`
}

// AppleWebApplyProfileReq 申请描述文件请求体。
type AppleWebApplyProfileReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64" example:"cn.ivfzhou.test"`
	// 类型
	Type string `json:"type,omitempty" binding:"profiletype" example:"IOS_APP_DEVELOPMENT"`
	// 平台
	Platform string `json:"platform,omitempty" binding:"appleplatform" example:"IOS"`
}

// AppleWebListDevicesReq 设备列表请求体。
type AppleWebListDevicesReq struct {
	// 页数
	PageNumber int `form:"pageNumber" binding:"min=1" example:"1"`
	// 一页任务条数
	PageSize int `form:"pageSize" binding:"min=1,max=100" example:"10"`
}

// AppleWebListDevicesRsp 设备列表响应体。
type AppleWebListDevicesRsp struct {
	// 设备信息
	List []*AppleWebListDevicesItem `json:"list,omitempty"`
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
}

// AppleWebListDevicesItem 设备信息。
type AppleWebListDevicesItem struct {
	// 设备类型
	Model string `json:"model,omitempty"`
	// 设备 ID
	UDID string `json:"udid,omitempty"`
	// 添加人
	User string `json:"user,omitempty" example:"zhangsan"`
	// 添加时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 备注
	Remark string `json:"remark,omitempty"`
	// 状态
	Status int `json:"status,omitempty" example:"1"`
}

// AppleWebApplyInHouseProfileReq 申请企业内测描述文件请求体。
type AppleWebApplyInHouseProfileReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64" example:"cn.ivfzhou.test"`
}

// AppleWebApplyPushCertificateReq 申请 Push 证书请求体。
type AppleWebApplyPushCertificateReq struct {
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" binding:"required,max=64" example:"cn.ivfzhou.test"`
	// 环境
	Environment int `json:"environment,omitempty" binding:"min=1,max=3" example:"1"`
}

// AppleWebDeleteBundleIDReq 删除 Bundle ID 请求体。
type AppleWebDeleteBundleIDReq struct {
	// Bundle ID。
	BundleID string `form:"bundleId" binding:"required,max=64" example:"cn.ivfzhou.test"`
}

// AppleWebRemoveCertificateReq 删除苹果证书请求体。
type AppleWebRemoveCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AppleWebListAppCertificatesRsp 获取应用证书和描述文件列表响应体。
type AppleWebListAppCertificatesRsp struct {
	// 证书信息
	List []*AppleWebListAppCertificatesItem `json:"list,omitempty"`
}

// AppleWebListAppCertificatesItem 证书和描述信息。
type AppleWebListAppCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 证书 ID
	AppleID string `json:"appleId,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 环境
	Environment int `json:"environment,omitempty" example:"1"`
	// 文件类型
	Category int `json:"category,omitempty" example:"1"`
	// 类型
	Type string `json:"type,omitempty" example:"iOS_APPLE_IN_HOUSE"`
	// Bundle ID
	BundleID string `json:"bundleId,omitempty" example:"cn.ivfzhou.test"`
	// 证书所有者
	CertificateOwner string `json:"certificateOwner,omitempty" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
	// 所属人
	User string `json:"user,omitempty" example:"zhangsan"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 过期时间
	ExpirationTime string `json:"expirationTime,omitempty" example:"2020-01-01 01:01:01"`
	// 描述文件内容
	ProfileContent string `json:"profileContent,omitempty"`
}

// AppleWebSubmitSigningJobReq 提交签名任务请求体。
type AppleWebSubmitSigningJobReq struct {
	// 描述文件 ID
	ProfileID string `json:"profileId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 待文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// AppleWebListSigningJobsReq 获取应用签名任务信息列表请求体。
type AppleWebListSigningJobsReq struct {
	// 页数
	PageNumber int `form:"pageNumber" binding:"min=1" example:"1"`
	// 一页任务条数
	PageSize int `form:"pageSize" binding:"min=1,max=100" example:"10"`
}

// AppleWebListSigningJobsRsp 获取应用签名任务信息列表响应体。
type AppleWebListSigningJobsRsp struct {
	// 任务信息
	List []*AppleWebListSigningJobsItem `json:"list,omitempty"`
	// 总数
	Count int `json:"count,omitempty" example:"10"`
}

// AppleWebListSigningJobsItem 签名任务信息。
type AppleWebListSigningJobsItem struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
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
}

// AppleWebRemoveProfileReq 删除描述文件请求体。
type AppleWebRemoveProfileReq struct {
	// 描述文件 ID。
	ProfileID string `form:"profileId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AppleWebRemovePushCertificateReq 删除 Push 证书请求体。
type AppleWebRemovePushCertificateReq struct {
	// 证书 ID。
	CertificateID string `form:"certificateId,omitempty" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// AppleWebDownloadCertificateReq 下载证书和描述文件请求体。
type AppleWebDownloadCertificateReq struct {
	// ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 类型
	Type int `form:"type" binding:"min=1,max=3" example:"1"`
}

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

func initApple() {
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebApplyBundleIDReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "请输入 apple bundle id",
				i18n.LanguageEnglish: "please enter apple bundle id",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 字符数过多",
				i18n.LanguageEnglish: "apple bundle id has too many characters",
			},
		},
		"Type": {
			"min": {
				i18n.LanguageChinese: "错误的环境类型",
				i18n.LanguageEnglish: "wrong environment type",
			},
			"max": {
				i18n.LanguageChinese: "错误的环境类型",
				i18n.LanguageEnglish: "wrong environment type",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebModifyBundleIDReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "apple bundle id 不存在于系统",
				i18n.LanguageEnglish: "apple bundle id does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 不存在于系统",
				i18n.LanguageEnglish: "apple bundle id does not exist in the system",
			},
		},
		"Capabilities": {
			"unique": {
				i18n.LanguageChinese: "不可存在重复的能力项",
				i18n.LanguageEnglish: "there must be no duplicate ability items",
			},
			"capability": {
				i18n.LanguageChinese: "存在非法的能力项",
				i18n.LanguageEnglish: "illegal ability items exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebApplyCertificateReq](), validator.TranslationMessage{
		"Type": {
			"certificatetype": {
				i18n.LanguageChinese: "证书类型非法",
				i18n.LanguageEnglish: "certificate type is illegal",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebRegisterDeviceReq](), validator.TranslationMessage{
		"UDID": {
			"required": {
				i18n.LanguageChinese: "请填写 UDID",
				i18n.LanguageEnglish: "please enter udid",
			},
			"max": {
				i18n.LanguageChinese: "UDID 字符数过多",
				i18n.LanguageEnglish: "too many characters in udid",
			},
		},
		"Device": {
			"appledevice": {
				i18n.LanguageChinese: "设备类型错误",
				i18n.LanguageEnglish: "invalid device type",
			},
		},
		"Remark": {
			"max": {
				i18n.LanguageChinese: "备注字符过多",
				i18n.LanguageEnglish: "too many characters in remark",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebApplyProfileReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "请选择 apple bundle id",
				i18n.LanguageEnglish: "please select apple bundle id",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 不存在",
				i18n.LanguageEnglish: "apple bundle id does not exist",
			},
		},
		"Type": {
			"profiletype": {
				i18n.LanguageChinese: "描述文件类型非法",
				i18n.LanguageEnglish: "invalid provision file type",
			},
		},
		"Platform": {
			"appleplatform": {
				i18n.LanguageChinese: "描述文件平台类型非法",
				i18n.LanguageEnglish: "invalid provision file platform",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebListDevicesReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebApplyInHouseProfileReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "请选择 apple bundle id",
				i18n.LanguageEnglish: "please select apple bundle id",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 不存在",
				i18n.LanguageEnglish: "apple bundle id does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebApplyPushCertificateReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "请选择 apple bundle id",
				i18n.LanguageEnglish: "please select apple bundle id",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 不存在",
				i18n.LanguageEnglish: "apple bundle id does not exist",
			},
		},
		"Environment": {
			"min": {
				i18n.LanguageChinese: "非法的证书环境",
				i18n.LanguageEnglish: "illegal apple push certificate environment",
			},
			"max": {
				i18n.LanguageChinese: "非法的证书环境",
				i18n.LanguageEnglish: "illegal apple push certificate environment",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebDeleteBundleIDReq](), validator.TranslationMessage{
		"BundleID": {
			"required": {
				i18n.LanguageChinese: "apple bundle id 不存在于系统",
				i18n.LanguageEnglish: "apple bundle id does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "apple bundle id 不存在于系统",
				i18n.LanguageEnglish: "apple bundle id does not exist in the system",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebRemoveCertificateReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebSubmitSigningJobReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebListSigningJobsReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebRemoveProfileReq](), validator.TranslationMessage{
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
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebRemovePushCertificateReq](), validator.TranslationMessage{
		"CertificateID": {
			"len": {
				i18n.LanguageChinese: "apple push 证书不存在",
				i18n.LanguageEnglish: "apple push certificate does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "push 证书不存在",
				i18n.LanguageEnglish: "apple push certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppleWebDownloadCertificateReq](), validator.TranslationMessage{
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
