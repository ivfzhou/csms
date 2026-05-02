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
	HTTPPathWindowsInternalGetWHQLJob                     = "/internal/windows/getWHQLJob"
	HTTPPathWindowsInternalGetWHQLJobToInitialTestMachine = "/internal/windows/getWHQLJobToInitialTestMachine"
	HTTPPathWindowsInternalUpdateWHQLJob                  = "/internal/windows/updateWHQLJob"
	HTTPPathWindowsInternalGetWHQLJobToStartTest          = "/internal/windows/getWHQLJobToStartTest"
	HTTPPathWindowsInternalGetTestingWHQLJobs             = "/internal/windows/getTestingWHQLJobs"
	HTTPPathWindowsInternalGetMachineEVCertificates       = "/internal/windows/getMachineEVCertificates"
	HTTPPathWindowsInternalGetWindowsSigningJob           = "/internal/windows/getWindowsSigningJob"
	HTTPPathWindowsInternalGetCertificate                 = "/internal/windows/getCertificate"
	HTTPPathWindowsInternalUpdateSigningJob               = "/internal/windows/updateSigningJob"
	HTTPPathWindowsAPISubmitSigningJob                    = "/api/windows/submitSigningJob"
	HTTPPathWindowsAPIGetSigningJobInformation            = "/api/windows/getSigningJobInformation"
	HTTPPathWindowsAPISubmitWHQLJob                       = "/api/windows/submitWHQLJob"
	HTTPPathWindowsAPIGetWHQLJobInformation               = "/api/windows/getWHQLJobInformation"
)

// WindowsWebUploadCertificateReq 上传个人 OV 证书请求体。
type WindowsWebUploadCertificateReq struct {
	// 证书内容，Base64 编码
	Certificate string `json:"certificate,omitempty" binding:"required"`
	// 证书密码
	Password string `json:"password,omitempty" binding:"min=1,max=64" example:"123456"`
}

// WindowsWebListCertificatesRsp 证书列表响应体。
type WindowsWebListCertificatesRsp struct {
	// 证书列表
	List []*WindowsWebListCertificatesItem `json:"list,omitempty"`
}

// WindowsWebListCertificatesItem Windows 证书信息。
type WindowsWebListCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
	// 证书类型
	Type int `json:"type,omitempty" example:"1"`
	// 指纹
	Fingerprint string `json:"fingerprint,omitempty" example:"e4fd2280d075253896580e0c061730032e2b9c7a"`
	// 所有者
	Owner string `json:"owner,omitempty" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
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

// WindowsWebDownloadCertificateReq 下载证书请求体。
type WindowsWebDownloadCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
}

// WindowsWebAddEVCertificateReq 添加 EV 证书请求体。
type WindowsWebAddEVCertificateReq struct {
	// 指纹
	SHA1 string `json:"sha1,omitempty" binding:"len=40" example:"e4fd2280d075253896580e0c061730032e2b9c7a"`
	// 所有者
	Owner string `json:"owner,omitempty" binding:"required" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
	// 颁发者
	Publisher string `json:"publisher,omitempty" binding:"required" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
	// 签名算法
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty" binding:"required" example:"SHA1WithRSA"`
	// 公钥算法
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty" binding:"required" example:"SHA1WithRSA"`
	// 密码
	Password string `json:"password,omitempty" binding:"required,min=6,max=64" example:"123456"`
	// 机器 IP
	MachineIP string `json:"machineIp,omitempty" binding:"ipv4" example:"172.16.1.1"`
	// 序列号
	SerialNumber string `json:"serialNumber,omitempty" binding:"required,max=1024" example:"6e4314c191e2fe571b60818914982e53e03e7a20"`
	// 版本号
	Version int `json:"version,omitempty" binding:"gt=0" example:"3"`
	// 生效时间
	NotBefore Time `json:"notBefore" binding:"required" example:"2026-01-20 09:48:00"`
	// 失效时间
	NotAfter Time `json:"notAfter" binding:"required,gtfield=NotBefore" example:"2026-01-20 09:48:00"`
	// 是用于微软网站校验的证书
	IsMicrosoftVerifyCertificate bool `json:"isMicrosoftVerifyCertificate,omitempty" example:"false"`
	// 类型
	Type int `json:"type,omitempty" binding:"oneof=1 3" example:"1"`
}

