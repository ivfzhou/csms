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
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

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
