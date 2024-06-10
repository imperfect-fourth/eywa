package eywa

import (
	"encoding/json"
	"fmt"
)

func Update[M Model, MP ModelPtr[M]]() UpdateQueryBuilder[M, string] {
	return UpdateQueryBuilder[M, string]{
		querySkeleton: querySkeleton[M, string]{
			modelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type UpdateQueryBuilder[M Model, FN FieldName[M]] struct {
	querySkeleton[M, FN]
}

func (uq UpdateQueryBuilder[M, FN]) Set(fields ...Field[M, FN]) UpdateQueryBuilder[M, FN] {
	uq.set = &set[M, FN]{fieldArr[M, FN](fields)}
	return uq
}

func (uq UpdateQueryBuilder[M, FN]) Where(w *WhereExpr) UpdateQueryBuilder[M, FN] {
	uq.where = &where{w}
	return uq
}

func (uq *UpdateQueryBuilder[M, FN]) marshalGQL() string {
	if uq.where == nil {
		uq.where = &where{Not(&WhereExpr{})}
	}
	return fmt.Sprintf(
		"update_%s",
		uq.querySkeleton.marshalGQL(),
	)
}

func (uq UpdateQueryBuilder[M, FN]) Select(field FN, fields ...FN) UpdateQuery[M, FN] {
	return UpdateQuery[M, FN]{
		uq:     &uq,
		fields: append(fields, field),
	}
}

type UpdateQuery[M Model, FN FieldName[M]] struct {
	uq     *UpdateQueryBuilder[M, FN]
	fields []FN
}

func (uq *UpdateQuery[M, FN]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\nreturning {\n%s\n}\n}",
		uq.uq.marshalGQL(),
		FieldNameArr[M, FN](uq.fields).marshalGQL(),
	)
}

func (uq *UpdateQuery[M, FN]) Query() string {
	return fmt.Sprintf(
		"mutation update_%s {\n%s\n}",
		uq.uq.modelName,
		uq.marshalGQL(),
	)
}

func (uq *UpdateQuery[M, FN]) Exec(client *Client) ([]M, error) {
	respBytes, err := client.do(uq.Query())
	if err != nil {
		return nil, err
	}

	type mutationReturning struct {
		Returning []M `json:"returning"`
	}
	type graphqlResponse struct {
		Data   map[string]mutationReturning `json:"data"`
		Errors []graphqlError               `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[fmt.Sprintf("update_%s", uq.uq.modelName)].Returning, nil
}
