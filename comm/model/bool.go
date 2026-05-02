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
)

var (
	// ErrBoolCannotBeNil 不能将数据库数据注入到 nil Bool 变量中。
	ErrBoolCannotBeNil = errors.New("cannot scan nil Bool")
	// ErrCannotScanToBool 数据库数据类型不能给 Bool 注入数据。
	ErrCannotScanToBool = errors.New("cannot scan to Bool")
)

// Bool 数据库字段开关。
type Bool bool

// Scan 数据库数据转换成 Go 变量中。
func (b *Bool) Scan(value any) error {
	if b == nil {
		return ErrBoolCannotBeNil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) <= 0 {
			*b = false
		} else if v[0] > 0 {
			*b = true
		} else {
			*b = false
		}
	default:
		return fmt.Errorf("type [%T] %w", value, ErrCannotScanToBool)
	}
	return nil
}

// Value Go 变量数据转成数据库数据。
func (b Bool) Value() (driver.Value, error) {
	return bool(b), nil
}
