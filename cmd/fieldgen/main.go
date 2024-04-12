package main

import (
	"fmt"
	"go/types"
	"os"
	re "regexp"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

var graphqlTagPattern = re.MustCompile(`graphql:"([^"]+)"`)

func main() {
	if len(os.Args) != 3 {
		panic("expected exactly one argument: <source type>")
	}
	typeName := os.Args[1]
	//	fmt.Println(typeName)
	//	graphqlName := os.Args[1]

	pkg, err := loadPackage()
	if err != nil {
		panic(err)
	}

	typeObj := pkg.Types.Scope().Lookup(typeName)
	if typeObj == nil {
		panic("tyoe not found in package")
	}
	//	fmt.Println(typeObj)
	typeStruct, ok := typeObj.Type().Underlying().(*types.Struct)
	//	fmt.Println(typeStruct)

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
		if _, ok := structField.Type().Underlying().(*types.Basic); ok {
			fmt.Printf(
				"%#v\n",
				jen.Func().Id(fmt.Sprintf("%s_%s", typeName, structField.Name())).Params().String().Block(
					jen.Return(jen.Lit(graphqlFieldName)),
				),
			)
		} else {
			fmt.Printf(
				"%#v\n",
				jen.Func().Id(fmt.Sprintf("%s_%s", typeName, structField.Name())).Params(
					jen.Id("fields").Id("func() string"),
				).String().Block(
					jen.Id("res").Op(":=").Index().String().Values(),
					jen.For(jen.List(jen.Id("_"), jen.Id("f")).Op(":=").Range().Id("fields")).Block(
						jen.Id("fields").Op("=").Append(jen.Id("fields"), jen.Id("f").Call()),
					),
					jen.Return(
						jen.Qual("fmt", "Sprintf").Call(
							jen.Lit(graphqlFieldName+"{\n%s}"),
							jen.Qual("strings", "Join").Call(
								jen.Id("fields"),
								jen.Lit("\n"),
							),
						),
					),
				),
			)
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
