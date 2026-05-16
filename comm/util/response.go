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

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	vt "github.com/go-playground/validator/v10"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/validator"
)

// Response HTTP 响应结构体。
type Response[T any] struct {
	// 响应码，大于 0 表示错误
	Code errs.Code `json:"code,omitempty" example:"0"`
	// 提示语
	Message string `json:"message,omitempty" example:"失败时的提示语"`
	// 响应数据
	Data *T `json:"data,omitempty"`
}

// ResponseCode 响应提示信息。没有配置对应提示语时，使用 [i18n.LanguageEnglish] 语言作为回退语言。
func ResponseCode(c *gin.Context, code errs.Code) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	language := ctxs.Language(c.Request.Context())
	message, ok := i18n.Get(code, i18n.Language(language))
	if !ok {
		message, _ = i18n.Get(code, i18n.LanguageEnglish)
	}
	c.JSON(http.StatusOK, &Response[any]{Message: message, Code: code})
}

// ResponseStatus 响应指定的响应码。
func ResponseStatus(c *gin.Context, status int) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	c.Status(status)
}

// ResponseStatusCode 响应指定的响应码。
func ResponseStatusCode(c *gin.Context, status int, code errs.Code) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	language := ctxs.Language(c.Request.Context())
	message, ok := i18n.Get(code, i18n.Language(language))
	if !ok {
		message, _ = i18n.Get(code, i18n.LanguageEnglish)
	}
	c.JSON(status, &Response[any]{Message: message, Code: code})
}

// ResponseStatusMsg 响应指定的响应码。
func ResponseStatusMsg(c *gin.Context, status int, msg string) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	c.String(status, msg)
}

// ResponseData 响应数据。
func ResponseData[T any](c *gin.Context, data *T) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	c.JSON(http.StatusOK, &Response[T]{Data: data})
}

// ResponseError 响应错误对象，并将错误信息保存进上下文。
func ResponseError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	rid := ctxs.RequestID(ctx)
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	c.Request = c.Request.WithContext(ctxs.WithError(ctx, err))
	language := ctxs.Language(ctx)

	// 根据错误对象类型，区别处理响应数据。
	var e vt.ValidationErrors
	if errors.As(err, &e) {
		c.JSON(http.StatusOK, &Response[any]{
			Message: validator.Translate(ctx, e),
			Code:    errs.ErrInvalidRequestParameters,
		})
		return
	}

	var e2 *errs.Error
	if errors.As(err, &e2) {
		if len(e2.Msg) > 0 {
			c.JSON(http.StatusOK, &Response[any]{Message: e2.Msg, Code: e2.Code})
		} else {
			message, ok := i18n.Get(e2.Code, i18n.Language(language))
			if !ok {
				message, ok = i18n.Get(e2.Code, i18n.LanguageEnglish)
			}
			if !ok {
				log.Warn(ctx, "cannot get i18n message", e2.Code)
			}
			c.JSON(http.StatusOK, &Response[any]{Message: message, Code: e2.Code})
		}
		return
	}

	log.Warn(ctx, "unknown error", err)
	message, ok := i18n.Get(errs.ErrUnknown, i18n.Language(language))
	if !ok {
		message, ok = i18n.Get(errs.ErrUnknown, i18n.LanguageEnglish)
	}
	c.JSON(http.StatusOK, &Response[any]{Message: message, Code: errs.ErrUnknown})
}

// ResponseAPIError 响应错误对象，并将错误信息保存进上下文。
func ResponseAPIError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	rid := ctxs.RequestID(ctx)
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	c.Request = c.Request.WithContext(ctxs.WithError(ctx, err))

	// 根据错误对象类型，区别处理响应数据。
	var e vt.ValidationErrors
	if errors.As(err, &e) {
		c.JSON(http.StatusBadRequest, &Response[any]{
			Message: validator.Translate(ctx, e),
			Code:    errs.ErrInvalidRequestParameters,
		})
		return
	}

	var e2 *errs.Error
	if errors.As(err, &e2) {
		httpCode := http.StatusInternalServerError
		if e2.Status > 0 {
			httpCode = e2.Status
		}
		if len(e2.Msg) > 0 {
			c.JSON(httpCode, &Response[any]{Message: e2.Msg, Code: e2.Code})
		} else {
			message, ok := i18n.Get(e2.Code, i18n.LanguageEnglish)
			if !ok {
				log.Warn(ctx, "cannot get i18n message", e2.Code)
			}
			c.JSON(httpCode, &Response[any]{Message: message, Code: e2.Code})
		}
		return
	}

	log.Warn(ctx, "unknown error", err)
	message, _ := i18n.Get(errs.ErrUnknown, i18n.LanguageEnglish)
	c.JSON(http.StatusInternalServerError, &Response[any]{Message: message, Code: errs.ErrUnknown})
}

// ResponseStream 响应流数据。
func ResponseStream(c *gin.Context, fileSize int64, fileName string, readCloser io.ReadCloser) {
	rid := ctxs.RequestID(c.Request.Context())
	c.Writer.Header().Set(consts.HTTPHeaderRequestID, rid)
	defer func() { log.ErrorIf(c.Request.Context(), readCloser.Close(), "failed to close io") }()
	c.DataFromReader(http.StatusOK, fileSize, "application/octet-stream", readCloser, map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	})
}
