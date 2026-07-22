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
	"fmt"
	"reflect"
	"sync"
	"time"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/query"
	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

const (
	methodTypeScan doMethod = iota + 1
	methodTypeCreateInBatches
	methodTypeCreate
	methodTypeTake
	methodTypeDelete
	methodTypeCount
	methodTypeUpdateColumnSimple
	methodTypeUserSearchByName
	methodTypeFind
	methodTypeFindByPage
	methodTypePreload
	methodTypeFirstOrCreate
	methodTypeSave
	methodTypeFirstOrInit
	methodTypeScanByPage
	methodTypeLast
	methodTypeFirst
	methodTypeFindInBatches
	methodTypeFindInBatch
	methodTypeEventGetTables
	methodTypeEventCount2
	methodTypeEventList
	methodTypeEventCountTypesWithDay
	methodTypeEventCountTypesWithWeek
	methodTypeEventCountTypesWithMonth
	methodTypeAndroidSigningJobGetTables
	methodTypeAppleSigningJobGetTables
	methodTypeAndroidSigningJobCount2
	methodTypeAndroidSigningJobList
	methodTypeAndroidSigningJobCountWithDay
	methodTypeAndroidSigningJobCountWithWeek
	methodTypeAndroidSigningJobCountWithMonth
	methodTypeAndroidSigningJobCostWithDay
	methodTypeAndroidSigningJobCostWithWeek
	methodTypeAndroidSigningJobCostWithMonth
	methodTypeAndroidSigningJobPassRateWithDay
	methodTypeAndroidSigningJobPassRateWithWeek
	methodTypeAndroidSigningJobPassRateWithMonth
	methodTypeWindowsSigningJobGetTables
	methodTypeWindowsSigningJobCount2
	methodTypeWindowsSigningJobList
	methodTypeWindowsSigningJobGetJobIDByStatus
	methodTypeWindowsSigningJobCountWithDay
	methodTypeWindowsSigningJobCountWithWeek
	methodTypeWindowsSigningJobCountWithMonth
	methodTypeWindowsSigningJobCostWithDay
	methodTypeWindowsSigningJobCostWithWeek
	methodTypeWindowsSigningJobCostWithMonth
	methodTypeWindowsSigningJobPassRateWithDay
	methodTypeWindowsSigningJobPassRateWithWeek
	methodTypeWindowsSigningJobPassRateWithMonth
	methodTypeWhqlJobCountWithDay
	methodTypeWhqlJobCountWithWeek
	methodTypeWhqlJobCountWithMonth
	methodTypeWhqlJobCostWithDay
	methodTypeWhqlJobCostWithWeek
	methodTypeWhqlJobCostWithMonth
	methodTypeWhqlJobPassRateWithDay
	methodTypeWhqlJobPassRateWithWeek
	methodTypeWhqlJobPassRateWithMonth
	methodTypeAppleSigningJobCount2
	methodTypeAppleSigningJobList
	methodTypeAppleSigningJobCountWithDay
	methodTypeAppleSigningJobCountWithWeek
	methodTypeAppleSigningJobCountWithMonth
	methodTypeAppleSigningJobCostWithDay
	methodTypeAppleSigningJobCostWithWeek
	methodTypeAppleSigningJobCostWithMonth
	methodTypeAppleSigningJobPassRateWithDay
	methodTypeAppleSigningJobPassRateWithWeek
	methodTypeAppleSigningJobPassRateWithMonth
)

