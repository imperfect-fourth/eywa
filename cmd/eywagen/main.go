package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"os"
	re "regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeNames  = flag.String("types", "", "comma-separated list of type names; must be set")
	outputFile = flag.String("output-file", "eywa_generated.go", "output file path for generated file.")
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage:")
	fmt.Fprint(os.Stderr, "\teywagen -types <comma separated list of type names>")
}

var tagPattern = re.MustCompile(`json:"([^"]+)"`)
var eywaTagPattern = re.MustCompile(`eywa:"([^"]+)"`)

const (
	genHeader           = "// generated by eywa. DO NOT EDIT. Any changes will be overwritten.\npackage "
	modelFieldNameConst = "const %s eywa.FieldName[%s] = \"%s\"\n"
	modelFieldFunc      = `
func %sField(val %s) eywa.Field[%s] {
	return eywa.Field[%s]{
		Name: "%s",
		Value: val,
	}
}
`
	modelScalarVarFunc = `
func %sVar(val %s) eywa.Field[%s] {
	return eywa.Field[%s]{
		Name: "%s",
		Value: eywa.QueryVar("%s", %s[%s](val)),
	}
}
`
	modelVarFunc = `
func %sVar[T interface{%s;eywa.TypedValue}](val %s) eywa.Field[%s] {
	return eywa.Field[%s]{
		Name: "%s",
		Value: eywa.QueryVar("%s", T{val}),
	}
}
`

	modelRelationshipNameFunc = `
func %s(subField eywa.FieldName[%s], subFields ...eywa.FieldName[%s]) eywa.FieldName[%s] {
	buf := bytes.NewBuffer([]byte("%s {\n"))
	buf.WriteString(string(subField))
	for _, f := range subFields {
		buf.WriteString("\n")
		buf.WriteString(string(f))
	}
	buf.WriteString("\n}")
	return eywa.FieldName[%s](buf.String())
}
`
)

