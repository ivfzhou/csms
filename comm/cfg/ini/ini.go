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

package ini

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	configuration                     = newImpl()
	configurationFilePath             = "config.ini"
	configurationFileWatcher          *fsnotify.Watcher
	closeConfigurationFileWatcherFlag atomic.Int32
	notifiers                         []func(cfg.Configurer)
	notifiersLocker                   sync.Mutex
)

// AddCommandFlag 添加配置文件命令行参数。
func AddCommandFlag() {
	flag.StringVar(&configurationFilePath, consts.CommandFlagConfigurationFilePath,
		"config.ini", "ini config file location")
}

// Initialize 解析服务配置文件，并监听配置文件改动。
// 配置文件解析 ini 格式失败，或监听配置文件失败，会退出程序。
func Initialize(ctx context.Context) {
	// 获取配置文件路径。
	var err error
	configurationFilePath, err = filepath.Abs(configurationFilePath)
	log.FatalIf(ctx, consts.ExitCodeInitialConfigError, err,
		"failed to get configuration file path", configurationFilePath)
	_, err = os.Stat(configurationFilePath)
	log.FatalIf(ctx, consts.ExitCodeInitialConfigError, err,
		"failed to state configuration file", configurationFilePath)

	// 解析配置文件。
	configuration, _, err = parse(ctx, configurationFilePath)
	log.FatalIf(ctx, consts.ExitCodeInitialConfigError, err,
		"failed to parse configuration file", configurationFilePath)

	// 监听配置文件改动。
	configurationFileWatcher, err = fsnotify.NewWatcher()
	log.FatalIf(ctx, consts.ExitCodeInitialConfigError, err, "failed to create config file watcher")
	err = configurationFileWatcher.Add(configurationFilePath)
	log.FatalIf(ctx, consts.ExitCodeInitialConfigError, err, "failed to watch config file", configurationFilePath)
	log.Info(ctx, "watch configuration file", configurationFilePath)
	go watchAndNotify(ctx)

	// 注册实现。
	cfg.RegisterImplement(get, closeWatch, addNotifier)
}

// 解析配置。
func parse(ctx context.Context, filePath string) (*Configuration, bool, error) {
	data := newImpl()
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "failed to read ini file", err, filePath)
		return data, false, err
	}
	if len(fileData) <= 0 {
		return data, false, nil
	}
	err = ini.StrictMapTo(data, fileData)
	if err != nil {
		log.Error(ctx, "failed to load ini file", err, filePath)
		return data, false, err
	}
	data.AppleAPIConfiguration.SecretValue = strings.ReplaceAll(data.AppleAPIConfiguration.SecretValue, `\n`, "\n")
	data.AppleConfiguration.ApplyCertificateCSRValue = strings.ReplaceAll(
		data.AppleConfiguration.ApplyCertificateCSRValue, `\n`, "\n")
	data.AppleConfiguration.CertificatePrivateKeyValue = strings.ReplaceAll(
		data.AppleConfiguration.CertificatePrivateKeyValue, `\n`, "\n")
	return data, true, nil
}

// 获取配置。
func get() cfg.Configurer {
	return configuration
}

// 关闭配置监听。
func closeWatch(ctx context.Context) {
	if !closeConfigurationFileWatcherFlag.CompareAndSwap(0, 1) {
		log.Warn(ctx, "ini config already closed")
		return
	}

	if configurationFileWatcher == nil {
		log.Warn(ctx, "ini configuration watcher is nil")
		return
	}

	log.Warn(ctx, "closing ini configuration file watch")
	log.ErrorIf(ctx, configurationFileWatcher.Close(), "failed to close ini config")
}

// 添加配置监听者。
func addNotifier(notifier func(cfg.Configurer)) {
	if notifier == nil {
		log.Warn(context.Background(), "discard nil ini configuration notifier")
		return
	}

	notifiersLocker.Lock()
	defer notifiersLocker.Unlock()

	notifiers = append(notifiers, notifier)
}

// 监听配置更新。
func watchAndNotify(ctx context.Context) {
	var err error
	var ok bool
	var event fsnotify.Event
	var newConfiguration *Configuration
	for {
		select {
		case event, ok = <-configurationFileWatcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Chmod:
				log.Warn(ctx, "configuration file chmod")
			case fsnotify.Remove:
				log.Warn(ctx, "configuration file removed")
			case fsnotify.Rename:
				log.Warn(ctx, "configuration file renamed", event.Name)
			case fsnotify.Write, fsnotify.Create:
				newConfiguration, ok, _ = parse(ctx, event.Name)
				if !ok {
					continue
				}
				configuration = newConfiguration
				notifyUpdate(ctx, newConfiguration)
				log.Warn(ctx, "update ini configuration successfully")
			default:
				log.Warn(ctx, "unknown watcher event", event.Op)
			}
		case err, ok = <-configurationFileWatcher.Errors:
			if ok {
				log.Error(ctx, "watching ini configuration file occur error", err)
			}
		}
	}
}

// 监听配置推送更新。
func notifyUpdate(ctx context.Context, newConfiguration cfg.Configurer) {
	notifiersLocker.Lock()
	defer notifiersLocker.Unlock()

	if len(notifiers) <= 0 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(notifiers))

	for _, v := range notifiers {
		go func(notify func(cfg.Configurer)) {
			defer wg.Done()
			defer func() {
				if p := recover(); p != nil {
					log.Error(ctx, "panic occurred when updating ini configuration", p)
				}
			}()
			notify(newConfiguration)
		}(v)
	}

	wg.Wait()
}
