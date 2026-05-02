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

package ini

import (
	"time"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
)

// Configuration ini 格式配置实现。
type Configuration struct {
	EnvironmentValue                  string                            `ini:"environment"`
	TimeZoneValue                     string                            `ini:"timeZone"`
	GatewayHostValue                  string                            `ini:"gatewayHost"`
	GatewayInternalServerPortValue    int                               `ini:"gatewayInternalServerPort"`
	TLSInsecureSkipVerifyValue        bool                              `ini:"tlsInsecureSkipVerify"`
	BackendConfiguration              BackendConfiguration              `ini:"backend"`
	FastlaneProxyConfiguration        FastlaneProxyConfiguration        `ini:"fastlaneProxy"`
	HLKManagerHostConfiguration       HLKManagerHostConfiguration       `ini:"hlkManagerHost"`
	HLKManagerTestConfiguration       HLKManagerTestConfiguration       `ini:"hlkManagerTest"`
	HLKManagerControllerConfiguration HLKManagerControllerConfiguration `ini:"hlkManagerController"`
	AppleConfiguration                AppleConfiguration                `ini:"apple"`
	LogConfiguration                  LogConfiguration                  `ini:"log"`
	MySQLConfiguration                MySQLConfiguration                `ini:"mysql"`
	RabbitMQConfiguration             RabbitMQConfiguration             `ini:"rabbitmq"`
	TusdConfiguration                 TusdConfiguration                 `ini:"tusd"`
	RedisConfiguration                RedisConfiguration                `ini:"redis"`
	AppleAPIConfiguration             AppleAPIConfiguration             `ini:"appleApi"`
	MicrosoftAPIConfiguration         MicrosoftAPIConfiguration         `ini:"microsoftApi"`
	SwaggerConfiguration              SwaggerConfiguration              `ini:"swagger"`
}

// BackendConfiguration ini 格式配置实现。
type BackendConfiguration struct {
	PortValue                              int           `ini:"port"`
	InternalPortValue                      int           `ini:"internalPort"`
	WriteTimeoutValue                      time.Duration `ini:"writeTimeout"`
	ReadTimeoutValue                       time.Duration `ini:"readTimeout"`
	MinimumRequestIntervalValue            time.Duration `ini:"minimumRequestInterval"`
	OpenAPIMaximumExpirationDurationValue  time.Duration `ini:"openApiMaximumExpirationDuration"`
	MaximumSendIntervalValue               time.Duration `ini:"maximumSendInterval"`
	UserAvatarMaximumSizeValue             int           `ini:"userAvatarMaximumSize"`
	AppLogoMaximumSizeValue                int           `ini:"appLogoMaximumSize"`
	FileUploadingMaximumIntervalValue      time.Duration `ini:"fileUploadingMaximumInterval"`
	WebSessionExpirationValue              time.Duration `ini:"webSessionExpiration"`
	MaximumAttestationJobIntervalValue     time.Duration `ini:"maximumAttestationJobInterval"`
	MaximumWHQLJobIntervalValue            time.Duration `ini:"maximumWHQLJobInterval"`
	WaitingDelayTimeOfDispatchingTestValue time.Duration `ini:"waitingDelayTimeOfDispatchingTest"`
}

// FastlaneProxyConfiguration ini 格式配置实现。
type FastlaneProxyConfiguration struct {
	PortValue int `ini:"port"`
}

// HLKManagerHostConfiguration ini 格式配置实现。
type HLKManagerHostConfiguration struct {
	PortValue int `ini:"port"`
}

// HLKManagerTestConfiguration ini 格式配置实现。
type HLKManagerTestConfiguration struct{}

// HLKManagerControllerConfiguration ini 格式配置实现。
type HLKManagerControllerConfiguration struct{}

// AppleConfiguration ini 格式配置实现。
type AppleConfiguration struct {
	ApplyCertificateCSRValue          string `ini:"applyCertificateCsr"`
	CertificatePrivateKeyValue        string `ini:"appleCertificatePrivateKey"`
	AppStoreTeamIDValue               string `ini:"appStoreTeamId"`
	InHouseTeamIDValue                string `ini:"appleInHouseTeamId"`
	AccountNameValue                  string `ini:"appleAccountName"`
	CommonProfileValue                string `ini:"appleCommonProfile"`
	CommonProfileIDValue              string `ini:"appleCommonProfileId"`
	CertificateIDOfCommonProfileValue string `ini:"appleCertificateIdOfCommonProfile"`
}

