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
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	appleJobIDs  sync.Map
	appleJobLock sync.Mutex
)

// StartAppleSignServer 启动 Apple 签名服务。
func StartAppleSignServer(ctx context.Context) <-chan struct{} {
	log.Info(ctx, "start apple sign server")

	// 声明队列。
	queue := cfg.Get().RabbitMQ().AppleSigningJobQueue()
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
						signAppleFile(ctxs.New(), delivery)
					}(delivery)
				}
			}
		}
	}()

	return closeChannel
}

// 执行签名。
func signAppleFile(ctx context.Context, delivery amqp.Delivery) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "sign panic", p, util.GetStackCallers())
		}
	}()

	jobID := string(delivery.Body)
	log.Info(ctx, "delivery received", jobID)

	// 加锁，避免同时处理多个同一任务。
	appleJobLock.Lock()
	defer appleJobLock.Unlock()

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
	appleSigningJob, err := httpBackendGetAppleSigningJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "failed to get apple signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	if appleSigningJob == nil {
		log.Error(ctx, "job not exists", jobID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 判断任务状态。
	if appleSigningJob.Status != model.AppleSigningJobStatusRunning {
		log.Error(ctx, "invalid job status", appleSigningJob.Status)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 校验出队次数。
	timesValue, _ := appleJobIDs.Load(jobID)
	times, _ := timesValue.(int)
	if times > consts.MaximinOutQueueTimes {
		failAppleJob(ctx, jobID, "任务处理次数超出限制")
		log.Error(ctx, "job exceeds maximin out queue times", times, jobID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}
	appleJobIDs.Store(jobID, times+1)

	// 获取证书和描述文件。
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "获取证书和描述文件")
	certificateAndProfile, err := httpBackendGetAppleCertificateAndProfile(ctx, appleSigningJob.ProfileID)
	if err != nil {
		log.Error(ctx, "failed to get apple certificate and profile", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	if certificateAndProfile == nil {
		failAppleJob(ctx, jobID, "证书或描述文件未找到")
		log.Error(ctx, "certificate and profile not exists", appleSigningJob.ProfileID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(filepath.Join(cc.ServiceNameSigner, consts.ModeApple, jobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书和描述文件写入磁盘。
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "证书和描述文件写入磁盘")
	certificatePath := filepath.Join(workspace, "a.certificate")
	certificateBytes, err := base64.StdEncoding.DecodeString(certificateAndProfile.Certificate)
	if err != nil {
		appendAppleJobLog(ctx, jobID, log.LevelError, "Base64 解码证书失败")
		log.Error(ctx, "failed to base64 decode certificate", err)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", jobID)
		return
	}
	err = os.WriteFile(certificatePath, certificateBytes, cc.FileMode)
	if err != nil {
		log.Error(ctx, "failed to write certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	profilePath := filepath.Join(workspace, "a.profile")
	profileBytes, err := base64.StdEncoding.DecodeString(certificateAndProfile.Profile)
	if err != nil {
		appendAppleJobLog(ctx, jobID, log.LevelError, "Base64 解码描述文件失败")
		log.Error(ctx, "failed to base64 decode profile", err)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", jobID)
		return
	}
	err = os.WriteFile(profilePath, profileBytes, cc.FileMode)
	if err != nil {
		log.Error(ctx, "failed to write profile", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 下载待签名文件。
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, appleSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 记录 zsign 版本信息。
	output, _ := exec.Command(consts.ZsignFilePath, "-v").CombinedOutput()
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "zsign 版本信息：%s", output)

	// 文件签名。
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "签名文件")
	fileName := filepath.Base(sourceFilePath)
	ext := filepath.Ext(sourceFilePath)
	fileName = fileName[:len(fileName)-len(ext)] + "_signed" + ext
	destinationFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName)
	output, err = exec.Command(consts.ZsignFilePath,
		"-k", certificatePath,
		"-m", profilePath,
		"-p", certificateAndProfile.Password,
		"-f",
		"-o", destinationFilePath,
		sourceFilePath,
	).CombinedOutput()
	if err != nil {
		failAppleJob(ctx, jobID, "签名文件失败：%v %s", err, output)
		log.Error(ctx, "failed to sign file", err, output)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}
	appendAppleJobLog(ctx, jobID, log.LevelInfo, "签名输出：%s", output)

	// 上传结果文件。
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
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeAppleSigning, filepath.Base(destinationFilePath),
		appleSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	// 更新任务。
	err = httpBackendUpdateAppleSigningJob(ctx, &bp.AppleInternalUpdateSigningJobReq{
		JobID:        jobID,
		Status:       model.AppleSigningJobStatusSuccess,
		AppendLog:    formatJobLog(log.LevelInfo, "签名成功"),
		FinishedTime: bp.Time(time.Now()),
		SignedFileID: signedFileID,
	})
	if err != nil {
		log.Error(ctx, "failed to update apple signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}

	appleJobIDs.Delete(jobID)
	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
	log.Info(ctx, "sign successfully")
}

// 更新任务为失败状态。
func failAppleJob(ctx context.Context, jobID string, format string, args ...any) {
	err := httpBackendUpdateAppleSigningJob(ctx, &bp.AppleInternalUpdateSigningJobReq{
		JobID:        jobID,
		Status:       model.AppleSigningJobStatusFailure,
		AppendLog:    formatJobLog(log.LevelError, format, args...),
		FinishedTime: bp.Time(time.Now()),
	})
	log.ErrorIf(ctx, err, "failed to update apple signing job", jobID)
}

// 打印任务日志。
func appendAppleJobLog(ctx context.Context, jobID string, level log.Level, format string, args ...any) {
	err := httpBackendUpdateAppleSigningJob(ctx, &bp.AppleInternalUpdateSigningJobReq{
		JobID:     jobID,
		AppendLog: formatJobLog(level, format, args...),
	})
	log.ErrorIf(ctx, err, "failed to append apple signing job log", jobID)
}
