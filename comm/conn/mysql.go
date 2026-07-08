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
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/query"
)

var (
	mysqlConnectionClosedFlag      atomic.Int32
	mysqlConnectionInitializedFlag atomic.Int32
	mysqlConnectionLock            sync.Mutex
	gormQuery                      = &query.DB{}
	mysqlDBConnection              *sql.DB
	mysqlAddress                   string
	mysqlMaximumIdle               int
	mysqlMaximumOpen               int
	mysqlMaximumLife               time.Duration
)

// InitializeMySQLConnection 建立与 MySQL 的连接，初始化 GROM。
// 连接 MySQL 服务失败，会退出程序。
func InitializeMySQLConnection(ctx context.Context) {
	mysqlConnectionLock.Lock()
	defer mysqlConnectionLock.Unlock()

	if mysqlConnectionClosedFlag.Load() > 0 {
		log.Warn(ctx, "mysql connection is closed, no need to initialize")
		return
	}

	if !mysqlConnectionInitializedFlag.CompareAndSwap(0, 1) {
		return
	}

	log.Info(ctx, "connecting to mysql server")

	// 获取 MySQL 连接地址。
	var mysqlMaskedAddress string
	mysqlAddress, mysqlMaskedAddress = getMySQLConnectionURL(cfg.Get())
	log.Info(ctx, "mysql server address is", mysqlMaskedAddress)

	// 获取 GORM 配置。
	gromOption := &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
		Logger:               log.GetGormLogger(),
		PrepareStmt:          true,
		DisableAutomaticPing: true,
		QueryFields:          true,
	}

	// 连接 MySQL 服务器。
	gormInstance, err := gorm.Open(mysql.Open(mysqlAddress), gromOption)
	if err != nil {
		time.Sleep(3 * time.Second)
		gormInstance, err = gorm.Open(mysql.Open(mysqlAddress), gromOption)
	}
	log.FatalIf(ctx, consts.ExitCodeInitialMySQLConnectionError, err, "failed to connect to mysql server")

	// 设置连接池。
	mysqlDBConnection, err = gormInstance.DB()
	log.FatalIf(ctx, consts.ExitCodeInitialMySQLConnectionError, err, "invalid mysql connection")
	mysqlMaximumIdle = cfg.Get().MySQL().MaximumIdle()
	mysqlMaximumOpen = cfg.Get().MySQL().MaximumOpen()
	mysqlMaximumLife = cfg.Get().MySQL().MaximumLife()
	mysqlDBConnection.SetMaxIdleConns(mysqlMaximumIdle)
	mysqlDBConnection.SetMaxOpenConns(mysqlMaximumOpen)
	mysqlDBConnection.SetConnMaxLifetime(mysqlMaximumLife)

	err = mysqlDBConnection.Ping()
	log.FatalIf(ctx, consts.ExitCodeInitialMySQLConnectionError, err, "failed to ping mysql server")

	// 设置默认数据库操作实例。
	gormQuery = query.NewDB(query.Use(gormInstance))

	// 监听配置更新，重新连接。
	cfg.RegisterNotifier(watchMySQLConfigurationUpdate)

	log.Info(ctx, "successfully connected to mysql server")
}

// MySQLTxClient 获取 GORM 操作数据库对象。若 ctx 开启了事务，则会在同一个事务下操作。
// 该函数不可并发调用。
func MySQLTxClient(ctx context.Context) *query.DB {
	if !consts.UnitTestMode() && ctxs.InTransaction(ctx) {
		tx := ctxs.DBClient(ctx)
		if tx == nil {
			log.Info(ctx, "start mysql transaction")
			tx = gormQuery.Begin()
			ctxs.SetDBClient(ctx, tx)
			return query.NewDB(tx.Query)
		}
		log.Debug(ctx, "get a mysql  transaction")
		return query.NewDB(tx.Query)
	}

	return gormQuery
}

// MySQLClient 获取 GORM 操作数据库对象。不会开启事务。
func MySQLClient(_ context.Context) *query.DB {
	return gormQuery
}

// CloseMySQLConnection 关闭与 MySQL 的连接。
func CloseMySQLConnection(ctx context.Context) {
	mysqlConnectionLock.Lock()
	defer mysqlConnectionLock.Unlock()

	if !mysqlConnectionClosedFlag.CompareAndSwap(0, 1) {
		return
	}

	log.Warn(ctx, "closing mysql connection")
	log.ErrorIf(ctx, mysqlDBConnection.Close(), "failed to close mysql connection")
}

// OneMySQLTCPSQL 等价于 gorm.DB.Connection。
func OneMySQLTCPSQL(ctx context.Context, fn func(q *query.DB) error) error {
	return query.GetDB(MySQLTxClient(ctx)).Connection(func(tx *gorm.DB) error {
		return fn(query.NewDB(query.Use(tx)))
	})
}

// BeginMySQLTx 标记数据库事务。
func BeginMySQLTx(ctx context.Context) context.Context {
	return ctxs.WithInTransaction(ctx)
}

