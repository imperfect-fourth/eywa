package eywatest

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/imperfect-fourth/eywa"
	"github.com/stretchr/testify/assert"
)

func TestSelectQuery(t *testing.T) {
	age := 10
	q := eywa.Get[testTable]().Limit(2).Offset(1).DistinctOn(testTable_Name).OrderBy(
		eywa.Desc[testTable](testTable_Name),
	).Where(
		eywa.Or(
			eywa.Eq[testTable](testTable_NameField("abcd")),
			eywa.Eq[testTable](testTable_AgeField(&age)),
		),
	).Select(testTable_Name)

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

func TestRelationshipSelectQuery(t *testing.T) {
	age := 10
	q := eywa.Get[testTable]().Limit(2).Offset(1).DistinctOn(testTable_Name).OrderBy(
		eywa.Desc[testTable](testTable_Name),
	).Where(
		eywa.Or(
			eywa.Eq[testTable](testTable_NameField("abcd")),
			eywa.Eq[testTable](testTable_AgeField(&age)),
		),
	).Select(
		testTable_Name,
		testTable_testTable2(
			testTable2_ID,
		),
	)

	expected := `query get_test_table {
test_table(limit: 2, offset: 1, distinct_on: name, where: {_or: [{name: {_eq: "abcd"}}, {age: {_eq: 10}}]}, order_by: {name: desc}) {
test_table2 {
id
}
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
	q := eywa.Update[testTable]().Where(
		eywa.Eq[testTable](testTable_IDField(3)),
	).Set(
		testTable_NameField("updatetest"),
		testTable_JsonBColVar[eywa.JSONBValue](jsonbcol{
			StrField:  "abcd",
			IntField:  2,
			BoolField: false,
			ArrField:  []int{1, 2, 3},
		}),
	).Select(
		testTable_Name,
		testTable_ID,
	)

	expected := `mutation update_test_table($testTable_JsonBCol: jsonb) {
update_test_table(where: {id: {_eq: 3}}, _set: {name: "updatetest", jsonb_col: $testTable_JsonBCol}) {
returning {
id
name
}
}
}`
	expectedVars := map[string]interface{}{
		"testTable_JsonBCol": jsonbcol{
			StrField:  "abcd",
			IntField:  2,
			BoolField: false,
			ArrField:  []int{1, 2, 3},
		},
	}
	if assert.Equal(t, expected, q.Query()) && assert.Equal(t, expectedVars, q.Variables()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql", &eywa.ClientOpts{
			Headers: map[string]string{
				"x-hasura-access-key": accessKey,
			},
		})

		resp, err := q.Exec(c)

		assert.NoError(t, err)
		n := 3
		assert.Equal(t, []testTable{{ID: n, Name: "updatetest"}}, resp)
	}
}

func TestInsertOneQuery(t *testing.T) {
	id := uuid.New()
	q := eywa.InsertOne(
		testTable2_IDField(id),
	).Select(
		testTable2_ID,
	)
	expected := fmt.Sprintf(`mutation insert_test_table2_one {
insert_test_table2_one(object: {id: "%s"}) {
id
}
}`, id.String())

	if assert.Equal(t, expected, q.Query()) {
		accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
		c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql", &eywa.ClientOpts{
			Headers: map[string]string{
				"x-hasura-access-key": accessKey,
			},
		})

		resp, err := q.Exec(c)
		assert.NoError(t, err)
		assert.Equal(t, &testTable2{ID: id}, resp)
	}
}

func TestInsertOneQueryOnConflict(t *testing.T) {
	id := uuid.New()
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	c := eywa.NewClient("https://aware-cowbird-80.hasura.app/v1/graphql", &eywa.ClientOpts{
		Headers: map[string]string{
			"x-hasura-access-key": accessKey,
		},
	})
	q := eywa.InsertOne(
		testTable2_IDField(id),
		testTable2_AgeField(10),
	).Select(
		testTable2_ID,
	)
	q.Exec(c)

	q2 := eywa.InsertOne(
		testTable2_IDField(id),
		testTable2_AgeField(20),
	).OnConflict(
		testTable2_PkeyConstraint,
		testTable2_Age,
	).Select(
		testTable2_ID,
		testTable2_Age,
	)
	expected := fmt.Sprintf(`mutation insert_test_table2_one {
insert_test_table2_one(object: {age: 20, id: "%s"}, on_conflict: {constraint: testTable2_pkey, update_columns: [age]}) {
age
id
}
}`, id.String())

	if assert.Equal(t, expected, q2.Query()) {
		resp, err := q2.Exec(c)
		assert.NoError(t, err)
		assert.Equal(t, &testTable2{ID: id, Age: 20}, resp)
	}
}
