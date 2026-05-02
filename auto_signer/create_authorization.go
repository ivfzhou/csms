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

package main

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"

	cl "gitee.com/ivfzhou/csms/comm/log"
)

// CreateAuthorization 获取请求凭证。
func CreateAuthorization(cfg *Configuration) (string, bool) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    cfg.Base.AppID,
		Subject:   cfg.Base.AccountID,
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpiredDuration)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	})
	tokenString, err := token.SignedString([]byte(cfg.Base.Secret))
	if err != nil {
		log.Println(cl.LevelError, "failed to create authorization token", err)
		return "", false
	}
	return "Bearer " + tokenString, true
}
