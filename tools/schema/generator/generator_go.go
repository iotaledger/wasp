// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
)

const (
	goImportCoreTypes  = "import \"github.com/iotaledger/wasp/packages/iscp\""
	goImportWasmLib    = "import \"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib\""
	goImportWasmClient = "import \"github.com/iotaledger/wasp/packages/vm/wasmclient\""
)

var goTypes = StringMap{
	"Address":   "wasmlib.ScAddress",
	"AgentID":   "wasmlib.ScAgentID",
	"ChainID":   "wasmlib.ScChainID",
	"Color":     "wasmlib.ScColor",
	"Hash":      "wasmlib.ScHash",
	"Hname":     "wasmlib.ScHname",
	"Int16":     "int16",
	"Int32":     "int32",
	"Int64":     "int64",
	"RequestID": "wasmlib.ScRequestID",
	"String":    "string",
}

var goKeys = StringMap{
	"Address":   "key",
	"AgentID":   "key",
	"ChainID":   "key",
	"Color":     "key",
	"Hash":      "key",
	"Hname":     "key",
	"Int16":     "??TODO",
	"Int32":     "wasmlib.Key32(key)",
	"Int64":     "??TODO",
	"RequestID": "key",
	"String":    "wasmlib.Key(key)",
}

var goTypeIds = StringMap{
	"Address":   "wasmlib.TYPE_ADDRESS",
	"AgentID":   "wasmlib.TYPE_AGENT_ID",
	"ChainID":   "wasmlib.TYPE_CHAIN_ID",
	"Color":     "wasmlib.TYPE_COLOR",
	"Hash":      "wasmlib.TYPE_HASH",
	"Hname":     "wasmlib.TYPE_HNAME",
	"Int16":     "wasmlib.TYPE_INT16",
	"Int32":     "wasmlib.TYPE_INT32",
	"Int64":     "wasmlib.TYPE_INT64",
	"RequestID": "wasmlib.TYPE_REQUEST_ID",
	"String":    "wasmlib.TYPE_STRING",
}

const (
	goTypeBytes = "wasmlib.TYPE_BYTES"
	goTypeMap   = "wasmlib.TYPE_MAP"
)

type GoGenerator struct {
	GenBase
}

func NewGoGenerator() *GoGenerator {
	g := &GoGenerator{}
	g.extension = ".go"
	g.funcRegexp = regexp.MustCompile(`^func (\w+).+$`)
	g.language = "Go"
	g.rootFolder = "go"
	g.gen = g
	return g
}

func (g *GoGenerator) flushGoConsts() {
	if len(g.s.ConstNames) == 0 {
		return
	}

	if len(g.s.ConstNames) == 1 {
		name := g.s.ConstNames[0]
		value := g.s.ConstValues[0]
		g.printf("\nconst %s = %s\n", name, value)
		g.s.flushConsts(func(name string, value string, padLen int) {})
		return
	}

	g.printf("\nconst (\n")
	g.s.flushConsts(func(name string, value string, padLen int) {
		g.printf("\t%s = %s\n", pad(name, padLen), value)
	})
	g.printf(")\n")
}

func (g *GoGenerator) generate() error {
	err := g.generateGoConsts(false)
	if err != nil {
		return err
	}
	err = g.generateGoTypes()
	if err != nil {
		return err
	}
	err = g.generateGoTypeDefs()
	if err != nil {
		return err
	}
	err = g.generateGoParams()
	if err != nil {
		return err
	}
	err = g.generateGoResults()
	if err != nil {
		return err
	}
	err = g.generateGoContract()
	if err != nil {
		return err
	}

	if !g.s.CoreContracts {
		err = g.generateGoKeys()
		if err != nil {
			return err
		}
		err = g.generateGoState()
		if err != nil {
			return err
		}
		err = g.generateGoLib()
		if err != nil {
			return err
		}
		err = g.generateFuncs()
		if err != nil {
			return err
		}

		// go-specific stuff
		return g.generateGoMain()
	}

	return nil
}

