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

type Field[M Model, FN FieldName[M]] struct {
	Name  FN
	Value interface{}
}
type fieldArr[M Model, FN FieldName[M]] []Field[M, FN]

func (fs fieldArr[M, FN]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(string(f.Name))
		buf.WriteString(": ")
		val, _ := json.Marshal(f.Value)
		buf.WriteString(string(val))
	}
	return buf.String()
}

func RawField[M Model](s string, v interface{}) Field[M, string] {
	return Field[M, string]{s, v}
}

type Queryable interface {
	Query() string
}

type querySkeleton[M Model, FN FieldName[M]] struct {
	modelName string
	// fields    ModelFieldArr[M, FN]
	queryArgs[M, FN]
}

func (qs querySkeleton[M, FN]) marshalGQL() string {
	return fmt.Sprintf("%s%s", qs.modelName, qs.queryArgs.marshalGQL())
}

func Select[M Model, MP ModelPtr[M]]() SelectQueryBuilder[M, string] {
	return SelectQueryBuilder[M, string]{
		querySkeleton: querySkeleton[M, string]{
			modelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

type SelectQueryBuilder[M Model, FN FieldName[M]] struct {
	querySkeleton[M, FN]
}

func (sq SelectQueryBuilder[M, FN]) DistinctOn(f string) SelectQueryBuilder[M, FN] {
	sq.distinctOn = (*distinctOn)(&f)
	return sq
}

func (sq SelectQueryBuilder[M, FN]) Offset(n int) SelectQueryBuilder[M, FN] {
	sq.offset = (*offset)(&n)
	return sq
}

func (sq SelectQueryBuilder[M, FN]) Limit(n int) SelectQueryBuilder[M, FN] {
	sq.limit = (*limit)(&n)
	return sq
}

func (sq SelectQueryBuilder[M, FN]) Where(w *WhereExpr) SelectQueryBuilder[M, FN] {
	sq.where = &where{w}
	return sq
}

func (sq SelectQueryBuilder[M, FN]) marshalGQL() string {
	return sq.querySkeleton.marshalGQL()
}

func (sq SelectQueryBuilder[M, FN]) Select(field FN, fields ...FN) SelectQuery[M, FN] {
	return SelectQuery[M, FN]{
		sq:     &sq,
		fields: append(fields, field),
	}
}

type SelectQuery[M Model, FN FieldName[M]] struct {
	sq     *SelectQueryBuilder[M, FN]
	fields []FN
}

func (sq SelectQuery[M, FN]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		sq.sq.marshalGQL(),
		FieldNameArr[M, FN](sq.fields).marshalGQL(),
	)
}

func (sq SelectQuery[M, FN]) Query() string {
	return fmt.Sprintf(
		"query get_%s {\n%s\n}",
		sq.sq.modelName,
		sq.marshalGQL(),
	)
}

func (sq SelectQuery[M, FN]) Exec(client *Client) ([]M, error) {
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
	return respObj.Data[sq.sq.modelName], nil
}
