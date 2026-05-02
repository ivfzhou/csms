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
	"crypto/md5"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

func TestUserAPI_WebRegister(t *testing.T) {
	const reqPath = "/web/user/register"
	createRequestBody := func(nameZh, nameEn, password, passwordConfirmation, department, avatarName *string,
		fileData []byte) (io.Reader, string) {
		reqBody := &bytes.Buffer{}
		reqBodyWriter := multipart.NewWriter(reqBody)
		if nameZh != nil {
			err := reqBodyWriter.WriteField("nameZh", *nameZh)
			if err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if nameEn != nil {
			if err := reqBodyWriter.WriteField("nameEn", *nameEn); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if password != nil {
			if err := reqBodyWriter.WriteField("password", *password); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if passwordConfirmation != nil {
			if err := reqBodyWriter.WriteField("passwordConfirmation", *passwordConfirmation); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if department != nil {
			if err := reqBodyWriter.WriteField("department", *department); err != nil {
				t.Error("write form field error", err)
				return nil, ""
			}
		}
		if avatarName != nil {
			avatar, err := reqBodyWriter.CreateFormFile("avatar", *avatarName)
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

	t.Run("正常请求", func(t *testing.T) {
		ctx := context.Background()
		nameZh, nameEn, password, passwordConfirmation, department, avatarName, avatarBytes :=
			"张三", "zhangsan", "123456", "123456", "/技术部", "zhangsan.jpg", GenerateJPEG(t, 10, 10)

		dbUserMocker := MockDBClient[model.User](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker := MockRedis(ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.CountOnce(0, nil)
		tusdMocker = tusdMocker.UploadFileOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.SAddOnce(1, nil)
		dbUserMocker = dbUserMocker.CreateOnce(nil)
		dbFileMocker = dbFileMocker.CreateOnce(nil)
		defer dbUserMocker.Reset()
		defer dbFileMocker.Reset()
		defer redisMocker.Reset()
		defer tusdMocker.Reset()

		reqBody, contentType := createRequestBody(
			&nameZh,
			&nameEn,
			&password,
			&passwordConfirmation,
			&department,
			&avatarName,
			avatarBytes,
		)
		CheckAndUnmarshalBody[any](t, ServeHTTP(ctx,
			CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)), consts.AlertRegisterUser)
	})

	t.Run("正常请求_使用PNG格式头像", func(t *testing.T) {
		ctx := context.Background()
		nameZh, nameEn, password, passwordConfirmation, department, avatarName, avatarBytes := "张三", "zhangsan",
			"123456", "123456", "/技术部", "zhangsan.png", GeneratePNG(t, 10, 10)

		dbUserMocker := MockDBClient[model.User](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker := MockRedis(ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.CountOnce(0, nil)
		tusdMocker = tusdMocker.UploadFileOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.SAddOnce(1, nil)
		dbUserMocker = dbUserMocker.CreateOnce(nil)
		dbFileMocker = dbFileMocker.CreateOnce(nil)
		defer dbUserMocker.Reset()
		defer dbFileMocker.Reset()
		defer redisMocker.Reset()
		defer tusdMocker.Reset()

		reqBody, contentType := createRequestBody(
			&nameZh,
			&nameEn,
			&password,
			&passwordConfirmation,
			&department,
			&avatarName,
			avatarBytes,
		)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.AlertRegisterUser,
		)
	})

	t.Run("异常测试_使用GIF格式头像", func(t *testing.T) {
		ctx := context.Background()
		nameZh, nameEn, password, passwordConfirmation, department, avatarName, avatarBytes :=
			"张三", "zhangsan", "123456", "123456", "/技术部", "zhangsan.gif", GenerateGIF(t, 10, 10)

		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		defer redisMocker.Reset()

		reqBody, contentType := createRequestBody(
			&nameZh,
			&nameEn,
			&password,
			&passwordConfirmation,
			&department,
			&avatarName,
			avatarBytes,
		)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrUserAvatarFormatNotSupported,
		)
	})

	t.Run("异常测试_头像过大", func(t *testing.T) {
		ctx := context.Background()

		nameZh, nameEn, password, passwordConfirmation, department, avatarName, avatarBytes :=
			"张三", "zhangsan", "123456", "123456", "/技术部", "zhangsan.jpg", GenerateJPEG(t, 10, 10)

		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("Server").
			FieldByName("UserAvatarMaximumSize").
			Set(1).
			Reset()
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		defer redisMocker.Reset()

		reqBody, contentType := createRequestBody(
			&nameZh,
			&nameEn,
			&password,
			&passwordConfirmation,
			&department,
			&avatarName,
			avatarBytes,
		)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrUserAvatarTooLarge,
		)
	})

	t.Run("异常测试_名字重复", func(t *testing.T) {
		ctx := context.Background()
		nameZh, nameEn, password, passwordConfirmation, department, avatarName, avatarBytes :=
			"张三", "zhangsan", "123456", "123456", "/技术部", "zhangsan.jpg", GenerateJPEG(t, 10, 10)

		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.CountOnce(1, nil)
		defer dbUserMocker.Reset()
		defer redisMocker.Reset()

		reqBody, contentType := createRequestBody(
			&nameZh,
			&nameEn,
			&password,
			&passwordConfirmation,
			&department,
			&avatarName,
			avatarBytes,
		)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			consts.ErrUserNameDuplicate,
		)
	})

	validateErrorRequest := func(t *testing.T,
		nameZh, nameEn, password, passwordConfirmation, department, avatarName *string, fileData []byte) {
		ctx := context.Background()
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		defer redisMocker.Reset()
		reqBody, contentType := createRequestBody(
			nameZh,
			nameEn,
			password,
			passwordConfirmation,
			department,
			avatarName,
			fileData,
		)
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostMultiFormRequest(ctx, reqPath, reqBody, contentType)),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                                                   string
		NameZh, NameEn, Password, PasswordConfirmation, Department, AvatarName *string
		FileData                                                               []byte
	}{
		{"未上传头像", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), nil, nil},
		{"中文名过短", TakeStringPtr("张"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"中文名过长", TakeStringPtr("张一二三四五六七八九一二三四五六七"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"中文名包含非中文字符", TakeStringPtr("张3"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"中文名缺失", nil, TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"英文名缺失", TakeStringPtr("张三"), nil, TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"英文名过短", TakeStringPtr("张三"), TakeStringPtr("zhang"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"英文名过长", TakeStringPtr("张三"), TakeStringPtr("zhang012345678901234567890123456789"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"英文名包含非法字符", TakeStringPtr("张三"), TakeStringPtr("zhang三"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"英文名格式错误", TakeStringPtr("张三"), TakeStringPtr("1zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"密码缺失", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), nil, TakeStringPtr("123456"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"密码确认缺失", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), nil, TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"密码过短", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("12345"), TakeStringPtr("12345"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"密码字符非法", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("12345" + string([]byte{128})), TakeStringPtr("12345" + string([]byte{128})), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"密码不一致", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123455"), TakeStringPtr("/技术部"), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"部门缺失", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), nil, TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"部门缺失", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr(util.FastRandomAlphaNumberString(1025)), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
		{"部门字符错误", TakeStringPtr("张三"), TakeStringPtr("zhangsan"), TakeStringPtr("123456"), TakeStringPtr("123456"), TakeStringPtr("/技术部" + string([]byte{128})), TakeStringPtr("zhangsan.jpg"), GenerateJPEG(t, 10, 10)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.NameZh, v.NameEn, v.Password, v.PasswordConfirmation, v.Department, v.AvatarName, v.FileData)
		})
	}
}

func TestUserAPI_WebLogin(t *testing.T) {
	const reqPath = "/web/user/login"
	checkResponseSetCookie := func(rsp *httptest.ResponseRecorder) {
		containsSkey := false
		containsUser := false
		for _, v := range rsp.Header().Values("Set-Cookie") {
			if strings.Contains(v, consts.HTTPHeaderSessionKey) {
				containsSkey = true
			} else if strings.Contains(v, consts.HTTPHeaderSessionUser) {
				containsUser = true
			}
		}
		if !containsSkey || !containsUser {
			t.Errorf("expect true, but got %v %v", containsSkey, containsUser)
		}
	}

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)
		redisMocker = redisMocker.SetOnce("", nil)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		rsp := ServeHTTP(
			ctx,
			CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebLoginReq{
				NameEn:   LoginUser.NameEn,
				Password: UserPassword,
			}),
		)
		CheckAndUnmarshalBody[any](t, rsp, consts.AlertLogin)
		checkResponseSetCookie(rsp)
	})

	t.Run("异常测试_密码错误", func(t *testing.T) {
		ctx := context.Background()

		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebLoginReq{
				NameEn:   LoginUser.NameEn,
				Password: UserPassword + "~",
			})),
			consts.ErrUserLoginFailed,
		)
	})

	t.Run("异常测试_用户不存在", func(t *testing.T) {
		ctx := context.Background()
		nameEn := "zhangsan"
		password := util.FastRandomAlphaNumberString(6)

		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.TakeOnce(nil, gorm.ErrRecordNotFound)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebLoginReq{
				NameEn:   nameEn,
				Password: password,
			})),
			consts.ErrUserLoginFailed,
		)
	})

	t.Run("admin登录", func(t *testing.T) {
		ctx := context.Background()
		nameEn := "admin"
		password := "admin"

		digest := md5.Sum([]byte(password))
		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbUserMocker = dbUserMocker.TakeOnce(&model.User{
			ID:             1,
			NameEn:         nameEn,
			PasswordDigest: hex.EncodeToString(digest[:]),
		}, nil)
		redisMocker = redisMocker.SetOnce("", nil)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequest(
			ctx,
			reqPath, &protocol.UserWebLoginReq{
				NameEn:   nameEn,
				Password: password,
			}),
		)
		CheckAndUnmarshalBody[any](t, rsp, consts.AlertLogin)
		checkResponseSetCookie(rsp)
	})

	validateErrorRequest := func(t *testing.T, nameEn, password string) {
		ctx := context.Background()
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		defer redisMocker.Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebLoginReq{
				NameEn:   nameEn,
				Password: password,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name     string
		NameEn   string
		Password string
	}{
		{"密码过短", "zhangsan", util.FastRandomAlphaNumberString(5)},
		{"密码字符非法", "zhangsan", util.FastRandomAlphaNumberString(5) + string([]byte{128})},
		{"用户名过短", "zhang", util.FastRandomAlphaNumberString(6)},
		{"用户名过长", "zhang012345678901234567890123456789", util.FastRandomAlphaNumberString(6)},
		{"用户名格式错误", "1zhangsan", util.FastRandomAlphaNumberString(6)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) { validateErrorRequest(t, v.NameEn, v.Password) })
	}
}

