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
	"gitee.com/ivfzhou/csms/backend/filter"
	"gitee.com/ivfzhou/csms/comm/model"
)

func initApiRoute(r *gin.RouterGroup) {
	apiFile(r)
	apiAndroid(r)
	apiWindows(r)
	apiApple(r)
}

func apiFile(r *gin.RouterGroup) {
	r = r.Group("/file", filter.APIAuthenticateFilter, filter.PermissionAPIAuthenticateFilter)

	addRoute(
		r,
		http.MethodGet,
		"/download",
		api.FileAPIDownload,
	)
	addRoute(
		r,
		http.MethodPost,
		"/initial",
		api.FileAPIInitial,
	)
	addRoute(
		r,
		http.MethodPost,
		"/uploadPart",
		api.FileAPIUploadPart,
	)
	addRoute(
		r,
		http.MethodGet,
		"/mergeParts",
		api.FileAPIMergeParts,
	)
}

func apiAndroid(r *gin.RouterGroup) {
	r = r.Group("/android", filter.APIAuthenticateFilter, filter.PermissionAPIAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate",
		[]int{model.CapabilityDownloadAndroidCertificate},
		api.AndroidAPIDownloadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAPKSigningJob",
		[]int{model.CapabilitySubmitAndroidSignJob},
		filter.DatabaseTransactionFilter,
		api.AndroidAPISubmitAPKSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAABSigningJob",
		[]int{model.CapabilitySubmitAndroidSignJob},
		filter.DatabaseTransactionFilter,
		api.AndroidAPISubmitAABSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAPKPatchSigningJob",
		[]int{model.CapabilitySubmitAndroidSignJob},
		filter.DatabaseTransactionFilter,
		api.AndroidAPISubmitAPKPatchSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getSigningJobInformation",
		[]int{model.CapabilityGetSignJobInfo},
		api.AndroidAPIGetSigningJobInformation,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCertificates",
		[]int{model.CapabilityGetCertificateInfo},
		api.AndroidAPIListCertificates,
	)
}

func apiWindows(r *gin.RouterGroup) {
	r = r.Group("/windows", filter.APIAuthenticateFilter, filter.PermissionAPIAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate",
		[]int{model.CapabilityDownloadWindowsOVCertificate},
		api.WindowsAPIDownloadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getCertificatePassword",
		[]int{model.CapabilityGetWindowsOVCertificatePassword},
		api.WindowsAPIGetCertificatePassword,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitSigningJob",
		[]int{model.CapabilitySubmitWindowsPESignJob},
		filter.DatabaseTransactionFilter,
		api.WindowsAPISubmitSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitWHQLJob",
		[]int{model.CapabilitySubmitWHQLSignJob},
		filter.DatabaseTransactionFilter,
		api.WindowsAPISubmitWHQLJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCertificates",
		[]int{model.CapabilityGetCertificateInfo},
		api.WindowsAPIListCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getSigningJobInformation",
		[]int{model.CapabilityGetSignJobInfo},
		api.WindowsAPIGetSigningJobInformation,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getWHQLJobInformation",
		[]int{model.CapabilityGetSignJobInfo},
		api.WindowsAPIGetWHQLJobInformation,
	)
}

func apiApple(r *gin.RouterGroup) {
	r = r.Group("/apple", filter.APIAuthenticateFilter, filter.PermissionAPIAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate",
		[]int{model.CapabilityDownloadAppleCertAndProvision},
		api.AppleAPIDownloadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitSigningJob",
		[]int{model.CapabilitySubmitAppleSignJob},
		filter.DatabaseTransactionFilter,
		api.AppleAPISubmitSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getSigningJobInformation",
		[]int{model.CapabilityGetSignJobInfo},
		api.AppleAPIGetSigningJobInformation,
	)
}
