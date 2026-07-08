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
	"strings"
	"time"

	"github.com/google/uuid"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
)

// ID 类型。
const (
	IDFile               IDType = "file"
	IDApp                IDType = "app"
	IDWindowsJob         IDType = "windows_job"
	IDWHQLJob            IDType = "whql_job"
	IDAndroidJob         IDType = "android_job"
	IDAppleJob           IDType = "apple_job"
	IDWindowsCertificate IDType = "windows_certificate"
	IDAndroidCertificate IDType = "android_certificate"
	IDAppleCertificate   IDType = "apple_certificate"
	IDAppleDevice        IDType = "apple_device"
)

type IDType string

// 生成唯一 ID。
func generateID(ctx context.Context, typ IDType) (string, error) {
	return generateIDWithTime(ctx, typ, time.Now())
}

// 生成唯一 ID。
func generateIDWithTime(ctx context.Context, typ IDType, t time.Time) (string, error) {
	yearMonth := t.Format("200601")
	for range 100 {
		id := ""
		switch typ {
		case IDApp, IDWindowsCertificate, IDAndroidCertificate, IDWHQLJob, IDAppleCertificate, IDAppleDevice:
			id = strings.ReplaceAll(uuid.NewString(), "-", "")
		default:
			id = yearMonth + strings.ReplaceAll(uuid.NewString(), "-", "")
		}
		redisResult, err := conn.RedisClient(ctx).SAdd(ctx, fmt.Sprintf(consts.RedisKeyIDFmt, typ), id).Result()
		if err != nil {
			log.Error(ctx, "failed to add data id to redis", err)
			return "", errs.NewWithError(consts.ErrSystem, err)
		}
		if redisResult > 0 && len(id) > 0 {
			return id, nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return "", errs.NewWithMsg(consts.ErrSystem, "no id available")
}

// 回收 ID。
func reclaimID(ctx context.Context, typ IDType, id string) error {
	err := conn.RedisClient(ctx).SRem(ctx, fmt.Sprintf(consts.RedisKeyIDFmt, typ), id).Err()
	if err != nil {
		log.Error(ctx, "deleting redis data ID failed", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}
