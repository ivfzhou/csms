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

// AppleInternalGetSigningJob 获取任务信息。
func AppleInternalGetSigningJob(ctx context.Context, req *protocol.AppleInternalGetSigningJobReq) (
	appleSigningJob *model.AppleSigningJob, err error) {

	// 获取任务信息。
	{
		log.Info(ctx, "get apple job")
		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob.Table(model.GetAppleSigningJobByID(req.JobID))
		appleSigningJob, err = appleSigningJobDo.WithContext(ctx).Where(
			appleSigningJobDo.JobID.Eq(req.JobID),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errs.IsMySQLError(err, mysql.ErrNoSuchTable) {
				err = errs.NewWithStatus(consts.ErrParameterInvalid, http.StatusNotFound)
				return
			}
			log.Error(ctx, "failed to retrieve apple signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// AppleInternalGetCertificateAndProfile 获取证书和描述文件信息。
func AppleInternalGetCertificateAndProfile(ctx context.Context, req *protocol.AppleInternalGetCertificateAndProfileReq) (
	rsp *protocol.AppleInternalGetCertificateAndProfileRsp, err error) {

	// 查库，获取描述文件信息。
	var appleProfile *model.AppleProfile
	{
		log.Info(ctx, "get apple profile")
		appleProfileDo := conn.MySQLClient(ctx).AppleProfile
		appleProfile, err = appleProfileDo.WithContext(ctx).Select(
			appleProfileDo.Content,
			appleProfileDo.CertificateID,
		).Where(
			appleProfileDo.ID.Eq(req.ProfileID),
			appleProfileDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "profile not found")
				return
			}
			log.Error(ctx, "failed to retrieve apple profile from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 查库，获取证书信息。
	var appleCertificate *model.AppleCertificate
	{
		log.Info(ctx, "get apple certificate")
		appleCertificateDo := conn.MySQLClient(ctx).AppleCertificate
		appleCertificate, err = appleCertificateDo.WithContext(ctx).Select(
			appleCertificateDo.Content,
			appleCertificateDo.AesKeyID,
			appleCertificateDo.Password,
		).Where(
			appleCertificateDo.ID.Eq(appleProfile.CertificateID),
			appleCertificateDo.DeletedTime.IsNull(),
		).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errs.NewWithStatusMsg(consts.ErrParameterInvalid, http.StatusNotFound, "certificate not found")
				return
			}
			log.Error(ctx, "failed to retrieve apple certificate from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 解密证书。
	{
		log.Info(ctx, "decrypt apple certificate")
		var secret []byte
		secret, err = getAESSecret(ctx, appleCertificate.AesKeyID)
		if err != nil {
			return
		}
		appleCertificate.Content, err = util.AESCBCDecrypt(appleCertificate.Content, secret)
		if err != nil {
			log.Error(ctx, "failed to decrypt apple certificate", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// Base64 编码。
	var certificate, profile string
	{
		log.Info(ctx, "base64 encode apple certificate")
		certificate = base64.StdEncoding.EncodeToString(appleCertificate.Content)
		profile = base64.StdEncoding.EncodeToString(appleProfile.Content)
		rsp = &protocol.AppleInternalGetCertificateAndProfileRsp{
			Certificate: certificate,
			Profile:     profile,
			Password:    appleCertificate.Password,
		}
	}

	return
}

// AppleInternalUpdateSigningJob 更新任务信息。
func AppleInternalUpdateSigningJob(ctx context.Context, req *protocol.AppleInternalUpdateSigningJobReq) (err error) {
	// 更新任务。
	{
		log.Info(ctx, "update apple signing job")
		appleSigningJobDo := conn.MySQLClient(ctx).AppleSigningJob.Table(model.GetAppleSigningJobByID(req.JobID))
		assignExprs := make([]field.AssignExpr, 0, 5)
		if req.Status > 0 {
			assignExprs = append(assignExprs, appleSigningJobDo.Status.Value(req.Status))
		}
		logString := util.TrimBlank(req.AppendLog)
		if len(logString) > 0 {
			assignExprs = append(assignExprs, query.Concat(appleSigningJobDo.Log, logString+"\n"))
		}
		if len(req.SignedFileID) > 0 {
			assignExprs = append(assignExprs, appleSigningJobDo.SignedFileID.Value(req.SignedFileID))
		}
		finishedTime := time.Time(req.FinishedTime)
		if !finishedTime.IsZero() {
			assignExprs = append(assignExprs, appleSigningJobDo.FinishedTime.Value(finishedTime))
		}
		if len(assignExprs) <= 0 {
			log.Error(ctx, "no updated values")
			return
		}
		_, err = appleSigningJobDo.WithContext(ctx).Where(
			appleSigningJobDo.JobID.Eq(req.JobID),
		).UpdateColumnSimple(
			assignExprs...,
		)
		if err != nil {
			log.Error(ctx, "failed to update apple signing job from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}
