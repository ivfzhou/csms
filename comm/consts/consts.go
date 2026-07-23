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
	"runtime"
	"strings"
	"sync"
)

// 程序退出码。
const (
	ExitCodeInitialConfigError = 1 + iota
	ExitCodeInitialMySQLConnectionError
	ExitCodeInitialMQConnectionError
	ExitCodeInitialRedisConnectionError
	ExitCodeInitialTusdConnectionError
	ExitCodeInitialValidatorError
	ExitCodeLocalIPNotFound
	ExitCodeLoadTimezoneError
	ExitCodeHTTPListenError
	ExitCodeLoadRedisScriptError
	ExitCodeStartCronJobError
	ExitCodeInitialMessageFileError
	ExitCodeHLKManagerNoMode
	ExitCodeHLKManagerInvalidSystem
	ExitCodeSignServerInvalidMode
	ExitCodeDeclareQueueError
)

// 文件扩展名。
const (
	ExtensionAPK  = ".apk"
	ExtensionAAB  = ".aab"
	ExtensionCAB  = ".cab"
	ExtensionSYS  = ".sys"
	ExtensionHLKX = ".hlkx"
	ExtensionZIP  = ".zip"
	ExtensionINF  = ".inf"
	ExtensionPFX  = ".pfx"
	ExtensionJKS  = ".jks"
	ExtensionDDF  = ".ddf"
	ExtensionCAT  = ".cat"
)

// 服务名。
const (
	ServiceNameBackend  = "backend"
	ServiceNameFastlane = "fastlane_proxy"
	ServiceNameHLK      = "hlk_manager"
	ServiceNameSigner   = "sign_server"
)

// 命令行参数。
const (
	// 配置文件路径参数。
	CommandFlagConfigurationFilePath = "config"
	// 服务运行使用 IP。
	CommandFlagLocalIP = "localIP"
	// 测试模式。
	CommandFlagLocalTestMode = "localTestMode"
	// 不限制请求速度。
	CommandFlagSkipRateLimit = "skipRateLimit"
	// i18n 语言配置文件夹。
	CommandFlagMessageFilesDirectory = "messageFilesDirectory"
	// Java 文件路径。
	CommandFlagJavaBinaryPath = "javaBinaryPath"
	// keytool 文件路径。
	CommandFlagKeytoolBinaryPath = "keytoolBinaryPath"
	// 谷歌 pepk.jar 文件路径。
	CommandFlagPepkJarPath = "pepkJarPath"
	// 运行 pepk.jar 的 java 解释器文件路径。
	CommandFlagJavaBinaryPathForPepk = "javaBinaryPathForPepk"
	// certmgr.exe 文件路径。
	CommandFlagCertmgrFilePath = "certmgrFilePath"
	// makecert.exe 文件路径。
	CommandFlagMakecertFilePath = "makecertFilePath"
	// signtool.exe 文件路径。
	CommandFlagSigntoolFilePath = "signtoolFilePath"
	// hlk_tool.ps1 文件路径。
	CommandFlagHLKToolScriptFilePath = "hlkToolScriptFilePath"
	// winevsigner.exe 文件路径。
	CommandFlagWinevsignerFilePath = "winevsignerFilePath"
	// zsign 文件路径。
	CommandFlagZsignFilePath = "zsignFilePath"
	// apksigner 文件路径。
	CommandFlagApksignerFilePath = "apksignerFilePath"
	// jarsigner 文件路径。
	CommandFlagJarsignerFilePath = "jarsignerFilePath"
	// java home 目录。
	CommandFlagJavaHomeFilePath = "javaHomeFilePath"
	// inf2Cat.exe 文件路径。
	CommandFlagInf2CatFilePath = "inf2CatFilePath"
	// makecab.exe 文件路径。
	CommandFlagMakecabFilePath = "makecabFilePath"
	// cabextract 文件路径。
	CommandFlagCabextractFilePath = "cabextractFilePath"
	// fastlane 文件路径。
	CommandFlagFastlaneBinaryPath = "fastlaneBinaryPath"
)

