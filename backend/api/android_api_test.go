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
	"encoding/base64"
	"io"
	"net/http"
	"testing"
	"time"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestAndroidAPI_WebApplyCertificate(t *testing.T) {
	const reqPath = "/web/android/applyCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			TakeOnce(&model.AndroidOrganization{Owner: "CN=ivfzhou"}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AesKey](ctx).
			LastOnce(&model.AesKey{
				ID:          1,
				Secret:      util.RandomBytes(16),
				CreatedTime: time.Now(),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebApplyCertificateReq{
				Type:    model.AndroidCertificateTypeDebug,
				OwnerID: 1,
				Alias:   "abc",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ, ownerID int, alias string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
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
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebApplyCertificateReq{
				Type:    typ,
				OwnerID: ownerID,
				Alias:   alias,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name    string
		Type    int
		OwnerID int
		Alias   string
	}{
		{"类型缺失", 0, 1, "abc"},
		{"类型错误", 3, 1, "abc"},
		{"主体缺失", 1, 0, "abc"},
		{"别名缺失", 1, 1, ""},
		{"别名字符非法", 1, 1, "～"},
		{"别名字符过多", 1, 1, util.FastRandomAlphaNumberString(65)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.OwnerID, v.Alias)
		})
	}
}

func TestAndroidAPI_WebAddOrganization(t *testing.T) {
	const reqPath = "/web/android/addOrganization"

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
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AndroidWebAddOrganizationReq{
				CommonName: "abc",
				DName:      "abc",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, commonName, dName string) {
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
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.AndroidWebAddOrganizationReq{
				CommonName: commonName,
				DName:      dName,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name       string
		CommonName string
		DName      string
	}{
		{"通用名缺失", "", "abc"},
		{"通用名过长", util.FastRandomAlphaNumberString(33), "abc"},
		{"DName缺失", "abc", ""},
		{"DName过长", "abc", util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CommonName, v.DName)
		})
	}
}

func TestAndroidAPI_WebListOrganizations(t *testing.T) {
	const reqPath = "/web/android/listOrganizations"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			FindOnce([]*model.AndroidOrganization{{}, {}, {}}, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListOrganizationsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("want list, got empty list")
		}
	})
}

func TestAndroidAPI_WebRemoveOrganization(t *testing.T) {
	const reqPath = "/web/android/removeOrganization"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		id := 1

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			EvalshaOnce(true, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.AndroidOrganization](ctx).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: id})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			EvalshaOnce(true, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequest(ctx, reqPath, protocol.AndroidWebRemoveOrganizationReq{ID: id})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID缺失", 0},
		{"ID非法", -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}

func TestAndroidAPI_WebUploadCertificate(t *testing.T) {
	const reqPath = "/web/android/uploadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AesKey](ctx).
			LastOnce(&model.AesKey{
				ID:          1,
				Secret:      util.RandomBytes(16),
				CreatedTime: time.Now(),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebUploadCertificateReq{
				Type:        model.AndroidCertificateTypeDebug,
				Storepass:   androidKeystoreStorepass,
				Keypass:     androidKeystoreKeypass,
				Certificate: androidCertificate,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ int, storepass, keypass, certificate string) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.AesKey](ctx).
			LastOnce(&model.AesKey{
				ID:          1,
				Secret:      util.RandomBytes(16),
				CreatedTime: time.Now(),
			}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebUploadCertificateReq{
				Type:        typ,
				Storepass:   storepass,
				Keypass:     keypass,
				Certificate: certificate,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		Type        int
		Storepass   string
		Keypass     string
		Certificate string
	}{
		{"类型缺失", 0, "123456", "123456", androidCertificate},
		{"类型错误", -1, "123456", "123456", androidCertificate},
		{"Storepass错误", model.AndroidCertificateTypeDebug, "", "123456", androidCertificate},
		{"Storepass过长", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(33), "123456", androidCertificate},
		{"Storepass过短", model.AndroidCertificateTypeDebug, util.FastRandomAlphaNumberString(5), "123456", androidCertificate},
		{"Keypass错误", model.AndroidCertificateTypeDebug, "123456", "", androidCertificate},
		{"Keypass过长", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(33), androidCertificate},
		{"Keypass过短", model.AndroidCertificateTypeDebug, "123456", util.FastRandomAlphaNumberString(5), androidCertificate},
		{"证书缺失", model.AndroidCertificateTypeDebug, "123456", "123456", ""},
		{"证书错误", model.AndroidCertificateTypeDebug, "123456", "123456", "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.Storepass, v.Keypass, v.Certificate)
		})
	}
}

func TestAndroidAPI_WebListCertificates(t *testing.T) {
	const reqPath = "/web/android/listCertificates"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			FindOnce([]*model.AndroidCertificate{{}, {}}, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListCertificatesRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil)),
			0,
		)
		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
	})
}