func TestUserAPI_WebGetInformation(t *testing.T) {
	const reqPath = "/web/user/getInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		dbUserMocker := MockDBClient[model.User](ctx)
		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.GetOnce(Session, nil)
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.UserWebGetInformationRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.NameEn != LoginUser.NameEn {
			t.Errorf("expect %v, but got %v", LoginUser.NameEn, rspBodyObj.Data.NameEn)
		}
		if rspBodyObj.Data.NameZh != LoginUser.NameZh {
			t.Errorf("expect %v, but got %v", LoginUser.NameZh, rspBodyObj.Data.NameZh)
		}
		if rspBodyObj.Data.Department != LoginUser.Department {
			t.Errorf("expect %v, but got %v", LoginUser.Department, rspBodyObj.Data.Department)
		}
		if rspBodyObj.Data.AvatarFileID != LoginUser.AvatarFileID {
			t.Errorf("expect %v, but got %v", LoginUser.AvatarFileID, rspBodyObj.Data.AvatarFileID)
		}
	})

	t.Run("异常测试_未登录", func(t *testing.T) {
		ctx := context.Background()

		redisMocker := MockRedis(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.GetOnce("", redis.Nil)
		defer redisMocker.Reset()

		CheckAndUnmarshalBody[protocol.UserWebGetInformationRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			consts.ErrNeedLogin,
		)
	})
}

