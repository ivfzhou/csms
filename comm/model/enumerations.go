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

package model

// 应用状态。
const (
	AppStatusValid = 1 + iota
	AppStatusInvalid
	AppStatusApproving
	AppStatusRejected
)

// 应用平台类型。
const (
	AppPlatformWindows = 1 + iota
	AppPlatformApple
	AppPlatformAndroid
)

// 文件类型。
const (
	FileTypeUserAvatar = iota + 1
	FileTypeAppLogo
	FileTypeAndroidSigning
	FileTypeWindowsSigning
	FileTypeAppleSigning
	FileTypeHLKLog
	FileTypeMicrosoftSigning
)

// 请求凭证账号权限。
const (
	// Windows
	CapabilityDownloadWindowsOVCertificate = 1 + iota
	CapabilityGetWindowsOVCertificatePassword
	CapabilitySubmitWindowsPESignJob
	CapabilitySubmitWHQLSignJob

	// 安卓
	CapabilityDownloadAndroidCertificate
	CapabilitySubmitAndroidSignJob

	// 苹果
	CapabilityDownloadAppleCertAndProvision
	CapabilitySubmitAppleSignJob

	// 公共
	CapabilityGetAppInfo
	CapabilityGetCertificateInfo
	CapabilityGetSignJobInfo
)

// 用户应用权限。
const (
	UserRoleSystemAdmin = iota + 1
	UserRoleAppAdmin
	UserRoleAppMember
	UserRoleAppSigner
)

// 待办状态。
const (
	TodoStatusProcessing = 1 + iota
	TodoStatusRejected
	TodoStatusApproved
)

// 待办类型。
const (
	TodoTypeRegisterApp = 1 + iota
	TodoTypeJoinApp
	TodoTypeApplySigner
	TodoTypeRegisterAppleDevice
	TodoTypeActivateApp
)

// 事件类型。
const (
	EventTypeRegisterApp = 1 + iota
	EventTypeInvalidateApp
	EventTypeEnableApp
	EventTypeUpdateApp

	EventTypeApplyOpenAPI
	EventTypeUpdateOpenAPI
	EventTypeRemoveOpenAPI
	EventTypeRenewOpenAPI
	EventTypeResetOpenAPI

	EventTypeUploadWindowsCertificate
	EventTypeRemoveWindowsCertificate
	EventTypeDownloadWindowsCertificate

	EventTypeUploadAndroidCertificate
	EventTypeApplyAndroidCertificate
	EventTypeRemoveAndroidCertificate
	EventTypeDownloadAndroidCertificate
	EventTypeDownloadGooglePlayCertificate
	EventTypeGetFacebookCertificateDigest

	EventTypeApplyAppleBundleID
	EventTypeModifyAppleBundleID
	EventTypeRemoveAppleBundleID
	EventTypeApplyProvision
	EventTypeApplyPushCertificate
	EventTypeDownloadAppleCertificate
	EventTypeDownloadProvision
	EventTypeRemoveProvision
	EventTypeRemovePushCertificate
	EventTypeRegisterAppleDevice
)

// 来源类型。
const (
	SourceWeb = 1 + iota
	SourceAPI
	SourceInternal
)

// Apple 测试设备状态。
const (
	AppleDeviceStatusOK = 1 + iota
	AppleDeviceStatusApproving
	AppleDeviceStatusRejected
)

// Apple 平台描述。
const (
	ApplePlatformIOSDescription       = "IOS"
	ApplePlatformMacOSDescription     = "MAC_OS"
	ApplePlatformUniversalDescription = "UNIVERSAL"
)

// Apple 平台。
const (
	ApplePlatformIOS = 1 + iota
	ApplePlatformMaxOS
	ApplePlatformUniversal
)

// Apple Bundle ID 类型。
const (
	AppleBundleIDTypeAppStore = 1 + iota
	AppleBundleIDTypeInHouse
)

// Apple 证书类型。
const (
	AppleCertificateTypeDistribution       = "DISTRIBUTION"
	AppleCertificateTypeDevelopment        = "DEVELOPMENT"
	AppleCertificateTypeIOSDevelopment     = "IOS_DEVELOPMENT"
	AppleCertificateTypeIOSDistribution    = "IOS_DISTRIBUTION"
	AppleCertificateTypeMacAppDistribution = "MAC_APP_DISTRIBUTION"
	AppleCertificateTypeMacAppDevelopment  = "MAC_APP_DEVELOPMENT"
)

// Apple 证书类别。
const (
	AppleCertificateCategorySigning = 1 + iota
	AppleCertificateCategoryPush
)

