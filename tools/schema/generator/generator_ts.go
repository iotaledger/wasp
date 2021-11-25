// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"

	"github.com/iotaledger/wasp/tools/schema/generator/tstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type TypeScriptGenerator struct {
	GenBase
}

func NewTypeScriptGenerator() *TypeScriptGenerator {
	g := &TypeScriptGenerator{}
	g.extension = ".ts"
	g.funcRegexp = regexp.MustCompile(`^export function (\w+).+$`)
	g.language = "TypeScript"
	g.rootFolder = "ts"
	g.gen = g
	return g
}

func (g *TypeScriptGenerator) init(s *model.Schema) {
	g.GenBase.init(s, tstemplates.TypeDependent, tstemplates.Templates)
}

func (g *TypeScriptGenerator) generateLanguageSpecificFiles() error {
	err := g.createSourceFile("index", true)
	if err != nil {
		return err
	}

	tsconfig := "tsconfig.json"
	return g.createFile(g.folder+tsconfig, false, func() {
		g.emit(tsconfig)
	})
}
