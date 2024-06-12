package eywa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}

type Model interface {
	ModelName() string
}

type ModelPtr[T Model] interface {
	*T
	Model
}

type ModelFieldName[M Model] string
type FieldName[M Model] interface {
	string | ModelFieldName[M]
}
type FieldNameArr[M Model, FN FieldName[M]] []FN

func (fa FieldNameArr[M, FN]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fa {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(string(f))
	}
	return buf.String()
}

type field[M Model] struct {
	Name  string
	Value interface{}
}

func (f field[M]) marshalValueGQL() string {
	if val, ok := f.Value.(queryVar[M]); ok {
		return val.name
	}
	if val, ok := f.Value.(gqlMarshaler); ok {
		return val.marshalGQL()
	}

	val, _ := json.Marshal(f.Value)
	vt := reflect.TypeOf(f.Value)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	if vt.Kind() == reflect.Struct || vt.Kind() == reflect.Map {
		val, _ = json.Marshal(string(val))
	}
	return string(val)
}

func (f field[M]) marshalGQL() string {
	return fmt.Sprintf("%s: %s", f.Name, f.marshalValueGQL())
}

type fieldArr[M Model] []field[M]

func (fs fieldArr[M]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(f.marshalGQL())
	}
	return buf.String()
}

func Field[M Model](s string, v interface{}) field[M] {
	return field[M]{s, v}
}

type Queryable interface {
	Query() string
	Variables() map[string]interface{}
}

type QuerySkeleton[M Model, FN FieldName[M]] struct {
	ModelName string
	// fields    ModelFieldArr[M, FN]
	queryArgs[M, FN]
}

func (qs QuerySkeleton[M, FN]) marshalGQL() string {
	return fmt.Sprintf("%s%s", qs.ModelName, qs.queryArgs.marshalGQL())
}

func Get[M Model, MP ModelPtr[M]]() GetQueryBuilder[M, ModelFieldName[M]] {
	return GetQueryBuilder[M, ModelFieldName[M]]{
		QuerySkeleton: QuerySkeleton[M, ModelFieldName[M]]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type GetQueryBuilder[M Model, FN FieldName[M]] struct {
	QuerySkeleton[M, FN]
}

func (sq GetQueryBuilder[M, FN]) DistinctOn(f FN) GetQueryBuilder[M, FN] {
	sq.distinctOn = &distinctOn[M, FN]{f}
	return sq
}

func (sq GetQueryBuilder[M, FN]) Offset(n int) GetQueryBuilder[M, FN] {
	sq.offset = (*offset)(&n)
	return sq
}

func (sq GetQueryBuilder[M, FN]) Limit(n int) GetQueryBuilder[M, FN] {
	sq.limit = (*limit)(&n)
	return sq
}

func (sq GetQueryBuilder[M, FN]) Where(w *WhereExpr) GetQueryBuilder[M, FN] {
	sq.where = &where{w}
	return sq
}

func (sq GetQueryBuilder[M, FN]) marshalGQL() string {
	return sq.QuerySkeleton.marshalGQL()
}

func (sq GetQueryBuilder[M, FN]) Select(field FN, fields ...FN) GetQuery[M, FN] {
	return GetQuery[M, FN]{
		sq:     &sq,
		fields: append(fields, field),
	}
}

type GetQuery[M Model, FN FieldName[M]] struct {
	sq     *GetQueryBuilder[M, FN]
	fields []FN
}

func (sq GetQuery[M, FN]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		sq.sq.marshalGQL(),
		FieldNameArr[M, FN](sq.fields).marshalGQL(),
	)
}

func (sq GetQuery[M, FN]) Query() string {
	return fmt.Sprintf(
		"query get_%s {\n%s\n}",
		sq.sq.ModelName,
		sq.marshalGQL(),
	)
}

func (sq GetQuery[M, FN]) Variables() map[string]interface{} {
	return nil
}

func (sq GetQuery[M, FN]) Exec(client *Client) ([]M, error) {
	respBytes, err := client.do(sq)
	if err != nil {
		return nil, err
	}

	type graphqlResponse struct {
		Data   map[string][]M `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}
	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}

	if len(respObj.Errors) > 0 {
		gqlErrs := make([]error, 0, len(respObj.Errors))
		for _, e := range respObj.Errors {
			gqlErrs = append(gqlErrs, errors.New(e.Message))
		}
		return nil, errors.Join(gqlErrs...)
	}

	return respObj.Data[sq.sq.ModelName], nil
}