// Apple 描述文件类型。
const (
	AppleProfileTypeIOSAppDevelopment = "IOS_APP_DEVELOPMENT"
	AppleProfileTypeMacAppDevelopment = "MAC_APP_DEVELOPMENT"
	AppleProfileTypeIOSAppAdhoc       = "IOS_APP_ADHOC"
	AppleProfileTypeIOSAppStore       = "IOS_APP_STORE"
	AppleProfileTypeMacAppStore       = "MAC_APP_STORE"
	AppleProfileTypeIOSAppInHouse     = "IOS_APP_INHOUSE"
)

// Apple Push 证书环境。
const (
	AppleCertificateEnvironmentProduction = 1 + iota
	AppleCertificateEnvironmentDevelopment
	AppleCertificateEnvironmentInHouse
)

// Windows 证书类型。
const (
	WindowsCertificateTypeCompanyEV = 1 + iota
	WindowsCertificateTypeCompanyOV
	WindowsCertificateTypePersonalEV
	WindowsCertificateTypePersonalOV
)

// Windows 签名任务类型。
const (
	WindowsSigningJobTypePE = 1 + iota
	WindowsSigningJobTypeAttestation
	WindowsSigningJobTypePEAndAttestation
	WindowsSigningJobTypeHLKX
)

// Windows 签名任务状态。
const (
	WindowsSigningJobStatusSigning = 1 + iota
	WindowsSigningJobStatusWaitCabSign
	WindowsSigningJobStatusCabSigning
	WindowsSigningJobStatusAttestationWaiting
	WindowsSigningJobStatusAttestationSigning
	WindowsSigningJobStatusFailure
	WindowsSigningJobStatusSuccess
)

// HLK 测试系统类型。
const (
	WHQLJobTestSystemWindowsServer2016_64 = "Windows Server 2016_64"
	WHQLJobTestSystemWindowsServer2019_64 = "Windows Server 2019_64"
	WHQLJobTestSystemWindowsServer2022_64 = "Windows Server 2022_64"
	WHQLJobTestSystemWindows10_1607_32    = "Windows 10 1607_32"
	WHQLJobTestSystemWindows10_1607_64    = "Windows 10 1607_64"
	WHQLJobTestSystemWindows10_1809_64    = "Windows 10 1809_64"
	WHQLJobTestSystemWindows10_22H2_64    = "Windows 10 22H2_64"
	WHQLJobTestSystemWindows10_22H2_32    = "Windows 10 22H2_32"
	WHQLJobTestSystemWindows11_22H2_64    = "Windows 11 22H2_64"
)

// WHQL 任务状态。
const (
	WHQLJobStatusWaitingTest = 1 + iota
	WHQLJobStatusInitiallingTestMachine
	WHQLJobStatusFinishInitiallingTestMachine
	WHQLJobStatusDispatchHLKTest
	WHQLJobStatusHLKTesting
	WHQLJobStatusFinishTest
	WHQLJobStatusHLKXFileSinging
	WHQLJobStatusWHQLSigning
	WHQLJobStatusFailure
	WHQLJobStatusSuccess
)

// WHQL 任务类型。
const (
	WHQLJobTypeHLKAndWHQL = 1 + iota
	WHQLJobTypeOnlyWHQL
)

// 安卓证书类型。
const (
	AndroidCertificateTypeDebug = 1 + iota
	AndroidCertificateTypeRelease
)

// 安卓签名任务类型。
const (
	AndroidSigningJobTypeAPK = 1 + iota
	AndroidSigningJobTypeAAB
	AndroidSigningJobTypePatch
)

// 安卓签名任务状态。
const (
	AndroidSigningJobStatusSigning = 1 + iota
	AndroidSigningJobStatusSuccess
	AndroidSigningJobStatusFailure
)

// Apple 签名任务状态。
const (
	AppleSigningJobStatusRunning = 1 + iota
	AppleSigningJobStatusSuccess
	AppleSigningJobStatusFailure
)

const (
	AppleDeviceTypeMac        = "MAC"
	AppleDeviceTypeIpad       = "IPAD"
	AppleDeviceTypeIpod       = "IPOD"
	AppleDeviceTypeIphone     = "IPHONE"
	AppleDeviceTypeAppleTV    = "APPLE_TV"
	AppleDeviceTypeAppleWatch = "APPLE_WATCH"
)

var AllAppleDeviceTypes = []string{
	AppleDeviceTypeMac,
	AppleDeviceTypeIpad,
	AppleDeviceTypeIpod,
	AppleDeviceTypeIphone,
	AppleDeviceTypeAppleTV,
	AppleDeviceTypeAppleWatch,
}

