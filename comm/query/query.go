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

package query

import (
	"context"

	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/comm/consts"
)

type APIAccountDo struct {
	aPIAccount
	mocker[IAPIAccountDo]
}

type APIAuthorizationDo struct {
	aPIAuthorization
	mocker[IAPIAuthorizationDo]
}

type AesKeyDo struct {
	aesKey
	mocker[IAesKeyDo]
}

type AndroidCertificateDo struct {
	androidCertificate
	mocker[IAndroidCertificateDo]
}

type AndroidOrganizationDo struct {
	androidOrganization
	mocker[IAndroidOrganizationDo]
}

type AndroidSigningJobDo struct {
	androidSigningJob
	mocker[IAndroidSigningJobDo]
}

type AppDo struct {
	app
	mocker[IAppDo]
}

type AppleBundleIDDo struct {
	appleBundleID
	mocker[IAppleBundleIDDo]
}

type AppleCertificateDo struct {
	appleCertificate
	mocker[IAppleCertificateDo]
}

type AppleDeviceDo struct {
	appleDevice
	mocker[IAppleDeviceDo]
}

type AppleProfileDo struct {
	appleProfile
	mocker[IAppleProfileDo]
}

type AppleSigningJobDo struct {
	appleSigningJob
	mocker[IAppleSigningJobDo]
}

type EventDo struct {
	event
	mocker[IEventDo]
}

type FileDo struct {
	file
	mocker[IFileDo]
}

type NoticeDo struct {
	notice
	mocker[INoticeDo]
}

type TodoDo struct {
	todo
	mocker[ITodoDo]
}

type UserDo struct {
	user
	mocker[IUserDo]
}

type UserRoleDo struct {
	userRole
	mocker[IUserRoleDo]
}

type WhqlJobDo struct {
	whqlJob
	mocker[IWhqlJobDo]
}

type WindowsCertificateDo struct {
	windowsCertificate
	mocker[IWindowsCertificateDo]
}

type WindowsCertificateAuthorizationDo struct {
	windowsCertificateAuthorization
	mocker[IWindowsCertificateAuthorizationDo]
}

type WindowsSigningJobDo struct {
	windowsSigningJob
	mocker[IWindowsSigningJobDo]
}

type DB struct {
	*Query
	APIAccount                      APIAccountDo
	APIAuthorization                APIAuthorizationDo
	AesKey                          AesKeyDo
	AndroidCertificate              AndroidCertificateDo
	AndroidOrganization             AndroidOrganizationDo
	AndroidSigningJob               AndroidSigningJobDo
	App                             AppDo
	AppleBundleID                   AppleBundleIDDo
	AppleCertificate                AppleCertificateDo
	AppleDevice                     AppleDeviceDo
	AppleProfile                    AppleProfileDo
	AppleSigningJob                 AppleSigningJobDo
	Event                           EventDo
	File                            FileDo
	Notice                          NoticeDo
	Todo                            TodoDo
	User                            UserDo
	UserRole                        UserRoleDo
	WhqlJob                         WhqlJobDo
	WindowsCertificate              WindowsCertificateDo
	WindowsCertificateAuthorization WindowsCertificateAuthorizationDo
	WindowsSigningJob               WindowsSigningJobDo
}

type mocker[Do any] struct {
	skipTablenameFn   bool
	forUnitTestMockFn func(context.Context) Do
}

type contexter[I any] interface {
	WithContext(context.Context) I
}

// GetDB 获取 *gorm.DB。
func GetDB(q *DB) *gorm.DB {
	return q.db
}

func NewDB(q *Query) *DB {
	return &DB{
		Query: q,
		APIAccount: APIAccountDo{
			aPIAccount: q.APIAccount,
		},
		APIAuthorization: APIAuthorizationDo{
			aPIAuthorization: q.APIAuthorization,
		},
		AesKey: AesKeyDo{
			aesKey: q.AesKey,
		},
		AndroidCertificate: AndroidCertificateDo{
			androidCertificate: q.AndroidCertificate,
		},
		AndroidOrganization: AndroidOrganizationDo{
			androidOrganization: q.AndroidOrganization,
		},
		AndroidSigningJob: AndroidSigningJobDo{
			androidSigningJob: q.AndroidSigningJob,
		},
		App: AppDo{
			app: q.App,
		},
		AppleBundleID: AppleBundleIDDo{
			appleBundleID: q.AppleBundleID,
		},
		AppleCertificate: AppleCertificateDo{
			appleCertificate: q.AppleCertificate,
		},
		AppleDevice: AppleDeviceDo{
			appleDevice: q.AppleDevice,
		},
		AppleProfile: AppleProfileDo{
			appleProfile: q.AppleProfile,
		},
		AppleSigningJob: AppleSigningJobDo{
			appleSigningJob: q.AppleSigningJob,
		},
		Event: EventDo{
			event: q.Event,
		},
		File: FileDo{
			file: q.File,
		},
		Notice: NoticeDo{
			notice: q.Notice,
		},
		Todo: TodoDo{
			todo: q.Todo,
		},
		User: UserDo{
			user: q.User,
		},
		UserRole: UserRoleDo{
			userRole: q.UserRole,
		},
		WhqlJob: WhqlJobDo{
			whqlJob: q.WhqlJob,
		},
		WindowsCertificate: WindowsCertificateDo{
			windowsCertificate: q.WindowsCertificate,
		},
		WindowsCertificateAuthorization: WindowsCertificateAuthorizationDo{
			windowsCertificateAuthorization: q.WindowsCertificateAuthorization,
		},
		WindowsSigningJob: WindowsSigningJobDo{
			windowsSigningJob: q.WindowsSigningJob,
		},
	}
}

