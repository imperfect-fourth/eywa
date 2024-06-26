package eywatest

import "github.com/google/uuid"

//go:generate ../eywagen -types testTable,testTable2
type testTable struct {
	Name       string      `json:"name"`
	Age        *int        `json:"age"`
	ID         int         `json:"id,omitempty"`
	iD         int32       `json:"idd,omitempty"`
	custom     *customType `json:"custom"`
	testTable2 *testTable2 `json:"testTable2"`
	JsonBCol   jsonbcol    `json:"jsonb_col"`
	RR         R           `json:"r"`
}

type R string

func (t testTable) ModelName() string {
	return "test_table"
}

type customType struct{}

type testTable2 struct {
	ID uuid.UUID `json:"id"`
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
