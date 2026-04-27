package eywa

type GQLMarshaler interface {
	MarshalGQL() []byte
}
