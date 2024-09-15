package unsafe

import (
	"os"
	"testing"

	"github.com/imperfect-fourth/eywa"
	"github.com/stretchr/testify/assert"
)

type testTable struct {
	Name     string          `json:"name"`
	Age      int             `json:"age"`
	ID       *int            `json:"id,omitempty"`
	State    eywa.HasuraEnum `json:"state"`
	JsonBCol jsonbcol        `json:"jsonb_col"`
}

func (t testTable) ModelName() string {
	return "test_table"
}

type jsonbcol struct {
	StrField  string `json:"str_field"`
	IntField  int    `json:"int_field"`
	BoolField bool   `json:"bool_field"`
	ArrField  []int  `json:"arr_field,omitempty"`
}

func TestSelectQuery(t *testing.T) {
	q := Get[testTable]().Limit(2).Offset(1).DistinctOn("name").OrderBy(eywa.Desc[testTable]("name")).Where(
		eywa.Or(
			eywa.Eq[testTable](eywa.RawField{"name", "abcd"}),
			eywa.Eq[testTable](eywa.RawField{"age", 10}),
		),
	).Select("name")

	expected := `query get_test_table {
test_table(limit: 2, offset: 1, distinct_on: name, where: {_or: [{name: {_eq: "abcd"}}, {age: {_eq: 10}}]}, order_by: {name: desc}) {
name
}
}`
	if assert.Equal(t, expected, q.Query()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql", &eywa.ClientOpts{
			Headers: map[string]string{
				"x-hasura-access-key": accessKey,
			},
		})

		resp, err := q.Exec(c)

		assert.NoError(t, err)
		assert.Equal(t, []testTable{{Name: "abcd"}, {Name: "abc"}}, resp)
	}
}

func TestUpdateQuery(t *testing.T) {
	q := Update[testTable]().Where(
		eywa.Eq[testTable](eywa.RawField{"id", 3}),
	).Set(
		eywa.RawField{"name", "updatetest"},
		eywa.RawField{"state", eywa.HasuraEnum("state1")},
		eywa.RawField{"jsonb_col", jsonbcol{
			StrField:  "abcd",
			IntField:  2,
			BoolField: false,
			ArrField:  []int{1, 2, 3},
		},
		}).Select("name", "id")

	expected := `mutation update_test_table {
update_test_table(where: {id: {_eq: 3}}, _set: {name: "updatetest", state: state1, jsonb_col: "{\"str_field\":\"abcd\",\"int_field\":2,\"bool_field\":false,\"arr_field\":[1,2,3]}"}) {
returning {
id
name
}
}
}`
	if assert.Equal(t, expected, q.Query()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql", &eywa.ClientOpts{
			Headers: map[string]string{
				"x-hasura-access-key": accessKey,
			},
		})

		resp, err := q.Exec(c)

		assert.NoError(t, err)
		n := 3
		assert.Equal(t, []testTable{{ID: &n, Name: "updatetest"}}, resp)
	}
}
