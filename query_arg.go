package eywa

import (
	"fmt"
	"reflect"
	"strings"
)

type queryArgs[M Model] struct {
	limit      *limit
	offset     *offset
	distinctOn *distinctOn[M]
	where      *where
	orderBy    *orderBy
	set        *set[M]
	object     *object[M]
	onConflict *onConflict[M]
}

func (qa queryArgs[M]) MarshalGQL() string {
	var args []string
	args = appendArg(args, qa.limit)
	args = appendArg(args, qa.offset)
	args = appendArg(args, qa.distinctOn)
	args = appendArg(args, qa.where)
	args = appendArg(args, qa.orderBy)
	args = appendArg(args, qa.set)
	args = appendArg(args, qa.object)
	args = appendArg(args, qa.onConflict)

	if len(args) == 0 {
		return ""
	}
	return fmt.Sprintf("(%s)", strings.Join(args, ", "))
}

func appendArg(arr []string, arg queryArg) []string {
	if arg == nil || reflect.ValueOf(arg).IsNil() {
		return arr
	}
	s := arg.MarshalGQL()
	if s == "" {
		return arr
	}
	return append(arr, s)
}

type queryArg interface {
	queryArgName() string
	MarshalGQL() string
}

type limit int

func (l limit) queryArgName() string {
	return "limit"
}
func (l limit) MarshalGQL() string {
	return fmt.Sprintf("%s: %d", l.queryArgName(), l)
}

type offset int

func (o offset) queryArgName() string {
	return "offset"
}
func (o offset) MarshalGQL() string {
	return fmt.Sprintf("%s: %d", o.queryArgName(), o)
}

type distinctOn[M Model] struct {
	field FieldName[M]
}

func (do distinctOn[M]) queryArgName() string {
	return "distinct_on"
}
func (do distinctOn[M]) MarshalGQL() string {
	return fmt.Sprintf("%s: %s", do.queryArgName(), do.field)
}

type where struct {
	*WhereExpr
}

func (w where) queryArgName() string {
	return "where"
}
func (w where) MarshalGQL() string {
	return fmt.Sprintf("%s: %s", w.queryArgName(), w.WhereExpr.MarshalGQL())
}

type set[M Model] struct {
	FieldArray[M]
}

func (s set[M]) queryArgName() string {
	return "_set"
}
func (s set[M]) MarshalGQL() string {
	if len(s.FieldArray) == 0 {
		return ""
	}
	return fmt.Sprintf("%s: {%s}", s.queryArgName(), s.FieldArray.MarshalGQL())
}

type operator string

const (
	eq  operator = "_eq"
	neq operator = "_neq"
	gt  operator = "_gt"
	gte operator = "_gte"
	lt  operator = "_lt"
	lte operator = "_lte"
)

func compare[M Model](oprtr operator, field Field[M]) *WhereExpr {
	return &WhereExpr{
		cmp: fmt.Sprintf("%s: {%s: %s}", field.GetName(), oprtr, field.GetValue()),
	}
}

func Eq[M Model](field Field[M]) *WhereExpr {
	return compare[M](eq, field)
}

func Neq[M Model](field Field[M]) *WhereExpr {
	return compare[M](neq, field)
}

func Gt[M Model](field Field[M]) *WhereExpr {
	return compare[M](gt, field)
}

func Gte[M Model](field Field[M]) *WhereExpr {
	return compare[M](gte, field)
}

func Lt[M Model](field Field[M]) *WhereExpr {
	return compare[M](lt, field)
}

func Lte[M Model](field Field[M]) *WhereExpr {
	return compare[M](lte, field)
}

func Not(w *WhereExpr) *WhereExpr {
	return &WhereExpr{
		not: w,
	}
}

func And(w ...*WhereExpr) *WhereExpr {
	return &WhereExpr{
		and: whereArr(w),
	}
}

func Or(w ...*WhereExpr) *WhereExpr {
	return &WhereExpr{
		or: whereArr(w),
	}
}

type WhereExpr struct {
	and whereArr
	or  whereArr
	not *WhereExpr
	cmp string
}

type whereArr []*WhereExpr

func (wa whereArr) MarshalGQL() string {
	stringArr := make([]string, 0, len(wa))
	for _, whereExpr := range wa {
		expr := whereExpr.MarshalGQL()
		if expr != "" {
			stringArr = append(stringArr, expr)
		}
	}
	return strings.Join(stringArr, ", ")
}

func (w *WhereExpr) MarshalGQL() string {
	if w == nil {
		return ""
	}
	if (w == &WhereExpr{}) {
		return "{}"
	}
	var stringArr []string

	andExpr := w.and.MarshalGQL()
	if andExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_and: [%s]", andExpr))
	}

	orExpr := w.or.MarshalGQL()
	if orExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_or: [%s]", orExpr))
	}

	notExpr := w.not.MarshalGQL()
	if notExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_not: %s", notExpr))
	}

	if w.cmp != "" {
		stringArr = append(stringArr, w.cmp)
	}
	expr := fmt.Sprintf("{%s}", strings.Join(stringArr, ", "))
	return expr
}

type OrderByExpr struct {
	order string
	field string
}

func (ob OrderByExpr) MarshalGQL() string {
	return fmt.Sprintf("%s: %s", ob.field, ob.order)
}

func Asc[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"asc", string(field)}
}
func AscNullsFirst[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"asc_nulls_first", string(field)}
}
func AscNullsLast[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"asc_nulls_last", string(field)}
}
func Desc[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"desc", string(field)}
}
func DescNullsFirst[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"desc_nulls_first", string(field)}
}
func DescNullsLast[M Model](field FieldName[M]) OrderByExpr {
	return OrderByExpr{"desc_nulls_last", string(field)}
}

type orderBy []OrderByExpr

func (oba orderBy) queryArgName() string {
	return "order_by"
}

func (oba orderBy) MarshalGQL() string {
	if len(oba) == 0 {
		return ""
	}
	stringArr := make([]string, 0, len(oba))
	for _, ob := range oba {
		expr := ob.MarshalGQL()
		if expr != "" {
			stringArr = append(stringArr, expr)
		}
	}
	return fmt.Sprintf("%s: {%s}", oba.queryArgName(), strings.Join(stringArr, ", "))
}

type object[M Model] struct {
	fields FieldArray[M]
}

func (o object[M]) queryArgName() string {
	return "object"
}

func (o object[M]) MarshalGQL() string {
	return fmt.Sprintf("%s: {%s}", o.queryArgName(), o.fields.MarshalGQL())
}

type onConflict[M Model] struct {
	constraint    Constraint[M]
	updateColumns FieldNameArray[M]
}

func (oc onConflict[M]) queryArgName() string {
	return "on_conflict"
}

func (oc onConflict[M]) MarshalGQL() string {
	if oc.updateColumns == nil {
		return fmt.Sprintf("%s: {constraint: %s}", oc.queryArgName(), string(oc.constraint))
	}
	return fmt.Sprintf("%s: {constraint: %s, update_columns: [%s]}", oc.queryArgName(), string(oc.constraint), oc.updateColumns.MarshalGQL())
}
