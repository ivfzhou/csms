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

func initWindowsInternalProtocol() {
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
