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

package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/ivfzhou/cron/v3"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

var (
	tasks = map[string]*task{
		// 每天凌晨两点，清理文件上传中产生的垃圾分片。
		"CleanUploadingFiles": {mutexWrapper(service.CronCleanUploadingFiles), "0 0 2 * * *"},
		// 每月底，自动创建下月数据库表。
		"CreateTables": {mutexWrapper(service.CronCreateTables), "0 0 2 31 * *"},
		// 每三十秒，提交 cab 文件签名。
		"SubmitCabFileSigning": {mutexWrapper(service.CronSubmitCabFileSigning), ""},
		// 每三十秒，提交微软 Attestation 签名。
		"StartAttestationJobs": {service.CronStartAttestationJobs, ""},
		// 每三十秒，检查微软 Attestation 签名结果。
		"CheckAttestationJobsResult": {service.CronCheckAttestationJobsResult, ""},
		// 每三十秒，检查 HLK 测试成功任务，将 hlkx 文件提交签名。
		"SubmitHLKXFileSigningJobs": {mutexWrapper(service.CronSubmitHLKXFileSigningJobs), ""},
		// 每三十秒，提交微软 WHQL 签名。
		"StartWHQLJobs": {service.CronStartWHQLJobs, ""},
		// 每三十秒，检查微软 WHQL 签名结果。
		"CheckWHQLJobsResult": {service.CronCheckWHQLJobsResult, ""},
	}
	c *cron.Cron
)

type task struct {
	fn   func(context.Context, string, time.Time)
	spec string
}

// Initialize 初始化定时任务。
func Initialize(ctx context.Context) {
	logger := log.GetCronLogger()
	c = cron.New(cron.WithSeconds(), cron.WithLocation(time.Local), cron.WithLogger(logger),
		cron.WithChain(cron.SkipIfStillRunning(logger)))

	hash := (util.IPv4ToNumber(util.LocalIP) % 6) * 10
	defaultSpec := fmt.Sprintf("%d,%d * * * * *", hash, (hash+30)%60)
	for k, v := range tasks {
		if len(v.spec) <= 0 {
			v.spec = defaultSpec
			log.Info(ctx, "cron task specification", k, v.spec)
		}
		log.Info(ctx, "add cron task", k)
		addJob(ctx, string(k), v.spec, v.fn)
	}

	c.Start()
}

// Close 关闭定时任务。
func Close(ctx context.Context) {
	log.Warn(ctx, "closing cron")
	<-c.Stop().Done()
}

// 添加任务。
func addJob(ctx context.Context, taskName, spec string, f func(context.Context, string, time.Time)) {
	_, err := c.AddFunc(spec, panicWrapper(taskName, f))
	if err != nil {
		log.Fatal(ctx, cc.ExitCodeStartCronJobError, "failed to add cron job", err)
	}
}

// 恐慌恢复。
func panicWrapper(jobName string, f func(context.Context, string, time.Time)) func(time.Time) {
	return func(t time.Time) {
		ctx := ctxs.New()
		log.Debug(ctx, "run cron job", jobName, t)
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error(ctx, "cron job occurs panic", recovered, t, util.GetStackCallers())
			}
		}()
		f(ctx, jobName, t)
	}
}

// 分布式锁下运行。
func mutexWrapper(f func(context.Context, string, time.Time)) func(context.Context, string, time.Time) {
	return func(ctx context.Context, taskName string, t time.Time) {
		// 加锁，避免同一时刻任务被多台机器运行。
		success, err := conn.RedisClient(ctx).SetNX(ctx, fmt.Sprintf(consts.RedisKeyCronLockFmt, taskName,
			t.Format("20060102150405")), util.LocalIP, consts.CronLockerExpiration).Result()
		if err != nil {
			log.Error(ctx, "failed to run redis command", err)
			return
		}
		if !success {
			log.Debug(ctx, "skip run cron job", taskName, t)
			return
		}

		f(ctx, taskName, t)
	}
}
