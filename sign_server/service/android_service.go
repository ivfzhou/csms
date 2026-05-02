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

package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/sign_server/consts"
)

var (
	androidJobIDs  sync.Map
	androidJobLock sync.Mutex
)

// StartAndroidSignServer 开启安装签名服务。
func StartAndroidSignServer(ctx context.Context) <-chan struct{} {
	log.Info(ctx, "start windows sign server")

	// 声明队列。
	queue := cfg.Get().RabbitMQ().AndroidSigningJobQueue()
	_, err := conn.RabbitMQClient(ctx).QueueDeclare(queue, true, false, false, false, nil)
	log.FatalIf(ctx, cc.ExitCodeDeclareQueueError, err, "declaring queue failed")

	// 启动监听。
	closeChannel := make(chan struct{})
	go func() {
		wg := sync.WaitGroup{}
		defer func() {
			wg.Wait()
			log.Info(ctx, "end consume")
			close(closeChannel)
		}()

		log.Info(ctx, "start consume")
		for {
			// 获取消费队列。
			var consumer <-chan amqp.Delivery
			consumer, err = conn.RabbitMQClient(ctx).ConsumeWithContext(
				ctx, queue, queue+"_"+util.LocalIP, false, false, false, false, nil)
			if err != nil {
				log.Error(ctx, "failed to consume queue", err)
				time.Sleep(3 * time.Second)
				continue
			}

			// 开启消费。
		F:
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				select {
				case <-ctx.Done():
					return
				case delivery, ok := <-consumer:
					if !ok {
						time.Sleep(100 * time.Millisecond)
						break F
					}
					wg.Add(1)
					go func(delivery amqp.Delivery) {
						defer wg.Done()
						signAndroidFile(ctxs.New(), delivery)
					}(delivery)
				}
			}
		}
	}()

	return closeChannel
}