// WindowsWebUploadCompanyCertificateReq 上传公司 OV 证书请求体。
type WindowsWebUploadCompanyCertificateReq struct {
	// 证书
	Certificate string `json:"certificate,omitempty" binding:"base64"`
	// 证书密码
	Password string `json:"password,omitempty" binding:"min=1,max=64" example:"123456"`
}

// WindowsWebListCompanyCertificatesRsp 获取后台管理中的 Windows 证书响应体。
type WindowsWebListCompanyCertificatesRsp struct {
	// 证书信息。
	List []*WindowsWebListCompanyCertificatesItem `json:"list,omitempty"`
}

// WindowsWebListCompanyCertificatesItem Windows 证书信息。
type WindowsWebListCompanyCertificatesItem struct {
	// 证书 ID
	ID string `json:"id,omitempty" example:"e4fd2280d075253896580e0c06173003"`
	// 类型
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
	// 过期时间
	NotAfter string `json:"notAfter,omitempty" example:"2026-01-20 09:48:00"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2026-01-20 09:48:00"`
	// 创建人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 密码
	Password string `json:"password,omitempty" example:"123456"`
	// 是否用于微软网站校验
	IsMSVerification bool `json:"isMSVerification,omitempty" example:"false"`
	// UKey 所在机器 IP
	MachineIP string `json:"machineIp,omitempty" example:"172.16.1.1"`
}

// WindowsWebGrantAppEVCertificateReq 授权应用使用个人 EV 证书请求体。
type WindowsWebGrantAppEVCertificateReq struct {
	// 应用 ID
	AppID string `json:"appId,omitempty" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"len=32,alphanum" example:"e4fd2280d075253896580e0c06173003"`
}

// WindowsWebGetCertificatePasswordReq 查看证书密码请求体。
type WindowsWebGetCertificatePasswordReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsWebGetCertificatePasswordRsp 查看证书密码响应体。
type WindowsWebGetCertificatePasswordRsp struct {
	// 密码
	Password string `json:"password,omitempty" example:"123456"`
}

// WindowsWebDownloadCompanyCertificateReq 下载公司证书请求体。
type WindowsWebDownloadCompanyCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsWebListGrantCertificateAppsReq 获取授权 Windows EV 证书应用列表请求体。
type WindowsWebListGrantCertificateAppsReq struct {
	// 应用 ID
	AppID string `form:"appId" binding:"omitempty,len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"omitempty,len=32,alphanum" example:"e4fd2280d075253896580e0c06173003"`
	// 每页条数
	PageSize int `form:"pageSize" binding:"gt=0,max=100" example:"10"`
	// 页数
	PageNumber int `form:"pageNumber" binding:"gt=0" example:"1"`
}

// WindowsWebListGrantCertificateAppsRsp 获取授权 Windows EV 证书应用列表响应体。
type WindowsWebListGrantCertificateAppsRsp struct {
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
	// 授权项
	List []*WindowsWebListGrantCertificateAppsItem `json:"list,omitempty"`
}

// WindowsWebListGrantCertificateAppsItem Windows 证书信息。
type WindowsWebListGrantCertificateAppsItem struct {
	// 应用 ID
	AppID string `json:"appId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" example:"e4fd2280d075253896580e0c06173003"`
	// 应用名
	AppName string `json:"appName,omitempty" example:"MyApp"`
	// 证书组织
	CertificateOrganization string `json:"certificateOrganization,omitempty" example:"C=CN, ST=HuNan, L=Changsha, O=ivfzhou, OU=ivfzhou, CN=windows personal ev certificate, emailAddress=ivfzhou@126.com"`
	// 授权时间
	GrantTime string `json:"grantTime,omitempty" example:"2026-01-20 09:48:00"`
	// 授权人
	User string `json:"user,omitempty" example:"zhangsan"`
}

