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

type Field[M Model] string
type ModelField[M Model] interface {
	Field[M] | string
}

type ModelFieldArr[M Model, MF ModelField[M]] []MF

func (mfs ModelFieldArr[M, MF]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, mf := range mfs {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(string(mf))
	}
	return buf.String()
}

type ModelFieldMap[M Model, MF ModelField[M]] map[MF]interface{}

func (mfs ModelFieldMap[M, MF]) marshalGQL() string {
	if mfs == nil {
		return "{}"
	}
	buf := bytes.NewBuffer([]byte("{"))
	first := true
	for k, v := range mfs {
		if !first {
			buf.WriteString(", ")
		} else {
			first = false
		}
		buf.WriteString(string(k))
		buf.WriteString(": ")
		valJson, _ := json.Marshal(v)
		buf.Write(valJson)
	}
	buf.WriteString("}")
	return buf.String()
}

type Queryable interface {
	Query() string
}

type querySkeleton[M Model, MF ModelField[M]] struct {
	modelName string
	// fields    ModelFieldArr[M, MF]
	queryArgs
}

func (qs querySkeleton[M, MF]) marshalGQL() string {
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

type SelectQueryBuilder[M Model, MF ModelField[M]] struct {
	querySkeleton[M, MF]
}

func (sq SelectQueryBuilder[M, MF]) queryModelName() string {
	return sq.modelName
}

func (sq SelectQueryBuilder[M, MF]) DistinctOn(f string) SelectQueryBuilder[M, MF] {
	sq.distinctOn = (*distinctOn)(&f)
	return sq
}

func (sq SelectQueryBuilder[M, MF]) Offset(n int) SelectQueryBuilder[M, MF] {
	sq.offset = (*offset)(&n)
	return sq
}

func (sq SelectQueryBuilder[M, MF]) Limit(n int) SelectQueryBuilder[M, MF] {
	sq.limit = (*limit)(&n)
	return sq
}

func (sq SelectQueryBuilder[M, MF]) Where(w *WhereExpr) SelectQueryBuilder[M, MF] {
	sq.where = &where{w}
	return sq
}

func (sq SelectQueryBuilder[M, MF]) marshalGQL() string {
	return sq.querySkeleton.marshalGQL()
}

func (sq SelectQueryBuilder[M, MF]) Select(field MF, fields ...MF) SelectQuery[M, MF] {
	return SelectQuery[M, MF]{
		sq:     &sq,
		fields: append(fields, field),
	}
}

type SelectQuery[M Model, MF ModelField[M]] struct {
	sq     *SelectQueryBuilder[M, MF]
	fields []MF
}

func (sq SelectQuery[M, MF]) marshalGQL() string {
	return fmt.Sprintf(
		"%s {\n%s\n}",
		sq.sq.marshalGQL(),
		ModelFieldArr[M, MF](sq.fields).marshalGQL(),
	)
}

func (sq SelectQuery[M, MF]) Query() string {
	return fmt.Sprintf(
		"query get_%s {\n%s\n}",
		sq.sq.modelName,
		sq.marshalGQL(),
	)
}

func (sq SelectQuery[M, MF]) Exec(client *Client) ([]M, error) {
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
