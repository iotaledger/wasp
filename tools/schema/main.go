// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
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

var (
	flagCore  = flag.Bool("core", false, "generate core contract interface")
	flagForce = flag.Bool("force", false, "force code generation")
	flagGo    = flag.Bool("go", false, "generate Go code")
	flagInit  = flag.String("init", "", "generate new schema file for smart contract named <string>")
	flagRust  = flag.Bool("rust", false, "generate Rust code")
	flagTs    = flag.Bool("ts", false, "generate TypScript code")
	flagType  = flag.String("type", "yaml", "type of schema file that will be generated. Values(yaml,json)")
)

func init() {
	flag.Parse()
}

func main() {
	err := generator.FindModulePath()
	if err != nil && *flagGo {
		log.Panic(err)
	}

	if *flagCore {
		generateCoreInterfaces()
		return
	}

	file, err := os.Open("schema.yaml")
	if err != nil {
		file, err = os.Open("schema.json")
	}
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
			if _, err := os.Stat(*flagInit); err == nil {
				log.Println("schema already exists")
				return
			}
			log.Panic(err)
		}
		return
	}

	flag.Usage()
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
		info, err := file.Stat()
		if err != nil {
			return err
		}
		s.SchemaTime = info.ModTime()

		// also force regeneration when schema tool itself is newer
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		info, err = os.Stat(exe)
		if err != nil {
			return err
		}
		if info.ModTime().After(s.SchemaTime) {
			s.SchemaTime = info.ModTime()
		}
	}

	// Preserve line number until here
	// comments are still preserved during generation
	if *flagGo {
		g := generator.NewGoGenerator(s)
		err = g.Generate()
		if err != nil {
			return err
		}
	}

	if *flagRust {
		g := generator.NewRustGenerator(s)
		err = g.Generate()
		if err != nil {
			return err
		}
	}

	if *flagTs {
		g := generator.NewTypeScriptGenerator(s)
		err = g.Generate()
		if err != nil {
			return err
		}
	}
	return nil
}

func generateSchemaNew() error {
	r := regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]+$")
	if !r.MatchString(*flagInit) {
		return fmt.Errorf("name contains path characters")
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
	switch *flagType {
	case "json":
		return WriteJSONSchema(schemaDef)
	case "yaml":
		return WriteYAMLSchema(schemaDef)
	}
	return errors.New("invalid schema type: " + *flagType)
}

func loadSchema(file *os.File) (s *model.Schema, err error) {
	fmt.Println("loading " + file.Name())
	schemaDef := &model.SchemaDef{}
	switch filepath.Ext(file.Name()) {
	case ".json":
		var jsonSchemaDef model.JSONSchemaDef
		err = json.NewDecoder(file).Decode(&jsonSchemaDef)
		schemaDef = jsonSchemaDef.ToSchemaDef()
		if err == nil && *flagType == "convert" {
			err = WriteYAMLSchema(schemaDef)
		}
	case ".yaml":
		fileByteArray, _ := io.ReadAll(file)
		schemaDef = model.NewSchemaDef()
		err = wasp_yaml.Unmarshal(fileByteArray, schemaDef)
		if err == nil && *flagType == "convert" {
			err = WriteJSONSchema(schemaDef)
		}
	default:
		err = errors.New("unexpected file type: " + file.Name())
	}
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

func WriteJSONSchema(schemaDef *model.SchemaDef) error {
	file, err := os.Create("schema.json")
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.Marshal(schemaDef.ToRawSchemaDef())
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	_, err = out.WriteTo(file)
	return err
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
