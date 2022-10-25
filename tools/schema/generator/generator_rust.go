// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"os"

	"github.com/iotaledger/wasp/tools/schema/generator/rstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type RustGenerator struct {
	GenBase
}

var _ IGenerator = new(RustGenerator)

func NewRustGenerator(s *model.Schema) *RustGenerator {
	g := &RustGenerator{}
	g.init(s, rstemplates.TypeDependent, rstemplates.Templates)
	return g
}

func (g *RustGenerator) Cleanup() {
	g.cleanCommonFiles()

	// now clean up language-specific files
	g.cleanFolder(g.folder + "../../main")
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

	g.keys["cargoMain"] = "Sc"
	cargoToml := "../Cargo.toml"
	err = g.createFile(g.folder+cargoToml, false, func() {
		g.emit(cargoToml)
	})
	if err != nil {
		return err
	}

	err = os.MkdirAll(g.folder+"../../main/src", 0o755)
	if err != nil {
		return err
	}

	err = g.createSourceFile("../../main/src/lib", !g.s.CoreContracts)
	if err != nil {
		return err
	}

	g.keys["cargoMain"] = "Main"
	g.folder += "../../main/src/"
	err = g.createFile(g.folder+cargoToml, false, func() {
		g.emit(cargoToml)
	})

	return err
}