// WindowsWebSubmitSigningJobReq 提交 Windows 签名请求体。
type WindowsWebSubmitSigningJobReq struct {
	// 签名类型
	SigningType int `json:"signingType,omitempty" binding:"oneof=1 2 3" example:"1"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" binding:"omitempty,len=32,alphanum" example:"e4fd2280d075253896580e0c06173003"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsWebListSigningJobsReq 获取 Windows 任务列表请求体。
type WindowsWebListSigningJobsReq struct {
	// 关键字
	KeyWord string `form:"keyWord"`
	// 任务类型
	SigningType int `form:"signingType" binding:"omitempty,oneof=1 2 3 4" example:"1"`
	// 任务状态
	Status int `form:"status" binding:"omitempty,oneof=1 2 3 4 5 6 7" example:"1"`
	// 页数
	PageNumber int `form:"pageNumber" binding:"min=1" example:"1"`
	// 一页任务条数
	PageSize int `form:"pageSize" binding:"min=1,max=100" example:"10"`
}

// WindowsWebListSigningJobsRsp 获取 Windows 任务列表响应体。
type WindowsWebListSigningJobsRsp struct {
	// 任务信息
	List []*WindowsWebListSigningJobsItem `json:"list,omitempty"`
	// 总数
	Count int `json:"count,omitempty" example:"10"`
}

// WindowsWebListSigningJobsItem Windows 签名任务信息。
type WindowsWebListSigningJobsItem struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 任务签名类型
	SigningType int `json:"signingType,omitempty" example:"1"`
	// 任务来源
	Source int `json:"source,omitempty" example:"1"`
	// 证书 ID
	CertificateID string `json:"certificateId,omitempty" example:"93aa1dc7f4ab32bdc5a3b0b78ecc35b5"`
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
}

// WindowsWebSubmitWHQLJobReq 提交 WHQL 任务请求体。
type WindowsWebSubmitWHQLJobReq struct {
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

// WindowsWebListWHQLJobsReq 获取 WHQL 任务列表请求体。
type WindowsWebListWHQLJobsReq struct {
	// 页数
	PageNumber int `form:"pageNumber" binding:"min=1" example:"1"`
	// 一页任务条数
	PageSize int `form:"pageSize" binding:"min=1,max=100" example:"10"`
}

// WindowsWebListWHQLJobsRsp 获取 WHQL 任务列表响应体。
type WindowsWebListWHQLJobsRsp struct {
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
	// 任务信息。
	List []*WindowsWebListWHQLJobsItem `json:"list,omitempty"`
}

// WindowsWebListWHQLJobsItem WHQL 任务信息。
type WindowsWebListWHQLJobsItem struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty"  example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 类型
	Type int `json:"type,omitempty"  example:"1"`
	// 来源
	Source int `json:"source,omitempty"  example:"1"`
	// 测试系统
	TestSystem string `json:"testSystem,omitempty"  example:"Windows 10 22H2_64"`
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
}

// WindowsWebRemoveCompanyCertificateReq 删除公司证书请求体。
type WindowsWebRemoveCompanyCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

// WindowsWebDeleteCertificateReq 删除证书请求体。
type WindowsWebDeleteCertificateReq struct {
	// 证书 ID
	CertificateID string `form:"certificateId" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
}

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

// WindowsInternalGetWHQLJobReq 获取 WHQL 任务信息请求体。
type WindowsInternalGetWHQLJobReq struct {
	// 任务 ID
	ID int `form:"id" binding:"gt=0"`
}

// WindowsInternalGetWHQLJobToInitialTestMachineReq 给 HLK 测试虚拟机初始化任务请求体。
type WindowsInternalGetWHQLJobToInitialTestMachineReq struct {
	// 测试系统版本
	System string `form:"system" binding:"hlktestsystem"`
}