// HTTP。
const (
	// 系统链路标识 HTTP 请求头字段名。
	HTTPHeaderRequestID = "X-Csms-Request-Id"
	// 客户端 IP。
	HTTPHeaderIP = "X-Real-IP"
	// 语言参数名。
	HTTPQueryLanguage = "language"
)

// 通用常量。
const (
	// 机器位宽。
	PointerSize = 4 << (^uintptr(0) >> 63)
	// 系统名称。
	SystemName = "csms"
	// 文件夹权限。
	DirectoryMode = 0700
	// 文件权限。
	FileMode = 0600
	// 时间格式。
	TimeFormat = "2006-01-02 15:04:05"
)

// MQ 消息头。
const (
	MQHeaderSendTime = "Send-Time"
)

// 通用变量。
var (
	// Bundle ID 所有能力项。
	AppleBundleIDCapabilities = map[string][2]string{
		"wireless accessory configuration": {"WIRELESS_ACCESSORY_CONFIGURATION", "wireless_accessory"},
		"wallet":                           {"WALLET", "wallet"},
		"user management":                  {"USER_MANAGEMENT", "user_management"},
		"time sensitive notifications":     {"USERNOTIFICATIONS_TIMESENSITIVE", "time_sensitive_notifications"},
		"system extension":                 {"SYSTEM_EXTENSION_INSTALL", "system_extension"},
		"sirikit":                          {"SIRIKIT", "siri_kit"},
		"sign in with apple":               {"APPLE_ID_AUTH", "sign_in_with_apple"},
		"push notifications":               {"PUSH_NOTIFICATIONS", "push_notification"},
		"personal vpn":                     {"PERSONAL_VPN", "personal_vpn"},
		"nfc tag reading":                  {"NFC_TAG_READING", "nfc_tag_reading"},
		"network extensions":               {"NETWORK_EXTENSIONS", "network_extension"},
		"multipath":                        {"MULTIPATH", "multipath"},
		"maps":                             {"MAPS", "maps"},
		"mdm managed associated domains":   {"MDM_MANAGED_ASSOCIATED_DOMAINS", "managed_associated_domains"},
		"low latency hls":                  {"COREMEDIA_HLS_LOW_LATENCY", "low_latency_hls"},
		"in app purchase":                  {"IN_APP_PURCHASE", "in_app_purchase"},
		"inter app autio":                  {"INTER_APP_AUDIO", "inter_app_audio"},
		"icloud":                           {"ICLOUD", "icloud"},
		"hot spot":                         {"HOT_SPOT", "hotspot"},
		"homekit":                          {"HOMEKIT", "home_kit"},
		"hls interstitial previews":        {"HLS_INTERSTITIAL_PREVIEW", "hls_interstitial_preview"},
		"healthKit estimate recalibration": {"HEALTHKIT_RECALIBRATE_ESTIMATES",
			"health_kit_estimate_recalibration"},
		"healthkit":                      {"HEALTHKIT", "health_kit"},
		"group activities":               {"GROUP_ACTIVITIES", "group_activities"},
		"game center":                    {"GAME_CENTER", "game_center"},
		"fonts":                          {"FONT_INSTALLATION", "fonts"},
		"file provider testingmode":      {"FILEPROVIDER_TESTINGMODE", "file_provider_testing_mode"},
		"family control":                 {"FAMILY_CONTROLS", "family_controls"},
		"extended virtual address space": {"EXTENDED_VIRTUAL_ADDRESSING", "extended_virtual_address_space"},
		"driverkit transport hid":        {"DRIVERKIT_TRANSPORT_HID_PUB", "driver_kit_transport_hid"},
		"driverkit family hid eventservice": {"DRIVERKIT_FAMILY_HIDEVENTSERVICE_PUB",
			"driver_kit_hid_event_service"},
		"driverkit family serial": {"DRIVERKIT_FAMILY_SERIAL_PUB",
			"driver_kit_family_serial"},
		"driverkit family networking": {"DRIVERKIT_FAMILY_NETWORKING_PUB",
			"driver_kit_family_networking"},
		"driverkit family hid device": {"DRIVERKIT_FAMILY_HIDDEVICE_PUB",
			"driver_kit_family_hid_device"},
		"driverkit":               {"DRIVERKIT_PUBLIC", "driver_kit"},
		"data protestion":         {"DATA_PROTECTION", "data_protection"},
		"custom network protocol": {"NETWORK_CUSTOM_PROTOCOL", "custom_network_protocol"},
		"communication notifications": {"USERNOTIFICATIONS_COMMUNICATION",
			"communication_notifications"},
		"classkit":                                          {"CLASSKIT", "class_kit"},
		"autofill credential provider":                      {"AUTOFILL_CREDENTIAL_PROVIDER", "auto_fill_credential"},
		"associated domains":                                {"ASSOCIATED_DOMAINS", "associated_domains"},
		"app groups":                                        {"APP_GROUPS", "app_group"},
		"app attest":                                        {"APP_ATTEST", "app_attest"},
		"apple pay payment processing":                      {"APPLE_PAY", "apple_pay"},
		"access wifi information":                           {"ACCESS_WIFI_INFORMATION", "access_wifi"},
		"driverkit family scsicontroller":                   {"DRIVERKIT_FAMILY_SCSICONTROLLER_PUB", ""},
		"weatherKit":                                        {"WEATHERKIT", ""},
		"driverKit USB transport":                           {"DRIVERKIT_USBTRANSPORT_PUB", ""},
		"increased memory limit":                            {"INCREASED_MEMORY_LIMIT", ""},
		"driverkit communicates with drivers":               {"DRIVERKIT_COMMUNICATESWITHDRIVERS", ""},
		"media device discovery":                            {"MEDIA_DEVICE_DISCOVERY", ""},
		"driverKit allow third party userclients":           {"DRIVERKIT_ALLOWTHIRDPARTY_USERCLIENTS", ""},
		"on demand install capable for app clip extensions": {"ONDEMANDINSTALL_EXTENSIONS", ""},
		"tap to present ID on iPhone (Display Only)":        {"TAP_TO_DISPLAY_ID", ""},
		"shallow depth and pressure":                        {"SHALLOW_DEPTH_PRESSURE", ""},
		"matter allow setup payload":                        {"MATTER_ALLOW_SETUP_PAYLOAD", ""},
		"messages collaboration":                            {"MESSAGES_COLLABORATION", ""},
		"vmnet":                                             {"VMNET", ""},
		"sensitive content analysis":                        {"SENSITIVE_CONTENT_ANALYSIS", ""},
		"mac catalyst":                                      {"MARZIPAN", ""},
		"5G network slicing":                                {"NETWORK_SLICING", ""},
	}

	unitTestMode       bool
	setUniTestModeOnce sync.Once
	testMode           bool
	skipRateLimit      bool
)

