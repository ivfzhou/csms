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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"github.com/redis/go-redis/v9"
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
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

// 文件上传信息。
type uploadingFileInfo struct {
	// 文件名
	Name string `json:"name,omitempty"`
	// 文件大小
	Size int64 `json:"size,omitempty"`
	// 上传用户 ID
	User int `json:"user,omitempty"`
	// 凭证 ID
	AccountID int `json:"accountId,omitempty"`
	// 应用 ID
	AppID int `json:"appId,omitempty"`
	// 文件 MD5 值
	MD5 string `json:"md5,omitempty"`
	// 上传类型
	Type int `json:"type,omitempty"`
	// 上传时间
	TimeSecond int64 `json:"timeSecond,omitempty"`
}

// FileWebDownload 下载。
func FileWebDownload(ctx context.Context, req *protocol.FileWebDownloadReq) (fileObj *FileInfo, err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown request context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 从数据库中获取文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.AppID,
			fileDo.TusdID,
			fileDo.Name,
			fileDo.Size,
			fileDo.Type,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.New(consts.ErrFileNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 根据文件类型校验下载权限。
	{
		log.Info(ctx, "verify file download request")
		var hasRight bool
		switch file.Type {
		case model.FileTypeUserAvatar: // 允许下载。
		case model.FileTypeAndroidSigning, model.FileTypeWindowsSigning, model.FileTypeAppleSigning,
			model.FileTypeMicrosoftSigning, model.FileTypeAppLogo, model.FileTypeHLKLog:
			hasRight, err = UserHasAnyRight(ctx, user.ID, file.AppID,
				model.UserRoleAppAdmin,
				model.UserRoleAppMember,
				model.UserRoleSystemAdmin,
			)
			if err != nil {
				return
			}
			if !hasRight {
				log.Warn(ctx, "no permission to download file")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
		}
	}

	// 从 Tusd 下载文件。
	var tusdResult *tus.GetResult
	{
		log.Info(ctx, "download file from tusd")
		tusdResult, err = conn.TusdClient(ctx).Get(ctx, &tus.GetRequest{Location: file.TusdID})
		if err != nil {
			log.Error(ctx, "failed to download file from tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if tusdResult.ContentLength != file.Size {
			log.Warn(ctx, "file size is different", file.Size, tusdResult.ContentLength)
		}
	}

	fileObj = &FileInfo{Name: file.Name, Size: int64(tusdResult.ContentLength), Reader: tusdResult.Body}

	return
}

// FileWebInitial 初始化上传。
func FileWebInitial(ctx context.Context, req *protocol.FileWebInitialReq) (rsp *protocol.FileWebInitialRsp, err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown request context")
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取应用信息。
	var app *model.App
	{
		if len(req.AppID) > 0 {
			log.Info(ctx, "get app information")
			app, err = GetAppInfoByID(ctx, req.AppID)
			if err != nil {
				return
			}
		}
	}

	// 按文件类型校验文件。
	{
		log.Info(ctx, "verify uploading file request")
		switch req.Type {
		case model.FileTypeUserAvatar:
			// 校验文件格式。
			fileExt := strings.ToLower(strings.Trim(filepath.Ext(req.Name), "."))
			if !slices.Contains(consts.SupportUserAvatarFmt, fileExt) {
				err = errs.New(consts.ErrUserAvatarFormatNotSupported)
				return
			}

			// 校验文件大小。
			userAvatarMaximumSize := cfg.Get().Backend().UserAvatarMaximumSize()
			if req.Size > int64(userAvatarMaximumSize) {
				log.Warn(ctx, "user avatar file size too big, max size is", userAvatarMaximumSize)
				err = errs.New(consts.ErrUserAvatarTooLarge)
				return
			}
		case model.FileTypeAppLogo:
			// 校验应用状态。
			if app == nil || app.ID <= 0 || app.Status != model.AppStatusValid {
				log.Warn(ctx, "app state is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 校验文件格式。
			fileExt := strings.ToLower(strings.Trim(filepath.Ext(req.Name), "."))
			if !slices.Contains(consts.SupportAppLogoFmt, fileExt) {
				err = errs.New(consts.ErrAppLogoFormatNotSupported)
				return
			}

			// 校验文件大小。
			appLogoMaximumSize := cfg.Get().Backend().AppLogoMaximumSize()
			if req.Size > int64(appLogoMaximumSize) {
				log.Warn(ctx, "app logo file size too big, max size is", appLogoMaximumSize)
				err = errs.New(consts.ErrAppLogoTooLarge)
				return
			}
		case model.FileTypeAndroidSigning, model.FileTypeAppleSigning, model.FileTypeWindowsSigning:
			// 校验应用状态。
			if app == nil || app.ID <= 0 || app.Status != model.AppStatusValid {
				log.Warn(ctx, "app state is invalid")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}

			// 用户要具有相关权限。
			var hasRight bool
			hasRight, err = UserHasAnyRight(ctx, user.ID, app.ID, model.UserRoleAppAdmin, model.UserRoleAppMember)
			if err != nil {
				return
			}
			if !hasRight {
				err = errs.New(consts.ErrPermissionDenied)
				return
			}
		case model.FileTypeHLKLog, model.FileTypeMicrosoftSigning:
			// 不允许上传。
			log.Warn(ctx, "cannot upload this type file")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 检查数据库中是否存在相同的文件。
	var now time.Time
	{
		log.Info(ctx, "check if there is identical file in database")
		now = time.Now()
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableName(now))
		conditions := make([]gen.Condition, 0, 6)
		conditions = append(conditions,
			fileDo.Md5.Eq(strings.ToLower(req.MD5)),
			fileDo.Name.Eq(filepath.Base(req.Name)),
			fileDo.Size.Eq(int(req.Size)),
			fileDo.Type.Eq(req.Type),
			fileDo.UserID.Eq(user.ID),
		)
		if app != nil {
			conditions = append(conditions, fileDo.AppID.Eq(app.ID))
		}
		var file *model.File
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.FileID,
		).Where(conditions...).Order(fileDo.ID.Desc()).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if file != nil {
			rsp = &protocol.FileWebInitialRsp{FileID: file.FileID, Exist: true}
			return
		}
	}

	// 缓存文件上传信息。
	var fileID string
	{
		log.Info(ctx, "cache file upload information")
		log.Info(ctx, "generate file id")
		fileID, err = generateID(ctx, IDFile)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "reclaim file id failed", fileID)
			}
		}()
		cachedFileInfo := &uploadingFileInfo{
			Name:       req.Name,
			Size:       req.Size,
			MD5:        req.MD5,
			User:       user.ID,
			Type:       req.Type,
			TimeSecond: now.Unix(),
		}
		if app != nil {
			cachedFileInfo.AppID = app.ID
		}
		fileInfoBytes, _ := json.Marshal(cachedFileInfo)
		err = conn.RedisClient(ctx).HSet(ctx, consts.RedisKeyFileUploadInfo, fileID, string(fileInfoBytes)).Err()
		if err != nil {
			log.Error(ctx, "failed to cache file upload information to redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	rsp = &protocol.FileWebInitialRsp{FileID: fileID}

	return
}

// FileWebUploadPart 上传分片。
func FileWebUploadPart(ctx context.Context, req *protocol.FileWebUploadPartReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown request context")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 获取文件上传缓存记录。
	var cachedFileInfo *uploadingFileInfo
	{
		log.Info(ctx, "get file cache information")
		var fileCache string
		fileCache, err = conn.RedisClient(ctx).HGet(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				log.Warn(ctx, "uploading file info is not found")
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to get file cache information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		cachedFileInfo = &uploadingFileInfo{}
		if err = json.Unmarshal([]byte(fileCache), cachedFileInfo); err != nil {
			log.Warn(ctx, "unserializing file cache information failed", err, fileCache)
		}
	}

	// 确保是同一个人上传。
	{
		log.Info(ctx, "validate file upload request")
		if cachedFileInfo.User != user.ID {
			log.Warn(ctx, "user who uploading file is incorrect", user.NameEn)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 上传时长不能太长。
	{
		maximumInterval := cfg.Get().Backend().FileUploadingMaximumInterval()
		if time.Since(time.Unix(cachedFileInfo.TimeSecond, 0)) > maximumInterval {
			log.Warn(ctx, "uploading file is too large for exceeded", maximumInterval, cachedFileInfo.TimeSecond)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 加锁，避免同序号分片覆盖。
	{
		log.Info(ctx, "lock uploading file part")
		lockKey := fmt.Sprintf(consts.RedisKeyFileUploadPartLockFmt, req.FileID, req.ChunkNumber)
		var success bool
		success, err = conn.RedisLock(ctx, lockKey, 0, time.Hour)
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer func() {
			_, err2 := conn.RedisUnlock(ctx, lockKey)
			log.ErrorIf(ctx, err2, "deleting redis key failed")
		}()
		if !success {
			log.Warn(ctx, "uploading duplicate file chunk")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 校验，确保该序号分片还未上传。
	{
		log.Info(ctx, "check file chunk number")
		var chunkNumbers []string
		chunkNumbers, err = conn.RedisClient(ctx).ZRangeByScore(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			&redis.ZRangeBy{Min: strconv.Itoa(req.ChunkNumber), Max: strconv.Itoa(req.ChunkNumber)},
		).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Error(ctx, "failed to get file chunk information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(chunkNumbers) > 0 {
			log.Warn(ctx, "file part chunk exists", chunkNumbers)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 上传到 Tusd。
	var tusdID string
	{
		log.Info(ctx, "upload file part to tusd")
		var fileReader multipart.File
		fileReader, err = req.Chunk.Open()
		if err != nil {
			log.Error(ctx, "failed to open file part", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		tusdID, err = conn.TusdClient(ctx).UploadPartByIO(ctx, fileReader, int(req.Chunk.Size))
		if err != nil {
			log.Error(ctx, "failed to upload file part to tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer func() {
			// 失败则丢弃分片。
			if err != nil {
				log.ErrorIf(ctx, conn.TusdClient(ctx).DiscardParts(ctx, []string{tusdID}),
					"deleting file part failed", tusdID)
			}
		}()
	}

	// 记录分片信息。
	{
		log.Info(ctx, "cache file chunk information")
		err = conn.RedisClient(ctx).ZAdd(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			redis.Z{Score: float64(req.ChunkNumber), Member: toFileChunkMember(tusdID, req.Chunk.Size)},
		).Err()
		if err != nil {
			log.Error(ctx, "failed to cache file chunk information to redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// FileWebMergeParts 合并分片。
func FileWebMergeParts(ctx context.Context, req *protocol.FileWebMergePartsReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown request context")
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 获取文件上传缓存信息。
	var cachedFileInfo *uploadingFileInfo
	{
		log.Info(ctx, "get file cache information")
		var fileCache string
		fileCache, err = conn.RedisClient(ctx).HGet(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				log.Warn(ctx, "uploading file info is not found", req.FileID)
				err = errs.New(consts.ErrParameterInvalid)
				return
			}
			log.Error(ctx, "failed to get file cache information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		cachedFileInfo = &uploadingFileInfo{}
		if err = json.Unmarshal([]byte(fileCache), cachedFileInfo); err != nil {
			log.Warn(ctx, "failed to unserialize file cache information", err, fileCache)
		}
	}

	// 校验，是同一个人上传。
	{
		log.Info(ctx, "verify file merge request")
		if cachedFileInfo.User != user.ID {
			log.Warn(ctx, "user who uploading file is incorrect", user.NameEn)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 校验，上传时长在范围内。
	{
		uploadingMaximumInterval := cfg.Get().Backend().FileUploadingMaximumInterval()
		if time.Since(time.Unix(cachedFileInfo.TimeSecond, 0)) > uploadingMaximumInterval {
			log.Warn(ctx, "uploading file is too large for exceeded", uploadingMaximumInterval)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
	}

	// 加锁，避免多个合并请求。
	{
		log.Info(ctx, "lock file merge request")
		lockKey := fmt.Sprintf(consts.RedisKeyFileUploadLockFmt, req.FileID)
		var success bool
		success, err = conn.RedisLock(ctx, lockKey, 0, time.Hour)
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if !success {
			log.Warn(ctx, "multiple merge request", req.FileID)
			err = errs.New(consts.ErrParameterInvalid)
			return
		}
		defer func() {
			_, err2 := conn.RedisUnlock(ctx, lockKey)
			log.ErrorIf(ctx, err2, "deleting redis key failed")
		}()
	}

	// 获取分片信息。
	var chunkNumbers []redis.Z
	{
		log.Info(ctx, "get file chunk information")
		chunkNumbers, err = conn.RedisClient(ctx).ZRangeWithScores(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			0,
			-1,
		).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Error(ctx, "failed to get file chunk information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 检查分片序号。
	var fileChunkTusdIDs []string
	var members []any
	{
		log.Info(ctx, "verify file chunk")
		fileChunkTusdIDs = make([]string, 0, len(chunkNumbers))
		fileSize := 0
		members = make([]any, len(chunkNumbers))
		for i, v := range chunkNumbers {
			members[i] = v.Member
			if int(v.Score) != i+1 {
				log.Warn(ctx, "file chunk is not in orderly")
				err = errs.New(consts.ErrFilePartNotOrder)
				return
			}
			tusdID, size := getFileChunkMember(v.Member)
			if len(tusdID) <= 0 {
				log.Error(ctx, "file chunk info is invalid", v.Member)
				err = errs.New(consts.ErrSystem)
				return
			}
			fileSize += size
			fileChunkTusdIDs = append(fileChunkTusdIDs, tusdID)
		}
		// 校验，分片数据大小与约定的要一致。
		if int64(fileSize) != cachedFileInfo.Size {
			log.Warn(ctx, "file size is not equaled")
			err = errs.New(consts.ErrFileSizeInvalid)
			return
		}
	}

	// 请求 Tusd 合并分片。
	var tusdID string
	{
		log.Info(ctx, "merge file chunks in tusd")
		tusdID, err = conn.TusdClient(ctx).MergeParts(ctx, fileChunkTusdIDs)
		if err != nil {
			log.Error(ctx, "failed to merge file chunks in tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存文件信息到数据库。
	{
		log.Info(ctx, "save file information")
		if err = createFile(ctx, &model.File{
			FileID:      req.FileID,
			TusdID:      tusdID,
			UserID:      cachedFileInfo.User,
			AppID:       cachedFileInfo.AppID,
			Name:        cachedFileInfo.Name,
			Md5:         cachedFileInfo.MD5,
			Size:        int(cachedFileInfo.Size),
			Type:        cachedFileInfo.Type,
			CreatedTime: time.Unix(cachedFileInfo.TimeSecond, 0),
		}); err != nil {
			return
		}
	}

	// 删除缓存信息。
	{
		log.ErrorIf(ctx, conn.RedisClient(ctx).HDel(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Err(),
			"remove file upload cache failed", req.FileID)
	}

	// 删除分片缓存。
	{
		log.ErrorIf(ctx, conn.RedisClient(ctx).
			ZRem(ctx, fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID), members...).Err(),
			"remove file chunk cache information failed", req.FileID)
	}

	return
}

// FileAPIDownload 下载。
func FileAPIDownload(ctx context.Context, req *protocol.FileAPIDownloadReq) (fileObj *FileInfo, err error) {
	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库，获取文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Name,
			fileDo.Size,
			fileDo.TusdID,
			fileDo.Type,
		).Where(
			fileDo.FileID.Eq(req.FileID),
			fileDo.AppID.Eq(app.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrFileNotFound, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 根据文件类型校验下载权限。
	{
		log.Info(ctx, "verify file download request")
		var hasRight bool
		var capability int
		switch file.Type {
		case model.FileTypeAndroidSigning, model.FileTypeWindowsSigning, model.FileTypeAppleSigning,
			model.FileTypeMicrosoftSigning:
			capability = model.CapabilityGetSignJobInfo
		default:
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
				"cannot download this type file")
			return
		}
		hasRight, err = APIAccountHasAnyRight(ctx, apiAccount.ID, capability)
		if err != nil {
			return
		}
		if !hasRight {
			err = errs.NewWithStatus(consts.ErrPermissionDenied, http.StatusUnauthorized)
			return
		}
	}

	// 从 Tusd 下载文件。
	var tusdResult *tus.GetResult
	{
		log.Info(ctx, "download file from tusd")
		tusdResult, err = conn.TusdClient(ctx).Get(ctx, &tus.GetRequest{Location: file.TusdID})
		if err != nil {
			log.Error(ctx, "failed to download file from tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if tusdResult.ContentLength != file.Size {
			log.Warn(ctx, "file size is different", file.Size, tusdResult.ContentLength)
		}
	}

	fileObj = &FileInfo{Name: file.Name, Size: int64(tusdResult.ContentLength), Reader: tusdResult.Body}

	return
}

// FileAPIInitial 初始化上传。
func FileAPIInitial(ctx context.Context, req *protocol.FileAPIInitialReq) (rsp *protocol.FileAPIInitialRsp, err error) {
	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 校验应用。
	{
		log.Info(ctx, "verify app")
		if app.Status != model.AppStatusValid {
			err = errs.NewWithStatus(consts.ErrAppStatusNotInvalid, http.StatusForbidden)
			return
		}
	}

	// 按文件类型校验文件。
	{
		log.Info(ctx, "verify uploading file request")
		var hasRight bool
		var capabilities []int
		switch req.Type {
		case model.FileTypeAndroidSigning:
			capabilities = []int{model.CapabilitySubmitAndroidSignJob}
		case model.FileTypeAppleSigning:
			capabilities = []int{model.CapabilitySubmitAppleSignJob}
		case model.FileTypeWindowsSigning:
			capabilities = []int{model.CapabilitySubmitWindowsPESignJob, model.CapabilitySubmitWHQLSignJob}
		default:
			// 不允许上传。
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusBadRequest)
			return
		}
		hasRight, err = APIAccountHasAnyRight(ctx, apiAccount.ID, capabilities[0], capabilities[1:]...)
		if err != nil {
			return
		}
		if !hasRight {
			err = errs.NewWithStatus(consts.ErrPermissionDenied, http.StatusUnauthorized)
			return
		}
	}

	// 检查数据库中是否存在相同的文件。
	var now time.Time
	{
		log.Info(ctx, "check if there is identical file in database")
		now = time.Now()
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableName(now))
		var file *model.File
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.FileID,
		).Where(
			fileDo.Md5.Eq(strings.ToLower(req.MD5)),
			fileDo.Name.Eq(filepath.Base(req.Name)),
			fileDo.Size.Eq(int(req.Size)),
			fileDo.Type.Eq(req.Type),
			fileDo.APIAccountID.Eq(apiAccount.ID),
			fileDo.AppID.Eq(app.ID),
		).Order(fileDo.ID.Desc()).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if file != nil {
			rsp = &protocol.FileAPIInitialRsp{FileID: file.FileID, Exist: true}
			return
		}
	}

	// 缓存文件上传信息。
	var fileID string
	{
		log.Info(ctx, "cache file upload information")
		log.Info(ctx, "generate file id")
		fileID, err = generateID(ctx, IDFile)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "reclaim file id failed", fileID)
			}
		}()
		fileInfoBytes, _ := json.Marshal(&uploadingFileInfo{
			Name:       req.Name,
			Size:       req.Size,
			MD5:        req.MD5,
			Type:       req.Type,
			TimeSecond: now.Unix(),
			AccountID:  apiAccount.ID,
			AppID:      app.ID,
		})
		err = conn.RedisClient(ctx).HSet(ctx, consts.RedisKeyFileUploadInfo, fileID, string(fileInfoBytes)).Err()
		if err != nil {
			log.Error(ctx, "failed to cache file upload information in redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	rsp = &protocol.FileAPIInitialRsp{FileID: fileID}

	return
}

// FileAPIUploadPart 上传分片。
func FileAPIUploadPart(ctx context.Context, req *protocol.FileAPIUploadPartReq) (err error) {
	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取文件上传缓存记录。
	var cachedFileInfo *uploadingFileInfo
	{
		log.Info(ctx, "get file cache information")
		var fileCache string
		fileCache, err = conn.RedisClient(ctx).HGet(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"no uploading file found")
				return
			}
			log.Error(ctx, "failed to get file cache information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		cachedFileInfo = &uploadingFileInfo{}
		if err = json.Unmarshal([]byte(fileCache), cachedFileInfo); err != nil {
			log.Warn(ctx, "unserializing file cache information failed", err, fileCache)
		}
	}

	// 确保是同一个人上传。
	{
		log.Info(ctx, "validate file upload request")
		if cachedFileInfo.AccountID != apiAccount.ID {
			log.Warn(ctx, "user who uploading file is incorrect", apiAccount.AccountID)
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "not the same user")
			return
		}
	}

	// 上传时长不能太长。
	{
		maximumInterval := cfg.Get().Backend().FileUploadingMaximumInterval()
		if time.Since(time.Unix(cachedFileInfo.TimeSecond, 0)) > maximumInterval {
			log.Warn(ctx, "uploading file is too large for exceeded", maximumInterval, cachedFileInfo.TimeSecond)
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusRequestTimeout)
			return
		}
	}

	// 加锁，避免同序号分片覆盖。
	{
		log.Info(ctx, "lock uploading file part")
		lockKey := fmt.Sprintf(consts.RedisKeyFileUploadPartLockFmt, req.FileID, req.ChunkNumber)
		var success bool
		success, err = conn.RedisLock(ctx, lockKey, 0, time.Hour)
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		defer func() {
			_, err2 := conn.RedisUnlock(ctx, lockKey)
			log.ErrorIf(ctx, err2, "deleting redis key failed")
		}()
		if !success {
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusLocked)
			return
		}
	}

	// 校验，确保该序号分片还未上传。
	var chunkNumbers []string
	{
		log.Info(ctx, "check file chunk number")
		chunkNumbers, err = conn.RedisClient(ctx).ZRangeByScore(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			&redis.ZRangeBy{Min: strconv.Itoa(req.ChunkNumber), Max: strconv.Itoa(req.ChunkNumber)},
		).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Error(ctx, "failed to get file chunk information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if len(chunkNumbers) > 0 {
			err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest, "file part chunk exists")
			return
		}
	}

	// 上传到 Tusd。
	var tusdID string
	{
		log.Info(ctx, "upload file part to tusd")
		var fileReader multipart.File
		fileReader, err = req.Chunk.Open()
		if err != nil {
			log.Error(ctx, "failed to open file part", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		tusdID, err = conn.TusdClient(ctx).UploadPartByIO(ctx, fileReader, int(req.Chunk.Size))
		if err != nil {
			log.Error(ctx, "failed to upload file part to tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		// 失败则丢弃分片。
		defer func() {
			if err != nil {
				log.ErrorIf(ctx, conn.TusdClient(ctx).DiscardParts(ctx, []string{tusdID}),
					"deleting file part failed", tusdID)
			}
		}()
	}

	// 记录分片信息。
	{
		log.Info(ctx, "cache file chunk information")
		err = conn.RedisClient(ctx).ZAdd(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			redis.Z{Score: float64(req.ChunkNumber), Member: toFileChunkMember(tusdID, req.Chunk.Size)},
		).Err()
		if err != nil {
			log.Error(ctx, "failed to cache file chunk information to redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// FileAPIMergeParts 合并分片。
func FileAPIMergeParts(ctx context.Context, req *protocol.FileAPIMergePartsReq) (err error) {
	// 获取上下文信息。
	var apiAccount *model.APIAccount
	var app *model.App
	{
		log.Info(ctx, "get context information")
		apiAccount = ctxs.APIAccount(ctx)
		app = ctxs.App(ctx)
		if apiAccount == nil || app == nil {
			log.Warn(ctx, "unknown context", app, apiAccount)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 获取文件上传缓存信息。
	var cachedFileInfo *uploadingFileInfo
	{
		log.Info(ctx, "get file cache information")
		var fileCache string
		fileCache, err = conn.RedisClient(ctx).HGet(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusBadRequest,
					"uploading file not found")
				return
			}
			log.Error(ctx, "failed to get file cache information from redis", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		cachedFileInfo = &uploadingFileInfo{}
		if err = json.Unmarshal([]byte(fileCache), cachedFileInfo); err != nil {
			log.Warn(ctx, "failed to unserialize file cache information", err, fileCache)
		}
	}

	// 校验，是同一个人上传。
	{
		log.Info(ctx, "verify file merge request")
		if cachedFileInfo.AccountID != apiAccount.ID {
			log.Warn(ctx, "user who uploading file is incorrect", apiAccount)
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusBadRequest)
			return
		}
	}

	// 校验，上传时长不范围内。
	{
		uploadingMaximumInterval := cfg.Get().Backend().FileUploadingMaximumInterval()
		if time.Since(time.Unix(cachedFileInfo.TimeSecond, 0)) > uploadingMaximumInterval {
			log.Warn(ctx, "uploading file is too large for exceeded", uploadingMaximumInterval)
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusRequestTimeout)
			return
		}
	}

	// 加锁，避免多个合并请求。
	{
		log.Info(ctx, "lock file merge request")
		lockKey := fmt.Sprintf(consts.RedisKeyFileUploadLockFmt, req.FileID)
		var success bool
		success, err = conn.RedisLock(ctx, lockKey, 0, time.Hour)
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if !success {
			log.Warn(ctx, "multiple merge request", req.FileID)
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusLocked)
			return
		}
		defer func() {
			_, err2 := conn.RedisUnlock(ctx, lockKey)
			log.ErrorIf(ctx, err2, "deleting redis key failed")
		}()
	}

	// 获取分片信息。
	var chunkNumbers []redis.Z
	{
		log.Info(ctx, "get file chunk information")
		chunkNumbers, err = conn.RedisClient(ctx).ZRangeWithScores(
			ctx,
			fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
			0,
			-1,
		).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Error(ctx, "failed to get file chunk information", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 检查分片序号。
	var fileChunkTusdIDs []string
	var members []any
	{
		log.Info(ctx, "verify file chunk")
		fileChunkTusdIDs = make([]string, 0, len(chunkNumbers))
		fileSize := 0
		members = make([]any, len(chunkNumbers))
		for i, v := range chunkNumbers {
			members[i] = v.Member
			if int(v.Score) != i+1 {
				err = errs.NewWithStatus(consts.ErrFilePartNotOrder, http.StatusBadRequest)
				return
			}
			tusdID, size := getFileChunkMember(v.Member)
			if len(tusdID) <= 0 {
				log.Error(ctx, "file chunk info is invalid", v.Member)
				err = errs.New(consts.ErrSystem)
				return
			}
			fileSize += size
			fileChunkTusdIDs = append(fileChunkTusdIDs, tusdID)
		}
		// 校验，分片数据大小与约定的要一致。
		if int64(fileSize) != cachedFileInfo.Size {
			err = errs.NewWithStatusMsg(consts.ErrFileSizeInvalid, http.StatusBadRequest, "file size is not equaled")
			return
		}
	}

	// 请求 Tusd 合并分片。
	var tusdID string
	{
		log.Info(ctx, "merge file chunk in tusd")
		tusdID, err = conn.TusdClient(ctx).MergeParts(ctx, fileChunkTusdIDs)
		if err != nil {
			log.Error(ctx, "failed to merge file chunk in tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 保存文件信息到数据库。
	{
		log.Info(ctx, "save file information")
		if err = createFile(ctx, &model.File{
			FileID:       req.FileID,
			TusdID:       tusdID,
			APIAccountID: cachedFileInfo.AccountID,
			AppID:        cachedFileInfo.AppID,
			Name:         cachedFileInfo.Name,
			Md5:          cachedFileInfo.MD5,
			Size:         int(cachedFileInfo.Size),
			Type:         cachedFileInfo.Type,
			CreatedTime:  time.Unix(cachedFileInfo.TimeSecond, 0),
		}); err != nil {
			return
		}
	}

	// 删除缓存信息。
	{
		log.ErrorIf(ctx, conn.RedisClient(ctx).HDel(ctx, consts.RedisKeyFileUploadInfo, req.FileID).Err(),
			"remove file upload cache failed", req.FileID)
	}

	// 删除分片缓存。
	{
		log.ErrorIf(ctx, conn.RedisClient(ctx).
			ZRem(ctx, fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID), members...).Err(),
			"remove file chunk cache information failed", req.FileID)
	}

	return
}

// FileInternalDownload 下载。
func FileInternalDownload(ctx context.Context, req *protocol.FileInternalDownloadReq) (fileObj *FileInfo, err error) {
	// 查库，获取文件信息。
	var file *model.File
	{
		log.Info(ctx, "get file information")
		fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(req.FileID))
		file, err = fileDo.WithContext(ctx).Select(
			fileDo.Name,
			fileDo.Size,
			fileDo.TusdID,
			fileDo.Type,
		).Where(
			fileDo.FileID.Eq(req.FileID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrFileNotFound, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve file information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 从 Tusd 下载文件。
	var tusdResult *tus.GetResult
	{
		log.Info(ctx, "download file from tusd")
		tusdResult, err = conn.TusdClient(ctx).Get(ctx, &tus.GetRequest{Location: file.TusdID})
		if err != nil {
			log.Error(ctx, "failed to download file from tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if tusdResult.ContentLength != file.Size {
			log.Warn(ctx, "file size is different", file.Size, tusdResult.ContentLength)
		}
	}

	fileObj = &FileInfo{Name: file.Name, Size: int64(tusdResult.ContentLength), Reader: tusdResult.Body}

	return
}

// FileInternalUpload 文件上传。
func FileInternalUpload(ctx context.Context, req *protocol.FileInternalUploadReq) (fileID string, err error) {

	// 上传到 Tusd。
	var tusdID, ms5Summary string
	{
		log.Info(ctx, "upload file to tusd")
		hash := md5.New()
		reader := io.TeeReader(req.Body, hash)
		tusdID, err = conn.TusdClient(ctx).MultipleUploadFromReader(ctx, reader)
		util.CloseIO(ctx, req.Body)
		if err != nil {
			log.Error(ctx, "failed to upload file to tusd", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		ms5Summary = hex.EncodeToString(hash.Sum(nil))
		defer func() {
			if err != nil {
				_, err2 := conn.TusdClient(ctx).Delete(ctx, &tus.DeleteRequest{Location: tusdID})
				log.ErrorIf(ctx, err2, "delete tusd file failed")
			}
		}()
	}

	// 文件保存入库。
	{
		log.Info(ctx, "save file to database")
		now := time.Now()
		fileID, err = generateIDWithTime(ctx, IDFile, now)
		if err != nil {
			return
		}
		err = createFile(ctx, &model.File{
			FileID:      fileID,
			TusdID:      tusdID,
			Name:        req.Name,
			AppID:       req.AppID,
			Md5:         ms5Summary,
			Size:        int(req.Size),
			Type:        req.Type,
			CreatedTime: now,
		})
		if err != nil {
			return
		}
	}

	return
}

// GetFileNamesByIDs 获取文件名。
func GetFileNamesByIDs(ctx context.Context, fileIDs []string) (map[string]string, error) {
	if len(fileIDs) <= 0 {
		return make(map[string]string), nil
	}
	fileIDToName := make(map[string]string, len(fileIDs))
	for k, v := range getFilesTable(fileIDs) {
		fileDo := conn.MySQLClient(ctx).File.Table(k)
		files, err := fileDo.WithContext(ctx).Select(
			fileDo.FileID,
			fileDo.Name,
		).Where(
			fileDo.FileID.In(v...),
		).Find()
		if err != nil && !errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
			log.Error(ctx, "failed to retrieve file names from database", err, files)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		for _, info := range files {
			fileIDToName[info.FileID] = info.Name
		}
	}
	return fileIDToName, nil
}

func createFile(ctx context.Context, fileInfo *model.File) error {
	fileTxDo := conn.MySQLTxClient(ctx).File.Table(model.GetFileTableNameByID(fileInfo.FileID))
	fields := make([]field.Expr, 0, 8)
	fields = append(fields, fileTxDo.FileID, fileTxDo.TusdID)
	if fileInfo.Type > 0 {
		fields = append(fields, fileTxDo.Type)
	}
	if fileInfo.UserID > 0 {
		fields = append(fields, fileTxDo.UserID)
	}
	if fileInfo.AppID > 0 {
		fields = append(fields, fileTxDo.AppID)
	}
	if fileInfo.APIAccountID > 0 {
		fields = append(fields, fileTxDo.APIAccountID)
	}
	if fileInfo.Size > 0 {
		fields = append(fields, fileTxDo.Size)
	}
	if len(fileInfo.Name) > 0 {
		fields = append(fields, fileTxDo.Name)
	}
	if len(fileInfo.Md5) > 0 {
		fields = append(fields, fileTxDo.Md5)
	}
	if !fileInfo.CreatedTime.IsZero() {
		fields = append(fields, fileTxDo.CreatedTime)
	}
	err := fileTxDo.WithContext(ctx).Select(fields...).Create(fileInfo)
	if err != nil {
		log.Error(ctx, "failed to save file information to database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}

// CronCleanUploadingFiles 清理文件垃圾分片。
func CronCleanUploadingFiles(ctx context.Context, _ string, _ time.Time) {
	// 清理过期了，还没上传完的任务。
	log.Debug(ctx, "clean up expired tasks that have not been uploaded yet")
	fileIDToInfo, err := conn.RedisClient(ctx).HGetAll(ctx, consts.RedisKeyFileUploadInfo).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to run redis command", err)
		return
	}
	for fileID, fileInfo := range fileIDToInfo {
		fileCachedInfo := &uploadingFileInfo{}
		if err = json.Unmarshal([]byte(fileInfo), fileCachedInfo); err != nil {
			log.Warn(ctx, "failed to unmarshal file cache information", err, fileInfo)
		}

		// 判断是否过期。
		if time.Since(time.Unix(fileCachedInfo.TimeSecond, 0)) >
			cfg.Get().Backend().FileUploadingMaximumInterval()+time.Minute {
			log.Info(ctx, "clean upload file", fileID)
			cleanExpiredUploadFiles(ctx, fileID)
		}
	}

	// 清理在刚好在合并分片时上传的多余分片。
	log.Info(ctx, "clean up file parts during file merging process")
	redisResult, err := conn.RedisClient(ctx).Keys(ctx, consts.RedisKeyFileUploadPartInfoKeyPrefix+"*").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to run redis command", err)
		return
	}
	var exist bool
	for _, v := range redisResult {
		fileID := strings.TrimPrefix(v, consts.RedisKeyFileUploadInfo)
		// 查询缓存是否存在，不存在就表明是多余的分片。
		exist, err = conn.RedisClient(ctx).HExists(ctx, consts.RedisKeyFileUploadInfo, fileID).Result()
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			continue
		}
		if !exist {
			log.Info(ctx, "clean file parts", fileID)
			var redisResult2 []redis.Z
			redisResult2, err = conn.RedisClient(ctx).ZRangeWithScores(ctx,
				fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, fileID), 0, -1).Result()
			if err != nil && !errors.Is(err, redis.Nil) {
				log.Error(ctx, "failed to get file parts cache information", err)
				continue
			}
			deleteFilePartCacheAndTusd(ctx, fileID,
				util.ListTo(redisResult2, func(e redis.Z) any { return e.Member }))
		}
	}
}

func cleanExpiredUploadFiles(ctx context.Context, fileID string) {
	// 加锁，避免该文件正在合并。
	log.Info(ctx, "get processing lock")
	lockKey := fmt.Sprintf(consts.RedisKeyFileUploadLockFmt, fileID)
	success, err := conn.RedisLock(ctx, lockKey, 0, time.Minute)
	if err != nil {
		log.Error(ctx, "failed to run redis command", err)
		return
	}
	if !success {
		log.Warn(ctx, "uploading file is in merging")
		return
	}
	defer func() {
		_, err = conn.RedisUnlock(ctx, lockKey)
		log.ErrorIf(ctx, err, "deleting cache has no effect")
	}()

	// 删除文件上传缓存信息。
	log.Info(ctx, "delete file upload cache")
	redisResult, err := conn.RedisClient(ctx).HDel(ctx, consts.RedisKeyFileUploadInfo, fileID).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to run redis command", err)
		return
	}
	if redisResult <= 0 {
		log.Warn(ctx, "deleting file upload information has no effect")
		return
	}
	defer func() {
		// 回收文件 ID。
		log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id", fileID)
	}()

	// 获取文件的所有分片缓存数据。
	log.Info(ctx, "delete all file parts information")
	redisResult2, err := conn.RedisClient(ctx).ZRangeWithScores(ctx,
		fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, fileID), 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to run redis command", err)
		return
	}
	deleteFilePartCacheAndTusd(ctx, fileID, util.ListTo(redisResult2, func(e redis.Z) any { return e.Member }))
}

func deleteFilePartCacheAndTusd(ctx context.Context, fileID string, members []any) {
	if len(members) <= 0 {
		log.Warn(ctx, "no file parts information need to delete")
		return
	}

	// 删除在 Tusd 中的文件分片。
	log.Info(ctx, "delete file parts in tusd")
	tusdIDs := make([]string, 0, len(members))
	for _, v := range members {
		tusdID, _ := getFileChunkMember(v)
		if len(tusdID) <= 0 {
			log.Error(ctx, "unknown tusd id", v)
			continue
		}
		tusdIDs = append(tusdIDs, tusdID)
	}
	if len(tusdIDs) > 0 {
		log.ErrorIf(ctx, conn.TusdClient(ctx).DiscardParts(ctx, tusdIDs),
			"failed to delete file parts in tusd", tusdIDs)
	}

	log.Info(ctx, "delete file parts cache")
	redisResult, err := conn.RedisClient(ctx).ZRem(ctx,
		fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, fileID), members...).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error(ctx, "failed to run redis command", err)
	}
	if int64(len(members)) != redisResult {
		log.Error(ctx, "the number of deleted file parts is inconsistent", len(members), redisResult)
	}
}

func toFileChunkMember(tusdID string, fileSize int64) string {
	return fmt.Sprintf("%s,%d", tusdID, fileSize)
}

func getFileChunkMember(member any) (tusdID string, fileSize int) {
	pair := strings.Split(fmt.Sprint(member), ",")
	if len(pair) != 2 {
		return "", 0
	}
	fileSize, err := strconv.Atoi(pair[1])
	if err != nil {
		return "", 0
	}
	return pair[0], fileSize
}

func getFilesTable(fileIDs []string) map[string][]string {
	tableNameToFileIDs := make(map[string][]string, len(fileIDs)/2)
	for _, fileID := range fileIDs {
		tableName := model.GetFileTableNameByID(fileID)
		tableNameToFileIDs[tableName] = append(tableNameToFileIDs[tableName], fileID)
	}
	return tableNameToFileIDs
}
