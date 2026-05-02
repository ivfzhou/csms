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

// OpenWebApplyReq 申请请求体。
type OpenWebApplyReq struct {
	// 凭证 ID
	AccountID string `json:"accountId,omitempty" binding:"required,max=64,varname" example:"test_account"`
	// 授权项
	Authorities []int `json:"authorities,omitempty" binding:"omitempty,unique,dive,gt=0,max=23" example:"1,2,3"`
	// 请求源，逗号分割
	RequestIP string `json:"requestIp,omitempty" binding:"required,max=256" example:"*,192.168.0.1,127.0.0.1-127.0.0.128"`
	// 请求频率
	Frequency int `json:"frequency,omitempty" binding:"required,min=10,max=500" example:"100"`
}

// OpenWebApplyRsp 申请响应体。
type OpenWebApplyRsp struct {
	// 凭证密码
	Password string `json:"password,omitempty"`
}

// OpenWebUpdateReq 修改请求体。
type OpenWebUpdateReq struct {
	// 凭证 ID
	AccountID string `json:"accountId,omitempty" binding:"required,max=64,varname" example:"test_account"`
	// 授权项
	Authorities []int `json:"authorities,omitempty" binding:"omitempty,unique,dive,gt=0,max=23" example:"1,2,3"`
	// 请求源
	RequestIP string `json:"requestIp,omitempty" binding:"required,max=256" example:"*,192.168.0.1,127.0.0.1-127.0.0.128"`
	// 请求频率
	Frequency int `json:"frequency,omitempty" binding:"required,min=10,max=500" example:"100"`
}

// OpenWebGetInformationReq 获取请求凭证信息请求体。
type OpenWebGetInformationReq struct {
	// 凭证 ID
	AccountID string `form:"accountId" binding:"required,max=64,varname" example:"test_account"`
}

// OpenWebGetInformationRsp 获取请求凭证信息响应体。
type OpenWebGetInformationRsp struct {
	// 应用 ID
	AppID string `json:"appId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 凭证 ID
	AccountID string `json:"accountId,omitempty" example:"test_account"`
	// 授权项
	Authorities []int `json:"authorities,omitempty" example:"1,2,3"`
	// 请求源
	RequestIP string `json:"requestIp,omitempty" example:"*,192.168.0.1,127.0.0.1-127.0.0.128"`
	// 请求频率
	Frequency int `json:"frequency,omitempty" example:"100"`
	// 创建人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 过期时间
	Expiration string `json:"expiration,omitempty" example:"2020-01-01 01:01:01"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
}

// OpenWebListReq 获取请求凭证列表请求体。
type OpenWebListReq struct {
	// 页数
	PageNumber int `form:"pageNumber" binding:"gt=0" example:"1"`
	// 每页条数
	PageSize int `form:"pageSize" binding:"gt=0,max=100" example:"10"`
}

// OpenWebListRsp 获取请求凭证列表响应体。
type OpenWebListRsp struct {
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
	// 每项信息
	List []*OpenWebListItem `json:"list,omitempty"`
}

