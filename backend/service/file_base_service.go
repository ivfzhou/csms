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

	"github.com/pingcap/tidb/parser/mysql"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
)

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

func getFilesTable(fileIDs []string) map[string][]string {
	tableNameToFileIDs := make(map[string][]string, len(fileIDs)/2)
	for _, fileID := range fileIDs {
		tableName := model.GetFileTableNameByID(fileID)
		tableNameToFileIDs[tableName] = append(tableNameToFileIDs[tableName], fileID)
	}
	return tableNameToFileIDs
}
