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

	"gitee.com/ivfzhou/csms/fastlane_proxy/api"
)

func initInternalRoute(r *gin.RouterGroup) {
	internalBundle(r)
	internalProfile(r)
	internalCertificate(r)
}

func internalBundle(r *gin.RouterGroup) {
	r = r.Group("/bundle")

	r.POST("/applyInHouse", api.BundleApplyInHouse)
	r.DELETE("removeInHouse", api.BundleRemoveInHouse)
	r.POST("modifyInHouseCapabilities", api.BundleModifyInHouseCapabilities)
}

func internalProfile(r *gin.RouterGroup) {
	r = r.Group("/profile")

	r.POST("/applyInHouse", api.ProfileApplyInHouse)
	r.DELETE("/removeInHouse", api.ProfileRemoveInHouse)
}

func internalCertificate(r *gin.RouterGroup) {
	r = r.Group("/certificate")

	r.POST("/applyPush", api.CertificateApplyPush)
	r.DELETE("/removePush", api.CertificateRemovePush)
}
