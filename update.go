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

type UpdateQueryBuilder[M Model, MF ModelField[M]] struct {
	querySkeleton[M, MF]
}

func (uq UpdateQueryBuilder[M, MF]) queryModelName() string {
	return uq.modelName
}

func (uq UpdateQueryBuilder[M, MF]) Set(s map[string]interface{}) UpdateQueryBuilder[M, MF] {
	uq.set = set(s)
	return uq
}

func (uq UpdateQueryBuilder[M, MF]) Where(w *WhereExpr) UpdateQueryBuilder[M, MF] {
	uq.where = &where{w}
	return uq
}

func (uq *UpdateQueryBuilder[M, MF]) marshalGQL() string {
	if uq.where == nil {
		uq.where = &where{Not(&WhereExpr{})}
	}
	return fmt.Sprintf(
		"update_%s",
		uq.querySkeleton.marshalGQL(),
	)
}

func (uq UpdateQueryBuilder[M, MF]) Select(field MF, fields ...MF) UpdateQuery[M, MF] {
	return UpdateQuery[M, MF]{
		uq:     &uq,
		fields: append(fields, field),
	}
}

type UpdateQuery[M Model, MF ModelField[M]] struct {
	uq     *UpdateQueryBuilder[M, MF]
	fields []MF
}

func (uq *UpdateQuery[M, MF]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\nreturning {\n%s\n}\n}",
		uq.uq.marshalGQL(),
		ModelFieldArr[M, MF](uq.fields).marshalGQL(),
	)
}

func (uq *UpdateQuery[M, MF]) Query() string {
	return fmt.Sprintf(
		"mutation update_%s {\n%s\n}",
		uq.uq.modelName,
		uq.marshalGQL(),
	)
}

func (uq *UpdateQuery[M, MF]) Exec(client *Client) ([]M, error) {
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
