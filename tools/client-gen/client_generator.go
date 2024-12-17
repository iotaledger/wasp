package client_gen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func (g *TypeGenerator) generateEnumType(enumType reflect.Type) {
	enumTypes := bcs.EnumTypes[enumType]
	if enumTypes == nil {
		return
	}

	variants := make([]string, 0)

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
			g.GenerateType(implType)
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
		typeStr := getQualifiedTypeName(implType)
		variants = append(variants, fmt.Sprintf("\t%s: %s", variantName, typeStr))
	}

	enumName := getQualifiedTypeName(enumType)
	enumDef := fmt.Sprintf("const %s = bcs.enum('%s', {\n%s\n});",
		enumName, enumName, strings.Join(variants, ",\n"))

	g.output = append(g.output, enumDef)
}

func (g *TypeGenerator) getBCSType(t reflect.Type) string {
	t = dereferenceType(t)

	if bcsType, exists := g.getOverrideOrEnumType(t); exists {
		return bcsType
	}

	return g.getBaseBCSType(t)
}

func (g *TypeGenerator) getOverrideOrEnumType(t reflect.Type) (string, bool) {
	if overrideType, exists := g.isOverriddenType(t); exists {
		return overrideType, true
	}

	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		return getQualifiedTypeName(t), true
	}

	return "", false
}

func (g *TypeGenerator) getBaseBCSType(t reflect.Type) string {
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
		return "bcs.fixedBytes(1)"
	case reflect.Int16:
		return "bcs.fixedBytes(2)"
	case reflect.Int32:
		return "bcs.fixedBytes(4)"
	case reflect.Int64:
		return "bcs.fixedBytes(8)"
	case reflect.String:
		return "bcs.string()"
	case reflect.Slice:
		return g.handleSliceType(t)
	case reflect.Array:
		return g.handleArrayType(t)
	case reflect.Map:
		return fmt.Sprintf("bcs.map(%s, %s)", g.getBCSType(t.Key()), g.getBCSType(t.Elem()))
	case reflect.Struct:
		return getQualifiedTypeName(t)
	default:
		return fmt.Sprintf("/* Unsupported type: %v */", t.Kind())
	}
}

func (g *TypeGenerator) handleSliceType(t reflect.Type) string {
	if t.Elem().Kind() == reflect.Uint8 {
		return "bcs.bytes()"
	}
	return fmt.Sprintf("bcs.vector(%s)", g.getBCSType(t.Elem()))
}

func (g *TypeGenerator) handleArrayType(t reflect.Type) string {
	if t.Elem().Kind() == reflect.Uint8 {
		return fmt.Sprintf("bcs.fixedBytes(%d)", t.Len())
	}
	return fmt.Sprintf("bcs.fixedArray(%s, %d)", g.getBCSType(t.Elem()), t.Len())
}
