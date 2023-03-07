// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
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

func (g *RustGenerator) Build() error {
	wasm := g.s.PackageName + "wasm"
	fmt.Printf("building %s_bg.wasm\n", wasm)
	args := "build rs/" + wasm
	return g.build("wasm-pack", args)
}

func (g *RustGenerator) Cleanup() {
	g.cleanCommonFiles()
	g.cleanSourceFile("lib")

	g.keys["cargoMain"] = "Impl"
	g.cleanSourceFileIfSame(g.generateCargoToml)
	g.cleanSourceFileIfSame(g.generateLicense)
	g.cleanSourceFileIfSame(g.generateReadmeMd)

	// now clean up Wasm VM host stub crate
	g.generateCommonFolder("wasm", false)
	g.cleanFolder(g.folder)
}

func (g *RustGenerator) GenerateImplementation() error {
	err := g.generateImplementation()
	if err != nil {
		return err
	}

	err = g.createSourceFile("lib", true)
	if err != nil {
		return err
	}

	err = g.generateCargoFiles("Impl")
	if err != nil {
		return err
	}
	return nil
}

func (g *RustGenerator) GenerateInterface() error {
	err := g.generateInterface()
	if err != nil {
		return err
	}

	if g.s.CoreContracts {
		return g.createSourceFile("mod", true)
	}

	err = g.createSourceFile("lib", true, "mod")
	if err != nil {
		return err
	}

	err = g.generateCargoFiles("Lib")
	if err != nil {
		return err
	}
	return nil
}

func (g *RustGenerator) GenerateWasmStub() error {
	g.generateCommonFolder("wasm", true)
	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}

	// would have preferred to use main.rs, but don't want to generate both a lib.rs
	// AND a main.rs, so we generate a different lib.rs by using a different macro name
	err = g.createSourceFile("lib", !g.s.CoreContracts, "main")
	if err != nil {
		return err
	}

	return g.generateCargoFiles("Wasm")
}

func (g *RustGenerator) generateCargoFiles(cargoMain string) error {
	g.keys["cargoMain"] = cargoMain

	err := g.generateCargoToml()
	if err != nil {
		return err
	}

	err = g.generateLicense()
	if err != nil {
		return err
	}

	return g.generateReadmeMd()
}

func (g *RustGenerator) generateCargoToml() error {
	const cargoToml = "../Cargo.toml"
	return g.createFile(g.folder+cargoToml, false, func() {
		g.emit(cargoToml)
	})
}

func (g *RustGenerator) generateLicense() error {
	const license = "../LICENSE"
	return g.createFile(g.folder+license, false, func() {
		g.emit(license)
	})
}

func (g *RustGenerator) generateReadmeMd() error {
	const readMe = "../README.md"
	return g.createFile(g.folder+readMe, false, func() {
		g.emit(readMe + " " + g.keys["cargoMain"])
	})
}
