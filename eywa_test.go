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
