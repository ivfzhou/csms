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
	"net/http"
	"strings"
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

// WindowsInternalGetWHQLJob 获取 WHQL 任务信息。
func WindowsInternalGetWHQLJob(ctx context.Context, req *protocol.WindowsInternalGetWHQLJobReq) (
	whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务信息。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "job not found")
				return
			}
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetWHQLJobToInitialTestMachine 给 HLK 测试虚拟机初始化任务。
func WindowsInternalGetWHQLJobToInitialTestMachine(ctx context.Context,
	req *protocol.WindowsInternalGetWHQLJobToInitialTestMachineReq) (whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusWaitingTest),
			whqlJobDo.TestSystem.Eq(req.System),
		).Order(whqlJobDo.ID.Asc()).Limit(1).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve whql job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if whqlJob == nil {
			// 没有任务可初始化。
			err = nil
			return
		}
	}

	// 更新任务状态。
	{
		log.Info(ctx, "update whql job status")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
			whqlJobDo.Status.Eq(model.WHQLJobStatusWaitingTest),
			whqlJobDo.TestSystem.Eq(req.System),
		).UpdateColumnSimple(
			whqlJobDo.Status.Value(model.WHQLJobStatusInitiallingTestMachine),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if sqlResult.RowsAffected <= 0 {
			// 没有获取到任务，被其它进程抢了。
			whqlJob = nil
			return
		}
		whqlJob.Status = model.WHQLJobStatusInitiallingTestMachine
	}

	return
}

// WindowsInternalUpdateWHQLJob 更新任务。
func WindowsInternalUpdateWHQLJob(ctx context.Context, req *protocol.WindowsInternalUpdateWHQLJobReq) (err error) {
	// 更新任务。
	{
		log.Info(ctx, "update whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, whqlJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(whqlJobDo.Log, logString+"\n"))
		}
		if len(req.TestMachineName) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.TestMachineName.Value(req.TestMachineName))
		}
		if len(req.ServiceName) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.ServiceName.Value(req.ServiceName))
		}
		if len(req.HLKLogFileID) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.HlkLogFileID.Value(req.HLKLogFileID))
		}
		finishedTestTime := time.Time(req.FinishedTestTime)
		if !finishedTestTime.IsZero() {
			assignExprs = append(assignExprs, whqlJobDo.FinishTestTime.Value(finishedTestTime))
		}
		if len(req.HLKXFileID) > 0 {
			assignExprs = append(assignExprs, whqlJobDo.HlkxFileID.Value(req.HLKXFileID))
		}
		if len(assignExprs) <= 0 {
			log.Warn(ctx, "no fields need updating")
			err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusBadRequest)
			return
		}
		_, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetWHQLJobToStartTest 获取任务，调度测试。
func WindowsInternalGetWHQLJobToStartTest(ctx context.Context, req *protocol.WindowsInternalGetWHQLJobToStartTestReq) (
	whqlJob *model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql job")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJob, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusFinishInitiallingTestMachine),
			whqlJobDo.TestSystem.In(req.Systems...),
			whqlJobDo.UpdatedTime.Lt(time.Now().Add(-cfg.Get().Backend().WaitingDelayTimeOfDispatchingTest())),
		).Order(whqlJobDo.UpdatedTime.Asc()).Limit(1).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve whql job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if whqlJob == nil {
			err = nil
			// 没有任务可初始化。
			return
		}
	}

	// 更新任务状态。
	{
		log.Info(ctx, "update whql job status")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		var sqlResult gen.ResultInfo
		sqlResult, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.ID.Eq(whqlJob.ID),
			whqlJobDo.Status.Eq(model.WHQLJobStatusFinishInitiallingTestMachine),
			whqlJobDo.TestSystem.In(req.Systems...),
		).UpdateColumnSimple(
			whqlJobDo.Status.Value(model.WHQLJobStatusDispatchHLKTest),
		)
		if err != nil {
			log.Error(ctx, "failed to update whql job status in database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}

		if sqlResult.RowsAffected <= 0 {
			// 没有获取到任务，被其它进程抢了。
			whqlJob = nil
			return
		}
		whqlJob.Status = model.WHQLJobStatusDispatchHLKTest
	}

	return
}

