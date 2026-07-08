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

package errs

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-sql-driver/mysql"

	"gitee.com/ivfzhou/csms/comm/consts"
)

// Error 错误对象。
type Error struct {
	// HTTP 响应码
	Status int
	// 错误码，大于零表示有错误
	Code Code
	// 内部错误对象
	Err error
	// 错误描述
	Msg string
	// 创建错误对象的调用方位置
	Caller string
}

// New 创建错误对象。
func New(code Code) error {
	return &Error{Code: code, Caller: getCaller()}
}

// NewWithMsg 创建错误对象。
func NewWithMsg(code Code, msg string) error {
	return &Error{Code: code, Msg: msg, Caller: getCaller()}
}

// NewWithError 创建错误对象。
func NewWithError(code Code, err error) error {
	return &Error{Code: code, Err: err, Caller: getCaller()}
}

// NewWithStatus 创建错误对象。
func NewWithStatus(code Code, status int) error {
	return &Error{Code: code, Status: status, Caller: getCaller()}
}

// NewWithStatusMsg 创建错误对象。
func NewWithStatusMsg(code Code, status int, msg string) error {
	return &Error{Code: code, Status: status, Caller: getCaller(), Msg: msg}
}

// NewWithErrorStatusMsg 创建错误对象。
func NewWithErrorStatusMsg(code Code, status int, msg string, err error) error {
	return &Error{Code: code, Status: status, Err: err, Caller: getCaller(), Msg: msg}
}

// IsMySQLError 是否是 MySQL 的错误对象。
func IsMySQLError(err error, num int) bool {
	var e *mysql.MySQLError
	return errors.As(err, &e) && e.Number == uint16(num)
}

// String 获取错误对象的描述信息。
// 如果 err 是 [Error]，返回 Error.Msg 或者 Error.Err 的内容。否则返回 err.Error()。
func String(err error) string {
	var e *Error
	if errors.As(err, &e) && e != nil {
		if len(e.Msg) > 0 {
			return e.Msg
		}
		if e.Err != nil {
			return fmt.Sprintf("%v", e.Err)
		}
	}
	return fmt.Sprintf("%v", e.Err)
}

// Error 接口实现。
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	elems := make([]string, 0, 5)
	if e.Status != 0 {
		elems = append(elems, fmt.Sprintf("%d", e.Status))
	}
	if e.Code != 0 {
		elems = append(elems, fmt.Sprintf("%d", e.Code))
	}
	if len(e.Msg) > 0 {
		elems = append(elems, fmt.Sprintf("%s", e.Msg))
	}
	if len(e.Caller) > 0 {
		elems = append(elems, fmt.Sprintf("%s", e.Caller))
	}
	if e.Err != nil {
		elems = append(elems, fmt.Sprintf("%v", e.Err))
	}

	return fmt.Sprintf("{%s}", strings.Join(elems, " "))
}

// Unwrap 接口实现。
func (e *Error) Unwrap() error {
	return e.Err
}

func getCaller() string {
	_, file, line, _ := runtime.Caller(2)
	index := strings.LastIndex(file, consts.SystemName)
	if index != -1 {
		return fmt.Sprintf("%s:%d", strings.Trim(file[index+len(consts.SystemName):], "/"), line)
	}
	return ""
}
