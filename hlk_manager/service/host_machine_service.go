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
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
)

var allVirtualMachineLogFile = sync.Map{}

// HostInternalReportLog 接收虚拟机写来的日志，并写入硬盘中。
func HostInternalReportLog(ctx context.Context, data []byte) (err error) {
	// 获取日志文件句柄。
	var file *lumberjack.Logger
	{
		log.Info(ctx, "get file object")
		requestIP := strings.ReplaceAll(ctxs.RequestIP(ctx), ".", "_")
		fileAny, _ := allVirtualMachineLogFile.LoadOrStore(requestIP, &lumberjack.Logger{
			Filename:   filepath.Join(filepath.Dir(cfg.Get().Log().Name()), requestIP+".log"),
			MaxSize:    cfg.Get().Log().FileMaximumSizeByMegabytes(),
			MaxAge:     cfg.Get().Log().FileMaximumAgeByDays(),
			MaxBackups: cfg.Get().Log().FileMaximumBackups(),
			LocalTime:  true,
		})
		file = fileAny.(*lumberjack.Logger)
	}

	// 将日志写入文件。
	{
		log.Info(ctx, "write log to file")
		var n int64
		n, err = io.Copy(file, bytes.NewReader(data))
		if err != nil {
			log.Error(ctx, "writing log to disk failed", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if n != int64(len(data)) {
			log.Error(ctx, "not all written in disk", n, len(data))
		}
	}

	return
}

// HostInternalRetoreTestMachine 重置测试机到检查点。
func HostInternalRetoreTestMachine(ctx context.Context) (err error) {
	// 运行 PowerShell 命令回退虚拟机到检查点。
	{
		testMachine := strings.ReplaceAll(ctxs.RequestIP(ctx), ".", "_")
		log.Info(ctx, "run powershell restore test machine", testMachine)
		var output []byte
		output, err = util.RunPowerShellCommands(ctx, fmt.Sprintf("Restore-VMSnapshot -VMName '%s' -Name %s",
			testMachine, consts.VirtualMachineCheckPointName), "Y")
		if err != nil {
			log.Error(ctx, "restoring test machine failed", err, output)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		log.Debug(ctx, "restoring test machine successfully", output)
	}

	return
}

// CloseVirtualMachineLogFile 关闭虚拟机日志文件句柄。
func CloseVirtualMachineLogFile(ctx context.Context) {
	allVirtualMachineLogFile.Range(func(k, v any) bool {
		file := v.(*lumberjack.Logger)
		log.ErrorIf(ctx, file.Close(), "failed to close log file")
		return true
	})
}