type DBClientMocker[Table any] interface {
	ScanOnce(fn func(any), err error) DBClientMocker[Table]
	CreateInBatchesOnce(err error) DBClientMocker[Table]
	CreateOnce(err error) DBClientMocker[Table]
	TakeOnce(u *Table, err error) DBClientMocker[Table]
	DeleteOnce(r gen.ResultInfo, err error) DBClientMocker[Table]
	CountOnce(c int64, err error) DBClientMocker[Table]
	UpdateColumnSimpleOnce(r gen.ResultInfo, err error) DBClientMocker[Table]
	UserSearchByNameOnce(u []*Table, err error) DBClientMocker[Table]
	FindOnce(u []*Table, err error) DBClientMocker[Table]
	EventGetTablesOnce([]string, error) DBClientMocker[Table]
	AndroidSigningJobGetTablesOnce([]string, error) DBClientMocker[Table]
	WindowsSigningJobGetTablesOnce([]string, error) DBClientMocker[Table]
	AppleSigningJobGetTablesOnce([]string, error) DBClientMocker[Table]
	EventCount2Once(r int, err error) DBClientMocker[Table]
	AndroidSigningJobCount2Once(r int, err error) DBClientMocker[Table]
	EventListOnce(r []*Table, err error) DBClientMocker[Table]
	EventCountTypesWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	EventCountTypesWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	EventCountTypesWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobListOnce(r []*Table, err error) DBClientMocker[Table]
	AndroidSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCount2Once(r int, err error) DBClientMocker[Table]
	WindowsSigningJobListOnce(r []*Table, err error) DBClientMocker[Table]
	WindowsSigningJobGetJobIDByStatusOnce(r []string, err error) DBClientMocker[Table]
	WindowsSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WindowsSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	WhqlJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCount2Once(r int, err error) DBClientMocker[Table]
	AppleSigningJobListOnce(r []*Table, err error) DBClientMocker[Table]
	AppleSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	AppleSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	LastOnce(t *Table, err error) DBClientMocker[Table]
	Reset()
}

type doer interface {
	append(d *doResultData)
}

type doMethod int

type doResultData struct {
	method doMethod
	values []any
}

type commonDo[Table, Do any] struct {
	gen.DO
	typeName  string
	datasLock *sync.Mutex
	datas     []*doResultData
	self      Do
}

type eventDo[Do any] struct {
	*commonDo[model.Event, Do]
}

type androidSigningJobDo[Do any] struct {
	*commonDo[model.AndroidSigningJob, Do]
}

type windowsSigningJobDo[Do any] struct {
	*commonDo[model.WindowsSigningJob, Do]
}

type whqlJobDo[Do any] struct {
	*commonDo[model.WhqlJob, Do]
}

type appleSigningJobDo[Do any] struct {
	*commonDo[model.AppleSigningJob, Do]
}

type userDo[Do any] struct {
	*commonDo[model.User, Do]
}

type dbClientMocker[Table, Do any] struct {
	do    doer
	reset func()
}

