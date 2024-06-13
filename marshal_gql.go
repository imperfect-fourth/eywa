package eywa

type gqlMarshaler interface {
	marshalGQL() string
}

type HasuraEnum string

func (he HasuraEnum) marshalGQL() string {
	_ = x(he)
	return string(he)
}

func x(q gqlMarshaler) string {
	return "abcd"
}
