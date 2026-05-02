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

// 缓存键。
const (
	RedisKeyAntiShake                   = "csms:backend:hash:antiShake"
	RedisKeySessionFmt                  = "csms:backend:string:session:%s"
	RedisKeyIDFmt                       = "csms:backend:set:id:%s"
	RedisKeyFileUploadInfo              = "csms:backend:hash:fileUploadInfo"
	RedisKeyFileUploadPartInfoFmt       = "csms:backend:sset:fileUploadPartInfo:%s"
	RedisKeyFileUploadPartInfoKeyPrefix = "csms:backend:sset:fileUploadPartInfo:"
	RedisKeyFileUploadLockFmt           = "csms:backend:string:fileUploadLock:%s"
	RedisKeyFileUploadPartLockFmt       = "csms:backend:string:fileUploadPartLock:%s_%d"
	RedisKeyCronLockFmt                 = "csms:backend:string:cronLock:%s_%s"
	RedisKeyApiAccessLimitPrefix        = "csms:backend:hash:apiAccessLimit:"
	RedisKeyAttestationStartLockFmt     = "csms:backend:string:attestationStartLock:%s"
	RedisKeyWHQLStartLockFmt            = "csms:backend:string:whqlStartLock:%s"
	RedisKeyAttestationCheckLockFmt     = "csms:backend:string:attestationCheckLock:%s"
	RedisKeyWHQLCheckLockFmt            = "csms:backend:string:whqlCheckLock:%s"
	RedisKeyMicrosoftAccessToken        = "csms:backend:string:microsoftAccessToken"
)
