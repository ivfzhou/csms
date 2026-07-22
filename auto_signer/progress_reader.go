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
	"fmt"
	"io"
	"strings"
	"time"
)

// ProgressReader 带进度报告的 io.Reader 包装器。
type ProgressReader struct {
	reader io.Reader
	// 总字节数
	total int64
	// 已读字节数
	read int64
	// 上次打印时的已读字节数
	lastRead int64
	// 开始时间
	startTime time.Time
	// 上次回调时间
	lastTime time.Time
	// 最小回调间隔
	minInterval time.Duration
	// 进度回调
	callback func(string)
	// 平均速度
	speed int64
}

// NewProgressReader 创建进度读取器。
func NewProgressReader(reader io.Reader, total int64, interval time.Duration, cb func(string)) *ProgressReader {
	now := time.Now()
	return &ProgressReader{
		reader:      reader,
		total:       total,
		startTime:   now,
		lastTime:    now,
		minInterval: interval,
		callback:    cb,
	}
}

// Read 实现 io.Reader 接口。
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)
	pr.report(false)
	return n, err
}

// Finish 完成进度报告，触发最后一次回调。
func (pr *ProgressReader) Finish() {
	pr.report(true)
}

func (pr *ProgressReader) GetSpeed() int64 {
	return pr.speed
}

// report 触发进度回调。
func (pr *ProgressReader) report(finished bool) {
	if !finished && time.Since(pr.lastTime) < pr.minInterval {
		return
	}
	if pr.callback == nil {
		return
	}

	// 瞬时速度（基于上次回调以来的增量）。
	instantBytes := pr.read - pr.lastRead
	instantElapsed := time.Since(pr.lastTime).Seconds()
	if instantElapsed < 0.001 {
		instantElapsed = 0.001
	}
	speed := float64(instantBytes) / instantElapsed

	// 百分比。
	percent := float64(0)
	if pr.total > 0 {
		percent = float64(pr.read) / float64(pr.total) * 100
	}

	// 进度条。
	barWidth := 20
	filled := int(percent / 100 * float64(barWidth))
	bar := strings.Repeat("=", filled)
	if filled < barWidth {
		bar += ">"
		bar += strings.Repeat(" ", barWidth-filled-1)
	}

	// 平均速度。
	elapsed := time.Since(pr.startTime).Seconds()
	if elapsed < 0.01 {
		elapsed = 0.01 // 避免除零
	}
	avgSpeed := float64(pr.read) / elapsed

	// 剩余时间估算。
	var eta string
	if avgSpeed > 0 && pr.total > 0 {
		remaining := float64(pr.total-pr.read) / avgSpeed
		eta = FormatDuration(time.Duration(remaining) * time.Second)
	} else {
		eta = "--"
	}

	// 更新字段。
	pr.lastRead = pr.read
	pr.lastTime = time.Now()
	pr.speed = int64(avgSpeed)

	pr.callback(fmt.Sprintf("写入文件 %.1f%% [%s] (%s/%s)，速度：%s/s 平均速度：%s/s 剩余：%s",
		percent,
		bar,
		FormatSize(pr.read),
		FormatSize(pr.total),
		FormatSize(int64(speed)),
		FormatSize(int64(avgSpeed)),
		eta,
	))
}