func TestAndroidAPI_WebDownloadCertificate(t *testing.T) {
	const reqPath = "/web/android/downloadCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			ScanOnce(func(v any) { *v.(*[]byte) = secret }, nil).
			Reset()
		encrypt, _ := util.AESCBCEncrypt(secret, []byte("abc"))
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, nil))

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
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebDownloadCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPI_WebDeleteCertificate(t *testing.T) {
	const reqPath = "/web/android/deleteCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPI_WebGetGooglePlayDeployCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayDeployCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil).
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil).
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar").
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set(`java`).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebGetGooglePlayDeployCertificateReq{
			CertificateID: util.FastRandomAlphaNumberString(32),
			PublicKey:     androidRSAPublicKey,
		}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebGetGooglePlayDeployCertificateReq{
				CertificateID: certificateID,
				PublicKey:     androidRSAPublicKey,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPI_WebGetGooglePlayCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil).
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rsp := ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebGetGooglePlayCertificateReq{CertificateID: certificateID}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebGetGooglePlayCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPI_WebGetGooglePlayUpgradeCertificate(t *testing.T) {
	const reqPath = "/web/android/getGooglePlayUpgradeCertificate"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil).
			TakeOnce(&model.AesKey{Secret: secret}, nil).
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil).
			Reset()
		defer mvt.Chain(&consts.PepkJarPath).
			Elem().
			Set("../../pepk.jar").
			Reset()
		defer mvt.Chain(&consts.JavaBinaryPathForPepk).
			Elem().
			Set(`java`).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rsp := ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx,
			reqPath, AppInfo.AppID, &protocol.AndroidWebGetGooglePlayUpgradeCertificateReq{
				DeployCertificateID: util.FastRandomAlphaNumberString(32),
				UploadCertificateID: util.FastRandomAlphaNumberString(32),
				PublicKey:           androidRSAPublicKey,
			}))

		if rsp.Code != http.StatusOK {
			t.Errorf("response code is not 200")
		}
		bs2, _ := io.ReadAll(rsp.Body)
		if len(bs2) <= 0 {
			t.Errorf("response body is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, deployCertificateID, uploadCertificateID, publicKey string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx,
				reqPath, AppInfo.AppID, &protocol.AndroidWebGetGooglePlayUpgradeCertificateReq{
					DeployCertificateID: deployCertificateID,
					UploadCertificateID: uploadCertificateID,
					PublicKey:           publicKey,
				})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                                                string
		DeployCertificateID, UploadCertificateID, PublicKey string
	}{
		{"部署证书ID缺失", "", util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"部署证书ID错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"部署证书ID非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(32), androidRSAPublicKey},
		{"升级证书ID缺失", util.FastRandomAlphaNumberString(32), "", androidRSAPublicKey},
		{"升级证书ID错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31), androidRSAPublicKey},
		{"升级证书ID非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(31) + "汉", androidRSAPublicKey},
		{"公钥缺失", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(32), ""},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.DeployCertificateID, v.UploadCertificateID, v.PublicKey)
		})
	}
}

func TestAndroidAPI_WebGetCertificateFacebookDigest(t *testing.T) {
	const reqPath = "/web/android/getCertificateFacebookDigest"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		certificateID := util.FastRandomAlphaNumberString(32)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		secret := util.RandomBytes(16)
		defer MockDBClient[model.AesKey](ctx).
			TakeOnce(&model.AesKey{Secret: secret}, nil).
			Reset()
		bs, _ := base64.StdEncoding.DecodeString(androidCertificate)
		encrypt, _ := util.AESCBCEncrypt(secret, bs)
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{Content: encrypt, Storepass: androidKeystoreStorepass, Keypass: androidKeystoreKeypass, Alias_: androidKeystoreAlias}, nil).
			Reset()
		defer mvt.Chain(&consts.KeytoolBinaryPath).
			Elem().
			Set("keytool").
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebGetCertificateFacebookDigestRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebGetCertificateFacebookDigestReq{CertificateID: certificateID})),
			0,
		)

		if len(rspBodyObj.Data.Digest) <= 0 {
			t.Errorf("digest is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, certificateID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebDeleteCertificateReq{CertificateID: certificateID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
	}{
		{"证书ID缺失", ""},
		{"证书ID错误", util.FastRandomAlphaNumberString(31)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID)
		})
	}
}

func TestAndroidAPI_WebSubmitAPKSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		signatureSchema := []int{1}
		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
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
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAPK,
				Type:   model.FileTypeAndroidSigning,
			}, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil).
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil).
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAPKSigningJobReq{
				SignatureSchema: signatureSchema,
				CertificateID:   certificateID,
				FileID:          fileID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAPKSigningJobReq{
				SignatureSchema: signatureSchema,
				CertificateID:   certificateID,
				FileID:          fileID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name            string
		SignatureSchema []int
		CertificateID   string
		FileID          string
	}{
		{"签名方案缺失", nil, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"签名方案重复", []int{1, 1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38)},
		{"证书ID缺失", []int{1}, "", util.FastRandomAlphaNumberString(38)},
		{"证书ID错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书ID非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件ID非法", []int{1}, util.FastRandomAlphaNumberString(32), ""},
		{"文件ID错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件ID非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidAPI_WebSubmitAABSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAABSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
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
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAAB,
				Type:   model.FileTypeAndroidSigning,
			}, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil).
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil).
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAABSigningJobReq{
				CertificateID: certificateID,
				FileID:        fileID,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, certificateID, fileID string) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAABSigningJobReq{
				CertificateID: certificateID,
				FileID:        fileID,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name          string
		CertificateID string
		FileID        string
	}{
		{"证书ID缺失", "", util.FastRandomAlphaNumberString(38)},
		{"证书ID错误", util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38)},
		{"证书ID非法", util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38)},
		{"文件ID非法", util.FastRandomAlphaNumberString(32), ""},
		{"文件ID错误", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37)},
		{"文件ID非法", util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.CertificateID, v.FileID)
		})
	}
}

