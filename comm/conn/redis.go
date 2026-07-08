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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	redisInitializedFlag atomic.Int32
	redisClosedFlag      atomic.Int32
	redisClient          = &redis.Client{}
	lockKeyToValues      = sync.Map{}
	redisConnectOptions  *redis.Options
	redisAddress         string
	redisUpdateLock      sync.Mutex
	redisLogLevel        int
)

// InitializeRedisConnection 初始化 Redis 服务连接。
// 连接失败会退出程序。
func InitializeRedisConnection(ctx context.Context) {
	if redisClosedFlag.Load() > 0 {
		return
	}

	if !redisInitializedFlag.CompareAndSwap(0, 1) {
		return
	}

	log.Info(ctx, "connecting to redis server")

	// 设置日志输出位置。
	redis.SetLogger(log.GetRedisLogger())
	redisLogLevel = getRedisLogLevel(log.GetLevel())
	setRedisLogLevel(redisLogLevel)

	// 解析链接。
	var err error
	var maskedAddress string
	redisConnectOptions, redisAddress, maskedAddress, err = getRedisAddress(cfg.Get())
	log.FatalIf(ctx, consts.ExitCodeInitialRedisConnectionError, err, "failed to parse redis url")

	// 连接 Redis。
	log.Info(ctx, "redis server address is", maskedAddress)
	redisClient = redis.NewClient(redisConnectOptions)

	// 测试连接。
	redisResult, err := redisClient.Ping(ctx).Result()
	log.FatalIf(ctx, consts.ExitCodeInitialRedisConnectionError, err, "failed to ping redis server")
	if redisResult != "PONG" {
		log.Fatal(ctx, consts.ExitCodeInitialRedisConnectionError, "pinging redis server result is unexpected",
			redisResult)
	}

	// 监听配置更新与重连。
	log.Info(ctx, "watch redis configuration update")
	cfg.RegisterNotifier(watchRedisConfigurationUpdate)

	log.Info(ctx, "successfully connected to redis server")
}

// RedisClient 获取 Redis 操作客户端。
func RedisClient(_ context.Context) *redis.Client {
	return redisClient
}

// CloseRedisConnection 关闭 Redis 连接。
func CloseRedisConnection(ctx context.Context) {
	redisUpdateLock.Lock()
	defer redisUpdateLock.Unlock()

	if !redisClosedFlag.CompareAndSwap(0, 1) {
		return
	}

	log.Warn(ctx, "closing redis connection")
	log.ErrorIf(ctx, redisClient.Close(), "close redis connection error")
}

// RedisLock 加分布式锁。
func RedisLock(ctx context.Context, key string, waitTime, expiration time.Duration) (bool, error) {
	now := time.Now()
	value := strings.ReplaceAll(uuid.NewString(), "-", "")
	success, err := redisClient.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, err
	}
	if success {
		// 加锁成功。
		lockKeyToValues.Store(key, value)
		return true, nil
	}

	// 重试加锁。
	for {
		if time.Since(now) <= waitTime {
			success, err = redisClient.SetNX(ctx, key, value, expiration).Result()
			if err != nil {
				return false, err
			}
			if success {
				lockKeyToValues.Store(key, value)
				return true, nil
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			return false, nil
		}
	}
}

// RedisUnlock 解分布式锁。
func RedisUnlock(ctx context.Context, key string) (bool, error) {
	value, _ := lockKeyToValues.Load(key)
	value2, _ := value.(string)
	return redisClient.Eval(
		ctx,
		`if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]); end; return 0;`,
		[]string{key},
		value2,
	).Bool()
}

// 获取 Redis 连接参数。
func getRedisAddress(configurer cfg.Configurer) (*redis.Options, string, string, error) {
	username := configurer.Redis().Username()
	password := configurer.Redis().Password()
	host := configurer.Redis().Host()
	port := configurer.Redis().Port()
	database := configurer.Redis().Database()
	address := fmt.Sprintf("redis://%s:%s@%s:%d/%d", username, password, host, port, database)
	maskedAddress := fmt.Sprintf("redis://%s:******@%s:%d/%d", username, host, port, database)

	options, err := redis.ParseURL(address)
	return options, address, maskedAddress, err
}

// 监听配置更新与重连。
func watchRedisConfigurationUpdate(configurer cfg.Configurer) {
	if redisClosedFlag.Load() > 0 {
		return
	}

	redisUpdateLock.Lock()
	defer redisUpdateLock.Unlock()

	ctx := ctxs.New()
	newLogLevel := getRedisLogLevel(log.GetLevel())
	if newLogLevel != redisLogLevel {
		log.Warn(ctx, "update redis log level")
		setRedisLogLevel(newLogLevel)
		redisLogLevel = newLogLevel
	}

	options, address, maskedAddress, err := getRedisAddress(configurer)
	if err != nil {
		log.Error(ctx, "invalid redis address", err, maskedAddress)
		return
	}
	if address == redisAddress {
		return
	}

	log.Info(ctx, "redis server address is", maskedAddress)
	newRedisClient := redis.NewClient(options)

	redisResult, err := newRedisClient.Ping(ctx).Result()
	if redisResult != "PONG" || err != nil {
		log.ErrorIf(ctx, newRedisClient.Close(), "failed to close redis client")
		log.Error(ctx, "failed to ping redis server", err)
		return
	}

	log.ErrorIf(ctx, redisClient.Close(), "failed to close redis client")
	redisClient = newRedisClient
	redisConnectOptions = options
	redisAddress = address
	log.Warn(ctx, "successfully connected to redis server")
}

// 设置 Redis 日志等级。
func setRedisLogLevel(level int) {
	switch level {
	case 0:
		redis.SetLogLevel(0)
	case 1:
		redis.SetLogLevel(1)
	case 2:
		redis.SetLogLevel(2)
	case 3:
		redis.SetLogLevel(3)
	default:
		redis.SetLogLevel(3)
	}
}

// 获取 Redis 日志。
func getRedisLogLevel(level log.Level) int {
	switch level {
	case log.LevelError:
		return 0
	case log.LevelWarn:
		return 1
	case log.LevelInfo:
		return 2
	case log.LevelDebug:
		return 3
	default:
		return 3
	}
}
