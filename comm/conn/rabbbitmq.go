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

package conn

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	rabbitMQClosedFlag         int32
	rabbitMQInitializedFlag    int32
	rabbitMQUpdateLock         sync.Mutex
	rabbitMQCloseNotifyChannel chan *amqp.Error
	rabbitMQContext            context.Context
	rabbitMQContextCancel      context.CancelFunc
	rabbitMQAddress            string
	rabbitMQMaskedAddress      string
	rabbitMQConnection         *amqp.Connection
	rabbitMQChannel            *amqp.Channel
	rabbitMQInstance           RabbitMQAPI
	rabbitMQPrefetchCount      int
)

// RabbitMQAPI 消息队列。
type RabbitMQAPI interface {
	QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (
		amqp.Queue, error)
	PublishWithContext(_ context.Context, exchange string, key string, mandatory bool, immediate bool,
		msg amqp.Publishing) error
	ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool,
		args amqp.Table) (<-chan amqp.Delivery, error)
}

// InitializeRabbitMQConnection 建立与 RabbitMQ 服务器的连接。
// 连接建立失败，会退出程序。
func InitializeRabbitMQConnection(ctx context.Context) {
	rabbitMQUpdateLock.Lock()
	defer rabbitMQUpdateLock.Unlock()

	if atomic.LoadInt32(&rabbitMQClosedFlag) > 0 {
		return
	}

	if !atomic.CompareAndSwapInt32(&rabbitMQInitializedFlag, 0, 1) {
		return
	}

	log.Info(ctx, "connecting to rabbitmq server")

	// 设置日志输出位置。
	amqp.SetLogger(log.GetRabbitMQLogger())

	// 建立 MQ 连接。
	rabbitMQAddress, rabbitMQMaskedAddress = getRabbitMQAddress(cfg.Get())
	log.Info(ctx, "rabbitmq server address is", rabbitMQMaskedAddress)
	var err error
	if rabbitMQConnection, err = amqp.Dial(rabbitMQAddress); err != nil {
		time.Sleep(3 * time.Second)
		rabbitMQConnection, err = amqp.Dial(rabbitMQAddress)
	}
	log.FatalIf(ctx, consts.ExitCodeInitialMQConnectionError, err, "failed to connect to rabbitmq server")

	// 打开通道。
	if rabbitMQChannel, err = rabbitMQConnection.Channel(); err != nil {
		time.Sleep(3 * time.Second)
		rabbitMQChannel, err = rabbitMQConnection.Channel()
	}
	log.FatalIf(ctx, consts.ExitCodeInitialMQConnectionError, err, "failed to create rabbitmq channel")

	// 设置并发量。
	rabbitMQPrefetchCount = cfg.Get().RabbitMQ().PrefetchCount()
	err = rabbitMQChannel.Qos(rabbitMQPrefetchCount, 0, true)
	log.FatalIf(ctx, consts.ExitCodeInitialMQConnectionError, err, "failed to set rabbitmq qos")

	// 监听连接关闭。
	rabbitMQCloseNotifyChannel = rabbitMQChannel.NotifyClose(make(chan *amqp.Error, 1))
	rabbitMQContext, rabbitMQContextCancel = context.WithCancel(context.Background())
	rabbitMQInstance = rabbitMQChannel
	log.Info(ctx, "listen rabbitmq connection closure and reconnect")
	go listenAndReconnectRabbitMQ()

	// 注册连接配置更新监听。
	cfg.RegisterNotifier(watchRabbitMQConfigurationUpdate)

	log.Info(ctx, "successfully connected to rabbitmq server")
}

// RabbitMQClient 获取 RabbitMQ 客户端实例。
func RabbitMQClient(_ context.Context) RabbitMQAPI {
	return rabbitMQInstance
}

