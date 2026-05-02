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

package consts

import (
	"flag"
	"time"

	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
)

// HTTP 字段。
const (
	HTTPHeaderSessionKey  = "csms_session"
	HTTPHeaderSessionUser = "csms_user"
	HTTPPathAppID         = "appId"
)

const (
	// 请求凭证续期时长。
	APIAccountExpirationTime = 30 * 24 * time.Hour
	// 定时任务加锁键的过期时间。
	CronLockerExpiration = time.Hour
	// 证书加密密钥更换时间间隔。
	ASEKeyRotationTime = 30 * 24 * time.Hour
	// 密码长度。
	AESKeyLength = 16
	// 安卓密钥库密码长度。
	KeystorePasswordLength = 6
	// Apple 证书密码长度。
	AppleCertificatePasswordLength = 6
	// Apple 通配符 Bundle ID。
	AppleWildcardBundleID = "*"
	// Apple 服务器地址。
	AppleServerDomain = "https://api.appstoreconnect.apple.com"
	// 请求凭证账号密钥长度。
	ApiAccountSecretLength = 128
	// 用户密码盐长度。
	UserPasswordSaltLength = 128
	// 用户会话长度。
	UserSessionLength = 128
	// 请求凭证 JWT 算法。
	APIAuthorizationAlgorithm = "HS256"
)

var (
	// java 执行程序文件路径。
	JavaBinaryPath string
	// keytool 执行程序文件路径。
	KeytoolBinaryPath string
	// pepk.jar 文件路径。
	PepkJarPath string
	// 运行 pepk.jar 的 Java 运行时文件路径。
	JavaBinaryPathForPepk string
	// 支持的用户头像文件格式。
	SupportUserAvatarFmt = []string{"jpeg", "png", "jpg"}
	// 支持的应用图标文件格式。
	SupportAppLogoFmt = []string{"jpeg", "png", "jpg"}
	// cabextract 文件路径。
	CabextractFilePath string
	// AppleProfileTypes 描述文件类型。
	AppleProfileTypes = map[string]map[string]string{
		model.ApplePlatformUniversalDescription: {
			model.AppleProfileTypeIOSAppDevelopment: "iOS 开发描述文件",
			model.AppleProfileTypeMacAppDevelopment: "macOS 开发描述文件",
			model.AppleProfileTypeIOSAppAdhoc:       "iOS Adhoc 描述文件",
			model.AppleProfileTypeIOSAppStore:       "iOS 发布描述文件",
			model.AppleProfileTypeMacAppStore:       "macOS 发布描述文件",
		},
		model.ApplePlatformIOSDescription: {
			model.AppleProfileTypeIOSAppDevelopment: "iOS 开发描述文件",
			model.AppleProfileTypeIOSAppAdhoc:       "iOS Adhoc 描述文件",
			model.AppleProfileTypeIOSAppStore:       "iOS 发布描述文件",
		},
		model.ApplePlatformMacOSDescription: {
			model.AppleProfileTypeMacAppDevelopment: "macOS 开发描述文件",
			model.AppleProfileTypeMacAppStore:       "macOS 发布描述文件",
		},
	}
)

// AddCommandFlag 增加程序命令参数。
func AddCommandFlag() {
	flag.StringVar(&JavaBinaryPath, cc.CommandFlagJavaBinaryPath, "java", "Java binary path")
	flag.StringVar(&CabextractFilePath, cc.CommandFlagCabextractFilePath, "./extract", "cabextract file path")
	flag.StringVar(&KeytoolBinaryPath, cc.CommandFlagKeytoolBinaryPath, "keytool", "keytool binary path")
	flag.StringVar(&PepkJarPath, cc.CommandFlagPepkJarPath, "./pepk.jar", "pepk.jar file path")
	flag.StringVar(&JavaBinaryPathForPepk, cc.CommandFlagJavaBinaryPathForPepk, "java",
		"java file path for running pepk.jar")
}
