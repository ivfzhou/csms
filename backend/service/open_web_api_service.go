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
	"errors"
	"strings"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

// OpenWebApply 申请请求凭证。
func OpenWebApply(ctx context.Context, req *protocol.OpenWebApplyReq) (rsp *protocol.OpenWebApplyRsp, err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app status")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 校验授权项。
	{
		log.Info(ctx, "validate authorities")
		switch app.Platform {
		case model.AppPlatformAndroid:
			if !util.ContainsAll(model.AllAndroidAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case model.AppPlatformApple:
			if !util.ContainsAll(model.AllAppleAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case model.AppPlatformWindows:
			if !util.ContainsAll(model.AllWindowsAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		}
	}

	// 校验请求源 IP。
	var newIPs []string
	{
		log.Info(ctx, "validate ip parameter")
		req.RequestIP = util.TrimBlank(req.RequestIP)
		ips := strings.Split(req.RequestIP, ",")
		if len(ips) <= 0 {
			log.Warn(ctx, "invalid ip parameter")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		newIPs = make([]string, 0, len(ips))
		for _, v := range ips {
			v = util.TrimBlank(v)
			if v == "*" {
				newIPs = append(newIPs, v)
				continue
			}

			pair := strings.Split(v, "-")
			if len(pair) == 2 {
				pair[0] = util.TrimBlank(pair[0])
				pair[1] = util.TrimBlank(pair[1])
				if util.IPv4ToNumber(pair[0]) <= 0 || util.IPv4ToNumber(pair[1]) <= 0 {
					log.Info(ctx, "invalid ip parameter")
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
				newIPs = append(newIPs, strings.Join(pair, "-"))
				continue
			}

			if util.IPv4ToNumber(v) <= 0 {
				log.Info(ctx, "invalid ip parameter")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			newIPs = append(newIPs, v)
		}
	}

	// 查询数据库，检查不存在同名的请求凭证账号。
	{
		log.Info(ctx, "check that api account id does not exist")
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		var count int64
		count, err = apiAccountTxDo.WithContext(ctx).Where(
			apiAccountTxDo.AppID.Eq(app.ID),
			apiAccountTxDo.AccountID.Eq(req.AccountID),
		).Clauses(query.ForUpdate()).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve api account from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			err = errs.New(consts.ErrAuthIDAlreadyExist)
			return
		}
	}

	// 将请求凭证账号信息保存到数据库中。
	var apiAccount *model.APIAccount
	var now time.Time
	{
		log.Info(ctx, "save api account")
		now = time.Now()
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		apiAccount = &model.APIAccount{
			AppID:       app.ID,
			UserID:      user.ID,
			AccountID:   req.AccountID,
			IP:          newIPs,
			Frequency:   req.Frequency,
			Secret:      util.RandomPrintableASCIINoSpaceString(consts.ApiAccountSecretLength),
			ExpiredTime: now.Add(consts.APIAccountExpirationTime),
			CreatedTime: now,
		}
		if err = apiAccountTxDo.WithContext(ctx).Select(
			apiAccountTxDo.AppID,
			apiAccountTxDo.UserID,
			apiAccountTxDo.AccountID,
			apiAccountTxDo.IP,
			apiAccountTxDo.Frequency,
			apiAccountTxDo.Secret,
			apiAccountTxDo.ExpiredTime,
			apiAccountTxDo.CreatedTime,
		).Create(apiAccount); err != nil {
			log.Error(ctx, "failed to save api account to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将请求凭证账号授权项保存到数据库中。
	{
		log.Info(ctx, "save api authorizations")
		apiAuthorizations := util.ListTo(req.Authorities, func(e int) *model.APIAuthorization {
			return &model.APIAuthorization{
				APIAccountID: apiAccount.ID,
				Capability:   e,
			}
		})
		if len(apiAuthorizations) > 0 {
			err = conn.MySQLTxClient(ctx).APIAuthorization.WithContext(ctx).
				CreateInBatches(apiAuthorizations, cfg.Get().MySQL().MaximumNumberOfPerSQLInsert())
			if err != nil {
				log.Error(ctx, "failed to save api authorizations to database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 将应用事件信息保存到数据库中。
	{
		log.Info(ctx, "add app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeApplyOpenAPI,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:  app.Name,
			EventUser: user.NameEn,
			EventAuth: req.AccountID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"frequency":   req.Frequency,
				"authorities": req.Authorities,
				"requestIp":   apiAccount.IP,
			}),
		})
		if err != nil {
			return
		}
	}

	rsp = &protocol.OpenWebApplyRsp{Password: apiAccount.Secret}

	return
}

// OpenWebUpdate 修改。
func OpenWebUpdate(ctx context.Context, req *protocol.OpenWebUpdateReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app status")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 校验授权项。
	{
		log.Info(ctx, "verify api account authorities")
		switch app.Platform {
		case model.AppPlatformAndroid:
			if !util.ContainsAll(model.AllAndroidAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case model.AppPlatformApple:
			if !util.ContainsAll(model.AllAppleAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		case model.AppPlatformWindows:
			if !util.ContainsAll(model.AllWindowsAppCapabilities, req.Authorities...) {
				log.Warn(ctx, "authorities is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		}
	}

	// 校验请求源 IP。
	var newIPs []string
	{
		log.Info(ctx, "validate ip parameter")
		req.RequestIP = util.TrimBlank(req.RequestIP)
		ips := strings.Split(req.RequestIP, ",")
		if len(ips) <= 0 {
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		newIPs = make([]string, 0, len(ips))
		for _, v := range ips {
			v = util.TrimBlank(v)
			if v == "*" {
				newIPs = append(newIPs, v)
				continue
			}
			pair := strings.Split(v, "-")
			if len(pair) == 2 {
				pair[0] = util.TrimBlank(pair[0])
				pair[1] = util.TrimBlank(pair[1])
				if util.IPv4ToNumber(pair[0]) <= 0 || util.IPv4ToNumber(pair[1]) <= 0 {
					err = errs.New(consts.ErrParameterInvalid)
					return
				}
				newIPs = append(newIPs, strings.Join(pair, "-"))
				continue
			}
			if util.IPv4ToNumber(v) <= 0 {
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			newIPs = append(newIPs, v)
		}
	}

	// 从数据库中获取请求凭证账号信息。
	var apiAccount *model.APIAccount
	{
		log.Info(ctx, "get api account information")
		apiAccountTxDo := conn.MySQLClient(ctx).APIAccount
		apiAccount, err = apiAccountTxDo.WithContext(ctx).Select(
			apiAccountTxDo.ID,
			apiAccountTxDo.IP,
		).Where(
			apiAccountTxDo.AppID.Eq(app.ID),
			apiAccountTxDo.AccountID.Eq(req.AccountID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve api account information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 更新数据库中请求凭证账号信息。
	var now time.Time
	{
		log.Info(ctx, "update api account")
		now = time.Now()
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		_, err = apiAccountTxDo.WithContext(ctx).Where(
			apiAccountTxDo.ID.Eq(apiAccount.ID),
		).UpdateColumnSimple(
			apiAccountTxDo.IP.Value(model.StringList(newIPs)),
			apiAccountTxDo.Frequency.Value(req.Frequency),
			apiAccountTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update api account information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 删除数据库中请求凭证账号的授权项。
	{
		log.Info(ctx, "delete old api account")
		apiAuthorizationTxDo := conn.MySQLTxClient(ctx).APIAuthorization
		_, err = apiAuthorizationTxDo.WithContext(ctx).Where(
			apiAuthorizationTxDo.APIAccountID.Eq(apiAccount.ID),
		).Delete()
		if err != nil {
			log.Error(ctx, "failed to delete api account authorities in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将新的请求凭证账号授权项保存到数据库。
	{
		apiAuthorizations := util.ListTo(req.Authorities, func(e int) *model.APIAuthorization {
			return &model.APIAuthorization{
				APIAccountID: apiAccount.ID,
				Capability:   e,
			}
		})
		if len(apiAuthorizations) > 0 {
			log.Info(ctx, "add api authorizations")
			apiAuthorizationTxDo := conn.MySQLTxClient(ctx).APIAuthorization
			err = apiAuthorizationTxDo.WithContext(ctx).
				CreateInBatches(apiAuthorizations, cfg.Get().MySQL().MaximumNumberOfPerSQLInsert())
			if err != nil {
				log.Error(ctx, "failed to add api authorization to database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 将应用事件保存到数据库中。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeUpdateOpenAPI,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser: user.NameEn,
			EventApp:  app.Name,
			EventAuth: req.AccountID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"frequency":   req.Frequency,
				"authorities": req.Authorities,
				"requestIp":   apiAccount.IP,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// OpenWebGetInformation 获取请求凭证信息。
func OpenWebGetInformation(ctx context.Context, req *protocol.OpenWebGetInformationReq) (
	rsp *protocol.OpenWebGetInformationRsp, err error) {

	// 获取上下文信息。
	var appInfo *model.App
	var userInfo *model.User
	{
		log.Info(ctx, "get context information")
		appInfo = ctxs.App(ctx)
		userInfo = ctxs.User(ctx)
		if appInfo == nil || userInfo == nil {
			log.Warn(ctx, "unknown context", userInfo, appInfo)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 从数据中获取请求凭证账号信息。
	var apiAccountInfo *model.APIAccount
	{
		log.Info(ctx, "get api account information from database")
		apiAccountQuery := conn.MySQLClient(ctx).APIAccount
		apiAccountInfo, err = apiAccountQuery.WithContext(ctx).Select(
			apiAccountQuery.ID,
			apiAccountQuery.UserID,
			apiAccountQuery.IP,
			apiAccountQuery.Frequency,
			apiAccountQuery.ExpiredTime,
			apiAccountQuery.CreatedTime,
		).Where(
			apiAccountQuery.AppID.Eq(appInfo.ID),
			apiAccountQuery.AccountID.Eq(req.AccountID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "api account not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve api account information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 从数据库中获取请求凭证账号授权项。
	var authorities []int
	{
		log.Info(ctx, "get authorities")
		apiAuthorizationQuery := conn.MySQLClient(ctx).APIAuthorization
		err = apiAuthorizationQuery.WithContext(ctx).Select(
			apiAuthorizationQuery.Capability,
		).Where(
			apiAuthorizationQuery.APIAccountID.Eq(apiAccountInfo.ID),
		).Scan(&authorities)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve api authorizations from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取用户名。
	var userName string
	{
		userName = userInfo.NameEn
		if apiAccountInfo.UserID != userInfo.ID {
			log.Info(ctx, "get user name from database")
			userQuery := conn.MySQLClient(ctx).User
			err = userQuery.WithContext(ctx).Select(
				userQuery.NameEn,
			).Where(
				userQuery.ID.Eq(apiAccountInfo.UserID),
			).Scan(&userName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user name from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	rsp = &protocol.OpenWebGetInformationRsp{
		AppID:       appInfo.AppID,
		AccountID:   req.AccountID,
		Authorities: authorities,
		RequestIP:   strings.Join(apiAccountInfo.IP, ","),
		Frequency:   apiAccountInfo.Frequency,
		Creator:     userName,
		Expiration:  formatTime(&apiAccountInfo.ExpiredTime),
		CreatedTime: formatTime(&apiAccountInfo.CreatedTime),
	}

	return
}

// OpenWebList 获取请求凭证信息列表。
func OpenWebList(ctx context.Context, req *protocol.OpenWebListReq) (rsp *protocol.OpenWebListRsp, err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 从数据库中获取请求凭证账号信息。
	var apiAccounts []*model.APIAccount
	{
		log.Info(ctx, "get api account list information")
		rsp = &protocol.OpenWebListRsp{}
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		rsp.Count, err = apiAccountDo.WithContext(ctx).Where(
			apiAccountDo.AppID.Eq(app.ID),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to count the number of api accounts in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		apiAccounts, err = apiAccountDo.WithContext(ctx).Select(
			apiAccountDo.IP,
			apiAccountDo.UserID,
			apiAccountDo.AccountID,
			apiAccountDo.ID,
			apiAccountDo.ExpiredTime,
			apiAccountDo.CreatedTime,
			apiAccountDo.Frequency,
		).Where(
			apiAccountDo.AppID.Eq(app.ID),
		).Order(apiAccountDo.ID.Desc()).Offset((req.PageNumber - 1) * req.PageSize).Limit(req.PageSize).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve api accounts from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(apiAccounts) <= 0 {
			return
		}
	}

	// 从数据库中获取请求凭证账号的授权项。
	var userIDs []int
	var accountIDToCapabilities map[int][]int
	{
		log.Info(ctx, "get api authorizations")
		var accountIDs []int
		accountIDs, userIDs = util.ListTo2(apiAccounts,
			func(e *model.APIAccount) (int, int) { return e.ID, e.UserID })
		type sqlResult struct {
			AccountID    int
			Capabilities string
		}
		apiAuthorizationDo := conn.MySQLClient(ctx).APIAuthorization
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		var capabilities []*sqlResult
		err = apiAuthorizationDo.WithContext(ctx).Select(
			apiAuthorizationDo.APIAccountID.As("AccountID"),
			apiAuthorizationDo.Capability.GroupConcat().As("Capabilities"),
		).Join(apiAccountDo, apiAccountDo.ID.EqCol(apiAuthorizationDo.APIAccountID)).Where(
			apiAuthorizationDo.APIAccountID.In(util.CleanNumbers(accountIDs)...),
		).Group(apiAuthorizationDo.APIAccountID).Scan(&capabilities)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve api authorizations from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		accountIDToCapabilities = util.ListToMap(capabilities, func(e *sqlResult) (int, []int) {
			return e.AccountID, util.CleanNumbers(util.ListTo(strings.Split(e.Capabilities, ","),
				func(e string) int { return util.Atoi(e) }))
		})
	}

	// 从数据库中，获取用户名。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users information")
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		rsp.List = make([]*protocol.OpenWebListItem, len(apiAccounts))
		for i, v := range apiAccounts {
			rsp.List[i] = &protocol.OpenWebListItem{
				AccountID:   v.AccountID,
				Authorities: accountIDToCapabilities[v.ID],
				Frequency:   v.Frequency,
				RequestIP:   strings.Join(v.IP, ","),
				Creator:     userIDToName[v.UserID],
				Expiration:  formatTime(&v.ExpiredTime),
				CreatedTime: formatTime(&v.CreatedTime),
			}
		}
	}

	return
}

// OpenWebRenewal 续期。
func OpenWebRenewal(ctx context.Context, req *protocol.OpenWebRenewalReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 查询数据库获取请求凭证账号信息。
	var apiAccountID int
	{
		log.Info(ctx, "get api account information")
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		err = apiAccountDo.WithContext(ctx).Select(
			apiAccountDo.ID,
		).Where(
			apiAccountDo.AppID.Eq(app.ID),
			apiAccountDo.AccountID.Eq(req.AccountID),
		).Scan(&apiAccountID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "api account not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve api account from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 更新请求凭证账号的过期时间。
	var now time.Time
	var expiration time.Time
	{
		log.Info(ctx, "update api account expiration time")
		now = time.Now()
		expiration = now.Add(consts.APIAccountExpirationTime)
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		_, err = apiAccountTxDo.WithContext(ctx).Where(
			apiAccountTxDo.ID.Eq(apiAccountID),
		).UpdateColumnSimple(
			apiAccountTxDo.ExpiredTime.Value(expiration),
			apiAccountTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update api account expiration time", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件信息记录到数据库。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRenewOpenAPI,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser: user.NameEn,
			EventApp:  app.Name,
			EventAuth: req.AccountID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"expiration": formatTime(&expiration),
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// OpenWebReset 重置密钥。
func OpenWebReset(ctx context.Context, req *protocol.OpenWebResetReq) (rsp *protocol.OpenWebResetRsp, err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 从数据库中获取请求凭证账号信息。
	var apiAccountID int
	{
		log.Info(ctx, "get api account information")
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		err = apiAccountDo.WithContext(ctx).Select(
			apiAccountDo.ID,
		).Where(
			apiAccountDo.AppID.Eq(app.ID),
			apiAccountDo.AccountID.Eq(req.AccountID),
		).Scan(&apiAccountID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "api account not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve api account information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 更新数据库中请求凭证账号密钥。
	var secret string
	var now time.Time
	{
		log.Info(ctx, "update api account secret")
		now = time.Now()
		secret = util.RandomPrintableASCIINoSpaceString(consts.ApiAccountSecretLength)
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		var sqlResult gen.ResultInfo
		sqlResult, err = apiAccountTxDo.WithContext(ctx).Where(
			apiAccountTxDo.ID.Eq(apiAccountID),
		).UpdateColumnSimple(
			apiAccountTxDo.Secret.Value(secret),
			apiAccountTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update api account information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "updating api account secret has no effect")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	// 保存应用事件信息到数据库。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeResetOpenAPI,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser: user.NameEn,
			EventApp:  app.Name,
			EventAuth: req.AccountID,
		})
		if err != nil {
			return
		}
	}

	rsp = &protocol.OpenWebResetRsp{Password: secret}

	return
}

// OpenWebRemove 删除。
func OpenWebRemove(ctx context.Context, req *protocol.OpenWebRemoveReq) (err error) {
	// 获取上下文信息。
	var app *model.App
	var user *model.User
	{
		log.Info(ctx, "get context information")
		app = ctxs.App(ctx)
		user = ctxs.User(ctx)
		if app == nil || user == nil {
			log.Warn(ctx, "unknown context", user, app)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
	}

	// 从数据库中查询请求凭证账号信息。
	var apiAccount *model.APIAccount
	{
		log.Info(ctx, "get api account information")
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		apiAccount, err = apiAccountDo.WithContext(ctx).Select(
			apiAccountDo.ID,
			apiAccountDo.IP,
			apiAccountDo.AccountID,
			apiAccountDo.Frequency,
			apiAccountDo.ExpiredTime,
			apiAccountDo.CreatedTime,
		).Where(
			apiAccountDo.AppID.Eq(app.ID),
			apiAccountDo.AccountID.Eq(req.AccountID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "api account not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve api account information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 删除数据库中的请求凭证账号。
	{
		log.Info(ctx, "remove api account")
		apiAccountTxDo := conn.MySQLTxClient(ctx).APIAccount
		var sqlResult gen.ResultInfo
		sqlResult, err = apiAccountTxDo.WithContext(ctx).Where(
			apiAccountTxDo.ID.Eq(apiAccount.ID),
		).Delete()
		if err != nil {
			log.Error(ctx, "failed to delete api account information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "deleting api account information has no effect")
			return
		}
	}

	// 删除数据库中的请求凭证账号授权项。
	var authorities []int
	{
		log.Info(ctx, "remove api authorizations")
		apiAuthorizationTxDo := conn.MySQLTxClient(ctx).APIAuthorization
		err = apiAuthorizationTxDo.WithContext(ctx).Select(
			apiAuthorizationTxDo.Capability,
		).Where(
			apiAuthorizationTxDo.APIAccountID.Eq(apiAccount.ID),
		).Scan(&authorities)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve api authorizations in database", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		_, err = apiAuthorizationTxDo.WithContext(ctx).Where(
			apiAuthorizationTxDo.APIAccountID.Eq(apiAccount.ID),
		).Delete()
		if err != nil {
			log.Error(ctx, "failed to delete api authorizations in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 记录应用事件到数据库中。
	{
		log.Info(ctx, "save app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRemoveOpenAPI,
			CreatedTime: time.Now(),
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventUser: user.NameEn,
			EventApp:  app.Name,
			EventAuth: req.AccountID,
			EventDetail: util.GetPrintJSON(map[string]any{
				"requestIp":   apiAccount.IP,
				"frequency":   apiAccount.Frequency,
				"authorities": authorities,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// APIAccountHasAnyRight 凭证是否有权限。
func APIAccountHasAnyRight(ctx context.Context, apiAccountID int, perm int, perms ...int) (bool, error) {
	apiAuthorizationDo := conn.MySQLClient(ctx).APIAuthorization
	count, err := apiAuthorizationDo.WithContext(ctx).Where(
		apiAuthorizationDo.APIAccountID.Eq(apiAccountID),
		apiAuthorizationDo.Capability.In(append(perms, perm)...),
	).Count()
	if err != nil {
		log.Error(ctx, "failed to retrieve api authorization information", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	return count > 0, nil
}

// GetAPIAccountNamesByIDs 获取凭证名。
func GetAPIAccountNamesByIDs(ctx context.Context, apiAccountIDs []int) (map[int]string, error) {
	if len(apiAccountIDs) <= 0 {
		return make(map[int]string), nil
	}
	apiAccountDo := conn.MySQLClient(ctx).APIAccount
	apiAccounts, err := apiAccountDo.WithContext(ctx).Select(
		apiAccountDo.ID,
		apiAccountDo.AccountID,
	).Where(
		apiAccountDo.ID.In(apiAccountIDs...),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve api accounts from database", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return util.ListToMap(apiAccounts, func(e *model.APIAccount) (int, string) { return e.ID, e.AccountID }), nil
}
