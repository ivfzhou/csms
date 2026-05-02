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

package internal

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	writers         []WriteCloser
	writerLock      sync.RWMutex
	writerCloseFlag int32
)

// WriteCloser 日志打印者。
type WriteCloser interface {
	Write(string)
	IsColorful() bool
	Close(context.Context)
}

// RegisterWriter 添加日志输出。
func RegisterWriter(w WriteCloser) {
	ctx := ctxs.New()
	if w == nil {
		log.Warn(ctx, "writer cannot be nil")
		return
	}

	writerLock.Lock()
	defer writerLock.Unlock()

	if atomic.LoadInt32(&writerCloseFlag) > 0 {
		panic("cannot register writer, because writer is already closed")
	}

	if slices.Contains(writers, w) {
		log.Warn(ctx, "writer is already registered", w)
		return
	}

	writers = append(writers, w)
}

// SendBuilderToWriter 日志输出。
func SendBuilderToWriter(builder *Builder) {
	if builder == nil {
		ctx := ctxs.New()
		log.Warn(ctx, "builder cannot be nil")
		return
	}

	writerLock.RLock()
	defer writerLock.RUnlock()

	hasColorfulWriter := false
	onlyColorfulWriter := true
	for _, v := range writers {
		if v.IsColorful() {
			hasColorfulWriter = true
		} else {
			onlyColorfulWriter = false
		}
	}

	var logString, colorfulLogString string
	if onlyColorfulWriter {
		colorfulLogString = builder.BuildWithColorAndReclaim()
	} else if hasColorfulWriter {
		logString, colorfulLogString = builder.BuildTwiceAndReclaim()
	} else {
		logString = builder.BuildAndReclaim()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(writers))
	for _, v := range writers {
		go func(w WriteCloser) {
			defer wg.Done()
			defer func() {
				if p := recover(); p != nil {
					_, _ = fmt.Fprintf(os.Stderr, "writing log panicked %v\n", p)
				}
			}()

			if w.IsColorful() {
				w.Write(colorfulLogString)
			} else {
				w.Write(logString)
			}
		}(v)
	}
	wg.Wait()
}

// CloseWriter 关闭日志打印。
func CloseWriter(ctx context.Context) {
	writerLock.Lock()
	defer writerLock.Unlock()

	if !atomic.CompareAndSwapInt32(&writerCloseFlag, 0, 1) {
		log.Warn(ctx, "writer is already closed")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(writers))
	for _, v := range writers {
		go func(w WriteCloser) {
			defer wg.Done()
			defer func() {
				if p := recover(); p != nil {
					log.Error(ctx, "close writer panicked", p)
				}
			}()
			w.Close(ctx)
		}(v)
	}
	wg.Wait()
}
