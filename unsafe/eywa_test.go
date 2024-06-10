package unsafe

import (
	"os"
	"testing"

	"github.com/imperfect-fourth/eywa"
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

func TestSelectQuery(t *testing.T) {
	q := Get[testTable]().Limit(2).Offset(1).DistinctOn("age").Where(
		eywa.Or(
			eywa.Eq[testTable](eywa.RawField{"name", "abc"}),
			eywa.Eq[testTable](eywa.RawField{"age", 12}),
		),
	).Select("name")

	expected := `query get_test_table {
test_table(limit: 2, offset: 1, distinct_on: age, where: {_or: [{name: {_eq: "abc"}}, {age: {_eq: 12}}]}) {
name
}
}`
	if assert.Equal(t, expected, q.Query()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
			map[string]string{
				"x-hasura-access-key": accessKey,
			},
		)

		resp, err := q.Exec(c)

		assert.NoError(t, err)
		assert.Equal(t, []testTable{{Name: "a"}}, resp)
	}
}

func TestUpdateQuery(t *testing.T) {
	q := Update[testTable]().Where(
		eywa.Eq[testTable](eywa.RawField{"id", 3}),
	).Set(eywa.RawField{"name", "updatetest"}).Select("name", "id")

	expected := `mutation update_test_table {
update_test_table(where: {id: {_eq: 3}}, _set: {name: "updatetest"}) {
returning {
id
name
}
}
}`
	if assert.Equal(t, expected, q.Query()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
			map[string]string{
				"x-hasura-access-key": accessKey,
			},
		)

		resp, err := q.Exec(c)

		assert.NoError(t, err)
		n := 3
		assert.Equal(t, []testTable{{ID: &n, Name: "updatetest"}}, resp)
	}
}
