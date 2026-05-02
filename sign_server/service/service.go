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

package service

import (
	"fmt"
	"strings"
	"time"

	"gitee.com/ivfzhou/csms/comm/log"
)

func formatJobLog(level log.Level, format string, args ...any) string {
	str := fmt.Sprintf(format, args...)
	str = strings.Trim(str, `\r\n`)
	str = strings.Trim(str, `\n`)
	return fmt.Sprintf("%s %s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), level.String(), str)
}
