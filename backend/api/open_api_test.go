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
	"testing"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func TestOpenAPI_WebApply(t *testing.T) {
	const reqPath = "/web/open/apply"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		authID, authorities, requestIP, frequency :=
			"a"+util.FastRandomAlphaNumberString(5), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(nil, nil).
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.APIAuthorization](ctx).
			CreateInBatchesOnce(nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebApplyReq{
				AccountID:   authID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			0,
		)

		if len(rspBodyObj.Data.Password) <= 0 {
			t.Errorf("password is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, accountID string, authorities []int, requestIP string, frequency int) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebApplyReq{
				AccountID:   accountID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		AccountID   string
		Authorities []int
		RequestIP   string
		Frequency   int
	}{
		{"凭证ID缺失", "", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证ID字符非法", "1a", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项重复", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword, model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项非法", "authId", []int{-1}, "127.0.0.1", 100},
		{"请求源缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "", 100},
		{"请求源过多", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, util.FastRandomAlphaNumberString(257), 100},
		{"请求频率缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 0},
		{"请求频率过大", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 99999},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID, v.Authorities, v.RequestIP, v.Frequency)
		})
	}
}

func TestOpenAPI_WebUpdate(t *testing.T) {
	const reqPath = "/web/open/update"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		authID, authorities, requestIP, frequency :=
			"a"+util.FastRandomAlphaNumberString(5), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(&model.APIAccount{}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.APIAuthorization](ctx).
			DeleteOnce(gen.ResultInfo{}, nil).
			CreateInBatchesOnce(nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebUpdateReq{
				AccountID:   authID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string, authorities []int, requestIP string, frequency int) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.OpenWebApplyRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequestWithApp(ctx, reqPath, AppInfo.AppID, &protocol.OpenWebUpdateReq{
				AccountID:   accountID,
				Authorities: authorities,
				RequestIP:   requestIP,
				Frequency:   frequency,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name        string
		AccountID   string
		Authorities []int
		RequestIP   string
		Frequency   int
	}{
		{"凭证ID缺失", "", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65), []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"凭证ID字符非法", "1a", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项重复", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword, model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 100},
		{"授权项非法", "authId", []int{-1}, "127.0.0.1", 100},
		{"请求源缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "", 100},
		{"请求源过多", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, util.FastRandomAlphaNumberString(257), 100},
		{"请求频率缺失", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 0},
		{"请求频率过大", "authId", []int{model.CapabilityGetWindowsOVCertificatePassword}, "127.0.0.1", 99999},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID, v.Authorities, v.RequestIP, v.Frequency)
		})
	}
}

func TestOpenAPI_WebGetInformation(t *testing.T) {
	const reqPath = "/web/open/getInformation"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		accountID := "a" + util.FastRandomAlphaNumberString(5)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			ScanOnce(func(v any) { *v.(*string) = "" }, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(&model.APIAccount{}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil).
			Reset()
		defer MockDBClient[model.APIAuthorization](ctx).
			FindOnce([]*model.APIAuthorization{{}}, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebGetInformationReq{AccountID: accountID})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebGetInformationReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证ID缺失", ""},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65)},
		{"凭证ID字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebList(t *testing.T) {
	const reqPath = "/web/open/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		pageNumber, pageSize := 1, 10

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{}, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			CountOnce(1, nil).
			FindOnce([]*model.APIAccount{{}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil).
			Reset()
		defer MockDBClient[model.APIAuthorization](ctx).
			ScanOnce(func(v any) {}, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, pageSize, pageNumber int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleSystemAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageSize, PageNumber int
	}{
		{"页数非法", 10, 0},
		{"每页条数非法", 0, 1},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageSize, v.PageNumber)
		})
	}
}

func TestOpenAPI_WebRenewal(t *testing.T) {
	const reqPath = "/web/open/renewal"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		accountID := "a" + util.FastRandomAlphaNumberString(5)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(&model.APIAccount{}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRenewalReq{AccountID: accountID})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRenewalReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证ID缺失", ""},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65)},
		{"凭证ID字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebRemove(t *testing.T) {
	const reqPath = "/web/open/remove"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		accountID := "a" + util.FastRandomAlphaNumberString(5)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(&model.APIAccount{}, nil).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.APIAuthorization](ctx).
			ScanOnce(func(v any) {}, nil).
			DeleteOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRemoveReq{AccountID: accountID})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateDeleteRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebRemoveReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证ID缺失", ""},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65)},
		{"凭证ID字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}

func TestOpenAPI_WebReset(t *testing.T) {
	const reqPath = "/web/open/reset"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		accountID := "a" + util.FastRandomAlphaNumberString(5)

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			EvalshaOnce(true, nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.APIAccount](ctx).
			TakeOnce(&model.APIAccount{}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(AppInfo, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			CreateOnce(nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.OpenWebResetRsp](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebResetReq{AccountID: accountID})),
			0,
		)

		if len(rspBodyObj.Data.Password) <= 0 {
			t.Errorf("password is empty")
		}
	})

	validateErrorRequest := func(t *testing.T, accountID string) {
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
		defer MockDBClient[model.UserRole](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{model.UserRoleAppAdmin} }, nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreateGetRequestWithApp(ctx, reqPath, AppInfo.AppID, protocol.OpenWebResetReq{AccountID: accountID})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name      string
		AccountID string
	}{
		{"凭证ID缺失", ""},
		{"凭证ID过长", util.FastRandomAlphaNumberString(65)},
		{"凭证ID字符非法", "1a"},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AccountID)
		})
	}
}