// 执行签名。
func signAndroidFile(ctx context.Context, delivery amqp.Delivery) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "sign panic", p, util.GetStackCallers())
		}
	}()

	jobID := string(delivery.Body)
	log.Info(ctx, "delivery received", jobID)

	// 加锁，避免同时处理多个同一任务。
	androidJobLock.Lock()
	defer androidJobLock.Unlock()

	// 校验消息出队时间。
	sendTime, _ := strconv.ParseInt(fmt.Sprint(delivery.Headers[cc.MQHeaderSendTime]), 10, 64)
	log.Info(ctx, "send time", time.Unix(sendTime, 0).Format(cc.TimeFormat))
	if sendTime <= 0 {
		time.Sleep(consts.MessageDelayTime)
	} else {
		delta := consts.MessageDelayTime - time.Since(time.Unix(sendTime, 0))
		time.Sleep(delta)
	}
	time.Sleep(3 * time.Second)

	// 获取任务信息。
	androidSigningJob, err := httpBackendGetAndroidSigningJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "failed to get android signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	if androidSigningJob == nil {
		log.Error(ctx, "job not exists", jobID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 判断任务状态。
	if androidSigningJob.Status != model.AndroidSigningJobStatusSigning {
		log.Error(ctx, "invalid job status", androidSigningJob.Status)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 校验出队次数。
	timesValue, _ := androidJobIDs.Load(jobID)
	times, _ := timesValue.(int)
	if times > consts.MaximinOutQueueTimes {
		failAndroidJob(ctx, jobID, "任务处理次数超过限制")
		log.Error(ctx, "job exceeds maximin out queue times", times, jobID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}
	androidJobIDs.Store(jobID, times+1)

	// 获取证书信息。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "获取证书信息")
	androidCertificate, err := httpBackendGetAndroidCertificate(ctx, androidSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	if androidCertificate == nil {
		log.Error(ctx, "certificate not found", androidSigningJob.CertificateID)
		failAndroidJob(ctx, jobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(filepath.Join(cc.ServiceNameSigner, consts.ModeAndroid, jobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "将证书写入硬盘")
	certificateFilePath := filepath.Join(workspace, androidCertificate.Alias_+cc.ExtensionJKS)
	err = os.WriteFile(certificateFilePath, androidCertificate.Content, cc.FileMode)
	if err != nil {
		log.Error(ctx, "failed to write certificate to disk", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 下载待签名文件。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, androidSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 文件签名。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "签名文件")
	var outputBytes []byte
	fileName := filepath.Base(sourceFilePath)
	ext := filepath.Ext(sourceFilePath)
	fileName = fileName[:len(fileName)-len(ext)] + "_signed" + ext
	destinationFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName)
	switch androidSigningJob.Type {
	case model.AndroidSigningJobTypeAAB:
		// 记录 jarsigner 版本。
		command := exec.Command(consts.JarsignerFilePath, "-verbose:all", "-version")
		if len(consts.JavaHomeFilePath) > 0 {
			command.Env = append(os.Environ(), "JAVA_HOME="+consts.JavaHomeFilePath)
		}
		outputBytes, err = command.Output()
		log.ErrorIf(ctx, err, "failed to get jarsigner version")
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "jarsigner 版本号：%s", outputBytes)

		// 签名。
		args := []string{
			consts.JarsignerFilePath,
			"-verbose:all",
			"-sigalg", "SHA256withRSA",
			"-digestalg", "SHA-256",
			"-keystore", certificateFilePath,
			"-storepass", androidCertificate.Storepass,
			"-keypass", androidCertificate.Keypass,
			sourceFilePath, androidCertificate.Alias_,
		}
		log.Info(ctx, "execute command",
			strings.ReplaceAll(
				strings.ReplaceAll(strings.Join(args, " "), androidCertificate.Storepass, "******"),
				androidCertificate.Keypass, "******"),
		)
		command = exec.Command(args[0], args[1:]...)
		if len(consts.JavaHomeFilePath) > 0 {
			command.Env = append(os.Environ(), "JAVA_HOME="+consts.JavaHomeFilePath)
		}
		outputBytes, err = command.Output()
		if err != nil {
			log.Error(ctx, "failed to sign aab file", err, outputBytes)
			failAndroidJob(ctx, jobID, "签名失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
			return
		}

		// 记录输出日志。
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "签名输出：%s", outputBytes)

		// 移动文件。
		log.ErrorIf(ctx, os.Rename(sourceFilePath, destinationFilePath), "failed to move file")
	case model.AndroidSigningJobTypeAPK, model.AndroidSigningJobTypePatch:
		// 记录 apksigner 版本。
		command := exec.Command(consts.ApksignerFilePath, "version")
		if len(consts.JavaHomeFilePath) > 0 {
			command.Env = append(os.Environ(), "JAVA_HOME="+consts.JavaHomeFilePath)
		}
		outputBytes, err = command.Output()
		log.ErrorIf(ctx, err, "failed to get apksigner version")
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "apksigner 版本号：%s", outputBytes)

		// 签名。
		args := make([]string, 1, 25)
		args[0] = consts.ApksignerFilePath
		args = append(args, "sign",
			"--verbose",
			"--ks", certificateFilePath,
			"--ks-key-alias", androidCertificate.Alias_,
			"--ks-pass", "env:KS_PASS",
			"--key-pass", "env:KEY_PASS",
			"--out", destinationFilePath,
			"--in", sourceFilePath,
		)
		for _, v := range androidSigningJob.SignatureSchemas {
			switch v {
			case consts.SignatureSchemaVersion1:
				args = append(args, "--v1-signing-enabled", "true")
			case consts.SignatureSchemaVersion2:
				args = append(args, "--v2-signing-enabled", "true")
			case consts.SignatureSchemaVersion3:
				args = append(args, "--v3-signing-enabled", "true")
			case consts.SignatureSchemaVersion4:
				args = append(args, "--v4-signing-enabled", "true")
			}
		}
		if androidSigningJob.Type == model.AndroidSigningJobTypePatch {
			args = append(args, "--min-sdk-version", strconv.Itoa(androidSigningJob.MinimumSdkLevel))
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		command = exec.Command(args[0], args[1:]...)
		command.Env = append(os.Environ(), "KS_PASS="+androidCertificate.Storepass,
			"KEY_PASS="+androidCertificate.Keypass, "JAVA_HOME="+consts.JavaHomeFilePath)
		outputBytes, err = command.Output()
		if err != nil {
			log.Error(ctx, "failed to sign apk file", err, outputBytes)
			failAndroidJob(ctx, jobID, "签名失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
			return
		}
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "签名输出：%s", outputBytes)
	default:
		log.Error(ctx, "invalid job type", androidSigningJob.Type)
		failAndroidJob(ctx, jobID, "错误的签名类型")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 校验签名。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "校验文件签名")
	switch androidSigningJob.Type {
	case model.AndroidSigningJobTypeAAB:
		args := []string{
			consts.JarsignerFilePath,
			"-verbose:all",
			"-verify",
			destinationFilePath, androidCertificate.Alias_,
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		command := exec.Command(args[0], args[1:]...)
		if len(consts.JavaHomeFilePath) > 0 {
			command.Env = append(os.Environ(), "JAVA_HOME="+consts.JavaHomeFilePath)
		}
		outputBytes, err = command.CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to verify aab file", err, outputBytes)
			failAndroidJob(ctx, jobID, "校验签名失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
			return
		}
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "检验文件签名输出：%s", outputBytes)
	case model.AndroidSigningJobTypeAPK, model.AndroidSigningJobTypePatch:
		args := make([]string, 1, 7)
		args[0] = consts.ApksignerFilePath
		args = append(args, "verify", "--verbose", "--in", destinationFilePath)
		if androidSigningJob.Type == model.AndroidSigningJobTypePatch {
			args = append(args, "--min-sdk-version", strconv.Itoa(androidSigningJob.MinimumSdkLevel))
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		command := exec.Command(args[0], args[1:]...)
		if len(consts.JavaHomeFilePath) > 0 {
			command.Env = append(os.Environ(), "JAVA_HOME="+consts.JavaHomeFilePath)
		}
		outputBytes, err = command.Output()
		if err != nil {
			log.Error(ctx, "failed to verify apk file", err, outputBytes)
			failAndroidJob(ctx, jobID, "校验签名失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
			return
		}
		appendAndroidJobLog(ctx, jobID, log.LevelInfo, "检验文件签名输出：%s", outputBytes)
	}

	// 上传结果文件。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(destinationFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	fileInfo, err := os.Stat(destinationFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeAndroidSigning, filepath.Base(destinationFilePath),
		androidSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 更新任务。
	appendAndroidJobLog(ctx, jobID, log.LevelInfo, "更新任务信息")
	err = httpBackendUpdateAndroidSigningJob(ctx, &bp.AndroidInternalUpdateSigningJobReq{
		JobID:        jobID,
		Status:       model.AndroidSigningJobStatusSuccess,
		FinishedTime: bp.Time(time.Now()),
		SignedFileID: signedFileID,
		AppendLog:    formatJobLog(log.LevelInfo, "签名成功"),
	})
	if err != nil {
		log.Error(ctx, "failed to update android signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", jobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
	androidJobIDs.Delete(jobID)
	log.Info(ctx, "sign successfully")
}

// 更新任务为失败状态。
func failAndroidJob(ctx context.Context, jobID string, format string, args ...any) {
	err := httpBackendUpdateAndroidSigningJob(ctx, &bp.AndroidInternalUpdateSigningJobReq{
		JobID:        jobID,
		Status:       model.AndroidSigningJobStatusFailure,
		AppendLog:    formatJobLog(log.LevelError, format, args...),
		FinishedTime: bp.Time(time.Now()),
	})
	log.ErrorIf(ctx, err, "failed to update android signing job", jobID)
}

// 打印任务日志。
func appendAndroidJobLog(ctx context.Context, jobID string, level log.Level, format string, args ...any) {
	err := httpBackendUpdateAndroidSigningJob(ctx, &bp.AndroidInternalUpdateSigningJobReq{
		JobID:     jobID,
		AppendLog: formatJobLog(level, format, args...),
	})
	log.ErrorIf(ctx, err, "failed to append android signing job log", jobID)
}
