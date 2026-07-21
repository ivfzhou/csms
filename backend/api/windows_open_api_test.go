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
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestWindowsAPIDownloadCertificate(t *testing.T) {
	const reqPath = "/api/windows/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("test certificate content"))
		mockCert := &model.WindowsCertificate{Content: encrypt, AesKeyID: 1} // 模拟数据库中的 Windows 证书记录。
		mockAesKey := &model.AesKey{Secret: secret}                          // 模拟数据库中的 AES 加密密钥记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中 Windows 证书信息。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                       // 查询数据库中证书解密密钥。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                   // 添加应用事件到数据库。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()
		defer dbEventMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.WindowsAPIDownloadCertificateReq{CertificateID: certificateID}))

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
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			protocol.WindowsAPIDownloadCertificateReq{CertificateID: certificateID}))

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

func TestWindowsAPIGetCertificatePassword(t *testing.T) {
	const reqPath = "/api/windows/getCertificatePassword"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		password := util.FastRandomAlphaNumberString(6)
		mockCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalEV, Password: password} // 模拟数据库中的 EV 证书记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                  // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                              // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                    // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                    // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                   // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                               // 校验 API 凭证权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)                     // 查询数据库中证书信息。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.CountOnce(1, nil) // 校验应用是否已获授权。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPIGetCertificatePasswordRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				protocol.WindowsAPIGetCertificatePasswordReq{CertificateID: certificateID})),
			0,
		)

		if rspBodyObj.Data.Password != password {
			t.Errorf("expect password %s, but got %s", password, rspBodyObj.Data.Password)
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			protocol.WindowsAPIGetCertificatePasswordReq{CertificateID: certificateID}))

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

func TestWindowsAPISubmitSigningJob(t *testing.T) {
	const reqPath = "/api/windows/submitSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockCert := &model.WindowsCertificate{Type: model.WindowsCertificateTypePersonalOV, AppID: AppInfo.ID} // 模拟数据库中的 OV 证书记录。
		mockFile := &model.File{
			Type:   model.FileTypeWindowsSigning,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   "test.dll",
			TusdID: util.FastRandomAlphaNumberString(38),
		} // 模拟数据库中的文件记录。
		peBinary := GenerateMinimalPE() // PE 文件二进制数据。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbWindowsSigningJobMocker := MockDBClient[model.WindowsSigningJob](ctx)
		tusdMocker := MockTusdClient(ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)               // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)           // 校验 API 凭证权限。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中证书信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                             // 查询数据库中文件信息。
		tusdMocker = tusdMocker.DownloadToFileOnce(peBinary, nil)                       // 下载 PE 文件校验格式。
		redisMocker = redisMocker.SAddOnce(1, nil)                                      // 生成任务 ID。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.CreateOnce(nil)           // 保存签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)                     // 发送签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbWindowsSigningJobMocker.Reset()
		defer tusdMocker.Reset()
		defer rabbitMQMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPISubmitSigningJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.WindowsAPISubmitSigningJobReq{
					SigningType:   model.WindowsSigningJobTypePE,
					CertificateID: certificateID,
					FileID:        fileID,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, signingType int, certificateID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			&protocol.WindowsAPISubmitSigningJobReq{
				SigningType:   signingType,
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
		SigningType   int
		CertificateID string
		FileID        string
	}{
		{"签名类型缺失", 0, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"签名类型错误", 4, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"文件 ID 缺失", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
		{"证书 ID 错误", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书 ID 非法", model.WindowsSigningJobTypePE, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SigningType, v.CertificateID, v.FileID)
		})
	}
}

func TestWindowsAPISubmitWHQLJob(t *testing.T) {
	const reqPath = "/api/windows/submitWHQLJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			Type:   model.FileTypeWindowsSigning,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   "test.sys",
			TusdID: util.FastRandomAlphaNumberString(38),
		} // 模拟数据库中的文件记录（.sys 类型）。
		peBinary := GenerateMinimalPE() // PE 文件二进制数据。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbWhqlJobMocker := MockDBClient[model.WhqlJob](ctx)
		tusdMocker := MockTusdClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库中文件信息。
		tusdMocker = tusdMocker.DownloadToFileOnce(peBinary, nil)              // 下载 .sys 文件校验 PE 格式。
		redisMocker = redisMocker.SAddOnce(1, nil)                             // 生成 WHQL 任务 ID。
		dbWhqlJobMocker = dbWhqlJobMocker.CreateOnce(nil)                      // 保存 WHQL 任务到数据库。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbWhqlJobMocker.Reset()
		defer tusdMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPISubmitWHQLJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.WindowsAPISubmitWHQLJobReq{
					SigningType: model.WHQLJobTypeHLKAndWHQL,
					TestSystem:  model.WHQLJobTestSystemWindows10_22H2_64,
					FileID:      fileID,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, signingType int, testSystem, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			&protocol.WindowsAPISubmitWHQLJobReq{
				SigningType: signingType,
				TestSystem:  testSystem,
				FileID:      fileID,
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
		Name        string
		SigningType int
		TestSystem  string
		FileID      string
	}{
		{"签名类型缺失", 0, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38)},
		{"签名类型非法", 3, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(38)},
		{"测试系统缺失", model.WHQLJobTypeHLKAndWHQL, "", util.FastRandomAlphaNumberString(38)},
		{"测试系统非法", model.WHQLJobTypeHLKAndWHQL, "InvalidSystem", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 缺失", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, ""},
		{"文件 ID 错误", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", model.WHQLJobTypeHLKAndWHQL, model.WHQLJobTestSystemWindows10_22H2_64, util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SigningType, v.TestSystem, v.FileID)
		})
	}
}

