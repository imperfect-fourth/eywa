package eywa

type GQLMarshaler interface {
	MarshalGQL() string
}

type HasuraEnum[T ~string] string

func (he HasuraEnum[T]) MarshalGQL() string {
	return string(he)
}
