package unsafe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testAction struct {
	Name string `json:"name"`
	ID   *int   `json:"id,omitempty"`
}

func (a testAction) ModelName() string {
	return "test_action"
}

func TestActionQuery(t *testing.T) {
	const state2 = "state2"

	q := Action[testAction]().QueryArgs(
		map[string]interface{}{
			"arg1": 4,
			"arg2": "value",
			"arg3": state2,
		},
	).Select("name")

	// 	expected := `query run_test_action {
	// test_action(arg1: 4, arg2: "value", arg3: enumvalue) {
	// name
	// }
	// }`
	query := q.Query()
	assert.Contains(t, query,
		`query run_test_action {
test_action(`,
		`arg1: 4`,
		`arg2: "value"`,
		`arg3: state2`,
		`) {
name
}
}`,
	)
}
