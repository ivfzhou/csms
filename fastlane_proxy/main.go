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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/ivfzhou/csms/comm/cfg"
	iniCfg "gitee.com/ivfzhou/csms/comm/cfg/ini"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	propertiesI18n "gitee.com/ivfzhou/csms/comm/i18n/properties"
	"gitee.com/ivfzhou/csms/comm/log"
	logImpl "gitee.com/ivfzhou/csms/comm/log/impl"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/fastlane_proxy/consts"
	"gitee.com/ivfzhou/csms/fastlane_proxy/protocol"
	"gitee.com/ivfzhou/csms/fastlane_proxy/route"
)

var (
	ctx, cancel       = context.WithCancel(ctxs.New())
	exitSignalChannel = make(chan os.Signal, 1)
)

func init() {
	// 解析命令行参数。
	cc.AddTestModeCommandFlag()
	iniCfg.AddCommandFlag()
	propertiesI18n.AddCommandFlag()
	util.AddIPCommandFlag()
	consts.AddCommandFlag()
	flag.Parse()

	// 初始化配置和日志。
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
	propertiesI18n.Initialize(ctx)
	protocol.Init(ctx)

	// 监听进程退出信号。
	signal.Notify(exitSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV /*, syscall.SIGUSR2*/)
}

func main() {
	// 初始化路由。
	router := route.Initialize(ctx)

	// 初始化 HTTP 服务错误通道。
	serverErrorChannel := make(chan error, 1)

	// HTTP 服务对象。
	var server *http.Server

	// 创建 HTTP 服务函数。
	port := cfg.Get().FastlaneProxy().Port()
	newServerFunc := func() {
		server = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		}
		go func() { serverErrorChannel <- server.ListenAndServe() }()
	}

	// 启动 HTTP 服务。
	newServerFunc()
	log.Info(ctx, "start http server listening on", port)

	// 第一次启动，判断启动是否成功，否则退出程序。
	time.Sleep(time.Second)
	select {
	case err := <-serverErrorChannel:
		log.Fatal(ctx, cc.ExitCodeHTTPListenError, "failed to start http server", err)
	default:
	}

	// 无限重启服务，除非接收到进程退出信号。
	for {
		select {
		// 接收到进程退出信号，退出服务。
		case <-exitSignalChannel:
			// 停止接收新的 HTTP 请求。
			cancel()

			// 关闭 HTTP 服务。
			log.Warn(ctx, "exiting server")
			err := server.Shutdown(context.Background())
			log.ErrorIf(ctx, err, "failed to shutdown server gracefully")

			// 释放连接资源。
			cfg.Close(ctx)
			i18n.Close(ctx)
			log.Warn(ctx, "server exit gracefully")
			log.Close(ctx)
			return

		// 接收到外部 HTTP 服务退出，重新启动服务。
		case err := <-serverErrorChannel:
			log.Error(ctx, "detected server error", err)
			err = server.Shutdown(ctx)
			log.ErrorIf(ctx, err, "failed to shutdown server gracefully")

			log.Warn(ctx, "recreate server and serve")
			time.Sleep(time.Second)
			newServerFunc()
		}
	}
}
