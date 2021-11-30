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

func NewClientGenerator() *ClientGenerator {
	g := &ClientGenerator{}
	g.extension = ".ts"
	g.language = "Client"
	g.rootFolder = "client"
	g.gen = g
	return g
}

func (g *ClientGenerator) init(s *model.Schema) {
	g.GenBase.init(s, clienttemplates.TypeDependent, clienttemplates.Templates)
}

func (g *ClientGenerator) Generate(s *model.Schema) error {
	g.gen.init(s)

	g.folder = g.rootFolder + "/"
	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.folder + "events" + g.extension)
	if err == nil && info.ModTime().After(s.SchemaTime) {
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
	return nil
}
