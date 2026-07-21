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
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

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