// WindowsInternalUpdateWHQLJobReq 更新任务请求体。
type WindowsInternalUpdateWHQLJobReq struct {
	// 任务 ID
	JobID int `json:"jobId,omitempty" binding:"gt=0"`
	// 追加日志
	AppendLog string `json:"appendLog,omitempty"`
	// 更新任务状态
	Status int `json:"status,omitempty" binding:"omitempty,min=1,max=10"`
	// 测试机名
	TestMachineName string `json:"testMachineName,omitempty" binding:"omitempty,max=15"`
	// 服务名
	ServiceName string `json:"serviceName,omitempty" binding:"omitempty,max=256"`
	// 日志文件 ID
	HLKLogFileID string `json:"hlkLogFileId,omitempty" binding:"omitempty,len=38,alphanum"`
	// HLKX 包文件 ID
	HLKXFileID string `json:"hlkxFileId,omitempty" binding:"omitempty,len=38,alphanum"`
	// 测试结束时间。
	FinishedTestTime Time `json:"finishedTestTime"`
}

// WindowsInternalGetWHQLJobToStartTestReq 获取任务，调度测试的请求体。
type WindowsInternalGetWHQLJobToStartTestReq struct {
	// 测试系统版本。
	Systems []string `form:"systems" binding:"min=1,unique,dive,min=1"`
}

// WindowsInternalGetTestingWHQLJobsReq 获取正在测试中的任务请求体。
type WindowsInternalGetTestingWHQLJobsReq struct {
	// 测试系统版本。
	Systems []string `form:"systems" binding:"min=1,unique,dive,min=1"`
}

// WindowsInternalGetMachineEVCertificatesReq 获取签名机器上在用的 EV UKey 请求体。
type WindowsInternalGetMachineEVCertificatesReq struct {
	// IP 地址
	IP string `form:"ip" binding:"ipv4"`
}

// WindowsInternalGetWindowsSigningJobReq 获取签名任务信息请求体。
type WindowsInternalGetWindowsSigningJobReq struct {
	// 任务 ID
	JobID string `form:"jobId" binding:"len=38,alphanum"`
}

// WindowsInternalGetCertificateReq 获取证书信息请求体。
type WindowsInternalGetCertificateReq struct {
	// 证书 ID
	ID int `form:"id" binding:"gt=0"`
}

// WindowsInternalUpdateSigningJobReq 更新签名任务请求体。
type WindowsInternalUpdateSigningJobReq struct {
	// 任务 ID
	JobID string `json:"jobId,omitempty" binding:"len=38,alphanum"`
	// 状态
	Status int `json:"status,omitempty" binding:"omitempty,gt=0,lte=8"`
	// 追加日志
	AppendLog string `json:"appendLog,omitempty"`
	// 结果文件 ID
	SignedFileID string `json:"signedFileId,omitempty" binding:"omitempty,len=38,alphanum"`
	// 结束时间
	FinishedTime Time `json:"finishedTime"`
	// PE 签名结束时间
	FinishedPESignTime Time `json:"finishedPESignTime"`
}

