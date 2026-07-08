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
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

func TestAppAPI_WebRegister(t *testing.T) {
	const reqPath = "/web/app/register"
	createRequestBody := func(
		name, logoName *string,
		platform *int,
		admins, members []string,
		logoData []byte,
	) (io.Reader, string) {
		reqBody := &bytes.Buffer{}
		reqBodyWriter := multipart.NewWriter(reqBody)
		if name != nil {
			err := reqBodyWriter.WriteField("name", *name)
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if platform != nil {
			if err := reqBodyWriter.WriteField("platform", strconv.Itoa(*platform)); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		for _, v := range admins {
			if err := reqBodyWriter.WriteField("admins", v); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		for _, v := range members {
			if err := reqBodyWriter.WriteField("members", v); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if logoName != nil {
			avatar, err := reqBodyWriter.CreateFormFile("logo", *logoName)
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
			written, err := avatar.Write(logoData)
			if err != nil {
				t.Error("write avatar file error", err)
				return nil, ""
			}
			if written != len(logoData) {
				t.Error("the number of bytes written does not meet the expectations", written, len(logoData))
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
		name, logoName, platform, admins, members, logoData := new("应用名"), new("logo_name.png"),
			TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)

		defer MockDBClient[model.App](ctx).
			CreateOnce(nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
			SAddOnce(1, nil).
			Reset()
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		defer MockDBClient[model.User](ctx).
			FindOnce(userInfos, nil).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			CreateInBatchesOnce(nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()
		defer MockTusdClient(ctx).
			UploadFileOnce(util.FastRandomAlphaNumberString(32), nil).
			Reset()

		reqBody, contentType := createRequestBody(name, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.AlertRegisterApp,
		)
	})

	t.Run("异常测试_图标文件过大", func(t *testing.T) {
		ctx := context.Background()
		name, logoName, platform, admins, members, logoData := new("应用名"), new("logo_name.png"),
			TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)

		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("Server").
			FieldByName("AppLogoMaximumSize").
			Set(1).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()

		reqBody, contentType := createRequestBody(name, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrAppLogoTooLarge,
		)
	})

	t.Run("异常测试_图标文件格式非法", func(t *testing.T) {
		ctx := context.Background()
		name, logoName, platform, admins, members, logoData := new("应用名"), new("logo_name.png"),
			TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GenerateGIF(t, 10, 10)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()

		reqBody, contentType := createRequestBody(name, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrAppLogoFormatNotSupported,
		)
	})

	t.Run("异常测试_用户不存在", func(t *testing.T) {
		ctx := context.Background()
		name, logoName, platform, admins, members, logoData := new("应用名"), new("logo_name.png"),
			TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		if len(userInfos) > 0 {
			userInfos = userInfos[:len(userInfos)-1]
		}
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce(userInfos, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()

		reqBody, contentType := createRequestBody(name, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrUserNotExists,
		)
	})

	validateErrorRequest := func(
		t *testing.T,
		name, logoName *string,
		platform *int,
		admins, members []string,
		logoData []byte,
	) {
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

		reqBody, contentType := createRequestBody(name, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name            string
		Name2, LogoName *string
		Platform        *int
		Admins, Members []string
		LogoData        []byte
	}{
		{"应用名缺失", nil, new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"应用名字符过多", new(util.FastRandomAlphaNumberString(65)), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"应用图标缺失", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"b", "c"}, nil},
		{"应用平台非法", new("应用名"), new("logo_name.png"), new(0), []string{"a", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"管理员字符非法", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"张", "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"管理员非法", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{util.FastRandomAlphaNumberString(33), "d"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"管理员重复", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "a"}, []string{"b", "c"}, GeneratePNG(t, 10, 10)},
		{"成员字符非法", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"张", "c"}, GeneratePNG(t, 10, 10)},
		{"成员非法", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{util.FastRandomAlphaNumberString(33), "c"}, GeneratePNG(t, 10, 10)},
		{"成员重复", new("应用名"), new("logo_name.png"), TakeIntPtr(model.AppPlatformWindows), []string{"a", "d"}, []string{"c", "c"}, GeneratePNG(t, 10, 10)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(
				t,
				v.Name2,
				v.LogoName,
				v.Platform,
				v.Admins,
				v.Members,
				v.LogoData,
			)
		})
	}
}

func TestAppAPI_WebUpdate(t *testing.T) {
	const reqPath = "/web/app/update"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		name, logoFileID, admins, members, logo :=
			"应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)

		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			SRemOnce(1, nil).
			Reset()
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce(userInfos, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			TakeOnce(&model.File{
				ID:     1,
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(logo)),
				ContentLength: len(logo),
			}, nil).
			DeleteFileOnce(nil).
			Reset()
		var hasSignerRoleUserIDs []int
		if len(userInfos) > 0 {
			hasSignerRoleUserIDs = []int{userInfos[0].ID}
		}
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			ScanOnce(func(v any) { *v.(*[]int) = hasSignerRoleUserIDs }, nil).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			CreateInBatchesOnce(nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("异常测试_应用状态非法", func(t *testing.T) {
		ctx := context.Background()
		name, logoFileID, admins, members, logo :=
			"应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Status").
			Set(model.AppStatusInvalid).
			Reset()
		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
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
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(logo)),
				ContentLength: len(logo),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.ErrAppStatusNotValid,
		)
	})

	t.Run("异常测试_应用图标文件过大", func(t *testing.T) {
		ctx := context.Background()
		name, logoFileID, admins, members, logo :=
			"应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)

		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("Server").
			FieldByName("AppLogoMaximumSize").
			Set(1).
			Reset()
		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
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
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(logo)),
				ContentLength: len(logo),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.ErrAppLogoTooLarge,
		)
	})

	t.Run("异常测试_应用图标文件格式不支持", func(t *testing.T) {
		ctx := context.Background()
		name, logoFileID, admins, members, logo :=
			"应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}, GenerateGIF(t, 10, 10)

		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
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
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(logo)),
				ContentLength: len(logo),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.ErrAppLogoFormatNotSupported,
		)
	})

	t.Run("异常测试_用户不存在", func(t *testing.T) {
		ctx := context.Background()
		name, logoFileID, admins, members, logo :=
			"应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)

		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			Reset()
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		if len(userInfos) > 0 {
			userInfos = userInfos[:len(userInfos)-1]
		}
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce(userInfos, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: util.FastRandomAlphaNumberString(38),
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Type:   model.FileTypeAppLogo,
			}, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(logo)),
				ContentLength: len(logo),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.ErrUserNotExists,
		)
	})

	validateErrorRequest := func(t *testing.T, name, logoFileID string, admins, members []string) {
		ctx := context.Background()
		defer MockRedis(ctx).
			GetOnce(Session, nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       name,
				LogoFileID: logoFileID,
				Admins:     admins,
				Members:    members,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name              string
		Name2, LogoFileID string
		Admins, Members   []string
	}{
		{"应用名缺失", "", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}},
		{"应用名过长", util.FastRandomAlphaNumberString(65), util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "d"}},
		{"应用图标ID缺失", "应用名", "", []string{"a", "b"}, []string{"c", "d"}},
		{"应用图标ID非法", "应用名", util.FastRandomAlphaNumberString(32), []string{"a", "b"}, []string{"c", "d"}},
		{"应用图标ID字符非法", "应用名", util.FastRandomAlphaNumberString(37) + "汉", []string{"a", "b"}, []string{"c", "d"}},
		{"重复的管理员", "应用名", util.FastRandomAlphaNumberString(38), []string{"a", "a"}, []string{"c", "d"}},
		{"管理员字符非法", "应用名", util.FastRandomAlphaNumberString(38), []string{"a汉", "b"}, []string{"c", "d"}},
		{"管理员字符过长", "应用名", util.FastRandomAlphaNumberString(38), []string{"a" + util.FastRandomAlphaNumberString(32), "b"}, []string{"c", "d"}},
		{"重复的成员", "应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c", "c"}},
		{"成员字符非法", "应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c汉", "d"}},
		{"成员字符过长", "应用名", util.FastRandomAlphaNumberString(38), []string{"a", "b"}, []string{"c" + util.FastRandomAlphaNumberString(32), "d"}},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2, v.LogoFileID, v.Admins, v.Members)
		})
	}
}

