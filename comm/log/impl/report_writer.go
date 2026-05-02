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

package impl

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
	"gitee.com/ivfzhou/csms/comm/util"
)

type reportWriter struct {
	reportURL string
}

// 新建实例。
func newReportWriter() internal.WriteCloser {
	w := &reportWriter{reportURL: getReportLogURL(cfg.Get())}
	cfg.RegisterNotifier(func(configurer cfg.Configurer) {
		reportURL := getReportLogURL(configurer)
		if reportURL == w.reportURL {
			return
		}

		ctx := ctxs.New()
		log.Warn(ctx, "update report log url", reportURL)
		w.reportURL = reportURL
	})
	return w
}

func (w *reportWriter) Write(s string) {
	status, rspBody, err := util.HTTPPost(context.Background(), w.reportURL,
		strings.NewReader(s), consts.HTTPHeaderIP, util.LocalIP)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "reporting log failed: %v\n", err)
		return
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		_, _ = fmt.Fprintf(os.Stderr, "reporting log failed: %d %s\n", status, rspBody)
	}
}

func (w *reportWriter) IsColorful() bool {
	return false
}

func (w *reportWriter) Close(context.Context) {}

// 获取上报地址。
func getReportLogURL(configurer cfg.Configurer) string {
	str := fmt.Sprintf("http://%s:%d/%s", configurer.GatewayHost(),
		configurer.GatewayInternalServerPort(), path.Join(consts.ServiceNameHLK, configurer.Log().ReportPath()))
	reportURL, err := url.Parse(str)
	if err != nil {
		ctx := ctxs.New()
		log.Error(ctx, "report log url is invalid", err, str)
		return ""
	}
	return reportURL.String()
}
