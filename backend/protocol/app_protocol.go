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
	"mime/multipart"
	"reflect"

	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/validator"
)

// AppWebRegisterReq 注册请求体。
type AppWebRegisterReq struct {
	// 应用名
	Name string `form:"name" binding:"required,max=64"`
	// 应用图标
	Logo *multipart.FileHeader `form:"logo" binding:"file"`
	// 平台
	Platform int `form:"platform" binding:"oneof=1 2 3"`
	// 管理员
	Admins []string `form:"admins" binding:"omitempty,unique,dive,min=1,max=32,varname"`
	// 成员
	Members []string `form:"members" binding:"omitempty,unique,dive,min=1,max=32,varname"`
}

// AppWebSearchReq 查询请求体。
type AppWebSearchReq struct {
	// 应用名
	Name string `json:"name,omitempty" binding:"max=64" example:"MyApp"`
	// 平台
	Platform []int `json:"platform,omitempty" binding:"omitempty,unique,dive,oneof=1 2 3" example:"1"`
	// 状态
	Status []int `json:"status,omitempty" binding:"omitempty,unique,dive,oneof=1 2 3 4" example:"1"`
}

// AppWebSearchRsp 查询响应体。
type AppWebSearchRsp struct {
	List []*AppWebSearchItem `json:"list,omitempty"`
}

// AppWebSearchItem 应用信息。
type AppWebSearchItem struct {
	// 应用 ID
	ID string `json:"id,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 应用名
	Name string `json:"name,omitempty" example:"MyApp"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 状态
	Status int `json:"status,omitempty" example:"1"`
}

// AppWebUpdateReq 更新请求体。
type AppWebUpdateReq struct {
	// 应用名
	Name string `json:"name,omitempty" binding:"required,max=64" example:"MyApp"`
	// 应用图标文件 ID
	LogoFileID string `json:"logoFileId,omitempty" binding:"len=38,alphanum" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 管理员
	Admins []string `json:"admins,omitempty" binding:"min=1,unique,dive,required,max=32,varname" example:"zhangsan"`
	// 成员
	Members []string `json:"members,omitempty" binding:"omitempty,unique,dive,min=1,max=32,varname" example:"zhangsan"`
}

// AppWebGetInformationRsp 获取应用信息响应体。
type AppWebGetInformationRsp struct {
	// 应用 ID
	AppID string `json:"appId,omitempty" example:"4ef83c03e2ce4f1f94c11168d1acd087"`
	// 应用名
	Name string `json:"name,omitempty" example:"MyApp"`
	// 应用图标文件 ID
	LogoFileID string `json:"logoFileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 应用平台
	Platform int `json:"platform,omitempty" example:"1"`
	// 应用管理员
	Admins map[string]string `json:"admins,omitempty" example:"zhangsan:张三"`
	// 应用成员
	Members map[string]string `json:"members,omitempty" example:"zhangsan:张三"`
	// 应用状态
	Status int `json:"status,omitempty" example:"1"`
}

// AppWebCountRsp 获取用户具有权限的应用个数响应体。
type AppWebCountRsp struct {
	// 总数
	Count int `json:"count,omitempty" example:"1"`
}

func initApp() {
	validator.AddTranslationMessage(reflect.TypeFor[AppWebRegisterReq](), validator.TranslationMessage{
		"Name": {
			"required": {
				i18n.LanguageChinese: "请填写应用名",
				i18n.LanguageEnglish: "please fill in the app name",
			},
			"max": {
				i18n.LanguageChinese: "应用名字符数过多",
				i18n.LanguageEnglish: "too many name characters in the app name",
			},
		},
		"Logo": {
			"file": {
				i18n.LanguageChinese: "请上传应用图标",
				i18n.LanguageEnglish: "please upload the app logo",
			},
		},
		"Platform": {
			"oneof": {
				i18n.LanguageChinese: "错误的应用平台类型",
				i18n.LanguageEnglish: "wrong app platform type",
			},
		},
		"Admins": {
			"unique": {
				i18n.LanguageChinese: "重复选择了应用管理员",
				i18n.LanguageEnglish: "repeatedly selected app administrator",
			},
			"min": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
		},
		"Members": {
			"unique": {
				i18n.LanguageChinese: "重复选择了应用成员",
				i18n.LanguageEnglish: "repeatedly selected app member",
			},
			"min": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppWebSearchReq](), validator.TranslationMessage{
		"Name": {
			"max": {
				i18n.LanguageChinese: "应用不存在于系统",
				i18n.LanguageEnglish: "app does not exist in the system",
			},
		},
		"Platform": {
			"oneof": {
				i18n.LanguageChinese: "应用平台类型非法",
				i18n.LanguageEnglish: "illegal app platform type",
			},
			"unique": {
				i18n.LanguageChinese: "应用平台类型重复",
				i18n.LanguageEnglish: "duplicate app platform types",
			},
		},
		"Status": {
			"oneof": {
				i18n.LanguageChinese: "应用状态类型非法",
				i18n.LanguageEnglish: "illegal app status type",
			},
			"unique": {
				i18n.LanguageChinese: "应用状态类型重复",
				i18n.LanguageEnglish: "duplicate app state types",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[AppWebUpdateReq](), validator.TranslationMessage{
		"Name": {
			"required": {
				i18n.LanguageChinese: "请填写应用名",
				i18n.LanguageEnglish: "please fill in the app name",
			},
			"max": {
				i18n.LanguageChinese: "应用名字符数过多",
				i18n.LanguageEnglish: "too many name characters in the app name",
			},
		},
		"LogoFileID": {
			"len": {
				i18n.LanguageChinese: "应用图标文件不存在于系统",
				i18n.LanguageEnglish: "app logo file id does not exist in the system",
			},
			"alphanum": {
				i18n.LanguageChinese: "应用图标文件不存在于系统",
				i18n.LanguageEnglish: "app logo file id does not exist in the system",
			},
		},
		"Admins": {
			"unique": {
				i18n.LanguageChinese: "重复选择了应用管理员",
				i18n.LanguageEnglish: "repeatedly selected app administrator",
			},
			"min": {
				i18n.LanguageChinese: "至少选择一个应用管理员",
				i18n.LanguageEnglish: "at least one app administrator is required",
			},
			"required": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "选择了不存在于系统的管理员",
				i18n.LanguageEnglish: "selected an administrator who does not exist in the system",
			},
		},
		"Members": {
			"unique": {
				i18n.LanguageChinese: "重复选择了应用成员",
				i18n.LanguageEnglish: "repeatedly selected app member",
			},
			"min": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
			"max": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
			"varname": {
				i18n.LanguageChinese: "选择了不存在于系统的成员",
				i18n.LanguageEnglish: "selected a member who does not exist in the system",
			},
		},
	})
}
