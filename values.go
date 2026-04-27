package eywa

type Enum string

func (e Enum) MarshalGQL() []byte {
	return []byte(e)
}
