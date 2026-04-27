package eywa

type Field[P SelectionSet] struct {
	name         string
	arguments    []byte
	selectionSet []byte
}

func NewField[P SelectionSet](name string) *Field[P] {
	return &Field[P]{
		name: name,
	}
}

func NewObjectField[P SelectionSet, O SelectionSet](name string) func(fields ...*Field[O]) *Field[P] {
	return func(fields ...*Field[O]) *Field[P] {
		return &Field[P]{
			name:         name,
			selectionSet: marshalGQLFields(fields),
		}
	}
}

func marshalGQLFields[P SelectionSet](fields []*Field[P]) []byte {
	ss := []byte("{\n")
	for _, f := range fields {
		ss = append(ss, f.MarshalGQL()...)
		ss = append(ss, '\n')
	}
	return append(ss, '}')
}

func (f Field[P]) MarshalGQL() []byte {
	b := make([]byte, 0, len(f.name)+len(f.arguments)+len(f.selectionSet)+1)
	b = append(b, f.name...)
	b = append(b, f.arguments...)
	return append(b, f.selectionSet...)
}

func WithArgs[P SelectionSet, A ArgumentSet](f *Field[P], args ...ArgumentOf[A]) *Field[P] {
	if len(args) == 0 {
		return f
	}
	argsBytes := []byte("(")
	for i, arg := range args {
		argsBytes = append(argsBytes, arg.MarshalGQL()...)
		if i != len(args)-1 {
			argsBytes = append(argsBytes, ", "...)
		}
	}
	f.setArguments(append(argsBytes, ')'))
	return f
}

func (f *Field[P]) setArguments(argsBytes []byte) {
	f.arguments = argsBytes
}
