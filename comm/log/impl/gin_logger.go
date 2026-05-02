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
	"bytes"
	"context"
	"io"

	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type ginLoggerImpl struct{}

// 新建实例。
func newGinLogger() io.Writer {
	return &ginLoggerImpl{}
}

func (l *ginLoggerImpl) Write(p []byte) (n int, err error) {
	internal.SendBuilderToWriter(internal.CreateBuilder(context.TODO(), string(bytes.Trim(p, "\n"))).SetLibrary("gin"))
	return len(p), nil
}
