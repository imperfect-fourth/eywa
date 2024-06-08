package eywa

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Update[T any, M Model[T]](m M) *UpdateQueryBuilder[T, M, string] {
	return &UpdateQueryBuilder[T, M, string]{
		querySkeleton: querySkeleton[T, M, string]{
			modelName: m.ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type UpdateQueryBuilder[T any, M Model[T], MF ModelField[T, M]] struct {
	querySkeleton[T, M, MF]
	set ModelFieldMap[T, M, MF]
}

func (uq UpdateQueryBuilder[T, M, MF]) queryModelName() string {
	return uq.modelName
}

func (uq UpdateQueryBuilder[T, M, MF]) Set(set map[MF]interface{}) UpdateQueryBuilder[T, M, MF] {
	uq.set = set
	return uq
}

func (uq UpdateQueryBuilder[T, M, MF]) Where(w *WhereExpr) UpdateQueryBuilder[T, M, MF] {
	uq.where = w
	return uq
}

func (uq *UpdateQueryBuilder[T, M, MF]) marshalGQL() string {
	var modifiers []string
	if uq.where != nil {
		modifiers = append(modifiers, fmt.Sprintf("where: %s", uq.where.marshalGQL()))
	} else {
		modifiers = append(modifiers, "where: {_not: {}}")
	}

	if uq.set != nil {
		modifiers = append(modifiers, fmt.Sprintf("_set: %s", uq.set.marshalGQL()))
	}

	modifier := strings.Join(modifiers, ", ")
	if modifier != "" {
		modifier = fmt.Sprintf("(%s)", modifier)
	}
	return fmt.Sprintf(
		"update_%s%s",
		uq.queryModelName(),
		modifier,
	)
}

func (uq UpdateQueryBuilder[T, M, MF]) Select(field MF, fields ...MF) UpdateQuery[T, M, MF] {
	return UpdateQuery[T, M, MF]{
		uq:     &uq,
		fields: append(fields, field),
	}
}

type UpdateQuery[T any, M Model[T], MF ModelField[T, M]] struct {
	uq     *UpdateQueryBuilder[T, M, MF]
	fields []MF
}

func (uq *UpdateQuery[T, M, MF]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\nreturning {\n%s\n}\n}",
		uq.uq.marshalGQL(),
		ModelFieldArr[T, M, MF](uq.fields).marshalGQL(),
	)
}

func (uq *UpdateQuery[T, M, MF]) Query() string {
	return fmt.Sprintf(
		"mutation update_%s {\n%s\n}",
		uq.uq.modelName,
		uq.marshalGQL(),
	)
}

func (uq *UpdateQuery[T, M, MF]) Exec(client *Client) ([]T, error) {
	respBytes, err := client.do(uq.Query())
	if err != nil {
		return nil, err
	}

	type mutationReturning struct {
		Returning []T `json:"returning"`
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
	fmt.Printf("%v", respObj)
	return respObj.Data[fmt.Sprintf("update_%s", uq.uq.modelName)].Returning, nil
}
