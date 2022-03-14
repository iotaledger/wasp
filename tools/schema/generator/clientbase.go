// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"os"
)

type ClientBase struct {
	GenBase
}

func (g *ClientBase) Generate() error {
	g.folder = g.rootFolder + "/" + g.s.PackageName + "client/"
	if g.s.CoreContracts {
		g.folder = g.rootFolder + "/wasmclient/" + g.s.PackageName + "/"
	}
	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.folder + "service" + g.extension)
	if err == nil && info.ModTime().After(g.s.SchemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return nil
	}

	fmt.Printf("generating %s code\n", g.language)
	return g.generateCode()
}

func (g *ClientBase) generateCode() error {
	err := g.createSourceFile("events", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	return g.createSourceFile("service", true)
}
