// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/rstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type RustGenerator struct {
	GenBase
}

func NewRustGenerator(s *model.Schema) *RustGenerator {
	g := &RustGenerator{}
	g.init(s, rstemplates.TypeDependent, rstemplates.Templates)
	return g
}

func (g *RustGenerator) Generate() error {
	err := g.generateCommonFiles()
	if err != nil {
		return err
	}

	// now generate language-specific files

	if g.s.CoreContracts {
		return g.createSourceFile("mod", true)
	}

	cargoToml := "Cargo.toml"
	return g.createFile(cargoToml, false, func() {
		g.emit(cargoToml)
	})
}