// CommitMySQLTx 提交事务。
func CommitMySQLTx(ctx context.Context) error {
	tx := ctxs.DBClient(ctx)
	if tx == nil {
		log.Warn(ctx, "no transaction found")
		return nil
	}
	return tx.Commit()
}

// RollbackMySQLTx 回滚事务。
func RollbackMySQLTx(ctx context.Context) error {
	tx := ctxs.DBClient(ctx)
	if tx == nil {
		log.Warn(ctx, "no transaction found")
		return nil
	}
	return tx.Rollback()
}

// 获取 MySQL 地址。
func getMySQLConnectionURL(configurer cfg.Configurer) (address string, maskedAddress string) {
	username := configurer.MySQL().Username()
	password := configurer.MySQL().Password()
	host := configurer.MySQL().Host()
	port := configurer.MySQL().Port()
	database := configurer.MySQL().Database()
	parameters := configurer.MySQL().Parameters()
	address = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", username, password, host, port, database, parameters)
	maskedAddress = fmt.Sprintf("%s:******@tcp(%s:%d)/%s?%s", username, host, port, database, parameters)
	return
}

// 监听配置更新并重连。
func watchMySQLConfigurationUpdate(configurer cfg.Configurer) {
	mysqlConnectionLock.Lock()
	defer mysqlConnectionLock.Unlock()

	ctx := ctxs.New()
	if mysqlConnectionClosedFlag.Load() > 0 {
		log.Warn(ctx, "mysql connection is closed, no need to update configuration")
		return
	}

	newMySQLAddress, newMaskedMySQLAddress := getMySQLConnectionURL(configurer)
	if newMySQLAddress != mysqlAddress {
		log.Info(ctx, "updating mysql connection", newMaskedMySQLAddress)
		gromOption := &gorm.Config{
			SkipDefaultTransaction: true,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "t_",
				SingularTable: true,
			},
			Logger:               log.GetGormLogger(),
			PrepareStmt:          true,
			DisableAutomaticPing: true,
			QueryFields:          true,
		}
		newGormInstance, err := gorm.Open(mysql.Open(newMySQLAddress), gromOption)
		if err != nil {
			time.Sleep(3 * time.Second)
			newGormInstance, err = gorm.Open(mysql.Open(newMySQLAddress), gromOption)
		}
		if err != nil {
			log.Error(ctx, "failed to connection mysql", err, newMaskedMySQLAddress)
			goto OnlyUpdatePoolParameters
		}

		var newMysqlDBConnection *sql.DB
		newMysqlDBConnection, err = newGormInstance.DB()
		if err != nil {
			log.Error(ctx, "failed to get mysql connection", err, newMaskedMySQLAddress)
			goto OnlyUpdatePoolParameters
		}

		err = mysqlDBConnection.Ping()
		if err != nil {
			log.Error(ctx, "failed to ping mysql server", err, newMaskedMySQLAddress)
			goto OnlyUpdatePoolParameters
		}

		mysqlMaximumIdle = configurer.MySQL().MaximumIdle()
		mysqlMaximumOpen = configurer.MySQL().MaximumOpen()
		mysqlMaximumLife = configurer.MySQL().MaximumLife()
		newMysqlDBConnection.SetMaxIdleConns(mysqlMaximumIdle)
		newMysqlDBConnection.SetMaxOpenConns(mysqlMaximumOpen)
		newMysqlDBConnection.SetConnMaxLifetime(mysqlMaximumLife)

		mysqlAddress = newMySQLAddress
		gormQuery = query.NewDB(query.Use(newGormInstance))
		log.ErrorIf(ctx, mysqlDBConnection.Close(), "failed to close mysql connection")
		mysqlDBConnection = newMysqlDBConnection
		log.Warn(ctx, "updating mysql connection successfully")

		return
	}

OnlyUpdatePoolParameters:
	newMySQLMaximumIdle := configurer.MySQL().MaximumIdle()
	newMySQLMaximumOpen := configurer.MySQL().MaximumOpen()
	newMySQLMaximumLife := configurer.MySQL().MaximumLife()
	if newMySQLMaximumIdle != mysqlMaximumIdle {
		log.Warn(ctx, "updating mysql connection maximum idle")
		mysqlDBConnection.SetMaxIdleConns(newMySQLMaximumIdle)
		mysqlMaximumIdle = newMySQLMaximumIdle
	}
	if newMySQLMaximumOpen != mysqlMaximumOpen {
		log.Warn(ctx, "updating mysql connection maximum open")
		mysqlDBConnection.SetMaxOpenConns(newMySQLMaximumOpen)
		mysqlMaximumOpen = newMySQLMaximumOpen
	}
	if newMySQLMaximumLife != mysqlMaximumLife {
		log.Warn(ctx, "updating mysql connection maximum life")
		mysqlDBConnection.SetConnMaxLifetime(newMySQLMaximumLife)
		mysqlMaximumLife = newMySQLMaximumLife
	}
}
