package eywa

import (
	"encoding/json"
	"errors"
	"fmt"
)

func Get[M Model, MP ModelPtr[M]]() GetQueryBuilder[M] {
	return GetQueryBuilder[M]{
		QuerySkeleton: QuerySkeleton[M]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type GetQueryBuilder[M Model] struct {
	QuerySkeleton[M]
}

func (sq GetQueryBuilder[M]) DistinctOn(f FieldName[M]) GetQueryBuilder[M] {
	sq.distinctOn = &distinctOn[M]{f}
	return sq
}

func (sq GetQueryBuilder[M]) Offset(n int) GetQueryBuilder[M] {
	sq.offset = (*offset)(&n)
	return sq
}

func (sq GetQueryBuilder[M]) Limit(n int) GetQueryBuilder[M] {
	sq.limit = (*limit)(&n)
	return sq
}

func (sq GetQueryBuilder[M]) OrderBy(o ...OrderByExpr) GetQueryBuilder[M] {
	orderByArr := orderBy(o)
	sq.orderBy = &orderByArr
	return sq
}

func (sq GetQueryBuilder[M]) Where(w *WhereExpr) GetQueryBuilder[M] {
	sq.where = &where{w}
	return sq
}

func (sq GetQueryBuilder[M]) MarshalGQL() string {
	return sq.QuerySkeleton.MarshalGQL()
}

func (sq GetQueryBuilder[M]) Select(field FieldName[M], fields ...FieldName[M]) GetQuery[M] {
	return GetQuery[M]{
		sq:     &sq,
		fields: append(fields, field),
	}
}

type GetQuery[M Model] struct {
	sq     *GetQueryBuilder[M]
	fields []FieldName[M]
}

func (sq GetQuery[M]) MarshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		sq.sq.MarshalGQL(),
		FieldNameArray[M](sq.fields).MarshalGQL(),
	)
}

func (sq GetQuery[M]) Query() string {
	return fmt.Sprintf(
		"query get_%s {\n%s\n}",
		sq.sq.ModelName,
		sq.MarshalGQL(),
	)
}

func (sq GetQuery[M]) Variables() map[string]interface{} {
	return nil
}

func (sq GetQuery[M]) Exec(client *Client) ([]M, error) {
	respBytes, err := client.Do(sq)
	if err != nil {
		return nil, err
	}

	type graphqlResponse struct {
		Data   map[string][]M `json:"data"`
		Errors []GraphQLError `json:"errors"`
	}

	respObj := graphqlResponse{}
	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}

	if len(respObj.Errors) > 0 {
		gqlErrs := make([]error, 0, len(respObj.Errors))
		for _, e := range respObj.Errors {
			gqlErrs = append(gqlErrs, errors.New(e.Message))
		}
		return nil, errors.Join(gqlErrs...)
	}

	return respObj.Data[sq.sq.ModelName], nil
}
