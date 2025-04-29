package main

import (
	"fmt"
	"reflect"
	"strings"

	bcs "github.com/iotaledger/bcs-go"
)

func (tg *TypeGenerator) generateArgsStruct(funcName string, suffix string, args []CompiledField) {
	dependencies := make([]string, 0)

	// Generate dependencies first and collect them
	for _, arg := range args {
		argType := dereferenceType(arg.Type)

		if _, isOverride := tg.isOverriddenType(argType); isOverride {
			continue
		}

		fmt.Println(argType.Name() + " - " + arg.Name)

		if argType.Kind() == reflect.Struct {
			depName := getQualifiedTypeName(argType)
			dependencies = append(dependencies, depName)
			tg.GenerateType(argType)
		} else if _, isEnum := bcs.EnumTypes[argType]; isEnum {
			depName := getQualifiedTypeName(argType)
			dependencies = append(dependencies, depName)
			tg.GenerateType(argType)
		}
	}

	fields := tg.generateStructFields(args)
	structName := fmt.Sprintf("%s%s", funcName, suffix)

	if !tg.generated[structName] {
		structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
			structName, structName, strings.Join(fields, ",\n"))

		tg.output = append(tg.output, TypeDefinition{
			Name:         structName,
			Definition:   structDef,
			Dependencies: dependencies,
		})
		tg.generated[structName] = true
	}
}

func (tg *TypeGenerator) generateStructFields(args []CompiledField) []string {
	fields := make([]string, 0, len(args))
	for _, arg := range args {
		typeStr := tg.getBCSType(arg.Type)
		if arg.IsOptional {
			typeStr = fmt.Sprintf("bcs.option(%s)", typeStr)
		}
		fields = append(fields, fmt.Sprintf("\t%s: %s", arg.Name, typeStr))
	}
	return fields
}

func (tg *TypeGenerator) GenerateType(t reflect.Type) {
	t = dereferenceType(t)
	qualifiedName := getQualifiedTypeName(t)

	if tg.generated[qualifiedName] {
		return
	}

	if _, isOverride := tg.isOverriddenType(t); isOverride {
		return
	}

	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		tg.generateEnumType(t)
		tg.generated[qualifiedName] = true
		return
	}

	if t.Kind() != reflect.Struct {
		return
	}

	tg.generateStructType(t, qualifiedName)
}

func (tg *TypeGenerator) generateStructType(t reflect.Type, qualifiedName string) {
	dependencies := make([]string, 0)

	// Generate nested types first and collect dependencies
	for i := 0; i < t.NumField(); i++ {
		fieldType := dereferenceType(t.Field(i).Type)

		if _, isOverride := tg.isOverriddenType(fieldType); isOverride {
			continue
		}

		if fieldType.Kind() == reflect.Struct {
			depName := getQualifiedTypeName(fieldType)
			if depName != qualifiedName { // avoid self-dependency
				dependencies = append(dependencies, depName)
				tg.GenerateType(fieldType)
			}
		} else if _, isEnum := bcs.EnumTypes[fieldType]; isEnum {
			depName := getQualifiedTypeName(fieldType)
			dependencies = append(dependencies, depName)
			tg.GenerateType(fieldType)
		}
	}

	// Generate struct fields
	fields := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typeStr := tg.getBCSType(field.Type)
		fieldName := strings.ToLower(field.Name[:1]) + field.Name[1:]
		fields = append(fields, fmt.Sprintf("\t%s: %s", fieldName, typeStr))
	}

	structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
		qualifiedName, qualifiedName, strings.Join(fields, ",\n"))

	tg.output = append(tg.output, TypeDefinition{
		Name:         qualifiedName,
		Definition:   structDef,
		Dependencies: dependencies,
	})
	tg.generated[qualifiedName] = true
}
