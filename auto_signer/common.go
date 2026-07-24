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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

// ReadAndUnmarshal 序列化数据并关闭流。
func ReadAndUnmarshal[T any](reader io.ReadCloser) *T {
	defer CloseIO(reader)
	result := new(T)
	bs, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("read error: %v\n", err)
	}
	err = json.Unmarshal(bs, result)
	if err != nil {
		log.Printf("unmarshal error: %v\n", err)
	}
	return result
}

// ReadAndClose 读取后关闭流。
func ReadAndClose(reader io.ReadCloser) []byte {
	defer CloseIO(reader)
	bs, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("read error: %v\n", err)
	}
	return bs
}

// CloseIO 关闭流。
func CloseIO(closer io.Closer) {
	if closer != nil {
		err := closer.Close()
		if err != nil {
			log.Printf("failed to close io: %v\n", err)
		}
	}
}

// 格式化文件大小。
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 格式化时间。
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func FormatOneline(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", `\r\n`), "\n", `\n`)
}
