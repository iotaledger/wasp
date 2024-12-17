package client_gen

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"math/big"
	"reflect"
	"strings"
)

type CoreContractFunction struct {
	ContractName string
	FunctionName string
	IsView       bool
	InputArgs    []CompiledField
	OutputArgs   []CompiledField
}

type CompiledField struct {
	Name       string
	ArgIndex   int
	Type       reflect.Type
	IsOptional bool
}

type CoreContractFunctionStructure interface {
	Inputs() []coreutil.FieldArg
	Outputs() []coreutil.FieldArg
	Hname() isc.Hname
	String() string
	ContractInfo() *coreutil.ContractInfo
	IsView() bool
}

func (g *TypeGenerator) generateEnumType(enumType reflect.Type) {
	// Get registered enum types from BCS
	enumTypes := bcs.EnumTypes[enumType]
	if enumTypes == nil {
		return
	}

	variants := make([]string, 0)

	// Generate struct types for each variant first
	for i := 1; i < len(enumTypes); i++ { // Skip index 0 (None)
		implType := enumTypes[i]
		if implType.Kind() == reflect.Ptr {
			implType = implType.Elem()
		}
		if implType.Kind() == reflect.Struct {
			g.GenerateType(implType)
		}
	}

	// Generate enum variants
	for i := 0; i < len(enumTypes); i++ {
		var (
			variantName string
			typeStr     string
		)

		if i == 0 {
			variantName = "NoType"
			typeStr = "null"
		} else {
			implType := enumTypes[i]
			if implType.Kind() == reflect.Ptr {
				implType = implType.Elem()
			}
			variantName = implType.Name()
			typeStr = getQualifiedTypeName(implType)
		}

		variants = append(variants, fmt.Sprintf("\t%s: %s", variantName, typeStr))
	}

	enumName := getQualifiedTypeName(enumType)
	enumDef := fmt.Sprintf("const %s = bcs.enum('%s', {\n%s\n});",
		enumName,
		enumName,
		strings.Join(variants, ",\n"))

	g.output = append(g.output, enumDef)
}

func getBCSType(t reflect.Type) string {
	// Dereference pointer if needed
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if it's a registered enum type first
	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		return getQualifiedTypeName(t)
	}

	switch t.Kind() {
	case reflect.Interface:
		// Check if it's a registered enum type
		if _, isEnum := bcs.EnumTypes[t]; isEnum {
			return getQualifiedTypeName(t)
		}
		// If it's not a registered enum, we might want to panic or handle differently
		return fmt.Sprintf("/* Unsupported type: %v %v */", t.Kind(), t.String())
	case reflect.Bool:
		return "bcs.bool()"

	// Unsigned integers
	case reflect.Uint8:
		return "bcs.u8()"
	case reflect.Uint16:
		return "bcs.u16()"
	case reflect.Uint32:
		return "bcs.u32()"
	case reflect.Uint64:
		return "bcs.u64()"

	case reflect.String:
		return "bcs.string()"

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "bcs.bytes()"
		}
		elemType := getBCSType(t.Elem())
		return fmt.Sprintf("bcs.vector(%s)", elemType)

	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("bcs.fixedBytes(%d)", t.Len())
		}
		elemType := getBCSType(t.Elem())
		return fmt.Sprintf("bcs.fixedArray(%s, %d)", elemType, t.Len())

	case reflect.Map:
		keyType := getBCSType(t.Key())
		valueType := getBCSType(t.Elem())
		return fmt.Sprintf("bcs.map(%s, %s)", keyType, valueType)

	case reflect.Struct:
		if t.String() == reflect.TypeOf(new(big.Int)).String() {
			return "bcs.u256()"
		}

		// Use qualified name for struct references
		fmt.Println(t.String() + " " + t.Name())
		return getQualifiedTypeName(t)

	default:
		return fmt.Sprintf("/* Unsupported type: %v */", t.Kind())
	}
}

// Helper function to create qualified type name
func getQualifiedTypeName(t reflect.Type) string {
	if t.PkgPath() != "" {
		// Convert package path's last segment and type name to PascalCase
		pkgName := t.PkgPath()[strings.LastIndex(t.PkgPath(), "/")+1:]
		return fmt.Sprintf("%s%s", strings.Title(pkgName), t.Name())
	}
	return t.Name()
}

