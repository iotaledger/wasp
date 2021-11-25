// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"

	"github.com/iotaledger/wasp/tools/schema/generator/rstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type RustGenerator struct {
	GenBase
}

func NewRustGenerator() *RustGenerator {
	g := &RustGenerator{}
	g.extension = ".rs"
	g.funcRegexp = regexp.MustCompile(`^pub fn (\w+).+$`)
	g.language = "Rust"
	g.rootFolder = "src"
	g.gen = g
	return g
}

func (g *RustGenerator) init(s *model.Schema) {
	g.GenBase.init(s, rstemplates.TypeDependent, rstemplates.Templates)
}

func (g *RustGenerator) funcName(f *model.Func) string {
	return snake(g.GenBase.funcName(f))
}

func (g *RustGenerator) generateLanguageSpecificFiles() error {
	if g.s.CoreContracts {
		return g.createSourceFile("mod", true)
	}

	cargoToml := "Cargo.toml"
	return g.createFile(cargoToml, false, func() {
		g.emit(cargoToml)
	})
}
