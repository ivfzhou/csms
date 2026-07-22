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
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var configFilePath string

// Configuration 配置。
type Configuration struct {
	// 基础
	Base struct {
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
			Type string `yaml:"type"`
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
	flag.StringVar(&configFilePath, "config", "config.yml", "config file location")
}

// ParseConfig 获取配置，
func ParseConfig() (*Configuration, string, int64, []string, error) {
	// 读取配置文件。
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, "", 0, nil, fmt.Errorf("请检查配置文件路径：%v", err)
	}

	// 反序列化配置数据。
	var cfgData Configuration
	if err = yaml.Unmarshal(fileData, &cfgData); err != nil {
		return nil, "", 0, nil, fmt.Errorf("请检查配置文件 YAML 格式：%v", err)
	}

	// 解析输入文件路径并获取文件大小。
	inFilePath, err := filepath.Abs(cfgData.Base.InFile)
	if err != nil {
		return nil, "", 0, nil, fmt.Errorf("请检查输入文件路径：%v", err)
	}
	fileInfo, err := os.Stat(inFilePath)
	if err != nil {
		return nil, "", 0, nil, fmt.Errorf("请检查输入文件是否存在：%v", err)
	}
	if fileInfo.IsDir() {
		return nil, "", 0, nil, fmt.Errorf("输入文件路径不能是文件夹：%v", err)
	}
	fileSize := fileInfo.Size()

	// 解析输出文件路径。
	outFilePath, err := filepath.Abs(cfgData.Base.OutFile)
	if err != nil {
		return nil, "", 0, nil, fmt.Errorf("请检查输出文件路径：%v", err)
	}

	// WHQL 配置校验。
	if cfgData.Base.JobType == JobTypeWHQL && len(cfgData.SignConfig.WHQL.TestConfigFilePath) > 0 {
		_, err = filepath.Abs(cfgData.SignConfig.WHQL.TestConfigFilePath)
		if err != nil {
			return nil, "", 0, nil, fmt.Errorf("请检查 HLK 测试配置文件路径：%v", err)
		}
	}

	// 生成请求凭证。
	token, err := CreateAuthorization(&cfgData)
	if err != nil {
		return nil, "", 0, nil, err
	}

	// 生成打印信息。
	info := make([]string, 0, 12)
	info = append(info, fmt.Sprintf("应用 ID: %s", cfgData.Base.AppID))
	info = append(info, fmt.Sprintf("凭证 ID: %s", cfgData.Base.AccountID))
	info = append(info, fmt.Sprintf("输入文件: %s (%s)", inFilePath, FormatSize(fileSize)))
	info = append(info, fmt.Sprintf("输出文件: %s", outFilePath))
	info = append(info, fmt.Sprintf("任务类型: %s", cfgData.Base.JobType))
	switch cfgData.Base.JobType {
	case JobTypeWindows:
		info = append(info, fmt.Sprintf("签名类型: %s", cfgData.SignConfig.Windows.Type))
		info = append(info, fmt.Sprintf("证书 ID: %s", cfgData.SignConfig.Windows.CertificateID))
	case JobTypeWHQL:
		info = append(info, fmt.Sprintf("签名类型: %s", cfgData.SignConfig.WHQL.Type))
		info = append(info, fmt.Sprintf("测试系统: %s", cfgData.SignConfig.WHQL.TestSystem))
	case JobTypeAndroid:
		info = append(info, fmt.Sprintf("签名类型: %s", cfgData.SignConfig.Android.Type))
		info = append(info, fmt.Sprintf("证书 ID: %s", cfgData.SignConfig.Android.CertificateID))
	case JobTypeApple:
		info = append(info, fmt.Sprintf("描述文件: %s", cfgData.SignConfig.Apple.ProvisionID))
	}

	return &cfgData, token, fileSize, info, nil
}
