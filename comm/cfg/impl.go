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

var defaultValue = &defaultConfiguration{
	EnvironmentValue:               DefaultEnvironment,
	TimeZoneValue:                  DefaultTimeZone,
	GatewayHostValue:               DefaultGatewayHost,
	GatewayInternalServerPortValue: DefaultGatewayInternalServerPort,
	BackendConfiguration: defaultBackendConfiguration{
		PortValue:                              DefaultBackendPort,
		InternalPortValue:                      DefaultBackendInternalPort,
		WriteTimeoutValue:                      DefaultBackendWriteTimeout,
		ReadTimeoutValue:                       DefaultBackendReadTimeout,
		MinimumRequestIntervalValue:            DefaultBackendMinimumRequestInterval,
		OpenAPIMaximumExpirationDurationValue:  DefaultBackendOpenAPIMaximumExpirationDuration,
		MaximumSendIntervalValue:               DefaultBackendMaximumSendInterval,
		UserAvatarMaximumSizeValue:             DefaultBackendUserAvatarMaximumSize,
		AppLogoMaximumSizeValue:                DefaultBackendAppLogoMaximumSize,
		FileUploadingMaximumIntervalValue:      DefaultBackendFileUploadingMaximumInterval,
		WebSessionExpirationValue:              DefaultBackendWebSessionExpiration,
		MaximumAttestationJobIntervalValue:     DefaultBackendMaximumAttestationJobInterval,
		MaximumWHQLJobIntervalValue:            DefaultBackendMaximumWHQLJobInterval,
		WaitingDelayTimeOfDispatchingTestValue: DefaultBackendWaitingDelayTimeOfDispatchingTest,
	},
	FastlaneProxyConfiguration: defaultFastlaneProxyConfiguration{
		PortValue: DefaultFastlaneProxyPort,
	},
	HLKManagerHostConfiguration: defaultHLKManagerHostConfiguration{
		PortValue: DefaultHLKManagerHostPort,
	},
	LogConfiguration: defaultLogConfiguration{
		FileMaximumSizeByMegabytesValue: DefaultLogFileMaximumSizeByMegabytes,
		FileMaximumBackupsValue:         DefaultLogFileMaximumBackups,
		FileMaximumAgeByDaysValue:       DefaultLogFileMaximumAgeByDays,
		LevelValue:                      DefaultLogLevel,
		SlowSQLThresholdValue:           DefaultLogSlowSQLThreshold,
	},
	MySQLConfiguration: defaultMySQLConfiguration{
		UsernameValue:                    DefaultMySQLUsername,
		PasswordValue:                    DefaultMySQLPassword,
		HostValue:                        DefaultMySQLHost,
		PortValue:                        DefaultMySQLPort,
		DatabaseValue:                    DefaultMySQLDatabase,
		MaximumIdleValue:                 DefaultMySQLMaximumIdle,
		MaximumOpenValue:                 DefaultMySQLMaximumOpen,
		MaximumLifeValue:                 DefaultMySQLMaximumLife,
		MaximumNumberOfPerSQLInsertValue: DefaultMySQLMaximumNumberOfPerSQLInsert,
	},
	RabbitMQConfiguration: defaultRabbitMQConfiguration{
		UsernameValue:                       DefaultRabbitMQUsername,
		PasswordValue:                       DefaultRabbitMQPassword,
		HostValue:                           DefaultRabbitMQHost,
		PortValue:                           DefaultRabbitMQPort,
		VirtualHostValue:                    DefaultRabbitMQVirtualHost,
		WindowsOVSigningJobQueueValue:       DefaultRabbitMQWindowsOVSigningJobQueue,
		WindowsEVSigningJobQueuePrefixValue: DefaultRabbitMQWindowsEVSigningJobQueuePrefix,
		AndroidSigningJobQueueValue:         DefaultRabbitMQAndroidSigningJobQueue,
		AppleSigningJobQueueValue:           DefaultRabbitMQAppleSigningJobQueue,
	},
	TusdConfiguration: defaultTusdConfiguration{
		HostValue: DefaultTusdHost,
		PortValue: DefaultTusdPort,
	},
	RedisConfiguration: defaultRedisConfiguration{
		HostValue: DefaultRedisHost,
		PortValue: DefaultRedisPort,
	},
	SwaggerConfiguration: defaultSwaggerConfiguration{
		HostValue:    DefaultSwaggerHost,
		SchemaValue:  DefaultSwaggerSchema,
		PortValue:    DefaultSwaggerPort,
		VersionValue: DefaultSwaggerVersion,
	},
}