func TestUserAPI_WebUpdate(t *testing.T) {
	const reqPath = "/web/user/update"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		nameZh := "张三"
		newAvatarID := util.FastRandomAlphaNumberString(38)
		department := "/技术部"
		avatarBytes := GenerateJPEG(t, 10, 10)
		password := "123456"

		dbUserMocker := MockDBClient[model.User](ctx)
		dbFileMocker := MockDBClient[model.File](ctx)
		redisMocker := MockRedis(ctx)
		tusdMocker := MockTusdClient(ctx)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil)
		redisMocker = redisMocker.GetOnce(Session, nil)
		dbUserMocker = dbUserMocker.TakeOnce(LoginUser, nil)
		redisMocker = redisMocker.EvalshaOnce(true, nil)
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{
			ID:     1,
			FileID: newAvatarID,
			TusdID: util.FastRandomAlphaNumberString(32),
			Type:   model.FileTypeUserAvatar,
			UserID: 1,
		}, nil)
		tusdMocker = tusdMocker.GetOnce(&tus.GetResult{
			HTTPStatus:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(avatarBytes)),
			ContentLength: len(avatarBytes),
		}, nil)
		dbUserMocker = dbUserMocker.UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil)
		dbFileMocker = dbFileMocker.TakeOnce(&model.File{
			ID:     1,
			FileID: util.FastRandomAlphaNumberString(38),
			TusdID: util.FastRandomAlphaNumberString(32),
			Type:   model.FileTypeUserAvatar,
			UserID: 1,
		}, nil)
		dbFileMocker = dbFileMocker.DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil)
		tusdMocker = tusdMocker.DeleteFileOnce(nil)
		redisMocker = redisMocker.SRemOnce(1, nil)
		defer redisMocker.Reset()
		defer dbUserMocker.Reset()
		defer dbFileMocker.Reset()
		defer tusdMocker.Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebUpdateReq{
				AvatarFileID:         newAvatarID,
				NameZh:               nameZh,
				Password:             password,
				PasswordConfirmation: password,
				Department:           department,
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("异常测试_头像过大", func(t *testing.T) {
		ctx := context.Background()
		nameZh := "张三"
		avatarID := util.FastRandomAlphaNumberString(38)
		newAvatarID := util.FastRandomAlphaNumberString(38)
		department := "/技术部"
		avatarBytes := GenerateJPEG(t, 10, 10)
		password := "123456"

		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("Server").
			FieldByName("UserAvatarMaximumSize").
			Set(1).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: newAvatarID,
				TusdID: util.FastRandomAlphaNumberString(32),
				Type:   model.FileTypeUserAvatar,
				UserID: 1,
			}, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(&model.User{
				ID:           1,
				NameEn:       LoginUser.NameEn,
				NameZh:       nameZh,
				AvatarFileID: avatarID,
				Department:   department,
			}, nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(avatarBytes)),
				ContentLength: len(avatarBytes),
			}, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebUpdateReq{
				AvatarFileID:         newAvatarID,
				NameZh:               nameZh,
				Password:             password,
				PasswordConfirmation: password,
				Department:           department,
			})),
			consts.ErrUserAvatarTooLarge,
		)
	})

	t.Run("异常测试_头像格式错误", func(t *testing.T) {
		ctx := context.Background()
		nameZh := "张三"
		avatarID := util.FastRandomAlphaNumberString(38)
		newAvatarID := util.FastRandomAlphaNumberString(38)
		department := "/技术部"
		avatarBytes := GenerateGIF(t, 10, 10)
		password := "123456"

		defer MockDBClient[model.User](ctx).
			TakeOnce(&model.User{
				ID:           1,
				NameEn:       LoginUser.NameEn,
				NameZh:       nameZh,
				AvatarFileID: avatarID,
				Department:   department,
			}, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				ID:     1,
				FileID: newAvatarID,
				TusdID: util.FastRandomAlphaNumberString(32),
				Type:   model.FileTypeUserAvatar,
				UserID: 1,
			}, nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockTusdClient(ctx).
			GetOnce(&tus.GetResult{
				HTTPStatus:    http.StatusOK,
				Body:          io.NopCloser(bytes.NewReader(avatarBytes)),
				ContentLength: len(avatarBytes),
			}, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebUpdateReq{
				AvatarFileID:         newAvatarID,
				NameZh:               nameZh,
				Password:             password,
				PasswordConfirmation: password,
				Department:           department,
			})),
			consts.ErrUserAvatarFormatNotSupported,
		)
	})

	t.Run("异常测试_未登录", func(t *testing.T) {
		ctx := context.Background()
		nameZh := "张三"
		newAvatarID := util.FastRandomAlphaNumberString(38)
		department := "/技术部"
		password := "123456"

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce("", redis.Nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebUpdateReq{
				AvatarFileID:         newAvatarID,
				NameZh:               nameZh,
				Password:             password,
				PasswordConfirmation: password,
				Department:           department,
			})),
			consts.ErrNeedLogin,
		)
	})

	validateErrorRequest := func(
		t *testing.T,
		nameZh, avatarFileID, department, password, password2, newAvatarFileID string,
	) {
		ctx := context.Background()
		defer MockDBClient[model.User](ctx).
			TakeOnce(&model.User{
				ID:           1,
				NameEn:       LoginUser.NameEn,
				NameZh:       nameZh,
				AvatarFileID: avatarFileID,
				Department:   department,
			}, nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.UserWebUpdateReq{
				NameZh:               nameZh,
				Password:             password,
				PasswordConfirmation: password2,
				Department:           department,
				AvatarFileID:         newAvatarFileID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                                                   string
		NameZh, AvatarFileID, Department, Password, Password2, NewAvatarFileID string
	}{
		{"头像缺失", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", ""},
		{"头像ID格式错误", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", "汉" + util.FastRandomAlphaNumberString(37)},
		{"头像ID长度错误", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", util.FastRandomAlphaNumberString(32)},
		{"中文名缺失", "", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"中文名过长", "张三一二三四五六七八九十一二三四五", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"中文名包含非法字符", "张3", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"密码过短", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "12345", "12345", util.FastRandomAlphaNumberString(38)},
		{"密码不一致", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123455", util.FastRandomAlphaNumberString(38)},
		{"密码不一致", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456", "123455", util.FastRandomAlphaNumberString(38)},
		{"密码包含非法字符", "张三", util.FastRandomAlphaNumberString(38), "/技术部", "123456" + string([]byte{128}), "123456" + string([]byte{128}), util.FastRandomAlphaNumberString(38)},
		{"部门缺失", "张三", util.FastRandomAlphaNumberString(38), "", "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"部门过长", "张三", util.FastRandomAlphaNumberString(38), util.FastRandomAlphaNumberString(1025), "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"部门过长", "张三", util.FastRandomAlphaNumberString(38), util.FastRandomAlphaNumberString(1025), "123456", "123456", util.FastRandomAlphaNumberString(38)},
		{"部门包含非法字符", "张三", util.FastRandomAlphaNumberString(38), "/技术部/" + string([]byte{128}), "123456", "123456", util.FastRandomAlphaNumberString(38)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(
				t,
				v.NameZh,
				v.AvatarFileID,
				v.Department,
				v.Password,
				v.Password2,
				v.NewAvatarFileID,
			)
		})
	}
}

func TestUserAPI_WebSearch(t *testing.T) {
	const reqPath = "/web/user/search"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()
		name := "z"
		result := map[string]string{
			"1": "2",
			"3": "4",
		}

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			UserSearchByNameOnce(util.MapToList(result, func(k, v string) *model.User {
				return &model.User{
					NameEn: k,
					NameZh: v,
				}
			}), nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.UserWebSearchRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.UserWebSearchReq{Name: name})),
			0,
		)

		if len(result) != len(rspBodyObj.Data.Users) {
			t.Errorf("result length is unexpected, want %v, got %v", len(result), len(rspBodyObj.Data.Users))
		}
		for k, v := range result {
			if rspBodyObj.Data.Users[k] != v {
				t.Errorf("result body is unexpected, want %v, got %v", v, rspBodyObj.Data.Users[k])
			}
		}
	})

	t.Run("异常测试_未登录", func(t *testing.T) {
		ctx := context.Background()
		name := "z"

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce("", redis.Nil).
			Reset()

		CheckAndUnmarshalBody[protocol.UserWebSearchRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, &protocol.UserWebSearchReq{Name: name})),
			consts.ErrNeedLogin,
		)
	})

	validateErrorRequest := func(t *testing.T, nameEn string) {
		ctx := context.Background()

		defer MockDBClient[model.User](ctx).
			TakeOnce(&model.User{
				ID:           1,
				NameEn:       LoginUser.NameEn,
				NameZh:       "张三",
				AvatarFileID: util.FastRandomAlphaNumberString(38),
				Department:   "/技术部",
			}, nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.UserWebSearchRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.UserWebSearchReq{Name: nameEn})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name  string
		Name2 string
	}{
		{"名字缺失", ""},
		{"名字过长", util.FastRandomAlphaNumberString(33)},
		{"名字包含非法字符", "张" + string([]byte{128})},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Name2)
		})
	}
}

func TestUserAPI_WebLogout(t *testing.T) {
	const reqPath = "/web/user/logout"
	checkResponseSetCookie := func(rsp *httptest.ResponseRecorder) {
		containsSkey := false
		containsUser := false
		for _, v := range rsp.Header().Values("Set-Cookie") {
			if strings.Contains(v, consts.HTTPHeaderSessionKey+"=;") {
				containsSkey = true
			} else if strings.Contains(v, consts.HTTPHeaderSessionUser+"=;") {
				containsUser = true
			}
		}
		if !containsSkey || !containsUser {
			t.Errorf("expect true, but got %v %v", containsSkey, containsUser)
		}
	}

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			DelOnce(1, nil).
			Reset()

		rsp := ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, nil))
		CheckAndUnmarshalBody[any](t, rsp, consts.AlertLogout)
		checkResponseSetCookie(rsp)
	})

	t.Run("异常测试_未登录", func(t *testing.T) {
		ctx := context.Background()

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce("", redis.Nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, nil)),
			consts.ErrNeedLogin,
		)
	})
}