func initWindows() {
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebUploadCertificateReq](), validator.TranslationMessage{
		"Certificate": {
			"required": {
				i18n.LanguageChinese: "证书缺失",
				i18n.LanguageEnglish: "certificate needed",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageChinese: "证书密码错误",
				i18n.LanguageEnglish: "certificate password incorrect",
			},
			"max": {
				i18n.LanguageChinese: "证书密码过长",
				i18n.LanguageEnglish: "certificate password is too long",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebDownloadCertificateReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebAddEVCertificateReq](), validator.TranslationMessage{
		"SHA1": {
			"len": {
				i18n.LanguageChinese: "证书指纹错误",
				i18n.LanguageEnglish: "certificate fingerprint error",
			},
		},
		"Publisher": {
			"required": {
				i18n.LanguageChinese: "请填写证书颁发者",
				i18n.LanguageEnglish: "please fill in the certificate issuer",
			},
		},
		"Owner": {
			"required": {
				i18n.LanguageChinese: "请填写证书所有者",
				i18n.LanguageEnglish: "please fill in the certificate owner",
			},
		},
		"SignatureAlgorithm": {
			"required": {
				i18n.LanguageChinese: "请填写证书签名算法",
				i18n.LanguageEnglish: "please fill in the certificate signing algorithm",
			},
		},
		"PublicKeyAlgorithm": {
			"required": {
				i18n.LanguageChinese: "请填写证书公钥算法",
				i18n.LanguageEnglish: "please fill in the certificate public key algorithm",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageChinese: "密码字符数过少",
				i18n.LanguageEnglish: "password has too few characters",
			},
			"max": {
				i18n.LanguageChinese: "密码字符数过多",
				i18n.LanguageEnglish: "too many password characters",
			},
		},
		"SerialNumber": {
			"required": {
				i18n.LanguageChinese: "请求输入序列号",
				i18n.LanguageEnglish: "serial number is required",
			},
			"max": {
				i18n.LanguageChinese: "序列号太长",
				i18n.LanguageEnglish: "serial number is too long",
			},
		},
		"Version": {
			"gt": {
				i18n.LanguageChinese: "请求输入正确的序列号",
				i18n.LanguageEnglish: "version is invalid",
			},
		},
		"MachineIP": {
			"ipv4": {
				i18n.LanguageChinese: "ip 地址非法",
				i18n.LanguageEnglish: "illegal ipv4 address",
			},
		},
		"NotBefore": {
			"required": {
				i18n.LanguageChinese: "请选择证书生效时间",
				i18n.LanguageEnglish: "please select the effective date of the certificate",
			},
		},
		"NotAfter": {
			"required": {
				i18n.LanguageChinese: "请选择证书失效时间",
				i18n.LanguageEnglish: "please select the certificate expiration time",
			},
			"gtfield": {
				i18n.LanguageChinese: "生效时间不能大于过期时间",
				i18n.LanguageEnglish: "effective time cannot exceed the expiration time",
			},
		},
		"Type": {
			"oneof": {
				i18n.LanguageChinese: "请选择证书类型",
				i18n.LanguageEnglish: "please select certificate type",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebUploadCompanyCertificateReq](), validator.TranslationMessage{
		"Certificate": {
			"base64": {
				i18n.LanguageChinese: "证书需为 base64 编码格式",
				i18n.LanguageEnglish: "certificate must be in base64 encoding format",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageChinese: "证书密码错误",
				i18n.LanguageEnglish: "certificate password incorrect",
			},
			"max": {
				i18n.LanguageChinese: "证书密码过长",
				i18n.LanguageEnglish: "certificate password is too long",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebGrantAppEVCertificateReq](), validator.TranslationMessage{
		"AppID": {
			"len": {
				i18n.LanguageChinese: "应用不存在于系统",
				i18n.LanguageEnglish: "app does not exist in the system",
			},
			"alphanum": {
				i18n.LanguageChinese: "应用不存在于系统",
				i18n.LanguageEnglish: "app does not exist in the system",
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
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebGetCertificatePasswordReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebDownloadCompanyCertificateReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebListGrantCertificateAppsReq](), validator.TranslationMessage{
		"AppID": {
			"len": {
				i18n.LanguageChinese: "应用不存在于系统",
				i18n.LanguageEnglish: "app does not exist in the system",
			},
			"alphanum": {
				i18n.LanguageChinese: "应用不存在于系统",
				i18n.LanguageEnglish: "app does not exist in the system",
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
		"PageSize": {
			"gt": {
				i18n.LanguageChinese: "请求每页条数非法",
				i18n.LanguageEnglish: "illegal number of items per page",
			},
			"max": {
				i18n.LanguageChinese: "每页条数过大",
				i18n.LanguageEnglish: "page size is too large",
			},
		},
		"PageNumber": {
			"gt": {
				i18n.LanguageChinese: "请求页数非法",
				i18n.LanguageEnglish: "illegal page number",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebSubmitSigningJobReq](), validator.TranslationMessage{
		"SigningType": {
			"oneof": {
				i18n.LanguageChinese: "未知的签名类型",
				i18n.LanguageEnglish: "unknown singing type",
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebListSigningJobsReq](), validator.TranslationMessage{
		"SigningType": {
			"oneof": {
				i18n.LanguageChinese: "未知的签名类型",
				i18n.LanguageEnglish: "unknown singing type",
			},
		},
		"Status": {
			"oneof": {
				i18n.LanguageChinese: "状态非法",
				i18n.LanguageEnglish: "invalid status",
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
				i18n.LanguageEnglish: "page count is too large",
				i18n.LanguageChinese: "每页条数过大",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebSubmitWHQLJobReq](), validator.TranslationMessage{
		"SigningType": {
			"oneof": {
				i18n.LanguageChinese: "未知的签名类型",
				i18n.LanguageEnglish: "unknown singing type",
			},
		},
		"TestSystem": {
			"hlktestsystem": {
				i18n.LanguageChinese: "未知的测试系统",
				i18n.LanguageEnglish: "unknown testing system",
			},
		},
		"ServiceName": {
			"max": {
				i18n.LanguageChinese: "服务名过长",
				i18n.LanguageEnglish: "service name is too long",
			},
		},
		"TestTarget": {
			"max": {
				i18n.LanguageChinese: "测试目标过长",
				i18n.LanguageEnglish: "test target is too long",
			},
		},
		"TestConfig": {
			"hlktestconfig": {
				i18n.LanguageChinese: "测试配置格式错误",
				i18n.LanguageEnglish: "invalid test config",
			},
			"max": {
				i18n.LanguageChinese: "配置内容过长",
				i18n.LanguageEnglish: "content is too large",
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebListWHQLJobsReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebRemoveCompanyCertificateReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsWebDeleteCertificateReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetWHQLJobReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetWHQLJobToInitialTestMachineReq](), validator.TranslationMessage{
		"System": {
			"hlktestsystem": {
				i18n.LanguageEnglish: "system is invalid",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalUpdateWHQLJobReq](), validator.TranslationMessage{
		"JobID": {
			"gt": {
				i18n.LanguageEnglish: "job id is invalid",
			},
		},
		"Status": {
			"min": {
				i18n.LanguageEnglish: "status is invalid",
			},
			"max": {
				i18n.LanguageEnglish: "status is invalid",
			},
		},
		"TestMachineName": {
			"max": {
				i18n.LanguageEnglish: "test machine name is invalid",
			},
		},
		"ServiceName": {
			"max": {
				i18n.LanguageEnglish: "service name is invalid",
			},
		},
		"HLKLogFileID": {
			"len": {
				i18n.LanguageEnglish: "hlk log file id is invalid",
			},
			"alphanum": {
				i18n.LanguageEnglish: "hlk log file id is invalid",
			},
		},
		"HLKXFileID": {
			"len": {
				i18n.LanguageEnglish: "hlkx file id is invalid",
			},
			"alphanum": {
				i18n.LanguageEnglish: "hlkx file id is invalid",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetWHQLJobToStartTestReq](), validator.TranslationMessage{
		"Systems": {
			"min": {
				i18n.LanguageEnglish: "systems is empty",
			},
			"unique": {
				i18n.LanguageEnglish: "systems duplicated",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetTestingWHQLJobsReq](), validator.TranslationMessage{
		"Systems": {
			"min": {
				i18n.LanguageEnglish: "systems is empty",
			},
			"unique": {
				i18n.LanguageEnglish: "systems duplicated",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetMachineEVCertificatesReq](), validator.TranslationMessage{
		"IP": {
			"ipv4": {
				i18n.LanguageEnglish: "ip is invalid",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetWindowsSigningJobReq](), validator.TranslationMessage{
		"JobID": {
			"alphanum": {
				i18n.LanguageEnglish: "job does not exist",
			},
			"len": {
				i18n.LanguageEnglish: "job does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalGetCertificateReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageEnglish: "certificate does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[WindowsInternalUpdateSigningJobReq](), validator.TranslationMessage{
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
			"lte": {
				i18n.LanguageEnglish: "status is invalid",
			},
		},
	})
}
