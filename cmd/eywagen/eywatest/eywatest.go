package eywatest

import (
	"github.com/google/uuid"
	"github.com/imperfect-fourth/eywa"
)

//go:generate ../eywagen -types testTable,testTable2
type testTable struct {
	Name       string            `json:"name"`
	Age        *int              `json:"age"`
	ID         int               `json:"id,omitempty",eywa:"pkey"`
	IDd        int32             `json:"idd,omitempty"`
	custom     *customType       `json:"custom"`
	testTable2 *testTable2       `json:"test_table2"`
	JsonBCol   jsonbcol          `json:"jsonb_col"`
	RR         R                 `json:"r"`
	Status     eywa.Enum[status] `json:"status"`
	F          X[string, int]    `json:"f"`
}

type status string

type X[T ~string, U ~int] string

var (
	state1 eywa.Enum[status] = "state1"
)

type R string

func (t testTable) ModelName() string {
	return "test_table"
}

type customType struct{}

type testTable2 struct {
	ID  uuid.UUID `json:"id",eywa:"pkey"`
	Age int       `json:"age"`
}

func (t testTable2) ModelName() string {
	return "test_table2"
}

type jsonbcol struct {
	StrField  string `json:"str_field"`
	IntField  int    `json:"int_field"`
	BoolField bool   `json:"bool_field"`
	ArrField  []int  `json:"arr_field,omitempty"`
}
