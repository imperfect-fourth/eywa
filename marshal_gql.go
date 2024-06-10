package eywa

type gqlMarshaller interface {
	marshalGQL() string
}

type HasuraEnum string

func (he HasuraEnum) marshalGQL() string {
	_ = x(he)
	return string(he)
}

func x(q gqlMarshaller) string {
	return "abcd"
}
