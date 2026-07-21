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
	"mime/multipart"
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
		chunkNumbers, err = conn.RedisClient(ctx).ZRangeArgs(
			ctx,
			redis.ZRangeArgs{
				Key:     fmt.Sprintf(consts.RedisKeyFileUploadPartInfoFmt, req.FileID),
				Start:   strconv.Itoa(req.ChunkNumber),
				Stop:    strconv.Itoa(req.ChunkNumber),
				ByScore: true,
			},
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
