// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/gotemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type GoGenerator struct {
	GenBase
}

func NewGoGenerator(s *model.Schema) *GoGenerator {
	g := &GoGenerator{}
	g.init(s, gotemplates.TypeDependent, gotemplates.Templates)
	return g
}

func (g *GoGenerator) Generate() error {
	err := g.generateCommonFiles()
	if err != nil {
		return err
	}

	// now generate language-specific files

	return g.createSourceFile("../main", !g.s.CoreContracts)
}
