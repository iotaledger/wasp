package testcore

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/samber/lo"
	"reflect"
	"regexp"
	"testing"
)

func getName(p reflect.Type) string {
	// Regex to run
	r := regexp.MustCompile(`\.(.*)\[`)

	// Return capture
	return r.FindStringSubmatch(p.String())[1]
}

var SupportedFieldOptional = reflect.TypeOf(coreutil.FieldOptional[any](""))
var SupportedField = reflect.TypeOf(coreutil.Field[any](""))

func isSupportedType(p reflect.Type) bool {
	fmt.Println(SupportedFieldOptional.Name())
	fmt.Println(SupportedFieldOptional.String())

	if getName(SupportedField) == getName(p) {
		return true
	}

	if getName(SupportedFieldOptional) == getName(p) {
		return true
	}

	return false
}

func isOptional(p reflect.Type) bool {
	return getName(SupportedFieldOptional) == getName(p)
}

type CoreContractFunction struct {
	ContractName string
	FunctionName string

	InputArgs  []CompiledField
	OutputArgs []CompiledField
}

type CompiledField struct {
	Name       string
	ArgIndex   int
	Type       string
	IsOptional bool
}

type CoreContractFunctionStructure interface {
	Inputs() []coreutil.FieldArg
	Outputs() []coreutil.FieldArg
	Hname() isc.Hname
	String() string
	ContractInfo() *coreutil.ContractInfo
}

func extractFields(fields []coreutil.FieldArg) []CompiledField {
	compiled := make([]CompiledField, len(fields))

	for i, input := range fields {
		fieldType := reflect.TypeOf(input)
		inputType := input.Type()

		fmt.Println(inputType)

		if !isSupportedType(fieldType) {
			panic(fmt.Errorf("%s is not supported -> fieldName: %s", fieldType.Name(), input.Name()))
		}

		compiled[i] = CompiledField{
			ArgIndex:   i,
			Name:       input.Name(),
			IsOptional: isOptional(fieldType),
			Type:       inputType.String(),
		}
	}

	return compiled
}

func constructCoreContractFunction(f CoreContractFunctionStructure) *CoreContractFunction {
	return &CoreContractFunction{
		ContractName: f.ContractInfo().Name,
		FunctionName: f.String(),
		InputArgs:    extractFields(f.Inputs()),
		OutputArgs:   extractFields(f.Outputs()),
	}
}

func TestA(t *testing.T) {
	k := constructCoreContractFunction(&accounts.FuncTransferAccountToChain)
	fmt.Println(string(lo.Must(json.Marshal(k))))
}