func mockWithContext[I any](ctx context.Context, mocker *mocker[I], do contexter[I]) I {
	if consts.UnitTestMode() && mocker.forUnitTestMockFn != nil {
		return mocker.forUnitTestMockFn(ctx)
	}
	return do.WithContext(ctx)
}

func (d *APIAccountDo) WithContext(ctx context.Context) IAPIAccountDo {
	return mockWithContext(ctx, &d.mocker, d.aPIAccount)
}

func (d *APIAuthorizationDo) WithContext(ctx context.Context) IAPIAuthorizationDo {
	return mockWithContext(ctx, &d.mocker, d.aPIAuthorization)
}

func (d *AesKeyDo) WithContext(ctx context.Context) IAesKeyDo {
	return mockWithContext(ctx, &d.mocker, d.aesKey)
}

func (d *AndroidCertificateDo) WithContext(ctx context.Context) IAndroidCertificateDo {
	return mockWithContext(ctx, &d.mocker, d.androidCertificate)
}

func (d *AndroidOrganizationDo) WithContext(ctx context.Context) IAndroidOrganizationDo {
	return mockWithContext(ctx, &d.mocker, d.androidOrganization)
}

func (d *AndroidSigningJobDo) WithContext(ctx context.Context) IAndroidSigningJobDo {
	return mockWithContext(ctx, &d.mocker, d.androidSigningJob)
}

func (d *AndroidSigningJobDo) Table(newTableName string) *AndroidSigningJobDo {
	if consts.UnitTestMode() && d.skipTablenameFn {
		return d
	}
	return &AndroidSigningJobDo{
		androidSigningJob: *d.androidSigningJob.Table(newTableName),
		mocker:            d.mocker,
	}
}

func (d *AppDo) WithContext(ctx context.Context) IAppDo {
	return mockWithContext(ctx, &d.mocker, d.app)
}

func (d *AppleBundleIDDo) WithContext(ctx context.Context) IAppleBundleIDDo {
	return mockWithContext(ctx, &d.mocker, d.appleBundleID)
}

func (d *AppleCertificateDo) WithContext(ctx context.Context) IAppleCertificateDo {
	return mockWithContext(ctx, &d.mocker, d.appleCertificate)
}

func (d *AppleDeviceDo) WithContext(ctx context.Context) IAppleDeviceDo {
	return mockWithContext(ctx, &d.mocker, d.appleDevice)
}

func (d *AppleProfileDo) WithContext(ctx context.Context) IAppleProfileDo {
	return mockWithContext(ctx, &d.mocker, d.appleProfile)
}

func (d *AppleSigningJobDo) WithContext(ctx context.Context) IAppleSigningJobDo {
	return mockWithContext(ctx, &d.mocker, d.appleSigningJob)
}

func (d *AppleSigningJobDo) Table(newTableName string) *AppleSigningJobDo {
	if consts.UnitTestMode() && d.skipTablenameFn {
		return d
	}
	return &AppleSigningJobDo{
		appleSigningJob: *d.appleSigningJob.Table(newTableName),
		mocker:          d.mocker,
	}
}

func (d *EventDo) WithContext(ctx context.Context) IEventDo {
	return mockWithContext(ctx, &d.mocker, d.event)
}

func (d *EventDo) Table(newTableName string) *EventDo {
	if consts.UnitTestMode() && d.skipTablenameFn {
		return d
	}
	return &EventDo{
		event:  *d.event.Table(newTableName),
		mocker: d.mocker,
	}
}

func (d *FileDo) WithContext(ctx context.Context) IFileDo {
	return mockWithContext(ctx, &d.mocker, d.file)
}

func (d *FileDo) Table(newTableName string) *FileDo {
	if consts.UnitTestMode() && d.skipTablenameFn {
		return d
	}
	return &FileDo{
		file:   *d.file.Table(newTableName),
		mocker: d.mocker,
	}
}

func (d *NoticeDo) WithContext(ctx context.Context) INoticeDo {
	return mockWithContext(ctx, &d.mocker, d.notice)
}

func (d *TodoDo) WithContext(ctx context.Context) ITodoDo {
	return mockWithContext(ctx, &d.mocker, d.todo)
}

func (d *UserDo) WithContext(ctx context.Context) IUserDo {
	return mockWithContext(ctx, &d.mocker, d.user)
}

func (d *UserRoleDo) WithContext(ctx context.Context) IUserRoleDo {
	return mockWithContext(ctx, &d.mocker, d.userRole)
}

func (d *WhqlJobDo) WithContext(ctx context.Context) IWhqlJobDo {
	return mockWithContext(ctx, &d.mocker, d.whqlJob)
}

func (d *WindowsCertificateDo) WithContext(ctx context.Context) IWindowsCertificateDo {
	return mockWithContext(ctx, &d.mocker, d.windowsCertificate)
}

func (d *WindowsCertificateAuthorizationDo) WithContext(ctx context.Context) IWindowsCertificateAuthorizationDo {
	return mockWithContext(ctx, &d.mocker, d.windowsCertificateAuthorization)
}

func (d *WindowsSigningJobDo) WithContext(ctx context.Context) IWindowsSigningJobDo {
	return mockWithContext(ctx, &d.mocker, d.windowsSigningJob)
}

func (d *WindowsSigningJobDo) Table(newTableName string) *WindowsSigningJobDo {
	if consts.UnitTestMode() && d.skipTablenameFn {
		return d
	}
	return &WindowsSigningJobDo{
		windowsSigningJob: *d.windowsSigningJob.Table(newTableName),
		mocker:            d.mocker,
	}
}
