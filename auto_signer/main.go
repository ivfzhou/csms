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
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
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
		fmt.Printf("%s", Version())
		return
	}

	// 初始化日志打印。
	taskID := strings.ReplaceAll(uuid.NewString(), "-", "")
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)
	log.SetPrefix(taskID + " ")
	logFile, _ := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND, 0666)
	if logFile != nil {
		log.SetOutput(logFile)
		defer CloseIO(logFile)
	} else {
		log.SetOutput(io.Discard)
	}

	startTime := time.Now()

	PrintHeader(taskID)
	fmt.Println()

	// 步骤 1：解析签名配置。
	step1 := NewStepRunner(1, 5, "解析签名配置")
	step1.Start()
	cfg, token, fileSize, info, err := ParseConfig()
	if err != nil {
		step1.Fail(err.Error())
		PrintErrorFooter(time.Since(startTime))
		os.Exit(ExitCodeParseYamlError)
	}
	step1.Done(info...)
	fmt.Println()

	// 步骤 2：上传待签名文件。
	step2 := NewStepRunner(2, 5, "上传待签名文件")
	step2.Start()
	token, fileID, info, err := UploadFile(cfg, token, fileSize, step2)
	if err != nil {
		step2.Fail(err.Error())
		PrintErrorFooter(time.Since(startTime))
		os.Exit(ExitCodeUploadFileError)
	}
	step2.Done(info...)
	fmt.Println()

	// 步骤 3：提交签名任务。
	step3 := NewStepRunner(3, 5, "提交签名任务")
	step3.Start()
	var jobID string
	switch cfg.Base.JobType {
	case JobTypeWindows:
		token, jobID, info, err = SubmitWindowsSigningJob(cfg, token, fileID)
	case JobTypeWHQL:
		token, jobID, info, err = SubmitWHQLJob(cfg, token, fileID)
	case JobTypeAndroid:
		token, jobID, info, err = SubmitAndroidSigningJob(cfg, token, fileID)
	case JobTypeApple:
		token, jobID, info, err = SubmitAppleSigningJob(cfg, token, fileID)
	default:
		step3.Fail(fmt.Sprintf("任务类型非法，请检查：%s", cfg.Base.JobType))
		os.Exit(ExitCodeInvalidConfigError)
	}
	if err != nil {
		step3.Fail(err.Error())
		PrintErrorFooter(time.Since(startTime))
		os.Exit(ExitCodeSubmitJobError)
	}
	step3.Done(info...)
	fmt.Println()

	// 步骤 4：监听签名结果。
	step4 := NewStepRunner(4, 5, "监听签名结果")
	step4.Start()
	var signedFileID string
	switch cfg.Base.JobType {
	case JobTypeWindows:
		token, signedFileID, info, err = ListenWindowsJob(cfg, token, jobID, step4)
	case JobTypeWHQL:
		token, signedFileID, info, err = ListenWHQLJob(cfg, token, jobID, step4)
	case JobTypeAndroid:
		token, signedFileID, info, err = ListenAndroidJob(cfg, token, jobID, step4)
	case JobTypeApple:
		token, signedFileID, info, err = ListenAppleJob(cfg, token, jobID, step4)
	default:
		step4.Fail(fmt.Sprintf("非法的任务类型：%s", cfg.Base.JobType))
		os.Exit(ExitCodeInvalidConfigError)
	}
	if err != nil {
		step4.Fail(err.Error())
		PrintErrorFooter(time.Since(startTime))
		os.Exit(ExitCodeListenJobError)
	}
	step4.Done(info...)
	fmt.Println()

	// 步骤 5：下载签名文件。
	step5 := NewStepRunner(5, 5, "下载签名文件")
	step5.Start()
	info, err = DownloadFile(cfg, token, signedFileID, step5)
	if err != nil {
		step5.Fail(err.Error())
		PrintErrorFooter(time.Since(startTime))
		os.Exit(ExitCodeDownloadFileError)
	}
	step5.Done(info...)
	fmt.Println()

	PrintSuccessFooter(time.Since(startTime))
}
