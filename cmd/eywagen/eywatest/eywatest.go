package eywatest

import "github.com/google/uuid"

//go:generate eywagen -types testTable
type testTable struct {
	Name   string      `json:"name"`
	Age    *int        `json:"age"`
	ID     uuid.UUID   `json:"id,omitempty"`
	custom *customType `json:"custom"`
}

func (t testTable) ModelName() string {
	return "test_table"
}

type customType struct{}
