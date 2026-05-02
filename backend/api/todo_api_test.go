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
	"io"
	"net/http"
	"strings"
	"testing"

	"gorm.io/gen"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func TestTodoAPI_WebCount(t *testing.T) {
	const reqPath = "/web/todo/count"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CountOnce(1, nil).
			CountOnce(2, nil).
			Reset()

		rspBodyObj := CheckAndUnmarshalBody[protocol.TodoWebCountRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, nil)),
			0,
		)

		if rspBodyObj.Data.Done != 2 {
			t.Errorf("want 2, but got %v", rspBodyObj.Data.Done)
		}
		if rspBodyObj.Data.NeedToDeal != 1 {
			t.Errorf("want 1, but got %v", rspBodyObj.Data.Done)
		}
	})
}

func TestTodoAPI_WebListDealt(t *testing.T) {
	const reqPath = "/web/todo/listDealt"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		appID, types, status, pageNumber, pageSize :=
			AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 1, 10

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}}, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CountOnce(1, nil).
			FindOnce([]*model.Todo{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			ScanOnce(func(v any) { *v.(*int) = AppInfo.ID }, nil).
			FindOnce([]*model.App{{}}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(1, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListDealtRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListDealtReq{
				AppID:      appID,
				Types:      types,
				Status:     status,
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, appID string, types, status []int, pageNumber, pageSize int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.TodoWebListDealtRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListDealtReq{
				AppID:      appID,
				Types:      types,
				Status:     status,
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		AppID                string
		Types, Status        []int
		PageNumber, PageSize int
	}{
		{"应用ID错误", "汉", []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 1, 10},
		{"待办类型错误", AppInfo.AppID, []int{-1}, []int{model.TodoStatusApproved}, 1, 10},
		{"待办状态错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{1}, 1, 10},
		{"每页条数错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 1, 0},
		{"页数错误", AppInfo.AppID, []int{model.TodoTypeActivateApp}, []int{model.TodoStatusApproved}, 0, 10},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.AppID, v.Types, v.Status, v.PageNumber, v.PageSize)
		})
	}
}

func TestTodoAPI_WebList(t *testing.T) {
	const reqPath = "/web/todo/list"

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
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CountOnce(1, nil).
			FindOnce([]*model.Todo{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			FindOnce([]*model.App{{}}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, pageNumber, pageSize int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.TodoWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebListReq{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                 string
		PageNumber, PageSize int
	}{
		{"每页条数错误", 1, 0},
		{"页数错误", 0, 10},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.PageNumber, v.PageSize)
		})
	}
}

func TestTodoAPI_WebDetail(t *testing.T) {
	const reqPath = "/web/todo/getDetail"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		id := 1

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			FindOnce([]*model.User{{}, {}}, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{ApproverID: LoginUser.ID}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			TakeOnce(&model.App{}, nil).
			Reset()
		defer MockDBClient[model.AppleDevice](ctx).
			TakeOnce(&model.AppleDevice{}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebGetDetailReq{
				ID: id,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.TodoWebGetDetailReq{
				ID: id,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID错误", 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}

func TestTodoAPI_WebDeal(t *testing.T) {
	const reqPath = "/web/todo/deal"

	t.Run("正常测试_注册应用", func(t *testing.T) {
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
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterApp}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_加入应用", func(t *testing.T) {
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
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeJoinApp}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
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
			CreateOnce(nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeApplySigner}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_启用应用", func(t *testing.T) {
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
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeActivateApp}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_注册设备", func(t *testing.T) {
		ctx := context.Background()

		key, _, err := GenerateECDSAKeyPEM("P256")
		if err != nil {
			t.Fatal(err)
		}
		defer mvt.Chain(cfg.Get()).
			Elem().
			FieldByName("AppleAPI").
			FieldByName("Secret").
			Set(key).
			Reset()
		defer MockHTTPClient(ctx).
			ResponseOnce(&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil).
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
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.AppleDevice](ctx).
			TakeOnce(&model.AppleDevice{}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			TakeOnce(&model.Todo{Status: model.TodoStatusProcessing, Candidates: []int{LoginUser.ID}, Type: model.TodoTypeRegisterAppleDevice}, nil).
			UpdateColumnSimpleOnce(gen.ResultInfo{RowsAffected: 1}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             1,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, id int) {
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
		CheckAndUnmarshalBody[protocol.TodoWebGetDetailRsp](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebDealReq{
				ID:             id,
				IsPass:         true,
				ApproveMessage: "~",
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name string
		ID   int
	}{
		{"ID错误", 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.ID)
		})
	}
}

func TestTodoAPI_WebCreate(t *testing.T) {
	const reqPath = "/web/todo/create"

	t.Run("正常测试_加入应用", func(t *testing.T) {
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
			FindOnce([]*model.UserRole{{}, {}}, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CountOnce(0, nil).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        model.TodoTypeJoinApp,
				AppID:       AppInfo.AppID,
				ApplyReason: "~",
			})),
			consts.AlertSuccess,
		)
	})

	t.Run("正常测试_申请签名权限", func(t *testing.T) {
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
			CountOnce(1, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()
		defer MockDBClient[model.Todo](ctx).
			CountOnce(0, nil).
			CreateOnce(nil).
			Reset()

		CheckAndUnmarshalBody[any](
			t,
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        model.TodoTypeApplySigner,
				AppID:       AppInfo.AppID,
				ApplyReason: "~",
			})),
			consts.AlertSuccess,
		)
	})

	validateErrorRequest := func(t *testing.T, typ int, appID, applyReason string) {
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
			ServeHTTP(ctx, CreatePostJSONRequest(ctx, reqPath, &protocol.TodoWebCreateReq{
				Type:        typ,
				AppID:       appID,
				ApplyReason: applyReason,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name               string
		Type               int
		AppID, ApplyReason string
	}{
		{"类型错误", 0, AppInfo.AppID, "~"},
		{"应用ID缺失", model.TodoTypeJoinApp, "", "~"},
		{"应用ID非法", model.TodoTypeJoinApp, "汉" + util.FastRandomAlphaNumberString(31), "~"},
		{"应用ID错误", model.TodoTypeJoinApp, util.FastRandomAlphaNumberString(33), "~"},
		{"理由缺失", model.TodoTypeJoinApp, AppInfo.AppID, ""},
		{"理由过长", model.TodoTypeJoinApp, AppInfo.AppID, util.FastRandomAlphaNumberString(257)},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.Type, v.AppID, v.ApplyReason)
		})
	}
}
