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
	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/filter"
	"gitee.com/ivfzhou/csms/comm/model"
)

var (
	permissionSystem   = []int{model.UserRoleSystemAdmin}
	permissionAppRead  = []int{model.UserRoleSystemAdmin, model.UserRoleAppAdmin, model.UserRoleAppMember}
	permissionAppAdmin = []int{model.UserRoleAppAdmin}
	permissionAppSign  = []int{model.UserRoleAppAdmin, model.UserRoleAppSigner}
	permissionApp      = []int{model.UserRoleAppAdmin, model.UserRoleAppMember}
)

func initWebRoute(r *gin.RouterGroup) {
	webUser(r)
	webApp(r)
	webFile(r)
	webEvent(r)
	webTodo(r)
	webOpen(r)
	webAndroid(r)
	webWindows(r)
	webApple(r)
	webNotice(r)
}

func webUser(r *gin.RouterGroup) {
	r = r.Group("/user")

	addRoute(
		r,
		http.MethodPost,
		"/register",
		filter.AntiShakeFilter,
		filter.DatabaseTransactionFilter,
		api.UserWebRegister,
	)
	addRoute(
		r,
		http.MethodPost,
		"/login",
		filter.AntiShakeFilter,
		api.UserWebLogin,
	)
	addRoute(
		r,
		http.MethodGet,
		"/getInformation",
		filter.WebAuthenticateFilter,
		api.UserWebGetInformation,
	)
	addRoute(
		r,
		http.MethodPost,
		"/update",
		filter.WebAuthenticateFilter,
		filter.AntiShakeFilter,
		filter.DatabaseTransactionFilter,
		api.UserWebUpdate,
	)
	addRoute(
		r,
		http.MethodGet,
		"/search",
		filter.WebAuthenticateFilter,
		api.UserWebSearch,
	)
	addRoute(
		r,
		http.MethodDelete,
		"/logout",
		filter.WebAuthenticateFilter,
		filter.AntiShakeFilter,
		api.UserWebLogout,
	)
}

func webApp(r *gin.RouterGroup) {
	r = r.Group("/app", filter.WebAuthenticateFilter)

	addRoute(
		r,
		http.MethodPost,
		"/register",
		filter.AntiShakeFilter,
		filter.DatabaseTransactionFilter,
		api.AppWebRegister,
	)
	addRoute(
		r,
		http.MethodPost,
		"/search",
		api.AppWebSearch,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/update/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppWebUpdate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getInformation/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AppWebGetInformation,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/invalidate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppWebInvalidate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/enable/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppWebEnable,
	)
	addRoute(
		r,
		http.MethodGet,
		"/count",
		api.AppWebCount,
	)
}

func webFile(r *gin.RouterGroup) {
	r = r.Group("/file", filter.WebAuthenticateFilter)

	addRoute(
		r,
		http.MethodGet,
		"/download",
		api.FileWebDownload,
	)
	addRoute(
		r,
		http.MethodPost,
		"/initial",
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.FileWebInitial,
	)
	addRoute(
		r,
		http.MethodPost,
		"/uploadPart",
		api.FileWebUploadPart,
	)
	addRoute(
		r,
		http.MethodGet,
		"/mergeParts",
		filter.AntiShakeFilter,
		api.FileWebMergeParts,
	)
}

func webEvent(r *gin.RouterGroup) {
	r = r.Group("/event", filter.WebAuthenticateFilter)

	addRoute(
		r,
		http.MethodGet,
		"/list",
		api.EventWebList,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/statistic",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.EventWebStatistic,
	)
}

func webTodo(r *gin.RouterGroup) {
	r = r.Group("/todo", filter.WebAuthenticateFilter)

	addRoute(
		r,
		http.MethodGet,
		"/count",
		api.TodoWebCount,
	)
	addRoute(
		r,
		http.MethodGet,
		"/list",
		api.TodoWebList,
	)
	addRoute(
		r,
		http.MethodGet,
		"/listDealt",
		api.TodoWebListDealt,
	)
	addRoute(
		r,
		http.MethodPost,
		"/create",
		filter.AntiShakeFilter,
		filter.DatabaseTransactionFilter,
		api.TodoWebCreate,
	)
	addRoute(
		r,
		http.MethodGet,
		"/getDetail",
		api.TodoWebGetDetail,
	)
	addRoute(
		r,
		http.MethodPost,
		"/deal",
		filter.AntiShakeFilter,
		filter.DatabaseTransactionFilter,
		api.TodoWebDeal,
	)
}

func webOpen(r *gin.RouterGroup) {
	r = r.Group("/open", filter.WebAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/apply/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.OpenWebApply,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/update/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.OpenWebUpdate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getInformation/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.OpenWebGetInformation,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/list/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.OpenWebList,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/renewal/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.OpenWebRenewal,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/reset/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.OpenWebReset,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/remove/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.OpenWebRemove,
	)
}

