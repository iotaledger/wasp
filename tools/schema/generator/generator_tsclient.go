// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/tsclienttemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type TsClientGenerator struct {
	ClientBase
}

func NewTsClientGenerator(s *model.Schema) *TsClientGenerator {
	g := &TsClientGenerator{}
	g.init(s, tsclienttemplates.TypeDependent, tsclienttemplates.Templates)
	return g
}

func (g *TsClientGenerator) Generate() error {
	err := g.ClientBase.Generate()
	if err != nil {
		return err
	}
	if g.s.CoreContracts {
		return nil
	}

	// now generate language-specific files

	err = g.createSourceFile("index", true)
	if err != nil {
		return err
	}

	tsconfig := "tsconfig.json"
	return g.createFile(g.folder+tsconfig, false, func() {
		g.emit(tsconfig)
	})
}
