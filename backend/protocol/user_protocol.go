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

// UserWebRegisterReq 注册请求体。
type UserWebRegisterReq struct {
	// 用户头像
	Avatar *multipart.FileHeader `form:"avatar" binding:"file"`
	// 中文名
	NameZh string `form:"nameZh" binding:"min=2,max=16,han"`
	// 英文名
	NameEn string `form:"nameEn" binding:"min=6,max=32,varname"`
	// 密码
	Password string `form:"password" binding:"min=6,utf8string"`
	// 二次确认密码
	PasswordConfirmation string `form:"passwordConfirmation" binding:"eqfield=Password"`
	// 部门
	Department string `form:"department" binding:"required,max=1024,hanasciiprint"`
}

// UserWebLoginReq 登录请求体。
type UserWebLoginReq struct {
	// 英文名
	NameEn string `json:"nameEn" binding:"min=6|eq=admin,max=32,varname" example:"zhangsan"`
	// 密码
	Password string `json:"password" binding:"utf8string,min=6|eq=admin" example:"123456"`
}

// UserWebGetInformationRsp 获取用户信息响应体。
type UserWebGetInformationRsp struct {
	// 头像文件 ID
	AvatarFileID string `json:"avatarFileId,omitempty" example:"2026013885c6abe6e642a6b32965e6f8afebcb"`
	// 中文名
	NameZh string `json:"nameZh,omitempty" example:"张三"`
	// 英文名
	NameEn string `json:"nameEn,omitempty" example:"zhangsan"`
	// 部门
	Department string `json:"department,omitempty" example:"/技术部/研发"`
}

// UserWebUpdateReq 更新个人信息请求体。
type UserWebUpdateReq struct {
	// 用户头像文件 ID
	AvatarFileID string `json:"avatarFileId,omitempty" binding:"len=38,alphanum" example:"2026013885c6abe6e642a6b32965e6f8afebcb"`
	// 中文名
	NameZh string `json:"nameZh,omitempty" binding:"min=2,max=16,han" example:"张三"`
	// 密码
	Password string `json:"password,omitempty" binding:"min=6,utf8string" example:"123456"`
	// 二次确认密码
	PasswordConfirmation string `json:"passwordConfirmation,omitempty" binding:"eqfield=Password" example:"123456"`
	// 部门
	Department string `json:"department,omitempty" binding:"required,max=1024,hanasciiprint" example:"/技术部/研发"`
}

// UserWebSearchReq 搜索用户请求体。
type UserWebSearchReq struct {
	// 用户英文名
	Name string `form:"name" binding:"min=1,max=32,hanasciiprint" example:"zhangsan"`
}

// UserWebSearchRsp 搜索用户响应体。
type UserWebSearchRsp struct {
	// 匹配的用户
	Users map[string]string `json:"users,omitempty" example:"zhangsan:张三"`
}