// LogConfiguration ini 格式配置实现。
type LogConfiguration struct {
	NameValue                       string        `ini:"name"`
	FileMaximumSizeByMegabytesValue int           `ini:"fileMaximumSizeByMegabytes"`
	FileMaximumBackupsValue         int           `ini:"fileMaximumBackups"`
	FileMaximumAgeByDaysValue       int           `ini:"fileMaximumAgeByDays"`
	LevelValue                      string        `ini:"level"`
	SlowSQLThresholdValue           time.Duration `ini:"slowSqlThreshold"`
	ConsoleColorfulValue            bool          `ini:"consoleColorful"`
	ReportPathValue                 string        `ini:"reportPath"`
}

// MySQLConfiguration ini 格式配置实现。
type MySQLConfiguration struct {
	UsernameValue                    string        `ini:"username"`
	PasswordValue                    string        `ini:"password"`
	HostValue                        string        `ini:"host"`
	PortValue                        int           `ini:"port"`
	DatabaseValue                    string        `ini:"database"`
	ParametersValue                  string        `ini:"parameters"`
	MaximumIdleValue                 int           `ini:"maximumIdle"`
	MaximumOpenValue                 int           `ini:"maximumOpen"`
	MaximumLifeValue                 time.Duration `ini:"maximumLife"`
	MaximumNumberOfPerSQLInsertValue int           `ini:"maximumNumberOfPerSQLInsert"`
}

// RabbitMQConfiguration ini 格式配置实现。
type RabbitMQConfiguration struct {
	UsernameValue                       string `ini:"username"`
	PasswordValue                       string `ini:"password"`
	HostValue                           string `ini:"host"`
	PortValue                           int    `ini:"port"`
	VirtualHostValue                    string `ini:"virtualHost"`
	PrefetchCountValue                  int    `ini:"prefetchCount"`
	WindowsOVSigningJobQueueValue       string `ini:"windowsOVSigningJobQueue"`
	WindowsEVSigningJobQueuePrefixValue string `ini:"windowsEVSigningJobQueuePrefix"`
	AndroidSigningJobQueueValue         string `ini:"androidSigningJobQueue"`
	AppleSigningJobQueueValue           string `ini:"appleSigningJobQueue"`
}

// TusdConfiguration ini 格式配置实现。
type TusdConfiguration struct {
	HostValue string `ini:"host"`
	PortValue int    `ini:"port"`
}

// RedisConfiguration ini 格式配置实现。
type RedisConfiguration struct {
	UsernameValue string `ini:"username"`
	PasswordValue string `ini:"password"`
	HostValue     string `ini:"host"`
	PortValue     int    `ini:"port"`
	DatabaseValue int    `ini:"database"`
}

// AppleAPIConfiguration ini 格式配置实现。
type AppleAPIConfiguration struct {
	IssuerIDValue string `ini:"issuerId"`
	KeyIDValue    string `ini:"keyId"`
	SecretValue   string `ini:"secret"`
}

// MicrosoftAPIConfiguration ini 格式配置实现。
type MicrosoftAPIConfiguration struct {
	GrantTypeValue    string `ini:"grantType"`
	ClientIDValue     string `ini:"clientId"`
	ClientSecretValue string `ini:"clientSecret"`
	ResourceValue     string `ini:"resource"`
	TenantIDValue     string `ini:"tenantId"`
}

// SwaggerConfiguration ini 格式配置实现。
type SwaggerConfiguration struct {
	HostValue     string `ini:"host"`
	SchemaValue   string `ini:"schema"`
	VersionValue  string `ini:"version"`
	PortValue     int    `ini:"port"`
	BasePathValue string `ini:"basePath"`
}

