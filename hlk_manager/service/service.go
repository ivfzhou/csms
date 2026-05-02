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
	"fmt"
	"path/filepath"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
)

func formatJobLog(level log.Level, format string, args ...any) string {
	str := fmt.Sprintf(format, args...)
	str = strings.Trim(str, `\r\n`)
	str = strings.Trim(str, `\n`)
	return fmt.Sprintf("%s %s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), level.String(), str)
}

func getMachineName() string {
	return strings.ReplaceAll(util.LocalIP, ".", "_")
}

func getMachinePoolName(testMachineName string) string {
	return fmt.Sprintf("pool_%s", testMachineName)
}

func getProjectName(jobID int) string {
	return fmt.Sprintf("project_%d", jobID)
}

func getTestInfoFilePath(jobID int) string {
	return filepath.Join(consts.TestInfoDirectoryPath, fmt.Sprintf("%d.json", jobID))
}

// 结束任务。
func failJob(ctx context.Context, jobID int, f string, args ...any) {
	log.Error(ctx, "fail job", jobID)
	err := httpBackendUpdateWHQLJob(ctx, &bp.WindowsInternalUpdateWHQLJobReq{
		JobID:     jobID,
		AppendLog: formatJobLog(log.LevelError, f, args...),
		Status:    model.WHQLJobStatusFailure,
	})
	if err != nil {
		log.Error(ctx, "updating job failed", err)
		return
	}
}
