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
	AccessChain         = "chain"
	AccessCreator       = "creator"
	AccessSelf          = "self"
	AliasThis           = "this"
	InitFunc            = "Init"
	KeyArray            = "array"
	KeyBaseType         = "basetype"
	KeyCore             = "core"
	KeyExist            = "exist"
	KeyMap              = "map"
	KeyMut              = "mut"
	KeyFunc             = "func"
	KeyMandatory        = "mandatory"
	KeyParam            = "param"
	KeyParams           = "params"
	KeyProxy            = "proxy"
	KeyPtrs             = "ptrs"
	KeyResult           = "result"
	KeyResults          = "results"
	KeyState            = "state"
	KeyStruct           = "struct"
	KeyStructs          = "structs"
	KeyThis             = "this"
	KeyTypeDef          = "typedef"
	KeyTypeDefs         = "typedefs"
	KeyView             = "view"
	KindFunc            = "Func"
	KindView            = "View"
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
	setFieldKeys()
	setFuncKeys()
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

func (g *GenBase) init(s *Schema) {
	g.s = s
	g.emitters = map[string]func(g *GenBase){}
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

var emitKeyRegExp = regexp.MustCompile(`\$[a-zA-Z_]+`)

func (g *GenBase) emit(template string) {
	lines := strings.Split(g.templates[template], "\n")
	for i := 1; i < len(lines)-1; i++ {
		line := lines[i]

		// replace any placeholder keys
		line = emitKeyRegExp.ReplaceAllStringFunc(line, func(key string) string {
			text, ok := g.keys[key[1:]]
			if ok {
				return text
			}
			return "???:" + key
		})

		// remove concatenation markers
		line = strings.ReplaceAll(line, "$+", "")

		// now process special commands
		if strings.HasPrefix(line, "$#") {
			if strings.HasPrefix(line, "$#emit ") {
				g.emit(strings.TrimSpace(line[7:]))
				continue
			}
			if strings.HasPrefix(line, "$#each ") {
				g.emitEach(line)
				continue
			}
			if strings.HasPrefix(line, "$#func ") {
				g.emitFunc(line)
				continue
			}
			if strings.HasPrefix(line, "$#if ") {
				g.emitIf(line)
				continue
			}
			if strings.HasPrefix(line, "$#set ") {
				g.emitSet(line)
				continue
			}
			g.println("???:" + line)
			continue
		}

		g.println(line)
	}
}

func (g *GenBase) emitEach(key string) {
	parts := strings.Split(key, " ")
	if len(parts) != 3 {
		g.println("???:" + key)
		return
	}

	template := parts[2]
	switch parts[1] {
	case KeyFunc:
		for _, g.currentFunc = range g.s.Funcs {
			g.gen.setFuncKeys()
			g.emit(template)
		}
	case KeyMandatory:
		mandatory := []*Field{}
		for _, g.currentField = range g.currentFunc.Params {
			if !g.currentField.Optional {
				mandatory = append(mandatory, g.currentField)
			}
		}
		g.emitFields(mandatory, template)
	case KeyParam:
		g.emitFields(g.currentFunc.Params, template)
	case KeyParams:
		g.emitFields(g.s.Params, template)
	case KeyResult:
		g.emitFields(g.currentFunc.Results, template)
	case KeyResults:
		g.emitFields(g.s.Results, template)
	case KeyState:
		g.emitFields(g.s.StateVars, template)
	case KeyStruct:
		g.emitFields(g.currentStruct.Fields, template)
	case KeyStructs:
		for _, g.currentStruct = range g.s.Structs {
			g.setKeyValues("strName", g.currentStruct.Name)
			g.emit(template)
		}
	case KeyTypeDef:
		g.emitFields(g.s.Typedefs, template)
	default:
		g.println("???:" + key)
	}
}

func (g *GenBase) emitFields(fields []*Field, template string) {
	for _, g.currentField = range fields {
		//if g.currentField.Alias == KeyThis {
		//	continue
		//}
		g.gen.setFieldKeys()
		g.emit(template)
	}
}

func (g *GenBase) emitFunc(key string) {
	parts := strings.Split(key, " ")
	if len(parts) != 2 {
		g.println("???:" + key)
		return
	}

	emitter, ok := g.emitters[parts[1]]
	if ok {
		emitter(g)
		return
	}
	g.println("???:" + key)
}

func (g *GenBase) emitIf(key string) {
	parts := strings.Split(key, " ")
	if len(parts) < 3 || len(parts) > 4 {
		g.println("???:" + key)
		return
	}

	conditionKey := parts[1]
	template := parts[2]

	condition := false
	switch conditionKey {
	case KeyArray:
		condition = g.currentField.Array
	case KeyBaseType:
		condition = g.currentField.TypeID != 0
	case KeyExist:
		proxy := g.keys[KeyProxy]
		condition = g.newTypes[proxy]
	case KeyMap:
		condition = g.currentField.MapKey != ""
	case KeyMut:
		condition = g.keys[KeyMut] == "Mutable"
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
	case KeyStructs:
		condition = len(g.s.Structs) != 0
	case KeyThis:
		condition = g.currentField.Alias == KeyThis
	case KeyTypeDef:
		condition = g.fieldIsTypeDef()
	case KeyTypeDefs:
		condition = len(g.s.Typedefs) != 0
	case KeyPtrs:
		condition = len(g.currentFunc.Params) != 0 || len(g.currentFunc.Results) != 0
	default:
		g.println("???:" + key)
		return
	}

	if condition {
		g.emit(template)
		return
	}

	// else branch?
	if len(parts) == 4 {
		template = parts[3]
		g.emit(template)
	}
}

func (g *GenBase) fieldIsTypeDef() bool {
	for _, typeDef := range g.s.Typedefs {
		if typeDef.Name == g.currentField.Type {
			g.currentField = typeDef
			g.gen.setFieldKeys()
			return true
		}
	}
	return false
}

func (g *GenBase) emitSet(line string) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		g.println("???:" + line)
		return
	}

	key := parts[1]
	value := line[len(parts[0])+len(key)+2:]
	g.keys[key] = value

	if key == KeyExist {
		g.newTypes[value] = true
	}
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

func (g *GenBase) setFieldKeys() {
	g.setKeyValues("fldName", g.currentField.Name)
	g.setKeyValues("fldType", g.currentField.Type)

	g.keys["fldAlias"] = g.currentField.Alias
	g.keys["FldComment"] = g.currentField.Comment
	g.keys["FldMapKey"] = g.currentField.MapKey

	g.keys["fldIndex"] = strconv.Itoa(g.s.KeyID)
	g.s.KeyID++
	g.keys["maxIndex"] = strconv.Itoa(g.s.KeyID)
}

func (g *GenBase) setFuncKeys() {
	g.setKeyValues("funcName", g.currentFunc.FuncName[4:])
	g.setKeyValues("kind", g.currentFunc.FuncName[:4])
	g.keys["funcHName"] = iscp.Hn(g.keys["funcName"]).String()
}

func (g *GenBase) setKeys() {
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

func (g *GenBase) setKeyValues(key, value string) {
	value = uncapitalize(value)
	g.keys[key] = value
	g.keys[capitalize(key)] = capitalize(value)
	g.keys[snake(key)] = snake(value)
	g.keys[upper(snake(key))] = upper(snake(value))
}
