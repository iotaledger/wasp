// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"os"

	"github.com/iotaledger/wasp/tools/schema/generator/tstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type TypeScriptGenerator struct {
	GenBase
}

var _ IGenerator = new(TypeScriptGenerator)

func NewTypeScriptGenerator(s *model.Schema) *TypeScriptGenerator {
	g := &TypeScriptGenerator{}
	g.init(s, tstemplates.TypeDependent, tstemplates.Templates)
	return g
}

func (g *TypeScriptGenerator) Cleanup() {
	g.cleanCommonFiles()

	// now clean up language-specific files
	g.cleanSourceFile("index")
	_ = os.Remove(g.folder + "../tsconfig.json")
}

func (g *TypeScriptGenerator) GenerateImplementation() error {
	err := g.generateImplementation()
	if err != nil {
		return err
	}
	err = g.createSourceFile("index", true, "indexImpl")
	if err != nil {
		return err
	}
	err = g.generateTsConfig("")
	if err != nil {
		return err
	}
	return nil
}

func (g *TypeScriptGenerator) GenerateInterface() error {
	err := g.generateInterface()
	if err != nil {
		return err
	}
	err = g.createSourceFile("index", true)
	if err != nil {
		return err
	}
	err = g.generateTsConfig("")
	if err != nil {
		return err
	}
	return nil
}

func (g *TypeScriptGenerator) GenerateWasmStub() error {
	err := g.createSourceFile("../main", !g.s.CoreContracts)
	if err != nil {
		return err
	}
	err = g.generateTsConfig("../")
	if err != nil {
		return err
	}
	return nil
}

func (g *TypeScriptGenerator) generateTsConfig(folder string) error {
	const tsconfig = "tsconfig.json"
	return g.createFile(g.folder+folder+tsconfig, false, func() {
		g.emit(tsconfig)
	})
}
