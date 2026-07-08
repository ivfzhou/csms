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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
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
	windowsJobIDs  sync.Map
	windowsJobLock sync.Mutex
	findPEHeader   = regexp.MustCompile(`PE[\w\W]{0,15}`)
)

// StartWindowsSignServer 开启 Windows 签名服务。
func StartWindowsSignServer(ctx context.Context) <-chan struct{} {
	log.Info(ctx, "start windows sign server")

	// 启动监听。
	closeChannel := make(chan struct{})
	go func() {
		defer close(closeChannel)

		for {
			slept := consumeWindowsMessage(ctx)

			select {
			case <-ctx.Done():
				return
			default:
			}

			if slept {
				time.Sleep(3 * time.Second)
			}
		}
	}()

	return closeChannel
}

// 消费签名任务。
func consumeWindowsMessage(ctx context.Context) (slept bool) {
	// 获取证书和队列。
	queues, err := getWindowsQueues(ctx)
	if err != nil {
		return true
	}

	// 声明队列。
	err = declareWindowsQueues(ctx, queues)
	if err != nil {
		return true
	}

	// 消费队列。
	queueToConsumer, err := consumeWindowsQueues(ctx, queues)
	if err != nil {
		return true
	}

	// 启动消费。
	log.Info(ctx, "start consume")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg := sync.WaitGroup{}
	for queue, consumer := range queueToConsumer {
		wg.Add(1)
		go func(queue string, consumer <-chan amqp.Delivery) {
			defer func() {
				log.Warn(ctx, "stop consume", queue)
				wg.Done()
			}()
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
						return
					}
					wg.Add(1)
					go func(delivery amqp.Delivery) {
						defer wg.Done()
						signWindowsFile(ctxs.New(), delivery)
					}(delivery)
				}
			}
		}(queue, consumer)
	}

	// 监听证书更新。
	wg.Go(func() {
		for {
			time.Sleep(10 * time.Second)

			// 监听退出信号。
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 获取证书和队列。
			var newQueues []string
			newQueues, err = getWindowsQueues(ctx)
			if err != nil {
				continue
			}

			// 比较。
			if len(newQueues) != len(queues) {
				slept = true
				cancel()
				log.Warn(ctx, "update consume")
				return
			}
			for _, v := range newQueues {
				if !slices.Contains(queues, v) {
					slept = true
					cancel()
					log.Warn(ctx, "update consume")
					return
				}
			}
		}
	})

	wg.Wait()
	log.Info(ctx, "end consume")

	return
}

// 获取证书和队列。
func getWindowsQueues(ctx context.Context) (queues []string, err error) {
	certificateFingerprints, err := httpBackendGetMachineEVCertificates(ctx, util.LocalIP)
	if err != nil {
		log.Error(ctx, "failed to get ev certificates", err)
		return nil, err
	}

	queues = make([]string, 1, len(certificateFingerprints)+1)
	queues[0] = cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
	evQueueNamePrefix := cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix()
	for _, v := range certificateFingerprints {
		queues = append(queues, evQueueNamePrefix+v)
	}

	return
}

// 声明队列。
func declareWindowsQueues(ctx context.Context, queues []string) (err error) {
	for _, v := range queues {
		_, err = conn.RabbitMQClient(ctx).QueueDeclare(v, true, false, false, true, amqp.Table{})
		if err != nil {
			log.Error(ctx, "failed to declare mq queue", err, v)
			return err
		}
	}

	return
}

// 消费队列。
func consumeWindowsQueues(ctx context.Context, queues []string) (
	queueToConsumer map[string]<-chan amqp.Delivery, err error) {

	queueToConsumer = make(map[string]<-chan amqp.Delivery, len(queues))
	for _, queue := range queues {
		log.Info(ctx, "consume queue", queue)
		var deliveryChannel <-chan amqp.Delivery
		deliveryChannel, err = conn.RabbitMQClient(ctx).ConsumeWithContext(
			ctx, queue, queue+"_"+util.LocalIP, false, false, false, false, amqp.Table{})
		if err != nil {
			log.Error(ctx, "failed to consume queue", err, queue)
			return nil, err
		}
		queueToConsumer[queue] = deliveryChannel
	}

	return
}

