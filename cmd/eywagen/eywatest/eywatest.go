package eywatest

import "github.com/google/uuid"

//go:generate eywagen -types testTable,testTable2
type testTable struct {
	Name       string      `json:"name"`
	Age        *int        `json:"age"`
	ID         uuid.UUID   `json:"id,omitempty"`
	custom     *customType `json:"custom"`
	testTable2 *testTable2 `json:"testTable2"`
}

func (t testTable) ModelName() string {
	return "test_table"
}

type customType struct{}

type testTable2 struct {
	ID *uuid.UUID `json:"id"`
}

func (t testTable2) ModelName() string {
	return "test_table2"
}
