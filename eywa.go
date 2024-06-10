package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
}

func NewClient(endpoint string, headers map[string]string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		headers:    headers,
	}
}

func (c *Client) do(q string) (*bytes.Buffer, error) {
	reqObj := graphqlRequest{
		Query: q,
	}

	var reqBytes bytes.Buffer
	err := json.NewEncoder(&reqBytes).Encode(&reqObj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint, &reqBytes)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBytes bytes.Buffer
	_, err = io.Copy(&respBytes, resp.Body)
	return &respBytes, err
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}

type Model interface {
	ModelName() string
}

type ModelPtr[T Model] interface {
	*T
	Model
}

type ModelFieldName[M Model] string
type FieldName[M Model] interface {
	string | ModelFieldName[M]
}
type FieldNameArr[M Model, FN FieldName[M]] []FN

func (fa FieldNameArr[M, FN]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fa {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(string(f))
	}
	return buf.String()
}

type RawField struct {
	Name  string
	Value interface{}
}

func (f RawField) GetName() string {
	return f.Name
}
func (f RawField) GetValue() string {
	if val, ok := f.Value.(gqlMarshaller); ok {
		return val.marshalGQL()
	}
	val, _ := json.Marshal(f.Value)
	return string(val)
}

type ModelField[M Model] struct {
	Name  string
	Value interface{}
}

func (f ModelField[M]) GetName() string {
	return f.Name
}
func (f ModelField[M]) GetValue() string {
	if val, ok := f.Value.(gqlMarshaller); ok {
		return val.marshalGQL()
	}
	val, _ := json.Marshal(f.Value)
	return string(val)
}

type Field[M Model] interface {
	RawField | ModelField[M]
	GetName() string
	GetValue() string
}

type fieldArr[M Model, F Field[M]] []F

func (fs fieldArr[M, MF]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(string(f.GetName()))
		buf.WriteString(": ")
		buf.WriteString(f.GetValue())
	}
	return buf.String()
}

//func RawField[M Model](s string, v interface{}) Field[M] {
//	return Field[M]{s, v}
//}

type Queryable interface {
	Query() string
}

type QuerySkeleton[M Model, FN FieldName[M], F Field[M]] struct {
	ModelName string
	// fields    ModelFieldArr[M, FN, F]
	queryArgs[M, FN, F]
}

func (qs QuerySkeleton[M, FN, F]) marshalGQL() string {
	return fmt.Sprintf("%s%s", qs.ModelName, qs.queryArgs.marshalGQL())
}

func Get[M Model, MP ModelPtr[M]]() GetQueryBuilder[M, ModelFieldName[M], ModelField[M]] {
	return GetQueryBuilder[M, ModelFieldName[M], ModelField[M]]{
		QuerySkeleton: QuerySkeleton[M, ModelFieldName[M], ModelField[M]]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type GetQueryBuilder[M Model, FN FieldName[M], F Field[M]] struct {
	QuerySkeleton[M, FN, F]
}

func (sq GetQueryBuilder[M, FN, F]) DistinctOn(f FN) GetQueryBuilder[M, FN, F] {
	sq.distinctOn = &distinctOn[M, FN]{f}
	return sq
}

func (sq GetQueryBuilder[M, FN, F]) Offset(n int) GetQueryBuilder[M, FN, F] {
	sq.offset = (*offset)(&n)
	return sq
}

func (sq GetQueryBuilder[M, FN, F]) Limit(n int) GetQueryBuilder[M, FN, F] {
	sq.limit = (*limit)(&n)
	return sq
}

func (sq GetQueryBuilder[M, FN, F]) Where(w *WhereExpr) GetQueryBuilder[M, FN, F] {
	sq.where = &where{w}
	return sq
}

func (sq GetQueryBuilder[M, FN, F]) marshalGQL() string {
	return sq.QuerySkeleton.marshalGQL()
}

func (sq GetQueryBuilder[M, FN, F]) Select(field FN, fields ...FN) GetQuery[M, FN, F] {
	return GetQuery[M, FN, F]{
		sq:     &sq,
		fields: append(fields, field),
	}
}

type GetQuery[M Model, FN FieldName[M], F Field[M]] struct {
	sq     *GetQueryBuilder[M, FN, F]
	fields []FN
}

func (sq GetQuery[M, FN, F]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		sq.sq.marshalGQL(),
		FieldNameArr[M, FN](sq.fields).marshalGQL(),
	)
}

func (sq GetQuery[M, FN, F]) Query() string {
	return fmt.Sprintf(
		"query get_%s {\n%s\n}",
		sq.sq.ModelName,
		sq.marshalGQL(),
	)
}

func (sq GetQuery[M, FN, F]) Exec(client *Client) ([]M, error) {
	respBytes, err := client.do(sq.Query())
	if err != nil {
		return nil, err
	}

	type graphqlResponse struct {
		Data   map[string][]M `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[sq.sq.ModelName], nil
}
