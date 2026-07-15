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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

func TestFileWebDownload(t *testing.T) {
	const reqPath = "/web/file/download"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		fileID, fileData, fileName :=
			util.FastRandomAlphaNumberString(38), GenerateBytes(4096), util.FastRandomAlphaNumberString(16)
		mockFile := &model.File{
			ID:     1,
			FileID: fileID,
			TusdID: util.FastRandomAlphaNumberString(32),
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Name:   fileName,
			Md5:    util.FastRandomAlphaNumberString(32),
			Size:   len(fileData),
			Type:   model.FileTypeUserAvatar,
		} // 模拟数据库中的文件记录。
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(fileData)),
			ContentLength: len(fileData),
		} // 模拟存储服务返回的文件下载结果。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库文件记录。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                    // 从存储服务获取文件内容。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbFileMocker.Reset()
		defer tusdMocker.Reset()

		rspBody, fileName2 := CheckAndReadBody(
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebDownloadReq{FileID: fileID})),
		)

		if !reflect.DeepEqual(rspBody, fileData) {
			t.Errorf("file data is not equal, %v %v", len(fileData), len(rspBody))
		}
		if !strings.Contains(fileName2, fileName) {
			t.Errorf("file name is not equal, %s %v", fileName2, fileName)
		}
	})

	validateErrorRequest := func(t *testing.T, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebDownloadReq{FileID: fileID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name   string
		FileID string
		Type   int
	}{
		{"文件 ID 缺失", "", model.FileTypeUserAvatar},
		{"文件 ID 过短", util.FastRandomAlphaNumberString(37), model.FileTypeUserAvatar},
		{"文件 ID 非法", util.FastRandomAlphaNumberString(37) + "汉", model.FileTypeUserAvatar},
	} {
		t.Run("异常请求_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID)
		})
	}
}

func TestFileWebInitial(t *testing.T) {
	const reqPath = "/web/file/initial"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		initName := "文件名.jpg"
		initSize := 1024
		initMD5 := "01234567890123456789012345678901"
		initType := model.FileTypeUserAvatar

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbFileMocker = dbFileMocker.TakeOnce(nil, nil)                         // 检查文件是否已存在（秒传检测）。
		redisMocker = redisMocker.SAddOnce(1, nil)                             // 缓存文件 ID 到 Redis。
		redisMocker = redisMocker.HSetOnce(1, nil)                             // 缓存文件上传信息到 Redis。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID,
				&protocol.FileWebInitialReq{
					Name: initName,
					Size: int64(initSize),
					MD5:  initMD5,
					Type: initType,
				},
			)),
			0,
		)

		if rspBodyObj.Data.Exist != false {
			t.Errorf("file existing is not expected, %v", rspBodyObj.Data.Exist)
		}
		if len(rspBodyObj.Data.FileID) != 38 {
			t.Errorf("file id is not equal, %v", rspBodyObj.Data.FileID)
		}
	})

	t.Run("正常测试_文件存在", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		initName := "文件名.jpg"
		initSize := 1024
		initMD5 := "01234567890123456789012345678901"
		initType := model.FileTypeUserAvatar
		mockFile := &model.File{ID: 1, FileID: fileID} // 模拟数据库中已存在的文件记录。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 检查文件已存在（命中秒传）。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID,
				&protocol.FileWebInitialReq{
					Name: initName,
					Size: int64(initSize),
					MD5:  initMD5,
					Type: initType,
				},
			)),
			0,
		)

		if rspBodyObj.Data.Exist != true {
			t.Errorf("file existing is not expected, %v", rspBodyObj.Data.Exist)
		}
		if rspBodyObj.Data.FileID != fileID {
			t.Errorf("file id is not equal, %v %v", rspBodyObj.Data.FileID, fileID)
		}
	})

	validateErrorRequest := func(t *testing.T, name, md5 string, typ, size int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()

		CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID,
				&protocol.FileWebInitialReq{
					Name: name,
					Size: int64(size),
					MD5:  md5,
					Type: typ,
				},
			)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name       string
		Name2, MD5 string
		Type, Size int
	}{
		{"文件名缺失", "", "01234567890123456789012345678901", model.FileTypeUserAvatar, 1024},
		{"文件名过长", util.FastRandomAlphaNumberString(257), "01234567890123456789012345678901", model.FileTypeUserAvatar, 1024},
		{"文件大小缺失", "文件名.jpg", "01234567890123456789012345678901", model.FileTypeUserAvatar, 0},
		{"文件 MD5 非法", "文件名.jpg", "0123456789012345678901234567890", model.FileTypeUserAvatar, 1024},
		{"文件类型非法", "文件名.jpg", "01234567890123456789012345678901", 0, 1024},
	} {
		t.Run("异常请求_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2, v.MD5, v.Type, v.Size)
		})
	}
}

