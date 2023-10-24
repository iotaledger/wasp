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
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/iotaledger/wasp/tools/schema/generator"
	"github.com/iotaledger/wasp/tools/schema/model"
	schemayaml "github.com/iotaledger/wasp/tools/schema/model/yaml"
)

const version = "schema tool version 1.1.11"

var (
	flagBuild   = flag.Bool("build", false, "build wasm target for specified languages")
	flagClean   = flag.Bool("clean", false, "clean up files that can be re-generated for specified languages")
	flagForce   = flag.Bool("force", false, "force code generation")
	flagGo      = flag.Bool("go", false, "generate Go code")
	flagInit    = flag.String("init", "", "generate new folder with schema file for smart contract named <string>")
	flagRust    = flag.Bool("rs", false, "generate Rust code")
	flagTs      = flag.Bool("ts", false, "generate TypScript code")
	flagVersion = flag.Bool("version", false, "show schema tool version")
)

func init() {
	flag.Parse()
}

func main() {
	err := runGenerator()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func addSubProjectToParentToml() error {
	const cargoTomlPath = "../Cargo.toml"
	_, err := os.Stat(cargoTomlPath)
	if err != nil {
		return nil
	}
	b, err := os.ReadFile(cargoTomlPath)
	if err != nil {
		return err
	}
	content := string(b)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// in case of Windows replace path separators
	cwd = strings.ReplaceAll(cwd, "\\", "/")
	projectName := path.Base(cwd)

	projectFormat := []string{
		"\"%s/rs/%s\"",
		"\"%s/rs/%simpl\"",
		"\"%s/rs/%swasm\"",
	}
	start := strings.Index(content, "members")
	memberContent := strings.Split(content[start:], "]")[0]
	end := start + len(memberContent)
	insertContent := ""

	changes := false
	for _, format := range projectFormat {
		path := fmt.Sprintf(format, projectName, projectName)
		if !strings.Contains(memberContent, path) {
			insertContent += "\t" + path + ",\n"
			changes = true
		}
	}
	if !changes {
		return nil
	}
	finalContent := content[:start] + memberContent + insertContent + content[end:]
	return os.WriteFile(cargoTomlPath, []byte(finalContent), 0o600)
}

func determineSchemaRegenerationTime(file *os.File, s *model.Schema) error {
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
	return nil
}

func generateCoreContractInterfaces() error {
	return filepath.WalkDir("interfaces", func(path string, d fs.DirEntry, err error) error {
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
		return generateSchema(file, true)
	})
}

func generateSchema(file *os.File, core ...bool) error {
	s, err := loadSchema(file)
	if err != nil {
		return err
	}
	s.CoreContracts = len(core) == 1 && core[0]

	err = determineSchemaRegenerationTime(file, s)
	if err != nil {
		return err
	}

	// Preserve line number until here
	// comments are still preserved during generation
	if *flagGo {
		g := generator.NewGoGenerator(s)
		err = generateSchemaFiles(g, s.CoreContracts)
		if err != nil {
			return err
		}
	}

	if *flagRust {
		g := generator.NewRustGenerator(s)
		err = generateSchemaFiles(g, s.CoreContracts)
		if err != nil {
			return err
		}
		// Add current contract to the workspace in the parent folder
		err = addSubProjectToParentToml()
		if err != nil {
			return err
		}
	}

	if *flagTs {
		g := generator.NewTypeScriptGenerator(s, "ts")
		err = generateSchemaFiles(g, s.CoreContracts)
		if err != nil {
			return err
		}
		if s.CoreContracts {
			// Note that we have a separate WasmLib for AssemblyScript.
			// The core contracts are identical except for package.json
			g = generator.NewTypeScriptGenerator(s, "as")
			err = generateSchemaFiles(g, s.CoreContracts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateSchemaFiles(g generator.IGenerator, core bool) error {
	if *flagClean {
		g.Cleanup()
		return nil
	}

	if !g.IsLatest() {
		err := g.GenerateInterface()
		if err != nil {
			return err
		}

		if core {
			return nil
		}

		err = g.GenerateImplementation()
		if err != nil {
			return err
		}

		err = g.GenerateTests()
		if err != nil {
			return err
		}

		err = g.GenerateWasmStub()
		if err != nil {
			return err
		}
	}

	if *flagBuild {
		return g.Build()
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
	schemaDef.Version = model.DefElt{Val: model.DefaultVersion}
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
	err = schemayaml.Unmarshal(fileByteArray, schemaDef)
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

func mustGenerateCoreContractInterfaces() bool {
	// special case when we run in /packages/wasmvm/wasmlib:
	// must generate WasmLib's built-in core contract interfaces
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	cwd = strings.ReplaceAll(cwd, "\\", "/")
	return strings.HasSuffix(cwd, "/packages/wasmvm/wasmlib")
}

func runGenerator() error {
	if *flagVersion {
		fmt.Println(version)
		return nil
	}

	err := generator.FindModulePath()
	if err != nil && *flagGo {
		return err
	}

	if mustGenerateCoreContractInterfaces() {
		if *flagBuild {
			return errors.New("cannot build core contracts")
		}
		return generateCoreContractInterfaces()
	}

	file, err := os.Open("schema.yaml")
	if err == nil {
		defer file.Close()
		if *flagInit != "" {
			return errors.New("schema definition file found")
		}
		return generateSchema(file)
	}

	if *flagInit != "" {
		_, err = os.Stat(strings.ToLower(*flagInit))
		if err == nil {
			return errors.New("contract folder already exists")
		}
		return generateSchemaNew()
	}

	// No schema file in current folder, walk all sub-folders to see
	// if there are schema files and do what's needed there instead.
	generated, err := walkSubFolders()
	if err != nil {
		return err
	}
	if !generated {
		flag.Usage()
		return fmt.Errorf("schema.yaml not found")
	}
	return nil
}

func walkSubFolders() (bool, error) {
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
	return generated, err
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
