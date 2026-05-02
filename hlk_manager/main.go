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
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
	"gitee.com/ivfzhou/csms/hlk_manager/route"
	"gitee.com/ivfzhou/csms/hlk_manager/service"
)

var (
	ctx, cancel       = context.WithCancel(ctxs.New())
	exitSignalChannel = make(chan os.Signal, 1)
	mode              string
	system            string
	systems           string
)

func init() {
	// 解析命令行参数。
	flag.StringVar(&mode, "mode", "", "required, mode is one of TestMachine, ControllerMachine or HostMachine")
	flag.StringVar(&system, "system", "", "required if mode is TestMachine, windows system version of test machine")
	flag.StringVar(&systems, "systems", "", "required if mode is ControllerMachine, which system versions can be controlled for machine testing, comma separated")
	iniCfg.AddCommandFlag()
	propertiesI18n.AddCommandFlag()
	consts.AddCommandFlag()
	util.AddIPCommandFlag()
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
	log.ErrorIf(ctx, err, "failed to load time location", timeZone)

	// 监听进程信号。
	signal.Notify(exitSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
}

func main() {
	switch mode {
	case consts.ModeTestMachine:
		logImpl.InitializeReportLog(ctx)
		closeChannel := service.TestMachineStart(ctx, system)

		select {
		case <-exitSignalChannel: // 接收到进程退出信号，退出服务。
			cancel()
		case <-closeChannel:
			cancel()
		}

		// 释放连接资源。
		log.Warn(ctx, "server exiting")
		<-closeChannel
		cfg.Close(ctx)
		log.Warn(ctx, "server exit gracefully")
		log.Close(ctx)
		return
	case consts.ModeControllerMachine:
		logImpl.InitializeReportLog(ctx)
		closeChannel := service.ControllerMachineStart(ctx, systems)

		select {
		case <-exitSignalChannel: // 接收到进程退出信号，退出服务。
			cancel()
		case <-closeChannel:
			cancel()
		}

		// 释放连接资源。
		log.Warn(ctx, "server exiting")
		<-closeChannel
		cfg.Close(ctx)
		log.Warn(ctx, "server exit gracefully")
		log.Close(ctx)
		return
	case consts.ModeHostMachine:
		// 初始化。
		propertiesI18n.Initialize(ctx)

		// 初始化路由。
		router := route.Initialize(ctx)

		// 初始化 HTTP 服务错误通道。
		serverErrorChannel := make(chan error, 1)

		// HTTP 服务对象。
		var server *http.Server

		// 创建 HTTP 服务函数。
		port := cfg.Get().HLKManagerHost().Port()
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
				service.CloseVirtualMachineLogFile(ctx)
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
	default:
		log.Fatal(ctx, cc.ExitCodeHLKManagerNoMode, "mode is invalid", mode)
	}
}