// CloseRabbitMQConnection 关闭 RabbitMQ 连接。
func CloseRabbitMQConnection(ctx context.Context) {
	rabbitMQUpdateLock.Lock()
	defer rabbitMQUpdateLock.Unlock()

	if !atomic.CompareAndSwapInt32(&rabbitMQClosedFlag, 0, 1) {
		return
	}

	log.Warn(ctx, "closing rabbitmq connection")

	// 关闭连接监听。
	rabbitMQContextCancel()
	time.Sleep(100 * time.Microsecond)

	log.ErrorIf(ctx, rabbitMQChannel.Close(), "failed to close rabbitmq channel")
	log.ErrorIf(ctx, rabbitMQConnection.Close(), "failed to close rabbitmq connection")
}

// MockRabbitMQClient 单测模式下替换实现。
func MockRabbitMQClient(client RabbitMQAPI) {
	if consts.UnitTestMode() {
		rabbitMQInstance = client
	}
}

// 监听连接中断并重连。
func listenAndReconnectRabbitMQ() {
	for {
		ctx := ctxs.New()
		select {
		// MQ 关闭，退出监听。
		case <-rabbitMQContext.Done():
			log.Warn(ctx, "exit listening rabbitmq connection")
			return
		// 通道关闭，重新连接。
		case err, ok := <-rabbitMQCloseNotifyChannel:
			func() {
				if !rabbitMQUpdateLock.TryLock() {
					return
				}

				// 重连失败，等待 10 秒。
				defer func() {
					if rabbitMQConnection.IsClosed() || rabbitMQChannel.IsClosed() {
						time.Sleep(10 * time.Second)
					}
				}()

				defer rabbitMQUpdateLock.Unlock()

				if ok {
					log.ErrorIf(ctx, err, "listened rabbitmq channel closure")
				}
				if rabbitMQConnection.IsClosed() {
					newConnection := reconnectRabbitMQ(ctx)
					if newConnection == nil {
						return
					}
					newChannel := reopenRabbitMQChannel(ctx, newConnection)
					if newChannel == nil {
						log.ErrorIf(ctx, newConnection.Close(), "failed to close rabbitmq connection")
						return
					}

					rabbitMQConnection = newConnection
					rabbitMQChannel = newChannel
					rabbitMQCloseNotifyChannel = newChannel.NotifyClose(make(chan *amqp.Error, 1))
					rabbitMQInstance = newChannel
					log.Info(ctx, "successfully reconnected rabbitmq server")
				} else {
					newChannel := reopenRabbitMQChannel(ctx, rabbitMQConnection)
					if newChannel == nil {
						return
					}

					rabbitMQChannel = newChannel
					rabbitMQCloseNotifyChannel = newChannel.NotifyClose(make(chan *amqp.Error, 1))
					rabbitMQInstance = newChannel
					log.Info(ctx, "successfully reopen rabbitmq channel")
				}
			}()
		}
	}
}

// 新建连接。
func reconnectRabbitMQ(ctx context.Context) *amqp.Connection {
	// RabbitMQ 是否关闭。
	select {
	case <-rabbitMQContext.Done():
		log.Info(ctx, "exit reconnecting rabbitmq connection")
		return nil
	default:
	}

	// 连接 RabbitMQ。
	conn, err := amqp.Dial(rabbitMQAddress)
	if err != nil {
		log.Error(ctx, "failed to connect rabbitmq", err)
		return nil
	}

	return conn
}

// 新建通道。
func reopenRabbitMQChannel(ctx context.Context, conn *amqp.Connection) *amqp.Channel {
	// MQ 是否关闭。
	select {
	case <-rabbitMQContext.Done():
		log.Info(ctx, "exit rechanneling rabbitmq connection")
		return nil
	default:
	}

	// 在重建连接时，连接又断了。
	for conn.IsClosed() {
		log.Error(ctx, "connection has been closed")
		return nil
	}

	// 打开通道。
	ch, err := conn.Channel()
	if err != nil {
		log.Error(ctx, "failed to rechannel rabbitmq", err)
		return nil
	}

	// 设置并发量。
	if err = ch.Qos(rabbitMQPrefetchCount, 0, true); err != nil {
		log.Error(ctx, "failed to set rabbitmq qos", err)
		return nil
	}

	return ch
}

