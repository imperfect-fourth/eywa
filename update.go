package eywa

import (
	"encoding/json"
	"fmt"
)

func update[M Model, FN FieldName[M], F Field[M]]() UpdateQueryBuilder[M, FN, F] {
	return UpdateQueryBuilder[M, FN, F]{
		querySkeleton: querySkeleton[M, FN, F]{
			modelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

func UpdateUnsafe[M Model, MP ModelPtr[M]]() UpdateQueryBuilder[M, string, RawField] {
	return update[M, string, RawField]()
}
func Update[M Model, MP ModelPtr[M]]() UpdateQueryBuilder[M, ModelFieldName[M], ModelField[M]] {
	return update[M, ModelFieldName[M], ModelField[M]]()
}

type UpdateQueryBuilder[M Model, FN FieldName[M], F Field[M]] struct {
	querySkeleton[M, FN, F]
}

func (uq UpdateQueryBuilder[M, FN, F]) Set(fields ...F) UpdateQueryBuilder[M, FN, F] {
	uq.set = &set[M, F]{fieldArr[M, F](fields)}
	return uq
}

func (uq UpdateQueryBuilder[M, FN, F]) Where(w *WhereExpr) UpdateQueryBuilder[M, FN, F] {
	uq.where = &where{w}
	return uq
}

func (uq *UpdateQueryBuilder[M, FN, F]) marshalGQL() string {
	if uq.where == nil {
		uq.where = &where{Not(&WhereExpr{})}
	}
	return fmt.Sprintf(
		"update_%s",
		uq.querySkeleton.marshalGQL(),
	)
}

func (uq UpdateQueryBuilder[M, FN, F]) Select(field FN, fields ...FN) UpdateQuery[M, FN, F] {
	return UpdateQuery[M, FN, F]{
		uq:     &uq,
		fields: append(fields, field),
	}
}

type UpdateQuery[M Model, FN FieldName[M], F Field[M]] struct {
	uq     *UpdateQueryBuilder[M, FN, F]
	fields []FN
}

func (uq *UpdateQuery[M, FN, F]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\nreturning {\n%s\n}\n}",
		uq.uq.marshalGQL(),
		FieldNameArr[M, FN](uq.fields).marshalGQL(),
	)
}

func (uq *UpdateQuery[M, FN, F]) Query() string {
	return fmt.Sprintf(
		"mutation update_%s {\n%s\n}",
		uq.uq.modelName,
		uq.marshalGQL(),
	)
}

func (uq *UpdateQuery[M, FN, F]) Exec(client *Client) ([]M, error) {
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
