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
	"time"

	"github.com/pingcap/tidb/parser/mysql"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

// AndroidInternalGetSigningJob 获取签名任务信息。
func AndroidInternalGetSigningJob(ctx context.Context, req *protocol.AndroidInternalGetSigningJobReq) (
	androidSigningJob *model.AndroidSigningJob, err error) {

	// 查库，获取信息。
	{
		log.Info(ctx, "get android job")
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob.Table(model.GetAndroidSigningJobByID(req.JobID))
		androidSigningJob, err = androidSigningJobDo.WithContext(ctx).Where(
			androidSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android job information from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	return
}

// AndroidInternalGetCertificate 获取安卓证书信息。
func AndroidInternalGetCertificate(ctx context.Context, req *protocol.AndroidInternalGetCertificateReq) (
	androidCertificate *model.AndroidCertificate, err error) {

	// 查库，获取证书信息。
	{
		log.Info(ctx, "get android certificate")
		androidCertificateDo := conn.MySQLClient(ctx).AndroidCertificate
		androidCertificate, err = androidCertificateDo.WithContext(ctx).Where(
			androidCertificateDo.ID.Eq(req.ID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve android certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		err = nil
	}

	// 解密证书内容。
	{
		if androidCertificate != nil && len(androidCertificate.Content) > 0 {
			var secret []byte
			secret, err = getAESSecret(ctx, androidCertificate.AesKeyID)
			if err != nil {
				return
			}
			androidCertificate.Content, err = util.AESCBCDecrypt(androidCertificate.Content, secret)
			if err != nil {
				log.Error(ctx, "failed to decrypt android certificate content", err)
				err = errs.NewWithError(consts.ErrSystem, err)
				return
			}
		}
	}

	return
}

// AndroidInternalUpdateSigningJob 更新任务信息。
func AndroidInternalUpdateSigningJob(ctx context.Context, req *protocol.AndroidInternalUpdateSigningJobReq) (
	err error) {

	// 更新任务。
	{
		log.Info(ctx, "update android signing job")
		androidSigningJobDo := conn.MySQLClient(ctx).AndroidSigningJob.Table(model.GetAndroidSigningJobByID(req.JobID))
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, androidSigningJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(androidSigningJobDo.Log, logString+"\n"))
		}
		if len(req.SignedFileID) > 0 {
			assignExprs = append(assignExprs, androidSigningJobDo.SignedFileID.Value(req.SignedFileID))
		}
		finishedTime := time.Time(req.FinishedTime)
		if !finishedTime.IsZero() {
			assignExprs = append(assignExprs, androidSigningJobDo.FinishedTime.Value(finishedTime))
		}
		if len(assignExprs) <= 0 {
			log.Error(ctx, "no updated values")
			return
		}
		_, err = androidSigningJobDo.WithContext(ctx).Where(
			androidSigningJobDo.JobID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update android signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}
