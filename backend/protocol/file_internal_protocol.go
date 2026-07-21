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
	"io"
	"reflect"

	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/validator"
)

const (
	HTTPPathFileInternalDownload        = "/internal/file/download"
	HTTPPathFileInternalUpload          = "/internal/file/upload"
	HTTPPathFileAPIInitializeUploadFile = "/api/file/initial"
	HTTPPathFileAPIUploadFilePart       = "/api/file/uploadPart"
	HTTPPathFileAPIMergeFilePart        = "/api/file/mergeParts"
	HTTPPathFileAPIDownload             = "/api/file/download"
)

// FileInternalDownloadReq 下载请求体。
type FileInternalDownloadReq struct {
	// 文件 ID
	FileID string `form:"fileId" binding:"alphanum,len=38" example:"2026014ef83c03e2ce4f1f94c11168d1acd087"`
}

// FileInternalUploadReq 文件上传请求体。
type FileInternalUploadReq struct {
	// 文件类型
	Type int `form:"type" binding:"gt=0,max=7"`
	// 文件名
	Name string `form:"name" binding:"required,max=256"`
	// 应用 ID
	AppID int `form:"appId" binding:"gt=0"`
	Body  io.ReadCloser
	Size  int64
}

func initFileInternalProtocol() {
	validator.AddTranslationMessage(reflect.TypeFor[FileInternalDownloadReq](), validator.TranslationMessage{
		"FileID": {
			"alphanum": {
				i18n.LanguageEnglish: "no download file information available",
			},
			"len": {
				i18n.LanguageEnglish: "no download file information available",
			},
		},
	})
	validator.AddTranslationMessage(reflect.TypeFor[FileInternalUploadReq](), validator.TranslationMessage{
		"Type": {
			"gt": {
				i18n.LanguageEnglish: "file type to be uploaded is incorrect",
			},
			"max": {
				i18n.LanguageEnglish: "file type to be uploaded is incorrect",
			},
		},
		"Name": {
			"required": {
				i18n.LanguageEnglish: "name of the file to be uploaded is missing",
			},
			"max": {
				i18n.LanguageEnglish: "name of the file to be uploaded has too many characters",
			},
		},
		"AppID": {
			"gt": {
				i18n.LanguageEnglish: "app id is missing or invalid",
			},
		},
	})
}
