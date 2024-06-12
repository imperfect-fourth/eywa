package eywa

import (
	"bytes"
	"fmt"
)

type queryVar[M Model] struct {
	name    string
	gqlType string
	value   interface{}
}

func (v queryVar[M]) marshalGQL() string {
	return fmt.Sprint("%s: %s", v.name, v.gqlType)
}

type queryVarArr[M Model] []queryVar[M]

func (vs queryVarArr[M]) marshalGQL() string {
	buf := bytes.NewBufferString("")
	for i, v := range vs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v.marshalGQL())
	}
	return buf.String()

}
