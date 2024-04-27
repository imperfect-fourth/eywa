package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
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

type Model[T any] interface {
	*T
	ModelName() string
}

type querySkeleton[T any, M Model[T]] struct {
	operationName string
	selectFields  []string
}

func (q *querySkeleton[T, M]) setSelectFields(fields []string) {
	q.selectFields = fields
}

type query[T any, M Model[T]] struct {
	*querySkeleton[T, M]

	model         M
	queryModifier map[string]interface{}
}

func Query[T any, M Model[T]](model M) *query[T, M] {
	q := &query[T, M]{
		querySkeleton: &querySkeleton[T, M]{
			operationName: "Get",
		},
		model:         model,
		queryModifier: map[string]interface{}{},
	}
	return q
}

func (q *query[T, M]) Select(fields ...string) *query[T, M] {
	q.setSelectFields(fields)
	return q
}

func (q *query[T, M]) Limit(n int) *query[T, M] {
	q.queryModifier["limit"] = n
	return q
}

func (q *query[T, M]) DistinctOn(field string) *query[T, M] {
	q.queryModifier["distinct_on"] = field
	return q
}

const (
	OrderAsc            = "asc"
	OrderAscNullsFirst  = "asc_nulls_first"
	OrderAscNullsLast   = "asc_nulls_last"
	OrderDesc           = "desc"
	OrderDescNullsFirst = "desc_nulls_first"
	OrderDescNullsLast  = "desc_nulls_last"
)

func (q *query[T, M]) OrderBy(orderBys map[string]string) *query[T, M] {
	orderByClause := ""
	for k, v := range orderBys {
		if orderByClause == "" {
			orderByClause = fmt.Sprintf("{%s: %s", k, v)
		} else {
			orderByClause = fmt.Sprintf("%s, %s: %s", orderByClause, k, v)
		}
	}
	orderByClause += "}"
	q.queryModifier["order_by"] = orderByClause
	return q
}

func (q *query[T, M]) Where(where *WhereExpr) *query[T, M] {
	q.queryModifier["where"] = where.build()
	return q
}

func (q *query[T, M]) build() string {
	baseQueryFormat := "query %s {%s%s {%s}}"

	modifierClause := ""
	for k, v := range q.queryModifier {
		if modifierClause == "" {
			modifierClause = fmt.Sprintf("(%s: %v", k, v)
		} else {
			modifierClause = fmt.Sprintf("%s, %s: %v", modifierClause, k, v)
		}
	}
	if modifierClause != "" {
		modifierClause = modifierClause + ")"
	}

	gql := fmt.Sprintf(
		baseQueryFormat,
		q.operationName,
		q.model.ModelName(),
		modifierClause,
		strings.Join(q.selectFields, "\n"),
	)
	fmt.Println(gql)
	return gql
}

func (q *query[T, M]) Exec(client *Client) ([]M, error) {
	respBytes, err := client.do(q.build())

	type graphqlResponse struct {
		Data   map[string][]M `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[q.model.ModelName()], nil
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}

const (
	Eq  = "_eq"
	Neq = "_new"
	Gt  = "_gt"
	Gte = "_gte"
	Lt  = "_lt"
	Lte = "_lte"
)

type Comparison map[string]map[string]interface{}

type WhereExpr struct {
	And         whereExprArr
	Or          whereExprArr
	Not         *WhereExpr
	Comparisons Comparison
}

type whereExprArr []*WhereExpr

func (w whereExprArr) build() string {
	exprArr := make([]string, 0, len(w))
	for _, exprElem := range w {
		exprBuild := exprElem.build()
		if exprBuild != "" {
			exprArr = append(exprArr, exprBuild)
		}
	}
	return strings.Join(exprArr, ", ")
}

func (w *WhereExpr) build() string {
	if w == nil {
		return ""
	}
	if (w == &WhereExpr{}) {
		return "{}"
	}
	var exprArr []string

	andExpr := w.And.build()
	if andExpr != "" {
		exprArr = append(exprArr, fmt.Sprintf("_and: [%s]", andExpr))
	}

	orExpr := w.Or.build()
	if orExpr != "" {
		exprArr = append(exprArr, fmt.Sprintf("_or: [%s]", orExpr))
	}

	notExpr := w.Not.build()
	if notExpr != "" {
		exprArr = append(exprArr, fmt.Sprintf("_not: %s", notExpr))
	}

	for field, cmprs := range w.Comparisons {
		cmpExprArr := make([]string, 0, len(cmprs))
		for cmpr, val := range cmprs {
			fmt.Println(reflect.TypeOf(val))
			if val == nil {
				cmpExprArr = append(cmpExprArr, fmt.Sprintf("%s: null", cmpr))
			} else if _, ok := val.(string); ok {
				cmpExprArr = append(cmpExprArr, fmt.Sprintf(`%s: %q`, cmpr, val))
			} else if v, ok := val.(fmt.Stringer); ok {
				cmpExprArr = append(cmpExprArr, fmt.Sprintf(`%s: %q`, cmpr, v.String()))
			} else {
				cmpExprArr = append(cmpExprArr, fmt.Sprintf("%s: %v", cmpr, val))
			}
		}
		exprArr = append(exprArr, fmt.Sprintf("%s: {%s}", field, strings.Join(cmpExprArr, ", ")))
	}

	expr := fmt.Sprintf("{%s}", strings.Join(exprArr, ", "))
	return expr
}

type queryByPk[T any, M Model[T]] struct {
	*querySkeleton[T, M]

	model M
	pk    map[string]interface{}
}

func QueryByPk[T any, M Model[T]](model M) *queryByPk[T, M] {
	return &queryByPk[T, M]{
		querySkeleton: &querySkeleton[T, M]{
			operationName: "GetByPk",
		},
		model: model,
	}
}

func (q *queryByPk[T, M]) Select(selectFields ...string) *queryByPk[T, M] {
	q.selectFields = selectFields
	return q
}

func (q *queryByPk[T, M]) Pk(pk map[string]interface{}) *queryByPk[T, M] {
	q.pk = pk
	return q
}

func (q *queryByPk[T, M]) build() string {
	baseQueryFormat := "query %s {%s(%s) {%s}}"

	pk := make([]string, 0, len(q.pk))
	for k, v := range q.pk {
		if v == nil {
			pk = append(pk, fmt.Sprintf("%s: null", k))
		}
		pk = append(pk, fmt.Sprintf("%s: %v", k, v))
	}
	return fmt.Sprintf(
		baseQueryFormat,
		q.operationName,
		fmt.Sprintf("%s_by_pk", q.model.ModelName()),
		strings.Join(pk, ", "),
		strings.Join(q.selectFields, "\n"),
	)
}

func (q *queryByPk[T, M]) Exec(c *Client) (M, error) {
	respBytes, err := c.do(q.build())

	type graphqlResponse struct {
		Data   map[string]M   `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[q.model.ModelName()+"_by_pk"], nil
}
