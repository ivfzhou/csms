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
	"debug/pe"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

// CronSubmitCabFileSigning 提交 cab 文件签名。
func CronSubmitCabFileSigning(ctx context.Context, _ string, _ time.Time) {
	// 获取需要提交签名的签名任务。
	log.Debug(ctx, "get windows signing jobs that waiting cab signing")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusWaitCabSign)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no cab signing job need to start")
		return
	}

	// 获取签名证书。
	log.Info(ctx, "find hlkx signing certificate from database")
	windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
	var windowsCertificate *model.WindowsCertificate
	windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
		windowsCertificateDo.Sha1,
	).Where(
		windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
		windowsCertificateDo.DeletedTime.IsNull(),
		windowsCertificateDo.NotAfter.Gte(time.Now()),
	).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "no windows certificate available")
		}
		return
	}

	// 签名队列。
	queue := cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + strings.ToUpper(windowsCertificate.Sha1)

	// 发送 mq 消息。
	for _, v := range windowsSignJobIDs {
		err = publishMessageToQueue(ctx, queue, []byte(v))
		if err != nil {
			continue
		}

		// 更新任务状态。
		windowsSigningJobDo2 := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(v))
		_, err = windowsSigningJobDo2.WithContext(ctx).Where(
			windowsSigningJobDo2.JobID.Eq(v),
			windowsSigningJobDo2.Status.Eq(model.WindowsSigningJobStatusWaitCabSign),
		).UpdateColumnSimple(
			windowsSigningJobDo2.Status.Value(model.WindowsSigningJobStatusCabSigning),
		)
		if err != nil {
			log.Error(ctx, "failed to update windows signing job", v)
			continue
		}
	}
}

// CronStartAttestationJobs 提交 Windows 签名任务中的 Attestation 签名。
func CronStartAttestationJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要提交 Attestation 签名的 Windows 签名任务。
	log.Debug(ctx, "get all windows sign job table names")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	log.Debug(ctx, "get windows sign jobs that require attestation signing from database")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusAttestationWaiting)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no attestation job need to start")
		return
	}

	// 处理任务。
	for _, v := range windowsSignJobIDs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock windows sign job", v)
			lockKey := fmt.Sprintf(consts.RedisKeyAttestationStartLockFmt, v)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire attestation lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire attestation lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release attestation lock")
				if !success {
					log.Error(ctx, "releasing attestation lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start attestation signing job", v)
			startAttestationJob(childCtx, v)
		}()
	}
}

// CronCheckAttestationJobsResult 检查 Attestation 签名结果。
func CronCheckAttestationJobsResult(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 Attestation 签名结果的任务。
	log.Debug(ctx, "get all windows sign job table names")
	tableNames, err := getAllWindowsSignJobTableNames(ctx)
	if err != nil {
		return
	}
	log.Debug(ctx, "get windows sign jobs that waiting attestation signing result from database")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob
	windowsSignJobIDs, err := windowsSigningJobDo.WithContext(ctx).
		GetJobIDByStatus(tableNames, model.WindowsSigningJobStatusAttestationSigning)
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		return
	}
	if len(windowsSignJobIDs) <= 0 {
		log.Debug(ctx, "no attestation job need to check")
		return
	}

	// 处理任务。
	for _, v := range windowsSignJobIDs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock windows sign job", v)
			lockKey := fmt.Sprintf(consts.RedisKeyAttestationCheckLockFmt, v)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire attestation lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire attestation lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release attestation lock")
				if !success {
					log.Error(ctx, "releasing attestation lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start attestation signing job", v)
			checkAttestationJobResult(childCtx, v)
		}()
	}
}

