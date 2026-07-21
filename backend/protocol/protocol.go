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

package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	vt "github.com/go-playground/validator/v10"

	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/comm/validator"
)

type Time time.Time

// Initialize 初始化。
func Initialize(ctx context.Context) {
	initAndroidWebProtocol()
	initAndroidOpenProtocol()
	initAndroidInternalProtocol()
	initAppWebProtocol()
	initAppleWebProtocol()
	initAppleOpenProtocol()
	initAppleInternalProtocol()
	initEventWebProtocol()
	initFileWebProtocol()
	initFileOpenProtocol()
	initFileInternalProtocol()
	initOpenWebProtocol()
	initTodoWebProtocol()
	initUserWebProtocol()
	initWindowsWebProtocol()
	initWindowsOpenProtocol()
	initWindowsInternalProtocol()
	initNoticeWebProtocol()

	// 注册校验标签。
	validator.RegisterValidation("han", func(fl vt.FieldLevel) bool {
		return util.IsAllHanCharacters(fmt.Sprint(fl.Field().Interface()))
	})
	validator.RegisterValidation("varname", func(fl vt.FieldLevel) bool {
		return util.IsVariableName(fmt.Sprint(fl.Field().Interface()))
	})
	validator.RegisterValidation("utf8", func(fl vt.FieldLevel) bool {
		return utf8.ValidString(fmt.Sprint(fl.Field().Interface()))
	})
	validator.RegisterValidation("hanasciiprint", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		for _, v := range []rune(val) {
			if unicode.Is(unicode.Han, v) || v >= ' ' && v <= '~' {
				continue
			}
			return false
		}
		return true
	})
	validator.RegisterValidation("capability", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := cc.AppleBundleIDCapabilities[val]
		return ok
	})
	validator.RegisterValidation("profiletype", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := model.AllAppleProfileTypes[val]
		return ok
	})
	validator.RegisterValidation("certificatetype", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := model.AllAppleCertificateTypes[val]
		return ok
	})
	validator.RegisterValidation("appleplatform", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := model.AllApplePlatformDescriptions[val]
		return ok
	})
	validator.RegisterValidation("hlktestsystem", func(fl vt.FieldLevel) bool {
		top := reflect.Indirect(fl.Top())
		if top.Kind() == reflect.Struct {
			signingType := top.FieldByName("SigningType")
			if signingType.IsValid() {
				if fmt.Sprint(signingType.Interface()) == fmt.Sprint(model.WHQLJobTypeOnlyWHQL) {
					return true
				}
			}
		}
		val := fmt.Sprint(fl.Field().Interface())
		_, ok := model.AllWHQLJobTestSystems[val]
		return ok
	})
	validator.RegisterValidation("hlktestconfig", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		var config cc.WHQLConfig
		err := json.Unmarshal([]byte(val), &config)
		return err == nil
	})
	validator.RegisterValidation("file", func(fl vt.FieldLevel) bool {
		val, ok := fl.Field().Interface().(multipart.FileHeader)
		if !ok || val.Size <= 0 || len(val.Filename) <= 0 {
			return false
		}
		return true
	})
	validator.RegisterValidation("utf8string", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		return utf8.ValidString(val) && !strings.Contains(val, "�")
	})
	validator.RegisterValidation("time", func(fl vt.FieldLevel) bool {
		val, _ := fl.Field().Interface().(time.Time)
		return val.UnixNano() > 0
	})
	validator.RegisterValidation("appledevice", func(fl vt.FieldLevel) bool {
		val := fmt.Sprint(fl.Field().Interface())
		return slices.Contains(model.AllAppleDeviceTypes, val)
	})

	validator.Init(ctx)
}

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	nt, err := time.Parse(cc.TimeFormat, string(bytes.Trim(b, "\"")))
	*t = Time(nt)
	return
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*t).Format(cc.TimeFormat) + `"`), nil
}
