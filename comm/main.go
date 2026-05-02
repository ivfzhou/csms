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
		OutPath:           "comm/query",
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
		// select * from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where 1 = 1
		// {{ if !begin.IsZero() }} and t.created_time >= @begin {{ end }}
		// {{ if !end.IsZero() }} and t.created_time <= @end {{ end }}
		// {{ if len(appIDs) > 0 }} and t.app_id in (@appIDs) {{ end }}
		// {{ if len(userIDs) > 0 }} and t.user_id in (@userIDs) {{ end }}
		// order by t.created_time desc, t.id desc limit @limit offset @offset
		List(tables []string, appIDs, userIDs []int, begin, end time.Time, limit, offset int) ([]*gen.T, error)

		// select count(*) from (
		// {{ for i, t := range tables }}
		//     select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where 1 = 1
		// {{ if !begin.IsZero() }} and t.created_time >= @begin {{ end }}
		// {{ if !end.IsZero() }} and t.created_time <= @end {{ end }}
		// {{ if len(appIDs) > 0 }} and t.app_id in (@appIDs) {{ end }}
		// {{ if len(userIDs) > 0 }} and t.user_id in (@userIDs) {{ end }}
		Count2(tables []string, appIDs, userIDs []int, begin, end time.Time) (int, error)

		// select t.type `type`, count(*) `count` from (
		// {{ for i, t := range tables }}
		//     select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.created_time >= @begin and t.created_time <= @end and t.type in (@types)
		// {{ if appID > 0 }} and t.app_id = @appID {{ end }}
		// group by t.type
		CountByTypes(tables []string, types []int, appID int, begin, end time.Time) ([]*gen.M, error)

		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db and TABLE_NAME like 't_event%'
		GetTables(db string) ([]string, error)
	}
	generator.ApplyInterface(func(EventQuery) {}, generator.GenerateModel("t_event"))

	type WindowsSigningJobQuery interface {
		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db
		// and TABLE_NAME like 't_windows_signing_job%'
		GetTables(db string) ([]string, error)

		// select t.job_id from (
		// {{ for i, t := range tables }}
		//    select * from @@t {{ if len(tables) - 1 != i }} union all {{ end }}
		// {{ end }} ) t
		// where t.status = @status
		GetJobIDByStatus(tables []string, status int) ([]string, error)

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
	}
	generator.ApplyInterface(func(WindowsSigningJobQuery) {}, generator.GenerateModel("t_windows_signing_job"))

	type AndroidSigningJobQuery interface {
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

		// select TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = @db
		// and TABLE_NAME like 't_android_signing_job%'
		GetTables(db string) ([]string, error)
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
	}
	generator.ApplyInterface(func(AppleSigningJobQuery) {}, generator.GenerateModel("t_apple_signing_job"))

	generator.ApplyBasic(generator.GenerateAllTable()...)
	generator.Execute()
}