// CronSubmitHLKXFileSigningJobs 将 whql_job 任务，通过 hlk 测试的 hlkx 文件提交签名。
func CronSubmitHLKXFileSigningJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 WHQL 签名结果的任务。
	log.Debug(ctx, "get windows whql jobs that waiting hlk signing")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.UserID,
		whqlJobDo.AppID,
		whqlJobDo.HlkxFileID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusFinishTest),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to check")
		return
	}

	// 查找签名证书。
	log.Info(ctx, "find hlkx signing certificate from database")
	windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
	var windowsCertificate *model.WindowsCertificate
	windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Select(
		windowsCertificateDo.ID,
		windowsCertificateDo.Type,
		windowsCertificateDo.Sha1,
	).Where(
		windowsCertificateDo.IsMicrosoftVerifyCertificate.Eq(model.Bool(true)),
		windowsCertificateDo.DeletedTime.IsNull(),
		windowsCertificateDo.NotAfter.Gte(time.Now()),
	).Order(windowsCertificateDo.NotAfter.Desc()).Limit(1).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "no windows certificate available")
		}
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			innerCtx := ctxs.New()
			log.Info(innerCtx, "create hlkx signing job", v.JobID)

			// 生成任务 ID。
			var jobID string
			jobID, err = generateID(innerCtx, IDWindowsJob)
			if err != nil {
				return
			}
			defer func() {
				if err != nil {
					log.ErrorIf(innerCtx, reclaimID(innerCtx, IDWindowsJob, jobID),
						"failed to reclaim windows sign job id", jobID)
				}
			}()

			// 开启事务。
			conn.BeginMySQLTx(innerCtx)
			defer func() {
				if err == nil {
					log.ErrorIf(innerCtx, conn.CommitMySQLTx(innerCtx), "commit tx failed")
				} else {
					_ = conn.RollbackMySQLTx(innerCtx)
				}
			}()

			// 创建任务。
			windowsSigningJob := &model.WindowsSigningJob{
				AppID:         v.AppID,
				JobID:         jobID,
				Type:          model.WindowsSigningJobTypeHLKX,
				UserID:        v.UserID,
				CertificateID: windowsCertificate.ID,
				FileID:        v.HlkxFileID,
				Source:        model.SourceInternal,
				Status:        model.WindowsSigningJobStatusSigning,
				CreatedTime:   time.Now(),
			}
			if err = createWindowsSignJob(innerCtx, windowsSigningJob); err != nil {
				log.Error(ctx, "failed to create windows sign job to database", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}

			// 更新 whql 任务。
			whqlJobDo = conn.MySQLTxClient(innerCtx).WhqlJob
			_, err = whqlJobDo.WithContext(innerCtx).Where(
				whqlJobDo.ID.Eq(v.ID),
			).UpdateColumnSimple(
				whqlJobDo.HlkxSignJobID.Value(jobID),
				whqlJobDo.Status.Value(model.WHQLJobStatusHLKXFileSinging),
			)
			if err != nil {
				log.Error(ctx, "failed to update whql job in database", err)
				return
			}

			// 发送任务消息。
			queue := cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
			if util.In(windowsCertificate.Type, model.WindowsCertificateTypePersonalEV,
				model.WindowsCertificateTypeCompanyEV) {
				queue = cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix() + windowsCertificate.Sha1
			}
			err = publishMessageToQueue(ctx, queue, []byte(jobID))
			if err != nil {
				return
			}
		}()
	}
}

// CronStartWHQLJobs 提交 WHQL 签名。
func CronStartWHQLJobs(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要提交 WHQL 签名的任务。
	log.Debug(ctx, "get whql job ids from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.JobID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusHLKXFileSinging),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to start")
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock whql job", v.JobID)
			lockKey := fmt.Sprintf(consts.RedisKeyWHQLStartLockFmt, v.JobID)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire whql lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire whql lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release whql lock")
				if !success {
					log.Error(ctx, "releasing whql lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start whql signing job", v.JobID, ctxs.RequestID(childCtx))
			startWHQLJob(childCtx, v.JobID)
		}()
	}
}

// CronCheckWHQLJobsResult 检查 WHQL 签名结果。
func CronCheckWHQLJobsResult(ctx context.Context, _ string, _ time.Time) {
	// 从数据库中，查出所有需要检查 WHQL 签名结果的任务。
	log.Debug(ctx, "get windows whql jobs that waiting microsoft signing result from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJobs, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.JobID,
	).Where(
		whqlJobDo.Status.Eq(model.WHQLJobStatusWHQLSigning),
	).Find()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows whql job information from database", err)
		return
	}
	if len(whqlJobs) <= 0 {
		log.Debug(ctx, "no whql job need to check")
		return
	}

	// 处理任务。
	for _, v := range whqlJobs {
		func() {
			// 单个任务加锁，避免同时处理。
			log.Info(ctx, "lock whql sign job", v.JobID)
			lockKey := fmt.Sprintf(consts.RedisKeyWHQLCheckLockFmt, v.JobID)
			var success bool
			success, err = conn.RedisLock(ctx, lockKey, 0, 30*time.Minute)
			if err != nil {
				log.Error(ctx, "failed to acquire whql lock", err)
				return
			}
			if !success {
				log.Info(ctx, "failed to acquire whql lock")
				return
			}
			defer func() {
				success, err = conn.RedisUnlock(ctx, lockKey)
				log.ErrorIf(ctx, err, "failed to release whql lock")
				if !success {
					log.Error(ctx, "releasing whql lock has no effect")
				}
			}()

			childCtx := ctxs.New()
			log.Info(ctx, "start checking whql result", v, ctxs.RequestID(childCtx))
			checkWHQLJobResult(childCtx, v.JobID)
		}()
	}
}