type TypeGenerator struct {
	generated map[string]bool
	output    []string
}

func NewTypeGenerator() *TypeGenerator {
	return &TypeGenerator{
		generated: make(map[string]bool),
		output:    make([]string, 0),
	}
}

// generateArgsStruct handles both input and output args
func (g *TypeGenerator) generateArgsStruct(funcName string, suffix string, args []CompiledField) {
	// First generate any struct types used in args
	for _, arg := range args {
		argType := arg.Type
		if argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		// Check if it's an enum
		if _, isEnum := bcs.EnumTypes[argType]; isEnum {
			g.GenerateType(argType)
			continue
		}

		// Handle struct types
		if argType.Kind() == reflect.Struct && argType.String() != reflect.TypeOf(new(big.Int)).String() {
			g.GenerateType(argType)
		}
	}
	fields := make([]string, 0)
	for _, arg := range args {
		var typeStr string
		argType := arg.Type
		if argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		// Check if it's an enum
		if _, isEnum := bcs.EnumTypes[argType]; isEnum {
			typeStr = getQualifiedTypeName(argType)
		} else if argType.Kind() == reflect.Struct && argType.String() != reflect.TypeOf(new(big.Int)).String() {
			typeStr = getQualifiedTypeName(argType)
		} else {
			typeStr = getBCSType(arg.Type)
		}

		// Wrap in optional if field is marked as optional
		if arg.IsOptional {
			typeStr = fmt.Sprintf("bcs.optional(%s)", typeStr)
		}

		fields = append(fields, fmt.Sprintf("\t%s: %s", arg.Name, typeStr))
	}

	structName := fmt.Sprintf("%s%s", funcName, suffix)
	structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
		structName,
		structName,
		strings.Join(fields, ",\n"))

	if !g.generated[structName] {
		g.output = append(g.output, structDef)
		g.generated[structName] = true
	}
}

func (g *TypeGenerator) GenerateFunction(cf CoreContractFunction) {
	// Generate input args struct
	if len(cf.InputArgs) > 0 {
		g.generateArgsStruct(cf.FunctionName, "Inputs", cf.InputArgs)
	}

	// Generate output args struct
	if len(cf.OutputArgs) > 0 {
		g.generateArgsStruct(cf.FunctionName, "Outputs", cf.OutputArgs)
	}
}
func (g *TypeGenerator) GenerateType(t reflect.Type) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	qualifiedName := getQualifiedTypeName(t)

	// Skip if already generated
	if g.generated[qualifiedName] {
		return
	}

	// Check if it's a registered enum type
	if _, isEnum := bcs.EnumTypes[t]; isEnum {
		g.generateEnumType(t)
		g.generated[qualifiedName] = true
		return
	}

	// Only continue for struct types
	if t.Kind() != reflect.Struct {
		return
	}

	// First generate any nested struct dependencies
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if field type is an enum
		if _, isEnum := bcs.EnumTypes[fieldType]; isEnum {
			g.GenerateType(fieldType)
			continue
		}

		// Handle nested structs
		if fieldType.Kind() == reflect.Struct && fieldType.String() != reflect.TypeOf(new(big.Int)).String() {
			g.GenerateType(fieldType)
		}
	}

	fields := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		var typeStr string
		if fieldType.Kind() == reflect.Struct && fieldType.String() != reflect.TypeOf(new(big.Int)).String() {
			// Use qualified name for struct references
			typeStr = getQualifiedTypeName(fieldType)
		} else {
			typeStr = getBCSType(fieldType)
		}

		fieldName := strings.ToLower(field.Name[:1]) + field.Name[1:]
		fields = append(fields, fmt.Sprintf("\t%s: %s", fieldName, typeStr))
	}

	structDef := fmt.Sprintf("const %s = bcs.struct('%s', {\n%s\n});",
		qualifiedName,
		qualifiedName,
		strings.Join(fields, ",\n"))

	g.output = append(g.output, structDef)
	g.generated[qualifiedName] = true
}

func (g *TypeGenerator) GetOutput() string {
	return strings.Join(g.output, "\n\n")
}