func (g *GoGenerator) generateGoArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if g.s.CoreContracts {
		return "wasmlib.TYPE_ARRAY16|" + varType
	}
	return "wasmlib.TYPE_ARRAY|" + varType
}

func (g *GoGenerator) generateGoConsts(test bool) error {
	err := g.create(g.Folder + "consts" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	packageName := "package test\n"
	importTypes := goImportCoreTypes
	if !test {
		packageName = g.packageName()
		importTypes = goImportWasmLib
	}

	// write file header
	g.println(copyright(true))
	g.println(packageName)
	g.println(importTypes)

	scName := g.s.Name
	if g.s.CoreContracts {
		// remove 'core' prefix
		scName = scName[4:]
	}
	g.s.appendConst("ScName", "\""+scName+"\"")
	if g.s.Description != "" {
		g.s.appendConst("ScDescription", "\""+g.s.Description+"\"")
	}
	hName := iscp.Hn(scName)
	hNameType := "wasmlib.ScHname"
	if test {
		hNameType = "iscp.Hname"
	}
	g.s.appendConst("HScName", hNameType+"(0x"+hName.String()+")")
	g.flushGoConsts()

	g.generateGoConstsFields(test, g.s.Params, "Param")
	g.generateGoConstsFields(test, g.s.Results, "Result")
	g.generateGoConstsFields(test, g.s.StateVars, "State")

	if len(g.s.Funcs) != 0 {
		for _, f := range g.s.Funcs {
			constName := capitalize(f.FuncName)
			g.s.appendConst(constName, "\""+f.String+"\"")
		}
		g.flushGoConsts()

		for _, f := range g.s.Funcs {
			constHname := "H" + capitalize(f.FuncName)
			g.s.appendConst(constHname, hNameType+"(0x"+f.Hname.String()+")")
		}
		g.flushGoConsts()
	}

	return nil
}

func (g *GoGenerator) generateGoConstsFields(test bool, fields []*Field, prefix string) {
	if len(fields) != 0 {
		for _, field := range fields {
			if field.Alias == AliasThis {
				continue
			}
			name := prefix + capitalize(field.Name)
			value := "\"" + field.Alias + "\""
			if !test {
				value = "wasmlib.Key(" + value + ")"
			}
			g.s.appendConst(name, value)
		}
		g.flushGoConsts()
	}
}

func (g *GoGenerator) generateGoContract() error {
	err := g.create(g.Folder + "contract" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(g.packageName())
	g.println(goImportWasmLib)

	for _, f := range g.s.Funcs {
		nameLen := f.nameLen(4)
		kind := f.Kind
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
		}
		g.printf("\ntype %sCall struct {\n", f.Type)
		g.printf("\t%s *wasmlib.Sc%s\n", pad(KindFunc, nameLen), kind)
		if len(f.Params) != 0 {
			g.printf("\t%s Mutable%sParams\n", pad("Params", nameLen), f.Type)
		}
		if len(f.Results) != 0 {
			g.printf("\tResults Immutable%sResults\n", f.Type)
		}
		g.printf("}\n")
	}

	g.generateGoContractFuncs()

	if g.s.CoreContracts {
		g.printf("\nfunc OnLoad() {\n")
		g.printf("\texports := wasmlib.NewScExports()\n")
		for _, f := range g.s.Funcs {
			constName := capitalize(f.FuncName)
			g.printf("\texports.Add%s(%s, wasmlib.%sError)\n", f.Kind, constName, f.Kind)
		}
		g.printf("}\n")
	}
	return nil
}

func (g *GoGenerator) generateGoContractFuncs() {
	g.println("\ntype Funcs struct{}")
	g.println("\nvar ScFuncs Funcs")
	for _, f := range g.s.Funcs {
		assign := "return"
		paramsID := "nil"
		if len(f.Params) != 0 {
			assign = "f :="
			paramsID = "&f.Params.id"
		}
		resultsID := "nil"
		if len(f.Results) != 0 {
			assign = "f :="
			resultsID = "&f.Results.id"
		}
		kind := f.Kind
		keyMap := ""
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
			keyMap = ", keyMap[:], idxMap[:]"
		}
		g.printf("\nfunc (sc Funcs) %s(ctx wasmlib.Sc%sCallContext) *%sCall {\n", f.Type, f.Kind, f.Type)
		g.printf("\t%s &%sCall{Func: wasmlib.NewSc%s(ctx, HScName, H%s%s%s)}\n", assign, f.Type, kind, f.Kind, f.Type, keyMap)
		if len(f.Params) != 0 || len(f.Results) != 0 {
			g.printf("\tf.Func.SetPtrs(%s, %s)\n", paramsID, resultsID)
			g.printf("\treturn f\n")
		}
		g.printf("}\n")
	}
}