// 执行签名。
func signWindowsFile(ctx context.Context, delivery amqp.Delivery) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "sign panic", p, util.GetStackCallers())
		}
	}()

	jobID := string(delivery.Body)
	log.Info(ctx, "delivery received", jobID)

	// 加锁，避免同时处理多个同一任务。
	windowsJobLock.Lock()
	defer windowsJobLock.Unlock()

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
	windowsSigningJob, err := httpBackendGetWindowsSigningJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "failed to get windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", jobID)
		return
	}
	if windowsSigningJob == nil {
		log.Error(ctx, "job not exists", jobID)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 判断任务状态。
	if !util.In(windowsSigningJob.Status,
		model.WindowsSigningJobStatusSigning, model.WindowsSigningJobStatusCabSigning) {
		log.Error(ctx, "invalid job status", windowsSigningJob.Status)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}

	// 校验出队次数。
	timesValue, _ := windowsJobIDs.Load(jobID)
	times, _ := timesValue.(int)
	if times > consts.MaximinOutQueueTimes {
		log.Error(ctx, "job exceeds maximin out queue times", times, jobID)
		failWindowsJob(ctx, jobID, "任务处理次数超过限制")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
		return
	}
	windowsJobIDs.Store(jobID, times+1)

	// 区分处理任务。
	switch windowsSigningJob.Type {
	case model.WindowsSigningJobTypePE:
		signWindowsFileForPEType(ctx, delivery, windowsSigningJob)
	case model.WindowsSigningJobTypePEAndAttestation:
		if windowsSigningJob.Status == model.WindowsSigningJobStatusSigning {
			signWindowsFileForPEAndAttestationType(ctx, delivery, windowsSigningJob)
		} else {
			signWindowsFileForPEAndAttestationType2(ctx, delivery, windowsSigningJob)
		}
	case model.WindowsSigningJobTypeAttestation:
		signWindowsFileForAttestationType(ctx, delivery, windowsSigningJob)
	case model.WindowsSigningJobTypeHLKX:
		signWindowsFileForHLKXType(ctx, delivery, windowsSigningJob)
	default:
		failWindowsJob(ctx, jobID, "未知的任务类型")
		log.Error(ctx, "invalid job type", windowsSigningJob.Type)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", jobID)
	}
}