func MockDBClient[Table any](ctx context.Context) DBClientMocker[Table] {
	typ := reflect.TypeFor[Table]()
	switch {
	case typ.ConvertibleTo(reflect.TypeFor[model.AesKey]()):
		return mockDBClient[Table, query.IAesKeyDo](ctx, "AesKey")
	case typ.ConvertibleTo(reflect.TypeFor[model.AndroidCertificate]()):
		return mockDBClient[Table, query.IAndroidCertificateDo](ctx, "AndroidCertificate")
	case typ.ConvertibleTo(reflect.TypeFor[model.AndroidOrganization]()):
		return mockDBClient[Table, query.IAndroidOrganizationDo](ctx, "AndroidOrganization")
	case typ.ConvertibleTo(reflect.TypeFor[model.AndroidSigningJob]()):
		return mockDBClient[Table, query.IAndroidSigningJobDo](ctx, "AndroidSigningJob")
	case typ.ConvertibleTo(reflect.TypeFor[model.APIAccount]()):
		return mockDBClient[Table, query.IAPIAccountDo](ctx, "APIAccount")
	case typ.ConvertibleTo(reflect.TypeFor[model.APIAuthorization]()):
		return mockDBClient[Table, query.IAPIAuthorizationDo](ctx, "APIAuthorization")
	case typ.ConvertibleTo(reflect.TypeFor[model.App]()):
		return mockDBClient[Table, query.IAppDo](ctx, "App")
	case typ.ConvertibleTo(reflect.TypeFor[model.AppleBundleID]()):
		return mockDBClient[Table, query.IAppleBundleIDDo](ctx, "AppleBundleID")
	case typ.ConvertibleTo(reflect.TypeFor[model.AppleCertificate]()):
		return mockDBClient[Table, query.IAppleCertificateDo](ctx, "AppleCertificate")
	case typ.ConvertibleTo(reflect.TypeFor[model.AppleDevice]()):
		return mockDBClient[Table, query.IAppleDeviceDo](ctx, "AppleDevice")
	case typ.ConvertibleTo(reflect.TypeFor[model.AppleProfile]()):
		return mockDBClient[Table, query.IAppleProfileDo](ctx, "AppleProfile")
	case typ.ConvertibleTo(reflect.TypeFor[model.AppleSigningJob]()):
		return mockDBClient[Table, query.IAppleSigningJobDo](ctx, "AppleSigningJob")
	case typ.ConvertibleTo(reflect.TypeFor[model.Event]()):
		return mockDBClient[Table, query.IEventDo](ctx, "Event")
	case typ.ConvertibleTo(reflect.TypeFor[model.UserRole]()):
		return mockDBClient[Table, query.IUserRoleDo](ctx, "UserRole")
	case typ.ConvertibleTo(reflect.TypeFor[model.File]()):
		return mockDBClient[Table, query.IFileDo](ctx, "File")
	case typ.ConvertibleTo(reflect.TypeFor[model.Notice]()):
		return mockDBClient[Table, query.INoticeDo](ctx, "Notice")
	case typ.ConvertibleTo(reflect.TypeFor[model.Todo]()):
		return mockDBClient[Table, query.ITodoDo](ctx, "Todo")
	case typ.ConvertibleTo(reflect.TypeFor[model.User]()):
		return mockDBClient[Table, query.IUserDo](ctx, "User")
	case typ.ConvertibleTo(reflect.TypeFor[model.UserRole]()):
		return mockDBClient[Table, query.IUserRoleDo](ctx, "UserRole")
	case typ.ConvertibleTo(reflect.TypeFor[model.WhqlJob]()):
		return mockDBClient[Table, query.IWhqlJobDo](ctx, "WhqlJob")
	case typ.ConvertibleTo(reflect.TypeFor[model.WindowsCertificate]()):
		return mockDBClient[Table, query.IWindowsCertificateDo](ctx, "WindowsCertificate")
	case typ.ConvertibleTo(reflect.TypeFor[model.WindowsCertificateAuthorization]()):
		return mockDBClient[Table, query.IWindowsCertificateAuthorizationDo](ctx, "WindowsCertificateAuthorization")
	case typ.ConvertibleTo(reflect.TypeFor[model.WindowsSigningJob]()):
		return mockDBClient[Table, query.IWindowsSigningJobDo](ctx, "WindowsSigningJob")
	default:
		panic("unhandled db do")
	}
}

