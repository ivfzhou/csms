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

func TestAppWebRegister(t *testing.T) {
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
		admins, members := []string{"a", "d"}, []string{"b", "c"}
		appName := new("应用名")
		logoName := new("logo_name.png")
		platform := TakeIntPtr(model.AppPlatformWindows)
		logoData := GeneratePNG(t, 10, 10)
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		tusdMocker := MockTusdClient(ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                    // 加载 Redis 限流脚本。       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中的系统管理员。
		dbUserMocker = dbUserMocker.FindOnce(userInfos, nil)                                      // 查询数据库中人员信息。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                // 添加应用 ID 到 Redis。
		redisMocker = redisMocker.SAddOnce(1, nil)                                                // 添加文件 ID 到 Redis。
		tusdMocker = tusdMocker.UploadFileOnce(util.FastRandomAlphaNumberString(32), nil)         // 上传应用图标文件到存储服务。
		dbFileMocker = dbFileMocker.CreateOnce(nil)                                               // 保存应用图标文件信息到数据库。
		dbTodoMocker = dbTodoMocker.CreateOnce(nil)                                               // 保存应用审批待办到数据库。
		dbAppMocker = dbAppMocker.CreateOnce(nil)                                                 // 保存应用信息到数据库。
		dbUserRoleMocker = dbUserRoleMocker.CreateInBatchesOnce(nil)                              // 添加数据库应用成员记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                             // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer tusdMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbEventMocker.Reset()

		reqBody, contentType := createRequestBody(appName, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.AlertRegisterApp,
		)
	})

	t.Run("异常测试_图标文件过大", func(t *testing.T) {
		ctx := context.Background()
		appName := new("应用名")
		logoName := new("logo_name.png")
		platform := TakeIntPtr(model.AppPlatformWindows)
		admins, members := []string{"a", "d"}, []string{"b", "c"}
		logoData := GeneratePNG(t, 10, 10)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appLogoMaximumSizeReset := mvt.Chain(cfg.Get()).Elem().FieldByName("BackendConfiguration").FieldByName("AppLogoMaximumSizeValue").Set(1)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                    // 加载 Redis 限流脚本。       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中的系统管理员。
		defer appLogoMaximumSizeReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		reqBody, contentType := createRequestBody(appName, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrAppLogoTooLarge,
		)
	})

	t.Run("异常测试_图标文件格式非法", func(t *testing.T) {
		ctx := context.Background()
		appName := new("应用名")
		logoName := new("logo_name.png")
		platform := TakeIntPtr(model.AppPlatformWindows)
		admins, members := []string{"a", "d"}, []string{"b", "c"}
		logoData := GenerateGIF(t, 10, 10)

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                    // 加载 Redis 限流脚本。       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中的系统管理员。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		reqBody, contentType := createRequestBody(appName, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrAppLogoFormatNotSupported,
		)
	})

	t.Run("异常测试_用户不存在", func(t *testing.T) {
		ctx := context.Background()
		admins, members := []string{"a", "d"}, []string{"b", "c"}
		appName := new("应用名")
		logoName := new("logo_name.png")
		platform := TakeIntPtr(model.AppPlatformWindows)
		logoData := GeneratePNG(t, 10, 10)
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

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                    // 加载 Redis 限流脚本。       // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                      // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中的系统管理员。
		dbUserMocker = dbUserMocker.FindOnce(userInfos, nil)                                      // 查询数据库中人员信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

		reqBody, contentType := createRequestBody(appName, logoName, platform, admins, members, logoData)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrUserNotExists,
		)
	})

	validateErrorRequest := func(t *testing.T, name, logoName *string, platform *int, admins, members []string, logoData []byte) {
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
			validateErrorRequest(t, v.Name2, v.LogoName, v.Platform, v.Admins, v.Members, v.LogoData)
		})
	}
}

