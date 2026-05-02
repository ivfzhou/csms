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

// TodoWebCountRsp 获取用户待办、已办数量响应体。
type TodoWebCountRsp struct {
	// 待处理数
	NeedToDeal int64 `json:"needToDeal,omitempty"`
	// 已处理数
	Done int64 `json:"done,omitempty"`
}

// TodoWebListReq 获取需要处理的待办请求体。
type TodoWebListReq struct {
	// 页数
	PageNumber int `form:"pageNumber" binding:"gt=0" example:"1"`
	// 每页条数
	PageSize int `form:"pageSize" binding:"gt=0,max=100" example:"10"`
}

// TodoWebListRsp 获取需要处理的待办响应体。
type TodoWebListRsp struct {
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
	// 待办信息
	List []*TodoWebListItem `json:"list,omitempty"`
}

// TodoWebListItem 待办信息。
type TodoWebListItem struct {
	// 待办 ID
	ID int `json:"id,omitempty" example:"1"`
	// 应用名
	AppName string `json:"appName,omitempty" example:"MyApp"`
	// 申请时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 申请人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 待办类型
	Type int `json:"type,omitempty" example:"1"`
	// 待办状态
	Status int `json:"status,omitempty" example:"1"`
}

// TodoWebListDealtReq 获取已处理的待办列表请求体。
type TodoWebListDealtReq struct {
	// 应用 ID
	AppID string `form:"appId" binding:"omitempty,len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 待办类型
	Types []int `form:"types" binding:"omitempty,unique,dive,gt=0,max=5" example:"2,3"`
	// 待办状态
	Status []int `form:"status" binding:"omitempty,unique,dive,ne=1,gt=0,max=3" example:"2,3"`
	// 页数
	PageNumber int `form:"pageNumber" binding:"gt=0" example:"1"`
	// 每页条数
	PageSize int `form:"pageSize" binding:"gt=0,max=100" example:"10"`
}

// TodoWebListDealtRsp 获取已处理的待办列表响应体。
type TodoWebListDealtRsp struct {
	// 总数
	Count int64 `json:"count,omitempty" example:"10"`
	// 待办信息
	List []*TodoWebListDealtItem `json:"list,omitempty"`
}

// TodoWebListDealtItem 待办信息。
type TodoWebListDealtItem struct {
	// 待办 ID
	ID int `json:"id,omitempty" example:"1"`
	// 应用名
	AppName string `json:"appName,omitempty" example:"MyApp"`
	// 申请时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 申请人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 待办类型
	Type int `json:"type,omitempty" example:"1"`
	// 待办状态
	Status int `json:"status,omitempty" example:"1"`
	// 处理人
	Approver string `json:"approver,omitempty" example:"zhangsan"`
	// 处理时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2020-01-01 01:01:01"`
}

// TodoWebCreateReq 创建待办请求体。
type TodoWebCreateReq struct {
	// 待办类型
	Type int `json:"type,omitempty" binding:"oneof=2 3" example:"2"`
	// 应用 ID
	AppID string `json:"appId,omitempty" binding:"len=32,alphanum" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 理由
	ApplyReason string `json:"applyReason,omitempty" binding:"required,max=256"`
}

// TodoWebGetDetailReq 获取待办详情请求体。
type TodoWebGetDetailReq struct {
	// 待办 ID
	ID int `form:"id" binding:"gt=0" example:"1"`
}

// TodoWebGetDetailRsp 获取待办详情响应体。
type TodoWebGetDetailRsp struct {
	// 待办 ID
	ID int `json:"id,omitempty" example:"1"`
	// 应用 ID
	AppID string `json:"appId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 应用名
	AppName string `json:"appName,omitempty" example:"MyApp"`
	// 申请时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 01:01:01"`
	// 申请人
	Creator string `json:"creator,omitempty" example:"zhangsan"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 待办类型
	Type int `json:"type,omitempty" example:"1"`
	// 待办状态
	Status int `json:"status,omitempty" example:"1"`
	// 申请人所属部门
	Department string `json:"department,omitempty" example:"/技术部"`
	// 可处理的人
	Candidates []string `json:"candidates,omitempty" example:"zhangsan"`
	// 处理人
	ApproveBy string `json:"approveBy,omitempty" example:"zhangsan"`
	// 测试设备 UDID
	DeviceUdid string `json:"deviceUdid,omitempty"`
	// 测试设备型号
	DeviceModel string `json:"deviceModel,omitempty"`
	// 申请理由
	ApplyReason string `json:"applyReason,omitempty"`
	// 审批语
	ApproveMsg string `json:"approveMsg,omitempty"`
	// 审批时间
	FinishedTime string `json:"finishedTime,omitempty" example:"2020-01-01 01:01:01"`
}

// TodoWebDealReq 审批请求体。
type TodoWebDealReq struct {
	// 待办 ID
	ID int `json:"id,omitempty" binding:"gt=0" example:"1"`
	// 是否通过
	IsPass bool `json:"isPass,omitempty" example:"true"`
	// 审批语
	ApproveMessage string `json:"approveMessage,omitempty"`
}

func initTodo() {
	validator.AddTranslationMessage(reflect.TypeFor[TodoWebListReq](), validator.TranslationMessage{
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
	validator.AddTranslationMessage(reflect.TypeFor[TodoWebListDealtReq](), validator.TranslationMessage{
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
		"Types": {
			"unique": {
				i18n.LanguageChinese: "重复地选择了相同待办类型",
				i18n.LanguageEnglish: "repeatedly selecting the same todo type",
			},
			"gt": {
				i18n.LanguageChinese: "选择了非法的待办类型",
				i18n.LanguageEnglish: "illegal todo type selected",
			},
			"max": {
				i18n.LanguageChinese: "选择了非法的待办类型",
				i18n.LanguageEnglish: "illegal todo type selected",
			},
		},
		"Status": {
			"ne": {
				i18n.LanguageChinese: "选择了非法的待办状态",
				i18n.LanguageEnglish: "illegal todo status selected",
			},
			"unique": {
				i18n.LanguageChinese: "重复地选择了相同待办状态",
				i18n.LanguageEnglish: "repeatedly selecting the same todo status",
			},
			"gt": {
				i18n.LanguageChinese: "选择了非法的待办状态",
				i18n.LanguageEnglish: "illegal todo status selected",
			},
			"max": {
				i18n.LanguageChinese: "选择了非法的待办状态",
				i18n.LanguageEnglish: "illegal todo status selected",
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
	validator.AddTranslationMessage(reflect.TypeFor[TodoWebCreateReq](), validator.TranslationMessage{
		"Type": {
			"oneof": {
				i18n.LanguageChinese: "非法的待办类型",
				i18n.LanguageEnglish: "illegal todo type",
			},
		},
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
		"ApplyReason": {
			"required": {
				i18n.LanguageChinese: "请填写申请理由",
				i18n.LanguageEnglish: "please fill in the reason",
			},
			"max": {
				i18n.LanguageChinese: "申请理由过长",
				i18n.LanguageEnglish: "reason too long",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[TodoWebGetDetailReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageChinese: "待办不存在",
				i18n.LanguageEnglish: "todo does not exist",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[TodoWebDealReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageChinese: "待办不存在",
				i18n.LanguageEnglish: "todo does not exist",
			},
		},
	})
}