func pkeyConstraint(typeName string) string {
	return fmt.Sprintf("var %s_PkeyConstraint = eywa.Constraint[%s](fmt.Sprintf(\"%%s_pkey\", (new(%s)).ModelName()))\n", typeName, typeName, typeName)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if *typeNames == "" {
		flag.Usage()
		os.Exit(2)
	}
	types := strings.Split(*typeNames, ",")

	pkg, err := loadPackage()
	if err != nil {
		panic(err)
	}

	header := bytes.NewBufferString(genHeader)
	header.WriteString(pkg.Name())
	header.WriteString("\n")

	contents := &fileContent{
		header:     header,
		importsMap: map[string]bool{"github.com/imperfect-fourth/eywa": true},
		imports:    bytes.NewBuffer([]byte{}),
		content:    bytes.NewBufferString(""),
	}
	for _, t := range types {
		if err := parseType(t, pkg, contents); err != nil {
			panic(err)
		}
	}
	if len(contents.importsMap) > 0 {
		contents.imports.WriteString("\nimport (\n")
		for pkgImport, ok := range contents.importsMap {
			if ok {
				contents.imports.WriteString(fmt.Sprintf("\t\"%s\"\n", pkgImport))
			}
		}
		contents.imports.WriteString(")\n\n")
	}
	if err := writeToFile(*outputFile, contents); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type fileContent struct {
	header     *bytes.Buffer
	importsMap map[string]bool
	imports    *bytes.Buffer
	content    *bytes.Buffer
}

var parsed = make(map[string]bool)

func parseType(typeName string, pkg *types.Package, contents *fileContent) error {
	if parsed[typeName] {
		return nil
	}
	parsed[typeName] = true

	typeObj := pkg.Scope().Lookup(typeName)
	if typeObj == nil {
		fmt.Printf("type %s not found in package, skipping...", typeName)
		return nil
	}
	typeStruct, ok := typeObj.Type().Underlying().(*types.Struct)
	if !ok {
		fmt.Printf("type %s is not a struct, skipping...", typeName)
		return nil
	}
	if types.NewMethodSet(types.NewPointer(typeObj.Type())).Lookup(pkg, "ModelName") == nil {
		fmt.Printf("struct type %s does not implement eywa.Model interface, skipping...", typeName)
		return nil
	}

	contents.content.WriteString("\n")
	recurseParse := make([]string, 0, typeStruct.NumFields())
	foundPkey := false
	pkey := ""
	for i := 0; i < typeStruct.NumFields(); i++ {
		tag := tagPattern.FindStringSubmatch(typeStruct.Tag(i))
		if tag == nil {
			continue
		}
		tagValues := strings.Split(tag[1], ",")
		if len(tagValues) == 0 {
			continue
		}
		fieldName := tagValues[0]
		if eywaTag := eywaTagPattern.FindStringSubmatch(typeStruct.Tag(i)); eywaTag != nil {
			if eywaTagValues := strings.Split(eywaTag[1], ","); len(eywaTagValues) != 0 {
				for _, v := range eywaTagValues {
					if v == "pkey" && foundPkey {
						return fmt.Errorf("model %s has two primary keys: %s, %s", typeName, pkey, fieldName)
					}
					if v == "pkey" {
						foundPkey = true
						contents.importsMap["fmt"] = true
						contents.content.WriteString(pkeyConstraint(typeName))
						pkey = fieldName
					}
				}
			}
		}
		field := typeStruct.Field(i)
		fieldType := field.Type()
		importPackages, fieldTypeNameFull := parseFieldTypeName(field.Type().String(), pkg.Path())
		for _, p := range importPackages {
			contents.importsMap[p] = true
		}
		fieldTypeName := fieldTypeNameFull
		if fieldTypeNameFull[0] == '*' {
			fieldTypeName = fieldTypeNameFull[1:]
		}
		fieldScalarGqlType := gqlType(fieldType.Underlying().String())

		// *struct -> struct, *[] -> [], *int -> int, etc
		if ptr, ok := fieldType.(*types.Pointer); ok {
			fieldType = ptr.Elem()
		}
		// []*x -> *x, []x -> x
		if slice, ok := fieldType.(*types.Slice); ok {
			fieldType = slice.Elem()
		} else if array, ok := fieldType.(*types.Array); ok {
			fieldType = array.Elem()
		}
		// struct -> *struct
		var fieldGqlType string
		if _, ok := fieldType.Underlying().(*types.Struct); ok {
			fieldType = types.NewPointer(fieldType)
			fieldGqlType = "eywa.JSONValue | eywa.JSONBValue"
		} else if _, ok := fieldType.Underlying().(*types.Map); ok {
			fieldGqlType = "eywa.JSONValue | eywa.JSONBValue"
		}

		switch fieldType := fieldType.(type) {
		case *types.Pointer:
			fieldMethodSet := types.NewMethodSet(fieldType)
			if m := fieldMethodSet.Lookup(pkg, "ModelName"); m != nil && m.Type().String() == "func() string" {
				contents.importsMap["bytes"] = true
				contents.content.WriteString(fmt.Sprintf(
					modelRelationshipNameFunc,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					fieldTypeName,
					fieldTypeName,
					typeName,
					fieldName,
					typeName,
				))
				recurseParse = append(recurseParse, fieldTypeName)
			} else {
				contents.content.WriteString(fmt.Sprintf(
					modelFieldNameConst,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					typeName,
					fieldName,
				))
				contents.content.WriteString(fmt.Sprintf(
					modelFieldFunc,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					fieldTypeNameFull,
					typeName,
					typeName,
					fieldName,
				))
				if fieldScalarGqlType != "" {
					contents.content.WriteString(fmt.Sprintf(
						modelScalarVarFunc,
						fmt.Sprintf("%s_%s", typeName, field.Name()),
						fieldTypeNameFull,
						typeName,
						typeName,
						fieldName,
						fmt.Sprintf("%s_%s", typeName, field.Name()),
						fmt.Sprintf("eywa.%s", fieldScalarGqlType),
						fieldTypeNameFull,
					))
				} else if fieldGqlType != "" {
					contents.content.WriteString(fmt.Sprintf(
						modelVarFunc,
						fmt.Sprintf("%s_%s", typeName, field.Name()),
						fieldGqlType,
						fieldTypeNameFull,
						typeName,
						typeName,
						fieldName,
						fmt.Sprintf("%s_%s", typeName, field.Name()),
					))
				}
			}
		default:
			contents.content.WriteString(fmt.Sprintf(
				modelFieldNameConst,
				fmt.Sprintf("%s_%s", typeName, field.Name()),
				typeName,
				fieldName,
			))
			contents.content.WriteString(fmt.Sprintf(
				modelFieldFunc,
				fmt.Sprintf("%s_%s", typeName, field.Name()),
				fieldTypeNameFull,
				typeName,
				typeName,
				fieldName,
			))
			if fieldScalarGqlType != "" {
				contents.content.WriteString(fmt.Sprintf(
					modelScalarVarFunc,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					fieldTypeNameFull,
					typeName,
					typeName,
					fieldName,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					fmt.Sprintf("eywa.%sVar", fieldScalarGqlType),
					fieldTypeNameFull,
				))
			} else if fieldGqlType != "" {
				contents.content.WriteString(fmt.Sprintf(
					modelVarFunc,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
					fieldGqlType,
					fieldTypeNameFull,
					typeName,
					typeName,
					fieldName,
					fmt.Sprintf("%s_%s", typeName, field.Name()),
				))
			}
		}
	}
	for _, t := range recurseParse {
		if err := parseType(t, pkg, contents); err != nil {
			return err
		}
	}
	return nil
}

func writeToFile(filename string, contents *fileContent) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := contents.header.WriteTo(f); err != nil {
		return err
	}
	if _, err := contents.imports.WriteTo(f); err != nil {
		return err
	}
	if _, err := contents.content.WriteTo(f); err != nil {
		return err
	}
	return nil
}

