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

package main

import "time"

// 程序退出码。
const (
	ExitCodeParseYamlError = 1 + iota
	ExitCodeUploadFileError
	ExitCodeInvalidConfigError
	ExitCodeSubmitJobError
	ExitCodeListenJobError
	ExitCodeDownloadFileError
)

// 签名任务类型。
const (
	JobTypeWindows = "windows"
	JobTypeWHQL    = "whql"
	JobTypeAndroid = "android"
	JobTypeApple   = "apple"
)

// 安卓签名任务类型。
const (
	AndroidTypeAPK   = "apk"
	AndroidTypeAAB   = "aab"
	AndroidTypePatch = "patch"
)

// Windows 任务类型。
const (
	WindowsSigningTypePE               = "pe"
	WindowsSigningTypeAttestation      = "attestation"
	WindowsSigningTypePEAndAttestation = "peAndAttestation"
)

// WHQL 任务类型。
const (
	WHQLTypeHLK        = "hlk"
	WHQLTypeTestAndHLK = "testAndHlk"
)

// 版本号。
const (
	Major = 1
	Minor = 0
	Patch = 0
)

const (
	UploadFilePartSize         = 8 * 1024 * 1024
	AccessTokenExpiredDuration = 2 * time.Hour
	HTTPRetryTimes             = 3
	LogFilePath                = "auto_signer.log"
)

var ServerAddress = "localhost"
