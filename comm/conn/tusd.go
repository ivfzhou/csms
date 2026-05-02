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

package conn

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

var (
	tusdClient          tus.TusClient
	tusdInitializedFlag int32
	tusdAddress         string
	tusdLogLevel        int
	tusdUpdateLock      sync.Mutex
)

// InitializeTusdConnection 初始化 Tusd 服务器连接。
func InitializeTusdConnection(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&tusdInitializedFlag, 0, 1) {
		return
	}

	log.Info(ctx, "connecting to tusd server")

	// 新建客户端。
	tusdLogLevel = log.LevelTusdFrom[log.GetLevel()]
	tusdAddress = getTusdAddress(cfg.Get())
	tusdClient = tus.NewClient(tusdAddress, tus.WithHTTPClient(util.GetHTTPClient()),
		tus.WithLogger(log.GetTusdLogger()), tus.WithLogLevel(tusdLogLevel))

	// 测试连接。
	log.Info(ctx, "option tusd server", tusdAddress)
	tusdResult, err := tusdClient.Options(ctx)
	if err != nil {
		log.Fatal(ctx, consts.ExitCodeInitialTusdConnectionError, "failed to option tusd server", err)
	}
	if tusdResult.HTTPStatus < http.StatusOK && tusdResult.HTTPStatus >= http.StatusMultipleChoices {
		log.Fatal(ctx, consts.ExitCodeInitialTusdConnectionError, "optioning tusd server result is unexpected",
			tusdResult)
	}

	// 监听配置更新。
	log.Info(ctx, "watch tusd configuration update")
	cfg.RegisterNotifier(watchTusdConfigurationUpdate)

	log.Warn(ctx, "successfully connected to tusd server", tusdResult)
}

// TusdClient 获取 Tusd 操作对象。
func TusdClient(_ context.Context) tus.TusClient {
	return tusdClient
}

// MockTusdClient 单测模式下替换实现。
func MockTusdClient(client tus.TusClient) {
	if consts.UnitTestMode() {
		tusdClient = client
	}
}

// 获取 Tusd 地址。
func getTusdAddress(configurer cfg.Configurer) string {
	return fmt.Sprintf("%s:%d", configurer.Tusd().Host(), configurer.Tusd().Port())
}

// 监听配置更新。
func watchTusdConfigurationUpdate(configurer cfg.Configurer) {
	tusdUpdateLock.Lock()
	defer tusdUpdateLock.Unlock()

	newLogLevel := log.LevelTusdFrom[log.GetLevel()]
	newAddress := getTusdAddress(configurer)

	if newAddress == tusdAddress && newLogLevel == tusdLogLevel {
		return
	}
	ctx := ctxs.New()
	newTusdClient := tus.NewClient(newAddress, tus.WithHTTPClient(util.GetHTTPClient()),
		tus.WithLogger(log.GetTusdLogger()), tus.WithLogLevel(newLogLevel))

	// 测试连接。
	log.Info(ctx, "option tusd server", newAddress)
	tusdResult, err := newTusdClient.Options(ctx)
	if err != nil {
		log.Error(ctx, "failed to option tusd server", err)
		return
	}
	if tusdResult.HTTPStatus < http.StatusOK && tusdResult.HTTPStatus >= http.StatusMultipleChoices {
		log.Error(ctx, "optioning tusd server result is unexpected", tusdResult)
		return
	}

	tusdClient = newTusdClient
	tusdAddress = newAddress
	tusdLogLevel = newLogLevel
	log.Info(ctx, "successfully connected to tusd server")
}
