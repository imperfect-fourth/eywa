package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLError struct {
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

type FieldName[M Model] string

type FieldNameArray[M Model] []FieldName[M]

func (fa FieldNameArray[M]) MarshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fa {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(string(f))
	}
	return buf.String()
}

type Field[M Model] struct {
	Name  string
	Value interface{}
}

func (f Field[M]) GetName() string {
	return f.Name
}

func (f Field[M]) GetValue() string {
	if var_, ok := f.Value.(queryVar); ok {
		return fmt.Sprintf("$%s", var_.name)
	}

	if val, ok := f.Value.(GQLMarshaler); ok {
		return val.MarshalGQL()
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

func (f Field[M]) GetRawValue() interface{} {
	return f.Value
}

type FieldArray[M Model] []Field[M]

func (fs FieldArray[M]) MarshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(f.GetName())
		buf.WriteString(": ")
		buf.WriteString(f.GetValue())
	}
	return buf.String()
}

type Queryable interface {
	Query() string
	Variables() map[string]interface{}
}

type QuerySkeleton[M Model] struct {
	ModelName string
	queryVars queryVarArr
	// fields    ModelFieldArr[M, FN, F]
	queryArgs[M]
}

func (qs QuerySkeleton[M]) MarshalGQL() string {
	return fmt.Sprintf("%s%s", qs.ModelName, qs.queryArgs.MarshalGQL())
}