func TestAndroidAPI_WebSubmitAPKPatchSigningJob(t *testing.T) {
	const reqPath = "/web/android/submitAPKPatchSigningJob"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		certificateID := util.FastRandomAlphaNumberString(32)
		fileID := util.FastRandomAlphaNumberString(38)
		minimumSDKVersion := 1
		signatureSchema := []int{1}

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			SAddOnce(1, nil).
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
		defer MockDBClient[model.File](ctx).
			TakeOnce(&model.File{
				UserID: LoginUser.ID,
				AppID:  AppInfo.ID,
				Name:   cc.ExtensionAPK,
				Type:   model.FileTypeAndroidSigning,
			}, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			TakeOnce(&model.AndroidCertificate{NotAfter: time.Now().Add(time.Hour)}, nil).
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			CreateOnce(nil).
			Reset()
		defer MockRabbitMQClient(ctx).
			PublishWithContextOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAPKPatchSigningJobReq{
				SignatureSchema:   signatureSchema,
				CertificateID:     certificateID,
				FileID:            fileID,
				MinimumSDKVersion: minimumSDKVersion,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, signatureSchema []int, certificateID, fileID string, minimumSDKVersion int) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.AndroidWebSubmitAPKPatchSigningJobReq{
				SignatureSchema:   signatureSchema,
				CertificateID:     certificateID,
				FileID:            fileID,
				MinimumSDKVersion: minimumSDKVersion,
			})),
			errs.ErrInvalidRequestParameters,
		)
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
		{"证书ID缺失", []int{1}, "", util.FastRandomAlphaNumberString(38), 1},
		{"证书ID错误", []int{1}, util.FastRandomAlphaNumberString(31), util.FastRandomAlphaNumberString(38), 1},
		{"证书ID非法", []int{1}, util.FastRandomAlphaNumberString(31) + "汉", util.FastRandomAlphaNumberString(38), 1},
		{"文件ID非法", []int{1}, util.FastRandomAlphaNumberString(32), "", 1},
		{"文件ID错误", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37), 1},
		{"文件ID非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(37) + "汉", 1},
		{"SDK版本缺失", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), 0},
		{"SDK版本非法", []int{1}, util.FastRandomAlphaNumberString(32), util.FastRandomAlphaNumberString(38), -1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.SignatureSchema, v.CertificateID, v.FileID, v.MinimumSDKVersion)
		})
	}
}

func TestAndroidAPI_WebListSigningJobs(t *testing.T) {
	const reqPath = "/web/android/listSigningJobs"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		keyWord := "~"
		status := 1
		certificateAlias := "~"
		pageNumber := 1
		pageSize := 1

		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
			Reset()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.AndroidCertificate](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.AndroidCertificate{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.APIAccount{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.AndroidSigningJob](ctx).
			AndroidSigningJobGetTablesOnce([]string{""}, nil).
			AndroidSigningJobCount2Once(1, nil).
			AndroidSigningJobListOnce([]*model.AndroidSigningJob{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.File](ctx).
			FindOnce([]*model.File{{}, {}}, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.AndroidWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebListSigningJobsReq{
				KeyWord:          keyWord,
				Status:           status,
				CertificateAlias: certificateAlias,
				PageNumber:       pageNumber,
				PageSize:         pageSize,
			})),
			0,
		)

		if len(rspBodyObj.Data.List) <= 0 {
			t.Errorf("list is empty")
		}
		if rspBodyObj.Data.Count <= 0 {
			t.Errorf("count is zero")
		}
	})

	validateErrorRequest := func(t *testing.T, keyWord string, status int, certificateAlias string, pageNumber, pageSize int) {
		ctx := context.Background()
		defer mvt.Chain(AppInfo).
			Elem().
			FieldByName("Platform").
			Set(model.AppPlatformAndroid).
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
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.AndroidWebListSigningJobsRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.AndroidWebListSigningJobsReq{
				KeyWord:          keyWord,
				Status:           status,
				CertificateAlias: certificateAlias,
				PageNumber:       pageNumber,
				PageSize:         pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name             string
		Status           int
		CertificateAlias string
		PageNumber       int
		PageSize         int
	}{
		{"状态非法", -1, "", 1, 1},
		{"别名过长", 0, util.FastRandomAlphaNumberString(65), 1, 1},
		{"页码非法", 0, "", 0, 1},
		{"页条数非法", 0, "", 1, 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, "", v.Status, v.CertificateAlias, v.PageNumber, v.PageSize)
		})
	}
}
