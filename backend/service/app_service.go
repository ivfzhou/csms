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
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"slices"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
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
	tus "gitee.com/ivfzhou/tus_client/v2"
)

// AppWebRegister 注册。
func AppWebRegister(ctx context.Context, req *protocol.AppWebRegisterReq) (err error) {
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

	// 查询系统管理员。
	var systemAdmins []int
	{
		log.Info(ctx, "get system administrators")
		userRoleDo := conn.MySQLClient(ctx).UserRole
		err = userRoleDo.WithContext(ctx).Select(
			userRoleDo.UserID,
		).Where(
			userRoleDo.Role.Eq(model.UserRoleSystemAdmin),
		).Scan(&systemAdmins)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve system administrators from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(systemAdmins) <= 0 {
			// 没有系统管理员。
			log.Error(ctx, "no system administrators were found")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取应用图标文件。
	var logoBytes []byte
	{
		log.Info(ctx, "get app logo file")
		var logoStream multipart.File
		logoStream, err = req.Logo.Open()
		if err != nil {
			log.Error(ctx, "opening app logo file failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.CloseIO(ctx, logoStream)
		logoBytes, err = io.ReadAll(logoStream)
		if err != nil {
			log.Error(ctx, "reading app logo file failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验应用图标文件。
	{
		log.Info(ctx, "verify app logo file")
		appLogoMaximumSize := cfg.Get().Backend().AppLogoMaximumSize()
		if len(logoBytes) > appLogoMaximumSize {
			log.Warn(ctx, "app logo file too big, max size is", appLogoMaximumSize)
			err = errs.New(consts.ErrAppLogoTooLarge)
			return
		}
		var logoFmt string
		_, logoFmt, err = image.Decode(bytes.NewReader(logoBytes))
		if err != nil {
			err = errs.NewWithError(consts.ErrAppLogoFormatNotSupported, err)
			return
		}
		if !slices.Contains(consts.SupportAppLogoFmt, logoFmt) {
			log.Warn(ctx, "app logo file format is invalid", logoFmt, ". support", consts.SupportAppLogoFmt)
			err = errs.New(consts.ErrAppLogoFormatNotSupported)
			return
		}
	}

	// 整理和校验应用管理员和成员。
	var adminIDs []int
	var memberIDs []int
	{
		log.Info(ctx, "tidy up members and administrators of app")
		userNames := util.CleanStrings(append(req.Admins, req.Members...))
		adminIDs = make([]int, 0, len(req.Admins)+1)
		memberIDs = make([]int, 0, len(req.Members))
		if len(userNames) > 0 {
			userDo := conn.MySQLClient(ctx).User
			var users []*model.User
			users, err = userDo.WithContext(ctx).Select(
				userDo.ID,
				userDo.NameEn,
			).Where(
				userDo.NameEn.In(userNames...),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve user information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if len(users) != len(userNames) {
				log.Warn(ctx, "usernames length is invalid", userNames)
				err = errs.New(consts.ErrUserNotExists)
				return
			}
			for _, v := range users {
				if slices.Contains(req.Admins, v.NameEn) {
					adminIDs = append(adminIDs, v.ID)
				}
				if slices.Contains(req.Members, v.NameEn) {
					memberIDs = append(memberIDs, v.ID)
				}
			}
		}
		// 将用户自己加入管理员。
		if !slices.Contains(adminIDs, user.ID) {
			adminIDs = append(adminIDs, user.ID)
		}
	}

	// 上传应用图标文件到 Tusd。
	var tusdID string
	{
		log.Info(ctx, "uploading app logo file to tusd")
		tusdID, err = conn.TusdClient(ctx).UploadFile(ctx, logoBytes)
		if err != nil {
			log.Error(ctx, "failed to upload app logo file to tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		// 注册失败就删除应用图标文件。
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, tusdID),
					"failed to delete app logo file in tusd", tusdID)
			}
		}()
	}

	// 将应用信息保存到数据库。
	var app *model.App
	var fileID string
	{
		log.Info(ctx, "save app information")
		fileID, err = generateID(ctx, IDFile)
		if err != nil {
			return
		}
		var appID string
		appID, err = generateID(ctx, IDApp)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				// 注册失败就回收文件 ID。
				log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id")
				log.ErrorIf(ctx, reclaimID(ctx, IDApp, appID), "failed to reclaim app id")
			}
		}()
		app = &model.App{
			AppID:       appID,
			UserID:      user.ID,
			Name:        req.Name,
			LogoFileID:  fileID,
			Platform:    req.Platform,
			Status:      model.AppStatusApproving,
			CreatedTime: time.Now(),
		}
		appTxDo := conn.MySQLTxClient(ctx).App
		if err = appTxDo.WithContext(ctx).Select(
			appTxDo.AppID,
			appTxDo.UserID,
			appTxDo.Name,
			appTxDo.LogoFileID,
			appTxDo.Platform,
			appTxDo.Status,
			appTxDo.CreatedTime,
		).Create(app); err != nil {
			log.Error(ctx, "failed to save app information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将文件信息保存到数据库。
	var fileName string
	{
		log.Info(ctx, "save file information")
		logoDigest := md5.Sum(logoBytes)
		fileName = filepath.Base(req.Logo.Filename)
		if err = createFile(ctx, &model.File{
			FileID:      fileID,
			TusdID:      tusdID,
			UserID:      user.ID,
			AppID:       app.ID,
			Name:        fileName,
			Md5:         hex.EncodeToString(logoDigest[:]),
			Size:        int(req.Logo.Size),
			Type:        model.FileTypeAppLogo,
			CreatedTime: app.CreatedTime,
		}); err != nil {
			return
		}
	}

	// 将待办信息保存到数据库。
	{
		log.Info(ctx, "save todo information")
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		err = todoTxDo.WithContext(ctx).Select(
			todoTxDo.AppID,
			todoTxDo.ApplierID,
			todoTxDo.Type,
			todoTxDo.Candidates,
			todoTxDo.Status,
			todoTxDo.CreatedTime,
		).Create(&model.Todo{
			AppID:       app.ID,
			ApplierID:   user.ID,
			Type:        model.TodoTypeRegisterApp,
			Candidates:  systemAdmins,
			Status:      model.TodoStatusProcessing,
			CreatedTime: app.CreatedTime,
		})
		if err != nil {
			log.Error(ctx, "failed to save todo information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用成员和管理员保存到数据库。
	{
		log.Info(ctx, "save app members and administrators")
		userRoles := make([]*model.UserRole, 0, len(adminIDs)+len(memberIDs))
		for _, v := range adminIDs {
			userRoles = append(userRoles, &model.UserRole{
				AppID:  app.ID,
				UserID: v,
				Role:   model.UserRoleAppAdmin,
			})
		}
		for _, v := range memberIDs {
			userRoles = append(userRoles, &model.UserRole{
				AppID:  app.ID,
				UserID: v,
				Role:   model.UserRoleAppMember,
			})
		}
		userRoleTxDo := conn.MySQLTxClient(ctx).UserRole
		err = userRoleTxDo.WithContext(ctx).CreateInBatches(userRoles, cfg.Get().MySQL().MaximumNumberOfPerSQLInsert())
		if err != nil {
			log.Error(ctx, "failed to save app members and administrators to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件信息保存到数据库。
	{
		log.Info(ctx, "save app event information")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeRegisterApp,
			CreatedTime: app.CreatedTime,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:  app.Name,
			EventUser: user.NameEn,
			EventDetail: util.GetPrintJSON(map[string]any{
				"name":         req.Name,
				"logoFileId":   fileID,
				"logoFileName": fileName,
				"platform":     model.AllAppPlatformDescriptions[req.Platform],
				"admins":       req.Admins,
				"members":      req.Members,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppWebSearch 查询。
func AppWebSearch(ctx context.Context, req *protocol.AppWebSearchReq) (rsp *protocol.AppWebSearchRsp, err error) {
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

	// 获取用户具有权限的应用 IDs。
	var appIDs []int
	{
		log.Info(ctx, "get apps with user permissions")
		var isSystemAdmin bool
		isSystemAdmin, err = IsSystemAdmin(ctx, user.ID)
		if err != nil {
			return
		}
		if !isSystemAdmin {
			appDo := conn.MySQLClient(ctx).App
			userRoleDo := conn.MySQLClient(ctx).UserRole
			err = appDo.WithContext(ctx).
				Select(appDo.ID).
				LeftJoin(userRoleDo, userRoleDo.AppID.EqCol(appDo.ID)).
				Where(
					userRoleDo.UserID.Eq(user.ID),
					userRoleDo.Role.In(model.UserRoleAppAdmin, model.UserRoleAppMember),
				).Scan(&appIDs)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(ctx, "failed to retrieve user roles information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			if len(appIDs) <= 0 {
				return
			}
		}
	}

	// 从数据库中查询应用信息。
	var apps []*model.App
	{
		log.Info(ctx, "get apps information")
		appDo := conn.MySQLClient(ctx).App
		statement := appDo.WithContext(ctx).Select(
			appDo.Name,
			appDo.Status,
			appDo.Platform,
			appDo.AppID,
		)
		if len(appIDs) > 0 {
			statement = statement.Where(appDo.ID.In(appIDs...))
		}
		if len(req.Platform) > 0 {
			statement = statement.Where(appDo.Platform.In(req.Platform...))
		}
		if len(req.Status) > 0 {
			statement = statement.Where(appDo.Status.In(req.Status...))
		}
		orderBys := make([]field.Expr, 0, 2)
		if len(req.Name) > 0 {
			statement = statement.Where(appDo.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
			orderBys = append(orderBys, query.CaseStringOrderBy(appDo.Name, req.Name))
		}
		orderBys = append(orderBys, appDo.ID.Desc())
		apps, err = statement.Order(orderBys...).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve apps information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.AppWebSearchItem, len(apps))
		for i, v := range apps {
			list[i] = &protocol.AppWebSearchItem{
				ID:       v.AppID,
				Name:     v.Name,
				Platform: v.Platform,
				Status:   v.Status,
			}
		}
		rsp = &protocol.AppWebSearchRsp{List: list}
	}

	return
}

// AppWebUpdate 更新。
func AppWebUpdate(ctx context.Context, req *protocol.AppWebUpdateReq) (err error) {
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

	// 获取应用图标文件信息。
	var file *model.File
	{
		if req.LogoFileID != app.LogoFileID {
			log.Info(ctx, "get app logo file")
			// 从数据库中查询文件信息。
			fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.LogoFileID))
			file, err = fileDo.WithContext(ctx).Select(
				fileDo.Type,
				fileDo.UserID,
				fileDo.AppID,
				fileDo.TusdID,
			).Where(
				fileDo.FileID.Eq(req.LogoFileID),
			).Take()
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Error(ctx, "failed to retrieve file information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

		}
	}

	// 校验应用图标。
	{
		if req.LogoFileID != app.LogoFileID {
			log.Info(ctx, "verify app logo file")
			if file == nil || file.Type != model.FileTypeAppLogo || file.UserID != user.ID || file.AppID != app.ID {
				log.Warn(ctx, "app logo file is not valid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 从 Tusd 下载应用头像文件，并校验文件格式。
			var tusdResult *tus.GetResult
			tusdResult, err = conn.TusdClient(ctx).Get(ctx, &tus.GetRequest{Location: file.TusdID})
			if err != nil {
				log.Error(ctx, "failed to download app logo file from tusd", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			defer util.CloseIO(ctx, tusdResult.Body)
			appLogoMaximumSize := cfg.Get().Backend().AppLogoMaximumSize()
			if tusdResult.ContentLength > appLogoMaximumSize {
				log.Warn(ctx, "app logo file too big, max size is", appLogoMaximumSize)
				err = errs.New(consts.ErrAppLogoTooLarge)
				return
			}
			var logoFmt string
			_, logoFmt, err = image.Decode(tusdResult.Body)
			if err != nil {
				err = errs.NewWithError(consts.ErrAppLogoFormatNotSupported, err)
				return
			}
			if !slices.Contains(consts.SupportAppLogoFmt, logoFmt) {
				log.Warn(ctx, "app logo file format is invalid", logoFmt, "support", consts.SupportAppLogoFmt)
				err = errs.New(consts.ErrAppLogoFormatNotSupported)
				return
			}

			// 更新成功后，删除老图标文件。
			defer func() {
				if err == nil {
					log.Info(ctx, "reclaim old app logo file")
					// 获取应用老图标文件信息。
					fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(app.LogoFileID))
					file2, err2 := fileDo.WithContext(ctx).Select(
						fileDo.TusdID,
					).Where(
						fileDo.FileID.Eq(app.LogoFileID),
					).Take()
					if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
						log.Error(ctx, "failed to retrieve app logo file information", err2)
						return
					}
					if file2 == nil {
						log.Warn(ctx, "app logo file does not exist", app.LogoFileID)
						return
					}

					// 删除应用老图标文件信息。
					fileDo = conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(app.LogoFileID))
					if _, err2 = fileDo.WithContext(ctx).Where(
						fileDo.FileID.Eq(app.LogoFileID),
					).Delete(); err2 != nil {
						log.Error(ctx, "failed to delete file information in database", err2)
						return
					}

					// 删除 Tusd 中的应用老图标文件。
					log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, file2.TusdID),
						"failed to delete app logo file in tusd", file2.TusdID)

					// 回收文件 ID。
					log.ErrorIf(ctx, reclaimID(ctx, IDFile, app.LogoFileID),
						"failed to reclaim app logo file id", app.LogoFileID)
				}
			}()
		}
	}

	// 整理和校验应用成员和管理员。
	var adminIDs []int
	var memberIDs []int
	{
		log.Info(ctx, "tidy up members and administrators of app")
		userNames := util.CleanStrings(append(req.Admins, req.Members...))
		adminIDs = make([]int, 0, len(req.Admins))
		memberIDs = make([]int, 0, len(req.Members))
		userDo := conn.MySQLClient(ctx).User
		var users []*model.User
		users, err = userDo.WithContext(ctx).Select(
			userDo.ID,
			userDo.NameEn,
		).Where(
			userDo.NameEn.In(userNames...),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve users information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(users) != len(userNames) {
			log.Warn(ctx, "usernames length is invalid", userNames)
			err = errs.New(consts.ErrUserNotExists)
			return
		}
		for _, v := range users {
			if slices.Contains(req.Admins, v.NameEn) {
				adminIDs = append(adminIDs, v.ID)
			}
			if slices.Contains(req.Members, v.NameEn) {
				memberIDs = append(memberIDs, v.ID)
			}
		}
	}

	// 更新数据库中的应用信息。
	var now time.Time
	{
		log.Info(ctx, "update app information")
		now = time.Now()
		appTxDo := conn.MySQLTxClient(ctx).App
		_, err = appTxDo.WithContext(ctx).Where(
			appTxDo.ID.Eq(app.ID),
		).UpdateColumnSimple(
			appTxDo.LogoFileID.Value(req.LogoFileID),
			appTxDo.Name.Value(req.Name),
			appTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update app information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 更新数据库中的应用成员权限信息。先全部删除，再添加。有签名权限的成员，保留签名权限。移除了的成员，删除签名权限。
	{
		log.Info(ctx, "update app user role information")
		userRoleTxDo := conn.MySQLTxClient(ctx).UserRole
		var hasAppSignUsers []int
		err = userRoleTxDo.WithContext(ctx).Select(
			userRoleTxDo.UserID,
		).Where(
			userRoleTxDo.AppID.Eq(app.ID),
			userRoleTxDo.Role.Eq(model.UserRoleAppSigner),
		).Scan(&hasAppSignUsers)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve user roles information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		canDeleteAppSignUsers := make([]int, 0, len(hasAppSignUsers))
		for _, v := range hasAppSignUsers {
			if !slices.Contains(adminIDs, v) && !slices.Contains(memberIDs, v) {
				canDeleteAppSignUsers = append(canDeleteAppSignUsers, v)
			}
		}
		if len(canDeleteAppSignUsers) > 0 {
			log.Info(ctx, "delete app sign user roles in database")
			if _, err = userRoleTxDo.WithContext(ctx).Where(
				userRoleTxDo.AppID.Eq(app.ID),
				userRoleTxDo.UserID.In(canDeleteAppSignUsers...),
				userRoleTxDo.Role.Eq(model.UserRoleAppSigner),
			).Delete(); err != nil {
				log.Error(ctx, "failed to delete app sign user roles information in database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
		if _, err = userRoleTxDo.WithContext(ctx).Where(
			userRoleTxDo.AppID.Eq(app.ID),
			userRoleTxDo.Role.In(model.UserRoleAppMember, model.UserRoleAppAdmin),
		).Delete(); err != nil {
			log.Error(ctx, "failed to delete app user role information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userRoleInfos := make([]*model.UserRole, 0, len(adminIDs)+len(memberIDs))
		for _, v := range adminIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{
				AppID:  app.ID,
				UserID: v,
				Role:   model.UserRoleAppAdmin,
			})
		}
		for _, v := range memberIDs {
			userRoleInfos = append(userRoleInfos, &model.UserRole{
				AppID:  app.ID,
				UserID: v,
				Role:   model.UserRoleAppMember,
			})
		}
		err = userRoleTxDo.WithContext(ctx).
			CreateInBatches(userRoleInfos, cfg.Get().MySQL().MaximumNumberOfPerSQLInsert())
		if err != nil {
			log.Error(ctx, "failed to add app user roles information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 更新数据库中的应用待办信息的可审批人员。
	{
		log.Info(ctx, "update approvers of app todo information")
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		_, err = todoTxDo.WithContext(ctx).Where(
			todoTxDo.AppID.Eq(app.ID),
			todoTxDo.Type.In(model.TodoTypeApplySigner, model.TodoTypeJoinApp),
			todoTxDo.Status.In(model.TodoStatusProcessing),
		).UpdateColumnSimple(
			todoTxDo.Candidates.Value(model.IntList(adminIDs)),
		)
		if err != nil {
			log.Error(ctx, "failed to update approvers of app todo information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件保存在数据库。
	{
		log.Info(ctx, "create app event")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeUpdateApp,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:  app.Name,
			EventUser: user.NameEn,
			EventDetail: util.GetPrintJSON(map[string]any{
				"name":       req.Name,
				"logoFileId": req.LogoFileID,
				"admins":     req.Admins,
				"members":    req.Members,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppWebGetInformation 获取应用信息。
func AppWebGetInformation(ctx context.Context) (rsp *protocol.AppWebGetInformationRsp, err error) {
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

	// 从数据库中获取应用成员和管理员。
	var userRoles []*model.UserRole
	var userIDToInfo map[int]*model.User
	{
		log.Info(ctx, "get app users information")
		userRoleDo := conn.MySQLClient(ctx).UserRole
		userRoles, err = userRoleDo.WithContext(ctx).Select(
			userRoleDo.UserID,
			userRoleDo.Role,
		).Where(
			userRoleDo.AppID.Eq(app.ID),
			userRoleDo.Role.In(model.UserRoleAppAdmin, model.UserRoleAppMember),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve app users information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userIDs := util.ListTo(userRoles, func(e *model.UserRole) int { return e.UserID })
		userDo := conn.MySQLClient(ctx).User
		var users []*model.User
		if len(userIDs) > 0 {
			users, err = userDo.WithContext(ctx).Select(
				userDo.ID,
				userDo.NameEn,
				userDo.NameZh,
			).Where(
				userDo.ID.In(userIDs...),
			).Find()
			if err != nil {
				log.Error(ctx, "failed to retrieve users information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
		userIDToInfo = util.ListAssociateBy(users, func(e *model.User) int { return e.ID })
	}

	// 整理数据。
	{
		rsp = &protocol.AppWebGetInformationRsp{
			AppID:      app.AppID,
			Name:       app.Name,
			LogoFileID: app.LogoFileID,
			Platform:   app.Platform,
			Status:     app.Status,
			Admins:     make(map[string]string, len(userIDToInfo)),
			Members:    make(map[string]string, len(userIDToInfo)),
		}
		for _, v := range userRoles {
			info := userIDToInfo[v.UserID]
			if info == nil {
				continue
			}
			switch v.Role {
			case model.UserRoleAppAdmin:
				rsp.Admins[info.NameEn] = info.NameZh
			case model.UserRoleAppMember:
				rsp.Members[info.NameEn] = info.NameZh
			default:
				log.Error(ctx, "unknown user role", v.Role)
			}
		}
	}

	return
}

// AppWebInvalidate 无效化。
func AppWebInvalidate(ctx context.Context) (err error) {
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

	// 更新数据库中的应用状态。
	var now time.Time
	{
		log.Info(ctx, "update app information")
		now = time.Now()
		appTxDo := conn.MySQLTxClient(ctx).App
		var sqlResult gen.ResultInfo
		sqlResult, err = appTxDo.WithContext(ctx).Where(
			appTxDo.ID.Eq(app.ID),
			appTxDo.Status.Eq(model.AppStatusValid),
		).UpdateColumnSimple(
			appTxDo.Status.Value(model.AppStatusInvalid),
			appTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update status of app in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "updating app information no rows effected")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	// 将应用事件信息保存到数据库。
	{
		log.Info(ctx, "save app event to database")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeInvalidateApp,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:  app.Name,
			EventUser: user.NameEn,
			EventDetail: util.GetPrintJSON(map[string]any{
				"appId": app.AppID,
			}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppWebEnable 启用。
func AppWebEnable(ctx context.Context) (err error) {
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
		if app.Status != model.AppStatusInvalid {
			err = errs.New(consts.ErrAppStatusNotInvalid)
			return
		}
	}

	// 查询系统管理员。
	var systemAdmins []int
	{
		log.Info(ctx, "get system administrators")
		userRoleDo := conn.MySQLClient(ctx).UserRole
		err = userRoleDo.WithContext(ctx).Select(
			userRoleDo.UserID,
		).Where(
			userRoleDo.Role.Eq(model.UserRoleSystemAdmin),
		).Scan(&systemAdmins)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve system administrators from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(systemAdmins) <= 0 {
			log.Error(ctx, "no system administrators were found")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 检查应用不存在相同类型正在处理中的待办。
	{
		log.Info(ctx, "check there are no pending todo of enabling app")
		var todo *model.Todo
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		todo, err = todoTxDo.WithContext(ctx).Where(
			todoTxDo.AppID.Eq(app.ID),
			todoTxDo.Type.Eq(model.TodoTypeActivateApp),
			todoTxDo.Status.Eq(model.TodoStatusProcessing),
		).Clauses(query.ForUpdate()).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve todo information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if todo != nil {
			err = errs.New(consts.ErrSameTodoExist)
			return
		}
	}

	// 更新应用状态为审批中。
	var now time.Time
	{
		log.Info(ctx, "update app status")
		now = time.Now()
		appTxDo := conn.MySQLTxClient(ctx).App
		var sqlResult gen.ResultInfo
		sqlResult, err = appTxDo.WithContext(ctx).Where(
			appTxDo.ID.Eq(app.ID),
			appTxDo.Status.Eq(app.Status),
		).UpdateColumnSimple(
			appTxDo.Status.Value(model.AppStatusApproving),
			appTxDo.UpdatedTime.Value(now),
		)
		if err != nil {
			log.Error(ctx, "failed to update app status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "updating app status no rows effected")
			err = errs.New(consts.ErrCommonFailure)
			return
		}
	}

	// 将待办信息保存到数据库。
	{
		log.Info(ctx, "create todo")
		todoTxDo := conn.MySQLTxClient(ctx).Todo
		err = todoTxDo.WithContext(ctx).Select(
			todoTxDo.AppID,
			todoTxDo.ApplierID,
			todoTxDo.Type,
			todoTxDo.Candidates,
			todoTxDo.Status,
			todoTxDo.CreatedTime,
		).Create(&model.Todo{
			AppID:       app.ID,
			ApplierID:   user.ID,
			Type:        model.TodoTypeActivateApp,
			Candidates:  systemAdmins,
			Status:      model.TodoStatusProcessing,
			CreatedTime: now,
		})
		if err != nil {
			log.Error(ctx, "failed to create todo information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 将应用事件信息保存到数据库。
	{
		log.Info(ctx, "save app event to database")
		err = createEvent(ctx, &model.Event{
			AppID:       app.ID,
			UserID:      user.ID,
			Type:        model.EventTypeEnableApp,
			CreatedTime: now,
			Source:      model.SourceWeb,
		}, map[EventField]any{
			EventApp:    app.Name,
			EventUser:   user.NameEn,
			EventDetail: util.GetPrintJSON(map[string]any{"appId": app.AppID}),
		})
		if err != nil {
			return
		}
	}

	return
}

// AppWebCount 获取用户具有权限的应用个数。
func AppWebCount(ctx context.Context) (rsp *protocol.AppWebCountRsp, err error) {
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

	// 从数据库中查询用户有权限的应用数量。
	var count int
	{
		log.Info(ctx, "retrieve the number of user apps from database")
		userRoleDo := conn.MySQLClient(ctx).UserRole
		err = userRoleDo.WithContext(ctx).Select(
			userRoleDo.AppID.Distinct().Count(),
		).Where(
			userRoleDo.UserID.Eq(user.ID),
			userRoleDo.AppID.IsNotNull(),
		).Scan(&count)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve the number of user apps from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	rsp = &protocol.AppWebCountRsp{Count: count}

	return
}

// GetAppInfoByID 获取应用信息。
func GetAppInfoByID(ctx context.Context, appID string) (*model.App, error) {
	if len(appID) <= 0 {
		return nil, nil
	}
	appDo := conn.MySQLClient(ctx).App
	app, err := appDo.WithContext(ctx).Where(appDo.AppID.Eq(appID)).Take()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error(ctx, "failed to retrieve app information from database", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return app, nil
}
