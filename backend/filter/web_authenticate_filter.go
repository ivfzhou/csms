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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
)

// WebAuthenticateFilter Web 接口会话鉴权
func WebAuthenticateFilter(c *gin.Context) {
	ctx := c.Request.Context()
	ip := ctxs.RequestIP(ctx)
	log.Info(ctx, "authenticate web request")

	// 获取请求的会话信息。
	skey, err := c.Cookie(consts.HTTPHeaderSessionKey)
	if err != nil {
		c.Abort()
		if errors.Is(err, http.ErrNoCookie) {
			util.ResponseCode(c, consts.ErrNeedLogin)
			return
		}
		log.Error(ctx, "failed to get session from cookie", err)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}
	userName, err := c.Cookie(consts.HTTPHeaderSessionUser)
	if err != nil {
		c.Abort()
		if errors.Is(err, http.ErrNoCookie) {
			util.ResponseCode(c, consts.ErrNeedLogin)
			return
		}
		log.Error(ctx, "failed to get username from cookie", err)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}

	// 获取缓存中的会话信息。
	session, err := conn.RedisClient(ctx).Get(ctx, fmt.Sprintf(consts.RedisKeySessionFmt, userName)).Result()
	if err != nil {
		c.Abort()
		if errors.Is(err, redis.Nil) {
			util.ResponseCode(c, consts.ErrNeedLogin)
			return
		}
		log.Error(ctx, "failed to get Redis data", err)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}
	var sessionInfo service.SessionInfo
	if err = json.Unmarshal([]byte(session), &sessionInfo); err != nil {
		c.Abort()
		log.Error(ctx, "failed to unserialize session information", err, session)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}

	// 校验会话信息。
	if sessionInfo.User != userName || sessionInfo.Session != skey || sessionInfo.IP != ip {
		c.Abort()
		util.ResponseCode(c, consts.ErrNeedLogin)
		return
	}

	// 获取数据库中的用户信息。
	userDo := conn.MySQLClient(ctx).User
	user, err := userDo.WithContext(ctx).Where(userDo.NameEn.Eq(userName)).Take()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Abort()
		log.Error(ctx, "failed to retrieve user information from database", err)
		util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}
	if user == nil || user.ID <= 0 {
		c.Abort()
		log.Warn(ctx, "unknown user", userName)
		util.ResponseCode(c, consts.ErrNeedLogin)
		return
	}

	// 将用户信息保存进上下文。
	ctx = ctxs.WithUser(ctx, user)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}
