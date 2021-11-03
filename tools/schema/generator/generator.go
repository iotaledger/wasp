// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
)

// TODO nested structs
// TODO handle case where owner is type AgentID[]

const (
	AccessChain         = "chain"
	AccessCreator       = "creator"
	AccessSelf          = "self"
	AliasThis           = "this"
	constPrefix         = "constPrefix"
	InitFunc            = "Init"
	KeyCore             = "core"
	KeyFunc             = "func"
	KeyParam            = "param"
	KeyParams           = "params"
	KeyPtrs             = "ptrs"
	KeyResult           = "result"
	KeyResults          = "results"
	KeyState            = "state"
	KeyStruct           = "struct"
	KeyTypeDef          = "typeDef"
	KeyView             = "view"
	KindFunc            = "Func"
	KindView            = "View"
	PrefixParam         = "Param"
	PrefixResult        = "Result"
	PrefixState         = "State"
	PropImmutable       = "Immutable"
	PropMutable         = "Mutable"
	SpecialFuncInit     = "funcInit"
	SpecialFuncSetOwner = "setOwner"
	SpecialViewGetOwner = "getOwner"
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
	generateProxyArray(field *Field, mutability, arrayType, proxyType string)
	generateProxyMap(field *Field, mutability, mapType, proxyType string)
	generateProxyReference(field *Field, mutability, typeName string)
	writeConsts()
	writeContract()
	writeInitialFuncs()
	writeKeys()
	writeLib()
	writeParams()
	writeResults()
	writeState()
	writeStructs()
	writeTypeDefs()
}

