package eywatest

import (
	"time"

	"github.com/google/uuid"
	"github.com/imperfect-fourth/eywa"
)

//go:generate ../eywagen -types testTable,testTable2
type testTable struct {
	Name       string                   `json:"name"`
	Age        *int                     `json:"age"`
	ID         int                      `json:"id,omitempty",eywa:"pkey"`
	IDd        int32                    `json:"idd,omitempty"`
	custom     *customType              `json:"custom"`
	customArr  []*customType            `json:"customarr"`
	testTable2 *testTable2              `json:"testTable2"`
	JsonBCol   jsonbcol                 `json:"jsonb_col"`
	Status     eywa.Enum[status]        `json:"status"`
	Generic    GenericType[string, int] `json:"generic_type"`
	ArrayCol   []string                 `json:"testarr"`
	timestamp  time.Time                `json:"timestamp"`
}

type status string

type GenericType[T ~string, U ~int] string

var (
	state1 eywa.Enum[status] = "state1"
)

func (t testTable) TableName() string {
	return "test_table"
}

func (t testTable) ModelName() string {
	return "test_table"
}

type customType struct{}

type testTable2 struct {
	ID  uuid.UUID `json:"id",eywa:"pkey"`
	Age int       `json:"age"`
}

func (t testTable2) TableName() string {
	return "testTable2"
}
func (t testTable2) ModelName() string {
	return "testTable2"
}

type jsonbcol struct {
	StrField  string `json:"str_field"`
	IntField  int    `json:"int_field"`
	BoolField bool   `json:"bool_field"`
	ArrField  []int  `json:"arr_field,omitempty"`
}