// PE 文件签名。
func signWindowsFileForPEType(ctx context.Context, delivery amqp.Delivery, windowsSigningJob *model.WindowsSigningJob) {
	// 获取证书信息。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "获取证书信息")
	windowsCertificate, err := httpBackendGetCertificate(ctx, windowsSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	if windowsCertificate == nil {
		log.Error(ctx, "certificate not found", windowsSigningJob.CertificateID)
		failWindowsJob(ctx, windowsSigningJob.JobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(
		filepath.Join(cc.ServiceNameSigner, consts.ModeWindows, windowsSigningJob.JobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	var certificateFilePath string
	if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV) {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将证书写入硬盘")
		certificateFilePath = filepath.Join(workspace, windowsCertificate.CommonName+cc.ExtensionPFX)
		err = os.WriteFile(certificateFilePath, windowsCertificate.Content, cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write certificate to disk", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
	}

	// 下载待签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, windowsSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 文件签名。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名文件")
	var outputBytes []byte
	switch windowsCertificate.Type {
	case model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/f", certificateFilePath,
			"/p", windowsCertificate.Password,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command",
			strings.ReplaceAll(strings.Join(args, " "), windowsCertificate.Password, "******"))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	default:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/sha1", windowsCertificate.Sha1,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	if err != nil {
		log.Error(ctx, "failed to sign file", err, outputBytes)
		failWindowsJob(ctx, windowsSigningJob.JobID, "签名失败：%v，%s", err, outputBytes)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名输出：%s", outputBytes)

	// 校验签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验文件签名")
	args := []string{consts.SigntoolFilePath, "verify", "/tw", "/v", sourceFilePath}
	log.Info(ctx, "execute command", strings.Join(args, " "))
	outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		log.Warn(ctx, "failed to verify file", err, outputBytes)
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelError, "校验签名输出：%v，%s", err, outputBytes)
	} else {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验签名输出：%s", outputBytes)
	}

	// 上传结果文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	fileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeWindowsSigning, filepath.Base(sourceFilePath),
		windowsSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 更新任务。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "更新任务")
	err = httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:        windowsSigningJob.JobID,
		Status:       model.WindowsSigningJobStatusSuccess,
		SignedFileID: signedFileID,
		AppendLog:    formatJobLog(log.LevelInfo, "签名成功"),
		FinishedTime: bp.Time(time.Now()),
	})
	if err != nil {
		log.Error(ctx, "failed to update windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
	windowsJobIDs.Delete(windowsSigningJob.JobID)
	log.Info(ctx, "sign successfully")
}

// PE 文件签名。
func signWindowsFileForPEAndAttestationType(ctx context.Context, delivery amqp.Delivery,
	windowsSigningJob *model.WindowsSigningJob) {

	// 获取证书信息。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "获取证书信息")
	windowsCertificate, err := httpBackendGetCertificate(ctx, windowsSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	if windowsCertificate == nil {
		log.Error(ctx, "certificate not found", windowsSigningJob.CertificateID)
		failWindowsJob(ctx, windowsSigningJob.JobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(
		filepath.Join(cc.ServiceNameSigner, consts.ModeWindows, windowsSigningJob.JobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	var certificateFilePath string
	if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV) {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将证书写入硬盘")
		certificateFilePath = filepath.Join(workspace, windowsCertificate.CommonName+cc.ExtensionPFX)
		err = os.WriteFile(certificateFilePath, windowsCertificate.Content, cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write certificate to disk", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
	}

	// 下载待签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, windowsSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 文件签名。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名文件")
	var outputBytes []byte
	switch windowsCertificate.Type {
	case model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/f", certificateFilePath,
			"/p", windowsCertificate.Password,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command",
			strings.ReplaceAll(strings.Join(args, " "), windowsCertificate.Password, "******"))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	default:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/sha1", windowsCertificate.Sha1,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	if err != nil {
		log.Error(ctx, "failed to sign file", err, outputBytes)
		failWindowsJob(ctx, windowsSigningJob.JobID, "签名失败：%v，%s", err, outputBytes)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名输出：%s", outputBytes)

	// 校验签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验文件签名")
	args := []string{consts.SigntoolFilePath, "verify", "/tw", "/v", sourceFilePath}
	log.Info(ctx, "execute command", strings.Join(args, " "))
	outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		log.Warn(ctx, "failed to verify file", err, outputBytes)
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelError, "校验签名输出：%v，%s", err, outputBytes)
	} else {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验签名输出：%s", outputBytes)
	}

	// 将 sys 文件转成 cab。
	if filepath.Ext(sourceFilePath) == cc.ExtensionSYS {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将 sys 文件转成 cab")

		// 生成 inf 文件。TODO: 模板是否通用。
		fileName := filepath.Base(sourceFilePath)
		fileExt := filepath.Ext(fileName)
		fileName = strings.TrimSuffix(fileName, fileExt)
		infFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionINF)
		err = os.WriteFile(infFilePath, fmt.Appendf(nil, consts.INFTemplate, fileName), cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write inf file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}

		// 获取 PE 头。TODO: 需优化实现。
		var fileStream *os.File
		fileStream, err = os.Open(sourceFilePath)
		if err != nil {
			log.Error(ctx, "failed to open file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		data := make([]byte, 1024)
		var n int
		n, err = io.ReadFull(fileStream, data)
		util.CloseIO(ctx, fileStream)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Error(ctx, "failed to read file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		data = data[:n]
		hits := findPEHeader.FindAllString(string(data), 1)
		osValue := "10_x64"
		if len(hits) > 0 && hits[0] == "L" {
			osValue = "10_X86"
		}

		// 运行 inf2Cat.exe。
		outputBytes, err = exec.Command(consts.Inf2CatFilePath, "/driver:"+workspace, "/os:"+osValue).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to exec inf2cat.exe", err, outputBytes)
			failWindowsJob(ctx, windowsSigningJob.JobID, "运行 inf2Cat.exe 失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		log.Info(ctx, "inf2cat.exe output", outputBytes)

		// 生成 ddf。TODO: 模板是否通用。
		ddfFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionDDF)
		catFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionCAT)
		err = os.WriteFile(ddfFilePath, fmt.Appendf(
			nil, consts.DDFTemplate, workspace, fileName, catFilePath, infFilePath, sourceFilePath), cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write ddf file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}

		// 运行 makecab.exe。
		outputBytes, err = exec.Command(consts.MakeecabFilePath, "/f", ddfFilePath, "/L", workspace).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to exec makecab.exe", err, outputBytes)
			failWindowsJob(ctx, windowsSigningJob.JobID, "运行 makecab.exe 失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		log.Info(ctx, "makecab.exe output", outputBytes)

		// 更新结果文件路径。
		sourceFilePath = filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionCAB)
	}

	// 上传结果文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	fileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeWindowsSigning, filepath.Base(sourceFilePath),
		windowsSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 更新任务。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "更新任务")
	err = httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:              windowsSigningJob.JobID,
		Status:             model.WindowsSigningJobStatusWaitCabSign,
		SignedFileID:       signedFileID,
		AppendLog:          formatJobLog(log.LevelInfo, "签名成功"),
		FinishedPESignTime: bp.Time(time.Now()),
	})
	if err != nil {
		log.Error(ctx, "failed to update windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
	windowsJobIDs.Delete(windowsSigningJob.JobID)
	log.Info(ctx, "sign successfully")
}

// cab 文件签名。
func signWindowsFileForPEAndAttestationType2(ctx context.Context, delivery amqp.Delivery,
	windowsSigningJob *model.WindowsSigningJob) {

	// 获取证书信息。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "获取证书信息")
	windowsCertificate, err := httpBackendGetCertificate(ctx, windowsSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	if windowsCertificate == nil {
		log.Error(ctx, "certificate not found", windowsSigningJob.CertificateID)
		failWindowsJob(ctx, windowsSigningJob.JobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(
		filepath.Join(cc.ServiceNameSigner, consts.ModeWindows, windowsSigningJob.JobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	var certificateFilePath string
	if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV) {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将证书写入硬盘")
		certificateFilePath = filepath.Join(workspace, windowsCertificate.CommonName+cc.ExtensionPFX)
		err = os.WriteFile(certificateFilePath, windowsCertificate.Content, cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write certificate to disk", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
	}

	// 下载待签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, windowsSigningJob.SignedFileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 文件签名。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名文件")
	var outputBytes []byte
	switch windowsCertificate.Type {
	case model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/f", certificateFilePath,
			"/p", windowsCertificate.Password,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command",
			strings.ReplaceAll(strings.Join(args, " "), windowsCertificate.Password, "******"))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	default:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/sha1", windowsCertificate.Sha1,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	if err != nil {
		log.Error(ctx, "failed to sign file", err, outputBytes)
		failWindowsJob(ctx, windowsSigningJob.JobID, "签名失败：%v，%s", err, outputBytes)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名输出：%s", outputBytes)

	// 校验签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验文件签名")
	args := []string{consts.SigntoolFilePath, "verify", "/tw", "/v", sourceFilePath}
	log.Info(ctx, "execute command", strings.Join(args, " "))
	outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		log.Warn(ctx, "failed to verify file", err, outputBytes)
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelError, "校验签名输出：%v，%s", err, outputBytes)
	} else {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验签名输出：%s", outputBytes)
	}

	// 上传结果文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	fileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeWindowsSigning, filepath.Base(sourceFilePath),
		windowsSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 更新任务。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "更新任务")
	err = httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:        windowsSigningJob.JobID,
		Status:       model.WindowsSigningJobStatusAttestationWaiting,
		SignedFileID: signedFileID,
		AppendLog:    formatJobLog(log.LevelInfo, "签名成功"),
	})
	if err != nil {
		log.Error(ctx, "failed to update windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
	windowsJobIDs.Delete(windowsSigningJob.JobID)
	log.Info(ctx, "sign successfully")
}

// cab 文件签名。
func signWindowsFileForAttestationType(ctx context.Context, delivery amqp.Delivery,
	windowsSigningJob *model.WindowsSigningJob) {

	// 获取证书信息。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "获取证书信息")
	windowsCertificate, err := httpBackendGetCertificate(ctx, windowsSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	if windowsCertificate == nil {
		log.Error(ctx, "certificate not found", windowsSigningJob.CertificateID)
		failWindowsJob(ctx, windowsSigningJob.JobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(
		filepath.Join(cc.ServiceNameSigner, consts.ModeWindows, windowsSigningJob.JobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	var certificateFilePath string
	if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV) {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将证书写入硬盘")
		certificateFilePath = filepath.Join(workspace, windowsCertificate.CommonName+cc.ExtensionPFX)
		err = os.WriteFile(certificateFilePath, windowsCertificate.Content, cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write certificate to disk", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
	}

	// 下载待签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, windowsSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 将 sys 文件转成 cab。
	if filepath.Ext(sourceFilePath) == cc.ExtensionSYS {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将 sys 文件转成 cab")

		// 生成 inf 文件。TODO: 模板是否通用。
		fileName := filepath.Base(sourceFilePath)
		fileExt := filepath.Ext(fileName)
		fileName = strings.TrimSuffix(fileName, fileExt)
		infFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionINF)
		err = os.WriteFile(infFilePath, fmt.Appendf(nil, consts.INFTemplate, fileName), cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write inf file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}

		// 获取 PE 头。TODO: 需优化实现。
		var fileStream *os.File
		fileStream, err = os.Open(sourceFilePath)
		if err != nil {
			log.Error(ctx, "failed to open file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		data := make([]byte, 1024)
		var n int
		n, err = io.ReadFull(fileStream, data)
		util.CloseIO(ctx, fileStream)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Error(ctx, "failed to read file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		data = data[:n]
		hits := findPEHeader.FindAllString(string(data), 1)
		osValue := "10_x64"
		if len(hits) > 0 && hits[0] == "L" {
			osValue = "10_X86"
		}

		// 运行 inf2Cat.exe。
		var outputBytes []byte
		outputBytes, err = exec.Command(consts.Inf2CatFilePath, "/driver:"+workspace, "/os:"+osValue).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to exec inf2cat.exe", err, outputBytes)
			failWindowsJob(ctx, windowsSigningJob.JobID, "运行 inf2Cat.exe 失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		log.Info(ctx, "inf2cat.exe output", outputBytes)

		// 生成 ddf。TODO: 模板是否通用。
		ddfFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionDDF)
		catFilePath := filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionCAT)
		err = os.WriteFile(ddfFilePath, fmt.Appendf(
			nil, consts.DDFTemplate, workspace, fileName, catFilePath, infFilePath, sourceFilePath), cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write ddf file", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}

		// 运行 makecab.exe。
		outputBytes, err = exec.Command(consts.MakeecabFilePath, "/f", ddfFilePath, "/L", workspace).CombinedOutput()
		if err != nil {
			log.Error(ctx, "failed to exec makecab.exe", err, outputBytes)
			failWindowsJob(ctx, windowsSigningJob.JobID, "运行 makecab.exe 失败：%v，%s", err, outputBytes)
			log.ErrorIf(ctx, delivery.Ack(false), "failed to nack message", windowsSigningJob.JobID)
			return
		}
		log.Info(ctx, "makecab.exe output", outputBytes)

		// 更新结果文件路径。
		sourceFilePath = filepath.Join(filepath.Dir(sourceFilePath), fileName+cc.ExtensionCAB)
	}

	// 文件签名。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名文件")
	var outputBytes []byte
	switch windowsCertificate.Type {
	case model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/f", certificateFilePath,
			"/p", windowsCertificate.Password,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command",
			strings.ReplaceAll(strings.Join(args, " "), windowsCertificate.Password, "******"))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	default:
		args := []string{
			consts.SigntoolFilePath, "sign",
			"/sha1", windowsCertificate.Sha1,
			"/fd", "sha256",
			"/td", "sha256",
			"/tr", consts.DefaultTimeServer,
			"/as", "/v", sourceFilePath,
		}
		log.Info(ctx, "execute command", strings.Join(args, " "))
		outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	if err != nil {
		log.Error(ctx, "failed to sign file", err, outputBytes)
		failWindowsJob(ctx, windowsSigningJob.JobID, "签名失败：%v，%s", err, outputBytes)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名输出：%s", outputBytes)

	// 校验签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验文件签名")
	args := []string{consts.SigntoolFilePath, "verify", "/tw", "/v", sourceFilePath}
	log.Info(ctx, "execute command", strings.Join(args, " "))
	outputBytes, err = exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		log.Warn(ctx, "failed to verify file", err, outputBytes)
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelError, "校验签名输出：%v，%s", err, outputBytes)
	} else {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "校验签名输出：%s", outputBytes)
	}

	// 上传结果文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	fileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeWindowsSigning, filepath.Base(sourceFilePath),
		windowsSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 更新任务。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "更新任务")
	err = httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:              windowsSigningJob.JobID,
		Status:             model.WindowsSigningJobStatusAttestationWaiting,
		SignedFileID:       signedFileID,
		AppendLog:          formatJobLog(log.LevelInfo, "签名成功"),
		FinishedPESignTime: bp.Time(time.Now()),
	})
	if err != nil {
		log.Error(ctx, "failed to update windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
	windowsJobIDs.Delete(windowsSigningJob.JobID)
	log.Info(ctx, "sign successfully")
}

// hlkx 文件签名。
func signWindowsFileForHLKXType(ctx context.Context, delivery amqp.Delivery,
	windowsSigningJob *model.WindowsSigningJob) {

	// 获取证书信息。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "获取证书信息")
	windowsCertificate, err := httpBackendGetCertificate(ctx, windowsSigningJob.CertificateID)
	if err != nil {
		log.Error(ctx, "failed to get windows certificate", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	if windowsCertificate == nil {
		log.Error(ctx, "certificate not found", windowsSigningJob.CertificateID)
		failWindowsJob(ctx, windowsSigningJob.JobID, "证书未找到")
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	// 准备工作区。
	workspace, err := util.GenerateTemporaryDirectory(
		filepath.Join(cc.ServiceNameSigner, consts.ModeWindows, windowsSigningJob.JobID))
	if err != nil {
		log.Error(ctx, "failed to generate temporary directory", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 清理工作区。
	defer util.RemoveDirectory(ctx, workspace)

	// 将证书写入磁盘。
	var certificateFilePath string
	if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalOV, model.WindowsCertificateTypeCompanyOV) {
		appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "将证书写入硬盘")
		certificateFilePath = filepath.Join(workspace, windowsCertificate.CommonName+cc.ExtensionPFX)
		err = os.WriteFile(certificateFilePath, windowsCertificate.Content, cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write certificate to disk", err)
			log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
			return
		}
	}

	// 下载待签名文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "下载待签名文件")
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, windowsSigningJob.FileID, workspace)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 文件签名。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名文件")
	args := []string{consts.WinevsignerFilePath, windowsCertificate.Sha1, sourceFilePath}
	log.Info(ctx, "execute command", strings.Join(args, " "))
	outputBytes, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		log.Error(ctx, "failed to sign file", err, outputBytes)
		failWindowsJob(ctx, windowsSigningJob.JobID, "签名失败：%v，%s", err, outputBytes)
		log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
		return
	}
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "签名输出：%s", outputBytes)

	// 上传结果文件。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "上传结果文件")
	fileStream, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to open file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	fileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}
	signedFileID, err := httpBackendUploadFile(ctx, model.FileTypeWindowsSigning, filepath.Base(sourceFilePath),
		windowsSigningJob.AppID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload file", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to nack message", windowsSigningJob.JobID)
		return
	}

	// 更新任务。
	appendWindowsJobLog(ctx, windowsSigningJob.JobID, log.LevelInfo, "更新任务")
	err = httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:        windowsSigningJob.JobID,
		Status:       model.WindowsSigningJobStatusSuccess,
		SignedFileID: signedFileID,
		AppendLog:    formatJobLog(log.LevelInfo, "签名成功"),
		FinishedTime: bp.Time(time.Now()),
	})
	if err != nil {
		log.Error(ctx, "failed to update windows signing job", err)
		log.ErrorIf(ctx, delivery.Nack(false, true), "failed to ack message", windowsSigningJob.JobID)
		return
	}

	log.ErrorIf(ctx, delivery.Ack(false), "failed to ack message", windowsSigningJob.JobID)
	windowsJobIDs.Delete(windowsSigningJob.JobID)
	log.Info(ctx, "sign successfully")
}

// 更新任务为失败状态。
func failWindowsJob(ctx context.Context, jobID string, format string, args ...any) {
	err := httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:        jobID,
		Status:       model.WindowsSigningJobStatusFailure,
		AppendLog:    formatJobLog(log.LevelError, format, args...),
		FinishedTime: bp.Time(time.Now()),
	})
	log.ErrorIf(ctx, err, "failed to update windows signing job")
}

// 打印任务日志。
func appendWindowsJobLog(ctx context.Context, jobID string, level log.Level, format string, args ...any) {
	err := httpBackendUpdateWindowsSigningJob(ctx, &bp.WindowsInternalUpdateSigningJobReq{
		JobID:     jobID,
		AppendLog: formatJobLog(level, format, args...),
	})
	log.ErrorIf(ctx, err, "failed to append windows signing job log", jobID)
}