// 凭证权限。
var (
	AllWindowsAppCapabilities = []int{
		CapabilityDownloadWindowsOVCertificate,
		CapabilityGetWindowsOVCertificatePassword,
		CapabilitySubmitWindowsPESignJob,
		CapabilitySubmitWHQLSignJob,
		CapabilityGetAppInfo,
		CapabilityGetCertificateInfo,
		CapabilityGetSignJobInfo,
	}
	AllAndroidAppCapabilities = []int{
		CapabilityDownloadAndroidCertificate,
		CapabilitySubmitAndroidSignJob,
		CapabilityGetAppInfo,
		CapabilityGetCertificateInfo,
		CapabilityGetSignJobInfo,
	}
	AllAppleAppCapabilities = []int{
		CapabilityDownloadAppleCertAndProvision,
		CapabilitySubmitAppleSignJob,
		CapabilityGetAppInfo,
		CapabilityGetCertificateInfo,
		CapabilityGetSignJobInfo,
	}
)

// AllAppPlatformDescriptions 平台描述
var AllAppPlatformDescriptions = map[int]string{
	AppPlatformWindows: "Windows",
	AppPlatformApple:   "Apple",
	AppPlatformAndroid: "Android",
}

// AllPushCertificateEnvironments Push 证书环境名称。
var AllPushCertificateEnvironments = map[int]string{
	AppleCertificateEnvironmentProduction:  "正式",
	AppleCertificateEnvironmentDevelopment: "开发",
	AppleCertificateEnvironmentInHouse:     "企业内测",
}

// AllWindowsCertificateDescriptions Windows 证书描述。
var AllWindowsCertificateDescriptions = map[int]string{
	WindowsCertificateTypeCompanyEV:  "公司EV",
	WindowsCertificateTypeCompanyOV:  "公司OV",
	WindowsCertificateTypePersonalEV: "个人EV",
	WindowsCertificateTypePersonalOV: "个人OV",
}

// AllAndroidCertificateTypeDescriptions 安卓证书类型。
var AllAndroidCertificateTypeDescriptions = map[int]string{
	AndroidCertificateTypeDebug:   "调试",
	AndroidCertificateTypeRelease: "发布",
}

// AllAppleBundleIDDescriptions Bundle ID 环境。
var AllAppleBundleIDDescriptions = map[int]string{
	AppleBundleIDTypeAppStore: "AppStore",
	AppleBundleIDTypeInHouse:  "企业内测",
}

// AllApplePlatformDescriptionToNumber Apple 平台。
var AllApplePlatformDescriptionToNumber = map[string]int{
	ApplePlatformIOSDescription:       ApplePlatformIOS,
	ApplePlatformMacOSDescription:     ApplePlatformMaxOS,
	ApplePlatformUniversalDescription: ApplePlatformUniversal,
}

// AllAppleCertificateTypes 苹果证书类型。
var AllAppleCertificateTypes = map[string]struct{}{
	AppleCertificateTypeDevelopment:        {},
	AppleCertificateTypeDistribution:       {},
	AppleCertificateTypeIOSDevelopment:     {},
	AppleCertificateTypeIOSDistribution:    {},
	AppleCertificateTypeMacAppDistribution: {},
	AppleCertificateTypeMacAppDevelopment:  {},
}

// AllAppleProfileTypes 描述文件类型。
var AllAppleProfileTypes = map[string]struct{}{
	AppleProfileTypeIOSAppDevelopment: {},
	AppleProfileTypeMacAppDevelopment: {},
	AppleProfileTypeIOSAppAdhoc:       {},
	AppleProfileTypeIOSAppStore:       {},
	AppleProfileTypeMacAppStore:       {},
}

// AllApplePlatformDescriptions 所有的 Apple 平台类型描述。
var AllApplePlatformDescriptions = map[string]struct{}{
	ApplePlatformIOSDescription:       {},
	ApplePlatformMacOSDescription:     {},
	ApplePlatformUniversalDescription: {},
}

// AllWHQLJobTestSystems 所有的 HLK 测试系统类型。
var AllWHQLJobTestSystems = map[string]struct{}{
	WHQLJobTestSystemWindowsServer2016_64: {},
	WHQLJobTestSystemWindowsServer2019_64: {},
	WHQLJobTestSystemWindowsServer2022_64: {},
	WHQLJobTestSystemWindows10_1607_32:    {},
	WHQLJobTestSystemWindows10_1607_64:    {},
	WHQLJobTestSystemWindows10_1809_64:    {},
	WHQLJobTestSystemWindows10_22H2_64:    {},
	WHQLJobTestSystemWindows10_22H2_32:    {},
	WHQLJobTestSystemWindows11_22H2_64:    {},
}
