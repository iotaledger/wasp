// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
)

// TODO nested structs
// TODO handle case where owner is type AgentID[]

const (
	AccessChain   = "chain"
	AccessCreator = "creator"
	AccessSelf    = "self"
	InitFunc      = "Init"
	KindFunc      = "Func"
	KindView      = "View"
)

var (
	ModuleCwd  = "???"
	ModuleName = "???"
	ModulePath = "???"
)

type Generator interface {
	init(s *Schema)
	funcName(f *Func) string
	generateLanguageSpecificFiles() error
	setFieldKeys()
	setFuncKeys()
}

type GenBase struct {
	currentField  *Field
	currentFunc   *Func
	currentStruct *Struct
	emitters      map[string]func(g *GenBase)
	extension     string
	file          *os.File
	folder        string
	funcRegexp    *regexp.Regexp
	gen           Generator
	keys          map[string]string
	language      string
	newTypes      map[string]bool
	rootFolder    string
	s             *Schema
	skipWarning   bool
	templates     map[string]string
}

func (g *GenBase) init(s *Schema, templates []map[string]string) {
	g.s = s
	g.emitters = map[string]func(g *GenBase){}
	g.newTypes = map[string]bool{}
	g.keys = map[string]string{}
	g.setCommonKeys()
	g.templates = map[string]string{}
	g.addTemplates(commonTemplates)
	for _, template := range templates {
		g.addTemplates(template)
	}
}

func (g *GenBase) addTemplates(t map[string]string) {
	for k, v := range t {
		g.templates[k] = v
	}
}

func (g *GenBase) close() {
	_ = g.file.Close()
}

func (g *GenBase) create(path string) (err error) {
	g.file, err = os.Create(path)
	return err
}

