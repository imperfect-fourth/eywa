package eywa

import (
	"encoding/json"
	"fmt"
	"strings"
)

type insert[T Model] struct {
	operationName   string
	returningFields []string
	affectedRows    bool
	insertOne       bool

	objects []T
}

type insertResponse[T Model] struct {
	AffectedRows int `json:"affected_rows"`
	Returning    []T `json:"returning"`
}

func Insert[T Model](object T, objects ...T) *insert[T] {
	return &insert[T]{
		operationName: "insert",
		objects:       append(objects, object),
		insertOne:     false,
	}
}

func InsertOne[T Model](object T) *insert[T] {
	return &insert[T]{
		operationName: "insertOne",
		objects:       append(objects, object),
		insertOne:     true,
	}
}

func (iq *insert[T]) Select(fields ...string) *insert[T] {
	iq.returningFields = fields
	return iq
}

func (iq *insert[T]) AffectedRows() *insert[T] {
	iq.affectedRows = true
	return iq
}

func (iq *insert[T]) build() (string, error) {
	var query string
	if iq.insertOne {
		obj, err := json.Marshal(iq.objects[0])
		if err != nil {
			return "", err
		}
		query = fmt.Sprintf(
			"mutation %s {insert_%s_one(object: %s) {%s}}",
			iq.operationName,
			iq.objects[0].ModelName(),
			string(obj),
			strings.Join(iq.returningFields, "\n"),
		)
	} else {
		insertQueryFormat := "mutation %s {insert_%s(objects: %s){%s}}"
		objs, err := json.Marshal(iq.objects)
		returnBlock := ""
		if len(iq.returningFields) != 0 {
			returnBlock = fmt.Sprintf("returning{%s}", returnBlock, strings.Join(iq.returningFields, "\n"))
		}
		if iq.affectedRows || returnBlock == "" {
			returnBlock = "affected_rows\n" + returnBlock
		}

		query = fmt.Sprintf(
			insertQueryFormat,
			iq.operationName,
			iq.objects[0].ModelName(),
			string(objs),
			returnBlock,
		)
	}
	return query, nil
}

func (iq *insert[T]) Exec(client *Client) ([]