func startAttestationJob(ctx context.Context, jobID string) {
	// 从数据中，获取任务信息。
	log.Info(ctx, "get windows sign job information from database", jobID)
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(jobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.ID,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.AppID,
		windowsSigningJobDo.JobID,
		windowsSigningJobDo.FileID,
		windowsSigningJobDo.Type,
		windowsSigningJobDo.SignedFileID,
		windowsSigningJobDo.CreatedTime,
	).Where(
		windowsSigningJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || windowsSigningJob == nil {
		log.Error(ctx, "windows sign job not found", err)
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify attestation job state")
	if windowsSigningJob.Status != model.WindowsSigningJobStatusAttestationWaiting {
		log.Warn(ctx, "windows sign job attestation status is not waiting")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	now := time.Now()
	if now.Sub(windowsSigningJob.CreatedTime) > cfg.Get().Backend().MaximumAttestationJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 查询数据库，获取文件信息。
	log.Info(ctx, "get file information from database")
	fileID := windowsSigningJob.SignedFileID
	fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(fileID))
	fileInfo, err := fileDo.WithContext(ctx).Select(
		fileDo.Name,
		fileDo.TusdID,
	).Where(
		fileDo.FileID.Eq(fileID),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "file not found", fileID)
			_, err = windowsSigningJobDo.WithContext(ctx).Where(
				windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
			).UpdateColumnSimple(
				windowsSigningJobDo.FinishedTime.Value(now),
				windowsSigningJobDo.UpdatedTime.Value(now),
				windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
				query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "待处理文件未找到")),
			)
			if err != nil {
				log.Error(ctx, "failed to update attestation job state", err)
			}
			return
		}
		return
	}

	// 下载文件到外存。
	log.Info(ctx, "download file to disk")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend,
		"attestation_signing_*"+filepath.Ext(fileInfo.Name))
	if err != nil {
		log.Error(ctx, "create a file path failed", err)
		return
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, fileInfo.TusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 从 cab 中提取 sys 文件。
	log.Info(ctx, "extract cab file", filePath)
	destinationDirectoryPath := filepath.Join(filepath.Dir(filePath),
		strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)))
	err = extractCab(filePath, destinationDirectoryPath)
	if err != nil {
		log.Error(ctx, "failed to extract cab file", err)
		return
	}
	defer util.RemoveDirectory(ctx, destinationDirectoryPath)

	// 校验 PE 文件格式。
	log.Info(ctx, "parse pe file machine type")
	peTyp := PETypeAmd64
	entries, err := os.ReadDir(destinationDirectoryPath)
	if err != nil {
		log.Error(ctx, "failed to read directory", err, destinationDirectoryPath)
		return
	}
	for _, v := range entries {
		if !strings.HasSuffix(strings.ToLower(v.Name()), cc.ExtensionSYS) {
			continue
		}
		func(filePath string) {
			var fileObj *pe.File
			fileObj, err = pe.Open(filePath)
			if err != nil {
				log.Error(ctx, "failed to open pe file", err)
				return
			}
			defer util.CloseIO(ctx, fileObj)
			switch fileObj.Machine {
			case pe.IMAGE_FILE_MACHINE_AMD64, pe.IMAGE_FILE_MACHINE_IA64:
				peTyp = PETypeAmd64
			case pe.IMAGE_FILE_MACHINE_ARM64, 0xA64E, 0xA641:
				peTyp = PETypeArm64
			case pe.IMAGE_FILE_MACHINE_I386, pe.IMAGE_FILE_MACHINE_LOONGARCH32, pe.IMAGE_FILE_MACHINE_RISCV32,
				0x160, 0x162, 0x184:
				peTyp = PETypeAmd32
			case pe.IMAGE_FILE_MACHINE_ARM:
				peTyp = PETypeArm32
			default:
				log.Warn(ctx, "unhandled pe machine", strconv.FormatUint(uint64(fileObj.Machine), 16))
			}
		}(filepath.Join(destinationDirectoryPath, v.Name()))
		if err != nil {
			log.Error(ctx, "parse sys file error", err)
			continue
		}
		break
	}
	log.Info(ctx, "machine code of sys file is", peTyp)

	// 创建微软产品。
	log.Info(ctx, "create microsoft product")
	productID, err := createAttestationProduct(ctx, strconv.Itoa(windowsSigningJob.AppID), peTyp)
	if err != nil {
		return
	}

	// 创建微软提交。
	log.Info(ctx, "create microsoft product's submission")
	submissionID, uploadURL, err := createMicrosoftSubmission(ctx, productID, windowsSigningJob.JobID)
	if err != nil {
		return
	}

	// 上传文件。
	log.Info(ctx, "upload file to microsoft server")
	err = uploadFileToMicrosoft(ctx, filePath, uploadURL)
	if err != nil {
		return
	}

	// 提交微软提交。
	log.Info(ctx, "commit a microsoft submission")
	err = commitMicrosoftSubmission(ctx, productID, submissionID)
	if err != nil {
		return
	}

	// 更新数据库中的任务进展信息。
	log.Info(ctx, "update windows sign job information in database")
	now = time.Now()
	_, err = windowsSigningJobDo.WithContext(ctx).Where(
		windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
	).UpdateColumnSimple(
		windowsSigningJobDo.UpdatedTime.Value(now),
		windowsSigningJobDo.ProductID.Value(productID),
		windowsSigningJobDo.SubmissionID.Value(submissionID),
		windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusAttestationSigning),
		query.Concat(windowsSigningJobDo.Log,
			formatJobLog(log.LevelInfo, "提交至微软方签名，PID：%s，SID：%s", productID, submissionID)),
	)
	if err != nil {
		log.Error(ctx, "failed to update windows sign job in database", err)
		return
	}
}

