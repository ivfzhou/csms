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
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"

	"gitee.com/ivfzhou/csms/comm/conn"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

const (
	redisScriptLoad redisMethod = iota + 1
	redisEvalsha
	redisSadd
	redisSet
	redisGet
	redisDel
	redisSrem
	redisHset
	redisHget
	redisSetnx
	redisZrangeByScore
	redisZadd
	redisEval
	redisZrangeWithScore
	redisHdel
	redisZrem
	redisZrangeArgs
)

type RedisMocker interface {
	ScriptLoadOnce(result string, err error) RedisMocker
	EvalshaOnce(result any, err error) RedisMocker
	SRemOnce(result int64, err error) RedisMocker
	SetOnce(result string, err error) RedisMocker
	GetOnce(result string, err error) RedisMocker
	DelOnce(result int64, err error) RedisMocker
	SAddOnce(result int64, err error) RedisMocker
	HSetOnce(result int64, err error) RedisMocker
	HGetOnce(result string, err error) RedisMocker
	SetNXOnce(result bool, err error) RedisMocker
	ZRangeByScoreOnce(result []string, err error) RedisMocker
	ZRangeArgsOnce(result []string, err error) RedisMocker
	ZAddOnce(result int64, err error) RedisMocker
	EvalOnce(result any, err error) RedisMocker
	ZRangeWithScores(result []redis.Z, err error) RedisMocker
	ZRem(result int64, err error) RedisMocker
	HDel(result int64, err error) RedisMocker
	Reset()
}

type redisMockerImpl struct {
	datas     []*redisResultData
	datasLock sync.Mutex
	reset     func()
}

type redisMethod int

type redisResultData struct {
	action redisMethod
	result any
	err    error
}

func MockRedis(ctx context.Context) RedisMocker {
	m := &redisMockerImpl{}
	m.reset = mvt.Chain(conn.RedisClient(ctx)).Elem().FieldByName("cmdable").Set(m.do).Reset
	return m
}

func dealRedisCmd[T any](data *redisResultData, v any) error {
	cmd := v.(interface {
		SetErr(error)
		SetVal(T)
		FullName() string
	})
	if data == nil {
		panic(fmt.Sprintf("unhandled redis command %v", cmd.FullName()))
	}
	cmd.SetErr(data.err)
	cmd.SetVal(data.result.(T))
	return data.err
}

func (m *redisMockerImpl) ScriptLoadOnce(result string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisScriptLoad, result: result, err: err})
	return m
}

func (m *redisMockerImpl) EvalshaOnce(result any, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisEvalsha, result: result, err: err})
	return m
}

func (m *redisMockerImpl) ZRangeWithScores(result []redis.Z, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisZrangeWithScore, result: result, err: err})
	return m
}

func (m *redisMockerImpl) EvalOnce(result any, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisEval, result: result, err: err})
	return m
}

func (m *redisMockerImpl) SRemOnce(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisSrem, result: result, err: err})
	return m
}

func (m *redisMockerImpl) SetOnce(result string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisSet, result: result, err: err})
	return m
}

func (m *redisMockerImpl) GetOnce(result string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisGet, result: result, err: err})
	return m
}

func (m *redisMockerImpl) DelOnce(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisDel, result: result, err: err})
	return m
}

func (m *redisMockerImpl) HDel(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisHdel, result: result, err: err})
	return m
}

func (m *redisMockerImpl) ZRem(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisZrem, result: result, err: err})
	return m
}

func (m *redisMockerImpl) HSetOnce(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisHset, result: result, err: err})
	return m
}

func (m *redisMockerImpl) HGetOnce(result string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisHget, result: result, err: err})
	return m
}

func (m *redisMockerImpl) SetNXOnce(result bool, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisSetnx, result: result, err: err})
	return m
}

func (m *redisMockerImpl) ZRangeByScoreOnce(result []string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisZrangeByScore, result: result, err: err})
	return m
}

func (m *redisMockerImpl) ZRangeArgsOnce(result []string, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisZrangeArgs, result: result, err: err})
	return m
}

func (m *redisMockerImpl) SAddOnce(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisSadd, result: result, err: err})
	return m
}

func (m *redisMockerImpl) ZAddOnce(result int64, err error) RedisMocker {
	m.datas = append(m.datas, &redisResultData{action: redisZadd, result: result, err: err})
	return m
}

func (m *redisMockerImpl) Reset() {
	m.reset()
}

func (m *redisMockerImpl) do(_ context.Context, cmd redis.Cmder) error {
	switch c := cmd.(type) {
	case *redis.StringCmd:
		if len(c.Args()) >= 2 && c.Args()[0] == "script" && c.Args()[1] == "load" {
			return dealRedisCmd[string](m.getRedisCmdData(redisScriptLoad), c)
		}
		if len(c.Args()) >= 2 && c.Args()[0] == "get" {
			return dealRedisCmd[string](m.getRedisCmdData(redisGet), c)
		}
		if len(c.Args()) >= 2 && c.Args()[0] == "hget" {
			return dealRedisCmd[string](m.getRedisCmdData(redisHget), c)
		}
	case *redis.Cmd:
		if len(c.Args()) >= 1 && c.Args()[0] == "evalsha" {
			return dealRedisCmd[any](m.getRedisCmdData(redisEvalsha), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "eval" {
			return dealRedisCmd[any](m.getRedisCmdData(redisEval), c)
		}
	case *redis.IntCmd:
		if len(c.Args()) >= 1 && c.Args()[0] == "sadd" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisSadd), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "del" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisDel), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "srem" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisSrem), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "hset" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisHset), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "zadd" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisZadd), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "zrem" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisZrem), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "hdel" {
			return dealRedisCmd[int64](m.getRedisCmdData(redisHdel), c)
		}
	case *redis.StatusCmd:
		if len(c.Args()) >= 1 && c.Args()[0] == "set" {
			return dealRedisCmd[string](m.getRedisCmdData(redisSet), c)
		}
	case *redis.BoolCmd:
		if len(c.Args()) >= 1 && c.Args()[0] == "setnx" {
			return dealRedisCmd[bool](m.getRedisCmdData(redisSetnx), c)
		}
		if len(c.Args()) >= 6 && c.Args()[0] == "set" && c.Args()[5] == "nx" {
			return dealRedisCmd[bool](m.getRedisCmdData(redisSetnx), c)
		}
	case *redis.StringSliceCmd:
		if len(c.Args()) >= 1 && c.Args()[0] == "zrangebyscore" {
			return dealRedisCmd[[]string](m.getRedisCmdData(redisZrangeByScore), c)
		}
		if len(c.Args()) >= 1 && c.Args()[0] == "zrange" {
			return dealRedisCmd[[]string](m.getRedisCmdData(redisZrangeArgs), c)
		}
	case *redis.ZSliceCmd:
		if len(c.Args()) >= 5 && c.Args()[0] == "zrange" && c.Args()[4] == "withscores" {
			return dealRedisCmd[[]redis.Z](m.getRedisCmdData(redisZrangeWithScore), c)
		}
	}
	panic(fmt.Sprintf("unhandled redis cmd %v", cmd.FullName()))
}

func (m *redisMockerImpl) getRedisCmdData(action redisMethod) *redisResultData {
	if len(m.datas) <= 0 {
		return nil
	}
	m.datasLock.Lock()
	defer m.datasLock.Unlock()
	for i, data := range m.datas {
		if data.action == action {
			header := m.datas[:i]
			tail := m.datas[i+1:]
			m.datas = append(header, tail...)
			return data
		}
	}
	return nil
}
