package eywa

import "fmt"

type Argument[A ArgumentSet, T any] struct {
	name string
	val  T
}

func NewArgument[A ArgumentSet, T any](name string) func(val T) Argument[A, T] {
	return func(val T) Argument[A, T] {
		return Argument[A, T]{
			name, val,
		}
	}
}

func (a Argument[A, T]) MarshalGQL() []byte {
	if vm, ok := any(a.val).(GQLMarshaler); ok {
		return fmt.Appendf([]byte{}, "%s: %v", a.name, vm.MarshalGQL())
	}
	return fmt.Appendf([]byte{}, "%s: %v", a.name, a.val)
}

func (a Argument[A, T]) ArgumentOf() *A { return nil }

type ArgumentSet interface {
	argumentSet()
}

type ArgumentOf[A ArgumentSet] interface {
	GQLMarshaler

	ArgumentOf() *A
}