func checkAttestationJobResult(ctx context.Context, jobID string) {
	// 查询数据库，获取任务信息。
	log.Info(ctx, "get windows sign job information from database", jobID)
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(jobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.ID,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.UpdatedTime,
		windowsSigningJobDo.CreatedTime,
		windowsSigningJobDo.ProductID,
		windowsSigningJobDo.SubmissionID,
		windowsSigningJobDo.AppID,
	).Where(
		windowsSigningJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || windowsSigningJob == nil {
		log.Error(ctx, "windows sign job not found")
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify attestation job state")
	if windowsSigningJob.Status != model.WindowsSigningJobStatusAttestationSigning {
		log.Warn(ctx, "windows sign job attestation status is not signing")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	beginTime := windowsSigningJob.UpdatedTime
	if beginTime.IsZero() {
		beginTime = windowsSigningJob.CreatedTime
	}
	now := time.Now()
	if now.Sub(beginTime) > cfg.Get().Backend().MaximumAttestationJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 查询微软方签名结果。
	log.Info(ctx, "query submission state from microsoft")
	finished, downloadURL, err := queryMicrosoftSubmission(
		ctx, windowsSigningJob.ProductID, windowsSigningJob.SubmissionID)
	if !finished {
		log.Warn(ctx, "windows sign job attestation status is not finished")
		return
	}

	// 处理签名结果。
	log.Info(ctx, "deal with signing result")
	now = time.Now()
	if err != nil {
		// 更新数据库任务状态。
		log.Warn(ctx, "microsoft attestation job failed", err)
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
		).UpdateColumnSimple(
			windowsSigningJobDo.FinishedTime.Value(now),
			windowsSigningJobDo.UpdatedTime.Value(now),
			windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusFailure),
			query.Concat(windowsSigningJobDo.Log,
				formatJobLog(log.LevelError, "微软 Attestation 签名失败：%v", errs.String(err))),
		)
		if err != nil {
			log.Error(ctx, "failed to update attestation job state", err)
		}
		return
	}

	// 下载签名结果文件。
	log.Info(ctx, "download signing result file")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "attestation_result_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return
	}
	httpCode, fileName, fileSize, fileMD5, err := util.HTTPGetToDisk(ctx, downloadURL, filePath)
	if err != nil || !(httpCode >= http.StatusOK && httpCode < http.StatusMultipleChoices) {
		log.Error(ctx, "failed to download signing result file", err, httpCode)
		return
	}

	// 将文件上传到 Tusd。
	log.Info(ctx, "upload signing result file to tusd")
	tusdID, err := conn.TusdClient(ctx).MultipleUploadFromFile(ctx, filePath)
	if err != nil {
		log.Error(ctx, "failed to upload signing result file", err)
		return
	}
	defer func() {
		// 若失败，删除文件。
		if err != nil {
			log.ErrorIf(ctx, conn.TusdClient(ctx).DeleteFile(ctx, tusdID), "failed to delete tusd file", tusdID)
		}
	}()

	// 保存文件信息到数据库中。
	fileID, err := generateID(ctx, IDFile)
	if err != nil {
		log.Error(ctx, "failed to generate file id", err)
		return
	}
	defer func() {
		// 失败回收 ID。
		if err != nil {
			log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id", fileID)
		}
	}()
	now = time.Now()
	err = createFile(ctx, &model.File{
		FileID:      fileID,
		TusdID:      tusdID,
		AppID:       windowsSigningJob.AppID,
		Name:        fileName,
		Md5:         fileMD5,
		Size:        int(fileSize),
		Type:        model.FileTypeMicrosoftSigning,
		CreatedTime: now,
	})
	if err != nil {
		return
	}

	// 更新数据库中任务状态。
	log.Info(ctx, "update windows sign job information in database")
	_, err = windowsSigningJobDo.WithContext(ctx).Where(
		windowsSigningJobDo.ID.Eq(windowsSigningJob.ID),
	).UpdateColumnSimple(
		windowsSigningJobDo.UpdatedTime.Value(now),
		windowsSigningJobDo.FinishedTime.Value(now),
		windowsSigningJobDo.SignedFileID.Value(fileID),
		windowsSigningJobDo.Status.Value(model.WindowsSigningJobStatusSuccess),
		query.Concat(windowsSigningJobDo.Log, formatJobLog(log.LevelInfo, "微软方签名成功")),
	)
	if err != nil {
		log.Error(ctx, "failed to update windows sign job in database", err)
		return
	}
}

