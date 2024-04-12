package main

import (
	"fmt"
	"go/types"
	"os"
	re "regexp"

	"golang.org/x/tools/go/packages"
)

var graphqlTagPattern = re.MustCompile(`graphql:"([^"]+)"`)

const basicTypeFieldFunc = `func %s() string {
	return "%s"
}
`

const structTypeFieldFunc = `func %s(subFieldFunc func() string, subFieldFuncs ...func() string) func() string {
	subFields := []string{subFieldFunc()}
	for _, f := range subFieldsFuncs {
		subFields = append(subFields, f())
	}
	return func() string {return "%s{\n"+strings.Join(subFields, "\n")+"}"}
}
`

func main() {
	if len(os.Args) != 2 {
		panic("expected exactly one argument: <source type>")
	}
	typeName := os.Args[1]

	pkg, err := loadPackage()
	if err != nil {
		panic(err)
	}

	typeObj := pkg.Types.Scope().Lookup(typeName)
	if typeObj == nil {
		panic("type not found in package")
	}
	typeStruct, ok := typeObj.Type().Underlying().(*types.Struct)

	if !ok {
		panic("type %v is not a struct")
	}

	for i := 0; i < typeStruct.NumFields(); i++ {
		graphqlTag := graphqlTagPattern.FindStringSubmatch(typeStruct.Tag(i))
		if graphqlTag == nil {
			continue
		}
		graphqlFieldName := graphqlTag[1]
		structField := typeStruct.Field(i)
		structFieldType := structField.Type()

		if ptr, ok := structFieldType.(*types.Pointer); ok {
			structFieldType = ptr.Elem()
		}

		switch structFieldType.(type) {
		case *types.Basic, *types.Map:
			fmt.Printf(
				basicTypeFieldFunc,
				fmt.Sprintf("%s_%s", typeName, structField.Name()),
				graphqlFieldName,
			)
		case *types.Struct, *types.Slice:
			fmt.Printf(
				structTypeFieldFunc,
				fmt.Sprintf("%s_%s", typeName, structField.Name()),
				graphqlFieldName,
			)
		default:
		}
	}
}

func loadPackage() (*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo, Tests: true}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("couldn't load package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package contains errors")
	}
	return pkgs[0], nil
}
