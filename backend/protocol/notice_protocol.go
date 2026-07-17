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

// NoticeWebLastRsp 获取通知响应体。
type NoticeWebLastRsp struct {
	// 通知内容
	Message string `json:"message,omitempty"`
}

// NoticeWebAddReq 添加通知请求体。
type NoticeWebAddReq struct {
	// 通知内容
	Message string `json:"message,omitempty" binding:"required,max=16777215" example:"any thing"`
	// 生效时间
	ActivatedTime Time `json:"activatedTime" binding:"required" example:"2026-07-17 12:00:00"`
	// 过期时间
	ExpiredTime Time `json:"expiredTime"  binding:"required,gtfield=ActivatedTime" example:"2026-07-18 12:00:00"`
}

// 通知列表响应体。
type NoticeWebListRsp struct {
	// 通知列表
	List []*NoticeWebListItem `json:"list,omitempty"`
}

// 通知信息。
type NoticeWebListItem struct {
	// 通知主码
	ID int `json:"id,omitempty"`
	// 通知内容
	Message string `json:"message,omitempty"`
	// 添加人
	User string `json:"user,omitempty"`
	// 生效时间
	ActivatedTime string `json:"activatedTime,omitempty"`
	// 失效时间
	ExpiredTime string `json:"expiredTime,omitempty"`
	// 添加时间
	CreatedTime string `json:"createdTime,omitempty"`
}

// 删除通知请求体。
type NoticeWebRemoveReq struct {
	// 通知主码
	ID int `form:"id" binding:"gt=0" example:"1"`
}

func initNotice() {
	validator.AddTranslationMessage(reflect.TypeFor[NoticeWebAddReq](), validator.TranslationMessage{
		"Message": {
			"required": {
				i18n.LanguageChinese: "请填写通知内容",
				i18n.LanguageEnglish: "please fill in notice content",
			},
			"max": {
				i18n.LanguageChinese: "通知内容字符数过多",
				i18n.LanguageEnglish: "notice content with too many characters",
			},
		},
		"ActivatedTime": {
			"required": {
				i18n.LanguageChinese: "请填写生效时间",
				i18n.LanguageEnglish: "please fill in the effective time",
			},
		},
		"ExpiredTime": {
			"required": {
				i18n.LanguageChinese: "请填写过期时间",
				i18n.LanguageEnglish: "please fill in the expiration date",
			},
			"gtfield": {
				i18n.LanguageChinese: "过期时间不能小于生效时间",
				i18n.LanguageEnglish: "the expiration time cannot be less than the effective time",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[NoticeWebRemoveReq](), validator.TranslationMessage{
		"ID": {
			"gt": {
				i18n.LanguageChinese: "错误的通知 ID",
				i18n.LanguageEnglish: "wrong notification ID",
			},
		},
	})
}