func startWHQLJob(ctx context.Context, jobID string) {
	// 从数据中，获取任务信息。
	log.Info(ctx, "get whql job information from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJob, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.Status,
		whqlJobDo.CreatedTime,
		whqlJobDo.HlkxSignJobID,
	).Where(
		whqlJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil {
		log.Error(ctx, "whql job not found")
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify whql job state")
	if whqlJob.Status != model.WHQLJobStatusHLKXFileSinging {
		log.Warn(ctx, "whql status is not in waiting")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify attestation job time consuming")
	if time.Since(whqlJob.CreatedTime) > cfg.Get().Backend().MaximumWHQLJobInterval() {
		log.Warn(ctx, "attestation job has exceeded deadline")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "任务超时终止")),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 查询 HLKX 文件签名结果。
	log.Info(ctx, "get hlkx signing job")
	windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.
		Table(model.GetWindowsSigningJobByID(whqlJob.HlkxSignJobID))
	windowsSigningJob, err := windowsSigningJobDo.WithContext(ctx).Select(
		windowsSigningJobDo.Log,
		windowsSigningJobDo.Status,
		windowsSigningJobDo.JobID,
		windowsSigningJobDo.SignedFileID,
		windowsSigningJobDo.AppID,
	).Where(
		windowsSigningJobDo.JobID.Eq(whqlJob.HlkxSignJobID),
	).Take()
	if err != nil {
		log.Error(ctx, "failed to retrieve windows sign job information from database", err)
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名信息未找到")),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 判断文件签名结果。
	log.Info(ctx, "check hlkx signing job result")
	switch windowsSigningJob.Status {
	case model.WindowsSigningJobStatusSuccess:
		log.Info(ctx, "hlkx signing job pass")
	case model.WindowsSigningJobStatusFailure:
		log.Error(ctx, "hlkx file signing failed")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名失败：%s", windowsSigningJob.Log)),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	default:
		log.Warn(ctx, "hlkx signing job running", jobID)
		return
	}

	// 查询数据库，获取文件信息。
	log.Info(ctx, "get file information from database")
	fileDo := conn.MySQLClient(ctx).File.Table(model.GetFileTableNameByID(windowsSigningJob.SignedFileID))
	file, err := fileDo.WithContext(ctx).Select(
		fileDo.Name,
		fileDo.TusdID,
	).Where(
		fileDo.FileID.Eq(windowsSigningJob.SignedFileID),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "file not found", windowsSigningJob.SignedFileID)
			_, err = whqlJobDo.WithContext(ctx).Where(
				whqlJobDo.ID.Eq(whqlJob.ID),
			).UpdateColumnSimple(
				whqlJobDo.FinishedTime.Value(time.Now()),
				whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
				query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "HLKX 文件签名结果文件未找到")),
			)
			if err != nil {
				log.Error(ctx, "failed to update whql job state", err)
			}
			return
		}
		return
	}

	// 下载文件到外存。
	log.Info(ctx, "download file to disk")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "whql_signing_*."+filepath.Ext(file.Name))
	if err != nil {
		log.Error(ctx, "create a file path failed", err)
		return
	}
	err = conn.TusdClient(ctx).DownloadToFile(ctx, file.TusdID, filePath)
	if err != nil {
		log.Error(ctx, "failed to download file", err)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 创建微软产品。
	log.Info(ctx, "create microsoft product")
	productID, err := createWHQLProduct(ctx, strconv.Itoa(windowsSigningJob.AppID))
	if err != nil {
		return
	}

	// 创建微软提交。
	log.Info(ctx, "create microsoft product's submission")
	submissionID, uploadURL, err := createMicrosoftSubmission(ctx, productID, windowsSigningJob.JobID)
	if err != nil {
		return
	}

	// 上传文件。
	log.Info(ctx, "upload file to microsoft server")
	err = uploadFileToMicrosoft(ctx, filePath, uploadURL)
	if err != nil {
		return
	}

	// 提交微软提交。
	log.Info(ctx, "commit a microsoft submission")
	err = commitMicrosoftSubmission(ctx, productID, submissionID)
	if err != nil {
		return
	}

	// 更新数据库中的任务进展信息。
	log.Info(ctx, "update whql job information in database")
	_, err = whqlJobDo.WithContext(ctx).Where(
		whqlJobDo.ID.Eq(whqlJob.ID),
	).UpdateColumnSimple(
		whqlJobDo.ProductID.Value(productID),
		whqlJobDo.SubmissionID.Value(submissionID),
		whqlJobDo.Status.Value(model.WHQLJobStatusWHQLSigning),
		query.Concat(whqlJobDo.Log,
			formatJobLog(log.LevelInfo, "提交至微软方签名，PID：%s，SID：%s", productID, submissionID)),
	)
	if err != nil {
		log.Error(ctx, "failed to update whql job in database", err)
		return
	}
}

