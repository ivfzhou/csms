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

package cfg

import "time"

// 运行环境。
const (
	EnvironmentLocal       = "local"
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

// Configurer 配置。
type Configurer interface {
	// Environment 运行环境。默认值 local
	Environment() string

	// TimeZone 时区，默认为 Asia/Shanghai
	TimeZone() string

	// GatewayHost Nginx 主机地址。默认 127.0.0.1
	GatewayHost() string

	// GatewayInternalServerPort 内部服务路由端口。默认 8000
	GatewayInternalServerPort() int

	// TLSInsecureSkipVerify 发送 HTTP 请求时，是否忽略 TLS 校验。默认值 false
	TLSInsecureSkipVerify() bool

	// Backend backend 服务配置
	Backend() BackendConfigurer

	// FastlaneProxy fastlane_proxy 服务配置
	FastlaneProxy() FastlaneProxyConfigurer

	// HLKManagerHost hlk_manager 宿主机配置
	HLKManagerHost() HLKManagerHostConfigurer

	// HLKManagerTest hlk_manager 测试机配置
	HLKManagerTest() HLKManagerTestConfigurer

	// HLKManagerController hlk_manager 控制机配置
	HLKManagerController() HLKManagerTestConfigurer

	// Apple 苹果账号相关配置
	Apple() AppleConfigurer

	// Log 日志配置
	Log() LogConfigurer

	// MySQL 数据库连接配置
	MySQL() MySQLConfigurer

	// RabbitMQ 消息队列连接配置
	RabbitMQ() RabbitMQConfigurer

	// Tusd Tusd 服务配置
	Tusd() TusdConfigurer

	// Redis Redis 连接配置
	Redis() RedisConfigurer

	// AppleAPI 苹果请求凭证
	AppleAPI() AppleAPIConfigurer

	// MicrosoftAPI 微软请求凭证账号信息
	MicrosoftAPI() MicrosoftAPIConfigurer

	// Swagger Swagger 配置
	Swagger() SwaggerConfigurer
}

// BackendConfigurer backend 服务配置。
type BackendConfigurer interface {
	// Port HTTP 监听端口。默认值 8090
	Port() int

	// InternalPort 内部服务 HTTP 监听端口。默认值 8091
	InternalPort() int

	// WriteTimeout HTTP 响应数据时限。默认值 30m
	WriteTimeout() time.Duration

	// ReadTimeout HTTP 读取数据时限。默认值 30s
	ReadTimeout() time.Duration

	// MinimumRequestInterval 两个页面请求之间最小请求间隔时间。默认值 800ms
	MinimumRequestInterval() time.Duration

	// OpenAPIMaximumExpirationDuration 请求凭证的最大有效时限。默认值 2h
	OpenAPIMaximumExpirationDuration() time.Duration

	// MaximumSendInterval 请求时间与到达时间的最大值。默认值 3m。
	MaximumSendInterval() time.Duration

	// UserAvatarMaximumSize 用户头像文件大小最大值。默认值 2mib
	UserAvatarMaximumSize() int

	// AppLogoMaximumSize 应用头像文件大小最大值。默认值 2mib
	AppLogoMaximumSize() int

	// FileUploadingMaximumInterval 文件上传时长最大值。默认值 24h
	FileUploadingMaximumInterval() time.Duration

	// WebSessionExpiration 页面登录会话最大存活时间。默认值 24h
	WebSessionExpiration() time.Duration

	// MaximumAttestationJobInterval 完成微软 Attestation 认证的最大时限。默认值 3 天
	MaximumAttestationJobInterval() time.Duration

	// MaximumWHQLJobInterval 完成微软 WHQL 认证的最大时限。默认值 30 天
	MaximumWHQLJobInterval() time.Duration

	// WaitingDelayTimeOfDispatchingTest 等待调度 HLK 测试延迟时间。默认 5 分钟。
	WaitingDelayTimeOfDispatchingTest() time.Duration
}

// FastlaneProxyConfigurer fastlane_proxy 服务配置。
type FastlaneProxyConfigurer interface {
	// Port HTTP 监听端口。默认值 8092
	Port() int
}

// HLKManagerHostConfigurer hlk_manager 宿主机配置。
type HLKManagerHostConfigurer interface {
	// Port HTTP 监听端口。默认值 8093
	Port() int
}

// HLKManagerTestConfigurer hlk_manager 测试机配置。
type HLKManagerTestConfigurer any

// HLKManagerControllerConfigurer hlk_manager 控制机配置。
type HLKManagerControllerConfigurer any

// AppleConfigurer 苹果账号相关配置。
type AppleConfigurer interface {
	// ApplyCertificateCSR 苹果证书的证书请求
	ApplyCertificateCSR() string

	// CertificatePrivateKey 苹果证书的证书对应的私钥
	CertificatePrivateKey() string

	// AppStoreTeamID 公司分发的团队 ID
	AppStoreTeamID() string

	// InHouseTeamID 企业内部分发的团队 ID
	InHouseTeamID() string

	// AccountName 苹果账号用户名
	AccountName() string

	// CommonProfile 通配符描述文件
	CommonProfile() string

	// CommonProfileID 通配符描述文件 ID
	CommonProfileID() string

	// CertificateIDOfCommonProfile 通配符描述文件的签名证书 ID
	CertificateIDOfCommonProfile() string
}

// LogConfigurer 日志配置。
type LogConfigurer interface {
	// Name 日志文件名
	Name() string

	// FileMaximumSizeByMegabytes 单个日志文件大小最大值。默认值 20
	FileMaximumSizeByMegabytes() int

	// FileMaximumBackups 日志文件数量最大值。默认值 50
	FileMaximumBackups() int

	// FileMaximumAgeByDays 日志文件最长存活时间。默认值 90
	FileMaximumAgeByDays() int

	// Level 日志打印等级。默认值 debug
	Level() string

	// SlowSQLThreshold 打印 SQL 慢查询语句时间阈值。默认值 100ms
	SlowSQLThreshold() time.Duration

	// ConsoleColorful 彩色日志打印。默认值 false
	ConsoleColorful() bool

	// ReportPath 日志上报地址。HLK 宿主机接收虚拟机的日志上报。
	ReportPath() string
}

// MySQLConfigurer 数据库连接配置。
type MySQLConfigurer interface {
	// Username 用户名。默认值 root
	Username() string

	// Password 连接密码。
	Password() string

	// Host 服务器主机。默认值 127.0.0.1
	Host() string

	// Port 服务端口号。默认值 3306
	Port() int

	// Database 数据库名。默认值 db_csms
	Database() string

	// Parameters 连接参数。
	Parameters() string

	// MaximumIdle 最大闲置连接数量。默认值 1
	MaximumIdle() int

	// MaximumOpen 最大连接数量。默认值 200
	MaximumOpen() int

	// MaximumLife 连接最长存活时间。默认值 24h
	MaximumLife() time.Duration

	// MaximumNumberOfPerSQLInsert 每次插入表记录的最大条数。默认 10
	MaximumNumberOfPerSQLInsert() int
}

// RabbitMQConfigurer RabbitMQ 消息队列连接配置。
type RabbitMQConfigurer interface {
	// Username 用户名。默认值 guest
	Username() string

	// Password 用户密码。默认值 guest
	Password() string

	// Host 服务器主机。默认值 127.0.0.1
	Host() string

	// Port 服务器端口号。默认值 5672
	Port() int

	// VirtualHost 使用的 vhost。默认值 csms
	VirtualHost() string

	// PrefetchCount Qos 配置。
	PrefetchCount() int

	// WindowsOVSigningJobQueue 队列名。默认值 queue_windows_ov
	WindowsOVSigningJobQueue() string

	// WindowsEVSigningJobQueuePrefix 队列名前缀。默认值 queue_windows_ev_
	WindowsEVSigningJobQueuePrefix() string

	// AndroidSigningJobQueue 队列名。默认值 queue_android
	AndroidSigningJobQueue() string

	// AppleSigningJobQueue 队列名。默认值 queue_apple
	AppleSigningJobQueue() string
}

// TusdConfigurer Tusd 服务配置。
type TusdConfigurer interface {
	// Host 服务器主机。默认值 127.0.0.1
	Host() string

	// Port 端口。默认值 8080
	Port() int
}

// RedisConfigurer Redis 连接配置。
type RedisConfigurer interface {
	// Username 用户名
	Username() string

	// Password 密码
	Password() string

	// Host 主机。默认值 127.0.0.1
	Host() string

	// Port 端口。默认值 6379
	Port() int

	// Database 数据库
	Database() int
}

// AppleAPIConfigurer 苹果请求凭证。
type AppleAPIConfigurer interface {
	// IssuerID 账号 ID
	IssuerID() string

	// KeyID 密钥 ID
	KeyID() string

	// Secret 密钥
	Secret() string
}

// MicrosoftAPIConfigurer 微软请求凭证账号信息。
type MicrosoftAPIConfigurer interface {
	// GrantType 账号类型
	GrantType() string

	// ClientID 账号 ID
	ClientID() string

	// ClientSecret 密钥
	ClientSecret() string

	// Resource 请求资源
	Resource() string

	// TenantID 租户 ID
	TenantID() string
}

// SwaggerConfigurer Swagger 配置。
type SwaggerConfigurer interface {
	// Host Swagger API 请求地址，默认为 127.0.0.1
	Host() string

	// Schema Swagger API 请求协议，默认 http。
	Schema() string

	// Port Swagger API 请求端口，默认 80。
	Port() int

	// Version Swagger API 版本号，默认 1.0
	Version() string

	// BasePath 基础请求路径
	BasePath() string
}
