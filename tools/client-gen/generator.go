package main

import (
	"math/big"
	"reflect"
	"time"
)

func NewTypeGenerator() *TypeGenerator {
	g := &TypeGenerator{
		generated:     make(map[string]bool),
		output:        make([]TypeDefinition, 0),
		typeOverrides: make(map[string]string),
	}
	g.registerDefaultOverrides()
	return g
}

func (tg *TypeGenerator) registerDefaultOverrides() {
	defaultOverrides := map[reflect.Type]string{
		reflect.TypeOf(new(big.Int)).Elem():   "bcs.u256()",
		reflect.TypeOf(new(time.Time)).Elem(): "bcs.u64()",
	}

	for t, bcsType := range defaultOverrides {
		tg.RegisterTypeOverride(t, bcsType)
	}
}

func (tg *TypeGenerator) RegisterTypeOverride(t reflect.Type, bcsType string) {
	tg.typeOverrides[t.String()] = bcsType
}

func (tg *TypeGenerator) isOverriddenType(t reflect.Type) (string, bool) {
	bcsType, exists := tg.typeOverrides[t.String()]
	return bcsType, exists
}

func (tg *TypeGenerator) GenerateFunction(cf CoreContractFunction) {
	if len(cf.InputArgs) > 0 {
		tg.generateArgsStruct(cf.FunctionName, "Inputs", cf.InputArgs)
	}
	if len(cf.OutputArgs) > 0 {
		tg.generateArgsStruct(cf.FunctionName, "Outputs", cf.OutputArgs)
	}
}
