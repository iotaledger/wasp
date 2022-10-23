// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/tstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type TypeScriptGenerator struct {
	GenBase
}

func NewTypeScriptGenerator(s *model.Schema) *TypeScriptGenerator {
	g := &TypeScriptGenerator{}
	g.init(s, tstemplates.TypeDependent, tstemplates.Templates)
	return g
}

func (g *TypeScriptGenerator) Generate() error {
	err := g.generateCommonFiles()
	if err != nil {
		return err
	}

	// now generate language-specific files

	err = g.createSourceFile("../main", !g.s.CoreContracts)
	if err != nil {
		return err
	}

	err = g.GenerateTsConfig("../")
	if err != nil {
		return err
	}

	err = g.GenerateTsConfig("")
	if err != nil {
		return err
	}

	return g.createSourceFile("index", true)
}

func (g *TypeScriptGenerator) GenerateTsConfig(folder string) error {
	tsconfig := "tsconfig.json"
	return g.createFile(g.folder+folder+tsconfig, false, func() {
		g.emit(tsconfig)
	})
}