func TestFileWebUploadPart(t *testing.T) {
	const reqPath = "/web/file/uploadPart"
	createRequestBody := func(fileID *string, chunkNumber *int, fileData []byte) (io.Reader, string) {
		reqBody := &bytes.Buffer{}
		reqBodyWriter := multipart.NewWriter(reqBody)
		if fileID != nil {
			err := reqBodyWriter.WriteField("fileId", *fileID)
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if chunkNumber != nil {
			if err := reqBodyWriter.WriteField("chunkNumber", strconv.Itoa(*chunkNumber)); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if fileData != nil {
			avatar, err := reqBodyWriter.CreateFormFile("chunk", "chunk")
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
			written, err := avatar.Write(fileData)
			if err != nil {
				t.Error("write avatar file error", err)
				return nil, ""
			}
			if written != len(fileData) {
				t.Error("the number of bytes written does not meet the expectations", written, len(fileData))
				return nil, ""
			}
		}
		if err := reqBodyWriter.Close(); err != nil {
			t.Error("failed to close multipart writer", err)
			return nil, ""
		}
		return reqBody, reqBodyWriter.FormDataContentType()
	}

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		uploadFileID := new(util.FastRandomAlphaNumberString(38))
		chunkNumber := new(1)
		chunkData := GenerateBytes(4096)
		bs, _ := json.Marshal(map[string]any{"user": LoginUser.ID, "timeSecond": time.Now().Unix()})

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                    // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                  // 查询数据库登录用户信息。
		redisMocker = redisMocker.HGetOnce(string(bs), nil)                                   // 获取上传文件缓存信息。
		redisMocker = redisMocker.SetNXOnce(true, nil)                                        // 获取分片上传锁。
		redisMocker = redisMocker.ZRangeArgsOnce(nil, nil)                                    // 查询已上传分片列表。
		tusdMocker = tusdMocker.UploadPartByIOOnce(util.FastRandomAlphaNumberString(32), nil) // 上传分片到存储服务。
		redisMocker = redisMocker.ZAddOnce(1, nil)                                            // 记录分片上传信息。
		redisMocker = redisMocker.EvalOnce(true, nil)                                         // 释放分片上传锁。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer tusdMocker.Reset()

		reqBody, contentType := createRequestBody(uploadFileID, chunkNumber, chunkData)
		CheckAndUnmarshalBody[any](t, ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)), 0)
	})

	validateErrorRequest := func(t *testing.T, fileID *string, chunkNumber *int, fileData []byte) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		reqBody, contentType := createRequestBody(fileID, chunkNumber, fileData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		FileID      *string
		ChunkNumber *int
		FileData    []byte
	}{
		{"分片缺失", new(util.FastRandomAlphaNumberString(38)), new(1), nil},
		{"分片大小为零", new(util.FastRandomAlphaNumberString(38)), new(1), []byte{}},
		{"文件 ID 错误", new(util.FastRandomAlphaNumberString(37)), new(1), GenerateBytes(4096)},
		{"文件 ID 字符非法", new("汉" + util.FastRandomAlphaNumberString(37)), new(1), GenerateBytes(4096)},
		{"分片序号错误", new(util.FastRandomAlphaNumberString(38)), new(-1), GenerateBytes(4096)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID, v.ChunkNumber, v.FileData)
		})
	}
}

