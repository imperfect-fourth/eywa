package eywa

type GQLMarshaler interface {
	MarshalGQL() string
}

type Enum[T ~string] string

func (e Enum[T]) MarshalGQL() string {
	return string(e)
}
