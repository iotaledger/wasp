// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"github.com/iotaledger/wasp/tools/schema/generator/goclienttemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type GoClientGenerator struct {
	ClientBase
}

func NewGoClientGenerator(s *model.Schema) *GoClientGenerator {
	g := &GoClientGenerator{}
	g.init(s, goclienttemplates.TypeDependent, goclienttemplates.Templates)
	return g
}
