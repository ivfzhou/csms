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

package api_test

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"gitee.com/ivfzhou/csms/comm/conn"
)

const (
	rabbitQueueDeclare rabbitMQMethod = 1 + iota
	rabbitPublishWithContext
)

type RabbitMQMocker interface {
	QueueDeclareOnce(amqp.Queue, error) RabbitMQMocker
	PublishWithContextOnce(error) RabbitMQMocker
	Reset()
}

type rabbitMQMockerImpl struct {
	datasLock sync.Mutex
	datas     []*rabbitMQResultData
	reset     func()
}

type rabbitMQResultData struct {
	fn     rabbitMQMethod
	result []any
}

type rabbitMQMethod int

func MockRabbitMQClient(ctx context.Context) RabbitMQMocker {
	m := &rabbitMQMockerImpl{}
	client := conn.RabbitMQClient(ctx)
	conn.MockRabbitMQClient(m)
	m.reset = func() { conn.MockRabbitMQClient(client) }
	return m
}

func (c *rabbitMQMockerImpl) QueueDeclareOnce(l amqp.Queue, e error) RabbitMQMocker {
	c.datas = append(c.datas, &rabbitMQResultData{
		fn:     rabbitQueueDeclare,
		result: []any{l, e},
	})
	return c
}

func (c *rabbitMQMockerImpl) PublishWithContextOnce(e error) RabbitMQMocker {
	c.datas = append(c.datas, &rabbitMQResultData{
		fn:     rabbitPublishWithContext,
		result: []any{e},
	})
	return c
}

func (c *rabbitMQMockerImpl) Reset() {
	c.reset()
}

func (c *rabbitMQMockerImpl) QueueDeclare(string, bool, bool, bool, bool, amqp.Table) (amqp.Queue, error) {
	data := c.getRabbitData(rabbitQueueDeclare)
	if data != nil {
		err, _ := data[1].(error)
		return data[0].(amqp.Queue), err
	}
	panic("unhandle QueueDeclare")
}

func (c *rabbitMQMockerImpl) PublishWithContext(context.Context, string, string, bool, bool, amqp.Publishing) error {
	data := c.getRabbitData(rabbitPublishWithContext)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	panic("unhandle PublishWithContext")
}

func (c *rabbitMQMockerImpl) getRabbitData(fn rabbitMQMethod) []any {
	if len(c.datas) <= 0 {
		return nil
	}
	c.datasLock.Lock()
	defer c.datasLock.Unlock()
	for i, data := range c.datas {
		if data.fn == fn {
			header := c.datas[:i]
			tail := c.datas[i+1:]
			c.datas = append(header, tail...)
			return data.result
		}
	}
	return nil
}
