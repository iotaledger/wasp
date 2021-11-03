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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iotaledger/wasp/tools/schema/generator"
	"gopkg.in/yaml.v2"
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
		fmt.Println(err)
		return
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
			fmt.Println("schema definition file already exists")
			return
		}
		err = generateSchema(file)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if *flagInit != "" {
		err = generateSchemaNew()
		if err != nil {
			fmt.Println(err)
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
		fmt.Println(err)
	}
}

func generateSchema(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	s, err := loadSchema(file)
	if err != nil {
		return err
	}
	s.CoreContracts = *flagCore

	s.SchemaTime = info.ModTime()
	if *flagForce {
		// make as if it has just been updated
		s.SchemaTime = time.Now()
	}

	if *flagTs {
		g := generator.NewTypeScriptGenerator()
		err = g.Generate(s)
		if err != nil {
			return err
		}
	}

	if *flagGo {
		g := generator.NewGoGenerator()
		err = g.Generate(s)
		if err != nil {
			return err
		}
	}

	if *flagRust {
		g := generator.NewRustGenerator()
		err = g.Generate(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateSchemaNew() error {
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

	schemaDef := &generator.SchemaDef{}
	schemaDef.Name = name
	schemaDef.Description = name + " description"
	schemaDef.Structs = make(generator.StringMapMap)
	schemaDef.Typedefs = make(generator.StringMap)
	schemaDef.State = make(generator.StringMap)
	schemaDef.State["owner"] = "AgentID // current owner of this smart contract"
	schemaDef.Funcs = make(generator.FuncDefMap)
	schemaDef.Views = make(generator.FuncDefMap)

	funcInit := &generator.FuncDef{}
	funcInit.Params = make(generator.StringMap)
	funcInit.Params["owner"] = "AgentID? // optional owner of this smart contract"
	schemaDef.Funcs["init"] = funcInit

	funcSetOwner := &generator.FuncDef{}
	funcSetOwner.Access = "owner // current owner of this smart contract"
	funcSetOwner.Params = make(generator.StringMap)
	funcSetOwner.Params["owner"] = "AgentID // new owner of this smart contract"
	schemaDef.Funcs["setOwner"] = funcSetOwner

	viewGetOwner := &generator.FuncDef{}
	viewGetOwner.Results = make(generator.StringMap)
	viewGetOwner.Results["owner"] = "AgentID // current owner of this smart contract"
	schemaDef.Views["getOwner"] = viewGetOwner
	switch *flagType {
	case "json":
		return WriteJSONSchema(schemaDef)
	case "yaml":
		return WriteYAMLSchema(schemaDef)
	}
	return errors.New("invalid schema type: " + *flagType)
}

func loadSchema(file *os.File) (s *generator.Schema, err error) {
	fmt.Println("loading " + file.Name())
	schemaDef := &generator.SchemaDef{}
	switch filepath.Ext(file.Name()) {
	case ".json":
		err = json.NewDecoder(file).Decode(schemaDef)
		if err == nil && *flagType == "convert" {
			err = WriteYAMLSchema(schemaDef)
		}
	case ".yaml":
		fileByteArray, _ := io.ReadAll(file)
		err = yaml.Unmarshal(fileByteArray, schemaDef)
		if err == nil && *flagType == "convert" {
			err = WriteJSONSchema(schemaDef)
		}
	default:
		err = errors.New("unexpected file type: " + file.Name())
	}
	if err != nil {
		return nil, err
	}

	s = generator.NewSchema()
	err = s.Compile(schemaDef)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func WriteJSONSchema(schemaDef *generator.SchemaDef) error {
	file, err := os.Create("schema.json")
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.Marshal(schemaDef)
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

func WriteYAMLSchema(schemaDef *generator.SchemaDef) error {
	file, err := os.Create("schema.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := yaml.Marshal(schemaDef)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	return err
}