func (g *GenBase) createSourceFile(name string, generator ...func()) error {
	err := g.create(g.folder + name + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// TODO take copyright from schema?

	// always add copyright to source file
	g.emit("copyright")
	if !g.skipWarning {
		g.emit("warning")
	}
	g.skipWarning = false
	g.s.KeyID = 0
	if len(generator) != 0 {
		generator[0]()
		return nil
	}
	g.emit(name + g.extension)
	return nil
}

func (g *GenBase) exists(path string) (err error) {
	_, err = os.Stat(path)
	return err
}

func (g *GenBase) Generate(s *Schema) error {
	g.gen.init(s)

	g.folder = g.rootFolder + "/"
	if g.rootFolder != "src" {
		module := strings.ReplaceAll(ModuleCwd, "\\", "/")
		module = module[strings.LastIndex(module, "/")+1:]
		g.folder += module + "/"
	}
	if g.s.CoreContracts {
		g.folder += g.s.Name + "/"
	}

	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.folder + "consts" + g.extension)
	if err == nil && info.ModTime().After(s.SchemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return nil
	}

	fmt.Printf("generating %s code\n", g.language)
	err = g.generateCode()
	if err != nil {
		return err
	}
	if !g.s.CoreContracts {
		err = g.generateTests()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GenBase) generateCode() error {
	err := g.createSourceFile("consts")
	if err != nil {
		return err
	}
	if len(g.s.Structs) != 0 {
		err = g.createSourceFile("structs")
		if err != nil {
			return err
		}
	}
	if len(g.s.Typedefs) != 0 {
		err = g.createSourceFile("typedefs")
		if err != nil {
			return err
		}
	}
	if len(g.s.Params) != 0 {
		err = g.createSourceFile("params")
		if err != nil {
			return err
		}
	}
	if len(g.s.Results) != 0 {
		err = g.createSourceFile("results")
		if err != nil {
			return err
		}
	}
	err = g.createSourceFile("contract")
	if err != nil {
		return err
	}

	if !g.s.CoreContracts {
		err = g.createSourceFile("keys")
		if err != nil {
			return err
		}
		err = g.createSourceFile("state")
		if err != nil {
			return err
		}
		err = g.createSourceFile("lib")
		if err != nil {
			return err
		}
		err = g.generateFuncs()
		if err != nil {
			return err
		}
	}

	return g.gen.generateLanguageSpecificFiles()
}

func (g *GenBase) generateFuncs() error {
	scFileName := g.folder + g.s.Name + g.extension
	err := g.open(g.folder + g.s.Name + g.extension)
	if err != nil {
		// generate initial SC function file
		g.skipWarning = true
		return g.createSourceFile(g.s.Name, func() {
			g.emit("funcs" + g.extension)
		})
	}

	// append missing function signatures to existing code file

	// scan existing file for signatures
	lines, existing, err := g.scanExistingCode()
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := g.folder + g.s.Name + ".bak"
	err = os.Rename(scFileName, scOriginal)
	if err != nil {
		return err
	}
	err = g.create(scFileName)
	if err != nil {
		return err
	}
	defer g.close()

	// make copy of file
	for _, line := range lines {
		g.println(line)
	}

	// append any new funcs
	for _, g.currentFunc = range g.s.Funcs {
		if existing[g.gen.funcName(g.currentFunc)] == "" {
			g.setFuncKeys()
			g.emit("funcSignature")
		}
	}

	return os.Remove(scOriginal)
}

func (g *GenBase) generateTests() error {
	err := os.MkdirAll("test", 0o755)
	if err != nil {
		return err
	}

	// do not overwrite existing file
	name := strings.ToLower(g.s.Name)
	filename := "test/" + name + "_test.go"
	err = g.exists(filename)
	if err == nil {
		return nil
	}

	err = g.create(filename)
	if err != nil {
		return err
	}
	defer g.close()

	g.emit("test.go")
	return nil
}

func (g *GenBase) open(path string) (err error) {
	g.file, err = os.Open(path)
	return err
}

func (g *GenBase) println(a ...interface{}) {
	_, _ = fmt.Fprintln(g.file, a...)
}

func (g *GenBase) scanExistingCode() ([]string, StringMap, error) {
	defer g.close()
	existing := make(StringMap)
	lines := make([]string, 0)
	scanner := bufio.NewScanner(g.file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := g.funcRegexp.FindStringSubmatch(line)
		if matches != nil {
			existing[matches[1]] = line
		}
		lines = append(lines, line)
	}
	err := scanner.Err()
	if err != nil {
		return nil, nil, err
	}
	return lines, existing, nil
}

func (g *GenBase) setCommonKeys() {
	g.keys["space"] = " "
	g.keys["package"] = g.s.Name
	g.keys["Package"] = g.s.FullName
	g.keys["module"] = ModuleName + strings.Replace(ModuleCwd[len(ModulePath):], "\\", "/", -1)
	scName := g.s.Name
	if g.s.CoreContracts {
		// strip off "core" prefix
		scName = scName[4:]
	}
	g.keys["scName"] = scName
	g.keys["hscName"] = iscp.Hn(scName).String()
	g.keys["scDesc"] = g.s.Description
}

func (g *GenBase) setFieldKeys() {
	g.setMultiKeyValues("fldName", g.currentField.Name)
	g.setMultiKeyValues("fldType", g.currentField.Type)

	g.keys["fldAlias"] = g.currentField.Alias
	g.keys["FldComment"] = g.currentField.Comment
	g.keys["FldMapKey"] = g.currentField.MapKey

	g.keys["fldIndex"] = strconv.Itoa(g.s.KeyID)
	g.s.KeyID++
	g.keys["maxIndex"] = strconv.Itoa(g.s.KeyID)
}

func (g *GenBase) setFuncKeys() {
	g.setMultiKeyValues("funcName", g.currentFunc.FuncName[4:])
	g.setMultiKeyValues("kind", g.currentFunc.FuncName[:4])
	g.keys["funcHName"] = iscp.Hn(g.keys["funcName"]).String()
	grant := g.currentFunc.Access
	comment := ""
	index := strings.Index(grant, "//")
	if index >= 0 {
		comment = fmt.Sprintf("    %s\n", grant[index:])
		grant = strings.TrimSpace(grant[:index])
	}
	g.keys["funcAccess"] = grant
	g.keys["funcAccessComment"] = comment
}

func (g *GenBase) setMultiKeyValues(key, value string) {
	value = uncapitalize(value)
	g.keys[key] = value
	g.keys[capitalize(key)] = capitalize(value)
	g.keys[snake(key)] = snake(value)
	g.keys[upper(snake(key))] = upper(snake(value))
}