type defaultConfiguration struct {
	EnvironmentValue                  string
	TimeZoneValue                     string
	GatewayHostValue                  string
	GatewayInternalServerPortValue    int
	TLSInsecureSkipVerifyValue        bool
	BackendConfiguration              defaultBackendConfiguration
	FastlaneProxyConfiguration        defaultFastlaneProxyConfiguration
	HLKManagerHostConfiguration       defaultHLKManagerHostConfiguration
	HLKManagerTestConfiguration       defaultHLKManagerTestConfiguration
	HLKManagerControllerConfiguration defaultHLKManagerControllerConfiguration
	AppleConfiguration                defaultAppleConfiguration
	LogConfiguration                  defaultLogConfiguration
	MySQLConfiguration                defaultMySQLConfiguration
	RabbitMQConfiguration             defaultRabbitMQConfiguration
	TusdConfiguration                 defaultTusdConfiguration
	RedisConfiguration                defaultRedisConfiguration
	AppleAPIConfiguration             defaultAppleAPIConfiguration
	MicrosoftAPIConfiguration         defaultMicrosoftAPIConfiguration
	SwaggerConfiguration              defaultSwaggerConfiguration
}

type defaultBackendConfiguration struct {
	PortValue                              int
	InternalPortValue                      int
	WriteTimeoutValue                      time.Duration
	ReadTimeoutValue                       time.Duration
	MinimumRequestIntervalValue            time.Duration
	OpenAPIMaximumExpirationDurationValue  time.Duration
	MaximumSendIntervalValue               time.Duration
	UserAvatarMaximumSizeValue             int
	AppLogoMaximumSizeValue                int
	FileUploadingMaximumIntervalValue      time.Duration
	WebSessionExpirationValue              time.Duration
	MaximumAttestationJobIntervalValue     time.Duration
	MaximumWHQLJobIntervalValue            time.Duration
	WaitingDelayTimeOfDispatchingTestValue time.Duration
}

type defaultFastlaneProxyConfiguration struct {
	PortValue int
}

type defaultHLKManagerHostConfiguration struct {
	PortValue int
}

type defaultHLKManagerTestConfiguration struct{}

type defaultHLKManagerControllerConfiguration struct{}

type defaultAppleConfiguration struct {
	ApplyCertificateCSRValue          string
	CertificatePrivateKeyValue        string
	AppStoreTeamIDValue               string
	InHouseTeamIDValue                string
	AccountNameValue                  string
	CommonProfileValue                string
	CommonProfileIDValue              string
	CertificateIDOfCommonProfileValue string
}

type defaultLogConfiguration struct {
	NameValue                       string
	FileMaximumSizeByMegabytesValue int
	FileMaximumBackupsValue         int
	FileMaximumAgeByDaysValue       int
	LevelValue                      string
	SlowSQLThresholdValue           time.Duration
	ConsoleColorfulValue            bool
	ReportPathValue                 string
}

type defaultMySQLConfiguration struct {
	UsernameValue                    string
	PasswordValue                    string
	HostValue                        string
	PortValue                        int
	DatabaseValue                    string
	ParametersValue                  string
	MaximumIdleValue                 int
	MaximumOpenValue                 int
	MaximumLifeValue                 time.Duration
	MaximumNumberOfPerSQLInsertValue int
}

type defaultRabbitMQConfiguration struct {
	UsernameValue                       string
	PasswordValue                       string
	HostValue                           string
	PortValue                           int
	VirtualHostValue                    string
	PrefetchCountValue                  int
	WindowsOVSigningJobQueueValue       string
	WindowsEVSigningJobQueuePrefixValue string
	AndroidSigningJobQueueValue         string
	AppleSigningJobQueueValue           string
}

type defaultTusdConfiguration struct {
	HostValue string
	PortValue int
}

type defaultRedisConfiguration struct {
	UsernameValue string
	PasswordValue string
	HostValue     string
	PortValue     int
	DatabaseValue int
}

type defaultAppleAPIConfiguration struct {
	IssuerIDValue string
	KeyIDValue    string
	SecretValue   string
}

type defaultMicrosoftAPIConfiguration struct {
	GrantTypeValue    string
	ClientIDValue     string
	ClientSecretValue string
	ResourceValue     string
	TenantIDValue     string
}

type defaultSwaggerConfiguration struct {
	HostValue     string
	SchemaValue   string
	PortValue     int
	VersionValue  string
	BasePathValue string
}

func (c *defaultConfiguration) Environment() string {
	return c.EnvironmentValue
}

func (c *defaultConfiguration) TimeZone() string {
	return c.TimeZoneValue
}

func (c *defaultConfiguration) GatewayHost() string {
	return c.GatewayHostValue
}

func (c *defaultConfiguration) GatewayInternalServerPort() int {
	return c.GatewayInternalServerPortValue
}

func (c *defaultConfiguration) TLSInsecureSkipVerify() bool {
	return c.TLSInsecureSkipVerifyValue
}

