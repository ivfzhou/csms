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

package ctxs

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
)

const (
	ctxKeyUser       ctxKey = "ctxKeyUser"
	ctxKeyAPIAccount ctxKey = "ctxKeyAPIAccount"
	ctxKeyApp        ctxKey = "ctxKeyApp"
	ctxKeyError      ctxKey = "ctxKeyError"
	ctxKeyRequestID  ctxKey = "ctxKeyRequestID"
	ctxKeyRequestIP  ctxKey = "ctxKeyRequestIP"
	ctxKeyRequestURI ctxKey = "ctxKeyRequestURI"
	ctxKeyDBClient   ctxKey = "ctxKeyDBClient"
	ctxKeyLanguage   ctxKey = "ctxKeyLanguage"
)

var (
	ErrContextCannotBeNil      = errors.New("context cannot be nil")
	ErrDBClientCannotBeNil     = errors.New("database client cannot be nil")
	ErrContextNotInTransaction = errors.New("transaction state is invalid")
)

type ctxKey string

type transactionInfo struct {
	isOpen bool
	tx     *query.QueryTx
}

// User 获取上下文中的用户信息。
func User(ctx context.Context) *model.User {
	if ctx == nil {
		return nil
	}
	userInfo, _ := ctx.Value(ctxKeyUser).(*model.User)
	return userInfo
}

// WithUser 向上下文中添加用户信息。
func WithUser(ctx context.Context, userInfo *model.User) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if userInfo == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyUser, userInfo)
}

// APIAccount 获取上下文中的请求凭证账号信息。
func APIAccount(ctx context.Context) *model.APIAccount {
	if ctx == nil {
		return nil
	}
	apiAccountInfo, _ := ctx.Value(ctxKeyAPIAccount).(*model.APIAccount)
	return apiAccountInfo
}

// WithAPIAccount 向上下文中添加请求凭证。
func WithAPIAccount(ctx context.Context, apiAccountInfo *model.APIAccount) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if apiAccountInfo == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyAPIAccount, apiAccountInfo)
}

// App 获取上下文中的应用信息。
func App(ctx context.Context) *model.App {
	if ctx == nil {
		return nil
	}
	appInfo, _ := ctx.Value(ctxKeyApp).(*model.App)
	return appInfo
}

// WithApp 向上下文中添加应用信息。
func WithApp(ctx context.Context, appInfo *model.App) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if appInfo == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyApp, appInfo)
}

// Error 获取上下文中的错误消息。
func Error(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	err, _ := ctx.Value(ctxKeyError).(error)
	return err
}

// WithError 向上下文中添加错误消息。
func WithError(ctx context.Context, err error) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if err == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyError, err)
}

// RequestID 获取上下文中的链路 ID。
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(ctxKeyRequestID).(string)
	return requestID
}

// WithRequestID 向上下文附带链路 ID。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(requestID) <= 0 {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyRequestID, requestID)
}

// RequestIP 获取上下文中的请求 IP。
func RequestIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestIP, _ := ctx.Value(ctxKeyRequestIP).(string)
	return requestIP
}

// WithRequestIP 向上下文附带链路请求 IP。
func WithRequestIP(ctx context.Context, requestIP string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(requestIP) <= 0 {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyRequestIP, requestIP)
}

// RequestURI 获取上下文中的请求 URI。
func RequestURI(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestURI, _ := ctx.Value(ctxKeyRequestURI).(string)
	return requestURI
}

// WithRequestURI 向上下文附带请求 URI。
func WithRequestURI(ctx context.Context, requestURI string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(requestURI) <= 0 {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyRequestURI, requestURI)
}

// Language 获取上下文的语言码。
func Language(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	language, _ := ctx.Value(ctxKeyLanguage).(string)
	return language
}

// WithLanguage 向上下文添加语言码。
func WithLanguage(ctx context.Context, language string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if language == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyLanguage, language)
}

// InTransaction 获取上下文中的事务标记。
func InTransaction(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	info, _ := ctx.Value(ctxKeyDBClient).(*transactionInfo)
	return info != nil && info.isOpen
}

// WithInTransaction 向上下文添加需要事务标记。
func WithInTransaction(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKeyDBClient, &transactionInfo{true, nil})
}

// DBClient 获取上下文中的数据库事务对象。
func DBClient(ctx context.Context) *query.QueryTx {
	if ctx == nil {
		return nil
	}
	info, _ := ctx.Value(ctxKeyDBClient).(*transactionInfo)
	if info != nil && info.isOpen {
		return info.tx
	}
	return nil
}

// SetDBClient 向上下文中添加事务。
func SetDBClient(ctx context.Context, tx *query.QueryTx) {
	if ctx == nil {
		panic(ErrContextCannotBeNil)
	}
	if tx == nil {
		panic(ErrDBClientCannotBeNil)
	}
	info, _ := ctx.Value(ctxKeyDBClient).(*transactionInfo)
	if info == nil || !info.isOpen || info.tx != nil {
		panic(ErrContextNotInTransaction)
	}

	info.tx = tx
}

// New 新建上下文对象。
func New() context.Context {
	ctx := context.Background()
	ctx = WithRequestID(ctx, strings.ReplaceAll(uuid.NewString(), "-", ""))
	return ctx
}

// Clone 克隆上下文对象，复制链路 ID，请求 IP、请求 URI、应用、用户、语言和请求凭证账号信息。
func Clone(ctx context.Context) context.Context {
	newCtx := New()
	newCtx = WithRequestID(newCtx, RequestID(ctx))
	newCtx = WithRequestIP(newCtx, RequestIP(ctx))
	newCtx = WithRequestURI(newCtx, RequestURI(ctx))
	newCtx = WithApp(newCtx, App(ctx))
	newCtx = WithUser(newCtx, User(ctx))
	newCtx = WithLanguage(newCtx, Language(ctx))
	newCtx = WithAPIAccount(newCtx, APIAccount(ctx))
	return newCtx
}
