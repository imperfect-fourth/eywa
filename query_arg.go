package eywa

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type queryArgs[M Model, FN FieldName[M]] struct {
	limit      *limit
	offset     *offset
	distinctOn *distinctOn[M, FN]
	where      *where
	set        *set[M]
}

func (qa queryArgs[M, FN]) marshalGQL() string {
	var args []string
	args = appendArg(args, qa.limit)
	args = appendArg(args, qa.offset)
	args = appendArg(args, qa.distinctOn)
	args = appendArg(args, qa.where)
	args = appendArg(args, qa.set)

	return fmt.Sprintf("(%s)", strings.Join(args, ", "))
}

func appendArg(arr []string, arg queryArg) []string {
	if arg == nil || reflect.ValueOf(arg).IsNil() {
		return arr
	}
	s := arg.marshalGQL()
	if s == "" {
		return arr
	}
	return append(arr, s)
}

type queryArg interface {
	queryArgName() string
	marshalGQL() string
}

type limit int

func (l limit) queryArgName() string {
	return "limit"
}
func (l limit) marshalGQL() string {
	return fmt.Sprintf("%s: %d", l.queryArgName(), l)
}

type offset int

func (o offset) queryArgName() string {
	return "offset"
}
func (o offset) marshalGQL() string {
	return fmt.Sprintf("%s: %d", o.queryArgName(), o)
}

type distinctOn[M Model, FN FieldName[M]] struct {
	field FN
}

func (do distinctOn[M, FN]) queryArgName() string {
	return "distinct_on"
}
func (do distinctOn[M, FN]) marshalGQL() string {
	return fmt.Sprintf("%s: %s", do.queryArgName(), do.field)
}

type where struct {
	*WhereExpr
}

func (w where) queryArgName() string {
	return "where"
}
func (w where) marshalGQL() string {
	return fmt.Sprintf("%s: %s", w.queryArgName(), w.WhereExpr.marshalGQL())
}

type set[M Model] struct {
	fieldArr[M]
}

func (s set[M]) queryArgName() string {
	return "_set"
}
func (s set[M]) marshalGQL() string {
	if s.fieldArr == nil || len(s.fieldArr) == 0 {
		return ""
	}
	return fmt.Sprintf("%s: {%s}", s.queryArgName(), s.fieldArr.marshalGQL())
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
	val, _ := json.Marshal(field.Value)
	return &WhereExpr{
		cmp: fmt.Sprintf("%s: {%s: %s}", field.Name, oprtr, string(val)),
	}
}

func Eq[M Model](field Field[M]) *WhereExpr {
	return compare(eq, field)
}

func Neq[M Model](field Field[M]) *WhereExpr {
	return compare(neq, field)
}

func Gt[M Model](field Field[M]) *WhereExpr {
	return compare(gt, field)
}

func Gte[M Model](field Field[M]) *WhereExpr {
	return compare(gte, field)
}

func Lt[M Model](field Field[M]) *WhereExpr {
	return compare(lt, field)
}

func Lte[M Model](field Field[M]) *WhereExpr {
	return compare(lte, field)
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

func (wa whereArr) marshalGQL() string {
	stringArr := make([]string, 0, len(wa))
	for _, whereExpr := range wa {
		expr := whereExpr.marshalGQL()
		if expr != "" {
			stringArr = append(stringArr, expr)
		}
	}
	return strings.Join(stringArr, ", ")
}

func (w *WhereExpr) marshalGQL() string {
	if w == nil {
		return ""
	}
	if (w == &WhereExpr{}) {
		return "{}"
	}
	var stringArr []string

	andExpr := w.and.marshalGQL()
	if andExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_and: [%s]", andExpr))
	}

	orExpr := w.or.marshalGQL()
	if orExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_or: [%s]", orExpr))
	}

	notExpr := w.not.marshalGQL()
	if notExpr != "" {
		stringArr = append(stringArr, fmt.Sprintf("_not: %s", notExpr))
	}

	if w.cmp != "" {
		stringArr = append(stringArr, w.cmp)
	}
	expr := fmt.Sprintf("{%s}", strings.Join(stringArr, ", "))
	return expr
}
