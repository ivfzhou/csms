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
	methodTypeWindowsSigningJobGetTables
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
	EventCount2Once(r int, err error) DBClientMocker[Table]
	AndroidSigningJobCount2Once(r int, err error) DBClientMocker[Table]
	EventListOnce(r []*Table, err error) DBClientMocker[Table]
	EventCountTypesWithDayOnce(r []map[string]any, err error) DBClientMocker[Table]
	EventCountTypesWithWeekOnce(r []map[string]any, err error) DBClientMocker[Table]
	EventCountTypesWithMonthOnce(r []map[string]any, err error) DBClientMocker[Table]
	AndroidSigningJobListOnce(r []*Table, err error) DBClientMocker[Table]
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
	panic("not implement yet")
}

func (c *windowsSigningJobDo[Do]) List([]string, int, string, int, []int, []int, int) ([]*model.WindowsSigningJob, error) {
	panic("not implement yet")
}

func (c *windowsSigningJobDo[Do]) Count2([]string, int, string, int, []int, []int) (int, error) {
	panic("not implement yet")
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

func (c *appleSigningJobDo[Do]) List([]string, int, int) ([]*model.AppleSigningJob, error) {
	panic("not implement yet")
}

func (c *appleSigningJobDo[Do]) Count2([]string, int) (int, error) {
	panic("not implement yet")
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
