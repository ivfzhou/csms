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

package util

import "unicode"

// IsAllHanCharacters 字符串全是汉字返回 true，否则返回 false。
func IsAllHanCharacters(s string) bool {
	for _, v := range []rune(s) {
		if !unicode.Is(unicode.Han, v) {
			return false
		}
	}
	return true
}

// IsVariableName 是否是合法变量名。
func IsVariableName(s string) bool {
	for i, v := range []rune(s) {
		if i > 0 {
			if v >= 'A' && v <= 'Z' || v >= 'a' && v <= 'z' || v >= '0' && v <= '9' || v == '_' {
				continue
			}
		} else {
			if v >= 'A' && v <= 'Z' || v >= 'a' && v <= 'z' || v == '_' {
				continue
			}
		}
		return false
	}
	return true
}
