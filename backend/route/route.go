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

package route

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gitee.com/ivfzhou/csms/backend/filter"
	cf "gitee.com/ivfzhou/csms/comm/filter"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/comm/validator"
)

// Initialize 初始化 HTTP 路由。
func Initialize(ctx context.Context) *gin.Engine {
	// 正式环境下设置 Gin 为发布模式，否则开启彩色 Gin 日志。
	if util.IsProductionEnvironment() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 设置 Gin 日志输出位置，和请求参数校验器。
	gin.DefaultWriter = log.GetGinLogger()
	binding.Validator = validator.Validator

	engine := gin.New()

	// 本地环境打印 Gin 日志。
	if util.IsLocalEnvironment() {
		engine.Use(gin.Logger())
		gin.ForceConsoleColor()
	}

	// 初始化 filter 包。
	filter.InitAntiShakeScript(ctx)
	filter.InitialAPIAuthenticateLimitScript(ctx)

	// 初始化 HTTP 路由。
	engine.Use(cf.Recover, cf.ExitFilter(ctx), cf.LogFormatterFilter)
	initWebRoute(engine.Group("/web", cf.LanguageFilter, filter.DateCheckFilter))
	initApiRoute(engine.Group("/api"))

	// Swagger 文档。
	if !util.IsProductionEnvironment() {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return engine
}

// InitializeInternal 初始化内部网络 HTTP 路由。
func InitializeInternal(ctx context.Context) *gin.Engine {
	engine := gin.New()

	// 本地环境打印 Gin 日志。
	if util.IsLocalEnvironment() {
		engine.Use(gin.Logger())
	}

	// 初始化 HTTP 路由。
	engine.Use(cf.Recover, cf.ExitFilter(ctx), cf.LogFormatterFilter)
	initInternalRoute(engine.Group("/internal"))

	return engine
}
