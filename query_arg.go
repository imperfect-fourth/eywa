package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type queryArgs struct {
	limit      *limit
	offset     *offset
	distinctOn *distinctOn
	where      *where
	set        set
}

func (qa queryArgs) marshalGQL() string {
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

type distinctOn string

func (do distinctOn) queryArgName() string {
	return "distinct_on"
}
func (do distinctOn) marshalGQL() string {
	return fmt.Sprintf("%s: %s", do.queryArgName(), do)
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

type set map[string]interface{}

func (s set) queryArgName() string {
	return "_set"
}
func (s set) marshalGQL() string {
	if s == nil {
		return ""
	}
	buf := bytes.NewBuffer([]byte{})
	first := true
	for k, v := range s {
		if !first {
			buf.WriteString(", ")
		} else {
			first = false
		}
		val, _ := json.Marshal(v)
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.Write(val)
	}
	return fmt.Sprintf("%s: {%s}", s.queryArgName(), buf.String())
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

func compare(oprtr operator, field string, value interface{}) *WhereExpr {
	val, _ := json.Marshal(value)
	return &WhereExpr{
		cmp: fmt.Sprintf("%s: {%s: %s}", field, oprtr, string(val)),
	}
}

func Eq(field string, value interface{}) *WhereExpr {
	return compare(eq, field, value)
}

func Neq(field string, value interface{}) *WhereExpr {
	return compare(neq, field, value)
}

func Gt(field string, value interface{}) *WhereExpr {
	return compare(gt, field, value)
}

func Gte(field string, value interface{}) *WhereExpr {
	return compare(gte, field, value)
}

func Lt(field string, value interface{}) *WhereExpr {
	return compare(lt, field, value)
}

func Lte(field string, value interface{}) *WhereExpr {
	return compare(lte, field, value)
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
