package eywa

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTable struct {
	Name string `graphql:"name"`
	Age  int    `graphql:"age"`
}

func (t testTable) ModelName() string {
	return "test_table"
}

func TestQuery(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	_, err := Query(&s).Select("name").Exec(c)

	assert.NoError(t, err)
}

func TestQueryLimit(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Query(&s).Select("name").Limit(1).Exec(c)

	if assert.NoError(t, err) {
		assert.Len(t, resp, 1)
	}
}

func TestQueryDistinctOn(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Query(&s).Select("name", "age").DistinctOn("age").Exec(c)
	expectedResp := []*testTable{
		{"efgh", 10}, {"abcd", 12},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestQueryOrderBy(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Query(&s).Select("name", "age").OrderBy(map[string]string{"age": OrderAsc}).Exec(c)
	expectedResp := []*testTable{
		{"efgh", 10}, {"abc", 10}, {"abcd", 12},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestQueryWhere(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	//	resp, err := Query(&s).Select("name", "age").Where(`{_or: [{name: {_eq: "abc"}}, {age: {_eq: 12}}]}`).Exec(c)
	resp, err := Query(&s).Select("name", "age").Where(
		&WhereExpr{
			Or: []*WhereExpr{
				{
					Comparisons: Comparison{
						"name": {
							Eq: "abc",
						},
					},
				},
				{
					Comparisons: Comparison{
						"age": {
							Eq: 12,
						},
					},
				},
			},
		},
	).Exec(c)
	expectedResp := []*testTable{
		{"abcd", 12}, {"abc", 10},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestQueryByPk(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := QueryByPk(&s).Pk(map[string]interface{}{"name": "abcd"}).Select("name", "age").Exec(c)
	expectedResp := &testTable{"abcd", 12}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}
