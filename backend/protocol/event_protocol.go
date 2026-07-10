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
	"time"

	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/validator"
)

const (
	TimeStepDay = iota + 1
	TimeStepWeek
	TimeStepMonth
)

// EventWebListReq 获取应用事件列表请求体。
type EventWebListReq struct {
	// 应用名或应用 ID
	App string `form:"app" binding:"max=64" example:"MyApp"`
	// 应用平台
	Platform int `form:"platform" binding:"min=0,max=3" example:"1"`
	// 事件来源
	Source int `form:"source" binding:"min=0,max=2" example:"1"`
	// 事件类型
	Type int `form:"type" binding:"min=0,max=28" example:"1"`
	// 操作人
	User string `form:"user" binding:"omitempty,max=32,han|varname" example:"张三"`
	// 事件开始时间
	BeginTime time.Time `form:"beginTime" time_format:"2006-01-02 15:04:05" binding:"time" example:"2020-01-01 00:00:00"`
	// 事件结束时间
	EndTime time.Time `form:"endTime" time_format:"2006-01-02 15:04:05" binding:"time,gtfield=BeginTime" example:"2020-01-01 00:00:00"`
	// 每页条数
	PageSize int `form:"pageSize" binding:"gt=0,max=100" example:"1"`
	// 页数
	PageNumber int `form:"pageNumber" binding:"gt=0" example:"10"`
}

// EventWebListRsp 获取应用事件列表响应体。
type EventWebListRsp struct {
	// 总数
	Count int `json:"count,omitempty" example:"10"`
	// 事件项
	List []*EventWebListItem `json:"list,omitempty"`
}

// EventWebListItem 应用事件信息。
type EventWebListItem struct {
	// 应用 ID
	AppID string `json:"appId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 应用名
	AppName string `json:"appName,omitempty" example:"MyApp"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 事件类型
	Type int `json:"type,omitempty" example:"1"`
	// 事件来源
	Source int `json:"source,omitempty" example:"1"`
	// 操作人
	User string `json:"user,omitempty" example:"zhangsan"`
	// 创建时间
	CreatedTime string `json:"createdTime,omitempty" example:"2020-01-01 00:00:00"`
	// 事件内容
	Content string `json:"content,omitempty"`
}

// EventWebStatisticReq 获取应用事件统计数量请求体。
type EventWebStatisticReq struct {
	// 应用 ID
	AppID string `form:"appId" binding:"omitempty,len=32,alphanum"`
	// 开始时间
	BeginTime time.Time `form:"beginTime" time_format:"2006-01-02" binding:"required"`
	// 结束时间
	EndTime time.Time `form:"endTime" time_format:"2006-01-02" binding:"gtfield=BeginTime"`
	// 时间粒度。
	TimeStep int `form:"timeStep" binding:"gt=0,max=3"`
}

// EventWebStatisticRsp 获取应用事件统计数量响应体。
type EventWebStatisticRsp struct {
	// 数据
	List []*EventWebStatisticItem `json:"list,omitempty"`
}

// EventWebStatisticItem 应用事件数量信息。
type EventWebStatisticItem struct {
	// 开始时间
	BeginTime string `json:"beginTime,omitempty"`
	// 应用新建数量
	CreateAppTimes int `json:"createAppTimes,omitempty"`
	// 上传 Windows 个人 OV 证书数量
	UploadWindowsCertificateTimes int `json:"uploadWindowsCertificateTimes,omitempty"`
	// 应用无效化数量
	InvalidAppTimes int `json:"invalidAppTimes,omitempty"`
	// 申请安卓证书数量
	ApplyAndroidCertificateTimes int `json:"applyAndroidCertificateTimes,omitempty"`
	// 申请苹果描述文件数量
	ApplyAppleProfileTimes int `json:"applyAppleProfileTimes,omitempty"`
	// 申请苹果推送证书数量
	ApplyApplePushCertificateTimes int `json:"applyApplePushCertificateTimes,omitempty"`
	// 上传安卓证书数量
	UploadAndroidCertificateTimes int `json:"uploadAndroidCertificateTimes,omitempty"`
}

func initEvent() {
	validator.AddTranslationMessage(reflect.TypeFor[EventWebListReq](), validator.TranslationMessage{
		"App": {
			"max": {
				i18n.LanguageChinese: "检索应用字符过多",
				i18n.LanguageEnglish: "too many search application characters",
			},
		},
		"Platform": {
			"min": {
				i18n.LanguageChinese: "平台类型非法",
				i18n.LanguageEnglish: "illegal platform type",
			},
			"max": {
				i18n.LanguageChinese: "平台类型非法",
				i18n.LanguageEnglish: "illegal platform type",
			},
		},
		"Source": {
			"min": {
				i18n.LanguageChinese: "事件来源类型非法",
				i18n.LanguageEnglish: "illegal source type of event",
			},
			"max": {
				i18n.LanguageChinese: "事件来源类型非法",
				i18n.LanguageEnglish: "illegal source type of event",
			},
		},
		"Type": {
			"min": {
				i18n.LanguageChinese: "事件来源类型非法",
				i18n.LanguageEnglish: "illegal source type of event",
			},
			"max": {
				i18n.LanguageChinese: "事件来源类型非法",
				i18n.LanguageEnglish: "illegal source type of event",
			},
		},
		"User": {
			"max": {
				i18n.LanguageChinese: "检索的操作人字符数过多",
				i18n.LanguageEnglish: "too many operator characters retrieved",
			},
			"han": {
				i18n.LanguageChinese: "检索的操作人包含非法字符",
				i18n.LanguageEnglish: "retrieval operator contains illegal characters",
			},
			"varname": {
				i18n.LanguageChinese: "检索的操作人包含非法字符",
				i18n.LanguageEnglish: "retrieval operator contains illegal characters",
			},
		},
		"BeginTime": {
			"time": {
				i18n.LanguageChinese: "请选择事件开始时间",
				i18n.LanguageEnglish: "please select the app event start time",
			},
		},
		"PageSize": {
			"gt": {
				i18n.LanguageChinese: "请求每页条数非法",
				i18n.LanguageEnglish: "illegal number of items per page",
			},
			"max": {
				i18n.LanguageChinese: "每页条数过大",
				i18n.LanguageEnglish: "page count is too large",
			},
		},
		"PageNumber": {
			"gt": {
				i18n.LanguageChinese: "请求页数非法",
				i18n.LanguageEnglish: "illegal page number",
			},
		},
		"EndTime": {
			"gtfield": {
				i18n.LanguageChinese: "结束时间不能比开始时间小",
				i18n.LanguageEnglish: "end time cannot be shorter than the start time",
			},
			"time": {
				i18n.LanguageChinese: "请选择事件结束时间",
				i18n.LanguageEnglish: "please select the app event end time",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[EventWebStatisticReq](), validator.TranslationMessage{
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
		"BeginTime": {
			"required": {
				i18n.LanguageChinese: "请选择开始时间",
				i18n.LanguageEnglish: "please choose a start time",
			},
		},
		"EndTime": {
			"gtfield": {
				i18n.LanguageChinese: "结束时间不能比开始时间小",
				i18n.LanguageEnglish: "end time cannot be shorter than the start time",
			},
		},
		"TimeStep": {
			"gt": {
				i18n.LanguageChinese: "错误的时间粒度",
				i18n.LanguageEnglish: "wrong time granularity",
			},
			"max": {
				i18n.LanguageChinese: "错误的时间粒度",
				i18n.LanguageEnglish: "wrong time granularity",
			},
		},
	})
}
