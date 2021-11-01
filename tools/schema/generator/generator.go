// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	AccessChain         = "chain"
	AccessCreator       = "creator"
	AccessSelf          = "self"
	AliasThis           = "this"
	InitFunc            = "Init"
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

//nolint:unused
var (
	camelRegExp  = regexp.MustCompile(`_[a-z]`)
	snakeRegExp  = regexp.MustCompile(`[a-z0-9][A-Z]`)
	snakeRegExp2 = regexp.MustCompile(`[A-Z][A-Z]+[a-z]`)
)

type Generator interface {
	funcName(f *Func) string
	generate() error
	generateFuncSignature(f *Func)
	generateInitialFuncs() error
}

type GenBase struct {
	extension  string
	file       *os.File
	Folder     string
	funcRegexp *regexp.Regexp
	gen        Generator
	language   string
	NewTypes   map[string]bool
	rootFolder string
	s          *Schema
}

func (g *GenBase) close() {
	_ = g.file.Close()
}

func (g *GenBase) create(path string) (err error) {
	g.file, err = os.Create(path)
	return err
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

func (g *GenBase) Generate(s *Schema, schemaTime time.Time) error {
	g.s = s
	g.NewTypes = make(map[string]bool)

	g.Folder = g.rootFolder + "/"
	if g.rootFolder != "src" {
		module := strings.ReplaceAll(ModuleCwd, "\\", "/")
		module = module[strings.LastIndex(module, "/")+1:]
		g.Folder += module + "/"
	}
	if g.s.CoreContracts {
		g.Folder += g.s.Name + "/"
	}

	err := os.MkdirAll(g.Folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.Folder + "consts" + g.extension)
	if err == nil && info.ModTime().After(schemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return nil
	}

	fmt.Printf("generating %s code\n", g.language)
	err = g.gen.generate()
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

func (g *GenBase) generateFuncs() error {
	scFileName := g.Folder + g.s.Name + g.extension
	err := g.open(g.Folder + g.s.Name + g.extension)
	if err != nil {
		// generate initial code file
		return g.gen.generateInitialFuncs()
	}

	// append missing function signatures to existing code file

	// scan existing file for signatures
	lines, existing, err := g.scanExistingCode()
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := g.Folder + g.s.Name + ".bak"
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
	for _, f := range g.s.Funcs {
		if existing[g.gen.funcName(f)] == "" {
			g.gen.generateFuncSignature(f)
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

//func (g *GenBase) print(a ...interface{}) {
//	_, _ = fmt.Fprint(g.file, a...)
//}

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

func calculatePadding(fields []*Field, types StringMap, snakeName bool) (nameLen, typeLen int) {
	for _, param := range fields {
		fldName := param.Name
		if snakeName {
			fldName = snake(fldName)
		}
		if nameLen < len(fldName) {
			nameLen = len(fldName)
		}
		fldType := param.Type
		if types != nil {
			fldType = types[fldType]
		}
		if typeLen < len(fldType) {
			typeLen = len(fldType)
		}
	}
	return
}

// convert lowercase snake case to camel case
//nolint:deadcode,unused
func camel(name string) string {
	return camelRegExp.ReplaceAllStringFunc(name, func(sub string) string {
		return strings.ToUpper(sub[1:])
	})
}

// capitalize first letter
func capitalize(name string) string {
	return upper(name[:1]) + name[1:]
}

// TODO take copyright from schema?
func copyright(noChange bool) string {
	text := "// Copyright 2020 IOTA Stiftung\n" +
		"// SPDX-License-Identifier: Apache-2.0\n"
	if noChange {
		text += "\n// (Re-)generated by schema tool\n" +
			"// >>>> DO NOT CHANGE THIS FILE! <<<<\n" +
			"// Change the json schema instead\n"
	}
	return text
}

// convert to lower case
func lower(name string) string {
	return strings.ToLower(name)
}

func FindModulePath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// we're going to walk up the path, make sure to restore
	ModuleCwd = cwd
	defer func() {
		_ = os.Chdir(ModuleCwd)
	}()

	file, err := os.Open("go.mod")
	for err != nil {
		err = os.Chdir("..")
		if err != nil {
			return fmt.Errorf("cannot find go.mod in cwd path")
		}
		prev := cwd
		cwd, err = os.Getwd()
		if err != nil {
			return err
		}
		if cwd == prev {
			// e.g. Chdir("..") gets us in a loop at Linux root
			return fmt.Errorf("cannot find go.mod in cwd path")
		}
		file, err = os.Open("go.mod")
	}

	// now file is the go.mod and cwd holds the path
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			ModuleName = strings.TrimSpace(line[len("module"):])
			ModulePath = cwd
			return nil
		}
	}

	return fmt.Errorf("cannot find module definition in go.mod")
}

// pad to specified size with spaces
func pad(name string, size int) string {
	for i := len(name); i < size; i++ {
		name += " "
	}
	return name
}

// convert camel case to lower case snake case
func snake(name string) string {
	name = snakeRegExp.ReplaceAllStringFunc(name, func(sub string) string {
		return sub[:1] + "_" + sub[1:]
	})
	name = snakeRegExp2.ReplaceAllStringFunc(name, func(sub string) string {
		n := len(sub)
		return sub[:n-2] + "_" + sub[n-2:]
	})
	return lower(name)
}

// uncapitalize first letter
func uncapitalize(name string) string {
	return lower(name[:1]) + name[1:]
}

// convert to upper case
func upper(name string) string {
	return strings.ToUpper(name)
}

func sortedFields(dict FieldMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(dict StringMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedFuncDescs(dict FuncDefMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedMaps(dict StringMapMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