// 获取 RabbitMQ 地址。
func getRabbitMQAddress(configurer cfg.Configurer) (rabbitMQAddress string, rabbitMQMaskedAddress string) {
	username := configurer.RabbitMQ().Username()
	password := configurer.RabbitMQ().Password()
	host := configurer.RabbitMQ().Host()
	port := configurer.RabbitMQ().Port()
	virtualHost := configurer.RabbitMQ().VirtualHost()
	rabbitMQAddress = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, virtualHost)
	rabbitMQMaskedAddress = fmt.Sprintf("amqp://%s:******@%s:%d/%s", username, host, port, virtualHost)
	return
}

// 监听配置更新并重连。
func watchRabbitMQConfigurationUpdate(configurer cfg.Configurer) {
	rabbitMQUpdateLock.Lock()
	defer rabbitMQUpdateLock.Unlock()

	ctx := ctxs.New()
	if atomic.LoadInt32(&rabbitMQClosedFlag) > 0 {
		log.Warn(ctx, "rabbitmq connection is closed, no need to update configuration")
		return
	}

	newRabbitMQAddress, newRabbitMQMaskedAddress := getRabbitMQAddress(configurer)
	if newRabbitMQAddress != rabbitMQAddress {
		log.Info(ctx, "updating rabbitmq server connection", newRabbitMQMaskedAddress)
		newRabbitMQConnection, err := amqp.Dial(newRabbitMQAddress)
		if err != nil {
			time.Sleep(3 * time.Second)
			newRabbitMQConnection, err = amqp.Dial(newRabbitMQAddress)
		}
		if err != nil {
			log.Error(ctx, "failed to connect to rabbitmq server", err, newRabbitMQMaskedAddress)
			goto OnlyUpdatePrefetchCount
		}

		var newRabbitMQChannel *amqp.Channel
		if newRabbitMQChannel, err = newRabbitMQConnection.Channel(); err != nil {
			time.Sleep(3 * time.Second)
			newRabbitMQChannel, err = newRabbitMQConnection.Channel()
		}
		if err != nil {
			log.Error(ctx, "failed to create rabbitmq channel", err, newRabbitMQMaskedAddress)
			log.ErrorIf(ctx, newRabbitMQConnection.Close(), "failed to close rabbitmq connection")
			goto OnlyUpdatePrefetchCount
		}

		newRabbitMQPrefetchCount := configurer.RabbitMQ().PrefetchCount()
		err = newRabbitMQChannel.Qos(newRabbitMQPrefetchCount, 0, true)
		if err != nil {
			log.Error(ctx, "failed to set rabbitmq qos", err)
			log.ErrorIf(ctx, newRabbitMQConnection.Close(), "failed to close rabbitmq connection")
			goto OnlyUpdatePrefetchCount
		}

		rabbitMQChannel = newRabbitMQChannel
		rabbitMQInstance = newRabbitMQChannel
		rabbitMQCloseNotifyChannel = newRabbitMQChannel.NotifyClose(make(chan *amqp.Error, 1))
		rabbitMQConnection = newRabbitMQConnection
		rabbitMQAddress, rabbitMQMaskedAddress = newRabbitMQAddress, newRabbitMQMaskedAddress
		rabbitMQPrefetchCount = newRabbitMQPrefetchCount
		log.Warn(ctx, "update rabbitmq server connection successfully", newRabbitMQMaskedAddress)
	}

OnlyUpdatePrefetchCount:
	newRabbitMQPrefetchCount := configurer.RabbitMQ().PrefetchCount()
	if rabbitMQPrefetchCount != newRabbitMQPrefetchCount {
		err := rabbitMQChannel.Qos(newRabbitMQPrefetchCount, 0, true)
		if err != nil {
			log.Error(ctx, "failed to set rabbitmq qos", err, newRabbitMQPrefetchCount)
			return
		}
		rabbitMQPrefetchCount = newRabbitMQPrefetchCount
		log.Warn(ctx, "updated rabbitmq server prefetch count", newRabbitMQPrefetchCount)
	}
}
