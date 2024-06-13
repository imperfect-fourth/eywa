package eywa

//type Type interface {
//	Type() string
//}
//
//type Boolean interface {
//	~bool
//}
//type NullableBoolean interface {
//	~*bool
//}
//type Int interface {
//	~int
//}
//type NullableInt interface {
//	~*int
//}
//type Float interface {
//	~float32 | ~float64
//}
//type NullableFloat interface {
//	~*float32 | ~*float64
//}
//type String interface {
//	~string
//}
//type NullableBoolean interface {
//	~*string
//}

type TypedValue interface {
	Type() string
	Value() interface{}
}

type scalarValue struct {
	name  string
	value interface{}
}

func (tv scalarValue) Type() string {
	return tv.name
}
func (tv scalarValue) Value() interface{} {
	return tv.value
}

func BooleanVar[T ~bool](val T) TypedValue {
	return scalarValue{"Boolean!", val}
}
func NullableBooleanVar[T ~*bool](val T) TypedValue {
	return scalarValue{"Boolean", val}
}
func IntVar[T ~int | ~int8 | ~int16 | ~int32 | ~int64](val T) TypedValue {
	return scalarValue{"Int!", val}
}
func NullableIntVar[T ~*int | ~*int8 | ~*int16 | ~*int32 | ~*int64](val T) TypedValue {
	return scalarValue{"Int", val}
}
func FloatVar[T ~float64 | ~float32](val T) TypedValue {
	return scalarValue{"Float!", val}
}
func NullableFloat[T ~*float64 | ~*float32](val T) TypedValue {
	return scalarValue{"Float", val}
}
func StringVar[T ~string](val T) TypedValue {
	return scalarValue{"String!", val}
}
func NullableStringVar[T ~*string](val T) TypedValue {
	return scalarValue{"String", val}
}
func JSONVar(val interface{}) TypedValue {
	return JSONValue{val}
}
func JSONBVar(val interface{}) TypedValue {
	return JSONBValue{val}
}

type JSONValue struct {
	Val interface{}
}

func (jv JSONValue) Type() string {
	return "json"
}
func (jv JSONValue) Value() interface{} {
	return jv.Val
}

type JSONBValue struct {
	Val interface{}
}

func (jv JSONBValue) Type() string {
	return "jsonb"
}
func (jv JSONBValue) Value() interface{} {
	return jv.Val
}