func (c *defaultConfiguration) Backend() BackendConfigurer {
	return &c.BackendConfiguration
}

func (c *defaultConfiguration) FastlaneProxy() FastlaneProxyConfigurer {
	return &c.FastlaneProxyConfiguration
}

func (c *defaultConfiguration) HLKManagerHost() HLKManagerHostConfigurer {
	return &c.HLKManagerHostConfiguration
}

func (c *defaultConfiguration) HLKManagerTest() HLKManagerTestConfigurer {
	return c.HLKManagerTestConfiguration
}

func (c *defaultConfiguration) HLKManagerController() HLKManagerTestConfigurer {
	return c.HLKManagerControllerConfiguration
}

func (c *defaultConfiguration) Apple() AppleConfigurer {
	return &c.AppleConfiguration
}

func (c *defaultConfiguration) Log() LogConfigurer {
	return &c.LogConfiguration
}

func (c *defaultConfiguration) MySQL() MySQLConfigurer {
	return &c.MySQLConfiguration
}

func (c *defaultConfiguration) RabbitMQ() RabbitMQConfigurer {
	return &c.RabbitMQConfiguration
}

func (c *defaultConfiguration) Tusd() TusdConfigurer {
	return &c.TusdConfiguration
}

func (c *defaultConfiguration) Redis() RedisConfigurer {
	return &c.RedisConfiguration
}

func (c *defaultConfiguration) AppleAPI() AppleAPIConfigurer {
	return &c.AppleAPIConfiguration
}

func (c *defaultConfiguration) MicrosoftAPI() MicrosoftAPIConfigurer {
	return &c.MicrosoftAPIConfiguration
}

func (c *defaultConfiguration) Swagger() SwaggerConfigurer {
	return &c.SwaggerConfiguration
}

func (c *defaultBackendConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultBackendConfiguration) InternalPort() int {
	return c.InternalPortValue
}

func (c *defaultBackendConfiguration) WriteTimeout() time.Duration {
	return c.WriteTimeoutValue
}

func (c *defaultBackendConfiguration) ReadTimeout() time.Duration {
	return c.ReadTimeoutValue
}

func (c *defaultBackendConfiguration) MinimumRequestInterval() time.Duration {
	return c.MinimumRequestIntervalValue
}

func (c *defaultBackendConfiguration) OpenAPIMaximumExpirationDuration() time.Duration {
	return c.OpenAPIMaximumExpirationDurationValue
}

func (c *defaultBackendConfiguration) MaximumSendInterval() time.Duration {
	return c.MaximumSendIntervalValue
}

func (c *defaultBackendConfiguration) UserAvatarMaximumSize() int {
	return c.UserAvatarMaximumSizeValue
}

func (c *defaultBackendConfiguration) AppLogoMaximumSize() int {
	return c.AppLogoMaximumSizeValue
}

func (c *defaultBackendConfiguration) FileUploadingMaximumInterval() time.Duration {
	return c.FileUploadingMaximumIntervalValue
}

func (c *defaultBackendConfiguration) WebSessionExpiration() time.Duration {
	return c.WebSessionExpirationValue
}

func (c *defaultBackendConfiguration) MaximumAttestationJobInterval() time.Duration {
	return c.MaximumAttestationJobIntervalValue
}

func (c *defaultBackendConfiguration) MaximumWHQLJobInterval() time.Duration {
	return c.MaximumWHQLJobIntervalValue
}

func (c *defaultBackendConfiguration) WaitingDelayTimeOfDispatchingTest() time.Duration {
	return c.WaitingDelayTimeOfDispatchingTestValue
}

func (c *defaultFastlaneProxyConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultHLKManagerHostConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultAppleConfiguration) ApplyCertificateCSR() string {
	return c.ApplyCertificateCSRValue
}

func (c *defaultAppleConfiguration) CertificatePrivateKey() string {
	return c.CertificatePrivateKeyValue
}

func (c *defaultAppleConfiguration) AppStoreTeamID() string {
	return c.AppStoreTeamIDValue
}

func (c *defaultAppleConfiguration) InHouseTeamID() string {
	return c.InHouseTeamIDValue
}

func (c *defaultAppleConfiguration) AccountName() string {
	return c.AccountNameValue
}

func (c *defaultAppleConfiguration) CommonProfile() string {
	return c.CommonProfileValue
}

func (c *defaultAppleConfiguration) CommonProfileID() string {
	return c.CommonProfileIDValue
}

func (c *defaultAppleConfiguration) CertificateIDOfCommonProfile() string {
	return c.CertificateIDOfCommonProfileValue
}

func (c *defaultLogConfiguration) Name() string {
	return c.NameValue
}

