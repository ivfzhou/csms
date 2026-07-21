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

package api_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestAndroidAPIDownloadCertificate(t *testing.T) {
	const reqPath = "/api/android/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("certificate content"))
		mockAesKey := &model.AesKey{Secret: secret} // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AndroidCertificate{
			Content:  encrypt,
			Category: model.AndroidCertificateTypeDebug,
			Alias_:   "alias",
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书数据。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书加密密钥。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.AndroidAPIDownloadCertificateReq{CertificateID: certificateID}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs, _ := io.ReadAll(rsp.Body)
		if len(bs) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.AndroidAPIDownloadCertificateReq{CertificateID: certificateID}))

		if rsp.Code != http.StatusBadRequest {
			t.Errorf("expect http code %d, but got %d", http.StatusBadRequest, rsp.Code)
		}
		var rspBodyObj util.Response[any]
		_ = json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
		if rspBodyObj.Code != errs.ErrInvalidRequestParameters {
			t.Errorf("expect %v, but got %v", errs.ErrInvalidRequestParameters, rspBodyObj.Code)
		}
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书 ID 缺失", ""},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPISubmitAPKSigningJob(t *testing.T) {
	const reqPath = "/api/android/submitAPKSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		signatureSchema := []int{1}
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（APK 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidAPISubmitAPKSigningJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.AndroidAPISubmitAPKSigningJobReq{
					SignatureSchema: signatureSchema,
					CertificateID:   certificateID,
					FileID:          fileID,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			&protocol.AndroidAPISubmitAPKSigningJobReq{
				SignatureSchema: signatureSchema,
				CertificateID:   certificateID,
				FileID:          fileID,
			}))

		if rsp.Code != http.StatusBadRequest {
			t.Errorf("expect http code %d, but got %d", http.StatusBadRequest, rsp.Code)
		}
		var rspBodyObj util.Response[any]
		_ = json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
		if rspBodyObj.Code != errs.ErrInvalidRequestParameters {
			t.Errorf("expect %v, but got %v", errs.ErrInvalidRequestParameters, rspBodyObj.Code)
		}
	}

	for _, v := range []struct {
		Name            string
		SignatureSchema []int
		CertificateID   string
		FileID          string
	}{
		{"签名方案缺失", nil, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"签名方案重复", []int{1, 1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 缺失", []int{1}, "", util.FastRandomAlphaNumberString(38)},
		{"证书 ID 错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidAPISubmitAABSigningJob(t *testing.T) {
	const reqPath = "/api/android/submitAABSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAAB,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（AAB 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidAPISubmitAABSigningJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.AndroidAPISubmitAABSigningJobReq{
					CertificateID: certificateID,
					FileID:        fileID,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			&protocol.AndroidAPISubmitAABSigningJobReq{
				CertificateID: certificateID,
				FileID:        fileID,
			}))

		if rsp.Code != http.StatusBadRequest {
			t.Errorf("expect http code %d, but got %d", http.StatusBadRequest, rsp.Code)
		}
		var rspBodyObj util.Response[any]
		_ = json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
		if rspBodyObj.Code != errs.ErrInvalidRequestParameters {
			t.Errorf("expect %v, but got %v", errs.ErrInvalidRequestParameters, rspBodyObj.Code)
		}
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
		FileID        string
	}{
		{"证书 ID 缺失", "", util.FastRandomAlphaNumberString(38)},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidAPISubmitAPKPatchSigningJob(t *testing.T) {
	const reqPath = "/api/android/submitAPKPatchSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		signatureSchema := []int{1}
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		minimumSDKVersion := 1
		mockFile := &model.File{
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   cc.ExtensionAPK,
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（APK 签名类型）。
		mockCert := &model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)} // 模拟数据库中的安卓证书记录（有效期未过期）。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中安卓证书信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 添加安卓签名任务 ID 到 Redis。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.CreateOnce(nil)           // 保存安卓签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送安卓签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidAPISubmitAPKPatchSigningJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.AndroidAPISubmitAPKPatchSigningJobReq{
					SignatureSchema:   signatureSchema,
					CertificateID:     certificateID,
					FileID:            fileID,
					MinimumSDKVersion: minimumSDKVersion,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string, minimumSDKVersion int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			&protocol.AndroidAPISubmitAPKPatchSigningJobReq{
				SignatureSchema:   signatureSchema,
				CertificateID:     certificateID,
				FileID:            fileID,
				MinimumSDKVersion: minimumSDKVersion,
			}))

		if rsp.Code != http.StatusBadRequest {
			t.Errorf("expect http code %d, but got %d", http.StatusBadRequest, rsp.Code)
		}
		var rspBodyObj util.Response[any]
		_ = json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
		if rspBodyObj.Code != errs.ErrInvalidRequestParameters {
			t.Errorf("expect %v, but got %v", errs.ErrInvalidRequestParameters, rspBodyObj.Code)
		}
	}

	for _, v := range []struct {
		Name              string
		SignatureSchema   []int
		CertificateID     string
		FileID            string
		MinimumSDKVersion int
	}{
		{"签名方案缺失", nil, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 1},
		{"签名方案重复", []int{1, 1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 缺失", []int{1}, "", util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38), 1},
		{"证书 ID 非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38), 1},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), "", 1},
		{"文件 ID 错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37), 1},
		{"文件 ID 非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉", 1},
		{"SDK 版本缺失", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 0},
		{"SDK 版本非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID, v.MinimumSDKVersion)
		})
	}
}

func TestAndroidAPIGetSigningJobInformation(t *testing.T) {
	const reqPath = "/api/android/getSigningJobInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		jobID := util.FastRandomAlphaNumberString(38)
		fileID := util.FastRandomAlphaNumberString(38)
		certificateID := util.FastRandomAlphaNumberString(32)
		certAlias := "alias"
		fileName := "test.apk"
		mockJob := &model.AndroidSigningJob{
			UserID:        APIAccount.ID,
			Source:        model.SourceAPI,
			CertificateID: 1,
			FileID:        fileID,
			Type:          model.AndroidSigningJobTypeAPK,
			CreatedTime:   time.Now(),
		} // 模拟数据库中的安卓签名任务记录（来源为 API）。
		mockCert := &model.AndroidCertificate{
			CertificateID: certificateID,
			Alias_:        certAlias,
		} // 模拟数据库中的安卓证书记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAndroidSigningJobMocker := MockDBClient[model.AndroidSigningJob](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                          // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                                      // 校验 API 凭证权限。
		dbAndroidSigningJobMocker = dbAndroidSigningJobMocker.TakeOnce(mockJob, nil)                               // 查询数据库中签名任务信息。
		dbAPIAccountMocker = dbAPIAccountMocker.ScanOnce(func(v any) { *v.(*string) = APIAccount.AccountID }, nil) // 查询数据库中 API 凭证账号名。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.TakeOnce(mockCert, nil)                            // 查询数据库中安卓证书信息。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = fileName }, nil)                         // 查询数据库中文件名。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAndroidSigningJobMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidAPIGetSigningJobInformationRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				protocol.AndroidAPIGetSigningJobInformationReq{JobID: jobID})),
			0,
		)

		if rspBodyObj.Data.FileName != fileName {
			t.Errorf("expect file name %s, but got %s", fileName, rspBodyObj.Data.FileName)
		}
		if rspBodyObj.Data.CertificateAlias != certAlias {
			t.Errorf("expect certificate alias %s, but got %s", certAlias, rspBodyObj.Data.CertificateAlias)
		}
	})

	validateErrorRequest := func(t *testing.T, jobID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.AndroidAPIGetSigningJobInformationReq{JobID: jobID}))

		if rsp.Code != http.StatusBadRequest {
			t.Errorf("expect http code %d, but got %d", http.StatusBadRequest, rsp.Code)
		}
		var rspBodyObj util.Response[any]
		_ = json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
		if rspBodyObj.Code != errs.ErrInvalidRequestParameters {
			t.Errorf("expect %v, but got %v", errs.ErrInvalidRequestParameters, rspBodyObj.Code)
		}
	}

	for _, v := range []struct {
		Name  string
		JobID string
	}{
		{"任务 ID 缺失", ""},
		{"任务 ID 错误", util.FastRandomAlphaNumberString(37)},
		{"任务 ID 非法", util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.JobID)
		})
	}
}

func TestAndroidAPIListCertificates(t *testing.T) {
	const reqPath = "/api/android/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockUserList := []*model.User{{ID: 1, NameEn: "zs"}, {ID: 2, NameEn: "ls"}} // 模拟数据库中的用户列表数据。
		mockCertList := []*model.AndroidCertificate{{}, {}}                         // 模拟数据库中的安卓证书列表数据。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAndroidCertificateMocker := MockDBClient[model.AndroidCertificate](ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformAndroid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                    // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                    // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                   // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)               // 校验 API 凭证权限。
		dbAndroidCertificateMocker = dbAndroidCertificateMocker.FindOnce(mockCertList, nil) // 查询数据库中安卓证书列表。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                             // 查询数据库中用户英文名。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAndroidCertificateMocker.Reset()
		defer dbUserMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidAPIListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})
}
