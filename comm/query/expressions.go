package query

import (
	"strconv"
	"time"

	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

// CaseStringOrderBy 生成带 CASE 的 OrderBy 语句。
func CaseStringOrderBy(f field.String, v string) field.Expr {
	return field.NewUnsafeFieldRaw("CASE WHEN ? = ? THEN 1 ELSE 2 END", f.RawExpr(), v)
}

// FindInSetWithNumber 生成 SQL FIND_IN_SET 语句。
func FindInSetWithNumber(f field.Field, v int) field.Expr {
	return field.NewUnsafeFieldRaw("FIND_IN_SET(?, ?) > 0", strconv.Itoa(v), f.RawExpr())
}

// ForUpdate 生成 SQL FOR UPDATE 语句。
func ForUpdate() clause.Expression {
	return clause.Locking{Strength: "UPDATE"}
}

// Concat 生成 SQL CONCAT 语句。
func Concat(f field.String, value string) field.AssignExpr {
	return f.SetCol(f.IfNull("").ConcatCol(field.NewUnsafeFieldRaw("?", value)))
}

// TimeDiffGte 生成时间字段与当前时间差值比较语句。
func TimeDiffGte(f field.Time, duration time.Duration) field.AssignExpr {
	return field.NewUnsafeFieldRaw("TIMESTAMPDIFF(SECOND, ?, NOW()) >= ? ", f.RawExpr(), int(duration.Seconds()))
}
