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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/ivfzhou/csms/comm/cfg"
	iniCfg "gitee.com/ivfzhou/csms/comm/cfg/ini"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	logImpl "gitee.com/ivfzhou/csms/comm/log/impl"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/sign_server/consts"
	"gitee.com/ivfzhou/csms/sign_server/service"
)

var (
	ctx, cancel       = context.WithCancel(ctxs.New())
	exitSignalChannel = make(chan os.Signal, 1)
	mode              string
)

func init() {
	// 解析命令行参数。
	flag.StringVar(&mode, "mode", "", "required, mode is one of Windows, Android or Apple")
	iniCfg.AddCommandFlag()
	util.AddIPCommandFlag()
	consts.AddCommandFlag()
	flag.Parse()

	// 初始化配置。
	iniCfg.Initialize(ctx)
	logImpl.Initialize(ctx)
	logImpl.InitializeConsoleLog(ctx)
	logImpl.InitializeFileLog(ctx)

	// 获取 IP 地址。
	util.InitializeLocalIP(ctx)
	if util.IPv4ToNumber(util.LocalIP) <= 0 {
		log.Fatal(ctx, cc.ExitCodeLocalIPNotFound, "failed to get local ip")
	}

	// 加载时区。
	var err error
	timeZone := cfg.Get().TimeZone()
	time.Local, err = time.LoadLocation(timeZone)
	log.FatalIf(ctx, cc.ExitCodeLoadTimezoneError, err, "failed to load time location", timeZone)

	// 初始化。
	conn.InitializeRabbitMQConnection(ctx)

	// 监听进程信号。
	signal.Notify(exitSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
}

func main() {
	var closeChannel <-chan struct{}
	switch mode {
	case consts.ModeWindows:
		closeChannel = service.StartWindowsSignServer(ctx)
	case consts.ModeAndroid:
		closeChannel = service.StartAndroidSignServer(ctx)
	case consts.ModeApple:
		closeChannel = service.StartAppleSignServer(ctx)
	default:
		log.Fatal(ctx, cc.ExitCodeSignServerInvalidMode, "mode is invalid", mode)
	}

	// 监听退出信号。
	select {
	case <-exitSignalChannel:
		cancel()
	case <-closeChannel:
		cancel()
	}
	log.Warn(ctx, "server exiting")
	<-closeChannel

	// 释放资源。
	cfg.Close(ctx)
	conn.CloseRabbitMQConnection(ctx)
	log.Warn(ctx, "server exit gracefully")
	log.Close(ctx)
}
