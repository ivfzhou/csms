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

package filter

import (
	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/consts"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// DatabaseTransactionFilter 数据库事务过滤器。
func DatabaseTransactionFilter(c *gin.Context) {
	ctx := c.Request.Context()
	if cc.UnitTestMode() {
		log.Warn(ctx, "unit test mode skip transaction")
		c.Next()
		return
	}

	// 事务会话设置到上下文中。
	log.Info(ctx, "mark in transaction")
	ctx = ctxs.WithInTransaction(ctx)
	c.Request = c.Request.WithContext(ctx)

	defer func() {
		ctx = c.Request.Context()
		tx := ctxs.DBClient(ctx)
		if tx == nil {
			log.Info(ctx, "no transaction found")
			return
		}

		// 检查是否提交事务。
		if p := recover(); p != nil || ctxs.Error(ctx) != nil {
			// 回滚事务。
			log.Warn(ctx, "rollback db transaction")
			log.ErrorIf(ctx, tx.Rollback(), "rollback transaction failed")
			if p != nil {
				c.Abort()
				log.Error(ctx, "handle request panic", p, util.GetStackCallers())
				util.ResponseError(c, errs.New(consts.ErrSystem))
			}
		} else {
			// 提交事务。
			log.Info(ctx, "commit db transaction")
			log.ErrorIf(ctx, tx.Commit(), "commit transaction failed")
		}
	}()

	// 继续业务逻辑。
	c.Next()
}
