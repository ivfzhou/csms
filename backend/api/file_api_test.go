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

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
	"github.com/redis/go-redis/v9"
)

func TestFileAPI_WebDownload(t *testing.T) {
	const reqPath = "/web/file/download"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		fileID, typ, fileData, fileName := util.FastRandomAlphaNumberString(38), model.FileTypeUserAvatar,
			GenerateBytes(4096), util.FastRandomAlphaNumberString(16)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: fileID,
				TusdID: util.FastRandomAlphaNumberString(32),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   fileName,
				Md5:    util.FastRandomAlphaNumberString(32),
				Size:   len(fileData),
				Type:   typ,
			}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(fileData)),
				ContentLength: len(fileData),
			}, nil).
			Reset()

		rspBody, fileName2 := CheckAndReadBody(
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebDownloadReq{
				FileID: fileID,
			})),
		)

		if !reflect.DeepEqual(rspBody, fileData) {
			t.Errorf("file data is not equal, %v %v", len(fileData), len(rspBody))
		}
		if !strings.Contains(fileName2, fileName) {
			t.Errorf("file name is not equal, %s %v", fileName2, fileName)
		}
	})

	validateErrorRequest := func(t *testing.T, fileID string, typ int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebDownloadReq{
				FileID: fileID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name   string
		FileID string
		Type   int
	}{
		{"文件ID缺失", "", model.FileTypeUserAvatar},
		{"文件ID过短", util.FastRandomAlphaNumberString(37), model.FileTypeUserAvatar},
		{"文件ID非法", util.FastRandomAlphaNumberString(37) + "汉", model.FileTypeUserAvatar},
		{"文件类型缺失", util.FastRandomAlphaNumberString(38), 0},
		{"文件类型非法", util.FastRandomAlphaNumberString(38), -9999},
	} {
		t.Run("异常请求_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID, v.Type)
		})
	}
}

func TestFileAPI_WebInitial(t *testing.T) {
	const reqPath = "/web/file/initial"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		name, size, md5, typ := "文件名.jpg", 1024, "01234567890123456789012345678901", model.FileTypeUserAvatar

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
			HSetOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(nil, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID, &protocol.FileWebInitialReq{
				Name: name,
				Size: int64(size),
				MD5:  md5,
				Type: typ,
			})),
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

		name, size, md5, typ, fileID := "文件名.jpg", 1024, "01234567890123456789012345678901", model.FileTypeUserAvatar,
			util.FastRandomAlphaNumberString(38)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: fileID,
			}, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID, &protocol.FileWebInitialReq{
				Name: name,
				Size: int64(size),
				MD5:  md5,
				Type: typ,
			})),
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
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.FileWebInitialRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath+"?"+consts.HTTPPathAppID+"="+AppInfo.AppID, &protocol.FileWebInitialReq{
				Name: name,
				Size: int64(size),
				MD5:  md5,
				Type: typ,
			})),
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
		{"文件名缺失", "文件名.jpg", "01234567890123456789012345678901", model.FileTypeUserAvatar, 0},
		{"文件MD5非法", "文件名.jpg", "0123456789012345678901234567890", model.FileTypeUserAvatar, 1024},
		{"文件类型非法", "文件名.jpg", "01234567890123456789012345678901", 0, 1024},
	} {
		t.Run("异常请求_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2, v.MD5, v.Type, v.Size)
		})
	}
}

func TestFileAPI_WebUploadPart(t *testing.T) {
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
		fileID, chunkNumber, filePart := util.FastRandomAlphaNumberString(38), 1, GenerateBytes(4096)

		bs, _ := json.Marshal(map[string]any{"user": LoginUser.ID, "timeSecond": time.Now().Unix()})
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			HGetOnce(string(bs), nil).
			SetNXOnce(true, nil).
			ZRangeByScoreOnce(nil, nil).
			ZAddOnce(1, nil).
			EvalOnce(true, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockTusdClient(ctx).
			UploadPartByIOOnce(util.FastRandomAlphaNumberString(32), nil).
			Reset()

		reqBody, contentType := createRequestBody(&fileID, &chunkNumber, filePart)
		CheckAndUnmarshalBody[any](t, ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)), 0)
	})

	validateErrorRequest := func(t *testing.T, fileID *string, chunkNumber *int, fileData []byte) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
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
		{"分片缺失", TakeStringPtr(util.FastRandomAlphaNumberString(38)), TakeIntPtr(1), nil},
		{"分片大小为零", TakeStringPtr(util.FastRandomAlphaNumberString(38)), TakeIntPtr(1), []byte{}},
		{"文件ID错误", TakeStringPtr(util.FastRandomAlphaNumberString(37)), TakeIntPtr(1), GenerateBytes(4096)},
		{"文件ID字符非法", TakeStringPtr("汉" + util.FastRandomAlphaNumberString(37)), TakeIntPtr(1), GenerateBytes(4096)},
		{"分片序号错误", TakeStringPtr(util.FastRandomAlphaNumberString(38)), TakeIntPtr(-1), GenerateBytes(4096)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID, v.ChunkNumber, v.FileData)
		})
	}
}

func TestFileAPI_WebMergeParts(t *testing.T) {
	const reqPath = "/web/file/mergeParts"

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
		bs, _ := json.Marshal(map[string]any{"user": LoginUser.ID, "timeSecond": time.Now().Unix(), "size": 3 * 1024})
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			HGetOnce(string(bs), nil).
			SetNXOnce(true, nil).
			ZRangeWithScores(redisZ, nil).
			ZAddOnce(1, nil).
			EvalshaOnce(true, nil).
			EvalOnce(true, nil).
			HDel(1, nil).
			ZRem(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockTusdClient(ctx).
			MergePartsOnce(util.FastRandomAlphaNumberString(32), nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.FileWebMergePartsReq{FileID: fileID})),
			consts.AlertFileUpload,
		)
	})

	validateErrorRequest := func(t *testing.T, fileID string) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			EvalshaOnce(true, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
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
		{"文件ID缺失", ""},
		{"文件ID错误", util.FastRandomAlphaNumberString(37)},
		{"文件ID字符非法", "汉" + util.FastRandomAlphaNumberString(37)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.FileID)
		})
	}
}
