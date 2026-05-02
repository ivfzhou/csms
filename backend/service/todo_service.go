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
	"slices"
	"strconv"
	"time"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

// TodoWebCount 获取用户待办、已办数量。
func TodoWebCount(ctx context.Context) (rsp *protocol.TodoWebCountRsp, err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询待办数量。
	{
		log.Info(ctx, "query user todos that user needs to handle")
		rsp = &protocol.TodoWebCountRsp{}
		todoDo := conn.MySQLClient(ctx).Todo
		rsp.NeedToDeal, err = todoDo.WithContext(ctx).Where(
			todoDo.Status.Eq(model.TodoStatusProcessing),
			query.FindInSetWithNumber(todoDo.Candidates, user.ID),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve todo information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取已处理的待办数。
	{
		log.Info(ctx, "query user todos that have been processed")
		todoDo := conn.MySQLClient(ctx).Todo
		rsp.Done, err = todoDo.WithContext(ctx).Where(
			todoDo.ApproverID.Eq(user.ID),
			todoDo.Status.In(model.TodoStatusApproved, model.TodoStatusRejected),
		).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve todo information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// TodoWebList 获取需要处理的待办。
func TodoWebList(ctx context.Context, req *protocol.TodoWebListReq) (rsp *protocol.TodoWebListRsp, err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查询用户需要处理的待办。
	var todos []*model.Todo
	{
		log.Info(ctx, "get user todos information")
		todoDo := conn.MySQLClient(ctx).Todo
		rsp = &protocol.TodoWebListRsp{}
		statement := todoDo.WithContext(ctx).Select(
			todoDo.ID,
			todoDo.CreatedTime,
			todoDo.ApplierID,
			todoDo.AppID,
			todoDo.Type,
			todoDo.Status,
		).Where(
			todoDo.Status.Eq(model.TodoStatusProcessing),
			query.FindInSetWithNumber(todoDo.Candidates, user.ID),
		)
		rsp.Count, err = statement.Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve todos information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		todos, err = statement.Order(todoDo.CreatedTime.Desc(), todoDo.ID.Desc()).
			Offset((req.PageNumber - 1) * req.PageSize).Limit(req.PageSize).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve todos information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(todos) <= 0 {
			return
		}
	}

	// 获取用户名。
	var appIDs []int
	var userIDToName map[int]string
	{
		log.Info(ctx, "get user information")
		var userIDs []int
		userIDs, appIDs = util.ListTo2(todos, func(e *model.Todo) (int, int) { return e.ApplierID, e.AppID })
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
	}

	// 获应用信息。
	var appIDToInfo map[int]*model.App
	{
		log.Info(ctx, "query app information from database")
		appDo := conn.MySQLClient(ctx).App
		var apps []*model.App
		apps, err = appDo.WithContext(ctx).Select(
			appDo.ID,
			appDo.Name,
			appDo.Platform,
		).Where(
			appDo.ID.In(util.CleanNumbers(appIDs)...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apps information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		appIDToInfo = util.ListAssociateBy(apps, func(e *model.App) int { return e.ID })
	}

	// 组装数据。
	{
		list := make([]*protocol.TodoWebListItem, len(todos))
		for i, v := range todos {
			appInfo, ok := appIDToInfo[v.AppID]
			if !ok {
				appInfo = &model.App{}
			}
			list[i] = &protocol.TodoWebListItem{
				ID:          v.ID,
				AppName:     appInfo.Name,
				CreatedTime: formatTime(&v.CreatedTime),
				Creator:     userIDToName[v.ApplierID],
				Platform:    appInfo.Platform,
				Type:        v.Type,
				Status:      v.Status,
			}
		}
		rsp.List = list
	}

	return
}

// TodoWebListDealt 获取已处理的待办列表。
func TodoWebListDealt(ctx context.Context, req *protocol.TodoWebListDealtReq) (
	rsp *protocol.TodoWebListDealtRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取应用 ID。
	var appID int
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app id")
			appDo := conn.MySQLClient(ctx).App
			err = appDo.WithContext(ctx).Select(
				appDo.ID,
			).Where(
				appDo.AppID.Eq(req.AppID),
			).Scan(&appID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve app id from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 判断用户是否有这个应用 ID 的权限。
	{
		if appID > 0 {
			log.Info(ctx, "verify user permission")
			var hasRight bool
			hasRight, err = UserHasAnyRight(ctx, user.ID, appID,
				model.UserRoleSystemAdmin, model.UserRoleAppAdmin, model.UserRoleAppMember)
			if err != nil {
				return
			}
			if !hasRight {
				log.Warn(ctx, "user does not have app right")
				return
			}
		}
	}

	// 查询数据库，获取符合条件的待办。
	var todos []*model.Todo
	{
		log.Info(ctx, "retrieve todos information from database")
		todoDo := conn.MySQLClient(ctx).Todo
		statement := todoDo.WithContext(ctx).Select(
			todoDo.AppID,
			todoDo.ApplierID,
			todoDo.ApproverID,
			todoDo.ID,
			todoDo.Type,
			todoDo.Status,
			todoDo.CreatedTime,
			todoDo.FinishedTime,
		)
		if appID > 0 {
			statement = statement.Where(todoDo.AppID.Eq(appID))
		}
		if len(req.Status) > 0 {
			statement = statement.Where(todoDo.Status.In(req.Status...))
		}
		if len(req.Types) > 0 {
			statement = statement.Where(todoDo.Type.In(req.Types...))
		}
		rsp = &protocol.TodoWebListDealtRsp{}
		// 只查询可由用户审批的待办。或者已经由用户处理了的待办。
		statement = statement.Where(
			todoDo.WithContext(ctx).Where(query.FindInSetWithNumber(todoDo.Candidates, user.ID)).
				Or(todoDo.ApproverID.Eq(user.ID)),
		)
		if rsp.Count, err = statement.Count(); err != nil {
			log.Error(ctx, "failed to query todos information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		todos, err = statement.Order(todoDo.CreatedTime.Desc(), todoDo.ID.Desc()).
			Offset((req.PageNumber - 1) * req.PageSize).Limit(req.PageSize).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve todo information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(todos) <= 0 {
			return
		}
	}

	// 获取应用信息。
	var userIDs []int
	var appIDToInfo map[int]*model.App
	{
		log.Info(ctx, "get app information")
		var appIDs []int
		userIDs, appIDs = util.ListToUnique2(todos,
			func(e *model.Todo) ([]int, int) { return []int{e.ApplierID, e.ApproverID}, e.AppID })
		appDo := conn.MySQLClient(ctx).App
		var apps []*model.App
		apps, err = appDo.WithContext(ctx).Select(
			appDo.ID,
			appDo.Name,
			appDo.Platform,
		).Where(
			appDo.ID.In(util.CleanNumbers(appIDs)...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve app information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		appIDToInfo = util.ListAssociateBy(apps, func(e *model.App) int { return e.ID })
	}

	// 获取用户信息。
	var userIDToName map[int]string
	{
		log.Info(ctx, "get users name")
		userIDToName, err = GetUserNamesByIDs(ctx, util.CleanNumbers(userIDs))
		if err != nil {
			return
		}
	}

	// 组装数据。
	{
		rsp.List = make([]*protocol.TodoWebListDealtItem, len(todos))
		for i, v := range todos {
			item := &protocol.TodoWebListDealtItem{
				ID:           v.ID,
				CreatedTime:  formatTime(&v.CreatedTime),
				Creator:      userIDToName[v.ApplierID],
				Type:         v.Type,
				Status:       v.Status,
				FinishedTime: formatTime(&v.FinishedTime),
				Approver:     userIDToName[v.ApproverID],
			}
			app, ok := appIDToInfo[v.AppID]
			if ok {
				item.AppName = app.Name
				item.Platform = app.Platform
			}
			rsp.List[i] = item
		}
	}

	return
}

// TodoWebCreate 创建待办。
func TodoWebCreate(ctx context.Context, req *protocol.TodoWebCreateReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取应用信息。
	var appID int
	{
		log.Info(ctx, "get app information")
		appDo := conn.MySQLClient(ctx).App
		var app *model.App
		app, err = appDo.WithContext(ctx).Select(
			appDo.ID,
			appDo.Status,
		).Where(
			appDo.AppID.Eq(req.AppID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "app not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve app information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if app.Status != model.AppStatusValid {
			err = errs.New(consts.ErrAppStatusNotValid)
			return
		}
		appID = app.ID
	}

	// 根据不同类型处理。
	{
		switch req.Type {
		case model.TodoTypeJoinApp:
			// 判断用户是否已是应用成员。
			log.Info(ctx, "get user app roles information")
			userRoleDo := conn.MySQLClient(ctx).UserRole
			var userRoleIDs []int
			err = userRoleDo.WithContext(ctx).Select(
				userRoleDo.Role,
			).Where(
				userRoleDo.UserID.Eq(user.ID),
				userRoleDo.AppID.Eq(appID),
			).Scan(&userRoleIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user roles information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if util.ContainsAny(userRoleIDs, model.UserRoleAppMember, model.UserRoleAppAdmin) {
				log.Warn(ctx, "user already in app")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 判断不存在相同的待办。
			log.Info(ctx, "get existing todo")
			todoTxDo := conn.MySQLTxClient(ctx).Todo
			var count int64
			count, err = todoTxDo.WithContext(ctx).Clauses(query.ForUpdate()).Where(
				todoTxDo.AppID.Eq(appID),
				todoTxDo.Type.Eq(model.TodoTypeJoinApp),
				todoTxDo.Status.Eq(model.TodoStatusProcessing),
				todoTxDo.ApplierID.Eq(user.ID),
			).Count()
			if count > 0 {
				err = errs.New(consts.ErrSameTodoExist)
				return
			}

			// 获取应用管理员。
			log.Info(ctx, "get app admins")
			var appAdminIDs []int
			err = userRoleDo.WithContext(ctx).Select(
				userRoleDo.UserID,
			).Where(
				userRoleDo.AppID.Eq(appID),
				userRoleDo.Role.Eq(model.UserRoleAppAdmin),
			).Scan(&appAdminIDs)
			if err != nil {
				log.Error(ctx, "failed to retrieve app admins in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if len(appAdminIDs) <= 0 {
				log.Error(ctx, "app admins not found")
				err = errs.New(consts.ErrSystem)
				return
			}

			// 添加待办记录。
			log.Info(ctx, "create todo in database")
			err = todoTxDo.WithContext(ctx).Select(
				todoTxDo.AppID,
				todoTxDo.Type,
				todoTxDo.ApplierID,
				todoTxDo.Candidates,
				todoTxDo.ApplyReason,
				todoTxDo.Status,
				todoTxDo.CreatedTime,
			).Create(&model.Todo{
				AppID:       appID,
				Type:        model.TodoTypeJoinApp,
				ApplierID:   user.ID,
				Candidates:  appAdminIDs,
				ApplyReason: req.ApplyReason,
				Status:      model.TodoStatusProcessing,
				CreatedTime: time.Now(),
			})
			if err != nil {
				log.Error(ctx, "failed to create todo in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		case model.TodoTypeApplySigner:
			// 判断用户是否是应用成员。
			log.Info(ctx, "get user app roles")
			userRoleDo := conn.MySQLClient(ctx).UserRole
			var count int64
			count, err = userRoleDo.WithContext(ctx).Where(
				userRoleDo.UserID.Eq(user.ID),
				userRoleDo.AppID.Eq(appID),
				userRoleDo.Role.Eq(model.UserRoleAppMember),
			).Count()
			if err != nil {
				log.Error(ctx, "failed to retrieve user roles information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if count <= 0 {
				log.Warn(ctx, "user is not an app member")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 判断不存在相同的待办。
			log.Info(ctx, "get existing todo")
			todoTxDo := conn.MySQLTxClient(ctx).Todo
			count, err = todoTxDo.WithContext(ctx).Clauses(query.ForUpdate()).Where(
				todoTxDo.AppID.Eq(appID),
				todoTxDo.Type.Eq(model.TodoTypeApplySigner),
				todoTxDo.Status.Eq(model.TodoStatusProcessing),
				todoTxDo.ApplierID.Eq(user.ID),
			).Count()
			if count > 0 {
				err = errs.New(consts.ErrSameTodoExist)
				return
			}

			// 获取应用管理员。
			log.Info(ctx, "get app admins")
			var appAdminIDs []int
			err = userRoleDo.WithContext(ctx).Select(
				userRoleDo.UserID,
			).Where(
				userRoleDo.AppID.Eq(appID),
				userRoleDo.Role.Eq(model.UserRoleAppAdmin),
			).Scan(&appAdminIDs)
			if err != nil {
				log.Error(ctx, "failed to retrieve app admins in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if len(appAdminIDs) <= 0 {
				log.Error(ctx, "app admins not found")
				err = errs.New(consts.ErrSystem)
				return
			}

			// 添加待办记录。
			log.Info(ctx, "create todo in database")
			err = todoTxDo.WithContext(ctx).Select(
				todoTxDo.AppID,
				todoTxDo.Type,
				todoTxDo.ApplierID,
				todoTxDo.Candidates,
				todoTxDo.ApplyReason,
				todoTxDo.Status,
				todoTxDo.CreatedTime,
			).Create(&model.Todo{
				AppID:       appID,
				Type:        model.TodoTypeApplySigner,
				ApplierID:   user.ID,
				Candidates:  appAdminIDs,
				ApplyReason: req.ApplyReason,
				Status:      model.TodoStatusProcessing,
				CreatedTime: time.Now(),
			})
			if err != nil {
				log.Error(ctx, "failed to create todo in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		default:
			log.Warn(ctx, "unknown todo type")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	return
}

// TodoWebGetDetail 获取待办详情。
func TodoWebGetDetail(ctx context.Context, req *protocol.TodoWebGetDetailReq) (
	rsp *protocol.TodoWebGetDetailRsp, err error) {

	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取待办详情。
	var todo *model.Todo
	{
		log.Info(ctx, "get todo detail information")
		todoDo := conn.MySQLClient(ctx).Todo
		todo, err = todoDo.WithContext(ctx).Select(
			todoDo.ApproverID,
			todoDo.Candidates,
			todoDo.AppID,
			todoDo.ApplierID,
			todoDo.ID,
			todoDo.Type,
			todoDo.Status,
			todoDo.CreatedTime,
			todoDo.FinishedTime,
		).Where(
			todoDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "todo not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve todo detail information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 判断用户具有查看待办的权限。
	{
		log.Info(ctx, "verify user permission")
		if todo.ApproverID != user.ID && !slices.Contains(todo.Candidates, user.ID) {
			log.Warn(ctx, "user does not have permission to view todo")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 获取应用信息。
	var app *model.App
	{
		log.Info(ctx, "get app information")
		appDo := conn.MySQLClient(ctx).App
		app, err = appDo.WithContext(ctx).Select(
			appDo.Name,
			appDo.AppID,
			appDo.Platform,
		).Where(
			appDo.ID.Eq(todo.AppID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve app information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if app == nil {
			app = &model.App{}
		}
	}

	// 获取用户信息。
	var userIDToInfo map[int]*model.User
	{
		log.Info(ctx, "get user information")
		userDo := conn.MySQLClient(ctx).User
		var users []*model.User
		users, err = userDo.WithContext(ctx).Select(
			userDo.ID,
			userDo.NameEn,
			userDo.Department,
		).Where(
			userDo.ID.In(append(todo.Candidates, todo.ApplierID, todo.ApproverID)...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve users information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userIDToInfo = util.ListAssociateBy(users, func(e *model.User) int { return e.ID })
	}

	// 获取测试设备信息。
	var deviceUdid string
	var deviceModel string
	{
		log.Info(ctx, "get apple device information")
		if todo.Type == model.TodoTypeRegisterAppleDevice {
			deviceID, _ := strconv.Atoi(todo.Information)
			if deviceID > 0 {
				appleDeviceDo := conn.MySQLClient(ctx).AppleDevice
				var appleDevice *model.AppleDevice
				appleDevice, err = appleDeviceDo.WithContext(ctx).Select(
					appleDeviceDo.Model,
					appleDeviceDo.Udid,
				).Where(
					appleDeviceDo.ID.Eq(deviceID),
				).Take()
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(ctx, "failed to retrieve apple device information from database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
				if appleDevice != nil {
					deviceUdid = appleDevice.Udid
					deviceModel = appleDevice.Model
				}
			}
		}
	}

	// 组装数据。
	{
		candidates := make([]string, 0, len(todo.Candidates))
		for _, v := range todo.Candidates {
			user = userIDToInfo[v]
			if user == nil {
				continue
			}
			candidates = append(candidates, user.NameEn)
		}
		user = userIDToInfo[todo.ApplierID]
		if user == nil {
			user = &model.User{}
		}
		user2 := userIDToInfo[todo.ApproverID]
		if user2 == nil {
			user2 = &model.User{}
		}
		rsp = &protocol.TodoWebGetDetailRsp{
			ID:           todo.ID,
			AppID:        app.AppID,
			AppName:      app.Name,
			CreatedTime:  formatTime(&todo.CreatedTime),
			Creator:      user.NameEn,
			Platform:     app.Platform,
			Type:         todo.Type,
			Status:       todo.Status,
			Department:   user.Department,
			Candidates:   candidates,
			ApproveBy:    user2.NameEn,
			DeviceUdid:   deviceUdid,
			DeviceModel:  deviceModel,
			ApplyReason:  todo.ApplyReason,
			ApproveMsg:   todo.ApproveMessage,
			FinishedTime: formatTime(&todo.FinishedTime),
		}
	}

	return
}

// TodoWebDeal 审批。
func TodoWebDeal(ctx context.Context, req *protocol.TodoWebDealReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取待办信息。
	var todo *model.Todo
	{
		log.Info(ctx, "query todo information from database")
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		todo, err = todoTxDo.WithContext(ctx).Clauses(query.ForUpdate()).Select(
			todoTxDo.Candidates,
			todoTxDo.Status,
			todoTxDo.Type,
			todoTxDo.AppID,
			todoTxDo.ApplierID,
			todoTxDo.Information,
		).Where(
			todoTxDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "todo not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to retrieve todo information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 判断用户具有审批待办的权限。
	{
		log.Info(ctx, "verify user permission")
		if !slices.Contains(todo.Candidates, user.ID) || todo.Status != model.TodoStatusProcessing {
			log.Info(ctx, "user does not have permission to deal")
			err = errs.New(consts.ErrPermissionDenied)
			return
		}
	}

	// 根据类型处理。
	var now time.Time
	{
		now = time.Now()
		switch todo.Type {
		// 更新应用状态。
		case model.TodoTypeRegisterApp:
			log.Info(ctx, "update app information")
			status := model.AppStatusRejected
			if req.IsPass {
				status = model.AppStatusValid
			}
			appTxDo := conn.MySQLTxClient(ctx).App
			var sqlResult gen.ResultInfo
			sqlResult, err = appTxDo.WithContext(ctx).Where(
				appTxDo.ID.Eq(todo.AppID),
				appTxDo.Status.Eq(model.AppStatusApproving),
			).UpdateColumnSimple(
				appTxDo.Status.Value(status),
				appTxDo.UpdatedTime.Value(now),
			)
			if err != nil {
				log.Error(ctx, "failed to update app information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if sqlResult.RowsAffected <= 0 {
				log.Warn(ctx, "no app updated")
				err = errs.New(consts.ErrSystem)
				return
			}

		// 给用户添加应用权限。
		case model.TodoTypeJoinApp:
			log.Info(ctx, "add user role")
			if req.IsPass {
				userRoleTxDo := conn.MySQLTxClient(ctx).UserRole
				err = userRoleTxDo.WithContext(ctx).Create(&model.UserRole{
					AppID:  todo.AppID,
					UserID: todo.ApplierID,
					Role:   model.UserRoleAppMember,
				})
				if err != nil {
					log.Error(ctx, "failed to add user role in database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
			}

		// 给用户添加应用签名权限。
		case model.TodoTypeApplySigner:
			log.Info(ctx, "add user role")
			if req.IsPass {
				userRoleTxDo := conn.MySQLTxClient(ctx).UserRole
				err = userRoleTxDo.WithContext(ctx).Create(&model.UserRole{
					AppID:  todo.AppID,
					UserID: todo.ApplierID,
					Role:   model.UserRoleAppSigner,
				})
				if err != nil {
					log.Error(ctx, "failed to add user role in database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
			}

		// 更新测试设备状态，申请注册设备。
		case model.TodoTypeRegisterAppleDevice:
			log.Info(ctx, "register apple device")
			appleDeviceID := util.Atoi(todo.Information)
			status := model.AppleDeviceStatusRejected
			var appleAPIResult *appleAPIResponse
			if req.IsPass {
				// 查询设备信息。
				log.Info(ctx, "get apple device information")
				appleDeviceDo := conn.MySQLClient(ctx).AppleDevice
				var appleDeviceInfo *model.AppleDevice
				appleDeviceInfo, err = appleDeviceDo.WithContext(ctx).Select(
					appleDeviceDo.Udid,
					appleDeviceDo.Platform,
				).Where(
					appleDeviceDo.ID.Eq(appleDeviceID),
				).Take()
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						err = errs.New(consts.ErrAppleDeviceNotFound)
						return
					}
					log.Error(ctx, "failed to retrieve apple device information from database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}

				// 注册设备。
				log.Info(ctx, "register apple device to apple")
				var deviceName string
				deviceName, err = generateID(ctx, IDAppleDevice)
				if err != nil {
					return
				}
				defer func() {
					if err != nil {
						log.ErrorIf(ctx, reclaimID(ctx, IDAppleDevice, deviceName),
							"failed to reclaim apple device name")
					}
				}()
				var token string
				token, err = generateAppleAPIToken(ctx)
				if err != nil {
					return
				}
				appleAPIResult, err = httpAppleAPIRegisterDevice(
					ctx, token, deviceName, appleDeviceInfo.Udid, appleDeviceInfo.Platform)
				if err != nil {
					return
				}

				status = model.AppleDeviceStatusOK
			}
			appleDeviceTxDo := conn.MySQLTxClient(ctx).AppleDevice
			var sqlResult gen.ResultInfo
			assignExprs := make([]field.AssignExpr, 0, 5)
			assignExprs = append(assignExprs,
				appleDeviceTxDo.Status.Value(status),
				appleDeviceTxDo.UpdatedTime.Value(now),
			)
			if appleAPIResult != nil {
				assignExprs = append(assignExprs,
					appleDeviceTxDo.InAppleID.Value(appleAPIResult.Data.ID),
					appleDeviceTxDo.Model.Value(appleAPIResult.Data.Attributes.Model),
					appleDeviceTxDo.Platform.Value(appleAPIResult.Data.Attributes.Platform),
				)
			}
			sqlResult, err = appleDeviceTxDo.WithContext(ctx).Where(
				appleDeviceTxDo.ID.Eq(appleDeviceID),
				appleDeviceTxDo.Status.Eq(model.AppleDeviceStatusApproving),
			).UpdateColumnSimple(assignExprs...)
			if err != nil {
				log.Error(ctx, "failed to update apple device information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if sqlResult.RowsAffected <= 0 {
				log.Warn(ctx, "no apple device updated")
				err = errs.New(consts.ErrSystem)
				return
			}

		// 更新应用状态。
		case model.TodoTypeActivateApp:
			log.Info(ctx, "update app information")
			if req.IsPass {
				appTxDo := conn.MySQLTxClient(ctx).App
				var sqlResult gen.ResultInfo
				sqlResult, err = appTxDo.WithContext(ctx).Where(
					appTxDo.ID.Eq(todo.AppID),
					appTxDo.Status.Eq(model.AppStatusApproving),
				).UpdateColumnSimple(
					appTxDo.Status.Value(model.AppStatusValid),
					appTxDo.UpdatedTime.Value(now),
				)
				if err != nil {
					log.Error(ctx, "failed to update app information in database", err)
					err = errs.NewWithError(consts.ErrSystem, err)
					return
				}
				if sqlResult.RowsAffected <= 0 {
					log.Warn(ctx, "no app updated")
					err = errs.New(consts.ErrSystem)
					return
				}
			}

		default:
			log.Error(ctx, "unhandled todo type", todo.Type)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 更新待办状态。
	{
		log.Info(ctx, "update todo information")
		status := model.TodoStatusRejected
		if req.IsPass {
			status = model.TodoStatusApproved
		}
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		var sqlResult gen.ResultInfo
		sqlResult, err = todoTxDo.WithContext(ctx).Where(
			todoTxDo.ID.Eq(req.ID),
			todoTxDo.Status.Eq(model.TodoStatusProcessing),
		).UpdateColumnSimple(
			todoTxDo.Status.Value(status),
			todoTxDo.ApproverID.Value(user.ID),
			todoTxDo.ApproveMessage.Value(req.ApproveMessage),
			todoTxDo.FinishedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update todo information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Error(ctx, "no todo updated")
		}
	}

	return
}