func TestAppWebSearch(t *testing.T) {
	const reqPath = "/web/app/search"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		searchName := ""
		searchPlatform := []int{}
		searchStatus := []int{}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)              // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)          // 加载 Redis 限流脚本。 // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                 // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                            // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(0, nil)                           // 校验是否为系统管理员。
		dbAppMocker = dbAppMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询用户有权限的应用 IDs。
		dbAppMocker = dbAppMocker.FindOnce([]*model.App{AppInfo}, nil)                  // 查询数据库应用信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbAppMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AppWebSearchRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AppWebSearchReq{
				Name:     searchName,
				Platform: searchPlatform,
				Status:   searchStatus,
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

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

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

func TestAppWebUpdate(t *testing.T) {
	const reqPath = "/web/app/update"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		admins, members, logo := []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)
		updateName := "应用名" // 更新应用请求参数。
		updateLogoFileID := util.FastRandomAlphaNumberString(38)
		mockNewFile := &model.File{
			ID:     1,
			FileID: util.FastRandomAlphaNumberString(38),
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Type:   model.FileTypeAppLogo,
		} // 模拟数据库中的应用图标文件记录（新文件）。
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(logo)),
			ContentLength: len(logo),
		} // 模拟存储服务返回的文件下载结果。
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1} // 模拟数据库操作影响行数。
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for _, v := range admins {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		for _, v := range members {
			userInfos = append(userInfos, &model.User{ID: rand.Intn(9999), NameEn: v})
		}
		var hasSignerRoleUserIDs []int
		if len(userInfos) > 0 {
			hasSignerRoleUserIDs = []int{userInfos[0].ID}
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		tusdMocker := MockTusdClient(ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                                    // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)                                // 加载 Redis 限流脚本。                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                                      // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                                       // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                                  // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                                      // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                                 // 校验应用管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(userInfos, nil)                                                  // 查询数据库中人员信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockNewFile, nil)                                                // 查询数据库中应用图标文件信息。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                                                   // 从存储服务下载应用图标文件。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)                               // 更新数据库应用信息。
		dbFileMocker = dbFileMocker.TakeOnce(mockNewFile, nil)                                                // 查询数据库中旧的应用图标文件信息。
		dbFileMocker = dbFileMocker.DeleteOnce(mockRowsAffected, nil)                                         // 删除数据库中旧的应用图标文件信息。
		tusdMocker = tusdMocker.DeleteFileOnce(nil)                                                           // 删除存储服务中旧的应用图标文件。
		redisMocker = redisMocker.SRemOnce(1, nil)                                                            // 删除 Redis 中旧文件 ID。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = hasSignerRoleUserIDs }, nil) // 查询有签名权限的人员。
		dbUserRoleMocker = dbUserRoleMocker.DeleteOnce(mockRowsAffected, nil)                                 // 删除需要移除应用签名权限的人员。
		dbUserRoleMocker = dbUserRoleMocker.DeleteOnce(mockRowsAffected, nil)                                 // 删除所有应用成员。
		dbUserRoleMocker = dbUserRoleMocker.CreateInBatchesOnce(nil)                                          // 添加应用成员。
		dbTodoMocker = dbTodoMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)                             // 更新待办中的审批人。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                                         // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbFileMocker.Reset()
		defer tusdMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbTodoMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       updateName,
				LogoFileID: updateLogoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("异常测试_应用状态非法", func(t *testing.T) {
		ctx := context.Background()
		updateName := "应用名" // 更新应用请求参数。
		updateLogoFileID := util.FastRandomAlphaNumberString(38)
		updateAdmins, updateMembers := []string{"a", "b"}, []string{"c", "d"}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		appStatusReset := mvt.Chain(AppInfo).Elem().FieldByName("Status").Set(model.AppStatusInvalid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		defer appStatusReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       updateName,
				LogoFileID: updateLogoFileID,
				Admins:     updateAdmins,
				Members:    updateMembers,
			})),
			consts.ErrAppStatusNotValid,
		)
	})

	t.Run("异常测试_应用图标文件过大", func(t *testing.T) {
		ctx := context.Background()
		logo := GeneratePNG(t, 10, 10)
		updateName := "应用名" // 更新应用请求参数。
		updateLogoFileID := util.FastRandomAlphaNumberString(38)
		updateAdmins, updateMembers := []string{"a", "b"}, []string{"c", "d"}
		mockFile := &model.File{
			ID:     1,
			FileID: util.FastRandomAlphaNumberString(38),
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Type:   model.FileTypeAppLogo,
		}
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(logo)),
			ContentLength: len(logo),
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		tusdMocker := MockTusdClient(ctx)
		appLogoMaximumSizeReset := mvt.Chain(cfg.Get()).Elem().FieldByName("BackendConfiguration").FieldByName("AppLogoMaximumSizeValue").Set(1)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库应用图标文件信息。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                    // 从存储服务下载应用图标文件。
		defer appLogoMaximumSizeReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer tusdMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       updateName,
				LogoFileID: updateLogoFileID,
				Admins:     updateAdmins,
				Members:    updateMembers,
			})),
			consts.ErrAppLogoTooLarge,
		)
	})

	t.Run("异常测试_应用图标文件格式不支持", func(t *testing.T) {
		ctx := context.Background()
		logo := GenerateGIF(t, 10, 10)
		updateName := "应用名" // 更新应用请求参数。
		updateLogoFileID := util.FastRandomAlphaNumberString(38)
		updateAdmins, updateMembers := []string{"a", "b"}, []string{"c", "d"}
		mockFile := &model.File{
			ID:     1,
			FileID: updateLogoFileID,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Type:   model.FileTypeAppLogo,
		}
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(logo)),
			ContentLength: len(logo),
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库应用图标文件信息。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                    // 从存储服务下载应用图标文件。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbFileMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer tusdMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       updateName,
				LogoFileID: updateLogoFileID,
				Admins:     updateAdmins,
				Members:    updateMembers,
			})),
			consts.ErrAppLogoFormatNotSupported,
		)
	})

	t.Run("异常测试_用户不存在", func(t *testing.T) {
		ctx := context.Background()
		admins, members, logo := []string{"a", "b"}, []string{"c", "d"}, GeneratePNG(t, 10, 10)
		updateName := "应用名" // 更新应用请求参数。
		updateLogoFileID := util.FastRandomAlphaNumberString(38)
		mockFile := &model.File{
			ID:     1,
			FileID: updateLogoFileID,
			UserID: LoginUser.ID,
			AppID:  AppInfo.ID,
			Type:   model.FileTypeAppLogo,
		}
		mockTusResult := &tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(logo)),
			ContentLength: len(logo),
		}
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

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		dbFileMocker = dbFileMocker.TakeOnce(mockFile, nil)                    // 查询数据库应用图标文件信息。
		tusdMocker = tusdMocker.GetOnce(mockTusResult, nil)                    // 从存储服务下载应用图标文件。
		dbUserMocker = dbUserMocker.FindOnce(userInfos, nil)                   // 查询数据库中人员信息。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbFileMocker.Reset()
		defer tusdMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AppWebUpdateReq{
				Name:       updateName,
				LogoFileID: updateLogoFileID,
				Admins:     admins,
				Members:    members,
			})),
			consts.ErrUserNotExists,
		)
	})

	validateErrorRequest := func(t *testing.T, name, logoFileID string, admins, members []string) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                       // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

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
		{"应用图标 ID 缺失", "应用名", "", []string{"a", "b"}, []string{"c", "d"}},
		{"应用图标 ID 非法", "应用名", util.FastRandomAlphaNumberString(32), []string{"a", "b"}, []string{"c", "d"}},
		{"应用图标 ID 字符非法", "应用名", util.FastRandomAlphaNumberString(37) + "汉", []string{"a", "b"}, []string{"c", "d"}},
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

