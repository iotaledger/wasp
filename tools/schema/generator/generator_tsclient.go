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
