package main

import (
	"reflect"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
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
	Type       reflect.Type `json:"-"`
	TypeName   string
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

type TypeOverride struct {
	SourceType reflect.Type
	BCSType    string
}

type TypeDefinition struct {
	Name         string
	Definition   string
	Dependencies []string
}

type TypeGenerator struct {
	generated     map[string]bool
	output        []TypeDefinition
	typeOverrides map[string]string
}