func TestAppAPI_WebGetInformation(t *testing.T) {
	const reqPath = "/web/app/getInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		admins, adminIDs, members, memberIDs := []string{"a", "b"}, []int{1, 2}, []string{"c", "d"}, []int{3, 4}

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for i, v := range admins {
			userInfos = append(userInfos, &model.User{ID: adminIDs[i], NameEn: v})
		}
		for i, v := range members {
			userInfos = append(userInfos, &model.User{ID: memberIDs[i], NameEn: v})
		}
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce(userInfos, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		userRoleInfos := make([]*model.UserRole, 0, len(admins)+len(members))
		for _, v := range adminIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{UserID: v, Role: model.UserRoleAppAdmin})
		}
		for _, v := range memberIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{UserID: v, Role: model.UserRoleAppMember})
		}
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			FindOnce(userRoleInfos, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppWebGetInformationRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)

		if rspBodyObj.Data.Name != AppInfo.Name {
			t.Errorf("name not equal, want %s, got %s", AppInfo.Name, rspBodyObj.Data.Name)
		}
		if rspBodyObj.Data.AppID != AppInfo.AppID {
			t.Errorf("app id not equal, want %v, got %v", AppInfo.AppID, rspBodyObj.Data.AppID)
		}
		if rspBodyObj.Data.LogoFileID != AppInfo.LogoFileID {
			t.Errorf("logo file id not equal, want %v, got %v", AppInfo.LogoFileID, rspBodyObj.Data.LogoFileID)
		}
		if rspBodyObj.Data.Status != AppInfo.Status {
			t.Errorf("status not equal, want %v, got %v", AppInfo.Status, rspBodyObj.Data.Status)
		}
		admins2 := util.MapToList(rspBodyObj.Data.Admins, func(k, v string) string { return k })
		sort.Strings(admins2)
		sort.Strings(admins)
		if !reflect.DeepEqual(admins2, admins) {
			t.Errorf("admins not equal, want %v, got %v", admins, rspBodyObj.Data.Admins)
		}
		members2 := util.MapToList(rspBodyObj.Data.Members, func(k, v string) string { return k })
		sort.Strings(members2)
		sort.Strings(members)
		if !reflect.DeepEqual(members2, members) {
			t.Errorf("members not equal, want %v, got %v", members, rspBodyObj.Data.Admins)
		}
	})
}