type GenBase struct {
	currentField  *Field
	currentFunc   *Func
	currentStruct *Struct
	emitters      map[string]func(g *GenBase) string
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

func (g *GenBase) init(s *Schema) {
	g.s = s
	g.emitters = map[string]func(g *GenBase) string{}
	g.newTypes = map[string]bool{}
	g.keys = map[string]string{}
	g.templates = map[string]string{}
	g.addTemplates(templates)
	g.setKeys()
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

func (g *GenBase) createSourceFile(name string, generator func()) error {
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
	generator()
	return nil
}

var (
	emitKeyRegExp = regexp.MustCompile(`\$[a-zA-Z_]+`)
	emitCmdRegExp = regexp.MustCompile(`\$#[^\n]*\n`)
)

func (g *GenBase) emit(template string) {
	g.printf(g.emitAll(template))
}

func (g *GenBase) emitAll(template string) string {
	template = g.templates[template]

	// first process special commands
	text := emitCmdRegExp.ReplaceAllStringFunc(template, func(key string) string {
		if strings.HasPrefix(key, "$#each ") {
			return g.emitEach(strings.TrimSpace(key[7:]))
		}
		if strings.HasPrefix(key, "$#emit ") {
			return g.emitAll(strings.TrimSpace(key[7:]))
		}
		if strings.HasPrefix(key, "$#func ") {
			return g.emitFunc(strings.TrimSpace(key[7:]))
		}
		if strings.HasPrefix(key, "$#if ") {
			return g.emitIf(strings.TrimSpace(key[5:]))
		}
		return "???:" + key
	})

	// then replace any remaining keys
	text = emitKeyRegExp.ReplaceAllStringFunc(text, func(key string) string {
		return g.emitKey(key)
	})

	// remove concatenation markers
	text = strings.ReplaceAll(text, "$+", "")

	// remove leading newline
	return strings.TrimPrefix(text, "\n")
}

func (g *GenBase) emitEach(key string) string {
	parts := strings.Split(key, " ")
	if len(parts) != 2 {
		return "???: " + key
	}

	template := parts[1]
	switch parts[0] {
	case "func":
		text := ""
		for _, g.currentFunc = range g.s.Funcs {
			g.setFuncKeys()
			text += g.emitAll(template)
		}
		return text
	case "mandatory":
		mandatory := []*Field{}
		for _, g.currentField = range g.currentFunc.Params {
			if !g.currentField.Optional {
				mandatory = append(mandatory, g.currentField)
			}
		}
		return g.emitFields(mandatory, template)
	case KeyParam:
		return g.emitFields(g.currentFunc.Params, template)
	case KeyParams:
		g.keys[constPrefix] = PrefixParam
		return g.emitFields(g.s.Params, template)
	case KeyResult:
		return g.emitFields(g.currentFunc.Results, template)
	case KeyResults:
		g.keys[constPrefix] = PrefixResult
		return g.emitFields(g.s.Results, template)
	case KeyState:
		g.keys[constPrefix] = PrefixState
		return g.emitFields(g.s.StateVars, template)
	case KeyStruct:
		text := ""
		for _, g.currentStruct = range g.s.Structs {
			g.setStructKeys()
			text += g.emitAll(template)
		}
		return text
	case KeyTypeDef:
		return g.emitFields(g.s.Typedefs, template)
	}
	return "???: " + key
}

func (g *GenBase) emitFields(fields []*Field, template string) string {
	text := ""
	for _, g.currentField = range fields {
		if g.currentField.Alias == "this" {
			continue
		}
		g.setFieldKeys()
		text += g.emitAll(template)
	}
	return text
}

func (g *GenBase) emitFunc(key string) string {
	f, ok := g.emitters[key]
	if ok {
		return f(g)
	}
	return "???: " + key
}

func (g *GenBase) emitIf(key string) string {
	parts := strings.Split(key, " ")
	if len(parts) < 2 || len(parts) > 3 {
		return "???: " + key
	}

	conditionKey := parts[0]
	template := parts[1]

	condition := false
	switch conditionKey {
	case KeyCore:
		condition = g.s.CoreContracts
	case KeyFunc, KeyView:
		condition = g.keys["kind"] == conditionKey
	case KeyParam:
		condition = len(g.currentFunc.Params) != 0
	case KeyParams:
		condition = len(g.s.Params) != 0
	case KeyResult:
		condition = len(g.currentFunc.Results) != 0
	case KeyResults:
		condition = len(g.s.Results) != 0
	case KeyState:
		condition = len(g.s.StateVars) != 0
	case KeyPtrs:
		condition = len(g.currentFunc.Params) != 0 || len(g.currentFunc.Results) != 0
	default:
		return "???: " + key
	}

	if condition {
		return g.emitAll(template)
	}
	if len(parts) == 3 {
		template = parts[2]
		return g.emitAll(template)
	}
	return ""
}

func (g *GenBase) emitKey(key string) string {
	text, ok := g.keys[key[1:]]
	if ok {
		return text
	}
	return "???: " + key
}

func (g *GenBase) exists(path string) (err error) {
	_, err = os.Stat(path)
	return err
}

func (g *GenBase) formatter(on bool) {
	if on {
		g.printf("\n// @formatter:%s\n", "on")
		return
	}
	g.printf("// @formatter:%s\n\n", "off")
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
	err := g.createSourceFile("consts", g.gen.writeConsts)
	if err != nil {
		return err
	}
	if len(g.s.Structs) != 0 {
		err = g.createSourceFile("structs", g.gen.writeStructs)
		if err != nil {
			return err
		}
	}
	if len(g.s.Typedefs) != 0 {
		err = g.createSourceFile("typedefs", g.gen.writeTypeDefs)
		if err != nil {
			return err
		}
	}
	if len(g.s.Params) != 0 {
		err = g.createSourceFile("params", g.gen.writeParams)
		if err != nil {
			return err
		}
	}
	if len(g.s.Results) != 0 {
		err = g.createSourceFile("results", g.gen.writeResults)
		if err != nil {
			return err
		}
	}
	err = g.createSourceFile("contract", g.gen.writeContract)
	if err != nil {
		return err
	}

	if !g.s.CoreContracts {
		err = g.createSourceFile("keys", g.gen.writeKeys)
		if err != nil {
			return err
		}
		err = g.createSourceFile("state", g.gen.writeState)
		if err != nil {
			return err
		}
		err = g.createSourceFile("lib", g.gen.writeLib)
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
		return g.createSourceFile(g.s.Name, g.gen.writeInitialFuncs)
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

func (g *GenBase) generateProxy(field *Field, mutability string) {
	if field.Array {
		proxyType := mutability + field.Type
		arrayType := "ArrayOf" + proxyType
		if !g.newTypes[arrayType] {
			g.newTypes[arrayType] = true
			g.gen.generateProxyArray(field, mutability, arrayType, proxyType)
		}
		g.gen.generateProxyReference(field, mutability, arrayType)
		return
	}

	if field.MapKey != "" {
		proxyType := mutability + field.Type
		mapType := "Map" + field.MapKey + "To" + proxyType
		if !g.newTypes[mapType] {
			g.newTypes[mapType] = true
			g.gen.generateProxyMap(field, mutability, mapType, proxyType)
		}
		g.gen.generateProxyReference(field, mutability, mapType)
	}
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

	module := ModuleName + strings.ReplaceAll(ModuleCwd[len(ModulePath):], "\\", "/")
	g.println("package test")
	g.println()
	g.println("import (")
	g.println("\t\"testing\"")
	g.println()
	g.printf("\t\"%s/go/%s\"\n", module, g.s.Name)
	g.println("\t\"github.com/iotaledger/wasp/packages/vm/wasmsolo\"")
	g.println("\t\"github.com/stretchr/testify/require\"")
	g.println(")")
	g.println()
	g.println("func TestDeploy(t *testing.T) {")
	g.printf("\tctx := wasmsolo.NewSoloContext(t, %s.ScName, %s.OnLoad)\n", name, name)
	g.printf("\trequire.NoError(t, ctx.ContractExists(%s.ScName))\n", name)
	g.println("}")

	return nil
}

func (g *GenBase) open(path string) (err error) {
	g.file, err = os.Open(path)
	return err
}

func (g *GenBase) printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(g.file, format, a...)
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

func (g *GenBase) setFuncKeys() {
	funcName := uncapitalize(g.currentFunc.FuncName[4:])
	g.keys["funcName"] = funcName
	g.keys["FuncName"] = capitalize(funcName)
	g.keys["func_name"] = snake(funcName)
	g.keys["FUNC_NAME"] = upper(snake(funcName))
	g.keys["funcHName"] = iscp.Hn(funcName).String()

	kind := lower(g.currentFunc.FuncName[:4])
	g.keys["kind"] = kind
	g.keys["Kind"] = capitalize(kind)
	g.keys["KIND"] = upper(kind)

	paramsID := "nil"
	if len(g.currentFunc.Params) != 0 {
		paramsID = "&f.Params.id"
	}
	g.keys["paramsID"] = paramsID

	resultsID := "nil"
	if len(g.currentFunc.Results) != 0 {
		resultsID = "&f.Results.id"
	}
	g.keys["resultsID"] = resultsID

	initFunc := ""
	initMap := ""
	if g.currentFunc.Type == InitFunc {
		initFunc = InitFunc
		initMap = ", keyMap[:], idxMap[:]"
	}
	g.keys["initFunc"] = initFunc
	g.keys["initMap"] = initMap
}

func (g *GenBase) setFieldKeys() {
	fldName := uncapitalize(g.currentField.Name)
	g.keys["fldName"] = fldName
	g.keys["FldName"] = capitalize(fldName)
	g.keys["fld_name"] = snake(fldName)
	g.keys["FLD_NAME"] = upper(snake(fldName))
	g.keys["fldAlias"] = g.currentField.Alias
}

func (g *GenBase) setKeys() {
	g.keys["package"] = g.s.Name
	g.keys["Package"] = g.s.FullName
	scName := g.s.Name
	if g.s.CoreContracts {
		// strip off "core" prefix
		scName = scName[4:]
	}
	g.keys["scName"] = scName
	g.keys["hscName"] = iscp.Hn(scName).String()
	g.keys["scDesc"] = g.s.Description
}

func (g *GenBase) setStructKeys() {
	name := uncapitalize(g.currentStruct.Name)
	g.keys["strName"] = name
	g.keys["StrName"] = capitalize(name)
	g.keys["str_name"] = snake(name)
	g.keys["STR_NAME"] = upper(snake(name))
}
