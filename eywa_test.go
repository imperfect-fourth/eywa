package eywa

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTable struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	ID   *int   `json:"id,omitempty"`
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

	resp, err := Query(&s).Select("name", "age").DistinctOn("age").Limit(5).Exec(c)
	if assert.NoError(t, err) {
		repeatCheck := make(map[int]bool)
		for _, r := range resp {
			if assert.False(t, repeatCheck[r.Age]) {
				repeatCheck[r.Age] = true
				continue
			}
			return
		}
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

	resp, err := Query(&s).Select("name", "age").OrderBy(map[string]string{"age": OrderAsc}).Limit(5).Exec(c)
	if assert.NoError(t, err) {
		asc := true
		prev := 0
		for _, r := range resp {
			if r.Age < prev {
				asc = false
			}
			prev = r.Age
		}

		assert.True(t, asc)
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
		{Name: "abcd", Age: 12}, {Name: "abc", Age: 10},
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

	resp, err := QueryByPk(&s).Pk(map[string]interface{}{"id": 1}).Select("name", "age").Exec(c)
	expectedResp := &testTable{Name: "abcd", Age: 12}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestInsertOne(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := InsertOne(&testTable{Name: "test", Age: 1}).Select("name", "age").Exec(c)
	expectedResp := &testTable{Name: "test", Age: 1}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestInsert(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Insert(&testTable{Name: "test", Age: 1}, &testTable{Name: "test2", Age: 2}).Select("name", "age").Exec(c)
	expectedResp := []*testTable{
		&testTable{Name: "test", Age: 1},
		&testTable{Name: "test2", Age: 2},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expectedResp, resp)
	}
}

func TestDelete(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	_, err := Insert(
		&testTable{Name: "testdelete", Age: 3},
		&testTable{Name: "testdelete", Age: 4},
		&testTable{Name: "testdelete", Age: 5},
	).Select("name").Exec(c)
	if !assert.NoError(t, err) {
		return
	}

	resp, err := Delete[testTable]().Where(&WhereExpr{
		Comparisons: Comparison{
			"name": {
				Eq: "testdelete",
			},
		},
	}).Select("name").Exec(c)
	if assert.NoError(t, err) {
		for _, r := range resp {
			assert.Equal(t, "testdelete", r.Name)
		}
	}
}