func TestWindowsAPIListCertificates(t *testing.T) {
	const reqPath = "/api/windows/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockUserList := []*model.User{{ID: 1, NameEn: "zs"}, {ID: 2, NameEn: "ls"}} // 模拟数据库中的用户列表数据。
		mockCertList := []*model.WindowsCertificate{{}, {}}                         // 模拟数据库中的 Windows 证书列表数据。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbWindowsCertificateAuthorizationMocker := MockDBClient[model.WindowsCertificateAuthorization](ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                                                      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                                                  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                                                        // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                                                        // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                                                       // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                                                                   // 校验 API 凭证权限。
		dbWindowsCertificateAuthorizationMocker = dbWindowsCertificateAuthorizationMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询应用已授权的 EV 证书 IDs。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.FindOnce(mockCertList, nil)                                                     // 查询数据库中 Windows 证书列表。
		dbUserMocker = dbUserMocker.FindOnce(mockUserList, nil)                                                                                 // 查询数据库中用户英文名。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbWindowsCertificateAuthorizationMocker.Reset()
		defer dbUserMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPIListCertificatesRsp](
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

func TestWindowsAPIGetSigningJobInformation(t *testing.T) {
	const reqPath = "/api/windows/getSigningJobInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		jobID := util.FastRandomAlphaNumberString(38)
		fileID := util.FastRandomAlphaNumberString(38)
		certID := util.FastRandomAlphaNumberString(32)
		fileName := "test.dll"
		mockJob := &model.WindowsSigningJob{
			UserID:        APIAccount.ID,
			Source:        model.SourceAPI,
			CertificateID: 1,
			FileID:        fileID,
			CreatedTime:   time.Now(),
		} // 模拟数据库中的签名任务记录（来源为 API）。
		mockCert := &model.WindowsCertificate{
			CertificateID: certID,
			CommonName:    "testCert",
		} // 模拟数据库中的证书记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWindowsSigningJobMocker := MockDBClient[model.WindowsSigningJob](ctx)
		dbWindowsCertificateMocker := MockDBClient[model.WindowsCertificate](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                          // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                                      // 校验 API 凭证权限。
		dbWindowsSigningJobMocker = dbWindowsSigningJobMocker.TakeOnce(mockJob, nil)                               // 查询数据库中签名任务信息。
		dbWindowsCertificateMocker = dbWindowsCertificateMocker.TakeOnce(mockCert, nil)                            // 查询数据库中证书信息。
		dbAPIAccountMocker = dbAPIAccountMocker.ScanOnce(func(v any) { *v.(*string) = APIAccount.AccountID }, nil) // 查询数据库中 API 凭证账号名。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = fileName }, nil)                         // 查询数据库中文件名。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWindowsSigningJobMocker.Reset()
		defer dbWindowsCertificateMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPIGetSigningJobInformationRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				protocol.WindowsAPIGetSigningJobInformationReq{JobID: jobID})),
			0,
		)

		if rspBodyObj.Data.FileName != fileName {
			t.Errorf("expect file name %s, but got %s", fileName, rspBodyObj.Data.FileName)
		}
		if rspBodyObj.Data.CertificateID != certID {
			t.Errorf("expect certificate id %s, but got %s", certID, rspBodyObj.Data.CertificateID)
		}
	})

	validateErrorRequest := func(t *testing.T, jobID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			protocol.WindowsAPIGetSigningJobInformationReq{JobID: jobID}))

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

func TestWindowsAPIGetWHQLJobInformation(t *testing.T) {
	const reqPath = "/api/windows/getWHQLJobInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		jobID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		fileName := "test.sys"
		mockJob := &model.WhqlJob{
			UserID:     APIAccount.ID,
			Source:     model.SourceAPI,
			FileID:     fileID,
			Type:       model.WHQLJobTypeHLKAndWHQL,
			TestSystem: model.WHQLJobTestSystemWindows10_22H2_64,
		} // 模拟数据库中的 WHQL 任务记录（来源为 API）。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbWhqlJobMocker := MockDBClient[model.WhqlJob](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                          // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                                      // 校验 API 凭证权限。
		dbWhqlJobMocker = dbWhqlJobMocker.TakeOnce(mockJob, nil)                                                   // 查询数据库中 WHQL 任务信息。
		dbAPIAccountMocker = dbAPIAccountMocker.ScanOnce(func(v any) { *v.(*string) = APIAccount.AccountID }, nil) // 查询数据库中 API 凭证账号名。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = fileName }, nil)                         // 查询源文件名。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = "" }, nil)                               // 查询 HLKX 包文件名（空）。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = "" }, nil)                               // 查询 HLK 日志文件名（空）。
		dbFileMocker = dbFileMocker.ScanOnce(func(v any) { *v.(*string) = "" }, nil)                               // 查询签名结果文件名（空）。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbWhqlJobMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.WindowsAPIGetWHQLJobInformationRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				protocol.WindowsAPIGetWHQLJobInformationReq{JobID: jobID})),
			0,
		)

		if rspBodyObj.Data.FileName != fileName {
			t.Errorf("expect file name %s, but got %s", fileName, rspBodyObj.Data.FileName)
		}
	})

	validateErrorRequest := func(t *testing.T, jobID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformWindows)
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
			protocol.WindowsAPIGetWHQLJobInformationReq{JobID: jobID}))

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
		{"任务 ID 错误", util.FastRandomAlphaNumberString(31)},
		{"任务 ID 非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.JobID)
		})
	}
}
