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

package main

import (
	"flag"
	"strings"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/rawsql"
)

// 根据 SQL 文件生成数据库实体对应的 Go 代码。
func main() {
	dbFilePath := flag.String("database", "database.sql", "database file location")
	flag.Parse()

	generator := gen.NewGenerator(gen.Config{
		OutPath:           "query",
		ModelPkgPath:      "model",
		WithUnitTest:      false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithoutContext | gen.WithQueryInterface,
	})
	db, _ := gorm.Open(
		rawsql.New(rawsql.Config{
			FilePath: []string{*dbFilePath},
		}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "t_",
				SingularTable: true,
			},
		},
	)
	generator.UseDB(db)
	generator.WithDataTypeMap(map[string]func(gorm.ColumnType) string{
		"int":       func(gorm.ColumnType) string { return "int" },
		"tinyint":   func(gorm.ColumnType) string { return "int" },
		"bigint":    func(gorm.ColumnType) string { return "int" },
		"smallint":  func(gorm.ColumnType) string { return "int" },
		"mediumint": func(gorm.ColumnType) string { return "int" },
		"varchar": func(columnType gorm.ColumnType) string {
			name := columnType.Name()
			switch name {
			case "candidates", "signature_schemas":
				comment, _ := columnType.Comment()
				if strings.HasPrefix(comment, "待办可审批的人") || strings.HasPrefix(comment, "APK 签名方案") {
					return "IntList"
				}
			case "ip", "capabilities":
				comment, _ := columnType.Comment()
				if strings.HasPrefix(comment, "凭证可以放行的 IP") {
					return "StringList"
				}
			}
			return "string"
		},
		"mediumtext": func(columnType gorm.ColumnType) string {
			name := columnType.Name()
			switch name {
			case "capabilities":
				comment, _ := columnType.Comment()
				if strings.HasPrefix(comment, "BundleID 的能力项") {
					return "StringList"
				}
			}
			return "string"
		},
		"bit": func(columnType gorm.ColumnType) string {
			precision, _, _ := columnType.DecimalSize()
			if precision == 1 {
				return "Bool"
			}
			return "[]byte"
		},
	})

	type UserQuery interface {
		// select name_en, name_zh from t_user
		// where name_en like concat('%', concat(@name, '%')) or name_zh like concat('%', concat(@name, '%'))
		// order by case when name_en = @name or name_zh = @name then 1 else 2 end, id desc
		// limit 10
		SearchByName(name string) ([]*gen.T, error)
	}
	generator.ApplyInterface(func(UserQuery) {}, generator.GenerateModel("t_user"))

	type EventQuery interface {
		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db and TABLE_NAME like 't_event%'
		GetTables(db string) ([]string, error)

		// select * from (
		// {{ for i, t := range tables }}
		//    select * from @@t
		//    {{ if len(tables) - 1 != i }} union all {{ end }}
		//    where 1 = 1
		//    {{ if !begin.IsZero() }} and created_time >= @begin {{ end }}
		//    {{ if !end.IsZero() }} and created_time <= @end {{ end }}
		//    {{ if len(appIDs) > 0 }} and app_id in (@appIDs) {{ end }}
		//    {{ if typ > 0 }} and type = @typ {{ end }}
		//    {{ if len(userIDs) > 0 }} and user_id in (@userIDs) {{ end }}
		//    {{ end }} ) t
		// order by t.created_time desc, t.id desc limit @limit offset @offset
		List(tables []string, appIDs, userIDs []int, begin, end time.Time, typ int, limit, offset int) ([]*gen.T, error)

		// select sum(count) from (
		// {{ for i, t := range tables }}
		//     select count(*) `count` from @@t
		//     where 1 = 1
		//     {{ if !begin.IsZero() }} and created_time >= @begin {{ end }}
		//     {{ if !end.IsZero() }} and created_time <= @end {{ end }}
		//     {{ if len(appIDs) > 0 }} and app_id in (@appIDs) {{ end }}
		//     {{ if typ > 0 }} and type = @typ {{ end }}
		//     {{ if len(userIDs) > 0 }} and user_id in (@userIDs) {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		Count2(tables []string, appIDs, userIDs []int, begin, end time.Time, typ int) (int, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and type in (@types)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountTypesWithDay(tables []string, types []int, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and type in (@types)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountTypesWithWeek(tables []string, types []int, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and type in (@types)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountTypesWithMonth(tables []string, types []int, appID int, begin, end time.Time) ([]gen.M, error)
	}
	generator.ApplyInterface(func(EventQuery) {}, generator.GenerateModel("t_event"))

	type WindowsSigningJobQuery interface {
		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db
		// and TABLE_NAME like 't_windows_signing_job%'
		GetTables(db string) ([]string, error)

		// select * from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID and t.source in (1, 2)
		// {{ if signingType > 0 }} and t.type = @signingType {{ end }}
		// {{ if status > 0 }} and t.status = @status {{ end }}
		// {{ if len(keyWord) > 0 }}
		//   and (
		//     t.job_id = @keyWord or t.file_id = @keyWord or t.signed_file_id = @keyWord
		//     {{ if len(certificateIDs) > 0 }} or t.certificate_id in (@certificateIDs) {{ end }}
		//     {{ if len(userIDs) > 0 }} or t.user_id in (@userIDs) {{ end }}
		//   )
		// {{ end }}
		// order by t.created_time desc, t.id desc limit @limit offset @offset
		List(tables []string, appID int, keyWord string, signingType, status int, certificateIDs []int,
			userIDs []int, limit, offset int) ([]*gen.T, error)

		// select count(*) from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID and t.source in (1, 2)
		// {{ if signingType > 0 }} and t.type = @signingType {{ end }}
		// {{ if status > 0 }} and t.status = @status {{ end }}
		// {{ if len(keyWord) > 0 }}
		//   and (
		//     t.job_id = @keyWord or t.file_id = @keyWord or t.signed_file_id = @keyWord
		//     {{ if len(certificateIDs) > 0 }} or t.certificate_id in (@certificateIDs) {{ end }}
		//     {{ if len(userIDs) > 0 }} or t.user_id in (@userIDs) {{ end }}
		//   )
		// {{ end }}
		Count2(tables []string, appID int, keyWord string, signingType, status int, certificateIDs []int,
			userIDs []int) (int, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m%d') `day`, cast(round(sum(case t.status when 7 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(sum(case t.status when 7 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m') `day`, cast(round(sum(case t.status when 7 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (6, 7)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.job_id from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.status = @status
		GetJobIDByStatus(tables []string, status int) ([]string, error)
	}
	generator.ApplyInterface(func(WindowsSigningJobQuery) {}, generator.GenerateModel("t_windows_signing_job"))

	type AndroidSigningJobQuery interface {
		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db
		// and TABLE_NAME like 't_android_signing_job%'
		GetTables(db string) ([]string, error)

		// select * from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID
		// {{ if status > 0 }} and t.status = @status {{ end }}
		// {{ if len(certificateIDs) > 0 }} and t.certificate_id in (@certificateIDs) {{ end }}
		// {{ if len(keyWord) > 0 }}
		//   and (
		//     t.job_id = @keyWord or t.file_id = @keyWord or t.signed_file_id = @keyWord
		//     {{ if len(userIDs) > 0 }} or t.user_id in (@userIDs) {{ end }}
		//   )
		// {{ end }}
		// order by t.created_time desc, t.id desc limit @limit offset @offset
		List(tables []string, appID int, keyWord string, status int, certificateIDs []int, userIDs []int,
			limit, offset int) ([]*gen.T, error)

		// select count(*) from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID
		// {{ if status > 0 }} and t.status = @status {{ end }}
		// {{ if len(certificateIDs) > 0 }} and t.certificate_id in (@certificateIDs) {{ end }}
		// {{ if len(keyWord) > 0 }}
		//   and (
		//     t.job_id = @keyWord or t.file_id = @keyWord or t.signed_file_id = @keyWord
		//     {{ if len(userIDs) > 0 }} or t.user_id in (@userIDs) {{ end }}
		//   )
		// {{ end }}
		Count2(tables []string, appID int, keyWord string, status int, certificateIDs []int, userIDs []int) (int, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, count(*) `count`, date_format(t.created_time, '%Y%m') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CountWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		CostWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m%d') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select t.type `type`, date_format(t.created_time, '%Y%m') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`, `type`
		// order by `day`
		PassRateWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)
	}
	generator.ApplyInterface(func(AndroidSigningJobQuery) {}, generator.GenerateModel("t_android_signing_job"))

	type AppleSigningJobQuery interface {
		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db
		// and TABLE_NAME like 't_apple_signing_job%'
		GetTables(db string) ([]string, error)

		// select * from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID
		// order by t.created_time desc, t.id desc limit @limit offset @offset
		List(tables []string, appID int, limit, offset int) ([]*gen.T, error)

		// select count(*) from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.app_id = @appID
		Count2(tables []string, appID int) (int, error)

		// select count(*) `count`, date_format(t.created_time, '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CountWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select count(*) `count`, date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CountWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select count(*) `count`, date_format(t.created_time, '%Y%m') `day` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CountWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(t.created_time, '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CostWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CostWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(t.created_time, '%Y%m') `day`, cast(round(avg(timestampdiff(SECOND, t.created_time, ifnull(t.finished_time, now()))), 0) as signed) `cost` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		CostWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(t.created_time, '%Y%m%d') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		PassRateWithDay(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(date_sub(t.created_time, INTERVAL (dayofweek(t.created_time)-2) DAY), '%Y%m%d') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		PassRateWithWeek(tables []string, appID int, begin, end time.Time) ([]gen.M, error)

		// select date_format(t.created_time, '%Y%m') `day`, cast(round(sum(case t.status when 2 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from (
		// {{ for i, t := range tables }}
		//     select * from @@t where created_time between @begin and @end and status in (2, 3)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		//     {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// group by `day`
		// order by `day`
		PassRateWithMonth(tables []string, appID int, begin, end time.Time) ([]gen.M, error)
	}
	generator.ApplyInterface(func(AppleSigningJobQuery) {}, generator.GenerateModel("t_apple_signing_job"))

	type WhqlJobQuery interface {
		// select type `type`, count(*) `count`, date_format(created_time, '%Y%m%d') `day` from t_whql_job
		// where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CountWithDay(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, count(*) `count`, date_format(date_sub(created_time, INTERVAL (dayofweek(created_time)-2) DAY), '%Y%m%d') `day` from t_whql_job
		// where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CountWithWeek(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, count(*) `count`, date_format(created_time, '%Y%m') `day` from t_whql_job
		// where created_time between @begin and @end
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CountWithMonth(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(created_time, '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, created_time, ifnull(finished_time, now()))), 0) as signed) `cost` from t_whql_job
		// where created_time between @begin and @end and status in (9, 10)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CostWithDay(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(date_sub(created_time, INTERVAL (dayofweek(created_time)-2) DAY), '%Y%m%d') `day`, cast(round(avg(timestampdiff(SECOND, created_time, ifnull(finished_time, now()))), 0) as signed) `cost` from t_whql_job
		// where created_time between @begin and @end and status in (9, 10)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CostWithWeek(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(created_time, '%Y%m') `day`, cast(round(avg(timestampdiff(SECOND, created_time, ifnull(finished_time, now()))), 0) as signed) `cost` from t_whql_job
		// where created_time between @begin and @end and status in (9, 10)
		//     {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		CostWithMonth(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(created_time, '%Y%m%d') `day`, cast(round(sum(case status when 10 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from t_whql_job
		//    where created_time between @begin and @end and status in (9, 10)
		//    {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		PassRateWithDay(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(date_sub(created_time, INTERVAL (dayofweek(created_time)-2) DAY), '%Y%m%d') `day`, cast(round(sum(case status when 10 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from t_whql_job
		//    where created_time between @begin and @end and status in (9, 10)
		//    {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		PassRateWithWeek(appID int, begin, end time.Time) ([]gen.M, error)

		// select type `type`, date_format(created_time, '%Y%m') `day`, cast(round(sum(case status when 10 then 1 else 0 end) * 10000 / count(*), 0) as signed) `rate` from t_whql_job
		//    where created_time between @begin and @end and status in (9, 10)
		//    {{ if appID > 0 }} and app_id = @appID {{ end }}
		// group by `day`, `type`
		// order by `day`
		PassRateWithMonth(appID int, begin, end time.Time) ([]gen.M, error)
	}
	generator.ApplyInterface(func(WhqlJobQuery) {}, generator.GenerateModel("t_whql_job"))

	generator.ApplyBasic(generator.GenerateAllTable()...)
	generator.Execute()
}