func loadPackage() (*types.Package, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo, Tests: true}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("couldn't load package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package contains errors")
	}
	return pkgs[0].Types, nil
}

func parseFieldTypeName(name, rootPkgPath string) (importPackages []string, typeName string) {
	genericTypeRegex := re.MustCompile(`^(.*?)(\[(.*)\])?$`)
	genericMatches := genericTypeRegex.FindStringSubmatch(name)

	rgx := re.MustCompile(`^(\*)?(.*/(.*))\.(.*)$`)
	matches := rgx.FindStringSubmatch(genericMatches[1])
	// basic types: int, string, etc
	if len(matches) == 0 {
		return nil, name
	}
	importPackages = []string{}
	pointer := matches[1]
	typePackagePath := matches[2]
	typePackageName := matches[3]
	typeName = matches[4]
	// if type has generic type parameters
	if genericMatches[2] != "" {
		typeParams := strings.Split(genericMatches[3], ", ")
		typeParamNames := []string{}
		for _, tp := range typeParams {
			tpImportPackages, tpTypeName := parseFieldTypeName(tp, rootPkgPath)
			importPackages = append(importPackages, tpImportPackages...)
			typeParamNames = append(typeParamNames, tpTypeName)
		}
		typeName = fmt.Sprintf("%s[%s]", typeName, strings.Join(typeParamNames, ", "))
	}
	// if type's source pkg is not the same as root package, import
	if rootPkgPath == typePackagePath {
		return importPackages, fmt.Sprintf("%s%s", pointer, typeName)
	}
	importPackages = append(importPackages, matches[2])
	return importPackages, fmt.Sprintf("%s%s.%s", pointer, typePackageName, typeName)
}

var gqlTypes = map[string]string{
	"bool":    "Boolean",
	"*bool":   "NullableBoolean",
	"int":     "Int",
	"*int":    "NullableInt",
	"float":   "Float",
	"*float":  "NullableFloat",
	"string":  "String",
	"*string": "NullableString",
}

func gqlType(fieldType string) string {
	for k, v := range gqlTypes {
		if strings.HasPrefix(fieldType, k) {
			return v
		}
	}
	return ""
}
