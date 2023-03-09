// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/iotaledger/wasp/tools/schema/generator"
	"github.com/iotaledger/wasp/tools/schema/model"
	wasp_yaml "github.com/iotaledger/wasp/tools/schema/model/yaml"
)

const version = "schema version 1.1.0"

var (
	flagBuild   = flag.Bool("build", false, "build wasm target")
	flagClean   = flag.Bool("clean", false, "clean up (re-)generated files")
	flagCore    = flag.Bool("core", false, "generate core contract interface")
	flagForce   = flag.Bool("force", false, "force code generation")
	flagGo      = flag.Bool("go", false, "generate Go code")
	flagInit    = flag.String("init", "", "generate new schema file for smart contract named <string>")
	flagRust    = flag.Bool("rs", false, "generate Rust code")
	flagTs      = flag.Bool("ts", false, "generate TypScript code")
	flagVersion = flag.Bool("version", false, "show schema tool version")
)

func init() {
	flag.Parse()
}

func main() {
	if *flagVersion {
		fmt.Println(version)
		return
	}

	err := generator.FindModulePath()
	if err != nil && *flagGo {
		log.Panic(err)
	}

	if *flagCore {
		generateCoreInterfaces()
		return
	}

	file, err := os.Open("schema.yaml")
	if err == nil {
		defer file.Close()
		if *flagInit != "" {
			log.Panic("schema definition file already exists")
		}
		err = generateSchema(file)
		if err != nil {
			log.Panic(err)
		}
		return
	}

	if *flagInit != "" {
		err = generateSchemaNew()
		if err != nil {
			if _, err2 := os.Stat(*flagInit); err2 == nil {
				log.Println("schema already exists")
				return
			}
			log.Panic(err)
		}
		return
	}

	// No schema file in current folder, walk all subfolders to see if there are
	// schema files and do what's needed there instead.
	if !walkSubFolders() {
		flag.Usage()
	}
}

func generateCoreInterfaces() {
	err := filepath.WalkDir("interfaces", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		return generateSchema(file)
	})
	if err != nil {
		log.Panic(err)
	}
}

func generateSchema(file *os.File) error {
	s, err := loadSchema(file)
	if err != nil {
		return err
	}
	s.CoreContracts = *flagCore

	s.SchemaTime = time.Now()
	if !*flagForce {
		// force regeneration when schema definition file is newer
		info, err2 := file.Stat()
		if err2 != nil {
			return err2
		}
		s.SchemaTime = info.ModTime()

		// also force regeneration when schema tool itself is newer
		exe, err2 := os.Executable()
		if err2 != nil {
			return err2
		}
		info, err2 = os.Stat(exe)
		if err2 != nil {
			return err2
		}
		if info.ModTime().After(s.SchemaTime) {
			s.SchemaTime = info.ModTime()
		}
	}

	// Preserve line number until here
	// comments are still preserved during generation
	if *flagGo {
		g := generator.NewGoGenerator(s)
		err = g.Generate(g, *flagBuild, *flagClean)
		if err != nil {
			return err
		}
	}

	if *flagRust {
		g := generator.NewRustGenerator(s)
		err = g.Generate(g, *flagBuild, *flagClean)
		if err != nil {
			return err
		}
	}

	if *flagTs {
		g := generator.NewTypeScriptGenerator(s, "ts")
		err = g.Generate(g, *flagBuild, *flagClean)
		if err != nil {
			return err
		}
		if s.CoreContracts {
			// note that we have a separate WasmLib for AssemblyScript
			// core contracts are identical except for package.json
			g = generator.NewTypeScriptGenerator(s, "as")
			err = g.Generate(g, false, *flagClean)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateSchemaNew() error {
	r := regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]+$")
	if !r.MatchString(*flagInit) {
		return errors.New("name contains path characters")
	}
	name := *flagInit
	fmt.Println("initializing " + name)

	subfolder := strings.ToLower(name)
	err := os.Mkdir(subfolder, 0o755)
	if err != nil {
		return err
	}
	err = os.Chdir(subfolder)
	if err != nil {
		return err
	}

	schemaDef := &model.SchemaDef{}
	schemaDef.Name = model.DefElt{Val: name}
	schemaDef.Description = model.DefElt{Val: name + " description"}
	schemaDef.Author = model.DefElt{Val: "Eric Hop <eric@iota.org>"}
	schemaDef.Structs = make(model.DefMapMap)
	schemaDef.Events = make(model.DefMapMap)
	schemaDef.Typedefs = make(model.DefMap)
	schemaDef.State = make(model.DefMap)

	defMapKey := model.DefElt{Val: "owner"}
	schemaDef.State[defMapKey] = &model.DefElt{Val: "AgentID // current owner of this smart contract"}
	schemaDef.Funcs = make(model.FuncDefMap)
	schemaDef.Views = make(model.FuncDefMap)

	funcInit := &model.FuncDef{}
	funcInit.Params = make(model.DefMap)
	funcInit.Params[defMapKey] = &model.DefElt{Val: "AgentID? // optional owner of this smart contract"}
	schemaDef.Funcs[model.DefElt{Val: "init"}] = funcInit

	funcSetOwner := &model.FuncDef{}
	funcSetOwner.Access = model.DefElt{Val: "owner // current owner of this smart contract"}
	funcSetOwner.Params = make(model.DefMap)
	funcSetOwner.Params[defMapKey] = &model.DefElt{Val: "AgentID // new owner of this smart contract"}
	schemaDef.Funcs[model.DefElt{Val: "setOwner"}] = funcSetOwner

	viewGetOwner := &model.FuncDef{}
	viewGetOwner.Results = make(model.DefMap)
	viewGetOwner.Results[defMapKey] = &model.DefElt{Val: "AgentID // current owner of this smart contract"}
	schemaDef.Views[model.DefElt{Val: "getOwner"}] = viewGetOwner
	return WriteYAMLSchema(schemaDef)
}

func loadSchema(file *os.File) (s *model.Schema, err error) {
	name := file.Name()
	if name == "schema.yaml" {
		cwd, _ := os.Getwd()
		folder := filepath.Base(cwd)
		name = folder + "/" + name
	}
	fmt.Println("loading " + name)
	fileByteArray, _ := io.ReadAll(file)
	schemaDef := model.NewSchemaDef()
	err = wasp_yaml.Unmarshal(fileByteArray, schemaDef)
	if err != nil {
		return nil, err
	}

	s = model.NewSchema()
	err = s.Compile(schemaDef)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func walkSubFolders() bool {
	generated := false
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != "schema.yaml" {
			return nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			log.Panic(err)
		}
		defer func() { _ = os.Chdir(cwd) }()

		sub := path[0 : len(path)-len(d.Name())-1]
		err = os.Chdir(sub)
		if err != nil {
			return err
		}
		file, err := os.Open(d.Name())
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		generated = true
		err = generator.FindModulePath()
		if err != nil {
			return err
		}
		return generateSchema(file)
	})
	if err != nil {
		log.Panic(err)
	}
	return generated
}

func WriteYAMLSchema(schemaDef *model.SchemaDef) error {
	file, err := os.Create("schema.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := yaml.Marshal(schemaDef.ToRawSchemaDef())
	if err != nil {
		return err
	}

	b = bytes.ReplaceAll(b, []byte("//"), []byte("#"))
	_, err = file.Write(b)
	return err
}
