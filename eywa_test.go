package eywa

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTable struct {
	Name string `graphql:"name"`
}

func (t testTable) ModelName() string {
	return "test_table"
}

func TestSelect(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)
	resp, err := Select(c, &s)
	expectedResp := []*testTable{{Name: "abcdefgh"}}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestQuery(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Query(&s).Select([]string{"name"}).Exec(c)
	expectedResp := []*testTable{{Name: "abcdefgh"}}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}

func TestQueryWithOptions(t *testing.T) {
	accessKey := os.Getenv("TEST_HGE_ACCESS_KEY")
	s := testTable{}
	c := NewClient("https://aware-cowbird-80.hasura.app/v1/graphql",
		map[string]string{
			"x-hasura-access-key": accessKey,
		},
	)

	resp, err := Query(&s, SelectFields[*testTable]([]string{"name"})).Exec(c)
	expectedResp := []*testTable{{Name: "abcdefgh"}}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedResp, resp)
	}
}
