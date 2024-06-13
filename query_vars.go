package eywa

import (
	"bytes"
	"fmt"
)

type queryVar struct {
	name  string
	value TypedValue
}

func (v queryVar) marshalGQL() string {
	return fmt.Sprintf("$%s: %s", v.name, v.value.Type())
}

type queryVarArr []queryVar

func (vs queryVarArr) marshalGQL() string {
	buf := bytes.NewBufferString("(")
	for i, v := range vs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v.marshalGQL())
	}
	buf.WriteString(")")
	return buf.String()

}

func QueryVar(name string, value TypedValue) queryVar {
	return queryVar{name, value}
}
