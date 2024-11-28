package eywa

import (
	"context"
	"encoding/json"
	"fmt"
)

func Update[M Model, MP ModelPtr[M]]() UpdateQueryBuilder[M] {
	return UpdateQueryBuilder[M]{
		QuerySkeleton: QuerySkeleton[M]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type UpdateQueryBuilder[M Model] struct {
	QuerySkeleton[M]
}

func (uq UpdateQueryBuilder[M]) Set(fields ...Field[M]) UpdateQueryBuilder[M] {
	uq.set = &set[M]{FieldArray[M](fields)}
	for _, f := range fields {
		if var_, ok := f.GetRawValue().(queryVar); ok {
			uq.queryVars = append(uq.queryVars, var_)
		}
	}
	return uq
}

func (uq UpdateQueryBuilder[M]) Where(w *WhereExpr) UpdateQueryBuilder[M] {
	uq.where = &where{w}
	return uq
}

func (uq *UpdateQueryBuilder[M]) MarshalGQL() string {
	if uq.where == nil {
		uq.where = &where{Not(&WhereExpr{})}
	}
	return fmt.Sprintf(
		"update_%s",
		uq.QuerySkeleton.MarshalGQL(),
	)
}

func (uq UpdateQueryBuilder[M]) Select(field FieldName[M], fields ...FieldName[M]) UpdateQuery[M] {
	return UpdateQuery[M]{
		uq:     &uq,
		fields: append(fields, field),
	}
}

type UpdateQuery[M Model] struct {
	uq     *UpdateQueryBuilder[M]
	fields []FieldName[M]
}

func (uq UpdateQuery[M]) MarshalGQL() string {
	return fmt.Sprintf(
		"%s {\nreturning {\n%s\n}\n}",
		uq.uq.MarshalGQL(),
		FieldNameArray[M](uq.fields).MarshalGQL(),
	)
}

func (uq UpdateQuery[M]) Query() string {
	return fmt.Sprintf(
		"mutation update_%s%s {\n%s\n}",
		uq.uq.ModelName,
		uq.uq.queryVars.MarshalGQL(),
		uq.MarshalGQL(),
	)
}

func (uq UpdateQuery[M]) Variables() map[string]interface{} {
	vars := map[string]interface{}{}
	for _, var_ := range uq.uq.queryVars {
		vars[var_.name] = var_.value.Value()
	}
	return vars
}

func (uq UpdateQuery[M]) Exec(client *Client) ([]M, error) {
	return uq.ExecWithContext(context.Background(), client)
}

func (uq UpdateQuery[M]) ExecWithContext(ctx context.Context, client *Client) ([]M, error) {
	respBytes, err := client.Do(ctx, uq)
	if err != nil {
		return nil, err
	}

	type mutationReturning struct {
		Returning []M `json:"returning"`
	}
	type graphqlResponse struct {
		Data   map[string]mutationReturning `json:"data"`
		Errors []GraphQLError               `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[fmt.Sprintf("update_%s", uq.uq.ModelName)].Returning, nil
}
