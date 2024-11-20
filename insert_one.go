package eywa

import (
	"encoding/json"
	"fmt"
)

func InsertOne[M Model, MP ModelPtr[M]](field Field[M], fields ...Field[M]) InsertOneQueryBuilder[M] {
	arr := FieldArray[M](fields)
	arr = append(arr, field)
	return InsertOneQueryBuilder[M]{
		QuerySkeleton: QuerySkeleton[M]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
			queryArgs: queryArgs[M]{
				object: &object[M]{arr},
			},
		},
	}
}

type InsertOneQueryBuilder[M Model] struct {
	QuerySkeleton[M]
}

//func (iq InsertOneQueryBuilder[M]) OnConstraint(constraint eywa.Constraint[M], field FieldName[M], fields ...FieldName[M]) InsertOneQuery[M] {
//	return InsertOneQuery[M]{
//		iq:     &iq,
//		fields: append(fields, field),
//	}
//}

func (iq *InsertOneQueryBuilder[M]) MarshalGQL() string {
	return fmt.Sprintf(
		"insert_%s_one%s",
		iq.QuerySkeleton.ModelName,
		iq.queryArgs.MarshalGQL(),
	)
}

func (iq InsertOneQueryBuilder[M]) Select(field FieldName[M], fields ...FieldName[M]) InsertOneQuery[M] {
	return InsertOneQuery[M]{
		iq:     &iq,
		fields: append(fields, field),
	}
}

type InsertOneQuery[M Model] struct {
	iq     *InsertOneQueryBuilder[M]
	fields []FieldName[M]
}

func (iq InsertOneQuery[M]) MarshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		iq.iq.MarshalGQL(),
		FieldNameArray[M](iq.fields).MarshalGQL(),
	)
}

func (iq InsertOneQuery[M]) Query() string {
	return fmt.Sprintf(
		"mutation insert_%s_one%s {\n%s\n}",
		iq.iq.ModelName,
		iq.iq.queryVars.MarshalGQL(),
		iq.MarshalGQL(),
	)
}

func (iq InsertOneQuery[M]) Variables() map[string]interface{} {
	vars := map[string]interface{}{}
	for _, var_ := range iq.iq.queryVars {
		vars[var_.name] = var_.value.Value()
	}
	return vars
}

func (iq InsertOneQuery[M]) Exec(client *Client) (*M, error) {
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
