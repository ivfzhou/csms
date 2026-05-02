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

package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTrans "github.com/go-playground/validator/v10/translations/en"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	// Validator 请求参数校验器。
	Validator *GinValidator

	translator            *ut.UniversalTranslator
	allTranslationMessage map[reflect.Type]TranslationMessage
	allTagValidator       map[Tag]func(fl validator.FieldLevel) bool
)

// GinValidator Gin 校验器。
type GinValidator struct {
	*validator.Validate
}

// StructFieldName 结构体字段名称。
type StructFieldName string

// Tag 校验规则。
type Tag string

// TranslationMessage 校验错误提示语载体。
type TranslationMessage map[StructFieldName]map[Tag]map[i18n.Language]string

// Init 初始化校验提示信息。若发生失败，会退出程序。
func Init(ctx context.Context) {
	Validator = &GinValidator{Validate: validator.New(validator.WithRequiredStructEnabled())}
	Validator.SetTagName("binding")

	// 初始化提示语翻译器。
	zhTran := zh.New()
	enTran := en.New()
	translator = ut.New(enTran, zhTran, enTran)
	err := translator.VerifyTranslations()
	if err != nil {
		log.Fatal(ctx, consts.ExitCodeInitialValidatorError, "failed to set validator translations", err)
	}

	zhTran2, _ := translator.GetTranslator(zhTran.Locale())
	if err = zhTrans.RegisterDefaultTranslations(Validator.Validate, zhTran2); err != nil {
		log.Fatal(ctx, consts.ExitCodeInitialValidatorError, "register default translation error", err)
	}
	enTran2, _ := translator.GetTranslator(enTran.Locale())
	if err = enTrans.RegisterDefaultTranslations(Validator.Validate, enTran2); err != nil {
		log.Fatal(ctx, consts.ExitCodeInitialValidatorError, "register default translation error", err)
	}

	// 转换错误提示语数据结构。
	messageMap := make(map[i18n.Language]map[Tag]map[reflect.Type]map[StructFieldName]string, len(allTranslationMessage))
	for structName, fields := range allTranslationMessage {
		for field, tags := range fields {
			for tag, languageMessages := range tags {
				for language, messages := range languageMessages {
					if messageMap[language] == nil {
						messageMap[language] = make(map[Tag]map[reflect.Type]map[StructFieldName]string)
					}
					if messageMap[language][tag] == nil {
						messageMap[language][tag] = make(map[reflect.Type]map[StructFieldName]string)
					}
					if messageMap[language][tag][structName] == nil {
						messageMap[language][tag][structName] = make(map[StructFieldName]string)
					}
					messageMap[language][tag][structName][field] = messages
				}
			}
		}
	}

	// 注册错误提示语进校验器。
	for language, tags := range messageMap {
		tran, found := translator.FindTranslator(string(language))
		if !found {
			continue
		}
		for tag, structs := range tags {
			err = Validator.RegisterTranslation(
				string(tag),
				tran,
				func(tran ut.Translator) error { return tran.Add(tag, "{0}", true) },
				func(tran ut.Translator, field validator.FieldError) string {
					pair := strings.Split(field.Namespace(), ".")
					for structName, fields := range structs {
						if pair[0] == structName.Name() {
							for fieldName, message := range fields {
								if pair[1] == string(fieldName) {
									return message
								}
								if strings.HasPrefix(pair[1], string(fieldName)) {
									// 匹配数组类型。
									matched, _ := regexp.MatchString(
										fmt.Sprintf("^%s\\[\\d+\\]$", fieldName), pair[1])
									if matched {
										return message
									}
								}
							}
							break
						}
					}
					message, err2 := tran.T(tag, field.Field())
					if err2 != nil {
						log.Error(ctx, "failed to translate", err2)
						// return field.Error()
						return ""
					}
					return message
				},
			)
			if err != nil {
				log.Fatal(ctx, consts.ExitCodeInitialValidatorError, "failed to register translator", err)
			}
		}
	}

	// 注册校验标签。
	for tag, fn := range allTagValidator {
		if err = Validator.RegisterValidation(string(tag), fn); err != nil {
			log.Fatal(ctx, consts.ExitCodeInitialValidatorError, "failed to register validator", err)
		}
	}
}

// RegisterValidation 注册校验标签。
func RegisterValidation(tag Tag, fn func(fl validator.FieldLevel) bool) {
	if fn == nil {
		return
	}
	if allTagValidator == nil {
		allTagValidator = make(map[Tag]func(fl validator.FieldLevel) bool)
	}
	allTagValidator[tag] = fn
}

// Translate 翻译错误提示语。
func Translate(ctx context.Context, err validator.ValidationErrors) string {
	if len(err) <= 0 {
		return ""
	}
	language := ctxs.Language(ctx)
	if len(language) <= 0 {
		language = string(i18n.LanguageEnglish)
	}
	tran, found := translator.GetTranslator(language)
	if !found {
		log.Warn(ctx, "cannot get translator", language)
	}

	// 只翻译第一个错误提示语。
	return err[0].Translate(tran)
}

// AddTranslationMessage 添加错误提示语。
func AddTranslationMessage(structType reflect.Type, message TranslationMessage) {
	if message == nil {
		return
	}
	if allTranslationMessage == nil {
		allTranslationMessage = make(map[reflect.Type]TranslationMessage, 50)
	}
	allTranslationMessage[structType] = message
}

// ValidateStruct 实现接口。校验结构体。
func (v *GinValidator) ValidateStruct(s any) error {
	return v.Validate.Struct(s)
}

// Engine 实现接口。返回校验器。
func (v *GinValidator) Engine() any {
	return v.Validate
}