// 新建实例。
func newImpl() *Configuration {
	return &Configuration{
		EnvironmentValue:               cfg.DefaultEnvironment,
		TimeZoneValue:                  cfg.DefaultTimeZone,
		GatewayHostValue:               cfg.DefaultGatewayHost,
		GatewayInternalServerPortValue: cfg.DefaultGatewayInternalServerPort,
		BackendConfiguration: BackendConfiguration{
			PortValue:                              cfg.DefaultBackendPort,
			InternalPortValue:                      cfg.DefaultBackendInternalPort,
			WriteTimeoutValue:                      cfg.DefaultBackendWriteTimeout,
			ReadTimeoutValue:                       cfg.DefaultBackendReadTimeout,
			MinimumRequestIntervalValue:            cfg.DefaultBackendMinimumRequestInterval,
			OpenAPIMaximumExpirationDurationValue:  cfg.DefaultBackendOpenAPIMaximumExpirationDuration,
			MaximumSendIntervalValue:               cfg.DefaultBackendMaximumSendInterval,
			UserAvatarMaximumSizeValue:             cfg.DefaultBackendUserAvatarMaximumSize,
			AppLogoMaximumSizeValue:                cfg.DefaultBackendAppLogoMaximumSize,
			FileUploadingMaximumIntervalValue:      cfg.DefaultBackendFileUploadingMaximumInterval,
			WebSessionExpirationValue:              cfg.DefaultBackendWebSessionExpiration,
			MaximumAttestationJobIntervalValue:     cfg.DefaultBackendMaximumAttestationJobInterval,
			MaximumWHQLJobIntervalValue:            cfg.DefaultBackendMaximumWHQLJobInterval,
			WaitingDelayTimeOfDispatchingTestValue: cfg.DefaultBackendWaitingDelayTimeOfDispatchingTest,
		},
		FastlaneProxyConfiguration: FastlaneProxyConfiguration{
			PortValue: cfg.DefaultFastlaneProxyPort,
		},
		HLKManagerHostConfiguration: HLKManagerHostConfiguration{
			PortValue: cfg.DefaultHLKManagerHostPort,
		},
		LogConfiguration: LogConfiguration{
			NameValue:                       consts.SystemName + ".log",
			FileMaximumSizeByMegabytesValue: cfg.DefaultLogFileMaximumSizeByMegabytes,
			FileMaximumBackupsValue:         cfg.DefaultLogFileMaximumBackups,
			FileMaximumAgeByDaysValue:       cfg.DefaultLogFileMaximumAgeByDays,
			LevelValue:                      cfg.DefaultLogLevel,
			SlowSQLThresholdValue:           cfg.DefaultLogSlowSQLThreshold,
		},
		MySQLConfiguration: MySQLConfiguration{
			UsernameValue:                    cfg.DefaultMySQLUsername,
			PasswordValue:                    cfg.DefaultMySQLPassword,
			HostValue:                        cfg.DefaultMySQLHost,
			PortValue:                        cfg.DefaultMySQLPort,
			DatabaseValue:                    cfg.DefaultMySQLDatabase,
			MaximumIdleValue:                 cfg.DefaultMySQLMaximumIdle,
			MaximumOpenValue:                 cfg.DefaultMySQLMaximumOpen,
			MaximumLifeValue:                 cfg.DefaultMySQLMaximumLife,
			MaximumNumberOfPerSQLInsertValue: cfg.DefaultMySQLMaximumNumberOfPerSQLInsert,
		},
		RabbitMQConfiguration: RabbitMQConfiguration{
			UsernameValue:                       cfg.DefaultRabbitMQUsername,
			PasswordValue:                       cfg.DefaultRabbitMQPassword,
			HostValue:                           cfg.DefaultRabbitMQHost,
			PortValue:                           cfg.DefaultRabbitMQPort,
			VirtualHostValue:                    cfg.DefaultRabbitMQVirtualHost,
			WindowsOVSigningJobQueueValue:       cfg.DefaultRabbitMQWindowsOVSigningJobQueue,
			WindowsEVSigningJobQueuePrefixValue: cfg.DefaultRabbitMQWindowsEVSigningJobQueuePrefix,
			AndroidSigningJobQueueValue:         cfg.DefaultRabbitMQAndroidSigningJobQueue,
			AppleSigningJobQueueValue:           cfg.DefaultRabbitMQAppleSigningJobQueue,
		},
		TusdConfiguration: TusdConfiguration{
			HostValue: cfg.DefaultTusdHost,
			PortValue: cfg.DefaultTusdPort,
		},
		RedisConfiguration: RedisConfiguration{
			HostValue: cfg.DefaultRedisHost,
			PortValue: cfg.DefaultRedisPort,
		},
		SwaggerConfiguration: SwaggerConfiguration{
			HostValue:    cfg.DefaultSwaggerHost,
			VersionValue: cfg.DefaultSwaggerVersion,
		},
	}
}

func (c *Configuration) Environment() string {
	return c.EnvironmentValue
}

func (c *Configuration) TimeZone() string {
	return c.TimeZoneValue
}

func (c *Configuration) GatewayHost() string {
	return c.GatewayHostValue
}

func (c *Configuration) GatewayInternalServerPort() int {
	return c.GatewayInternalServerPortValue
}

func (c *Configuration) TLSInsecureSkipVerify() bool {
	return c.TLSInsecureSkipVerifyValue
}

func (c *Configuration) Backend() cfg.BackendConfigurer {
	return &c.BackendConfiguration
}

