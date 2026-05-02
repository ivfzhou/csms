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
	"time"

	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

func TestEventAPI_WebList(t *testing.T) {
	const reqPath = "/web/event/list"

	t.Run("正常测试", func(t *testing.T) {
		ctx := context.Background()

		app, platform, source, typ, user, beginTime, endTime, pageSize, pageNumber :=
			util.FastRandomAlphaNumberString(32), model.AppPlatformAndroid, model.SourceWeb,
			model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1

		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.User{{}, {}, {}}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			EventGetTablesOnce([]string{"t_event"}, nil).
			EventCount2Once(1, nil).
			EventListOnce([]*model.Event{{Content: "{}"}, {Content: "{}"}, {Content: "{}"}}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(0, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.App{{}, {}, {}}, nil).
			Reset()

		CheckAndUnmarshalBody[protocol.EventWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebListReq{
				App:        app,
				Platform:   platform,
				Source:     source,
				Type:       typ,
				User:       user,
				BeginTime:  beginTime,
				EndTime:    endTime,
				PageSize:   pageSize,
				PageNumber: pageNumber,
			})),
			0,
		)
	})

	validateErrorRequest := func(t *testing.T, app string, platform, source, typ int, user string,
		beginTime, endTime time.Time, pageSize, pageNumber int) {
		ctx := context.Background()
		defer MockRedis(ctx).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			ScriptLoadOnce(util.FastRandomAlphaNumberString(32), nil).
			GetOnce(Session, nil).
			Reset()
		defer MockDBClient[model.User](ctx).
			TakeOnce(LoginUser, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.User{{}, {}, {}}, nil).
			Reset()
		defer MockDBClient[model.Event](ctx).
			EventGetTablesOnce([]string{"t_event"}, nil).
			EventCount2Once(1, nil).
			EventListOnce([]*model.Event{{Content: "{}"}, {Content: "{}"}, {Content: "{}"}}, nil).
			Reset()
		defer MockDBClient[model.UserRole](ctx).
			CountOnce(0, nil).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			Reset()
		defer MockDBClient[model.App](ctx).
			ScanOnce(func(v any) { *v.(*[]int) = []int{1} }, nil).
			FindOnce([]*model.App{{}, {}, {}}, nil).
			Reset()
		CheckAndUnmarshalBody[protocol.EventWebListRsp](
			t,
			ServeHTTP(ctx, CreateGetRequest(ctx, reqPath, protocol.EventWebListReq{
				App:        app,
				Platform:   platform,
				Source:     source,
				Type:       typ,
				User:       user,
				BeginTime:  beginTime,
				EndTime:    endTime,
				PageSize:   pageSize,
				PageNumber: pageNumber,
			})),
			errs.ErrInvalidRequestParameters,
		)
	}

	for _, v := range []struct {
		Name                  string
		app                   string
		platform, source, typ int
		user                  string
		beginTime, endTime    time.Time
		pageSize, pageNumber  int
	}{
		{"应用名过长", util.FastRandomAlphaNumberString(65), model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"应用平台非法", "应用名", -1, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件来源非法", "应用名", model.AppPlatformAndroid, -1, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件类型非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, -1, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"操作人名过长", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, util.FastRandomAlphaNumberString(33), time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"操作人名字符非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "a" + string([]byte{128}), time.Now(), time.Now().Add(time.Hour), 10, 1},
		{"事件开始时间缺失", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Time{}, time.Now().Add(time.Hour), 10, 1},
		{"事件结束时间缺失", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Time{}, 10, 1},
		{"事件结束时间早于开始时间", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(-time.Hour), 10, 1},
		{"每页条数非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 0, 1},
		{"页数非法", "应用名", model.AppPlatformAndroid, model.SourceWeb, model.EventTypeApplyOpenAPI, "zhangsan", time.Now(), time.Now().Add(time.Hour), 10, 0},
	} {
		t.Run("异常测试_"+v.Name, func(t *testing.T) {
			validateErrorRequest(t, v.app, v.platform, v.source, v.typ, v.user, v.beginTime, v.endTime, v.pageSize,
				v.pageNumber)
		})
	}
}
