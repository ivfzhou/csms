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
	"net/http"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/api"
)

func initInternalRoute(r *gin.RouterGroup) {
	internalWindows(r)
	internalFile(r)
	internalAndroid(r)
	internalApple(r)
}

func internalWindows(r *gin.RouterGroup) {
	r = r.Group("/windows")

	addRoute(r, http.MethodGet, "/getWHQLJob", api.WindowsInternalGetWHQLJob)
	addRoute(r, http.MethodGet, "/getWHQLJobToInitialTestMachine", api.WindowsInternalGetWHQLJobToInitialTestMachine)
	addRoute(r, http.MethodPost, "/updateWHQLJob", api.WindowsInternalUpdateWHQLJob)
	addRoute(r, http.MethodGet, "/getWHQLJobToStartTest", api.WindowsInternalGetWHQLJobToStartTest)
	addRoute(r, http.MethodGet, "/getTestingWHQLJobs", api.WindowsInternalGetTestingWHQLJobs)
	addRoute(r, http.MethodGet, "/getMachineEVCertificates", api.WindowsInternalGetMachineEVCertificates)
	addRoute(r, http.MethodGet, "/getWindowsSigningJob", api.WindowsInternalGetWindowsSigningJob)
	addRoute(r, http.MethodGet, "/getCertificate", api.WindowsInternalGetCertificate)
	addRoute(r, http.MethodPost, "/updateSigningJob", api.WindowsInternalUpdateSigningJob)
}

func internalFile(r *gin.RouterGroup) {
	r = r.Group("/file")

	addRoute(r, http.MethodGet, "/download", api.FileInternalDownload)
	addRoute(r, http.MethodPost, "/upload", api.FileInternalUpload)
}

func internalAndroid(r *gin.RouterGroup) {
	r = r.Group("/android")

	addRoute(r, http.MethodGet, "/getSigningJob", api.AndroidInternalGetSigningJob)
	addRoute(r, http.MethodGet, "/getCertificate", api.AndroidInternalGetCertificate)
	addRoute(r, http.MethodPost, "/updateSigningJob", api.AndroidInternalUpdateSigningJob)
}

func internalApple(r *gin.RouterGroup) {
	r = r.Group("/apple")

	addRoute(r, http.MethodGet, "/getSigningJob", api.AppleInternalGetSigningJob)
	addRoute(r, http.MethodGet, "/getCertificateAndProfile", api.AppleInternalGetCertificateAndProfile)
	addRoute(r, http.MethodPost, "/updateSigningJob", api.AppleInternalUpdateSigningJob)
}
