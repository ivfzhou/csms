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

package properties

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	messages              map[i18n.Language]map[errs.Code]string
	languages             []i18n.Language
	messageFileWatcher    *fsnotify.Watcher
	messageFilesDirectory string
	messageLocker         sync.RWMutex
	initializedFlag       atomic.Int32
	closedFlag            atomic.Int32
)

// AddCommandFlag 添加提示语文件所在文件夹的命令参数。
func AddCommandFlag() {
	flag.StringVar(&messageFilesDirectory, consts.CommandFlagMessageFilesDirectory, ".",
		"directory of message files to load")
}

// Initialize 解析消息提示语文件，并监听文件改动。
// 若处理失败，会退出程序。
func Initialize(ctx context.Context) {
	if !initializedFlag.CompareAndSwap(0, 1) {
		return
	}

	if closedFlag.Load() > 0 {
		return
	}

	languages = []i18n.Language{i18n.LanguageEnglish, i18n.LanguageChinese}
	messages = make(map[i18n.Language]map[errs.Code]string, len(languages))
	filePathToLanguage := make(map[string]i18n.Language, len(languages))
	for _, v := range languages {
		filePath := filepath.Join(messageFilesDirectory, string(v)+`_message.properties`)
		fileAbsolutePath, err := filepath.Abs(filePath)
		log.FatalIf(ctx, consts.ExitCodeInitialMessageFileError, err, "cannot get file absolute path", filePath)
		fileData, err := os.ReadFile(fileAbsolutePath)
		if err != nil {
			log.Warn(ctx, "failed to read message file", err, fileAbsolutePath)
			messages[v] = make(map[errs.Code]string)
		} else {
			messages[v] = parse(ctx, fileData)
			filePathToLanguage[fileAbsolutePath] = v
		}
	}

	// 启动监听。
	var err error
	messageFileWatcher, err = fsnotify.NewWatcher()
	log.FatalIf(ctx, consts.ExitCodeInitialMessageFileError, err, "failed to create message file watcher")
	for v := range filePathToLanguage {
		err = messageFileWatcher.Add(v)
		log.FatalIf(ctx, consts.ExitCodeInitialMessageFileError, err, "failed to watch message file", v)
		log.Info(ctx, "watch message file", v)
	}
	go watchPropertiesFileUpdate(ctx, filePathToLanguage)

	// 设置实现。
	i18n.RegisterImplement(get, closeWatch)
}

// 获取消息提示语。
func get(code errs.Code, language i18n.Language) (string, bool) {
	messageLocker.RLock()
	defer messageLocker.RUnlock()
	message, ok := messages[language]
	if ok {
		var oneClause string
		oneClause, ok = message[code]
		return oneClause, ok
	}
	return "", false
}

// 关闭对消息提示语文件的改动监听。
func closeWatch(ctx context.Context) {
	if !closedFlag.CompareAndSwap(0, 1) {
		return
	}

	log.Warn(ctx, "closing message file watch")
	log.ErrorIf(ctx, messageFileWatcher.Close(), "failed to close message file watcher")
}

// 解析 Properties 文件。
func parse(ctx context.Context, fileData []byte) map[errs.Code]string {
	// 切换换行符。
	fileData = bytes.ReplaceAll(fileData, []byte("\r\n"), []byte("\n"))

	// 去除空白字符。
	trimBlankFn := func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', '\v', '\f', '\a', '\b':
			return true
		}
		return false
	}
	fileData = bytes.TrimFunc(fileData, trimBlankFn)
	if len(fileData) == 0 {
		return make(map[errs.Code]string)
	}

	// 按行分割处理。
	entries := bytes.Split(fileData, []byte("\n"))
	result := make(map[errs.Code]string, len(entries))
	for i := range entries {
		// 去掉空白项。
		entry := entries[i]
		entry = bytes.TrimFunc(entry, trimBlankFn)
		if len(entry) == 0 {
			continue
		}

		// 去掉注释。
		if bytes.HasPrefix(entry, []byte{'#'}) {
			continue
		}

		// 转换数据。
		pair := bytes.Split(entry, []byte("="))
		if len(pair) != 2 {
			log.Warn(ctx, "message format is invalid", entry)
			continue
		}
		code, err := strconv.ParseInt(string(bytes.TrimFunc(pair[0], trimBlankFn)), 10, 64)
		if err != nil {
			log.Warn(ctx, "message format is invalid", entry)
			continue
		}

		result[errs.Code(code)] = string(bytes.TrimSpace(pair[1]))
	}

	return result
}

// 监听文件更新。
func watchPropertiesFileUpdate(ctx context.Context, filePathToLanguage map[string]i18n.Language) {
	for {
		select {
		case event, ok := <-messageFileWatcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Chmod:
				log.Warn(ctx, "message file chmod", event.Name)
			case fsnotify.Remove:
				log.Warn(ctx, "message file removed", event.Name)
			case fsnotify.Rename:
				log.Warn(ctx, "message file renamed", event.Name)
			case fsnotify.Write, fsnotify.Create:
				fileData, err := os.ReadFile(event.Name)
				if err != nil || len(fileData) <= 0 {
					log.ErrorIf(ctx, err, "failed to read message file")
					continue
				}
				data := parse(ctx, fileData)
				func() {
					messageLocker.Lock()
					defer messageLocker.Unlock()
					messages[filePathToLanguage[event.Name]] = data
				}()
				log.Warn(ctx, "update message file successfully", event.Name)
			default:
				log.Warn(ctx, "unknown watcher event", event.Op)
			}
		case err, ok := <-messageFileWatcher.Errors:
			if ok {
				log.Error(ctx, err, "watch properties file error")
			}
		}
	}
}
