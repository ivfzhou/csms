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

func TestAppleAPIDownloadCertificate(t *testing.T) {
	const reqPath = "/api/apple/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		certificateID := util.FastRandomAlphaNumberString(32)
		secret := util.RandomBytes(16)
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("test file content"))
		mockAesKey := &model.AesKey{Secret: secret}                        // 模拟数据库中的 AES 加密密钥记录。
		mockCert := &model.AppleCertificate{Content: encrypt, AesKeyID: 1} // 模拟数据库中的苹果证书记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAppleCertificateMocker := MockDBClient[model.AppleCertificate](ctx)
		dbAesKeyMocker := MockDBClient[model.AesKey](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)          // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)      // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                            // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                            // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)           // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)       // 校验 API 凭证权限。
		dbAppleCertificateMocker = dbAppleCertificateMocker.TakeOnce(mockCert, nil) // 查询数据库中苹果证书信息。
		dbAesKeyMocker = dbAesKeyMocker.TakeOnce(mockAesKey, nil)                   // 查询数据库中证书解密密钥。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAppleCertificateMocker.Reset()
		defer dbAesKeyMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.AppleAPIDownloadCertificateReq{
				CertificateID: certificateID,
				Type:          protocol.AppleFileTypeSigningCertificate,
			}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs, _ := io.ReadAll(rsp.Body)
		if len(bs) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string, fileType int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
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
			protocol.AppleAPIDownloadCertificateReq{
				CertificateID: certificateID,
				Type:          fileType,
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
		certificateID string
		fileType      int
	}{
		{"证书 ID 缺失", "", protocol.AppleFileTypeSigningCertificate},
		{"证书 ID 错误", util.FastRandomAlphaNumberString(31), protocol.AppleFileTypeSigningCertificate},
		{"证书 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", protocol.AppleFileTypeSigningCertificate},
		{"文件类型小于最小值", util.FastRandomAlphaNumberString(32), 0},
		{"文件类型大于最大值", util.FastRandomAlphaNumberString(32), 4},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.certificateID, v.fileType)
		})
	}
}

func TestAppleAPISubmitSigningJob(t *testing.T) {
	const reqPath = "/api/apple/submitSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		profileID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		mockProfile := &model.AppleProfile{Type: model.AppleProfileTypeIOSAppInHouse} // 模拟数据库中的描述文件记录（InHouse 类型跳过 Bundle ID 查询）。
		mockFile := &model.File{
			AppID: AppInfo.ID,
			Type:  model.FileTypeAppleSigning,
		} // 模拟数据库中的文件记录（Apple 签名类型）。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		rabbitMQMocker := MockRabbitMQClient(ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		dbAppleProfileMocker = dbAppleProfileMocker.LastOnce(mockProfile, nil) // 查询数据库中描述文件信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库中文件信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                             // Redis 生成唯一 ID。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.CreateOnce(nil)      // 保存 Apple 签名任务到数据库。
		rabbitMQMocker = rabbitMQMocker.PublishWithContextOnce(nil)            // 发送签名任务消息到消息队列。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()
		defer rabbitMQMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppleAPISubmitSigningJobRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.AppleAPISubmitSigningJobReq{
					ProfileID: profileID,
					FileID:    fileID,
				})),
			0,
		)

		if len(rspBodyObj.Data.JobID) <= 0 {
			t.Errorf("job id is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, profileID, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
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
			&protocol.AppleAPISubmitSigningJobReq{
				ProfileID: profileID,
				FileID:    fileID,
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
		Name      string
		profileID string
		fileID    string
	}{
		{"描述文件 ID 缺失", "", util.FastRandomAlphaNumberString(38)},
		{"描述文件 ID 错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"描述文件 ID 非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件 ID 缺失", util.FastRandomAlphaNumberString(32), ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.profileID, v.fileID)
		})
	}
}

func TestAppleAPIGetSigningJobInformation(t *testing.T) {
	const reqPath = "/api/apple/getSigningJobInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		jobID := util.FastRandomAlphaNumberString(38)
		fileID := util.FastRandomAlphaNumberString(38)
		signedFileID := util.FastRandomAlphaNumberString(38)
		profileIDStr := util.FastRandomAlphaNumberString(32)
		fileName := "test.ipa"
		signedFileName := "signed.ipa"
		mockJob := &model.AppleSigningJob{
			UserID:       APIAccount.ID,
			Source:       model.SourceAPI,
			ProfileID:    1,
			FileID:       fileID,
			SignedFileID: signedFileID,
			CreatedTime:  time.Now(),
		} // 模拟数据库中的苹果签名任务记录（来源为 API）。
		mockProfile := &model.AppleProfile{
			ProfileID: profileIDStr,
			BundleID:  0,
		} // 模拟数据库中的描述文件记录（BundleID 为 0 跳过 Bundle ID 查询）。
		mockFileList := []*model.File{
			{FileID: fileID, Name: fileName},
			{FileID: signedFileID, Name: signedFileName},
		} // 模拟数据库中的文件列表数据。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbAppleSigningJobMocker := MockDBClient[model.AppleSigningJob](ctx)
		dbAppleProfileMocker := MockDBClient[model.AppleProfile](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                         // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                     // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                           // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                           // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                                          // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                                      // 校验 API 凭证权限。
		dbAppleSigningJobMocker = dbAppleSigningJobMocker.TakeOnce(mockJob, nil)                                   // 查询数据库中签名任务信息。
		dbAppleProfileMocker = dbAppleProfileMocker.TakeOnce(mockProfile, nil)                                     // 查询数据库中描述文件信息。
		dbFileMocker = dbFileMocker.FindOnce(mockFileList, nil)                                                    // 查询数据库中文件名。
		dbAPIAccountMocker = dbAPIAccountMocker.ScanOnce(func(v any) { *v.(*string) = APIAccount.AccountID }, nil) // 查询数据库中 API 凭证账号名。
		defer appPlatformReset.Reset()
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbAppleSigningJobMocker.Reset()
		defer dbAppleProfileMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppleAPIGetSigningJobInformationRsp](
			t,
			ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				protocol.AppleAPIGetSigningJobInformationReq{JobID: jobID})),
			0,
		)

		if rspBodyObj.Data.FileName != fileName {
			t.Errorf("expect file name %s, but got %s", fileName, rspBodyObj.Data.FileName)
		}
		if rspBodyObj.Data.ProfileID != profileIDStr {
			t.Errorf("expect profile id %s, but got %s", profileIDStr, rspBodyObj.Data.ProfileID)
		}
	})

	validateErrorRequest := func(t *testing.T, jobID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		appPlatformReset := mvt.Chain(AppInfo).Elem().FieldByName("Platform").Set(model.AppPlatformApple)
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
			protocol.AppleAPIGetSigningJobInformationReq{JobID: jobID}))

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
