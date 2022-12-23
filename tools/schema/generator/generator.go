// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/tools/schema/model"
)

// TODO nested structs
// TODO handle case where owner is type AgentID[]

type IGenerator interface {
	Cleanup()
	GenerateImplementation() error
	GenerateInterface() error
	GenerateWasmStub() error
	IsLatest() bool
}

type GenBase struct {
	currentEvent  *model.Struct
	currentField  *model.Field
	currentFunc   *model.Func
	currentStruct *model.Struct
	emitters      map[string]func(g *GenBase)
	extension     string
	file          *os.File
	folder        string
	funcRegexp    *regexp.Regexp
	keys          model.StringMap
	language      string
	newTypes      map[string]bool
	rootFolder    string
	s             *model.Schema
	subFolder     string
	tab           int
	templates     model.StringMap
	typeDependent model.StringMapMap
}

const spaces = "                                             "

func (g *GenBase) init(s *model.Schema, typeDependent model.StringMapMap, templates []map[string]string) {
	g.s = s
	g.typeDependent = typeDependent

	g.emitters = map[string]func(g *GenBase){}
	g.keys = model.StringMap{}
	g.newTypes = map[string]bool{}
	g.templates = model.StringMap{}

	config := templates[0]
	g.language = config["language"]
	g.extension = config["extension"]
	g.rootFolder = config["rootFolder"]
	g.subFolder = config["subFolder"]
	g.funcRegexp = regexp.MustCompile(config["funcRegexp"])

	g.addTemplates(commonTemplates)
	for _, template := range templates {
		g.addTemplates(template)
	}

	g.setCommonKeys()
}

func (g *GenBase) addTemplates(t model.StringMap) {
	for k, v := range t {
		g.templates[k] = v
	}
}

func (g *GenBase) cleanCommonFiles() {
	g.generateCommonFolder("", false)
	g.cleanFolder(g.folder)

	g.generateCommonFolder("impl", true)
	g.cleanSourceFile("thunks")
	g.cleanSourceFile("../main")
	g.cleanFolder(g.folder + "../pkg")
}

func (g *GenBase) cleanFolder(folder string) {
	_ = os.RemoveAll(folder)
	_ = os.Remove(folder)
}

func (g *GenBase) cleanSourceFile(name string) {
	path := g.folder + name + g.extension
	_ = os.Remove(path)
}

func (g *GenBase) close() {
	_ = g.file.Close()
}

func (g *GenBase) createFile(path string, overwrite bool, generator func()) (err error) {
	if !overwrite && g.exists(path) == nil {
		return nil
	}
	g.file, err = os.Create(path)
	if err != nil {
		return err
	}
	defer g.close()
	generator()
	return nil
}

func (g *GenBase) createSourceFile(name string, condition bool, macro ...string) error {
	path := g.folder + name + g.extension
	if !condition {
		_ = os.Remove(path)
		return nil
	}
	if len(macro) == 1 {
		name = macro[0]
	}
	return g.createFile(path, true, func() {
		g.emit("copyright")
		g.emit("warning")
		g.emit(name + g.extension)
	})
}

func (g *GenBase) error(what string) {
	g.println("???:" + what)
}

func (g *GenBase) exists(path string) (err error) {
	_, err = os.Stat(path)
	return err
}

func (g *GenBase) funcName(f *model.Func) string {
	name := f.Kind + capitalize(f.Name)
	if g.language == "Rust" {
		name = snake(name)
	}
	return name
}

func (g *GenBase) IsLatest() bool {
	g.generateCommonFolder("", true)

	info, err := os.Stat(g.folder + "consts" + g.extension)
	if err == nil && info.ModTime().After(g.s.SchemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return true
	}

	fmt.Printf("generating %s code\n", g.language)
	return false
}

