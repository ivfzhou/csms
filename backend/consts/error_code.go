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

import "gitee.com/ivfzhou/csms/comm/errs"

// 错误码。
const (
	// 系统运行失败。
	ErrSystem errs.Code = iota + 200001
	// 客户端请求参数非法。
	ErrParameterInvalid
	// 申请苹果证书时，证书已存在。
	ErrAppleCertificateExists
	// 客户端日期请求头非法。
	ErrRequestDateInvalid
	// 客户端无权请求。
	ErrPermissionDenied
	// 客户端请求无合法登录凭证。
	ErrNeedLogin
	// 客户端请求频率过大。
	ErrRateLimitReached
	// 应用状态非正常，不允许操作。
	ErrAppStatusNotValid
	// 用户英文名重复。
	ErrUserNameDuplicate
	// 登录失败。
	ErrUserLoginFailed
	// 上传的文件分片不是有序的。
	ErrFilePartNotOrder
	// 上传的的文件大小与分片大小不一致。
	ErrFileSizeInvalid
	// 操作失败。
	ErrCommonFailure
	// 请求凭证账号 ID 已存在，不可再申请。
	ErrAuthIDAlreadyExist
	// 应用状态不合法。
	ErrAppStatusNotInvalid
	// 相同的待办正在处理中。
	ErrSameTodoExist
	// 证书密码错误。
	ErrCertPasswordInvalid
	// 证书内容错误。
	ErrCertFormatInvalid
	// 密钥库的解密密码错误。
	ErrKeystoreStorePassInvalid
	// 密钥库的密钥错误。
	ErrKeystoreKeyPassInvalid
	// Bundle ID 已被注册。
	ErrAppleBundleIDExist
	// 没有签名证书。
	ErrNoAppleSigningCertificate
	// 描述文件名称已被注册。
	ErrAppleProfileNameExist
	// Windows EV 证书已存在。
	ErrWindowsEVCertificateExist
	// 用户上传的文件太大。
	ErrUserAvatarTooLarge
	// 用户文件头像格式不支持。
	ErrUserAvatarFormatNotSupported
	// 描述文件未找到。
	ErrAppleProfileNotFound
	// 文件未找到。
	ErrFileNotFound
	// 应用图标文件太大。
	ErrAppLogoTooLarge
	// 应用图标文件格式不支持。
	ErrAppLogoFormatNotSupported
	// 用户不存在。
	ErrUserNotExists
	// 测试设备已被注册了。
	ErrAppleDeviceRegistered
	// 测试设备 ID 非法。
	ErrInvalidAppleDeviceUDID
	// 测试设备注册数量达到上限。
	ErrAppleDeviceRegisterReachLimit
	// 未找到测试设备。
	ErrAppleDeviceNotFound
	// Bundle ID 在使用，不能删除。
	ErrAppleBundleIDIsUsing
	// API 请求 IP 限制。
	ErrRequestIPNotAllowed
	// 请求凭证账号过期。
	ErrApiAccountExpired
	// 请求凭证过期。
	ErrApiAuthorizationExpired
	// 应用平台不符合预期。
	ErrAppPlatformNotSupported
)

// 提示语码。
const (
	// 注册用户。
	AlertRegisterUser errs.Code = -iota + -200001
	// 用户登录。
	AlertLogin
	// 登出。
	AlertLogout
	// 操作成功。
	AlertSuccess
	// 有效化应用。
	AlertEnableApp
	// 注册应用。
	AlertRegisterApp
	// 文件上传成功。
	AlertFileUpload
	// 修改 Bundle ID 存在失败。
	AlertErrorInModifyingAppleBundleID
	// 注册测试设备。
	AlertRegisterAppleDevice
)
