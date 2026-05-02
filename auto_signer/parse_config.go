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

package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	cl "gitee.com/ivfzhou/csms/comm/log"
)

var configFilePath string

// Configuration 配置。
type Configuration struct {
	// 基础
	Base struct {
		// 服务地址
		ServerAddress string `yaml:"serverAddress"`
		// 应用 ID
		AppID string `yaml:"appId"`
		// 请求凭证 ID
		AccountID string `yaml:"accountId"`
		// 请求凭证密钥
		Secret string `yaml:"secret"`
		// 任务类型
		JobType string `yaml:"jobType"`
		// 待签名文件
		InFile string `yaml:"inFile"`
		// 签名结果文件
		OutFile string `yaml:"outFile"`
	} `json:"base"`
	// 签名参数
	SignConfig struct {
		// Windows 签名
		Windows struct {
			// 任务类型
			SigningType string `yaml:"signingType"`
			// 证书 ID
			CertificateID string `yaml:"certificateId"`
		} `yaml:"windows"`
		// WHQL 认证
		WHQL struct {
			// 任务类型
			Type string `yaml:"type"`
			// 测试系统
			TestSystem string `yaml:"testSystem"`
			// 服务进程名
			ServiceName string `yaml:"serviceName"`
			// 测试标的
			TestTarget string `yaml:"testTarget"`
			// 测试配置文件路径
			TestConfigFilePath string `yaml:"testConfigFilePath"`
		} `yaml:"whql"`
		// 安卓签名配置
		Android struct {
			// 类型
			Type string `yaml:"type"`
			// 签名方案
			SignatureSchema []int `yaml:"signatureSchema"`
			// 证书 ID
			CertificateID string `yaml:"certificateId"`
			// Android API 版本
			MinimumSDKVersion int `json:"minimumSDKVersion"`
		} `yaml:"android"`
		// 苹果签名
		Apple struct {
			// 描述文件 ID
			ProvisionID string `yaml:"provisionId"`
		} `yaml:"apple"`
	} `yaml:"signConfig"`
}

// AddConfigCommandFlag 添加命令参数。
func AddConfigCommandFlag() {
	flag.StringVar(&configFilePath, "config", "config.yml", "required, config file location")
}

// ParseConfig 获取配置。
func ParseConfig() (*Configuration, bool) {
	// 读取配置文件。
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Println(cl.LevelError, "failed to read config file", err)
		return nil, false
	}

	// 反序列化配置数据。
	var cfgData Configuration
	if err = yaml.Unmarshal(fileData, &cfgData); err != nil {
		log.Println(cl.LevelError, "failed to parse config file data", err)
		return nil, false
	}

	return &cfgData, true
}
