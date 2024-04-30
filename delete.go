package eywa

import (
	"encoding/json"
	"fmt"
	"strings"
)

type deleteMutation[T any, M Model[T]] struct {
	*querySkeleton[T, M]
	where string
	model M
}

func Delete[T any, M Model[T]]() *deleteMutation[T, M] {
	var x T
	return &deleteMutation[T, M]{
		querySkeleton: &querySkeleton[T, M]{
			operationName: "delete",
		},
		model: &x,
	}
}

func (d *deleteMutation[T, M]) Select(fields ...string) *deleteMutation[T, M] {
	d.setSelectFields(fields)
	return d
}

func (d *deleteMutation[T, M]) Where(where *WhereExpr) *deleteMutation[T, M] {
	d.where = where.build()
	return d
}

const deleteMutationFormat string = "mutation %s {delete_%s(where: %s) {%s}}"

func (d *deleteMutation[T, M]) build() (string, error) {
	returnBlock := fmt.Sprintf("returning {%s}", strings.Join(d.selectFields, "\n"))
	mutation := fmt.Sprintf(
		deleteMutationFormat,
		d.operationName,
		d.model.ModelName(),
		d.where,
		returnBlock,
	)
	return mutation, nil
}

type DeleteResponse[T any, M Model[T]] struct {
	//	AffectedRows *int `json:"affected_rows"`
	Returning []M
}

func (d *deleteMutation[T, M]) Exec(client *Client) ([]M, error) {
	query, err := d.build()
	if err != nil {
		return nil, fmt.Errorf("couldn't build query: %w", err)
	}
	respBytes, err := client.do(query)
	if err != nil {
		return nil, fmt.Errorf("client query failed: %w", err)
	}
	type gqlResp struct {
		Data   map[string]*DeleteResponse[T, M] `json:"data"`
		Errors []graphqlError                   `json:"errors"`
	}
	var respObj gqlResp
	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[fmt.Sprintf("delete_%s", d.model.ModelName())].Returning, nil
}
