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

package service

import (
	"context"
	"fmt"
	"path"

	"gitee.com/ivfzhou/csms/comm/cfg"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/hlk_manager/protocol"
)

func httpHostRestoreMachine(ctx context.Context) error {
	reqURL := fmt.Sprintf("http://%s:%d/%s", cfg.Get().GatewayHost(), cfg.Get().GatewayInternalServerPort(),
		path.Join(cc.ServiceNameHLK, protocol.HTTPPathRetoreTestMachine))
	status, rspBody, err := util.HTTPGet(ctx, reqURL)
	if err != nil {
		return fmt.Errorf("http host machine failed %w %v %s", err, status, rspBody)
	}
	return nil
}
