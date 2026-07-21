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
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
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

// SessionInfo 用户登录信息。
type SessionInfo struct {
	// 用户 ID
	UserID int `json:"userId,omitempty"`
	// 用户英文名
	User string `json:"user,omitempty"`
	// 会话凭证
	Session string `json:"session,omitempty"`
	// 登录 IP
	IP string `json:"ip,omitempty"`
}

// UserWebRegister 注册。
func UserWebRegister(ctx context.Context, req *protocol.UserWebRegisterReq) (err error) {
	// 获取户头像。
	var avatarBytes []byte
	{
		log.Info(ctx, "get user avatar")
		var avatarStream multipart.File
		avatarStream, err = req.Avatar.Open()
		if err != nil {
			log.Error(ctx, "opening user avatar file failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer util.CloseIO(ctx, avatarStream)
		avatarBytes, err = io.ReadAll(avatarStream)
		if err != nil {
			log.Error(ctx, "reading user avatar file failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验用户头像。
	{
		log.Info(ctx, "verify user avatar")
		avatarMaximumSize := cfg.Get().Backend().UserAvatarMaximumSize()
		if len(avatarBytes) > avatarMaximumSize {
			err = errs.New(consts.ErrUserAvatarTooLarge)
			return
		}
		var avatarFmt string
		_, avatarFmt, err = image.Decode(bytes.NewReader(avatarBytes))
		if err != nil {
			err = errs.NewWithError(consts.ErrUserAvatarFormatNotSupported, err)
			return
		}
		if !slices.Contains(consts.SupportUserAvatarFmt, avatarFmt) {
			log.Warn(ctx, "invalid user avatar format", avatarFmt, ". support", consts.SupportUserAvatarFmt)
			err = errs.New(consts.ErrUserAvatarFormatNotSupported)
			return
		}
	}

	// 从数据库中查询英文名有没有被占用。
	{
		log.Info(ctx, "check whether the user name is occupied in the database")
		userTxDo := conn.MySQLTxClient(ctx).User
		var count int64
		count, err = userTxDo.WithContext(ctx).Where(
			userTxDo.NameEn.Eq(req.NameEn),
		).Clauses(query.ForUpdate()).Count()
		if err != nil {
			log.Error(ctx, "failed to retrieve user information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if count > 0 {
			err = errs.New(consts.ErrUserNameDuplicate)
			return
		}
	}

	// 将头像文件保存至 Tusd。
	var tusdID string
	{
		log.Info(ctx, "upload user avatar to tusd")
		tusdID, err = conn.TusdClient(ctx).UploadFile(ctx, avatarBytes)
		if err != nil {
			log.Error(ctx, "failed to upload user avatar file to tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		// 注册失败就删除文件。
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, tusdID), "deleting user avatar file failed",
					tusdID)
			}
		}()
	}

	// 将用户信息保存到数据库。
	var fileID string
	var now time.Time
	var userID int
	{
		log.Info(ctx, "save user information to database")
		fileID, err = generateID(ctx, IDFile)
		if err != nil {
			return
		}
		defer func() {
			// 注册失败就回收 ID。
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "reclaiming user avatar file id failed")
			}
		}()
		now = time.Now()
		salt := util.RandomPrintableASCIIString(consts.UserPasswordSaltLength)
		digest := md5.Sum([]byte(salt + req.Password))
		password := hex.EncodeToString(digest[:])
		user := &model.User{
			NameEn:         req.NameEn,
			NameZh:         req.NameZh,
			AvatarFileID:   fileID,
			Department:     req.Department,
			PasswordDigest: password,
			PasswordSalt:   salt,
			CreatedTime:    now,
		}
		userTxDo := conn.MySQLTxClient(ctx).User
		if err = userTxDo.WithContext(ctx).Select(
			userTxDo.NameEn,
			userTxDo.NameZh,
			userTxDo.AvatarFileID,
			userTxDo.Department,
			userTxDo.PasswordDigest,
			userTxDo.PasswordSalt,
			userTxDo.CreatedTime,
		).Create(user); err != nil {
			if errs.IsMySQLError(err, mysql.ErrDupEntry) {
				err = errs.New(consts.ErrUserNameDuplicate)
				return
			}
			log.Error(ctx, "failed to save user information to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userID = user.ID
	}

	// 将文件信息保存到数据库。
	{
		log.Info(ctx, "save user avatar file information to database")
		digest := md5.Sum(avatarBytes)
		if err = createFile(ctx, &model.File{
			FileID:      fileID,
			TusdID:      tusdID,
			UserID:      userID,
			Name:        filepath.Base(req.Avatar.Filename),
			Md5:         hex.EncodeToString(digest[:]),
			Size:        int(req.Avatar.Size),
			Type:        model.FileTypeUserAvatar,
			CreatedTime: now,
		}); err != nil {
			return
		}
	}

	return
}

// UserWebLogin 登录。
func UserWebLogin(ctx context.Context, req *protocol.UserWebLoginReq) (session string, err error) {
	// 从数据库中查询用户信息。
	var user *model.User
	{
		log.Info(ctx, "get user information")
		userDo := conn.MySQLClient(ctx).User
		user, err = userDo.WithContext(ctx).Select(
			userDo.ID,
			userDo.NameEn,
			userDo.PasswordDigest,
			userDo.PasswordSalt,
		).Where(
			userDo.NameEn.Eq(req.NameEn),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn(ctx, "user not found", req.NameEn)
				err = errs.New(consts.ErrUserLoginFailed)
				return
			}
			log.Error(ctx, "failed to retrieve user information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 校验登陆密码。
	{
		log.Info(ctx, "verify user password")
		if len(user.PasswordSalt) > 0 {
			delta := consts.UserPasswordSaltLength - len(user.PasswordSalt)
			if delta > 0 {
				user.PasswordSalt += strings.Repeat(" ", delta)
			}
		}
		digest := md5.Sum([]byte(user.PasswordSalt + req.Password))
		password := hex.EncodeToString(digest[:])
		if password != user.PasswordDigest {
			log.Warn(ctx, "user password incorrect")
			err = errs.New(consts.ErrUserLoginFailed)
			return
		}
	}

	// 将用户会话信息保存到缓存中。
	{
		session = util.RandomPrintableASCIIString(consts.UserSessionLength)
		log.Info(ctx, "cache user login information")
		sessionInfo := SessionInfo{
			UserID:  user.ID,
			User:    user.NameEn,
			Session: session,
			IP:      ctxs.RequestIP(ctx),
		}
		err = conn.RedisClient(ctx).Set(ctx, fmt.Sprintf(consts.RedisKeySessionFmt, user.NameEn),
			util.GetPrintJSON(sessionInfo), cfg.Get().Backend().WebSessionExpiration()).Err()
		if err != nil {
			log.Error(ctx, "failed to set user information to redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// UserWebGetInformation 获取用户信息。
func UserWebGetInformation(ctx context.Context) (rsp *protocol.UserWebGetInformationRsp, err error) {
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

	rsp = &protocol.UserWebGetInformationRsp{
		AvatarFileID: userInfo.AvatarFileID,
		NameZh:       userInfo.NameZh,
		NameEn:       userInfo.NameEn,
		Department:   userInfo.Department,
	}

	return
}

// UserWebUpdate 更新个人信息。
func UserWebUpdate(ctx context.Context, req *protocol.UserWebUpdateReq) (err error) {
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

	// 从数据库中获取新的头像文件信息。
	var file *model.File
	{
		if user.AvatarFileID != req.AvatarFileID {
			log.Info(ctx, "get file information")
			fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.AvatarFileID))
			file, err = fileDo.WithContext(ctx).Select(
				fileDo.Type,
				fileDo.UserID,
				fileDo.TusdID,
			).Where(
				fileDo.FileID.Eq(req.AvatarFileID),
			).Take()
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				log.Error(ctx, "failed to retrieve file information from database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	// 校验用户头像。
	{
		if user.AvatarFileID != req.AvatarFileID {
			log.Info(ctx, "verify user avatar file")
			if file == nil || file.Type != model.FileTypeUserAvatar || file.UserID != user.ID {
				log.Warn(ctx, "user avatar file is not valid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 下载新头像，校验文件格式。
			var tusdResult *tus.GetResult
			tusdResult, err = conn.TusdClient(ctx).Get(ctx, &tus.GetRequest{Location: file.TusdID})
			if err != nil {
				log.Error(ctx, "failed to download user avatar file from tusd", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
			defer util.CloseIO(ctx, tusdResult.Body)
			avatarMaximumSize := cfg.Get().Backend().UserAvatarMaximumSize()
			if tusdResult.ContentLength > avatarMaximumSize {
				log.Warn(ctx, "user avatar file too big, max size is", avatarMaximumSize)
				err = errs.New(consts.ErrUserAvatarTooLarge)
				return
			}
			var avatarFmt string
			_, avatarFmt, err = image.Decode(tusdResult.Body)
			if err != nil {
				err = errs.NewWithError(consts.ErrUserAvatarFormatNotSupported, err)
				return
			}
			if !slices.Contains(consts.SupportUserAvatarFmt, avatarFmt) {
				log.Info(ctx, "invalid user avatar format", avatarFmt, ". support", consts.SupportUserAvatarFmt)
				err = errs.New(consts.ErrUserAvatarFormatNotSupported)
				return
			}

			// 若更新成功，删除老的用户头像文件。
			defer func() {
				if err == nil {
					log.Info(ctx, "reclaim old user avatar file")

					// 从数据库中获取老头像文件信息。
					fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(user.AvatarFileID))
					file2, err2 := fileDo.WithContext(ctx).Select(
						fileDo.TusdID,
					).Where(
						fileDo.FileID.Eq(user.AvatarFileID),
					).Take()
					if err2 != nil && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
						log.Error(ctx, "failed to retrieve old user avatar file information from database", err2)
						return
					}
					if file2 == nil {
						log.Warn(ctx, "user avatar file does not exist", user.AvatarFileID)
						return
					}

					// 删除数据库中的用户老头像文件信息。
					fileDo = conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(user.AvatarFileID))
					if _, err2 = fileDo.WithContext(ctx).Where(
						fileDo.FileID.Eq(user.AvatarFileID),
					).Delete(); err2 != nil {
						log.Error(ctx, "failed to delete file information in database", err2, user.AvatarFileID)
						return
					}

					// 删除 Tusd 中的用户老头像文件。
					log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, file2.TusdID),
						"failed to delete user avatar file in tusd", file2.TusdID)

					// 回收文件 ID。
					log.ErrorIf(ctx, reclaimID(ctx, IDFile, user.AvatarFileID), "failed to reclaim user file id",
						user.AvatarFileID)
				}
			}()
		}
	}

	// 更新数据库中的用户信息。
	{
		log.Info(ctx, "update user information")
		passwordSalt := util.RandomPrintableASCIIString(consts.UserPasswordSaltLength)
		digest := md5.Sum([]byte(passwordSalt + req.Password))
		userTxDo := conn.MySQLTxClient(ctx).User
		_, err = userTxDo.WithContext(ctx).Where(
			userTxDo.ID.Eq(user.ID),
		).UpdateColumnSimple(
			userTxDo.AvatarFileID.Value(req.AvatarFileID),
			userTxDo.NameZh.Value(req.NameZh),
			userTxDo.Department.Value(req.Department),
			userTxDo.PasswordSalt.Value(passwordSalt),
			userTxDo.PasswordDigest.Value(hex.EncodeToString(digest[:])),
			userTxDo.UpdatedTime.Value(time.Now()),
		)
		if err != nil {
			log.Error(ctx, "failed to update user information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// UserWebSearch 搜索用户。
func UserWebSearch(ctx context.Context, req *protocol.UserWebSearchReq) (rsp *protocol.UserWebSearchRsp, err error) {
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

	// 从数据库中获取用户信息。
	var userNameToName map[string]string
	{
		log.Info(ctx, "search users")
		userDo := conn.MySQLClient(ctx).User
		var users []*model.User
		users, err = userDo.WithContext(ctx).SearchByName(req.Name)
		if err != nil {
			log.Error(ctx, "failed to retrieve users information in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		userNameToName = util.ListToMap(users, func(e *model.User) (string, string) { return e.NameEn, e.NameZh })
		rsp = &protocol.UserWebSearchRsp{Users: userNameToName}
	}

	return
}

// UserWebLogout 登出。
func UserWebLogout(ctx context.Context) (err error) {
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

	// 删除登录缓存信息。
	{
		log.Info(ctx, "delete user login information in redis")
		err = conn.RedisClient(ctx).Del(ctx, fmt.Sprintf(consts.RedisKeySessionFmt, user.NameEn)).Err()
		if err != nil {
			log.Error(ctx, "failed to delete redis key", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// IsSystemAdmin 是否是系统管理员。
func IsSystemAdmin(ctx context.Context, user int) (bool, error) {
	userRoleDo := conn.MySQLClient(ctx).UserRole
	count, err := userRoleDo.WithContext(ctx).Where(
		userRoleDo.UserID.Eq(user),
		userRoleDo.Role.Eq(model.UserRoleSystemAdmin),
		userRoleDo.AppID.IsNull(),
	).Count()
	if err != nil {
		log.Error(ctx, "failed to retrieve user roles from database", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	return count > 0, nil
}

// UserHasAnyRight 查库获取用户是否有权限。
func UserHasAnyRight(ctx context.Context, userID, appID int, perm int, perms ...int) (bool, error) {
	// 系统管理员特殊处理。
	perms2 := make([]int, 0, len(perms)+1)
	isSystemAdmin := false
	for _, v := range append(perms, perm) {
		if v == model.UserRoleSystemAdmin {
			isSystemAdmin = true
		} else {
			perms2 = append(perms2, v)
		}
	}
	perms2 = util.CleanNumbers(perms2)

	userRoleDo := conn.MySQLClient(ctx).UserRole
	statement := userRoleDo.WithContext(ctx)
	if len(perms2) > 0 {
		statement = statement.Where(
			userRoleDo.UserID.Eq(userID),
			userRoleDo.AppID.Eq(appID),
			userRoleDo.Role.In(perms2...),
		)
	}
	if isSystemAdmin {
		statement = statement.Or(
			userRoleDo.UserID.Eq(userID),
			userRoleDo.AppID.IsNull(),
			userRoleDo.Role.Eq(model.UserRoleSystemAdmin),
		)
	}
	count, err := statement.Count()
	if err != nil {
		log.Error(ctx, "failed to retrieve user roles from database", err)
		return false, errs.NewWithError(consts.ErrSystem, err)
	}
	return count > 0, nil
}

// GetUserNamesByIDs 获取用户名。
func GetUserNamesByIDs(ctx context.Context, userIDs []int) (map[int]string, error) {
	if len(userIDs) <= 0 {
		return make(map[int]string), nil
	}
	userDo := conn.MySQLClient(ctx).User
	users, err := userDo.WithContext(ctx).Select(
		userDo.ID,
		userDo.NameEn,
	).Where(
		userDo.ID.In(userIDs...),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve user names from database", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return util.ListToMap(users, func(e *model.User) (int, string) { return e.ID, e.NameEn }), nil
}