// OpenWebListItem 获取请求凭证列表的每项信息。
type OpenWebListItem struct {
	// 凭证 ID
	AccountID string `json:"accountId,omitempty" example:"test_account"`
	// 授权项
	Authorities []int `json:"authorities,omitempty" example:"1,2,3"`
	// 请求频率
	Frequency int `json:"frequency,omitempty" example:"100"`
	// 请求源
	RequestIP string `json:"requestIp,omitempty" example:"*,192.168.0.1,127.0.0.1-127.0.0.128"`
	// 创建人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 过期时间
	Expiration string `json:"expiration,omitempty" example:"2020-01-01 01:01:01"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
}

// OpenWebRenewalReq 请求凭证续期请求体。
type OpenWebRenewalReq struct {
	// 凭证 ID
	AccountID string `form:"accountId" binding:"required,max=64,varname" example:"test_account"`
}

// OpenWebResetReq 重置密钥请求体。
type OpenWebResetReq struct {
	// 凭证 ID
	AccountID string `form:"accountId" binding:"required,max=64,varname" example:"test_account"`
}

// OpenWebResetRsp 重置密钥响应体。
type OpenWebResetRsp struct {
	// 凭证密码
	Password string `json:"password,omitempty"`
}

// OpenWebRemoveReq 删除请求凭证请求体。
type OpenWebRemoveReq struct {
	// 凭证 ID
	AccountID string `form:"accountId" binding:"required,max=64,varname" example:"test_account"`
}

func initOpen() {
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebApplyReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "请填写凭证 ID",
				i18n.LanguageEnglish: "please fill in the auth id",
			},
			"max": {
				i18n.LanguageChinese: "请求凭证 ID 字符数过多",
				i18n.LanguageEnglish: "api account with too many characters",
			},
			"varname": {
				i18n.LanguageChinese: "请求凭证 ID 须由字母、数字和下划线组成，且不能以数字开头",
				i18n.LanguageEnglish: "api account must consist of letters, numbers, and underscores, and cannot start with a number",
			},
		},
		"Authorities": {
			"unique": {
				i18n.LanguageChinese: "重复选择了授权项",
				i18n.LanguageEnglish: "repeatedly selected authorization item",
			},
			"gt": {
				i18n.LanguageChinese: "选择了错误的功能授权项",
				i18n.LanguageEnglish: "selected the wrong api authorization item",
			},
			"max": {
				i18n.LanguageChinese: "选择了错误的功能授权项",
				i18n.LanguageEnglish: "selected the wrong api authorization item",
			},
		},
		"RequestIP": {
			"required": {
				i18n.LanguageChinese: "请填写请求源 IP",
				i18n.LanguageEnglish: "please fill in the request source ip address",
			},
			"max": {
				i18n.LanguageChinese: "请求源 IP 字符数过多",
				i18n.LanguageEnglish: "request source ip has too many characters",
			},
		},
		"Frequency": {
			"required": {
				i18n.LanguageChinese: "请填写请求频率",
				i18n.LanguageEnglish: "please fill in the request frequency",
			},
			"min": {
				i18n.LanguageChinese: "请求频率过小",
				i18n.LanguageEnglish: "request frequency too low",
			},
			"max": {
				i18n.LanguageChinese: "请求频率过大",
				i18n.LanguageEnglish: "request frequency too high",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebUpdateReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
		},
		"Authorities": {
			"unique": {
				i18n.LanguageChinese: "重复选择了授权项",
				i18n.LanguageEnglish: "repeatedly selected authorization item",
			},
			"gt": {
				i18n.LanguageChinese: "选择了错误的功能授权项",
				i18n.LanguageEnglish: "selected the wrong api authorization item",
			},
			"max": {
				i18n.LanguageChinese: "选择了错误的功能授权项",
				i18n.LanguageEnglish: "selected the wrong api authorization item",
			},
		},
		"RequestIP": {
			"required": {
				i18n.LanguageChinese: "请填写请求源 IP",
				i18n.LanguageEnglish: "please fill in the request source ip address",
			},
			"max": {
				i18n.LanguageChinese: "请求源 IP 字符数过多",
				i18n.LanguageEnglish: "request source ip has too many characters",
			},
		},
		"Frequency": {
			"required": {
				i18n.LanguageChinese: "请填写请求频率",
				i18n.LanguageEnglish: "please fill in the request frequency",
			},
			"min": {
				i18n.LanguageChinese: "请求频率过小",
				i18n.LanguageEnglish: "request frequency too low",
			},
			"max": {
				i18n.LanguageChinese: "请求频率过大",
				i18n.LanguageEnglish: "request frequency too high",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebGetInformationReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebListReq](), validator.TranslationMessage{
		"PageNumber": {
			"gt": {
				i18n.LanguageChinese: "错误的页数",
				i18n.LanguageEnglish: "wrong page number",
			},
		},
		"PageSize": {
			"gt": {
				i18n.LanguageChinese: "错误的每页条数",
				i18n.LanguageEnglish: "incorrect number of items per page",
			},
			"max": {
				i18n.LanguageChinese: "每页条数过大",
				i18n.LanguageEnglish: "page size is too large",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebRenewalReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebResetReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[OpenWebRemoveReq](), validator.TranslationMessage{
		"AccountID": {
			"required": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "凭证 ID 不存在于系统",
				i18n.LanguageEnglish: "api account does not exist in the system",
			},
		},
	})
}
