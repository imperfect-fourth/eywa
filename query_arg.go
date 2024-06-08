package eywa

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

type operator string

const (
	eq  operator = "_eq"
	neq          = "_neq"
	gt           = "_gt"
	gte          = "_gte"
	lt           = "_lt"
	lte          = "_lte"
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

type where struct {
	*WhereExpr
}

func (w where) queryArgName() string {
	return "where"
}
func (w where) marshalGQL() string {
	return fmt.Sprintf("%s: %s", w.queryArgName(), w.WhereExpr.marshalGQL())
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
