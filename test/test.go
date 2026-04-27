package eywatest

import "github.com/imperfect-fourth/eywa"

type User struct {
	eywa.SelectionSet

	ID       int     `json:"id"`
	Username string  `json:"username"`
	Address  *string `json:"address"`

	SubField []UserSubField `json:"sub_field"`
}

type UserSubField struct {
	eywa.SelectionSet

	ID         int    `json:"id"`
	SomeValues []bool `json:"some_values"`
}

type NotUser struct {
	eywa.SelectionSet

	ID int `json:"id"`
}

type orderBy string

const (
	Asc  orderBy = "asc"
	Desc orderBy = "desc"
)

type (
	EnumArg  eywa.Enum
	EnumArg1 eywa.Enum
)

const (
	enumValue  EnumArg  = "enum_value"
	enumValue1 EnumArg1 = "enum_value_1"
)

type arguments struct {
	eywa.ArgumentSet

	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	OrderBy orderBy `json:"order_by"`
	EnumArg EnumArg `json:"enum_arg"`
}

type QueryRoot struct {
	eywa.QueryRoot

	Users    []User    `json:"users"`
	NotUsers []NotUser `json:"not_users"`
}
type MutationRoot struct {
	eywa.MutationRoot
}
