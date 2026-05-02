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

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/cron"
	"gitee.com/ivfzhou/csms/backend/docs"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/route"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/cfg"
	iniCfg "gitee.com/ivfzhou/csms/comm/cfg/ini"
	"gitee.com/ivfzhou/csms/comm/conn"
	cs "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	propertiesI18n "gitee.com/ivfzhou/csms/comm/i18n/properties"
	"gitee.com/ivfzhou/csms/comm/log"
	logImpl "gitee.com/ivfzhou/csms/comm/log/impl"
	"gitee.com/ivfzhou/csms/comm/util"
)

var (
	ctx, cancel       = context.WithCancel(ctxs.New())
	exitSignalChannel = make(chan os.Signal, 1)
)

func init() {
	// 解析命令行参数。
	cs.AddTestModeCommandFlag()
	cs.AddSkipRateLimitCommandFlag()
	iniCfg.AddCommandFlag()
	propertiesI18n.AddCommandFlag()
	consts.AddCommandFlag()
	util.AddIPCommandFlag()
	flag.Parse()

	// 初始化配置和日志。
	iniCfg.Initialize(ctx)
	logImpl.InitializeConsoleLog(ctx)
	logImpl.InitializeFileLog(ctx)

	// 获取 IP 地址。
	util.InitializeLocalIP(ctx)
	if util.IPv4ToNumber(util.LocalIP) <= 0 {
		log.Fatal(ctx, cs.ExitCodeLocalIPNotFound, "failed to get local ip")
	}

	// 加载时区。
	var err error
	timeZone := cfg.Get().TimeZone()
	time.Local, err = time.LoadLocation(timeZone)
	log.FatalIf(ctx, cs.ExitCodeLoadTimezoneError, err, "failed to load time location", timeZone)

	// 初始化。
	propertiesI18n.Initialize(ctx)
	protocol.Initialize(ctx)
	conn.InitializeMySQLConnection(ctx)
	conn.InitializeRabbitMQConnection(ctx)
	conn.InitializeRedisConnection(ctx)
	conn.InitializeTusdConnection(ctx)
	cron.Initialize(ctx)
	service.Initialize(ctx)

	// 设置 Swagger 参数。
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.Get().Swagger().Host(), cfg.Get().Swagger().Port())
	docs.SwaggerInfo.Version = cfg.Get().Swagger().Version()
	docs.SwaggerInfo.BasePath = cfg.Get().Swagger().BasePath()
	docs.SwaggerInfo.Schemes = []string{cfg.Get().Swagger().Schema()}

	// 监听进程信号。
	signal.Notify(exitSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV /*, syscall.SIGUSR2*/)
}

// 启动程序入口。
//
//	@title			CSMS API
//	@version		1.0
//	@contact.name	ivfzhou
//	@contact.email	ivfzhou@126.com
//	@license.name	MulanPSL-2.0 license
//	@license.url	http://license.coscl.org.cn/MulanPSL2
//	@host			127.0.0.1:8090
//	@schemes		http
func main() {
	// 初始化路由。
	router := route.Initialize(ctx)
	internalRouter := route.InitializeInternal(ctx)

	// 初始化 HTTP 服务错误通道。
	serverErrorChannel := make(chan error, 1)
	internalServerErrorChannel := make(chan error, 1)

	// 内部和外部 HTTP 服务对象。
	var server *http.Server
	var internalServer *http.Server

	// 创建 HTTP 服务函数。
	port := cfg.Get().Backend().Port()
	newServerFunc := func() {
		server = &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  cfg.Get().Backend().ReadTimeout(),
			WriteTimeout: cfg.Get().Backend().WriteTimeout(),
		}
		go func() { serverErrorChannel <- server.ListenAndServe() }()
	}
	internalPort := cfg.Get().Backend().InternalPort()
	newInternalServerFunc := func() {
		internalServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", internalPort),
			Handler: internalRouter,
		}
		go func() { internalServerErrorChannel <- internalServer.ListenAndServe() }()
	}

	// 启动 HTTP 服务。
	newServerFunc()
	newInternalServerFunc()
	log.Info(ctx, "start http server listening on", port, "and internal port", internalPort)

	// 第一次启动，判断启动是否成功，否则退出程序。
	time.Sleep(time.Second)
	select {
	case err := <-serverErrorChannel:
		log.Fatal(ctx, cs.ExitCodeHTTPListenError, "failed to start http server", err)
	case err := <-internalServerErrorChannel:
		log.Fatal(ctx, cs.ExitCodeHTTPListenError, "failed to start internal http server", err)
	default:
	}

	// 无限重启服务，除非接收到退出信号。
	for {
		select {
		// 接收到进程退出信号，退出服务。
		case <-exitSignalChannel:
			// 停止接收新的 HTTP 请求。
			cancel()

			// 关闭 HTTP 服务。
			log.Warn(ctx, "exiting server")
			err := internalServer.Shutdown(context.Background())
			log.ErrorIf(ctx, err, "failed to shutdown internal server gracefully")
			err = server.Shutdown(context.Background())
			log.ErrorIf(ctx, err, "failed to shutdown server gracefully")

			// 释放连接资源。
			cfg.Close(ctx)
			cron.Close(ctx)
			conn.CloseRabbitMQConnection(ctx)
			conn.CloseMySQLConnection(ctx)
			conn.CloseRedisConnection(ctx)
			i18n.Close(ctx)
			log.Warn(ctx, "server exit gracefully")
			log.Close(ctx)
			return

		// 接收到内部 HTTP 服务退出，重新启动服务。
		case err := <-internalServerErrorChannel:
			log.Error(ctx, "detected internal server error", err)
			err = internalServer.Shutdown(ctx)
			log.ErrorIf(ctx, err, "failed to shutdown internal server gracefully")

			log.Warn(ctx, "recreate internal server and serve")
			time.Sleep(time.Second)
			newInternalServerFunc()

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