func TestAppWebGetInformation(t *testing.T) {
	const reqPath = "/web/app/getInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		admins, adminIDs, members, memberIDs := []string{"a", "b"}, []int{1, 2}, []string{"c", "d"}, []int{3, 4}
		userInfos := make([]*model.User, 0, len(admins)+len(members))
		for i, v := range admins {
			userInfos = append(userInfos, &model.User{ID: adminIDs[i], NameEn: v})
		}
		for i, v := range members {
			userInfos = append(userInfos, &model.User{ID: memberIDs[i], NameEn: v})
		}
		userRoleInfos := make([]*model.UserRole, 0, len(admins)+len(members))
		for _, v := range adminIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{UserID: v, Role: model.UserRoleAppAdmin})
		}
		for _, v := range memberIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{UserID: v, Role: model.UserRoleAppMember})
		}

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)     // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil) // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                        // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                   // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                       // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                  // 校验应用管理员权限。
		dbUserMocker = dbUserMocker.FindOnce(userInfos, nil)                   // 查询应用成员信息。
		dbUserRoleMocker = dbUserRoleMocker.FindOnce(userRoleInfos, nil)       // 查询应用的成员。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()

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

func TestAppWebInvalidate(t *testing.T) {
	const reqPath = "/web/app/invalidate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1} // 模拟数据库操作影响行数。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)      // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisRateLimitScriptSha, nil)  // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                        // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                         // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                    // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                        // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                   // 校验应用管理员权限。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil) // 更新数据库应用记录。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                           // 添加应用事件到数据库。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbEventMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.AlertSuccess,
		)
	})
}

