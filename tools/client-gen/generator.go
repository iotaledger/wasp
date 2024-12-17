package client_gen

import (
	"math/big"
	"reflect"
	"strings"
	"time"
)

func NewTypeGenerator() *TypeGenerator {
	g := &TypeGenerator{
		generated:     make(map[string]bool),
		output:        make([]string, 0),
		typeOverrides: make(map[string]string),
	}
	g.registerDefaultOverrides()
	return g
}

func (g *TypeGenerator) registerDefaultOverrides() {
	defaultOverrides := map[reflect.Type]string{
		reflect.TypeOf(new(big.Int)).Elem():   "bcs.u256()",
		reflect.TypeOf(new(time.Time)).Elem(): "bcs.u64()",
	}

	for t, bcsType := range defaultOverrides {
		g.RegisterTypeOverride(t, bcsType)
	}
}

func (g *TypeGenerator) RegisterTypeOverride(t reflect.Type, bcsType string) {
	g.typeOverrides[t.String()] = bcsType
}

func (g *TypeGenerator) isOverriddenType(t reflect.Type) (string, bool) {
	bcsType, exists := g.typeOverrides[t.String()]
	return bcsType, exists
}

func (g *TypeGenerator) GenerateFunction(cf CoreContractFunction) {
	if len(cf.InputArgs) > 0 {
		g.generateArgsStruct(cf.FunctionName, "Inputs", cf.InputArgs)
	}
	if len(cf.OutputArgs) > 0 {
		g.generateArgsStruct(cf.FunctionName, "Outputs", cf.OutputArgs)
	}
}

func (g *TypeGenerator) GetOutput() string {
	return strings.Join(g.output, "\n\n")
}
