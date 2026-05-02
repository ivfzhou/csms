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

// 默认值。
const (
	DefaultEnvironment               = "local"
	DefaultTimeZone                  = "Asia/Shanghai"
	DefaultGatewayHost               = "127.0.0.1"
	DefaultGatewayInternalServerPort = 8000

	DefaultBackendPort                              = 8090
	DefaultBackendInternalPort                      = 8091
	DefaultBackendWriteTimeout                      = 30 * time.Minute
	DefaultBackendReadTimeout                       = 30 * time.Second
	DefaultBackendMinimumRequestInterval            = 800 * time.Millisecond
	DefaultBackendOpenAPIMaximumExpirationDuration  = 2 * time.Hour
	DefaultBackendMaximumSendInterval               = 3 * time.Minute
	DefaultBackendUserAvatarMaximumSize             = 2 * 1024 * 1024
	DefaultBackendAppLogoMaximumSize                = 2 * 1024 * 1024
	DefaultBackendFileUploadingMaximumInterval      = 24 * time.Hour
	DefaultBackendWebSessionExpiration              = 24 * time.Hour
	DefaultBackendMaximumAttestationJobInterval     = 3 * 24 * time.Hour
	DefaultBackendMaximumWHQLJobInterval            = 30 * 24 * time.Hour
	DefaultBackendWaitingDelayTimeOfDispatchingTest = 5 * time.Minute

	DefaultFastlaneProxyPort = 8092

	DefaultHLKManagerHostPort = 8093

	DefaultLogFileMaximumSizeByMegabytes = 20
	DefaultLogFileMaximumBackups         = 50
	DefaultLogFileMaximumAgeByDays       = 90
	DefaultLogLevel                      = "debug"
	DefaultLogSlowSQLThreshold           = 100 * time.Millisecond

	DefaultMySQLUsername                    = "root"
	DefaultMySQLPassword                    = "root"
	DefaultMySQLHost                        = "127.0.0.1"
	DefaultMySQLPort                        = 3306
	DefaultMySQLDatabase                    = "db_csms"
	DefaultMySQLMaximumIdle                 = 1
	DefaultMySQLMaximumOpen                 = 200
	DefaultMySQLMaximumLife                 = 24 * time.Hour
	DefaultMySQLMaximumNumberOfPerSQLInsert = 10

	DefaultRabbitMQUsername                       = "guest"
	DefaultRabbitMQPassword                       = "guest"
	DefaultRabbitMQHost                           = "127.0.0.1"
	DefaultRabbitMQPort                           = 5672
	DefaultRabbitMQVirtualHost                    = "csms"
	DefaultRabbitMQWindowsOVSigningJobQueue       = "queue_windows_ov"
	DefaultRabbitMQWindowsEVSigningJobQueuePrefix = "queue_windows_ev_"
	DefaultRabbitMQAndroidSigningJobQueue         = "queue_android"
	DefaultRabbitMQAppleSigningJobQueue           = "queue_apple"

	DefaultTusdHost = "127.0.0.1"
	DefaultTusdPort = 8080

	DefaultRedisHost = "127.0.0.1"
	DefaultRedisPort = 6379

	DefaultSwaggerHost    = "127.0.0.1"
	DefaultSwaggerPort    = 80
	DefaultSwaggerSchema  = "http"
	DefaultSwaggerVersion = "1.0"
)
