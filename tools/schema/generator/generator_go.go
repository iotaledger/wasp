// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"

	"github.com/iotaledger/wasp/tools/schema/generator/gotemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type GoGenerator struct {
	GenBase
}

func NewGoGenerator() *GoGenerator {
	g := &GoGenerator{}
	g.extension = ".go"
	g.funcRegexp = regexp.MustCompile(`^func (\w+).+$`)
	g.language = "Go"
	g.rootFolder = "go"
	g.gen = g
	return g
}

func (g *GoGenerator) init(s *model.Schema) {
	g.GenBase.init(s, gotemplates.TypeDependent, gotemplates.Templates)
}

func (g *GoGenerator) generateLanguageSpecificFiles() error {
	return g.createSourceFile("../main", !g.s.CoreContracts)
}
