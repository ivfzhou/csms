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

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
	"gitee.com/ivfzhou/csms/hlk_manager/service"
)

// HostMachineReportLog 接收虚拟机写来的日志，并写入硬盘中。
func HostMachineReportLog(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := c.GetRawData()
	if err != nil {
		log.Error(ctx, "read body failed", err)
		util.ResponseAPIError(c, errs.NewWithError(consts.ErrSystem, err))
		return
	}
	err = service.HostInternalReportLog(ctx, data)
	if err != nil {
		log.Warn(ctx, "failed to report log", err, data)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusNoContent)
}

// HostMachineRetoreTestMachine 重置测试机到检查点。
func HostMachineRetoreTestMachine(c *gin.Context) {
	ctx := c.Request.Context()
	err := service.HostInternalRetoreTestMachine(ctx)
	if err != nil {
		log.Warn(ctx, "failed to restore test machine", err)
		util.ResponseAPIError(c, err)
		return
	}
	util.ResponseStatus(c, http.StatusNoContent)
}
