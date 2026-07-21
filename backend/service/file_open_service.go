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
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"github.com/redis/go-redis/v9"
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
