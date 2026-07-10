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

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

const (
	EventUser                  EventField = "user"
	EventApp                   EventField = "app"
	EventAuth                  EventField = "auth"
	EventAppleBundleID         EventField = "appleBundleID"
	EventCertificateCommonName EventField = "certificateCommonName"
	EventAlias                 EventField = "alias"
	EventProvisionType         EventField = "provisionType"
	EventAppleCertificateType  EventField = "appleCertificateType"
	EventAppleDeviceModel      EventField = "appleDeviceModel"
	EventDetail                EventField = "detail"
)

type EventField string

// EventWebList 检索应用事件。
func EventWebList(ctx context.Context, req *protocol.EventWebListReq) (rsp *protocol.EventWebListRsp, err error) {
	// 获取上下文信息。
	var userInfo *model.User
	{
		log.Info(ctx, "get context information")
		userInfo = ctxs.User(ctx)
		if userInfo == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 根据开始结束时间，过滤出要检索的应用事件表。
	var eventTables []string
	{
		log.Info(ctx, "get app event tables")
		eventTables, err = filterEventTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return
		}
		if len(eventTables) <= 0 {
			return
		}
	}

	// 从数据库中查出应用事件关联应用 IDs。
	var appIDs []int
	{
		log.Info(ctx, "get user app ids from database")
		appQuery := conn.MySQLClient(ctx).App
		statement := appQuery.WithContext(ctx)
		var isSystemAdmin bool
		isSystemAdmin, err = IsSystemAdmin(ctx, userInfo.ID)
		if err != nil {
			return
		}
		if !isSystemAdmin {
			userRoleQuery := conn.MySQLClient(ctx).UserRole
			var appIDs2 []int
			err = userRoleQuery.WithContext(ctx).Select(
				userRoleQuery.AppID,
			).Where(
				userRoleQuery.UserID.Eq(userInfo.ID),
				userRoleQuery.Role.In(model.UserRoleAppAdmin, model.UserRoleAppMember),
			).Scan(&appIDs2)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 用户没有应用，则无应用事件信息。
					return
				}
				log.Error(ctx, "failed to retrieve user roles from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			statement = statement.Where(appQuery.ID.In(appIDs2...))
		}
		if len(req.App) > 0 {
			statement = statement.Where(
				appQuery.WithContext(ctx).Where(appQuery.AppID.Eq(req.App)).Or(appQuery.Name.Like("%" + req.App + "%")),
			)
		}
		if req.Platform > 0 {
			statement = statement.Where(appQuery.Platform.Eq(req.Platform))
		}
		// 是系统管理员，且不过滤应用，则不用查应用 IDs。
		if !(isSystemAdmin && len(req.App) <= 0) {
			if err = statement.Select(appQuery.ID).Scan(&appIDs); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 用户没有应用，则无应用事件信息。
					return
				}
				log.Error(ctx, "failed to retrieve app ids from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 从数据库中查出应用事件关联的操作人。
	var userIDs []int
	{
		log.Info(ctx, "get user ids from database")
		userQuery := conn.MySQLClient(ctx).User
		if len(req.User) > 0 {
			err = userQuery.WithContext(ctx).Select(
				userQuery.ID,
			).Where(
				userQuery.NameEn.Like("%" + req.User + "%"),
			).Or(
				userQuery.NameZh.Like("%" + req.User + "%"),
			).Scan(&userIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user ids from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 从数据库中检索应用事件。
	var eventInfos []*model.Event
	{
		log.Info(ctx, "get app events from database")
		rsp = &protocol.EventWebListRsp{}
		eventQuery := conn.MySQLClient(ctx).Event
		rsp.Count, err = eventQuery.WithContext(ctx).Count2(
			eventTables, appIDs, userIDs, req.BeginTime, req.EndTime, req.Type)
		if err != nil {
			log.Error(ctx, "failed to query app event information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		eventInfos, err = eventQuery.WithContext(ctx).List(eventTables, appIDs, userIDs, req.BeginTime, req.EndTime,
			req.Type, req.PageSize, (req.PageNumber-1)*req.PageSize)
		if err != nil {
			log.Error(ctx, "failed to query app events from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(eventInfos) <= 0 {
			return
		}
	}

	// 从数据库中查出应用信息。
	var appIDToInfo map[int]*model.App
	{
		log.Info(ctx, "get user and app information from database")
		appIDs, userIDs = util.ListTo2(eventInfos, func(e *model.Event) (int, int) { return e.AppID, e.UserID })
		appQuery := conn.MySQLClient(ctx).App
		var appInfos []*model.App
		appInfos, err = appQuery.WithContext(ctx).Select(
			appQuery.AppID,
			appQuery.ID,
			appQuery.Name,
			appQuery.Platform,
		).Where(
			appQuery.ID.In(util.CleanNumbers(appIDs)...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to get user and app information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		appIDToInfo = util.ListAssociateBy(appInfos, func(e *model.App) int { return e.ID })
	}

	// 从数据库中查出用户信息。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get user name from database")
		userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.EventWebListItem, len(eventInfos))
		for i, v := range eventInfos {
			appInfo, ok := appIDToInfo[v.AppID]
			if !ok {
				appInfo = &model.App{}
			}
			list[i] = &protocol.EventWebListItem{
				AppID:       appInfo.AppID,
				AppName:     appInfo.Name,
				Platform:    appInfo.Platform,
				Type:        v.Type,
				Source:      v.Source,
				User:        userIDToName[v.UserID],
				CreatedTime: formatTime(&v.CreatedTime),
				Content:     getEventDescription(ctx, v),
			}
		}
		rsp.List = list
	}

	return
}

// EventWebStatistic 获取应用事件统计数量。
func EventWebStatistic(ctx context.Context, req *protocol.EventWebStatisticReq) (
	rsp *protocol.EventWebStatisticRsp, err error) {

	// 获取上下文信息。
	userInfo := ctxs.User(ctx)
	{
		log.Info(ctx, "get context information")
		if userInfo == nil {
			log.Warn(ctx, "unknown context")
			return nil, errs.New(consts.ErrSystem)
		}
	}

	// 查询应用信息。
	var appID int
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app information")
			appQuery := conn.MySQLClient(ctx).App
			err = appQuery.WithContext(ctx).Select(
				appQuery.ID,
			).Where(
				appQuery.AppID.Eq(req.AppID),
			).Scan(&appID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Warn(ctx, "app not found")
					return nil, errs.New(consts.ErrParameterInvalid)
				}
				log.Error(ctx, "failed to retrieve app information from database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
		}
	}

	// 查询数据库，获取事件数量。
	var sqlResult []map[string]any
	{
		log.Info(ctx, "get app events")

		// 包含结束日期的记录。
		req.EndTime = req.EndTime.AddDate(0, 0, 1).Add(-time.Second)

		var tableNames []string
		tableNames, err = filterEventTables(ctx, req.BeginTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		if len(tableNames) <= 0 {
			return &protocol.EventWebStatisticRsp{}, nil
		}
		eventQuery := conn.MySQLClient(ctx).Event
		switch req.TimeStep {
		case protocol.TimeStepDay:
			sqlResult, err = eventQuery.WithContext(ctx).CountTypesWithDay(tableNames, []int{
				model.EventTypeRegisterApp,
				model.EventTypeUploadWindowsCertificate,
				model.EventTypeInvalidateApp,
				model.EventTypeApplyAndroidCertificate,
				model.EventTypeApplyProvision,
				model.EventTypeApplyPushCertificate,
				model.EventTypeUploadAndroidCertificate,
			}, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepWeek:
			sqlResult, err = eventQuery.WithContext(ctx).CountTypesWithWeek(tableNames, []int{
				model.EventTypeRegisterApp,
				model.EventTypeUploadWindowsCertificate,
				model.EventTypeInvalidateApp,
				model.EventTypeApplyAndroidCertificate,
				model.EventTypeApplyProvision,
				model.EventTypeApplyPushCertificate,
				model.EventTypeUploadAndroidCertificate,
			}, appID, req.BeginTime, req.EndTime)
		case protocol.TimeStepMonth:
			sqlResult, err = eventQuery.WithContext(ctx).CountTypesWithMonth(tableNames, []int{
				model.EventTypeRegisterApp,
				model.EventTypeUploadWindowsCertificate,
				model.EventTypeInvalidateApp,
				model.EventTypeApplyAndroidCertificate,
				model.EventTypeApplyProvision,
				model.EventTypeApplyPushCertificate,
				model.EventTypeUploadAndroidCertificate,
			}, appID, req.BeginTime, req.EndTime)
		default:
			log.Warn(ctx, "unknown time step", req.TimeStep)
			return nil, errs.New(consts.ErrParameterInvalid)
		}
		if err != nil {
			log.Error(ctx, "failed to count event information from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}

	// 处理数据。
	var items []*protocol.EventWebStatisticItem
	{
		log.Info(ctx, "deal sql data")
		items = make([]*protocol.EventWebStatisticItem, 0, len(sqlResult)/2)
		item := &protocol.EventWebStatisticItem{}
		for _, v := range sqlResult {
			if v == nil {
				continue
			}
			day := fmt.Sprintf("%s", v["day"])
			var typ int
			typ, err = strconv.Atoi(fmt.Sprintf("%v", v["type"]))
			if err != nil {
				log.Error(ctx, "failed to convert event type to int", err, v["type"])
			}
			var count int
			count, err = strconv.Atoi(fmt.Sprintf("%v", v["count"]))
			if err != nil {
				log.Error(ctx, "failed to convert event count to int", err, v["count"])
			}
			var t time.Time
			switch req.TimeStep {
			case protocol.TimeStepDay:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepWeek:
				t, err = time.Parse("20060102", day)
			case protocol.TimeStepMonth:
				t, err = time.Parse("200601", day)
			}
			if err != nil {
				log.Error(ctx, "failed to parse day", err, day)
				continue
			}
			if len(item.BeginTime) <= 0 {
				item.BeginTime = formatDate(&t)
				items = append(items, item)
			}
			beginTime := formatDate(&t)
			if beginTime != item.BeginTime {
				item = &protocol.EventWebStatisticItem{BeginTime: beginTime}
				items = append(items, item)
			}
			switch typ {
			case model.EventTypeRegisterApp:
				item.CreateAppTimes = count
			case model.EventTypeUploadWindowsCertificate:
				item.UploadWindowsCertificateTimes = count
			case model.EventTypeInvalidateApp:
				item.InvalidAppTimes = count
			case model.EventTypeApplyAndroidCertificate:
				item.ApplyAndroidCertificateTimes = count
			case model.EventTypeApplyProvision:
				item.ApplyAppleProfileTimes = count
			case model.EventTypeApplyPushCertificate:
				item.ApplyApplePushCertificateTimes = count
			case model.EventTypeUploadAndroidCertificate:
				item.UploadAndroidCertificateTimes = count
			default: // noop
			}
		}
	}

	rsp = &protocol.EventWebStatisticRsp{List: items}

	return
}

func createEvent(ctx context.Context, event *model.Event, content map[EventField]any) error {
	contentBytes, _ := json.Marshal(content)
	event.Content = string(contentBytes)
	eventTxDo := conn.MySQLTxClient(ctx).Event.Table(model.GetEventTableName(event.CreatedTime))
	err := eventTxDo.WithContext(ctx).Create(event)
	if err != nil {
		log.Error(ctx, "failed to save app event information to database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}

func getEventDescription(ctx context.Context, e *model.Event) string {
	m := make(map[EventField]any)
	err := json.Unmarshal([]byte(e.Content), &m)
	if err != nil {
		log.Error(ctx, "failed to unmarshal app event detail", err, e.Content)
	}
	lang := i18n.Language(ctxs.Language(ctx))
	switch e.Type {
	case model.EventTypeRegisterApp:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】注册了应用【%v】：%v", m[EventUser], m[EventApp], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] register a app [%v]: %v", m[EventUser], m[EventApp], m[EventDetail])
	case model.EventTypeInvalidateApp:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】无效化了应用【%v】：%v",
				m[EventUser], m[EventApp], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] invalidate the app [%v]: %v",
			m[EventUser], m[EventApp], m[EventDetail])
	case model.EventTypeEnableApp:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】申请有效化应用【%v】：%v",
				m[EventUser], m[EventApp], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] enable a app [%v]: %v",
			m[EventUser], m[EventApp], m[EventDetail])
	case model.EventTypeUpdateApp:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】更新了应用【%v】：%v",
				m[EventUser], m[EventApp], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] update a app [%v]: %v",
			m[EventUser], m[EventApp], m[EventDetail])
	case model.EventTypeApplyOpenAPI:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在应用【%v】下创建了请求凭证【%v】：%v",
				m[EventUser], m[EventApp], m[EventAuth], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] create a auth id [%v] in app [%v]: [%v]",
			m[EventUser], m[EventAuth], m[EventApp], m[EventDetail])
	case model.EventTypeUpdateOpenAPI:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】更新了应用【%v】的请求凭证【%v】：%v",
				m[EventUser], m[EventApp], m[EventAuth], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] update auth id [%v] in app [%v]: %v",
			m[EventUser], m[EventAuth], m[EventApp], m[EventDetail])
	case model.EventTypeRemoveOpenAPI:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了应用【%v】的请求凭证【%v】：%v",
				m[EventUser], m[EventApp], m[EventAuth], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] remove auth id [%v] in app [%v]: %v",
			m[EventUser], m[EventAuth], m[EventApp], m[EventDetail])
	case model.EventTypeRenewOpenAPI:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】续期了应用【%v】的请求凭证【%v】：%v",
				m[EventUser], m[EventApp], m[EventAuth], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] renew auth id [%v] in app [%v]: %v",
			m[EventUser], m[EventAuth], m[EventApp], m[EventDetail])
	case model.EventTypeResetOpenAPI:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】重置了应用【%v】的请求凭证【%v】：%v",
				m[EventUser], m[EventApp], m[EventAuth], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] reset auth id [%v] in app [%v]: %v",
			m[EventUser], m[EventAuth], m[EventApp], m[EventDetail])
	case model.EventTypeUploadWindowsCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在 Windows 应用【%v】上传了证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventCertificateCommonName], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] upload Windows certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventCertificateCommonName], m[EventApp], m[EventDetail])
	case model.EventTypeRemoveWindowsCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了 Windows 应用【%v】的证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventCertificateCommonName], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] delete the Windows certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventCertificateCommonName], m[EventApp], m[EventDetail])
	case model.EventTypeDownloadWindowsCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】下载了 Windows 应用【%v】的证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventCertificateCommonName], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] download the Windows certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventCertificateCommonName], m[EventApp], m[EventDetail])
	case model.EventTypeUploadAndroidCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在 Android 应用【%v】上传了证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] upload Android certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeApplyAndroidCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在 Android 应用【%v】申请了证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] apply Android certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeRemoveAndroidCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了 Android 应用【%v】的证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] delete the Android certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeDownloadAndroidCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】下载了 Android 应用【%v】的证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] download the Android certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeDownloadGooglePlayCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】下载了 Android 应用【%v】的 GooglePlay 证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] download the Android GooglePlay certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeGetFacebookCertificateDigest:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】获取了 Android 应用【%v】证书【%v】的 Facebook 散列：%v",
				m[EventUser], m[EventApp], m[EventAlias], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] obtain the Facebook hash of certificate [%v] of app [%v]: %v",
			m[EventUser], m[EventAlias], m[EventApp], m[EventDetail])
	case model.EventTypeApplyAppleBundleID:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】申请了 Apple 应用【%v】的 BundleID【%v】：%v",
				m[EventUser], m[EventApp], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] apply a BundleID [%v] of Apple app [%v]: %v",
			m[EventUser], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeModifyAppleBundleID:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】修改了 Apple 应用【%v】的 BundleID【%v】：%v",
				m[EventUser], m[EventApp], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] modify the BundleID [%v] of Apple app [%v]: %v",
			m[EventUser], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeRemoveAppleBundleID:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了 Apple 应用【%v】的 BundleID【%v】：%v",
				m[EventUser], m[EventApp], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] delete the BundleID [%v] of Apple app [%v]: %v",
			m[EventUser], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeApplyProvision:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在 Apple 应用【%v】下申请了类型为【%v】和 BundleID 为【%v】的描述文件：%v",
				m[EventUser], m[EventApp], m[EventProvisionType], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf(
			"User [%v] apply for a provision file of type [%v] and BundleID [%v] of the Apple app [%v]: %v",
			m[EventUser], m[EventProvisionType], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeApplyPushCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在 Apple 应用【%v】下申请了 BundleID 为【%v】的苹果推送证书：%v",
				m[EventUser], m[EventApp], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] apply a Apple push certificate of BundleID [%v] of the Apple app [%v]: %v",
			m[EventUser], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeDownloadAppleCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】下载了应用【%v】下类型为【%v】的证书【%v】：%v",
				m[EventUser], m[EventApp], m[EventAppleCertificateType], m[EventCertificateCommonName], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] download a certificate [%v] of type [%v] under app [%v]: %v",
			m[EventUser], m[EventCertificateCommonName], m[EventAppleCertificateType], m[EventApp], m[EventDetail])
	case model.EventTypeDownloadProvision:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】下载了应用【%v】下类型为【%v】和 BundleID 为【%v】的描述文件：%v",
				m[EventUser], m[EventApp], m[EventProvisionType], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] download a provision file of type [%v] and BundleID [%v] of the app [%v]: %v",
			m[EventUser], m[EventProvisionType], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeRemoveProvision:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了应用【%v】下类型为【%v】和 BundleID 为【%v】的描述文件：%v",
				m[EventUser], m[EventApp], m[EventProvisionType], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] delete a provision file of type [%v] and BundleID [%v] of the app [%v]: %v",
			m[EventUser], m[EventProvisionType], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeRemovePushCertificate:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】删除了应用【%v】下 BundleID 为【%v】的推送证书：%v",
				m[EventUser], m[EventApp], m[EventAppleBundleID], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] delete the Apple push certificate of BundleID [%v] of app [%v]: %v",
			m[EventUser], m[EventAppleBundleID], m[EventApp], m[EventDetail])
	case model.EventTypeRegisterAppleDevice:
		if lang == i18n.LanguageChinese {
			return fmt.Sprintf("用户【%v】在应用【%v】下注册了测试设备【%v】：%v",
				m[EventUser], m[EventApp], m[EventAppleDeviceModel], m[EventDetail])
		}

		return fmt.Sprintf("User [%v] register a test apple device [%v] under the app [%v]: %v",
			m[EventUser], m[EventAppleDeviceModel], m[EventApp], m[EventDetail])
	}
	return ""
}

func filterEventTables(ctx context.Context, begin, end time.Time) ([]string, error) {
	// 从数据库中查处所有应用事件表。
	eventDo := conn.MySQLClient(ctx).Event
	allEventTables, err := eventDo.WithContext(ctx).GetTables(cfg.Get().MySQL().Database())
	if err != nil {
		log.Error(ctx, "failed to retrieve database tables", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	slices.Sort(allEventTables)
	if begin.IsZero() && end.IsZero() {
		return allEventTables, nil
	}

	// 过滤应用事件表。
	endTable := model.GetEventTableName(end)
	beginTable := model.GetEventTableName(begin)
	if !begin.IsZero() {
		for i := range allEventTables {
			if allEventTables[i] >= beginTable {
				allEventTables = allEventTables[i:]
				break
			}
		}
	}
	if !end.IsZero() {
		for i := len(allEventTables) - 1; i >= 0; i-- {
			if allEventTables[i] <= endTable {
				allEventTables = allEventTables[:i+1]
				break
			}
		}
	}

	log.Info(ctx, "gather app event tables", allEventTables, begin, end)
	return allEventTables, nil
}