func TestAppAPI_WebInvalidate(t *testing.T) {
	const reqPath = "/web/app/invalidate"

	t.Run("正常测试", func(t *testing.T) {
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
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.AlertSuccess,
		)
	})
}

func TestAppAPI_WebEnable(t *testing.T) {
	const reqPath = "/web/app/enable"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Status").
			Set(model.AppStatusInvalid).
			Reset()
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
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(nil, nil).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.AlertEnableApp,
		)
	})
}

func TestAppAPI_WebCount(t *testing.T) {
	const reqPath = "/web/app/count"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		count := 3

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*int) = count }, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppWebCountRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.Count != count {
			t.Errorf("want %d, got %d", count, rspBodyObj.Data.Count)
		}
	})
}

func TestAppAPI_WebSearch(t *testing.T) {
	const reqPath = "/web/app/search"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		name, platform, status := "", []int{}, []int{}

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(0, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.App{AppInfo}, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppWebSearchRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AppWebSearchReq{
				Name:     name,
				Platform: platform,
				Status:   status,
			})),
			0,
		)

		if rspBodyObj.Data.List[0].ID != AppInfo.AppID {
			t.Errorf("want %v, got %v", AppInfo.AppID, rspBodyObj.Data.List[0].ID)
		}
		if rspBodyObj.Data.List[0].Name != AppInfo.Name {
			t.Errorf("want %s, got %s", AppInfo.Name, rspBodyObj.Data.List[0].Name)
		}
		if rspBodyObj.Data.List[0].Status != AppInfo.Status {
			t.Errorf("want %v, got %v", AppInfo.Status, rspBodyObj.Data.List[0].Status)
		}
		if rspBodyObj.Data.List[0].Platform != AppInfo.Platform {
			t.Errorf("want %v, got %v", AppInfo.Platform, rspBodyObj.Data.List[0].Platform)
		}
	})

	validateErrorRequest := func(t *testing.T, name string, platform, status []int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.AppWebSearchRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AppWebSearchReq{
				Name:     name,
				Platform: platform,
				Status:   status,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name             string
		Name2            string
		Platform, Status []int
	}{
		{"应用名过长", util.FastRandomAlphaNumberString(65), []int{}, []int{}},
		{"平台重复", "", []int{1, 1}, []int{}},
		{"平台枚举非法", "", []int{4}, []int{}},
		{"状态重复", "", []int{}, []int{1, 1}},
		{"状态枚举非法", "", []int{}, []int{5}},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2, v.Platform, v.Status)
		})
	}
}
