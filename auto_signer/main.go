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
	"log"
	"os"

	cl "gitee.com/ivfzhou/csms/comm/log"
)

func init() {
	// 设置日志格式。
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)

	// 处理命令参数。
	AddVersionCommandFlag()
	AddConfigCommandFlag()
	usage := flag.Usage
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "CSMS 自动化签名程序 %s\n", Version())
		usage()
	}

	flag.Parse()
}

func main() {
	// 打印版本号。
	if printVersion {
		fmt.Printf("%s\n", Version())
		return
	}

	log.Println(cl.LevelInfo, "BEGIN SIGN")

	// 解析命令行参数，读取 yaml 文件，解析配置数据。
	log.Println(cl.LevelInfo, "parse config file")
	cfg, ok := ParseConfig()
	if !ok {
		os.Exit(ExitCodeParseYamlError)
	}

	// 获取请求凭证。
	log.Println(cl.LevelInfo, "create an access token")
	token, ok := CreateAuthorization(cfg)
	if !ok {
		os.Exit(ExitCodeGetAccessTokenError)
	}

	// 上传待签名文件。
	log.Println(cl.LevelInfo, "upload file")
	fileID, token, ok := UploadFile(cfg, token)
	if !ok {
		os.Exit(ExitCodeUploadFileError)
	}
	log.Println(cl.LevelInfo, "file id is", fileID)

	// 根据配置，提交签名任务。
	log.Println(cl.LevelInfo, "submit job")
	var jobID string
	switch cfg.Base.JobType {
	case JobTypeWindows:
		jobID, token, ok = SubmitWindowsSigningJob(cfg, token, fileID)
	case JobTypeWHQL:
		jobID, token, ok = SubmitWHQLJob(cfg, token, fileID)
	case JobTypeAndroid:
		jobID, token, ok = SubmitAndroidSigningJob(cfg, token, fileID)
	case JobTypeApple:
		jobID, token, ok = SubmitAppleSigningJob(cfg, token, fileID)
	default:
		log.Println(cl.LevelError, "invalid job type", cfg.Base.JobType)
		os.Exit(ExitCodeInvalidConfigError)
	}
	if !ok {
		os.Exit(ExitCodeSubmitJobError)
	}
	log.Println(cl.LevelInfo, "job id is", jobID)

	// 监听任务结果。
	log.Println(cl.LevelInfo, "listen job")
	var signedFileID string
	switch cfg.Base.JobType {
	case JobTypeWindows:
		signedFileID, token, ok = ListenWindowsJob(cfg, token, jobID)
	case JobTypeWHQL:
		signedFileID, token, ok = ListenWHQLJob(cfg, token, jobID)
	case JobTypeAndroid:
		signedFileID, token, ok = ListenAndroidJob(cfg, token, jobID)
	case JobTypeApple:
		signedFileID, token, ok = ListenAppleJob(cfg, token, jobID)
	default:
		log.Println(cl.LevelError, "invalid job type", cfg.Base.JobType)
		os.Exit(ExitCodeInvalidConfigError)
	}
	if !ok {
		os.Exit(ExitCodeListenJobError)
	}
	log.Println(cl.LevelInfo, "result file id is", signedFileID)

	// 下载签名结果文件。
	log.Println(cl.LevelInfo, "download file")
	ok = DownloadFile(cfg, token, signedFileID)
	if !ok {
		os.Exit(ExitCodeDownloadFileError)
	}

	log.Println(cl.LevelInfo, "END SIGN")
}