func TestFileWebMergeParts(t *testing.T) {
	const reqPath = "/web/file/mergeParts"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		// 模拟 Redis 中的分片列表数据。
		redisZ := make([]redis.Z, 0, 3)
		for i := range cap(redisZ) {
			redisZ = append(redisZ, redis.Z{
				Score:  float64(i + 1),
				Member: fmt.Sprintf("%s,%d", util.FastRandomAlphaNumberString(32), 1024),
			})
		}
		// 模拟 Redis 中的上传文件缓存信息。
		bs, _ := json.Marshal(map[string]any{"user": LoginUser.ID, "timeSecond": time.Now().Unix(), "size": 3 * 1024})

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		tusdMocker := MockTusdClient(ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)            // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                   // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                              // 查询数据库登录用户信息。
		redisMocker = redisMocker.HGetOnce(string(bs), nil)                               // 获取上传文件缓存信息。
		redisMocker = redisMocker.SetNXOnce(true, nil)                                    // 获取分片合并锁。
		redisMocker = redisMocker.ZRangeWithScores(redisZ, nil)                           // 查询已上传分片列表及大小。
		tusdMocker = tusdMocker.MergePartsOnce(util.FastRandomAlphaNumberString(32), nil) // 通知存储服务合并分片。
		redisMocker = redisMocker.ZAddOnce(1, nil)                                        // 记录分片合并结果。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                  // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.EvalOnce(true, nil)                                     // 释放分片合并锁。
		redisMocker = redisMocker.HDel(1, nil)                                            // 清除上传文件缓存信息。
		redisMocker = redisMocker.ZRem(1, nil)                                            // 清除分片缓存记录。
		dbFileMocker = dbFileMocker.CreateOnce(nil)                                       // 保存文件记录到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer tusdMocker.Reset()
		defer dbFileMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebMergePartsReq{FileID: fileID})),
			consts.AlertFileUpload,
		)
	})

	validateErrorRequest := func(t *testing.T, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebMergePartsReq{FileID: fileID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name   string
		FileID string
	}{
		{"文件 ID 缺失", ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(37)},
		{"文件 ID 字符非法", "汉" + util.FastRandomAlphaNumberString(37)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID)
		})
	}
}

func TestFileAPIDownload(t *testing.T) {
	const reqPath = "/api/file/download"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		fileData := GenerateBytes(4096)
		fileName := util.FastRandomAlphaNumberString(16)
		mockFile := &model.File{
			FileID: fileID,
			TusdID: util.FastRandomAlphaNumberString(32),
			AppID:  AppInfo.ID,
			Name:   fileName,
			Size:   len(fileData),
			Type:   model.FileTypeAndroidSigning,
		} // 模拟数据库中的文件记录（安卓签名类型）。
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(fileData)),
			ContentLength: len(fileData),
		} // 模拟存储服务返回的文件下载结果。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库中文件信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验文件下载权限。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                    // 从存储服务获取文件内容。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()
		defer tusdMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.FileAPIDownloadReq{FileID: fileID}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs, _ := io.ReadAll(rsp.Body)
		if len(bs) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.FileAPIDownloadReq{FileID: fileID}))

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
		Name   string
		FileID string
	}{
		{"文件 ID 缺失", ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(37)},
		{"文件 ID 字符非法", "汉" + util.FastRandomAlphaNumberString(37)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID)
		})
	}
}

