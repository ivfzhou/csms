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

package route

import (
	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/hlk_manager/api"
)

func initInternalRoute(r *gin.RouterGroup) {
	internalHostMachine(r)
}

func internalHostMachine(r *gin.RouterGroup) {
	r.POST("/host/reportLog", api.HostMachineReportLog)
	r.GET("/host/retoreTestMachine", api.HostMachineRetoreTestMachine)
}
