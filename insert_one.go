package eywa

import (
	"encoding/json"
	"fmt"
)

func InsertOne[M Model, MP ModelPtr[M]](field ModelField[M], fields ...ModelField[M]) InsertOneQueryBuilder[M, ModelFieldName[M], ModelField[M]] {
	arr := fieldArr[M, ModelField[M]](fields)
	arr = append(arr, field)
	return InsertOneQueryBuilder[M, ModelFieldName[M], ModelField[M]]{
		QuerySkeleton: QuerySkeleton[M, ModelFieldName[M], ModelField[M]]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
			queryArgs: queryArgs[M, ModelFieldName[M], ModelField[M]]{
				object: &object[M, ModelField[M]]{arr},
			},
		},
	}
}

type InsertOneQueryBuilder[M Model, FN FieldName[M], F Field[M]] struct {
	QuerySkeleton[M, FN, F]
}

func (iq *InsertOneQueryBuilder[M, FN, F]) MarshalGQL() string {
	return fmt.Sprintf(
		"insert_%s_one%s",
		iq.QuerySkeleton.ModelName,
		iq.queryArgs.MarshalGQL(),
	)
}

func (iq InsertOneQueryBuilder[M, FN, F]) Select(field FN, fields ...FN) InsertOneQuery[M, FN, F] {
	return InsertOneQuery[M, FN, F]{
		iq:     &iq,
		fields: append(fields, field),
	}
}

type InsertOneQuery[M Model, FN FieldName[M], F Field[M]] struct {
	iq     *InsertOneQueryBuilder[M, FN, F]
	fields []FN
}

func (iq InsertOneQuery[M, FN, F]) MarshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		iq.iq.MarshalGQL(),
		FieldNameArr[M, FN](iq.fields).MarshalGQL(),
	)
}

func (iq InsertOneQuery[M, FN, F]) Query() string {
	return fmt.Sprintf(
		"mutation insert_%s_one%s {\n%s\n}",
		iq.iq.ModelName,
		iq.iq.queryVars.MarshalGQL(),
		iq.MarshalGQL(),
	)
}

func (iq InsertOneQuery[M, FN, F]) Variables() map[string]interface{} {
	vars := map[string]interface{}{}
	for _, var_ := range iq.iq.queryVars {
		vars[var_.name] = var_.value.Value()
	}
	return vars
}

func (iq InsertOneQuery[M, FN, F]) Exec(client *Client) (*M, error) {
	respBytes, err := client.Do(iq)
	if err != nil {
		return nil, err
	}

	type graphqlResponse struct {
		Data   map[string]*M  `json:"data"`
		Errors []GraphQLError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[fmt.Sprintf("insert_%s_one", iq.iq.ModelName)], nil
}
