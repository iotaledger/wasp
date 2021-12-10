// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/schema/generator/clienttemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

type ClientGenerator struct {
	GenBase
}

func NewClientGenerator(s *model.Schema) *ClientGenerator {
	g := &ClientGenerator{}
	g.init(s, clienttemplates.TypeDependent, clienttemplates.Templates)
	return g
}

func (g *ClientGenerator) Generate() error {
	g.folder = g.rootFolder + "/"
	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.folder + "events" + g.extension)
	if err == nil && info.ModTime().After(g.s.SchemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return nil
	}

	fmt.Printf("generating %s code\n", g.language)
	return g.generateCode()
}

func (g *ClientGenerator) generateCode() error {
	err := g.createSourceFile("events", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("app", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("service", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	return nil
}