func initUser() {
	validator.AddTranslationMessage(reflect.TypeFor[UserWebRegisterReq](), validator.TranslationMessage{
		"Avatar": {
			"file": {
				i18n.LanguageChinese: "请上传用户头像",
				i18n.LanguageEnglish: "please upload a user avatar",
			},
		},
		"NameZh": {
			"min": {
				i18n.LanguageChinese: "中文名字符数不能少于两位",
				i18n.LanguageEnglish: "number of chinese name characters cannot be less than two",
			},
			"max": {
				i18n.LanguageChinese: "中文名字符数不能大于十六位",
				i18n.LanguageEnglish: "number of chinese name characters cannot exceed sixteen",
			},
			"han": {
				i18n.LanguageChinese: "中文名必须全是汉字",
				i18n.LanguageEnglish: "chinese names must be all in chinese characters",
			},
		},
		"NameEn": {
			"min": {
				i18n.LanguageChinese: "英文名字符数不能少于六位",
				i18n.LanguageEnglish: "number of english name characters cannot be less than six",
			},
			"max": {
				i18n.LanguageChinese: "英文名字符数不能大于三十二位",
				i18n.LanguageEnglish: "number of english name characters cannot be exceed thirty-two",
			},
			"varname": {
				i18n.LanguageChinese: "英文名必须由字母、数字和下划线组成，且不能以数字开头",
				i18n.LanguageEnglish: "english names must consist of letters, numbers, and underscores, and cannot start with a number",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageChinese: "密码字符数不能少于六位",
				i18n.LanguageEnglish: "number of password characters cannot be less than six",
			},
			"utf8string": {
				i18n.LanguageChinese: "密码必须是 utf-8 字符",
				i18n.LanguageEnglish: "password must be uft-8 characters",
			},
		},
		"PasswordConfirmation": {
			"eqfield": {
				i18n.LanguageChinese: "两次输入的密码不相同",
				i18n.LanguageEnglish: "passwords entered twice are not the same",
			},
		},
		"Department": {
			"required": {
				i18n.LanguageChinese: "请填写用户部门信息",
				i18n.LanguageEnglish: "please fill in user department information",
			},
			"max": {
				i18n.LanguageChinese: "用户部门字符数过多",
				i18n.LanguageEnglish: "too many characters in the user department",
			},
			"hanasciiprint": {
				i18n.LanguageChinese: "部门信息包含非法字符",
				i18n.LanguageEnglish: "user department information contains illegal characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[UserWebLoginReq](), validator.TranslationMessage{
		"NameEn": {
			"min=6|eq=admin": {
				i18n.LanguageChinese: "用户不存在",
				i18n.LanguageEnglish: "user does not exist",
			},
			"max": {
				i18n.LanguageChinese: "用户不存在",
				i18n.LanguageEnglish: "user does not exist",
			},
			"varname": {
				i18n.LanguageChinese: "用户不存在",
				i18n.LanguageEnglish: "user does not exist",
			},
		},
		"Password": {
			"min=6|eq=admin": {
				i18n.LanguageChinese: "密码错误",
				i18n.LanguageEnglish: "password error",
			},
			"utf8": {
				i18n.LanguageChinese: "密码错误",
				i18n.LanguageEnglish: "password error",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[UserWebUpdateReq](), validator.TranslationMessage{
		"AvatarFileID": {
			"len": {
				i18n.LanguageChinese: "用户头像文件不存在",
				i18n.LanguageEnglish: "user avatar file id does not exist",
			},
			"alphanum": {
				i18n.LanguageChinese: "用户头像文件不存在",
				i18n.LanguageEnglish: "user avatar file id does not exist",
			},
		},
		"NameZh": {
			"min": {
				i18n.LanguageChinese: "中文名字符数不能少于两位",
				i18n.LanguageEnglish: "number of chinese name characters cannot be less than two",
			},
			"max": {
				i18n.LanguageChinese: "中文名字符数不能大于十六位",
				i18n.LanguageEnglish: "number of chinese name characters cannot exceed sixteen",
			},
			"han": {
				i18n.LanguageChinese: "中文名必须全是汉字",
				i18n.LanguageEnglish: "chinese names must be all in chinese characters",
			},
		},
		"Password": {
			"min": {
				i18n.LanguageChinese: "密码字符数不能少于六位",
				i18n.LanguageEnglish: "number of password characters cannot be less than six",
			},
			"utf8": {
				i18n.LanguageChinese: "密码必须是 utf-8 字符",
				i18n.LanguageEnglish: "password must be utf-8 characters",
			},
		},
		"PasswordConfirmation": {
			"eqfield": {
				i18n.LanguageChinese: "两次输入的密码不相同",
				i18n.LanguageEnglish: "passwords entered twice are not the same",
			},
		},
		"Department": {
			"required": {
				i18n.LanguageChinese: "请填写部门信息",
				i18n.LanguageEnglish: "please fill in department information",
			},
			"max": {
				i18n.LanguageChinese: "部门字符数过多",
				i18n.LanguageEnglish: "too many characters in the department",
			},
			"hanasciiprint": {
				i18n.LanguageChinese: "部门信息包含非法字符",
				i18n.LanguageEnglish: "department information contains illegal characters",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[UserWebSearchReq](), validator.TranslationMessage{
		"Name": {
			"min": {
				i18n.LanguageChinese: "英文名字符数不能少于一位",
				i18n.LanguageEnglish: "number of english name characters cannot be less than one",
			},
			"max": {
				i18n.LanguageChinese: "英文名字符数不能大于三十二位",
				i18n.LanguageEnglish: "number of english name characters cannot be exceed thirty-two",
			},
			"varname": {
				i18n.LanguageChinese: "英文名必须由字母、数字和下划线组成，且不能以数字开头",
				i18n.LanguageEnglish: "english names must consist of letters, numbers, and underscores, and cannot start with a number",
			},
		},
	})
}