func (g *GenBase) Generate(gen IGenerator, clean bool) error {
	if clean {
		gen.Cleanup()
		return nil
	}

	if g.IsLatest() {
		return nil
	}

	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	err = gen.GenerateInterface()
	if err != nil {
		return err
	}

	if g.s.CoreContracts {
		return nil
	}

	g.generateCommonFolder("impl", true)
	err = os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	err = gen.GenerateImplementation()
	if err != nil {
		return err
	}

	err = g.generateTests()
	if err != nil {
		return err
	}

	return gen.GenerateWasmStub()
}

func (g *GenBase) generateCommonFolder(postfix string, withSubFolder bool) {
	g.folder = g.rootFolder + "/" + g.s.PackageName + postfix + "/"
	if g.s.CoreContracts {
		g.folder = g.rootFolder + "/wasmlib/" + g.s.PackageName + "/"
	}
	if withSubFolder && g.subFolder != "" {
		g.folder += g.subFolder + "/"
		if g.s.CoreContracts {
			g.folder = g.subFolder + "/" + g.s.PackageName + "/"
		}
	}
}

func (g *GenBase) generateImplementation() error {
	err := g.createSourceFile("thunks", true)
	if err != nil {
		return err
	}
	return g.generateFuncs(g.appendFuncs)
}

func (g *GenBase) generateInterface() error {
	err := g.createSourceFile("consts", true)
	if err != nil {
		return err
	}
	err = g.createSourceFile("events", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("eventhandlers", len(g.s.Events) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("structs", len(g.s.Structs) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("typedefs", len(g.s.Typedefs) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("params", len(g.s.Params) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("results", len(g.s.Results) != 0)
	if err != nil {
		return err
	}
	err = g.createSourceFile("state", !g.s.CoreContracts && len(g.s.StateVars) != 0)
	if err != nil {
		return err
	}
	return g.createSourceFile("contract", true)
}

func (g *GenBase) generateFuncs(appendFuncs func(existing model.StringMap)) error {
	scFileName := g.folder + "funcs" + g.extension
	if g.exists(scFileName) != nil {
		// generate initial SC function file
		return g.createFile(scFileName, false, func() {
			g.emit("copyright")
			g.emit("funcs" + g.extension)
		})
	}

	// append missing SC functions to existing code file

	// scan existing file for function names
	existing := make(model.StringMap)
	lines := make([]string, 0)
	err := g.scanExistingCode(scFileName, &existing, &lines)
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := g.folder + "funcs.bak"
	err = os.Rename(scFileName, scOriginal)
	if err != nil {
		return err
	}

	err = g.createFile(scFileName, false, func() {
		// make copy of original file
		for _, line := range lines {
			g.println(line)
		}

		// append any new funcs
		appendFuncs(existing)
	})
	if err != nil {
		return err
	}
	return os.Remove(scOriginal)
}

func (g *GenBase) appendFuncs(existing model.StringMap) {
	for _, g.currentFunc = range g.s.Funcs {
		if existing[g.funcName(g.currentFunc)] == "" {
			g.setFuncKeys(false, 0, 0)
			g.emit("funcSignature")
		}
	}
}

func (g *GenBase) generateTests() error {
	err := os.MkdirAll("test", 0o755)
	if err != nil {
		return err
	}

	// do not overwrite existing file
	name := strings.ToLower(g.s.PackageName)
	filename := "test/" + name + "_test.go"
	return g.createFile(filename, false, func() {
		g.emit("test.go")
	})
}

func (g *GenBase) openFile(path string, processor func() error) (err error) {
	g.file, err = os.Open(path)
	if err != nil {
		return err
	}
	defer g.close()
	return processor()
}

func (g *GenBase) println(a ...interface{}) {
	_, _ = fmt.Fprintln(g.file, a...)
}

func (g *GenBase) scanExistingCode(path string, existing *model.StringMap, lines *[]string) error {
	return g.openFile(path, func() error {
		scanner := bufio.NewScanner(g.file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := g.funcRegexp.FindStringSubmatch(line)
			if matches != nil {
				(*existing)[matches[1]] = line
			}
			*lines = append(*lines, line)
		}
		return scanner.Err()
	})
}
