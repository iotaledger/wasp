package main

import (
	"fmt"
	"reflect"
	"strings"

	bcs "github.com/iotaledger/bcs-go"
)

func (tg *TypeGenerator) generateEnumType(enumType reflect.Type) {
	enumTypes := bcs.EnumTypes[enumType]
	if enumTypes == nil {
		return
	}

	variants := make([]string, 0)
	dependencies := make([]string, 0)
	enumName := getQualifiedTypeName(enumType)

	// Generate struct types for variants first
	// Skip variant 0 as it's typically the None/empty variant
	for variantID, implType := range enumTypes {
		if variantID == 0 {
			continue
		}

		if implType.Kind() == reflect.Ptr {
			implType = implType.Elem()
		}
		if implType.Kind() == reflect.Struct {
			depName := getQualifiedTypeName(implType)
			if depName != enumName { // avoid self-dependency
				dependencies = append(dependencies, depName)
				tg.GenerateType(implType)
			}
		}
	}

	// Generate enum variants
	// Handle the None/empty variant first
	variants = append(variants, "\tNoType: null")

	// Then handle all other variants
	for variantID, implType := range enumTypes {
		if variantID == 0 {
			continue
		}

		if implType.Kind() == reflect.Ptr {
			implType = implType.Elem()
		}

		variantName := implType.Name()
		typeStr := tg.getBCSType(implType)
		variants = append(variants, fmt.Sprintf("\t%s: %s", variantName, typeStr))
	}

	enumDef := fmt.Sprintf("const %s = bcs.enum('%s', {\n%s\n});",
		enumName, enumName, strings.Join(variants, ",\n"))

	tg.output = append(tg.output, TypeDefinition{
		Name:         enumName,
		Definition:   enumDef,
		Dependencies: dependencies,
	})
}

func (tg *TypeGenerator) getBCSType(t reflect.Type) string {
	t = dereferenceType(t)

	if bcsType, exists := tg.getOverrideOrEnumType(t); exists {
		return bcsType
	}

	return tg.getBaseBCSType(t)
}

func (tg *TypeGenerator) getOverrideOrEnumType(t reflect.Type) (string, bool) {
	if overrideType, exists := tg.isOverriddenType(t); exists {
		return overrideType, true
	}

	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		return getQualifiedTypeName(t), true
	}

	return "", false
}

func (tg *TypeGenerator) getBaseBCSType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bcs.bool()"
	case reflect.Uint8:
		return "bcs.u8()"
	case reflect.Uint16:
		return "bcs.u16()"
	case reflect.Uint32:
		return "bcs.u32()"
	case reflect.Uint64:
		return "bcs.u64()"
	case reflect.Int8:
		return "bcs.bytes(1)"
	case reflect.Int16:
		return "bcs.bytes(2)"
	case reflect.Int32:
		return "bcs.bytes(4)"
	case reflect.Int64:
		return "bcs.bytes(8)"
	case reflect.String:
		return "bcs.string()"
	case reflect.Slice:
		return tg.handleSliceType(t)
	case reflect.Array:
		return tg.handleArrayType(t)
	case reflect.Map:
		return fmt.Sprintf("bcs.map(%s, %s)", tg.getBCSType(t.Key()), tg.getBCSType(t.Elem()))
	case reflect.Struct:
		return getQualifiedTypeName(t)
	default:
		return fmt.Sprintf("/* Unsupported type: %v */", t.Kind())
	}
}

func (tg *TypeGenerator) handleSliceType(t reflect.Type) string {
	tg.GenerateType(t.Elem())

	return fmt.Sprintf("bcs.vector(%s)", tg.getBCSType(t.Elem()))
}

func (tg *TypeGenerator) handleArrayType(t reflect.Type) string {
	tg.GenerateType(t.Elem())
	return fmt.Sprintf("bcs.fixedArray(%d, %s)", t.Len(), tg.getBCSType(t.Elem()))
}
