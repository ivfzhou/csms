/*
 * Copyright (c) 2024 ivfzhou
 * backend is licensed under Mulan PSL v2.
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
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/conn"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// OpenAPI 请求限流校验脚本。
const apiRequestRateLimitScript = `
-- 拼接存放数据使用的键
local key = '` + consts.RedisKeyApiAccessLimitPrefix + `'..KEYS[1];
-- 获取上次剩余请求量
local residue = tonumber(redis.call('hget', key, 'residue') or 0);
-- 获取上次请求时间
local lastAccessTime = tonumber(redis.call('hget', key, 'lastAccessTime') or 0);
-- 获取当前时间
local nowArr = redis.call('time');
local now = tonumber(nowArr[1]) * 1000000 + tonumber(nowArr[2]);
-- 计算出当前可用请求量
local genPerTime = tonumber(ARGV[1] or 0);
local max = tonumber(ARGV[2] or 0);
local canCost = math.min((now - lastAccessTime) * genPerTime + residue, max);
-- 判断是否可以放行
local need = tonumber(ARGV[3] or 0);
residue = canCost - need;
if residue >= 0 then
	redis.call('hmset', key, 'lastAccessTime', now, 'residue', residue);
	return 1
end
redis.call('hmset', key, 'lastAccessTime', now, 'residue', canCost);
return 0;`

// OpenAPI 请求限流校验脚本摘要值。
var apiRequestRateLimitScriptCmdSha string

// InitialAPIAuthenticateLimitScript 获取 Redis 脚本 Sha。
func InitialAPIAuthenticateLimitScript(ctx context.Context) {
	var err error
	apiRequestRateLimitScriptCmdSha, err = conn.RedisClient(ctx).ScriptLoad(ctx, apiRequestRateLimitScript).Result()
	if err != nil {
		log.Fatal(ctx, cc.ExitCodeLoadRedisScriptError, "failed to load api request rate limit script", err)
	}
}

// APIAuthenticateFilter API 访问鉴权。
func APIAuthenticateFilter(c *gin.Context) {
	ctx := c.Request.Context()
	log.Info(ctx, "authenticate api request")

	// 获取请求凭证。
	accessToken := c.Request.Header.Get("Authorization")
	if len(accessToken) <= 7 {
		c.Abort()
		util.ResponseStatus(c, http.StatusUnauthorized)
		return
	}
	accessToken = accessToken[7:]

	// 解析 JWT 凭证。
	log.Info(ctx, "parse http authorization token", accessToken)
	var apiAccount *model.APIAccount
	var app *model.App
	var authID string
	tokenObj, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		// 判断加密算法。
		if token.Header["alg"] != consts.APIAuthorizationAlgorithm {
			log.Error(ctx, "jwt algorithm is invalid", token.Header["alg"])
			return nil, errs.NewWithStatusMsg(consts.ErrPermissionDenied, http.StatusUnauthorized,
				"algorithm does not support")
		}

		// 获取 appID。
		appID, err := token.Claims.GetIssuer()
		if err != nil {
			log.Error(ctx, "failed to get jwt issuer", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}

		// 从数据库获取应用信息。
		app, err = service.GetAppInfoByID(ctx, appID)
		if err != nil {
			log.Error(ctx, "failed to retrieve app information from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		if app == nil || app.ID <= 0 {
			log.Warn(ctx, "no app was found in database", appID)
			return nil, errs.NewWithStatusMsg(consts.ErrPermissionDenied, http.StatusUnauthorized,
				"app of api account not found")
		}

		// 获取 authID。
		if authID, err = token.Claims.GetSubject(); err != nil {
			log.Error(ctx, "failed to get jwt subject", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}

		// 从数据库获取授权信息。
		apiAccountDo := conn.MySQLClient(ctx).APIAccount
		apiAccount, err = apiAccountDo.WithContext(ctx).Where(
			apiAccountDo.AppID.Eq(app.ID),
			apiAccountDo.AccountID.Eq(authID),
		).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(ctx, "failed to retrieve api account information from database", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		if apiAccount == nil || apiAccount.ID <= 0 {
			log.Warn(ctx, "no api account information found in database", authID)
			return nil, errs.NewWithStatusMsg(consts.ErrPermissionDenied, http.StatusUnauthorized,
				"api account not found")
		}

		// 授权是否过期。
		if time.Since(apiAccount.ExpiredTime) >= 0 {
			log.Warn(ctx, "api account expired", apiAccount.ExpiredTime)
			return nil, errs.NewWithStatus(consts.ErrApiAccountExpired, http.StatusUnauthorized)
		}

		// 签发时间是否大于当前时间。
		issuedAt, err := token.Claims.GetIssuedAt()
		if err != nil {
			log.Error(ctx, "failed to get jwt issuer at", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		if issuedAt == nil || time.Since(issuedAt.Time) <= 0 {
			log.Warn(ctx, "issuer time is invalid", issuedAt.Time)
			return nil, errs.NewWithStatusMsg(consts.ErrPermissionDenied, http.StatusUnauthorized,
				"issuer time is invalid")
		}

		// 校验凭证时效。
		expirationTime, err := token.Claims.GetExpirationTime()
		if err != nil {
			log.Error(ctx, "failed to get jwt expiration time", err)
			return nil, errs.NewWithError(consts.ErrSystem, err)
		}
		if expirationTime == nil ||
			expirationTime.Sub(issuedAt.Time) > cfg.Get().Backend().OpenAPIMaximumExpirationDuration() {
			log.Warn(ctx, "jwt expiration time is too long", expirationTime)
			return nil, errs.NewWithStatus(consts.ErrApiAuthorizationExpired, http.StatusUnauthorized)
		}

		return []byte(apiAccount.Secret), nil
	})
	if err != nil {
		c.Abort()
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			log.Warn(ctx, "api authorization token is malformed")
			util.ResponseStatusMsg(c, http.StatusUnauthorized, "access token malformed")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			log.Warn(ctx, "api authorization token signature is invalid")
			util.ResponseStatusMsg(c, http.StatusUnauthorized, "access token signature is invalid")
		case errors.Is(err, jwt.ErrTokenExpired):
			log.Warn(ctx, "api authorization token expired", accessToken)
			util.ResponseStatusMsg(c, http.StatusUnauthorized, "access token expired")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			log.Warn(ctx, "api authorization token is not valid yet", accessToken)
			util.ResponseStatusMsg(c, http.StatusUnauthorized, "access token is not valid yet")
		default:
			util.ResponseAPIError(c, err)
		}
		return
	}
	if !tokenObj.Valid {
		c.Abort()
		log.Warn(ctx, "api authorization token is not valid", accessToken)
		util.ResponseStatusMsg(c, http.StatusUnauthorized, "access token is not valid")
		return
	}

	// 校验请求 IP 是否在允许范围。
	log.Info(ctx, "verify api request ip")
	reqIP := ctxs.RequestIP(ctx)
	if !slices.Contains(apiAccount.IP, "*") {
		reqIPNum := util.IPv4ToNumber(reqIP)
		isPass := false
		for _, v := range apiAccount.IP {
			// 是否是 IP 段。
			if strings.Contains(v, "-") {
				ipArr := strings.Split(v, "-")
				begin := util.IPv4ToNumber(ipArr[0])
				end := util.IPv4ToNumber(ipArr[1])
				if reqIPNum >= begin && reqIPNum <= end {
					isPass = true
					break
				}
			} else if util.IPv4ToNumber(v) == reqIPNum {
				isPass = true
				break
			}
		}
		if !isPass {
			c.Abort()
			log.Warn(ctx, "api request ip", reqIP, "is not allowed")
			util.ResponseAPIError(c, errs.NewWithStatus(consts.ErrRequestIPNotAllowed, http.StatusForbidden))
			return
		}
	}

	// 请求限流校验。
	log.Info(ctx, "verify api request rate")
	if !cc.SkipRateLimit() {
		maxTokens := apiAccount.Frequency * 60 * 1000 * 1000
		var pass bool
		pass, err = conn.RedisClient(ctx).EvalSha(ctx, apiRequestRateLimitScriptCmdSha,
			[]string{strconv.Itoa(apiAccount.ID)}, maxTokens, maxTokens, 1).Bool()
		if err != nil {
			c.Abort()
			log.Error(ctx, "failed to eval api request rate limit script", err)
			util.ResponseError(c, errs.NewWithError(consts.ErrSystem, err))
			return
		}
		if !pass {
			c.Abort()
			util.ResponseStatus(c, http.StatusTooManyRequests)
			return
		}
	}

	// 信息记录上下文。
	ctx = ctxs.WithAPIAccount(ctx, apiAccount)
	ctx = ctxs.WithApp(ctx, app)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}
