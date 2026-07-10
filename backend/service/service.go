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
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	"gitee.com/ivfzhou/csms/comm/util"
)

type FileInfo struct {
	Name   string
	Size   int64
	Reader io.ReadCloser
}

// Initialize 服务启动运行逻辑。
func Initialize(ctx context.Context) {
	// 创建消息队列。
	err := initRabbitMQQueue(ctx)
	if err != nil {
		return
	}

	// 建表。
	err = createTables(ctx, time.Now())
	if err != nil {
		return
	}
}

// CronCreateTables 创建数据库表。
func CronCreateTables(ctx context.Context, _ string, _ time.Time) {
	err := createTables(ctx, time.Now().AddDate(0, 1, 0))
	log.ErrorIf(ctx, err, "create tables failed")
}

func initRabbitMQQueue(ctx context.Context) (err error) {
	var queue amqp.Queue
	ovSigningJobQueue := cfg.Get().RabbitMQ().WindowsOVSigningJobQueue()
	if len(ovSigningJobQueue) > 0 {
		queue, err = conn.RabbitMQClient(ctx).QueueDeclare(ovSigningJobQueue, true, false, false, true, amqp.Table{})
		if err != nil {
			log.Error(ctx, "failed to declare windows signing job ov queue", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		log.Info(ctx, "rabbitmq queue declared", queue.Name)
	}

	androidSigningJobQueue := cfg.Get().RabbitMQ().AndroidSigningJobQueue()
	if len(androidSigningJobQueue) > 0 {
		queue, err = conn.RabbitMQClient(ctx).QueueDeclare(
			androidSigningJobQueue, true, false, false, true, amqp.Table{})
		if err != nil {
			log.Error(ctx, "failed to declare android signing job queue", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		log.Info(ctx, "rabbitmq queue declared", queue.Name)
	}

	appleSigningJobQueue := cfg.Get().RabbitMQ().AppleSigningJobQueue()
	if len(appleSigningJobQueue) > 0 {
		queue, err = conn.RabbitMQClient(ctx).QueueDeclare(appleSigningJobQueue, true, false, false, true, amqp.Table{})
		if err != nil {
			log.Error(ctx, "failed to declare apple signing job queue", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
		log.Info(ctx, "rabbitmq queue declared", queue.Name)
	}

	// 从数据库中查出 EV 证书。
	var certificateFingerprints []string
	windowsCertificateQuery := conn.MySQLClient(ctx).WindowsCertificate
	err = windowsCertificateQuery.WithContext(ctx).Select(
		windowsCertificateQuery.Sha1.Upper().Distinct(),
	).Where(
		windowsCertificateQuery.DeletedTime.IsNull(),
		windowsCertificateQuery.Type.In(model.WindowsCertificateTypePersonalEV, model.WindowsCertificateTypeCompanyEV),
		windowsCertificateQuery.Sha1.Length().Gt(0),
	).Scan(&certificateFingerprints)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error(ctx, "failed to retrieve windows certificate fingerprints from database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	evSigningJobQueuePrefix := cfg.Get().RabbitMQ().WindowsEVSigningJobQueuePrefix()
	for _, v := range certificateFingerprints {
		queueName := evSigningJobQueuePrefix + v
		if len(queueName) > 0 {
			queue, err = conn.RabbitMQClient(ctx).QueueDeclare(queueName, true, false, false, true, amqp.Table{})
			if err != nil {
				log.Error(ctx, "failed to declare windows signing job ev queue", err)
				return errs.NewWithError(consts.ErrSystem, err)
			}
			log.Info(ctx, "rabbitmq queue declared", queue.Name)
		}
	}

	return nil
}

func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(cc.TimeFormat)
}

func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(cc.TimeFormat)[:10]
}

func getLastAESSecret(ctx context.Context) (*model.AesKey, error) {
	aesKeyQuery := conn.MySQLClient(ctx).AesKey
	aesKeyInfo, err := aesKeyQuery.WithContext(ctx).Where(
		aesKeyQuery.Secret.Length().Eq(consts.AESKeyLength),
	).Last()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			aesKeyInfo = &model.AesKey{
				Secret:      util.RandomBytes(consts.AESKeyLength),
				CreatedTime: time.Now(),
			}
			if err = aesKeyQuery.WithContext(ctx).Create(aesKeyInfo); err != nil {
				log.Error(ctx, "failed to create aes key to database", err)
				return nil, errs.NewWithError(consts.ErrSystem, err)
			}
			return aesKeyInfo, nil
		}
		log.Error(ctx, "failed to retrieve aes key information from database", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	if time.Since(aesKeyInfo.CreatedTime) > consts.ASEKeyRotationTime {
		aesKeyInfo = &model.AesKey{
			Secret:      util.RandomBytes(consts.AESKeyLength),
			CreatedTime: time.Now(),
		}
		if err = aesKeyQuery.WithContext(ctx).Create(aesKeyInfo); err != nil {
			log.Error(ctx, "failed to create aes key to database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
	}
	return aesKeyInfo, nil
}

func getAESSecret(ctx context.Context, id int) ([]byte, error) {
	aesKeyQuery := conn.MySQLClient(ctx).AesKey
	aesKeyInfo, err := aesKeyQuery.WithContext(ctx).Select(
		aesKeyQuery.Secret,
	).Where(
		aesKeyQuery.ID.Eq(id),
	).Take()
	if err != nil {
		log.Error(ctx, "failed to retrieve aes secret from database", err)
		return nil, errs.NewWithError(consts.ErrSystem, err)
	}
	return aesKeyInfo.Secret, nil
}

func createTables(ctx context.Context, t time.Time) error {
	// 创建文件表信息。
	err := createTable(ctx, model.GetFileTableName(t), &model.File{})
	if err != nil {
		return err
	}

	// 创建应用事件表。
	err = createTable(ctx, model.GetEventTableName(t), &model.Event{})
	if err != nil {
		return err
	}

	// 创建安卓签名任务表。
	err = createTable(ctx, model.GetAndroidSigningJobTableName(t), &model.AndroidSigningJob{})
	if err != nil {
		return err
	}

	// 创建苹果签名任务表。
	err = createTable(ctx, model.GetAppleSigningJobTableName(t), &model.AppleSigningJob{})
	if err != nil {
		return err
	}

	// 创建 Windows 签名任务表。
	err = createTable(ctx, model.GetWindowsSigningJobTableName(t), &model.WindowsSigningJob{})
	if err != nil {
		return err
	}

	return nil
}

func createTable(ctx context.Context, tableName string, pointer any) error {
	var exist bool
	db := query.GetDB(conn.MySQLClient(ctx)).WithContext(ctx)
	err := db.Table("information_schema.TABLES").
		Select("1").
		Where("TABLE_SCHEMA = ? and TABLE_NAME = ?", cfg.Get().MySQL().Database(), tableName).
		Find(&exist).Error
	if err != nil {
		log.Error(ctx, "failed to query database", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	if !exist {
		err = db.Table(tableName).AutoMigrate(pointer)
		if err != nil {
			log.Error(ctx, "failed to create table", err)
			return errs.NewWithError(consts.ErrSystem, err)
		}
	}
	return nil
}

func formatJobLog(level log.Level, format string, args ...any) string {
	format = fmt.Sprintf(format, args...)
	format = strings.ReplaceAll(format, "\r\n", `\r\n`)
	format = strings.ReplaceAll(format, "\n", `\n`)
	format = strings.Trim(format, `\r\n`)
	format = strings.Trim(format, `\n`)
	return fmt.Sprintf("%s %s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), level.String(), format)
}

func publishMessageToQueue(ctx context.Context, queue string, body []byte) error {
	err := conn.RabbitMQClient(ctx).PublishWithContext(ctx, "", queue, true, false,
		amqp.Publishing{
			Body: body,
			Headers: amqp.Table{
				cc.MQHeaderSendTime: time.Now().Unix(),
			},
		},
	)
	if err != nil {
		log.Error(ctx, "failed to publish message to rabbitmq", err)
		return errs.NewWithError(consts.ErrSystem, err)
	}
	return nil
}