func (g *GoGenerator) funcName(f *Func) string {
	return f.FuncName
}

// TODO handle case where owner is type AgentID[]
func (g *GoGenerator) generateFuncSignature(f *Func) {
	g.printf("\nfunc %s(ctx wasmlib.Sc%sContext, f *%sContext) {\n", f.FuncName, f.Kind, f.Type)
	switch f.FuncName {
	case SpecialFuncInit:
		g.printf("    if f.Params.Owner().Exists() {\n")
		g.printf("        f.State.Owner().SetValue(f.Params.Owner().Value())\n")
		g.printf("        return\n")
		g.printf("    }\n")
		g.printf("    f.State.Owner().SetValue(ctx.ContractCreator())\n")
	case SpecialFuncSetOwner:
		g.printf("    f.State.Owner().SetValue(f.Params.Owner().Value())\n")
	case SpecialViewGetOwner:
		g.printf("    f.Results.Owner().SetValue(f.State.Owner().Value())\n")
	default:
	}
	g.printf("}\n")
}

func (g *GoGenerator) generateInitialFuncs() error {
	err := g.create(g.Folder + g.s.Name + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(false))
	g.println(g.packageName())
	g.println(goImportWasmLib)

	for _, f := range g.s.Funcs {
		g.generateFuncSignature(f)
	}
	return nil
}

func (g *GoGenerator) generateGoKeys() error {
	err := g.create(g.Folder + "keys" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(g.packageName())
	g.println(goImportWasmLib)

	g.s.KeyID = 0
	g.generateGoKeysIndexes(g.s.Params, "Param")
	g.generateGoKeysIndexes(g.s.Results, "Result")
	g.generateGoKeysIndexes(g.s.StateVars, "State")
	g.flushGoConsts()

	size := g.s.KeyID
	g.printf("\nconst keyMapLen = %d\n", size)
	g.printf("\nvar keyMap = [keyMapLen]wasmlib.Key{\n")
	g.generateGoKeysArray(g.s.Params, "Param")
	g.generateGoKeysArray(g.s.Results, "Result")
	g.generateGoKeysArray(g.s.StateVars, "State")
	g.printf("}\n")
	g.printf("\nvar idxMap [keyMapLen]wasmlib.Key32\n")
	return nil
}

func (g *GoGenerator) generateGoKeysArray(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := prefix + capitalize(field.Name)
		g.printf("\t%s,\n", name)
		g.s.KeyID++
	}
}

func (g *GoGenerator) generateGoKeysIndexes(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := "Idx" + prefix + capitalize(field.Name)
		field.KeyID = g.s.KeyID
		value := strconv.Itoa(field.KeyID)
		g.s.KeyID++
		g.s.appendConst(name, value)
	}
}