func mockDBClient[Table, Do any](ctx context.Context, name string) DBClientMocker[Table] {
	var do doer
	var dbDo Do
	switch name {
	case "User":
		userDoVar := &userDo[Do]{&commonDo[model.User, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = userDoVar
		userDoVar.self = doI.(Do)
		dbDo = userDoVar.self
		do = userDoVar
	case "WindowsSigningJob":
		windowsSigningJobDoVar := &windowsSigningJobDo[Do]{&commonDo[model.WindowsSigningJob, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = windowsSigningJobDoVar
		windowsSigningJobDoVar.self = doI.(Do)
		dbDo = windowsSigningJobDoVar.self
		do = windowsSigningJobDoVar
	case "WhqlJob":
		whqlJobDoVar := &whqlJobDo[Do]{&commonDo[model.WhqlJob, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = whqlJobDoVar
		whqlJobDoVar.self = doI.(Do)
		dbDo = whqlJobDoVar.self
		do = whqlJobDoVar
	case "AppleSigningJob":
		appleSigningJobDoVar := &appleSigningJobDo[Do]{&commonDo[model.AppleSigningJob, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = appleSigningJobDoVar
		appleSigningJobDoVar.self = doI.(Do)
		dbDo = appleSigningJobDoVar.self
		do = appleSigningJobDoVar
	case "Event":
		eventDoVar := &eventDo[Do]{&commonDo[model.Event, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = eventDoVar
		eventDoVar.self = doI.(Do)
		dbDo = eventDoVar.self
		do = eventDoVar
	case "AndroidSigningJob":
		androidSigningJobDoVar := &androidSigningJobDo[Do]{&commonDo[model.AndroidSigningJob, Do]{datasLock: new(sync.Mutex), typeName: name}}
		var doI any = androidSigningJobDoVar
		androidSigningJobDoVar.self = doI.(Do)
		dbDo = androidSigningJobDoVar.self
		do = androidSigningJobDoVar
	default:
		commonDoVar := &commonDo[Table, Do]{datasLock: new(sync.Mutex), typeName: name}
		var doI any = commonDoVar
		commonDoVar.self = doI.(Do)
		dbDo = commonDoVar.self
		do = commonDoVar
	}
	reset2 := mvt.Chain(conn.MySQLClient(ctx)).Elem().FieldByName(name).FieldByName("mocker").
		FieldByName("skipTablenameFn").Set(true).Reset
	reset := mvt.Chain(conn.MySQLClient(ctx)).Elem().FieldByName(name).FieldByName("mocker").
		FieldByName("forUnitTestMockFn").Set(func(context.Context) Do { return dbDo }).Reset
	return &dbClientMocker[Table, Do]{do, func() {
		if reset2 != nil {
			reset2()
		}
		reset()
	}}
}

func (m *dbClientMocker[Table, Do]) ScanOnce(fn func(any), err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeScan, values: []any{fn, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) CreateInBatchesOnce(err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeCreateInBatches, values: []any{err}})
	return m
}

func (m *dbClientMocker[Table, Do]) CreateOnce(err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeCreate, values: []any{err}})
	return m
}

func (m *dbClientMocker[Table, Do]) TakeOnce(t *Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeTake, values: []any{t, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) DeleteOnce(r gen.ResultInfo, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeDelete, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) CountOnce(c int64, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeCount, values: []any{c, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) UpdateColumnSimpleOnce(r gen.ResultInfo, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeUpdateColumnSimple, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) UserSearchByNameOnce(t []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeUserSearchByName, values: []any{t, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) FindOnce(t []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeFind, values: []any{t, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventGetTablesOnce(r []string, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventGetTables, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobGetTablesOnce(r []string, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobGetTables, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventCount2Once(r int, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventCount2, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCount2Once(r int, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCount2, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) LastOnce(t *Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeLast, values: []any{t, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventListOnce(r []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventList, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventCountTypesWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventCountTypesWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventCountTypesWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventCountTypesWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) EventCountTypesWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeEventCountTypesWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobListOnce(r []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobList, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCountWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCountWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCountWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCostWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCostWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobCostWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobPassRateWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobPassRateWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AndroidSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAndroidSigningJobPassRateWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobGetTablesOnce(r []string, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobGetTables, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCount2Once(r int, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCount2, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobListOnce(r []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobList, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobGetJobIDByStatusOnce(r []string, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobGetJobIDByStatus, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCountWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCountWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCountWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCostWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCostWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobCostWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobPassRateWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobPassRateWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WindowsSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWindowsSigningJobPassRateWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCountWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCountWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCountWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCostWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCostWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobCostWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobPassRateWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobPassRateWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) WhqlJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeWhqlJobPassRateWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobGetTablesOnce(r []string, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobGetTables, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCount2Once(r int, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCount2, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobListOnce(r []*Table, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobList, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCountWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCountWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCountWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCountWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCountWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCountWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCostWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCostWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCostWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCostWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobCostWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobCostWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobPassRateWithDayOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobPassRateWithDay, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobPassRateWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobPassRateWithWeek, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) AppleSigningJobPassRateWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table] {
	m.do.append(&doResultData{method: methodTypeAppleSigningJobPassRateWithMonth, values: []any{r, err}})
	return m
}

func (m *dbClientMocker[Table, Do]) Reset() {
	m.reset()
}

func (c *commonDo[Table, Do]) Create(...*Table) error {
	data := c.getResultData(methodTypeCreate)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	panic(fmt.Sprintf("unhandle db %s Create", c.typeName))
}

func (c *commonDo[Table, Do]) CreateInBatches([]*Table, int) error {
	data := c.getResultData(methodTypeCreateInBatches)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	panic(fmt.Sprintf("unhandle db %s CreateInBatches", c.typeName))
}

func (c *commonDo[Table, Do]) Take() (*Table, error) {
	data := c.getResultData(methodTypeTake)
	if data != nil {
		user, _ := data[0].(*Table)
		err, _ := data[1].(error)
		return user, err
	}
	panic(fmt.Sprintf("unhandle db %s Take", c.typeName))
}

func (c *commonDo[Table, Do]) Find() ([]*Table, error) {
	data := c.getResultData(methodTypeFind)
	if data != nil {
		users, _ := data[0].([]*Table)
		err, _ := data[1].(error)
		return users, err
	}
	panic(fmt.Sprintf("unhandle db %s Find", c.typeName))
}

func (c *commonDo[Table, Do]) Delete(...*Table) (gen.ResultInfo, error) {
	data := c.getResultData(methodTypeDelete)
	if data != nil {
		r, _ := data[0].(gen.ResultInfo)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s Delete", c.typeName))
}

func (c *commonDo[Table, Do]) FindByPage(int, int) ([]*Table, int64, error) {
	data := c.getResultData(methodTypeFindByPage)
	if data != nil {
		tables, _ := data[0].([]*Table)
		count, _ := data[1].(int64)
		err, _ := data[2].(error)
		return tables, count, err
	}
	panic(fmt.Sprintf("unhandle db %s FindByPage", c.typeName))
}

func (c *commonDo[Table, Do]) Count() (int64, error) {
	data := c.getResultData(methodTypeCount)
	if data != nil {
		err, _ := data[1].(error)
		return data[0].(int64), err
	}
	panic(fmt.Sprintf("unhandle db %s Count", c.typeName))
}

func (c *commonDo[Table, Do]) UpdateColumnSimple(...field.AssignExpr) (gen.ResultInfo, error) {
	data := c.getResultData(methodTypeUpdateColumnSimple)
	if data != nil {
		r, _ := data[0].(gen.ResultInfo)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s UpdateColumnSimple", c.typeName))
}

func (c *commonDo[Table, Do]) FindInBatch(int, func(gen.Dao, int) error) ([]*Table, error) {
	data := c.getResultData(methodTypeFindInBatch)
	if data != nil {
		tables, _ := data[0].([]*Table)
		err, _ := data[1].(error)
		return tables, err
	}
	panic(fmt.Sprintf("unhandle db %s FindInBatch", c.typeName))
}

func (c *commonDo[Table, Do]) FindInBatches(*[]*Table, int, func(gen.Dao, int) error) error {
	data := c.getResultData(methodTypeFindInBatches)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	panic(fmt.Sprintf("unhandle db %s FindInBatches", c.typeName))
}

func (c *commonDo[Table, Do]) First() (*Table, error) {
	data := c.getResultData(methodTypeFirst)
	if data != nil {
		table, _ := data[0].(*Table)
		err, _ := data[1].(error)
		return table, err
	}
	panic(fmt.Sprintf("unhandle db %s First", c.typeName))
}

func (c *commonDo[Table, Do]) Last() (*Table, error) {
	data := c.getResultData(methodTypeLast)
	if data != nil {
		table, _ := data[0].(*Table)
		err, _ := data[1].(error)
		return table, err
	}
	panic(fmt.Sprintf("unhandle db %s Last", c.typeName))
}

func (c *commonDo[Table, Do]) ScanByPage(any, int, int) (int64, error) {
	data := c.getResultData(methodTypeScanByPage)
	if data != nil {
		r, _ := data[0].(int64)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s ScanByPage", c.typeName))
}

func (c *commonDo[Table, Do]) FirstOrInit() (*Table, error) {
	data := c.getResultData(methodTypeFirstOrInit)
	if data != nil {
		user, _ := data[0].(*Table)
		err, _ := data[1].(error)
		return user, err
	}
	panic(fmt.Sprintf("unhandle db %s FirstOrInit", c.typeName))
}

func (c *commonDo[Table, Do]) Save(...*Table) error {
	data := c.getResultData(methodTypeSave)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	panic(fmt.Sprintf("unhandle db %s Save", c.typeName))
}

func (c *commonDo[Table, Do]) FirstOrCreate() (*Table, error) {
	data := c.getResultData(methodTypeFirstOrCreate)
	if data != nil {
		user, _ := data[0].(*Table)
		err, _ := data[1].(error)
		return user, err
	}
	panic(fmt.Sprintf("unhandle db %s FirstOrCreate", c.typeName))
}

func (c *commonDo[Table, Do]) Scan(result any) error {
	data := c.getResultData(methodTypeScan)
	if data != nil {
		fn := data[0].(func(any))
		fn(result)
		err, _ := data[1].(error)
		return err
	}
	panic(fmt.Sprintf("unhandle db %s Scan", c.typeName))
}

func (c *commonDo[Table, Do]) Preload(args ...field.RelationField) Do {
	data := c.getResultData(methodTypePreload)
	if data != nil {
		fn, _ := data[0].(func(...field.RelationField))
		fn(args...)
	} else {
		panic(fmt.Sprintf("unhandle db %s Preload", c.typeName))
	}
	return c.self
}

func (c *commonDo[Table, Do]) Attrs(...field.AssignExpr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Assign(...field.AssignExpr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Joins(...field.RelationField) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Returning(any, ...string) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Debug() Do {
	return c.self
}

func (c *commonDo[Table, Do]) WithContext(context.Context) Do {
	return c.self
}

func (c *commonDo[Table, Do]) ReadDB() Do {
	return c.self
}

func (c *commonDo[Table, Do]) WriteDB() Do {
	return c.self
}

func (c *commonDo[Table, Do]) Session(*gorm.Session) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Clauses(...clause.Expression) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Not(...gen.Condition) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Or(...gen.Condition) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Select(...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Where(...gen.Condition) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Order(...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Distinct(...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Omit(...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Join(schema.Tabler, ...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) LeftJoin(schema.Tabler, ...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) RightJoin(schema.Tabler, ...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Group(...field.Expr) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Having(...gen.Condition) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Limit(int) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Offset(int) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Scopes(...func(gen.Dao) gen.Dao) Do {
	return c.self
}

func (c *commonDo[Table, Do]) Unscoped() Do {
	return c.self
}

func (c *commonDo[Table, Do]) getResultData(method doMethod) []any {
	if len(c.datas) <= 0 {
		return nil
	}
	c.datasLock.Lock()
	defer c.datasLock.Unlock()
	for i, data := range c.datas {
		if data.method == method {
			header := c.datas[:i]
			tail := c.datas[i+1:]
			c.datas = append(header, tail...)
			return data.values
		}
	}
	return nil
}

func (c *commonDo[Table, Do]) append(d *doResultData) {
	c.datas = append(c.datas, d)
}

func (c *eventDo[Do]) List([]string, []int, []int, time.Time, time.Time, int, int, int) ([]*model.Event, error) {
	data := c.getResultData(methodTypeEventList)
	if data != nil {
		r, _ := data[0].([]*model.Event)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s List", c.typeName))
}

func (c *eventDo[Do]) Count2([]string, []int, []int, time.Time, time.Time, int) (int, error) {
	data := c.getResultData(methodTypeEventCount2)
	if data != nil {
		r, _ := data[0].(int)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s Count2", c.typeName))
}

func (c *eventDo[Do]) CountTypesWithDay([]string, []int, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeEventCountTypesWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s CountTypesWithDay", c.typeName))
}

func (c *eventDo[Do]) CountTypesWithWeek([]string, []int, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeEventCountTypesWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s CountTypesWithWeek", c.typeName))
}

func (c *eventDo[Do]) CountTypesWithMonth([]string, []int, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeEventCountTypesWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s CountTypesWithMonth", c.typeName))
}

func (c *eventDo[Do]) GetTables(string) ([]string, error) {
	data := c.getResultData(methodTypeEventGetTables)
	if data != nil {
		r, _ := data[0].([]string)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s EventGetTables", c.typeName))
}

func (c *androidSigningJobDo[Do]) List([]string, int, string, int, []int, []int, int, int) ([]*model.AndroidSigningJob, error) {
	data := c.getResultData(methodTypeAndroidSigningJobList)
	if data != nil {
		r, _ := data[0].([]*model.AndroidSigningJob)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s List", c.typeName))
}

func (c *androidSigningJobDo[Do]) Count2([]string, int, string, int, []int, []int) (int, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCount2)
	if data != nil {
		r, _ := data[0].(int)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s Count2", c.typeName))
}

func (c *androidSigningJobDo[Do]) GetTables(string) ([]string, error) {
	data := c.getResultData(methodTypeAndroidSigningJobGetTables)
	if data != nil {
		r, _ := data[0].([]string)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobGetTables", c.typeName))
}

func (c *androidSigningJobDo[Do]) CountWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCountWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCountWithDay", c.typeName))
}

func (c *androidSigningJobDo[Do]) CountWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCountWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCountWithWeek", c.typeName))
}

func (c *androidSigningJobDo[Do]) CountWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCountWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCountWithMonth", c.typeName))
}

func (c *androidSigningJobDo[Do]) CostWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCostWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCostWithDay", c.typeName))
}

func (c *androidSigningJobDo[Do]) CostWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCostWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCostWithWeek", c.typeName))
}

func (c *androidSigningJobDo[Do]) CostWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobCostWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobCostWithMonth", c.typeName))
}

func (c *androidSigningJobDo[Do]) PassRateWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobPassRateWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobPassRateWithDay", c.typeName))
}

func (c *androidSigningJobDo[Do]) PassRateWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobPassRateWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobPassRateWithWeek", c.typeName))
}

func (c *androidSigningJobDo[Do]) PassRateWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAndroidSigningJobPassRateWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AndroidSigningJobPassRateWithMonth", c.typeName))
}

func (c *userDo[Do]) SearchByName(string) ([]*model.User, error) {
	data := c.getResultData(methodTypeUserSearchByName)
	if data != nil {
		tables, _ := data[0].([]*model.User)
		err, _ := data[1].(error)
		return tables, err
	}
	panic(fmt.Sprintf("unhandle db %s SearchByName", c.typeName))
}

func (c *windowsSigningJobDo[Do]) GetJobIDByStatus([]string, int) ([]string, error) {
	data := c.getResultData(methodTypeWindowsSigningJobGetJobIDByStatus)
	if data != nil {
		r, _ := data[0].([]string)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobGetJobIDByStatus", c.typeName))
}

func (c *windowsSigningJobDo[Do]) List([]string, int, string, int, int, []int, []int, int, int) ([]*model.WindowsSigningJob, error) {
	data := c.getResultData(methodTypeWindowsSigningJobList)
	if data != nil {
		r, _ := data[0].([]*model.WindowsSigningJob)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s List", c.typeName))
}

func (c *windowsSigningJobDo[Do]) Count2([]string, int, string, int, int, []int, []int) (int, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCount2)
	if data != nil {
		r, _ := data[0].(int)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s Count2", c.typeName))
}

func (c *windowsSigningJobDo[Do]) GetTables(string) ([]string, error) {
	data := c.getResultData(methodTypeWindowsSigningJobGetTables)
	if data != nil {
		r, _ := data[0].([]string)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobGetTables", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CountWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCountWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCountWithDay", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CountWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCountWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCountWithWeek", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CountWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCountWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCountWithMonth", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CostWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCostWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCostWithDay", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CostWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCostWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCostWithWeek", c.typeName))
}

func (c *windowsSigningJobDo[Do]) CostWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobCostWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobCostWithMonth", c.typeName))
}

func (c *windowsSigningJobDo[Do]) PassRateWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobPassRateWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobPassRateWithDay", c.typeName))
}

func (c *windowsSigningJobDo[Do]) PassRateWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobPassRateWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobPassRateWithWeek", c.typeName))
}

func (c *windowsSigningJobDo[Do]) PassRateWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWindowsSigningJobPassRateWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WindowsSigningJobPassRateWithMonth", c.typeName))
}

func (c *whqlJobDo[Do]) CountWithDay(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCountWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCountWithDay", c.typeName))
}

func (c *whqlJobDo[Do]) CountWithWeek(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCountWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCountWithWeek", c.typeName))
}

func (c *whqlJobDo[Do]) CountWithMonth(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCountWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCountWithMonth", c.typeName))
}

func (c *whqlJobDo[Do]) CostWithDay(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCostWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCostWithDay", c.typeName))
}

func (c *whqlJobDo[Do]) CostWithWeek(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCostWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCostWithWeek", c.typeName))
}

func (c *whqlJobDo[Do]) CostWithMonth(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobCostWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobCostWithMonth", c.typeName))
}

func (c *whqlJobDo[Do]) PassRateWithDay(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobPassRateWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobPassRateWithDay", c.typeName))
}

func (c *whqlJobDo[Do]) PassRateWithWeek(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobPassRateWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobPassRateWithWeek", c.typeName))
}

func (c *whqlJobDo[Do]) PassRateWithMonth(int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeWhqlJobPassRateWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s WhqlJobPassRateWithMonth", c.typeName))
}

func (c *appleSigningJobDo[Do]) List(_ []string, _ int, _ int, _ int) ([]*model.AppleSigningJob, error) {
	data := c.getResultData(methodTypeAppleSigningJobList)
	if data != nil {
		r, _ := data[0].([]*model.AppleSigningJob)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s List", c.typeName))
}

func (c *appleSigningJobDo[Do]) Count2([]string, int) (int, error) {
	data := c.getResultData(methodTypeAppleSigningJobCount2)
	if data != nil {
		r, _ := data[0].(int)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s Count2", c.typeName))
}

func (c *appleSigningJobDo[Do]) GetTables(string) ([]string, error) {
	data := c.getResultData(methodTypeAppleSigningJobGetTables)
	if data != nil {
		r, _ := data[0].([]string)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobGetTables", c.typeName))
}

func (c *appleSigningJobDo[Do]) CountWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCountWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCountWithDay", c.typeName))
}

func (c *appleSigningJobDo[Do]) CountWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCountWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCountWithWeek", c.typeName))
}

func (c *appleSigningJobDo[Do]) CountWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCountWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCountWithMonth", c.typeName))
}

func (c *appleSigningJobDo[Do]) CostWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCostWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCostWithDay", c.typeName))
}

func (c *appleSigningJobDo[Do]) CostWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCostWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCostWithWeek", c.typeName))
}

func (c *appleSigningJobDo[Do]) CostWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobCostWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobCostWithMonth", c.typeName))
}

func (c *appleSigningJobDo[Do]) PassRateWithDay([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobPassRateWithDay)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobPassRateWithDay", c.typeName))
}

func (c *appleSigningJobDo[Do]) PassRateWithWeek([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobPassRateWithWeek)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobPassRateWithWeek", c.typeName))
}

func (c *appleSigningJobDo[Do]) PassRateWithMonth([]string, int, time.Time, time.Time) ([]map[string]any, error) {
	data := c.getResultData(methodTypeAppleSigningJobPassRateWithMonth)
	if data != nil {
		r, _ := data[0].([]map[string]any)
		err, _ := data[1].(error)
		return r, err
	}
	panic(fmt.Sprintf("unhandle db %s AppleSigningJobPassRateWithMonth", c.typeName))
}
