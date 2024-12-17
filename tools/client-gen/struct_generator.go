package client_gen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func (g *TypeGenerator) generateArgsStruct(funcName string, suffix string, args []CompiledField) {
	// Generate dependencies first
	g.generateDependencies(args)

	fields := g.generateStructFields(args)
	structName := fmt.Sprintf("%s%s", funcName, suffix)

	if !g.generated[structName] {
		structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
			structName, structName, strings.Join(fields, ",\n"))
		g.output = append(g.output, structDef)
		g.generated[structName] = true
	}
}

func (g *TypeGenerator) generateDependencies(args []CompiledField) {
	for _, arg := range args {
		argType := dereferenceType(arg.Type)

		if _, isOverride := g.isOverriddenType(argType); isOverride {
			continue
		}

		if _, isEnum := bcs.EnumTypes[argType]; isEnum {
			g.GenerateType(argType)
		} else if argType.Kind() == reflect.Struct {
			g.GenerateType(argType)
		}
	}
}

func (g *TypeGenerator) generateStructFields(args []CompiledField) []string {
	fields := make([]string, 0, len(args))
	for _, arg := range args {
		typeStr := g.getBCSType(arg.Type)
		if arg.IsOptional {
			typeStr = fmt.Sprintf("bcs.option(%s)", typeStr)
		}
		fields = append(fields, fmt.Sprintf("\t%s: %s", arg.Name, typeStr))
	}
	return fields
}

func (g *TypeGenerator) GenerateType(t reflect.Type) {
	t = dereferenceType(t)
	qualifiedName := getQualifiedTypeName(t)

	if g.generated[qualifiedName] {
		return
	}

	if _, isOverride := g.isOverriddenType(t); isOverride {
		return
	}

	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		g.generateEnumType(t)
		g.generated[qualifiedName] = true
		return
	}

	if t.Kind() != reflect.Struct {
		return
	}

	g.generateStructType(t, qualifiedName)
}

func (g *TypeGenerator) generateStructType(t reflect.Type, qualifiedName string) {
	// Generate nested types first
	for i := 0; i < t.NumField(); i++ {
		fieldType := dereferenceType(t.Field(i).Type)

		if _, isOverride := g.isOverriddenType(fieldType); isOverride {
			continue
		}

		if _, isEnum := bcs.EnumTypes[fieldType]; isEnum {
			g.GenerateType(fieldType)
		} else if fieldType.Kind() == reflect.Struct {
			g.GenerateType(fieldType)
		}
	}

	// Generate struct fields
	fields := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typeStr := g.getBCSType(field.Type)
		fieldName := strings.ToLower(field.Name[:1]) + field.Name[1:]
		fields = append(fields, fmt.Sprintf("\t%s: %s", fieldName, typeStr))
	}

	structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
		qualifiedName, qualifiedName, strings.Join(fields, ",\n"))

	g.output = append(g.output, structDef)
	g.generated[qualifiedName] = true
}
