// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/gotemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type GoGenerator struct {
	GenBase
}

var _ IGenerator = new(GoGenerator)

func NewGoGenerator(s *model.Schema) *GoGenerator {
	g := &GoGenerator{}
	g.init(s, gotemplates.TypeDependent, gotemplates.Templates)
	return g
}

func (g *GoGenerator) Cleanup() {
	g.cleanCommonFiles()

	// now clean up language-specific files
}

func (g *GoGenerator) GenerateImplementation() error {
	err := g.generateImplementation()
	if err != nil {
		return err
	}
	return nil
}

func (g *GoGenerator) GenerateInterface() error {
	err := g.generateInterface()
	if err != nil {
		return err
	}
	return nil
}

func (g *GoGenerator) GenerateWasmStub() error {
	return g.createSourceFile("../main", !g.s.CoreContracts)
}