func (c *defaultLogConfiguration) FileMaximumSizeByMegabytes() int {
	return c.FileMaximumSizeByMegabytesValue
}

func (c *defaultLogConfiguration) FileMaximumBackups() int {
	return c.FileMaximumBackupsValue
}

func (c *defaultLogConfiguration) FileMaximumAgeByDays() int {
	return c.FileMaximumAgeByDaysValue
}

func (c *defaultLogConfiguration) Level() string {
	return c.LevelValue
}

func (c *defaultLogConfiguration) SlowSQLThreshold() time.Duration {
	return c.SlowSQLThresholdValue
}

func (c *defaultLogConfiguration) ConsoleColorful() bool {
	return c.ConsoleColorfulValue
}

func (c *defaultLogConfiguration) ReportPath() string {
	return c.ReportPathValue
}

func (c *defaultMySQLConfiguration) Username() string {
	return c.UsernameValue
}

func (c *defaultMySQLConfiguration) Password() string {
	return c.PasswordValue
}

func (c *defaultMySQLConfiguration) Host() string {
	return c.HostValue
}

func (c *defaultMySQLConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultMySQLConfiguration) Database() string {
	return c.DatabaseValue
}

func (c *defaultMySQLConfiguration) Parameters() string {
	return c.ParametersValue
}

func (c *defaultMySQLConfiguration) MaximumIdle() int {
	return c.MaximumIdleValue
}

func (c *defaultMySQLConfiguration) MaximumOpen() int {
	return c.MaximumOpenValue
}

func (c *defaultMySQLConfiguration) MaximumLife() time.Duration {
	return c.MaximumLifeValue
}

func (c *defaultMySQLConfiguration) MaximumNumberOfPerSQLInsert() int {
	return c.MaximumNumberOfPerSQLInsertValue
}

func (c *defaultRabbitMQConfiguration) Username() string {
	return c.UsernameValue
}

func (c *defaultRabbitMQConfiguration) Password() string {
	return c.PasswordValue
}

func (c *defaultRabbitMQConfiguration) Host() string {
	return c.HostValue
}

func (c *defaultRabbitMQConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultRabbitMQConfiguration) VirtualHost() string {
	return c.VirtualHostValue
}

func (c *defaultRabbitMQConfiguration) PrefetchCount() int {
	return c.PrefetchCountValue
}

func (c *defaultRabbitMQConfiguration) WindowsOVSigningJobQueue() string {
	return c.WindowsOVSigningJobQueueValue
}

func (c *defaultRabbitMQConfiguration) WindowsEVSigningJobQueuePrefix() string {
	return c.WindowsEVSigningJobQueuePrefixValue
}

func (c *defaultRabbitMQConfiguration) AndroidSigningJobQueue() string {
	return c.AndroidSigningJobQueueValue
}

func (c *defaultRabbitMQConfiguration) AppleSigningJobQueue() string {
	return c.AppleSigningJobQueueValue
}

func (c *defaultTusdConfiguration) Host() string {
	return c.HostValue
}

func (c *defaultTusdConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultRedisConfiguration) Username() string {
	return c.UsernameValue
}

func (c *defaultRedisConfiguration) Password() string {
	return c.PasswordValue
}

func (c *defaultRedisConfiguration) Host() string {
	return c.HostValue
}

func (c *defaultRedisConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultRedisConfiguration) Database() int {
	return c.DatabaseValue
}

func (c *defaultAppleAPIConfiguration) IssuerID() string {
	return c.IssuerIDValue
}

func (c *defaultAppleAPIConfiguration) KeyID() string {
	return c.KeyIDValue
}

func (c *defaultAppleAPIConfiguration) Secret() string {
	return c.SecretValue
}

func (c *defaultMicrosoftAPIConfiguration) GrantType() string {
	return c.GrantTypeValue
}

func (c *defaultMicrosoftAPIConfiguration) ClientID() string {
	return c.ClientIDValue
}

func (c *defaultMicrosoftAPIConfiguration) ClientSecret() string {
	return c.ClientSecretValue
}

func (c *defaultMicrosoftAPIConfiguration) Resource() string {
	return c.ResourceValue
}

func (c *defaultMicrosoftAPIConfiguration) TenantID() string {
	return c.TenantIDValue
}

func (c *defaultSwaggerConfiguration) Host() string {
	return c.HostValue
}

func (c *defaultSwaggerConfiguration) Port() int {
	return c.PortValue
}

func (c *defaultSwaggerConfiguration) Schema() string {
	return c.SchemaValue
}

func (c *defaultSwaggerConfiguration) Version() string {
	return c.VersionValue
}

func (c *defaultSwaggerConfiguration) BasePath() string {
	return c.BasePathValue
}