func TestAppWebEnable(t *testing.T) {
	const reqPath = "/web/app/enable"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		mockRowsAffected := gen.ResultInfo{RowsAffected: 1} // 模拟数据库操作影响行数。

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbAppMocker := MockDBClient[model.App](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		dbEventMocker := MockDBClient[model.Event](ctx)
		dbTodoMocker := MockDBClient[model.Todo](ctx)
		appStatusReset := mvt.Chain(AppInfo).Elem().FieldByName("Status").Set(model.AppStatusInvalid)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                        // 加载 Redis 限流脚本。
		redisMocker = redisMocker.EvalshaOnce(true, nil)                                          // 执行防抖过滤 Redis Lua 脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                           // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                      // 查询数据库登录用户信息。
		dbAppMocker = dbAppMocker.TakeOnce(AppInfo, nil)                                          // 查询数据库应用信息。
		dbUserRoleMocker = dbUserRoleMocker.CountOnce(1, nil)                                     // 校验应用管理员权限。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil) // 查询数据库中的系统管理员。
		dbAppMocker = dbAppMocker.UpdateColumnSimpleOnce(mockRowsAffected, nil)                   // 更新数据库应用状态。
		dbTodoMocker = dbTodoMocker.TakeOnce(nil, nil)                                            // 检查应用是否存在启用待办。
		dbTodoMocker = dbTodoMocker.CreateOnce(nil)                                               // 创建应用启用审批待办。
		dbEventMocker = dbEventMocker.CreateOnce(nil)                                             // 添加应用事件到数据库。
		defer appStatusReset.Reset()
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbAppMocker.Reset()
		defer dbUserRoleMocker.Reset()
		defer dbEventMocker.Reset()
		defer dbTodoMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			consts.AlertEnableApp,
		)
	})
}

func TestAppWebCount(t *testing.T) {
	const reqPath = "/web/app/count"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		count := 3

		redisMocker := MockRedis(ctx)
		dbUserMocker := MockDBClient[model.User](ctx)
		dbUserRoleMocker := MockDBClient[model.UserRole](ctx)
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                   // 加载 Redis 防抖脚本。
		redisMocker = redisMocker.ScriptLoadOnce(RedisShakeScriptSha, nil)                   // 加载 Redis 限流脚本。
		redisMocker = redisMocker.GetOnce(Session, nil)                                      // 获取 Redis 用户会话数据。
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)                                 // 查询数据库登录用户信息。
		dbUserRoleMocker = dbUserRoleMocker.ScanOnce(func(v any) { *v.(*int) = count }, nil) // 查询用户应用数量。
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbUserRoleMocker.Reset()

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
