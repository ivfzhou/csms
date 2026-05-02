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
	"io"
	"log"

	cl "gitee.com/ivfzhou/csms/comm/log"
)

// ReadAndUnmarshal 序列化数据并关闭流。
func ReadAndUnmarshal[T any](reader io.ReadCloser) *T {
	defer CloseIO(reader)
	result := new(T)
	bs, err := io.ReadAll(reader)
	if err != nil {
		log.Println(cl.LevelError, "read error", err)
	}
	err = json.Unmarshal(bs, result)
	if err != nil {
		log.Println(cl.LevelError, "unmarshal error", err)
	}
	return result
}

// ReadAndClose 读取后关闭流。
func ReadAndClose(reader io.ReadCloser) []byte {
	defer CloseIO(reader)
	bs, err := io.ReadAll(reader)
	if err != nil {
		log.Println(cl.LevelError, "read error", err)
	}
	return bs
}

// CloseIO 关闭流。
func CloseIO(closer io.Closer) {
	if closer != nil {
		err := closer.Close()
		if err != nil {
			log.Println(cl.LevelError, "failed to close io", err)
		}
	}
}