// WindowsInternalGetTestingWHQLJobs 获取正在测试中的任务。
func WindowsInternalGetTestingWHQLJobs(ctx context.Context, req *protocol.WindowsInternalGetTestingWHQLJobsReq) (
	whqlJobs []*model.WhqlJob, err error) {

	// 查库，获取任务。
	{
		log.Info(ctx, "get whql jobs")
		whqlJobDo := conn.MySQLClient(ctx).WhqlJob
		whqlJobs, err = whqlJobDo.WithContext(ctx).Where(
			whqlJobDo.Status.Eq(model.WHQLJobStatusHLKTesting),
			whqlJobDo.TestSystem.In(req.Systems...),
			whqlJobDo.UpdatedTime.Lt(time.Now()),
		).Order(whqlJobDo.UpdatedTime.Asc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve whql jobs from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// WindowsInternalGetMachineEVCertificates 获取签名机器上在用的 EV UKey。
func WindowsInternalGetMachineEVCertificates(ctx context.Context,
	req *protocol.WindowsInternalGetMachineEVCertificatesReq) (sha1s []string, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get windows certificates")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		var windowsCertificates []*model.WindowsCertificate
		windowsCertificates, err = windowsCertificateDo.WithContext(ctx).Select(
			windowsCertificateDo.Sha1.Distinct(),
		).Where(
			windowsCertificateDo.IP.Eq(req.IP),
			windowsCertificateDo.Type.In(model.WindowsCertificateTypePersonalEV, model.WindowsCertificateTypeCompanyEV),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve windows certificates from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		sha1s = util.ListTo(windowsCertificates,
			func(e *model.WindowsCertificate) string { return strings.ToUpper(e.Sha1) })
		sha1s = util.CleanStrings(sha1s)
	}

	return
}

// WindowsInternalGetWindowsSigningJob 获取签名任务信息。
func WindowsInternalGetWindowsSigningJob(ctx context.Context, req *protocol.WindowsInternalGetWindowsSigningJobReq) (
	windowsSigningJob *model.WindowsSigningJob, err error) {

	// 查库，获取任务信息。
	{
		log.Info(ctx, "get windows signing job")
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(req.JobID))
		windowsSigningJob, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve windows signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	return
}

// WindowsInternalGetCertificate 获取证书信息。
func WindowsInternalGetCertificate(ctx context.Context, req *protocol.WindowsInternalGetCertificateReq) (
	windowsCertificate *model.WindowsCertificate, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get windows certificate")
		windowsCertificateDo := conn.MySQLClient(ctx).WindowsCertificate
		windowsCertificate, err = windowsCertificateDo.WithContext(ctx).Where(
			windowsCertificateDo.ID.Eq(req.ID),
			windowsCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve windows certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if windowsCertificate == nil {
			err = nil
			return
		}
	}

	// 解密证书。
	{
		if windowsCertificate.AesKeyID > 0 {
			var secret []byte
			secret, err = getAESSecret(ctx, windowsCertificate.AesKeyID)
			if err != nil {
				return
			}
			windowsCertificate.Content, err = util.AESCBCDecrypt(windowsCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt windows certificate", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	return
}

// WindowsInternalUpdateSigningJob 更新签名任务。
func WindowsInternalUpdateSigningJob(ctx context.Context, req *protocol.WindowsInternalUpdateSigningJobReq) (
	err error) {

	// 更新任务。
	{
		log.Info(ctx, "update windows signing job")
		windowsSigningJobDo := conn.MySQLClient(ctx).WindowsSigningJob.Table(model.GetWindowsSigningJobByID(req.JobID))
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, windowsSigningJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(windowsSigningJobDo.Log, logString+"\n"))
		}
		if len(req.SignedFileID) > 0 {
			assignExprs = append(assignExprs, windowsSigningJobDo.SignedFileID.Value(req.SignedFileID))
		}
		finishedTime := time.Time(req.FinishedTime)
		if !finishedTime.IsZero() {
			assignExprs = append(assignExprs, windowsSigningJobDo.FinishedTime.Value(finishedTime))
		}
		finishedPETime := time.Time(req.FinishedPESignTime)
		if !finishedPETime.IsZero() {
			assignExprs = append(assignExprs, windowsSigningJobDo.FinishPeTime.Value(finishedPETime))
		}
		if len(assignExprs) <= 0 {
			log.Error(ctx, "no updated values")
			return
		}
		assignExprs = append(assignExprs, windowsSigningJobDo.UpdatedTime.Value(time.Now()))
		_, err = windowsSigningJobDo.WithContext(ctx).Where(
			windowsSigningJobDo.JobID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update windows signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}
