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
	"strconv"
	"strings"
)

var (
	// ErrIntListCannotBeNil 不能将数据库数据注入到 nil IntList 变量中。
	ErrIntListCannotBeNil = errors.New("cannot scan nil IntList")
	// ErrCannotScanToIntList 数据库数据类型不能给 IntList 注入数据。
	ErrCannotScanToIntList = errors.New("cannot scan to IntList")
)

// IntList 代表数据库字符串类型的字段。按逗号分割的数组列表。
type IntList []int

// Scan 数据库数据转换成 Go 变量中。
func (i *IntList) Scan(value any) error {
	if i == nil {
		return ErrIntListCannotBeNil
	}
	var arr []string
	switch v := value.(type) {
	case []byte:
		arr = strings.Split(string(v), ",")
	case string:
		arr = strings.Split(v, ",")
	default:
		return fmt.Errorf("type [%T] %w", value, ErrCannotScanToIntList)
	}
	*i = make([]int, 0, len(arr))
	for _, v := range arr {
		num, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		*i = append(*i, num)
	}
	return nil
}

// Value Go 变量数据转成数据库数据。
func (i IntList) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	arr := make([]string, 0, len(i))
	for _, v := range i {
		arr = append(arr, strconv.Itoa(v))
	}
	return strings.Join(arr, ","), nil
}