// AddTestModeCommandFlag 添加测试模式命令参数。
func AddTestModeCommandFlag() {
	flag.BoolVar(&testMode, CommandFlagLocalTestMode, false, "specify server in test mode")
}

// AddSkipRateLimitCommandFlag 添加不限制请求速度命令参数。
func AddSkipRateLimitCommandFlag() {
	flag.BoolVar(&skipRateLimit, CommandFlagSkipRateLimit, false, "whether the request rate is not limited")
}

// UnitTestMode 单测模式。
func UnitTestMode() bool {
	setUniTestModeOnce.Do(func() {
		callers := make([]uintptr, 32)
		n := runtime.Callers(3, callers)
		callers = callers[:n]
		frames := runtime.CallersFrames(callers)
		for {
			frame, more := frames.Next()
			index := strings.LastIndex(frame.File, SystemName)
			if index != -1 {
				if strings.HasSuffix(strings.Trim(frame.File[index+len(SystemName):], "/"), "_test.go") {
					unitTestMode = true
					break
				}
			}

			if !more {
				break
			}
		}
	})
	return unitTestMode
}

// TestMode 测试模式。
func TestMode() bool {
	return testMode
}

// SkipRateLimit 不限制请求速度。
func SkipRateLimit() bool {
	return skipRateLimit
}