func TestFileAPIInitial(t *testing.T) {
	const reqPath = "/api/file/initial"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		initName := "test.apk"
		initSize := int64(1024)
		initMD5 := "01234567890123456789012345678901"
		initType := model.FileTypeAndroidSigning

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验文件上传权限。
		dbFileMocker = dbFileMocker.TakeOnce(nil, nil)                         // 检查文件是否已存在（秒传检测）。
		redisMocker = redisMocker.SAddOnce(1, nil)                             // Redis 生成文件 ID。
		redisMocker = redisMocker.HSetOnce(1, nil)                             // 缓存文件上传信息到 Redis。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileAPIInitialRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.FileAPIInitialReq{
					Name: initName,
					Size: initSize,
					MD5:  initMD5,
					Type: initType,
				})),
			0,
		)

		if rspBodyObj.Data.Exist != false {
			t.Errorf("file existing is not expected, %v", rspBodyObj.Data.Exist)
		}
		if len(rspBodyObj.Data.FileID) != 38 {
			t.Errorf("file id length is not 38, %v", rspBodyObj.Data.FileID)
		}
	})

	t.Run("正常测试_文件存在", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		initName := "test.apk"
		initSize := int64(1024)
		initMD5 := "01234567890123456789012345678901"
		initType := model.FileTypeAndroidSigning
		mockFile := &model.File{ID: 1, FileID: fileID} // 模拟数据库中已存在的文件记录。

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验文件上传权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 检查文件已存在（命中秒传）。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer dbFileMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileAPIInitialRsp](
			t,
			ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
				CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
				&protocol.FileAPIInitialReq{
					Name: initName,
					Size: initSize,
					MD5:  initMD5,
					Type: initType,
				})),
			0,
		)

		if rspBodyObj.Data.Exist != true {
			t.Errorf("file existing is expected, %v", rspBodyObj.Data.Exist)
		}
		if rspBodyObj.Data.FileID != fileID {
			t.Errorf("expect file id %s, but got %s", fileID, rspBodyObj.Data.FileID)
		}
	})

	validateErrorRequest := func(t *testing.T, name, md5 string, typ, size int) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIPostJSONRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			&protocol.FileAPIInitialReq{
				Name: name,
				Size: int64(size),
				MD5:  md5,
				Type: typ,
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
		Name       string
		Name2, MD5 string
		Type, Size int
	}{
		{"文件名缺失", "", "01234567890123456789012345678901", model.FileTypeAndroidSigning, 1024},
		{"文件名过长", util.FastRandomAlphaNumberString(257), "01234567890123456789012345678901", model.FileTypeAndroidSigning, 1024},
		{"文件大小缺失", "test.apk", "01234567890123456789012345678901", model.FileTypeAndroidSigning, 0},
		{"文件 MD5 非法", "test.apk", "0123456789012345678901234567890", model.FileTypeAndroidSigning, 1024},
		{"文件类型缺失", "test.apk", "01234567890123456789012345678901", 0, 1024},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2, v.MD5, v.Type, v.Size)
		})
	}
}

func TestFileAPIUploadPart(t *testing.T) {
	const reqPath = "/api/file/uploadPart"
	createRequestBody := func(fileID *string, chunkNumber *int, fileData []byte) (io.Reader, string) {
		reqBody := &bytes.Buffer{}
		reqBodyWriter := multipart.NewWriter(reqBody)
		if fileID != nil {
			err := reqBodyWriter.WriteField("fileId", *fileID)
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if chunkNumber != nil {
			if err := reqBodyWriter.WriteField("chunkNumber", strconv.Itoa(*chunkNumber)); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if fileData != nil {
			avatar, err := reqBodyWriter.CreateFormFile("chunk", "chunk")
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
			written, err := avatar.Write(fileData)
			if err != nil {
				t.Error("write avatar file error", err)
				return nil, ""
			}
			if written != len(fileData) {
				t.Error("the number of bytes written does not meet the expectations", written, len(fileData))
				return nil, ""
			}
		}
		if err := reqBodyWriter.Close(); err != nil {
			t.Error("failed to close multipart writer", err)
			return nil, ""
		}
		return reqBody, reqBodyWriter.FormDataContentType()
	}

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		uploadFileID := new(util.FastRandomAlphaNumberString(38))
		chunkNumber := new(1)
		chunkData := GenerateBytes(4096)
		bs, _ := json.Marshal(map[string]any{"accountId": APIAccount.ID, "timeSecond": time.Now().Unix()})

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                      // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                      // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                     // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)                 // 校验 API 凭证权限。
		redisMocker = redisMocker.HGetOnce(string(bs), nil)                                   // 获取上传文件缓存信息。
		redisMocker = redisMocker.SetNXOnce(true, nil)                                        // 获取分片上传锁。
		redisMocker = redisMocker.ZRangeArgsOnce(nil, nil)                                    // 查询已上传分片列表。
		tusdMocker = tusdMocker.UploadPartByIOOnce(util.FastRandomAlphaNumberString(32), nil) // 上传分片到存储服务。
		redisMocker = redisMocker.ZAddOnce(1, nil)                                            // 记录分片上传信息。
		redisMocker = redisMocker.EvalOnce(true, nil)                                         // 释放分片上传锁。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer tusdMocker.Reset()

		reqBody, contentType := createRequestBody(uploadFileID, chunkNumber, chunkData)
		rsp := ServeHTTP(ctx, CreateAPIPostMultiFormRequest(ctx, reqPath, reqBody, contentType))

		if rsp.Code != http.StatusNoContent {
			t.Errorf("expect http code %d, but got %d", http.StatusNoContent, rsp.Code)
		}
	})

	validateErrorRequest := func(t *testing.T, fileID *string, chunkNumber *int, fileData []byte) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		reqBody, contentType := createRequestBody(fileID, chunkNumber, fileData)
		rsp := ServeHTTP(ctx, CreateAPIPostMultiFormRequest(ctx, reqPath, reqBody, contentType))

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
		FileID      *string
		ChunkNumber *int
		FileData    []byte
	}{
		{"分片缺失", new(util.FastRandomAlphaNumberString(38)), new(1), nil},
		{"分片大小为零", new(util.FastRandomAlphaNumberString(38)), new(1), []byte{}},
		{"文件 ID 错误", new(util.FastRandomAlphaNumberString(37)), new(1), GenerateBytes(4096)},
		{"文件 ID 字符非法", new("汉" + util.FastRandomAlphaNumberString(37)), new(1), GenerateBytes(4096)},
		{"分片序号错误", new(util.FastRandomAlphaNumberString(38)), new(-1), GenerateBytes(4096)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID, v.ChunkNumber, v.FileData)
		})
	}
}

