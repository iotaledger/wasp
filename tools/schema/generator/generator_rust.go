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
	g.cleanFolder(g.folder + "../../" + g.s.PackageName + "_main")
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

	err = g.GenerateCargoToml("Sc")
	if err != nil {
		return err
	}

	g.folder += "../../" + g.s.PackageName + "_main/src/"
	err = os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}

	// would have preferred to use main.rs, but don't want to generate both a lib.rs
	// AND a main.rs so the mainRs template is tricked into generating a different
	// lib.rs than the actual lib.rs by using a relative path
	err = g.createSourceFile("../src/lib", !g.s.CoreContracts)
	if err != nil {
		return err
	}

	return g.GenerateCargoToml("Main")
}

func (g *RustGenerator) GenerateCargoToml(cargoMain string) error {
	const cargoToml = "../Cargo.toml"
	g.keys["cargoMain"] = cargoMain
	err := g.createFile(g.folder+cargoToml, false, func() {
		g.emit(cargoToml)
	})
	if err != nil {
		return err
	}

	const license = "../LICENSE"
	return g.createFile(g.folder+license, false, func() {
		g.emit(license)
	})
}
