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

// FileAPIDownloadReq 下载请求体。
type FileAPIDownloadReq struct {
	// 文件 ID
	FileID string `form:"fileId" binding:"alphanum,len=38" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// FileAPIInitialReq 初始化上传请求体。
type FileAPIInitialReq struct {
	// 文件名
	Name string `json:"name,omitempty" binding:"required,max=256" example:"avatar.png"`
	// 文件大小
	Size int64 `json:"size,omitempty" binding:"gt=0" example:"1024"`
	// 文件 MD5 值
	MD5 string `json:"md5,omitempty" binding:"md5" example:"73b8928185abcc728c7fedb882ba531c"`
	// 文件类型
	Type int `json:"type,omitempty" binding:"gt=0,max=7" example:"1"`
}

// FileAPIInitialRsp 初始化上传响应体。
type FileAPIInitialRsp struct {
	// 文件是否存在
	Exist bool `json:"exist,omitempty" example:"true"`
	// 文件 ID
	FileID string `json:"fileId,omitempty" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// FileAPIUploadPartReq 上传分片请求体。
type FileAPIUploadPartReq struct {
	// 文件分片
	Chunk *multipart.FileHeader `form:"chunk" binding:"file"`
	// 文件 ID
	FileID string `form:"fileId" binding:"alphanum,len=38" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
	// 分片序号
	ChunkNumber int `form:"chunkNumber" binding:"gt=0" example:"1"`
}

// FileAPIMergePartsReq 合并分片请求体。
type FileAPIMergePartsReq struct {
	// 文件 ID
	FileID string `form:"fileId" binding:"alphanum,len=38" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

func initFileOpenProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[FileAPIDownloadReq](), validator.TranslationMessage{
		"FileID": {
			"alphanum": {
				i18n.LanguageChinese: "无下载文件信息",
				i18n.LanguageEnglish: "no download file information available",
			},
			"len": {
				i18n.LanguageChinese: "无下载文件信息",
				i18n.LanguageEnglish: "no download file information available",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[FileAPIInitialReq](), validator.TranslationMessage{
		"Name": {
			"required": {
				i18n.LanguageChinese: "待上传的文件的名称缺失",
				i18n.LanguageEnglish: "name of the file to be uploaded is missing",
			},
			"max": {
				i18n.LanguageChinese: "待上传的文件的名称字符数过多",
				i18n.LanguageEnglish: "name of the file to be uploaded has too many characters",
			},
		},
		"Size": {
			"gt": {
				i18n.LanguageChinese: "待上传的文件的大小错误",
				i18n.LanguageEnglish: "size of the file to be uploaded is incorrect",
			},
		},
		"MD5": {
			"md5": {
				i18n.LanguageChinese: "待上传的文件的 md5 错误",
				i18n.LanguageEnglish: "md5 of the file to be uploaded is incorrect",
			},
		},
		"Type": {
			"gt": {
				i18n.LanguageChinese: "待上传的文件类型错误",
				i18n.LanguageEnglish: "file type to be uploaded is incorrect",
			},
			"max": {
				i18n.LanguageChinese: "待上传的文件类型错误",
				i18n.LanguageEnglish: "file type to be uploaded is incorrect",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[FileAPIUploadPartReq](), validator.TranslationMessage{
		"Chunk": {
			"file": {
				i18n.LanguageChinese: "待上传的分片数据缺失",
				i18n.LanguageEnglish: "missing file chunk data to be uploaded",
			},
		},
		"FileID": {
			"alphanum": {
				i18n.LanguageChinese: "不存在待上传文件",
				i18n.LanguageEnglish: "there is no file to be uploaded",
			},
			"len": {
				i18n.LanguageChinese: "不存在待上传文件",
				i18n.LanguageEnglish: "there is no file to be uploaded",
			},
		},
		"ChunkNumber": {
			"gt": {
				i18n.LanguageChinese: "非法的文件分片序号",
				i18n.LanguageEnglish: "illegal file chunk number",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[FileAPIMergePartsReq](), validator.TranslationMessage{
		"FileID": {
			"alphanum": {
				i18n.LanguageChinese: "不存在待上传文件",
				i18n.LanguageEnglish: "there is no file to be uploaded",
			},
			"len": {
				i18n.LanguageChinese: "不存在待上传文件",
				i18n.LanguageEnglish: "there is no file to be uploaded",
			},
		},
	})
}