func (g *GoGenerator) generateGoLib() error {
	err := g.create(g.Folder + "lib" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println("//nolint:dupl")
	g.println(g.packageName())
	g.println(goImportWasmLib)

	g.printf("\nfunc OnLoad() {\n")
	g.printf("\texports := wasmlib.NewScExports()\n")
	for _, f := range g.s.Funcs {
		constName := capitalize(f.FuncName)
		g.printf("\texports.Add%s(%s, %sThunk)\n", f.Kind, constName, f.FuncName)
	}

	g.printf("\n\tfor i, key := range keyMap {\n")
	g.printf("\t\tidxMap[i] = key.KeyID()\n")
	g.printf("\t}\n")

	g.printf("}\n")

	// generate parameter structs and thunks to set up and check parameters
	for _, f := range g.s.Funcs {
		g.generateGoThunk(f)
	}
	return nil
}

func (g *GoGenerator) generateGoMain() error {
	err := g.create(g.Folder + "../main" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	module := ModuleName + strings.Replace(ModuleCwd[len(ModulePath):], "\\", "/", -1)

	// write file header
	g.println(copyright(true))
	g.println("// +build wasm")

	g.println("\npackage main")
	g.println()
	g.println(goImportWasmClient)
	g.printf("\nimport \"%s/go/%s\"\n", module, g.s.Name)

	g.printf("\nfunc main() {\n")
	g.printf("}\n")

	g.printf("\n//export on_load\n")
	g.printf("func onLoad() {\n")
	g.printf("\th := &wasmclient.WasmVMHost{}\n")
	g.printf("\th.ConnectWasmHost()\n")
	g.printf("\t%s.OnLoad()\n", g.s.Name)
	g.printf("}\n")

	return nil
}

func (g *GoGenerator) generateGoParams() error {
	err := g.create(g.Folder + "params" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.printf(g.packageName())

	totalParams := 0
	for _, f := range g.s.Funcs {
		totalParams += len(f.Params)
	}
	if totalParams != 0 {
		g.println()
		g.println(goImportWasmLib)
	}

	for _, f := range g.s.Funcs {
		if len(f.Params) == 0 {
			continue
		}
		g.generateGoStruct(f.Params, PropImmutable, f.Type, "Params")
		g.generateGoStruct(f.Params, PropMutable, f.Type, "Params")
	}

	return nil
}

func (g *GoGenerator) generateGoProxy(field *Field, mutability string) {
	if field.Array {
		g.generateGoProxyArray(field, mutability)
		return
	}

	if field.MapKey != "" {
		g.generateGoProxyMap(field, mutability)
	}
}

func (g *GoGenerator) generateGoProxyArray(field *Field, mutability string) {
	proxyType := mutability + field.Type
	arrayType := "ArrayOf" + proxyType
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\ntype %s%s = %s\n", mutability, field.Name, arrayType)
	}
	if g.NewTypes[arrayType] {
		// already generated this array
		return
	}
	g.NewTypes[arrayType] = true

	g.printf("\ntype %s struct {\n", arrayType)
	g.printf("\tobjID int32\n")
	g.printf("}\n")

	if mutability == PropMutable {
		g.printf("\nfunc (a %s) Clear() {\n", arrayType)
		g.printf("\twasmlib.Clear(a.objID)\n")
		g.printf("}\n")
	}

	g.printf("\nfunc (a %s) Length() int32 {\n", arrayType)
	g.printf("\treturn wasmlib.GetLength(a.objID)\n")
	g.printf("}\n")

	if field.TypeID == 0 {
		g.generateGoProxyArrayNewType(field, proxyType, arrayType)
		return
	}

	// array of predefined type
	g.printf("\nfunc (a %s) Get%s(index int32) wasmlib.Sc%s {\n", arrayType, field.Type, proxyType)
	g.printf("\treturn wasmlib.NewSc%s(a.objID, wasmlib.Key32(index))\n", proxyType)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoProxyArrayNewType(field *Field, proxyType, arrayType string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := goTypeMap
		if subtype.Array {
			varType = goTypeIds[subtype.Type]
			if varType == "" {
				varType = goTypeBytes
			}
			varType = g.generateGoArrayType(varType)
		}
		g.printf("\nfunc (a %s) Get%s(index int32) %s {\n", arrayType, field.Type, proxyType)
		g.printf("\tsubID := wasmlib.GetObjectID(a.objID, wasmlib.Key32(index), %s)\n", varType)
		g.printf("\treturn %s{objID: subID}\n", proxyType)
		g.printf("}\n")
		return
	}

	g.printf("\nfunc (a %s) Get%s(index int32) %s {\n", arrayType, field.Type, proxyType)
	g.printf("\treturn %s{objID: a.objID, keyID: wasmlib.Key32(index)}\n", proxyType)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoProxyMap(field *Field, mutability string) {
	proxyType := mutability + field.Type
	mapType := "Map" + field.MapKey + "To" + proxyType
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\ntype %s%s = %s\n", mutability, field.Name, mapType)
	}
	if g.NewTypes[mapType] {
		// already generated this map
		return
	}
	g.NewTypes[mapType] = true

	keyType := goTypes[field.MapKey]
	keyValue := goKeys[field.MapKey]

	g.printf("\ntype %s struct {\n", mapType)
	g.printf("\tobjID int32\n")
	g.printf("}\n")

	if mutability == PropMutable {
		g.printf("\nfunc (m %s) Clear() {\n", mapType)
		g.printf("\twasmlib.Clear(m.objID)\n")
		g.printf("}\n")
	}

	if field.TypeID == 0 {
		g.generateGoProxyMapNewType(field, proxyType, mapType, keyType, keyValue)
		return
	}

	// map of predefined type
	g.printf("\nfunc (m %s) Get%s(key %s) wasmlib.Sc%s {\n", mapType, field.Type, keyType, proxyType)
	g.printf("\treturn wasmlib.NewSc%s(m.objID, %s.KeyID())\n", proxyType, keyValue)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoProxyMapNewType(field *Field, proxyType, mapType, keyType, keyValue string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := goTypeMap
		if subtype.Array {
			varType = goTypeIds[subtype.Type]
			if varType == "" {
				varType = goTypeBytes
			}
			varType = g.generateGoArrayType(varType)
		}
		g.printf("\nfunc (m %s) Get%s(key %s) %s {\n", mapType, field.Type, keyType, proxyType)
		g.printf("\tsubID := wasmlib.GetObjectID(m.objID, %s.KeyID(), %s)\n", keyValue, varType)
		g.printf("\treturn %s{objID: subID}\n", proxyType)
		g.printf("}\n")
		return
	}

	g.printf("\nfunc (m %s) Get%s(key %s) %s {\n", mapType, field.Type, keyType, proxyType)
	g.printf("\treturn %s{objID: m.objID, keyID: %s.KeyID()}\n", proxyType, keyValue)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoResults() error {
	err := g.create(g.Folder + "results" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.printf(g.packageName())

	results := 0
	for _, f := range g.s.Funcs {
		results += len(f.Results)
	}
	if results != 0 {
		g.println()
		g.println(goImportWasmLib)
	}

	for _, f := range g.s.Funcs {
		if len(f.Results) == 0 {
			continue
		}
		g.generateGoStruct(f.Results, PropImmutable, f.Type, "Results")
		g.generateGoStruct(f.Results, PropMutable, f.Type, "Results")
	}
	return nil
}

func (g *GoGenerator) generateGoState() error {
	err := g.create(g.Folder + "state" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.printf(g.packageName())
	if len(g.s.StateVars) != 0 {
		g.println()
		g.println(goImportWasmLib)
	}

	g.generateGoStruct(g.s.StateVars, PropImmutable, g.s.FullName, "State")
	g.generateGoStruct(g.s.StateVars, PropMutable, g.s.FullName, "State")
	return nil
}

// TODO nested structs
func (g *GoGenerator) generateGoStruct(fields []*Field, mutability, typeName, kind string) {
	typeName = mutability + typeName + kind
	kind = strings.TrimSuffix(kind, "s")

	// first generate necessary array and map types
	for _, field := range fields {
		g.generateGoProxy(field, mutability)
	}

	g.printf("\ntype %s struct {\n", typeName)
	g.printf("\tid int32\n")
	g.printf("}\n")

	for _, field := range fields {
		varName := capitalize(field.Name)
		varID := "idxMap[Idx" + kind + varName + "]"
		if g.s.CoreContracts {
			varID = kind + varName + ".KeyID()"
		}
		varType := goTypeIds[field.Type]
		if varType == "" {
			varType = goTypeBytes
		}
		if field.Array {
			varType = g.generateGoArrayType(varType)
			arrayType := "ArrayOf" + mutability + field.Type
			g.printf("\nfunc (s %s) %s() %s {\n", typeName, varName, arrayType)
			g.printf("\tarrID := wasmlib.GetObjectID(s.id, %s, %s)\n", varID, varType)
			g.printf("\treturn %s{objID: arrID}\n", arrayType)
			g.printf("}\n")
			continue
		}
		if field.MapKey != "" {
			varType = goTypeMap
			mapType := "Map" + field.MapKey + "To" + mutability + field.Type
			g.printf("\nfunc (s %s) %s() %s {\n", typeName, varName, mapType)
			mapID := "s.id"
			if field.Alias != AliasThis {
				mapID = "mapID"
				g.printf("\tmapID := wasmlib.GetObjectID(s.id, %s, %s)\n", varID, varType)
			}
			g.printf("\treturn %s{objID: %s}\n", mapType, mapID)
			g.printf("}\n")
			continue
		}

		proxyType := mutability + field.Type
		if field.TypeID == 0 {
			g.printf("\nfunc (s %s) %s() %s {\n", typeName, varName, proxyType)
			g.printf("\treturn %s{objID: s.id, keyID: %s}\n", proxyType, varID)
			g.printf("}\n")
			continue
		}

		g.printf("\nfunc (s %s) %s() wasmlib.Sc%s {\n", typeName, varName, proxyType)
		g.printf("\treturn wasmlib.NewSc%s(s.id, %s)\n", proxyType, varID)
		g.printf("}\n")
	}
}

func (g *GoGenerator) generateGoThunk(f *Func) {
	nameLen := f.nameLen(5)
	mutability := PropMutable
	if f.Kind == KindView {
		mutability = PropImmutable
	}
	g.printf("\ntype %sContext struct {\n", f.Type)
	if len(f.Params) != 0 {
		g.printf("\t%s Immutable%sParams\n", pad("Params", nameLen), f.Type)
	}
	if len(f.Results) != 0 {
		g.printf("\tResults Mutable%sResults\n", f.Type)
	}
	g.printf("\t%s %s%sState\n", pad("State", nameLen), mutability, g.s.FullName)
	g.printf("}\n")

	g.printf("\nfunc %sThunk(ctx wasmlib.Sc%sContext) {\n", f.FuncName, f.Kind)
	g.printf("\tctx.Log(\"%s.%s\")\n", g.s.Name, f.FuncName)

	if f.Access != "" {
		g.generateGoThunkAccessCheck(f)
	}

	g.printf("\tf := &%sContext{\n", f.Type)

	if len(f.Params) != 0 {
		g.printf("\t\tParams: Immutable%sParams{\n", f.Type)
		g.printf("\t\t\tid: wasmlib.OBJ_ID_PARAMS,\n")
		g.printf("\t\t},\n")
	}

	if len(f.Results) != 0 {
		g.printf("\t\tResults: Mutable%sResults{\n", f.Type)
		g.printf("\t\t\tid: wasmlib.OBJ_ID_RESULTS,\n")
		g.printf("\t\t},\n")
	}

	g.printf("\t\tState: %s%sState{\n", mutability, g.s.FullName)
	g.printf("\t\t\tid: wasmlib.OBJ_ID_STATE,\n")
	g.printf("\t\t},\n")

	g.printf("\t}\n")

	for _, param := range f.Params {
		if !param.Optional {
			name := capitalize(param.Name)
			g.printf("\tctx.Require(f.Params.%s().Exists(), \"missing mandatory %s\")\n", name, param.Name)
		}
	}

	g.printf("\t%s(ctx, f)\n", f.FuncName)
	g.printf("\tctx.Log(\"%s.%s ok\")\n", g.s.Name, f.FuncName)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoThunkAccessCheck(f *Func) {
	grant := f.Access
	index := strings.Index(grant, "//")
	if index >= 0 {
		g.printf("\t%s\n", grant[index:])
		grant = strings.TrimSpace(grant[:index])
	}
	switch grant {
	case AccessSelf:
		grant = "ctx.AccountID()"
	case AccessChain:
		grant = "ctx.ChainOwnerID()"
	case AccessCreator:
		grant = "ctx.ContractCreator()"
	default:
		g.printf("\taccess := ctx.State().GetAgentID(wasmlib.Key(\"%s\"))\n", grant)
		g.printf("\tctx.Require(access.Exists(), \"access not set: %s\")\n", grant)
		grant = "access.Value()"
	}
	g.printf("\tctx.Require(ctx.Caller() == %s, \"no permission\")\n\n", grant)
}

func (g *GoGenerator) generateGoTypes() error {
	if len(g.s.Structs) == 0 {
		return nil
	}

	err := g.create(g.Folder + "types" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.println(g.packageName())
	g.println(goImportWasmLib)

	for _, typeDef := range g.s.Structs {
		g.generateGoType(typeDef)
	}

	return nil
}

func (g *GoGenerator) generateGoType(typeDef *Struct) {
	nameLen, typeLen := calculatePadding(typeDef.Fields, goTypes, false)

	g.printf("\ntype %s struct {\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		fldName := pad(capitalize(field.Name), nameLen)
		fldType := goTypes[field.Type]
		if field.Comment != "" {
			fldType = pad(fldType, typeLen)
		}
		g.printf("\t%s %s%s\n", fldName, fldType, field.Comment)
	}
	g.printf("}\n")

	// write encoder and decoder for struct
	g.printf("\nfunc New%sFromBytes(bytes []byte) *%s {\n", typeDef.Name, typeDef.Name)
	g.printf("\tdecode := wasmlib.NewBytesDecoder(bytes)\n")
	g.printf("\tdata := &%s{}\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		name := capitalize(field.Name)
		g.printf("\tdata.%s = decode.%s()\n", name, field.Type)
	}
	g.printf("\tdecode.Close()\n")
	g.printf("\treturn data\n}\n")

	g.printf("\nfunc (o *%s) Bytes() []byte {\n", typeDef.Name)
	g.printf("\treturn wasmlib.NewBytesEncoder().\n")
	for _, field := range typeDef.Fields {
		name := capitalize(field.Name)
		g.printf("\t\t%s(o.%s).\n", field.Type, name)
	}
	g.printf("\t\tData()\n}\n")

	g.generateGoTypeProxy(typeDef, false)
	g.generateGoTypeProxy(typeDef, true)
}

func (g *GoGenerator) generateGoTypeProxy(typeDef *Struct, mutable bool) {
	typeName := PropImmutable + typeDef.Name
	if mutable {
		typeName = PropMutable + typeDef.Name
	}

	g.printf("\ntype %s struct {\n", typeName)
	g.printf("\tobjID int32\n")
	g.printf("\tkeyID wasmlib.Key32\n")
	g.printf("}\n")

	g.printf("\nfunc (o %s) Exists() bool {\n", typeName)
	g.printf("\treturn wasmlib.Exists(o.objID, o.keyID, wasmlib.TYPE_BYTES)\n")
	g.printf("}\n")

	if mutable {
		g.printf("\nfunc (o %s) SetValue(value *%s) {\n", typeName, typeDef.Name)
		g.printf("\twasmlib.SetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES, value.Bytes())\n")
		g.printf("}\n")
	}

	g.printf("\nfunc (o %s) Value() *%s {\n", typeName, typeDef.Name)
	g.printf("\treturn New%sFromBytes(wasmlib.GetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES))\n", typeDef.Name)
	g.printf("}\n")
}

func (g *GoGenerator) generateGoTypeDefs() error {
	if len(g.s.Typedefs) == 0 {
		return nil
	}

	err := g.create(g.Folder + "typedefs" + g.extension)
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.println(g.packageName())
	g.println(goImportWasmLib)

	for _, subtype := range g.s.Typedefs {
		g.generateGoProxy(subtype, PropImmutable)
		g.generateGoProxy(subtype, PropMutable)
	}

	return nil
}

func (g *GoGenerator) packageName() string {
	return fmt.Sprintf("package %s\n", g.s.Name)
}