func TestFileAPIMergeParts(t *testing.T) {
	const reqPath = "/api/file/mergeParts"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		fileID := util.FastRandomAlphaNumberString(38)
		redisZ := make([]redis.Z, 0, 3)
		for i := range cap(redisZ) {
			redisZ = append(redisZ, redis.Z{
				Score:  float64(i + 1),
				Member: fmt.Sprintf("%s,%d", util.FastRandomAlphaNumberString(32), 1024),
			})
		}
		bs, _ := json.Marshal(map[string]any{"accountId": APIAccount.ID, "timeSecond": time.Now().Unix(), "size": 3 * 1024})

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		tusdMocker := MockTusdClient(ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)            // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                  // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                  // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)                 // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)             // 校验 API 凭证权限。
		redisMocker = redisMocker.HGetOnce(string(bs), nil)                               // 获取上传文件缓存信息。
		redisMocker = redisMocker.SetNXOnce(true, nil)                                    // 获取分片合并锁。
		redisMocker = redisMocker.ZRangeWithScores(redisZ, nil)                           // 查询已上传分片列表及大小。
		tusdMocker = tusdMocker.MergePartsOnce(util.FastRandomAlphaNumberString(32), nil) // 通知存储服务合并分片。
		redisMocker = redisMocker.ZAddOnce(1, nil)                                        // 记录分片合并结果。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                  // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.EvalOnce(true, nil)                                     // 释放分片合并锁。
		redisMocker = redisMocker.HDel(1, nil)                                            // 清除上传文件缓存信息。
		redisMocker = redisMocker.ZRem(1, nil)                                            // 清除分片缓存记录。
		dbFileMocker = dbFileMocker.CreateOnce(nil)                                       // 保存文件记录到数据库。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()
		defer tusdMocker.Reset()
		defer dbFileMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.FileAPIMergePartsReq{FileID: fileID}))

		if rsp.Code != http.StatusNoContent {
			t.Errorf("expect http code %d, but got %d", http.StatusNoContent, rsp.Code)
		}
	})

	validateErrorRequest := func(t *testing.T, fileID string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbAPIAccountMocker := MockDBClient[model.APIAccount](ctx)
		dbAPIAuthorizationMocker := MockDBClient[model.APIAuthorization](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行 API 请求限流脚本。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbAPIAccountMocker = dbAPIAccountMocker.TakeOnce(APIAccount, nil)      // 查询数据库 API 凭证信息。
		dbAPIAuthorizationMocker = dbAPIAuthorizationMocker.CountOnce(1, nil)  // 校验 API 凭证权限。
		defer redisMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbAPIAccountMocker.Reset()
		defer dbAPIAuthorizationMocker.Reset()

		rsp := ServeHTTP(ctx, CreateAPIGetRequest(ctx, reqPath,
			CreateAPIAuthorization(AppInfo.AppID, APIAccount.AccountID, APIAccount.Secret),
			protocol.FileAPIMergePartsReq{FileID: fileID}))

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
		Name   string
		FileID string
	}{
		{"文件 ID 缺失", ""},
		{"文件 ID 错误", util.FastRandomAlphaNumberString(37)},
		{"文件 ID 字符非法", "汉" + util.FastRandomAlphaNumberString(37)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID)
		})
	}
}
