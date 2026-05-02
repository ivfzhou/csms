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

package impl

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type fileWriter struct {
	file       *lumberjack.Logger
	updateLock sync.Mutex
}

// 新建实例。
func newFileWriter() internal.WriteCloser {
	w := &fileWriter{file: &lumberjack.Logger{
		Filename:   cfg.Get().Log().Name(),
		MaxSize:    cfg.Get().Log().FileMaximumSizeByMegabytes(),
		MaxAge:     cfg.Get().Log().FileMaximumAgeByDays(),
		MaxBackups: cfg.Get().Log().FileMaximumBackups(),
		LocalTime:  true,
	}}
	cfg.RegisterNotifier(func(configurer cfg.Configurer) {
		name := configurer.Log().Name()
		fileMaximumSizeByMegabytes := configurer.Log().FileMaximumSizeByMegabytes()
		fileMaximumAgeByDays := configurer.Log().FileMaximumAgeByDays()
		fileMaximumBackups := configurer.Log().FileMaximumBackups()

		w.updateLock.Lock()
		defer w.updateLock.Unlock()

		var newCreate bool
		if w.file.Filename != name {
			newCreate = true
		}
		if w.file.MaxSize != fileMaximumSizeByMegabytes {
			newCreate = true
		}
		if w.file.MaxAge != fileMaximumAgeByDays {
			newCreate = true
		}
		if w.file.MaxBackups != fileMaximumBackups {
			newCreate = true
		}

		if !newCreate {
			return
		}

		ctx := ctxs.New()
		log.Warn(ctx, "update file log writer")
		log.ErrorIf(ctx, w.file.Close(), "closing file writer failed")

		w.file = &lumberjack.Logger{
			Filename:   name,
			MaxSize:    fileMaximumSizeByMegabytes,
			MaxAge:     fileMaximumAgeByDays,
			MaxBackups: fileMaximumBackups,
			LocalTime:  true,
		}
	})
	return w
}

func (w *fileWriter) Write(s string) {
	written, err := io.Copy(w.file, strings.NewReader(s))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "writing log to file output failed: %v\n", err)
	}
	if written != int64(len(s)) {
		_, _ = fmt.Fprintf(os.Stderr, "writing log to file output failed: written %d, actual %d\n", written, len(s))
	}
}

func (w *fileWriter) IsColorful() bool {
	return false
}

func (w *fileWriter) Close(ctx context.Context) {
	w.updateLock.Lock()
	defer w.updateLock.Unlock()

	log.ErrorIf(ctx, w.file.Close(), "closing file writer failed")
}
