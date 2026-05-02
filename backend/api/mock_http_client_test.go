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
	"net/http"
	"sync"

	"gitee.com/ivfzhou/csms/comm/util"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

type HTTPMocker interface {
	ResponseOnce(*http.Response, error) HTTPMocker
	Reset()
}

type httpResultDate struct {
	rsp *http.Response
	err error
}

type httpMockerImpl struct {
	datas []*httpResultDate
	lock  sync.Mutex
	reset func()
}

func MockHTTPClient(_ context.Context) HTTPMocker {
	m := &httpMockerImpl{}
	reset := mvt.Chain(util.GetHTTPClient()).Elem().FieldByName("Transport").Set(m).Reset
	m.reset = reset
	return m
}

func (m *httpMockerImpl) ResponseOnce(rsp *http.Response, err error) HTTPMocker {
	m.datas = append(m.datas, &httpResultDate{rsp: rsp, err: err})
	return m
}

func (m *httpMockerImpl) Reset() {
	m.reset()
}

func (m *httpMockerImpl) RoundTrip(*http.Request) (*http.Response, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.datas) == 0 {
		panic("unhandled http client")
	}
	data := m.datas[0]
	m.datas = m.datas[1:]
	return data.rsp, data.err
}