func (c *Configuration) FastlaneProxy() cfg.FastlaneProxyConfigurer {
	return &c.FastlaneProxyConfiguration
}

func (c *Configuration) HLKManagerHost() cfg.HLKManagerHostConfigurer {
	return &c.HLKManagerHostConfiguration
}

func (c *Configuration) HLKManagerTest() cfg.HLKManagerTestConfigurer {
	return c.HLKManagerTestConfiguration
}

func (c *Configuration) HLKManagerController() cfg.HLKManagerTestConfigurer {
	return c.HLKManagerControllerConfiguration
}

func (c *Configuration) Apple() cfg.AppleConfigurer {
	return &c.AppleConfiguration
}

func (c *Configuration) Log() cfg.LogConfigurer {
	return &c.LogConfiguration
}

func (c *Configuration) MySQL() cfg.MySQLConfigurer {
	return &c.MySQLConfiguration
}

func (c *Configuration) RabbitMQ() cfg.RabbitMQConfigurer {
	return &c.RabbitMQConfiguration
}

func (c *Configuration) Tusd() cfg.TusdConfigurer {
	return &c.TusdConfiguration
}

func (c *Configuration) Redis() cfg.RedisConfigurer {
	return &c.RedisConfiguration
}

func (c *Configuration) AppleAPI() cfg.AppleAPIConfigurer {
	return &c.AppleAPIConfiguration
}

func (c *Configuration) MicrosoftAPI() cfg.MicrosoftAPIConfigurer {
	return &c.MicrosoftAPIConfiguration
}

func (c *Configuration) Swagger() cfg.SwaggerConfigurer {
	return &c.SwaggerConfiguration
}

func (c *BackendConfiguration) Port() int {
	return c.PortValue
}

func (c *BackendConfiguration) InternalPort() int {
	return c.InternalPortValue
}

func (c *BackendConfiguration) WriteTimeout() time.Duration {
	return c.WriteTimeoutValue
}

func (c *BackendConfiguration) ReadTimeout() time.Duration {
	return c.ReadTimeoutValue
}

func (c *BackendConfiguration) MinimumRequestInterval() time.Duration {
	return c.MinimumRequestIntervalValue
}

func (c *BackendConfiguration) OpenAPIMaximumExpirationDuration() time.Duration {
	return c.OpenAPIMaximumExpirationDurationValue
}

func (c *BackendConfiguration) MaximumSendInterval() time.Duration {
	return c.MaximumSendIntervalValue
}

func (c *BackendConfiguration) UserAvatarMaximumSize() int {
	return c.UserAvatarMaximumSizeValue
}

func (c *BackendConfiguration) AppLogoMaximumSize() int {
	return c.AppLogoMaximumSizeValue
}

func (c *BackendConfiguration) FileUploadingMaximumInterval() time.Duration {
	return c.FileUploadingMaximumIntervalValue
}

func (c *BackendConfiguration) WebSessionExpiration() time.Duration {
	return c.WebSessionExpirationValue
}

func (c *BackendConfiguration) MaximumAttestationJobInterval() time.Duration {
	return c.MaximumAttestationJobIntervalValue
}

func (c *BackendConfiguration) MaximumWHQLJobInterval() time.Duration {
	return c.MaximumWHQLJobIntervalValue
}

func (c *BackendConfiguration) WaitingDelayTimeOfDispatchingTest() time.Duration {
	return c.WaitingDelayTimeOfDispatchingTestValue
}

func (c *FastlaneProxyConfiguration) Port() int {
	return c.PortValue
}

func (c *HLKManagerHostConfiguration) Port() int {
	return c.PortValue
}

func (c *AppleConfiguration) ApplyCertificateCSR() string {
	return c.ApplyCertificateCSRValue
}

func (c *AppleConfiguration) CertificatePrivateKey() string {
	return c.CertificatePrivateKeyValue
}

func (c *AppleConfiguration) AppStoreTeamID() string {
	return c.AppStoreTeamIDValue
}

func (c *AppleConfiguration) InHouseTeamID() string {
	return c.InHouseTeamIDValue
}

func (c *AppleConfiguration) AccountName() string {
	return c.AccountNameValue
}

func (c *AppleConfiguration) CommonProfile() string {
	return c.CommonProfileValue
}

func (c *AppleConfiguration) CommonProfileID() string {
	return c.CommonProfileIDValue
}

func (c *AppleConfiguration) CertificateIDOfCommonProfile() string {
	return c.CertificateIDOfCommonProfileValue
}

func (c *LogConfiguration) Name() string {
	return c.NameValue
}