func checkWHQLJobResult(ctx context.Context, jobID string) {
	// 查询数据库，获取任务信息。
	log.Info(ctx, "get whql job information from database")
	whqlJobDo := conn.MySQLClient(ctx).WhqlJob
	whqlJob, err := whqlJobDo.WithContext(ctx).Select(
		whqlJobDo.ID,
		whqlJobDo.Status,
		whqlJobDo.CreatedTime,
		whqlJobDo.ProductID,
		whqlJobDo.SubmissionID,
		whqlJobDo.AppID,
	).Where(
		whqlJobDo.JobID.Eq(jobID),
	).Take()
	if err != nil || whqlJob == nil {
		log.Error(ctx, "failed to retrieve whql job information from database", err)
		return
	}

	// 校验任务状态。
	log.Info(ctx, "verify whql job state")
	if whqlJob.Status != model.WHQLJobStatusWHQLSigning {
		log.Warn(ctx, "whql job status is not in microsoft signing")
		return
	}

	// 校验任务耗时。
	log.Info(ctx, "verify whql job time consuming")
	if time.Since(whqlJob.CreatedTime) > cfg.Get().Backend().MaximumWHQLJobInterval() {
		log.Warn(ctx, "whql job has exceeded deadline")
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError,
				"任务超时终止，最大时限为：%s", util.FormatDuration(cfg.Get().Backend().MaximumWHQLJobInterval()))),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job state", err)
		}
		return
	}

	// 请求微软服务器查询结果。
	log.Info(ctx, "request microsoft server to get result")
	finished, downloadURL, err := queryMicrosoftSubmission(ctx, whqlJob.ProductID, whqlJob.SubmissionID)
	if !finished {
		log.Info(ctx, "whql job is not over yet")
		return
	}
	if err != nil {
		log.Warn(ctx, "whql job failed", err)
		// 更新数据库任务信息。
		log.Info(ctx, "update whql job in database")
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
		).UpdateColumnSimple(
			whqlJobDo.FinishedTime.Value(time.Now()),
			whqlJobDo.Status.Value(model.WHQLJobStatusFailure),
			query.Concat(whqlJobDo.Log, formatJobLog(log.LevelError, "微软审核不通过：%v", errs.String(err))),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job information in database", err)
			return
		}
		if sqlResult.RowsAffected != 1 {
			log.Error(ctx, "updating whql job information has no effect in database", sqlResult.RowsAffected)
		}
		return
	}

	// 下载微软签名结果文件。
	log.Info(ctx, "download microsoft signed file")
	filePath, err := util.GenerateTemporaryFile(cc.ServiceNameBackend, "whql_signed_*")
	if err != nil {
		log.Error(ctx, "failed to generate temporary file path", err)
		return
	}
	httpCode, fileName, fileSize, fileMD5, err := util.HTTPGetToDisk(ctx, downloadURL, filePath)
	if err != nil {
		log.Error(ctx, "failed to download microsoft audit result file", err, downloadURL)
		return
	}
	if httpCode != http.StatusOK {
		log.Error(ctx, "failed to download microsoft audit result file", err, downloadURL, httpCode)
		return
	}
	defer util.RemoveFile(ctx, filePath)

	// 将文件上传到 Tusd。
	log.Info(ctx, "upload file to tusd", fileName, fileSize, fileMD5)
	tusdID, err := conn.TusdClient(ctx).MultipleUploadFromFile(ctx, filePath)
	if err != nil {
		log.Error(ctx, "failed to upload microsoft audit result file to tusd", err)
		return
	}
	defer func() {
		if err != nil {
			_, err2 := conn.TusdClient(ctx).Delete(ctx, &tus.DeleteRequest{Location: tusdID})
			log.ErrorIf(ctx, err2, "failed to delete tusd file", tusdID)
		}
	}()

	// 保存文件信息到数据库。
	log.Info(ctx, "save file information to database")
	fileID, err := generateID(ctx, IDFile)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			log.ErrorIf(ctx, reclaimID(ctx, IDFile, fileID), "failed to reclaim file id")
		}
	}()
	ctxTx := conn.BeginMySQLTx(ctx)
	defer func() {
		if err == nil {
			log.ErrorIf(ctxTx, conn.CommitMySQLTx(ctxTx), "failed to commit transaction")
		} else {
			log.ErrorIf(ctx, conn.RollbackMySQLTx(ctxTx), "rollback transaction failed")
		}
	}()
	err = createFile(ctxTx, &model.File{
		FileID:      fileID,
		TusdID:      tusdID,
		AppID:       whqlJob.AppID,
		Name:        fileName,
		Md5:         fileMD5,
		Size:        int(fileSize),
		Type:        model.FileTypeMicrosoftSigning,
		CreatedTime: time.Now(),
	})
	if err != nil {
		return
	}

	// 更新数据库中的任务信息。
	log.Info(ctx, "update whql job in database")
	var sqlResult gen.ResultInfo
	whqlJobTxDo := conn.MySQLTxClient(ctxTx).WhqlJob
	sqlResult, err = whqlJobTxDo.WithContext(ctx).Where(
		whqlJobDo.ID.Eq(whqlJob.ID),
	).UpdateColumnSimple(
		whqlJobDo.FinishedTime.Value(time.Now()),
		whqlJobDo.SignedFileID.Value(fileID),
		whqlJobDo.Status.Value(model.WHQLJobStatusSuccess),
		query.Concat(whqlJobDo.Log, formatJobLog(log.LevelInfo, "微软审核通过")),
	)
	if err != nil {
		log.Error(ctx, "failed to update whql job information in database", err)
		return
	}
	if sqlResult.RowsAffected <= 0 {
		log.Error(ctx, "updating whql job information result is not as expected in database", sqlResult.RowsAffected)
	}
}

// 解析 CAB 文件。使用 expand 和 cabextract 命令提取。
func extractCab(cabFilePath string, destinationDirectoryPath string) (err error) {
	err = os.MkdirAll(destinationDirectoryPath, cc.DirectoryMode)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		_, err = exec.Command("expand", cabFilePath, "-F:*", destinationDirectoryPath).CombinedOutput()
	} else {
		_, err = exec.Command(consts.CabextractFilePath, "-d", destinationDirectoryPath, cabFilePath).CombinedOutput()
	}
	return
}
