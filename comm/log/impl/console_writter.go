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

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type consoleWriter struct {
	colorful bool
}

// 新建实例。
func newConsoleWriter() internal.WriteCloser {
	w := &consoleWriter{colorful: cfg.Get().Log().ConsoleColorful()}
	cfg.RegisterNotifier(func(configurer cfg.Configurer) {
		colorful := configurer.Log().ConsoleColorful()
		if colorful != w.colorful {
			ctx := ctxs.New()
			log.Warn(ctx, "change console colorful", colorful)
			w.colorful = colorful
		}
	})
	return w
}

func (w *consoleWriter) Write(s string) {
	written, err := io.Copy(os.Stdout, strings.NewReader(s))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "writing log to standard output failed: %v\n", err)
	}
	if written != int64(len(s)) {
		_, _ = fmt.Fprintf(os.Stderr, "writing log to standard output failed: written %d, actual %d\n", written, len(s))
	}
}

func (w *consoleWriter) IsColorful() bool {
	return w.colorful
}

func (w *consoleWriter) Close(context.Context) {}