func webAndroid(r *gin.RouterGroup) {
	r = r.Group("/android", filter.WebAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/addOrganization",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebAddOrganization,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listOrganizations",
		permissionSystem,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebListOrganizations,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyCertificate/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebApplyCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/uploadCertificate/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebUploadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCertificates/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebListCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebDownloadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getGooglePlayCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebGetGooglePlayCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/getGooglePlayDeployCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebGetGooglePlayDeployCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/getGooglePlayUpgradeCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebGetGooglePlayUpgradeCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getCertificateFacebookDigest/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebGetCertificateFacebookDigest,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAPKSigningJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebSubmitAPKSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAABSigningJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebSubmitAABSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitAPKPatchSigningJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebSubmitAPKPatchSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listSigningJobs/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebListSigningJobs,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/removeOrganization",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AndroidWebRemoveOrganization,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/deleteCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AndroidWebDeleteCertificate,
	)
}

func webWindows(r *gin.RouterGroup) {
	r = r.Group("/windows", filter.WebAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/uploadCertificate/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.WindowsWebUploadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCertificates/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebListCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebDownloadCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/addEVCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebAddEVCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/uploadCompanyCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebUploadCompanyCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCompanyCertificates",
		permissionSystem,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebListCompanyCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/grantAppEVCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebGrantAppEVCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/getCertificatePassword/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebGetCertificatePassword,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCompanyCertificate",
		permissionSystem,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebDownloadCompanyCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listGrantCertificateApps",
		permissionSystem,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebListGrantCertificateApps,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitSigningJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.WindowsWebSubmitSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listSigningJobs/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebListSigningJobs,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitWHQLJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.WindowsWebSubmitWHQLJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listWHQLJobs/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.WindowsWebListWHQLJobs,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/removeCompanyCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.WindowsWebRemoveCompanyCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/deleteCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.WindowsWebDeleteCertificate,
	)
}

func webApple(r *gin.RouterGroup) {
	r = r.Group("/apple", filter.WebAuthenticateFilter)

	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyBundleID/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebApplyBundleID,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/modifyBundleID/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebModifyBundleID,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebApplyCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listBundleIDs/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebListBundleIDs,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listCertificates",
		permissionSystem,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebListCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/registerDevice/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebRegisterDevice,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listDevices/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebListDevices,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyProfile/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebApplyProfile,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyInHouseProfile/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebApplyInHouseProfile,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyCommonProfile/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebApplyCommonProfile,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/applyPushCertificate/:"+consts.HTTPPathAppID,
		permissionApp,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebApplyPushCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/deleteBundleID/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebDeleteBundleID,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/removeCertificate",
		permissionSystem,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebRemoveCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listAppCertificates/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebListAppCertificates,
	)
	addRouteWithPermissions(
		r,
		http.MethodPost,
		"/submitSigningJob/:"+consts.HTTPPathAppID,
		permissionAppSign,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebSubmitSigningJob,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/listSigningJobs/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebListSigningJobs,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/removeProfile/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebRemoveProfile,
	)
	addRouteWithPermissions(
		r,
		http.MethodDelete,
		"/removePushCertificate/:"+consts.HTTPPathAppID,
		permissionAppAdmin,
		filter.AntiShakeFilter,
		filter.PermissionWebAuthenticateFilter,
		filter.DatabaseTransactionFilter,
		api.AppleWebRemovePushCertificate,
	)
	addRouteWithPermissions(
		r,
		http.MethodGet,
		"/downloadCertificate/:"+consts.HTTPPathAppID,
		permissionAppRead,
		filter.PermissionWebAuthenticateFilter,
		api.AppleWebDownloadCertificate,
	)
}

func webNotice(r *gin.RouterGroup) {
	r = r.Group("/notice")

	addRoute(
		r,
		http.MethodGet,
		"/last",
		api.NoticeWebLast,
	)
}

func addRoute(r *gin.RouterGroup, method string, path string, handlers ...gin.HandlerFunc) {
	addRouteWithPermissions(r, method, path, nil, handlers...)
}

func addRouteWithPermissions(r *gin.RouterGroup, method string, path string, perms []int, handlers ...gin.HandlerFunc) {
	if len(handlers) <= 0 {
		panic("handlers must have at least one handler")
	}
	if len(perms) > 0 {
		filter.AddPathAuthorities(r.BasePath()+path, perms...)
	}
	switch method {
	case http.MethodGet:
		r.GET(path, handlers...)
	case http.MethodPost:
		r.POST(path, handlers...)
	case http.MethodPut:
		r.PUT(path, handlers...)
	case http.MethodPatch:
		r.PATCH(path, handlers...)
	case http.MethodHead:
		r.HEAD(path, handlers...)
	case http.MethodOptions:
		r.OPTIONS(path, handlers...)
	case http.MethodDelete:
		r.DELETE(path, handlers...)
	default:
		panic("unknown method: " + method)
	}
}
