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

package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrStringListCannotBeNil 不能将数据库数据注入到 nil StringList 变量中。
	ErrStringListCannotBeNil = errors.New("cannot scan nil StringList")
	// ErrCannotScanToStringList 数据库数据类型不能给 StringList 注入数据。
	ErrCannotScanToStringList = errors.New("cannot scan to StringList")
)

// StringList 代表数据库字符串类型的字段。按逗号分割的字符串列表。
type StringList []string

// Scan 实现接口。
func (s *StringList) Scan(value any) error {
	if s == nil {
		return ErrStringListCannotBeNil
	}
	var arr []string
	switch v := value.(type) {
	case []byte:
		arr = strings.Split(string(v), ",")
	case string:
		arr = strings.Split(v, ",")
	default:
		return fmt.Errorf("type %T %w", value, ErrCannotScanToStringList)
	}
	*s = arr
	return nil
}

// Value 实现接口。
func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return strings.Join(s, ","), nil
}