func (c *LogConfiguration) FileMaximumSizeByMegabytes() int {
	return c.FileMaximumSizeByMegabytesValue
}

func (c *LogConfiguration) FileMaximumBackups() int {
	return c.FileMaximumBackupsValue
}

func (c *LogConfiguration) FileMaximumAgeByDays() int {
	return c.FileMaximumAgeByDaysValue
}

func (c *LogConfiguration) Level() string {
	return c.LevelValue
}

func (c *LogConfiguration) SlowSQLThreshold() time.Duration {
	return c.SlowSQLThresholdValue
}

func (c *LogConfiguration) ConsoleColorful() bool {
	return c.ConsoleColorfulValue
}

func (c *LogConfiguration) ReportPath() string {
	return c.ReportPathValue
}

func (c *MySQLConfiguration) Username() string {
	return c.UsernameValue
}

func (c *MySQLConfiguration) Password() string {
	return c.PasswordValue
}

func (c *MySQLConfiguration) Host() string {
	return c.HostValue
}

func (c *MySQLConfiguration) Port() int {
	return c.PortValue
}

func (c *MySQLConfiguration) Database() string {
	return c.DatabaseValue
}

func (c *MySQLConfiguration) Parameters() string {
	return c.ParametersValue
}

func (c *MySQLConfiguration) MaximumIdle() int {
	return c.MaximumIdleValue
}

func (c *MySQLConfiguration) MaximumOpen() int {
	return c.MaximumOpenValue
}

func (c *MySQLConfiguration) MaximumLife() time.Duration {
	return c.MaximumLifeValue
}

func (c *MySQLConfiguration) MaximumNumberOfPerSQLInsert() int {
	return c.MaximumNumberOfPerSQLInsertValue
}

func (c *RabbitMQConfiguration) Username() string {
	return c.UsernameValue
}

func (c *RabbitMQConfiguration) Password() string {
	return c.PasswordValue
}

func (c *RabbitMQConfiguration) Host() string {
	return c.HostValue
}

func (c *RabbitMQConfiguration) Port() int {
	return c.PortValue
}

func (c *RabbitMQConfiguration) VirtualHost() string {
	return c.VirtualHostValue
}

func (c *RabbitMQConfiguration) PrefetchCount() int {
	return c.PrefetchCountValue
}

func (c *RabbitMQConfiguration) WindowsOVSigningJobQueue() string {
	return c.WindowsOVSigningJobQueueValue
}

func (c *RabbitMQConfiguration) WindowsEVSigningJobQueuePrefix() string {
	return c.WindowsEVSigningJobQueuePrefixValue
}

func (c *RabbitMQConfiguration) AndroidSigningJobQueue() string {
	return c.AndroidSigningJobQueueValue
}

func (c *RabbitMQConfiguration) AppleSigningJobQueue() string {
	return c.AppleSigningJobQueueValue
}

func (c *TusdConfiguration) Host() string {
	return c.HostValue
}

func (c *TusdConfiguration) Port() int {
	return c.PortValue
}

func (c *RedisConfiguration) Username() string {
	return c.UsernameValue
}

func (c *RedisConfiguration) Password() string {
	return c.PasswordValue
}

func (c *RedisConfiguration) Host() string {
	return c.HostValue
}

func (c *RedisConfiguration) Port() int {
	return c.PortValue
}

func (c *RedisConfiguration) Database() int {
	return c.DatabaseValue
}

func (c *AppleAPIConfiguration) IssuerID() string {
	return c.IssuerIDValue
}

func (c *AppleAPIConfiguration) KeyID() string {
	return c.KeyIDValue
}

func (c *AppleAPIConfiguration) Secret() string {
	return c.SecretValue
}

func (c *MicrosoftAPIConfiguration) GrantType() string {
	return c.GrantTypeValue
}

func (c *MicrosoftAPIConfiguration) ClientID() string {
	return c.ClientIDValue
}

func (c *MicrosoftAPIConfiguration) ClientSecret() string {
	return c.ClientSecretValue
}

func (c *MicrosoftAPIConfiguration) Resource() string {
	return c.ResourceValue
}

func (c *MicrosoftAPIConfiguration) TenantID() string {
	return c.TenantIDValue
}

func (c *SwaggerConfiguration) Host() string {
	return c.HostValue
}

func (c *SwaggerConfiguration) Schema() string {
	return c.SchemaValue
}

func (c *SwaggerConfiguration) Port() int {
	return c.PortValue
}

func (c *SwaggerConfiguration) Version() string {
	return c.VersionValue
}

func (c *SwaggerConfiguration) BasePath() string {
	return c.BasePathValue
}
