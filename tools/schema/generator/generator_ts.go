// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/schema/generator/tstemplates"
	"github.com/iotaledger/wasp/tools/schema/model"
)

const (
	packageJSON  = "package.json"
	tsconfigJSON = "tsconfig.json"
)

type TypeScriptGenerator struct {
	GenBase
}

var _ IGenerator = new(TypeScriptGenerator)

func NewTypeScriptGenerator(s *model.Schema, rootFolder string) *TypeScriptGenerator {
	g := &TypeScriptGenerator{}
	config := tstemplates.Templates[0]
	config["rootFolder"] = rootFolder
	g.init(s, tstemplates.TypeDependent, tstemplates.Templates)
	return g
}

func (g *TypeScriptGenerator) Build() error {
	err := os.MkdirAll("ts/pkg", 0o755)
	if err != nil {
		return err
	}
	wasm := g.s.PackageName + "_ts.wasm"
	fmt.Printf("building %s\n", wasm)
	args := "asc ts/main.ts --lib d:/work/node_modules -O --outFile ts/pkg/" + wasm
	return g.build("npx", args)
}

func (g *TypeScriptGenerator) Cleanup() {
	g.cleanCommonFiles()

	// now clean up language-specific files
	g.cleanSourceFileIfSame(g.createImplIndex)
	g.cleanSourceFileIfSame(g.createImplTsConfig)
	_ = os.Remove(g.folder + "../" + tsconfigJSON)
}

func (g *TypeScriptGenerator) GenerateImplementation() error {
	err := g.generateImplementation()
	if err != nil {
		return err
	}
	err = g.createImplIndex()
	if err != nil {
		return err
	}
	err = g.createImplTsConfig()
	if err != nil {
		return err
	}
	return nil
}

func (g *TypeScriptGenerator) createImplTsConfig() error {
	return g.generateConfig("", tsconfigJSON)
}

func (g *TypeScriptGenerator) createImplIndex() error {
	return g.createSourceFile("index", true, "indexImpl")
}

func (g *TypeScriptGenerator) GenerateInterface() error {
	err := g.generateInterface()
	if err != nil {
		return err
	}
	err = g.createSourceFile("index", true)
	if err != nil {
		return err
	}
	err = g.generateConfig("", tsconfigJSON)
	if err != nil {
		return err
	}
	if !g.s.CoreContracts || g.rootFolder == "ts" {
		err = g.generateConfig("", packageJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *TypeScriptGenerator) GenerateWasmStub() error {
	err := g.createSourceFile("../main", !g.s.CoreContracts)
	if err != nil {
		return err
	}
	err = g.generateConfig("../", tsconfigJSON)
	if err != nil {
		return err
	}
	return nil
}

func (g *TypeScriptGenerator) generateConfig(folder string, name string) error {
	return g.createFile(g.folder+folder+name, false, func() {
		g.emit(name)
	})
}
