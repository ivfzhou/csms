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

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/i18n"
	"gitee.com/ivfzhou/csms/comm/log"
)

// LanguageFilter 将客户端语言信息保存进上下文中。
func LanguageFilter(c *gin.Context) {
	language := c.Query(consts.HTTPQueryLanguage)
	if len(language) <= 0 {
		language = string(i18n.LanguageEnglish)
	}
	log.Info(c.Request.Context(), "set request language", language)
	c.Request = c.Request.WithContext(ctxs.WithLanguage(c.Request.Context(), language))
	c.Next()
}
