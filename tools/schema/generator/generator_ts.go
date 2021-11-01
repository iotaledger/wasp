// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
)

const (
	tsImportSelf    = "import * as sc from \"./index\";"
	tsImportWasmLib = "import * as wasmlib from \"wasmlib\""
)

var tsFuncRegexp = regexp.MustCompile(`^export function (\w+).+$`)

var tsTypes = StringMap{
	"Address":   "wasmlib.ScAddress",
	"AgentID":   "wasmlib.ScAgentID",
	"ChainID":   "wasmlib.ScChainID",
	"Color":     "wasmlib.ScColor",
	"Hash":      "wasmlib.ScHash",
	"Hname":     "wasmlib.ScHname",
	"Int16":     "i16",
	"Int32":     "i32",
	"Int64":     "i64",
	"RequestID": "wasmlib.ScRequestID",
	"String":    "string",
}

var tsInits = StringMap{
	"Address":   "new wasmlib.ScAddress()",
	"AgentID":   "new wasmlib.ScAgentID()",
	"ChainID":   "new wasmlib.ScChainID()",
	"Color":     "new wasmlib.ScColor(0)",
	"Hash":      "new wasmlib.ScHash()",
	"Hname":     "new wasmlib.ScHname(0)",
	"Int16":     "0",
	"Int32":     "0",
	"Int64":     "0",
	"RequestID": "new wasmlib.ScRequestID()",
	"String":    "\"\"",
}

var tsKeys = StringMap{
	"Address":   "key",
	"AgentID":   "key",
	"ChainID":   "key",
	"Color":     "key",
	"Hash":      "key",
	"Hname":     "key",
	"Int16":     "??TODO",
	"Int32":     "new wasmlib.Key32(key)",
	"Int64":     "??TODO",
	"RequestID": "key",
	"String":    "wasmlib.Key32.fromString(key)",
}

var tsTypeIds = StringMap{
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
	tsTypeBytes = "wasmlib.TYPE_BYTES"
	tsTypeMap   = "wasmlib.TYPE_MAP"
)

type TypeScriptGenerator struct {
	Generator
}

func NewTypeScriptGenerator(s *Schema) *TypeScriptGenerator {
	return &TypeScriptGenerator{Generator{s: s}}
}

func (g *TypeScriptGenerator) flushTsConsts() {
	if len(g.s.ConstNames) == 0 {
		return
	}

	g.println()
	g.s.flushConsts(func(name string, value string, padLen int) {
		g.printf("export const %s = %s;\n", pad(name, padLen), value)
	})
}

func (g *TypeScriptGenerator) GenerateTs() error {
	g.s.NewTypes = make(map[string]bool)

	err := g.generateTsConsts()
	if err != nil {
		return err
	}
	err = g.generateTsTypes()
	if err != nil {
		return err
	}
	err = g.generateTsTypeDefs()
	if err != nil {
		return err
	}
	err = g.generateTsParams()
	if err != nil {
		return err
	}
	err = g.generateTsResults()
	if err != nil {
		return err
	}
	err = g.generateTsContract()
	if err != nil {
		return err
	}

	if !g.s.CoreContracts {
		err = g.generateTsKeys()
		if err != nil {
			return err
		}
		err = g.generateTsState()
		if err != nil {
			return err
		}
		err = g.generateTsLib()
		if err != nil {
			return err
		}
		err = g.generateTsFuncs()
		if err != nil {
			return err
		}
	}

	// typescript-specific stuff
	err = g.generateTsConfig()
	if err != nil {
		return err
	}
	return g.generateTsIndex()
}

func (g *TypeScriptGenerator) generateTsArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if g.s.CoreContracts {
		return "wasmlib.TYPE_ARRAY16|" + varType
	}
	return "wasmlib.TYPE_ARRAY|" + varType
}

func (g *TypeScriptGenerator) generateTsConfig() error {
	err := g.exists(g.s.Folder + "tsconfig.json")
	if err == nil {
		// already exists
		return nil
	}

	err = g.create(g.s.Folder + "tsconfig.json")
	if err != nil {
		return err
	}
	defer g.close()

	g.println("{")
	g.println("  \"extends\": \"assemblyscript/std/assembly.json\",")
	g.println("  \"include\": [\"./*.ts\"]")
	g.println("}")

	return nil
}

func (g *TypeScriptGenerator) generateTsConsts() error {
	err := g.create(g.s.Folder + "consts.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)

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
	hNameType := "new wasmlib.ScHname"
	g.s.appendConst("HScName", hNameType+"(0x"+hName.String()+")")
	g.flushTsConsts()

	g.generateTsConstsFields(g.s.Params, "Param")
	g.generateTsConstsFields(g.s.Results, "Result")
	g.generateTsConstsFields(g.s.StateVars, "State")

	if len(g.s.Funcs) != 0 {
		for _, f := range g.s.Funcs {
			constName := capitalize(f.FuncName)
			g.s.appendConst(constName, "\""+f.String+"\"")
		}
		g.flushTsConsts()

		for _, f := range g.s.Funcs {
			constHname := "H" + capitalize(f.FuncName)
			g.s.appendConst(constHname, hNameType+"(0x"+f.Hname.String()+")")
		}
		g.flushTsConsts()
	}

	return nil
}

func (g *TypeScriptGenerator) generateTsConstsFields(fields []*Field, prefix string) {
	if len(fields) != 0 {
		for _, field := range fields {
			if field.Alias == AliasThis {
				continue
			}
			name := prefix + capitalize(field.Name)
			value := "\"" + field.Alias + "\""
			g.s.appendConst(name, value)
		}
		g.flushTsConsts()
	}
}

func (g *TypeScriptGenerator) generateTsContract() error {
	err := g.create(g.s.Folder + "contract.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	for _, f := range g.s.Funcs {
		kind := f.Kind
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
		}
		g.printf("\nexport class %sCall {\n", f.Type)
		g.printf("    func: wasmlib.Sc%s = new wasmlib.Sc%s(sc.HScName, sc.H%s%s);\n", kind, kind, f.Kind, f.Type)
		if len(f.Params) != 0 {
			g.printf("    params: sc.Mutable%sParams = new sc.Mutable%sParams();\n", f.Type, f.Type)
		}
		if len(f.Results) != 0 {
			g.printf("    results: sc.Immutable%sResults = new sc.Immutable%sResults();\n", f.Type, f.Type)
		}
		g.printf("}\n")

		if !g.s.CoreContracts {
			mutability := PropMutable
			if f.Kind == KindView {
				mutability = PropImmutable
			}
			g.printf("\nexport class %sContext {\n", f.Type)
			if len(f.Params) != 0 {
				g.printf("    params: sc.Immutable%sParams = new sc.Immutable%sParams();\n", f.Type, f.Type)
			}
			if len(f.Results) != 0 {
				g.printf("    results: sc.Mutable%sResults = new sc.Mutable%sResults();\n", f.Type, f.Type)
			}
			g.printf("    state: sc.%s%sState = new sc.%s%sState();\n", mutability, g.s.FullName, mutability, g.s.FullName)
			g.printf("}\n")
		}
	}

	g.generateTsContractFuncs()
	return nil
}

func (g *TypeScriptGenerator) generateTsContractFuncs() {
	g.println("\nexport class ScFuncs {")
	for _, f := range g.s.Funcs {
		g.printf("\n    static %s(ctx: wasmlib.Sc%sCallContext): %sCall {\n", uncapitalize(f.Type), f.Kind, f.Type)
		g.printf("        let f = new %sCall();\n", f.Type)

		paramsID := "null"
		if len(f.Params) != 0 {
			paramsID = "f.params"
		}
		resultsID := "null"
		if len(f.Results) != 0 {
			resultsID = "f.results"
		}
		if len(f.Params) != 0 || len(f.Results) != 0 {
			g.printf("        f.func.setPtrs(%s, %s);\n", paramsID, resultsID)
		}
		g.printf("        return f;\n")
		g.printf("    }\n")
	}
	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsFuncs() error {
	scFileName := g.s.Folder + g.s.Name + ".ts"
	err := g.open(scFileName)
	if err != nil {
		// generate initial code file
		return g.generateTsFuncsNew(scFileName)
	}

	// append missing function signatures to existing code file

	// scan existing file for signatures
	lines, existing, err := g.scanExistingCode(tsFuncRegexp)
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := g.s.Folder + g.s.Name + ".bak"
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
		if existing[f.FuncName] == "" {
			g.generateTsFuncSignature(f)
		}
	}

	return os.Remove(scOriginal)
}

func (g *TypeScriptGenerator) generateTsFuncSignature(f *Func) {
	g.printf("\nexport function %s(ctx: wasmlib.Sc%sContext, f: sc.%sContext): void {\n", f.FuncName, f.Kind, f.Type)
	switch f.FuncName {
	case SpecialFuncInit:
		g.printf("    if (f.params.owner().exists()) {\n")
		g.printf("        f.state.owner().setValue(f.params.owner().value());\n")
		g.printf("        return;\n")
		g.printf("    }\n")
		g.printf("    f.state.owner().setValue(ctx.contractCreator());\n")
	case SpecialFuncSetOwner:
		g.printf("    f.state.owner().setValue(f.params.owner().value());\n")
	case SpecialViewGetOwner:
		g.printf("    f.results.owner().setValue(f.state.owner().value());\n")
	default:
	}
	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsFuncsNew(scFileName string) error {
	err := g.create(scFileName)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(false))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	for _, f := range g.s.Funcs {
		g.generateTsFuncSignature(f)
	}
	return nil
}

func (g *TypeScriptGenerator) generateTsIndex() error {
	err := g.create(g.s.Folder + "index.ts")
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))

	if !g.s.CoreContracts {
		g.printf("export * from \"./%s\";\n\n", g.s.Name)
	}

	g.println("export * from \"./consts\";")
	g.println("export * from \"./contract\";")
	if !g.s.CoreContracts {
		g.println("export * from \"./keys\";")
		g.println("export * from \"./lib\";")
	}
	if len(g.s.Params) != 0 {
		g.println("export * from \"./params\";")
	}
	if len(g.s.Results) != 0 {
		g.println("export * from \"./results\";")
	}
	if !g.s.CoreContracts {
		g.println("export * from \"./state\";")
		if len(g.s.Structs) != 0 {
			g.println("export * from \"./types\";")
		}
		if len(g.s.Typedefs) != 0 {
			g.println("export * from \"./typedefs\";")
		}
	}
	return nil
}

func (g *TypeScriptGenerator) generateTsKeys() error {
	err := g.create(g.s.Folder + "keys.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	g.s.KeyID = 0
	g.generateTsKeysIndexes(g.s.Params, "Param")
	g.generateTsKeysIndexes(g.s.Results, "Result")
	g.generateTsKeysIndexes(g.s.StateVars, "State")
	g.flushTsConsts()

	g.printf("\nexport let keyMap: string[] = [\n")
	g.generateTsKeysArray(g.s.Params, "Param")
	g.generateTsKeysArray(g.s.Results, "Result")
	g.generateTsKeysArray(g.s.StateVars, "State")
	g.printf("];\n")
	g.printf("\nexport let idxMap: wasmlib.Key32[] = new Array(keyMap.length);\n")
	return nil
}

func (g *TypeScriptGenerator) generateTsKeysArray(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := prefix + capitalize(field.Name)
		g.printf("    sc.%s,\n", name)
		g.s.KeyID++
	}
}

func (g *TypeScriptGenerator) generateTsKeysIndexes(fields []*Field, prefix string) {
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

func (g *TypeScriptGenerator) generateTsLib() error {
	err := g.create(g.s.Folder + "lib.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	g.printf("\nexport function on_call(index: i32): void {\n")
	g.printf("    return wasmlib.onCall(index);\n")
	g.printf("}\n")

	g.printf("\nexport function on_load(): void {\n")
	g.printf("    let exports = new wasmlib.ScExports();\n")
	for _, f := range g.s.Funcs {
		constName := capitalize(f.FuncName)
		g.printf("    exports.add%s(sc.%s, %sThunk);\n", f.Kind, constName, f.FuncName)
	}

	g.printf("\n    for (let i = 0; i < sc.keyMap.length; i++) {\n")
	g.printf("        sc.idxMap[i] = wasmlib.Key32.fromString(sc.keyMap[i]);\n")
	g.printf("    }\n")

	g.printf("}\n")

	// generate parameter structs and thunks to set up and check parameters
	for _, f := range g.s.Funcs {
		g.generateTsThunk(f)
	}
	return nil
}

func (g *TypeScriptGenerator) generateTsParams() error {
	err := g.create(g.s.Folder + "params.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	for _, f := range g.s.Funcs {
		if len(f.Params) == 0 {
			continue
		}
		g.generateTsStruct(f.Params, PropImmutable, f.Type, "Params")
		g.generateTsStruct(f.Params, PropMutable, f.Type, "Params")
	}

	return nil
}

func (g *TypeScriptGenerator) generateTsProxy(field *Field, mutability string) {
	if field.Array {
		g.generateTsProxyArray(field, mutability)
		arrayType := "ArrayOf" + mutability + field.Type
		g.generateTsProxyReference(field, mutability, arrayType)
		return
	}

	if field.MapKey != "" {
		g.generateTsProxyMap(field, mutability)
		mapType := "Map" + field.MapKey + "To" + mutability + field.Type
		g.generateTsProxyReference(field, mutability, mapType)
	}
}

func (g *TypeScriptGenerator) generateTsProxyArray(field *Field, mutability string) {
	proxyType := mutability + field.Type
	arrayType := "ArrayOf" + proxyType
	if g.s.NewTypes[arrayType] {
		// already generated this array
		return
	}
	g.s.NewTypes[arrayType] = true

	g.printf("\nexport class %s {\n", arrayType)
	g.printf("    objID: i32;\n")

	g.printf("\n    constructor(objID: i32) {\n")
	g.printf("        this.objID = objID;\n")
	g.printf("    }\n")

	if mutability == PropMutable {
		g.printf("\n    clear(): void {\n")
		g.printf("        wasmlib.clear(this.objID);\n")
		g.printf("    }\n")
	}

	g.printf("\n    length(): i32 {\n")
	g.printf("        return wasmlib.getLength(this.objID);\n")
	g.printf("    }\n")

	if field.TypeID == 0 {
		g.generateTsProxyArrayNewType(field, proxyType)
		g.printf("}\n")
		return
	}

	// array of predefined type
	g.printf("\n    get%s(index: i32): wasmlib.Sc%s {\n", field.Type, proxyType)
	g.printf("        return new wasmlib.Sc%s(this.objID, new wasmlib.Key32(index));\n", proxyType)
	g.printf("    }\n")

	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsProxyArrayNewType(field *Field, proxyType string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := tsTypeMap
		if subtype.Array {
			varType = tsTypeIds[subtype.Type]
			if varType == "" {
				varType = tsTypeBytes
			}
			varType = g.generateTsArrayType(varType)
		}
		g.printf("\n    get%s(index: i32): sc.%s {\n", field.Type, proxyType)
		g.printf("        let subID = wasmlib.getObjectID(this.objID, new wasmlib.Key32(index), %s);\n", varType)
		g.printf("        return new sc.%s(subID);\n", proxyType)
		g.printf("    }\n")
		return
	}

	g.printf("\n    get%s(index: i32): sc.%s {\n", field.Type, proxyType)
	g.printf("        return new sc.%s(this.objID, new wasmlib.Key32(index));\n", proxyType)
	g.printf("    }\n")
}

func (g *TypeScriptGenerator) generateTsProxyMap(field *Field, mutability string) {
	proxyType := mutability + field.Type
	mapType := "Map" + field.MapKey + "To" + proxyType
	if g.s.NewTypes[mapType] {
		// already generated this map
		return
	}
	g.s.NewTypes[mapType] = true

	keyType := tsTypes[field.MapKey]
	keyValue := tsKeys[field.MapKey]

	g.printf("\nexport class %s {\n", mapType)
	g.printf("    objID: i32;\n")

	g.printf("\n    constructor(objID: i32) {\n")
	g.printf("        this.objID = objID;\n")
	g.printf("    }\n")

	if mutability == PropMutable {
		g.printf("\n    clear(): void {\n")
		g.printf("        wasmlib.clear(this.objID)\n")
		g.printf("    }\n")
	}

	if field.TypeID == 0 {
		g.generateTsProxyMapNewType(field, proxyType, keyType, keyValue)
		g.printf("}\n")
		return
	}

	// map of predefined type
	g.printf("\n    get%s(key: %s): wasmlib.Sc%s {\n", field.Type, keyType, proxyType)
	g.printf("        return new wasmlib.Sc%s(this.objID, %s.getKeyID());\n", proxyType, keyValue)
	g.printf("    }\n")

	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsProxyMapNewType(field *Field, proxyType, keyType, keyValue string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := tsTypeMap
		if subtype.Array {
			varType = tsTypeIds[subtype.Type]
			if varType == "" {
				varType = tsTypeBytes
			}
			varType = g.generateTsArrayType(varType)
		}
		g.printf("\n    get%s(key: %s): sc.%s {\n", field.Type, keyType, proxyType)
		g.printf("        let subID = wasmlib.getObjectID(this.objID, %s.getKeyID(), %s);\n", keyValue, varType)
		g.printf("        return new sc.%s(subID);\n", proxyType)
		g.printf("    }\n")
		return
	}

	g.printf("\n    get%s(key: %s): sc.%s {\n", field.Type, keyType, proxyType)
	g.printf("        return new sc.%s(this.objID, %s.getKeyID());\n", proxyType, keyValue)
	g.printf("    }\n")
}

func (g *TypeScriptGenerator) generateTsProxyReference(field *Field, mutability, typeName string) {
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\nexport class %s%s extends %s {\n};\n", mutability, field.Name, typeName)
	}
}

func (g *TypeScriptGenerator) generateTsResults() error {
	err := g.create(g.s.Folder + "results.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	for _, f := range g.s.Funcs {
		if len(f.Results) == 0 {
			continue
		}
		g.generateTsStruct(f.Results, PropImmutable, f.Type, "Results")
		g.generateTsStruct(f.Results, PropMutable, f.Type, "Results")
	}
	return nil
}

func (g *TypeScriptGenerator) generateTsState() error {
	err := g.create(g.s.Folder + "state.ts")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	g.generateTsStruct(g.s.StateVars, PropImmutable, g.s.FullName, "State")
	g.generateTsStruct(g.s.StateVars, PropMutable, g.s.FullName, "State")
	return nil
}

// TODO nested structs
func (g *TypeScriptGenerator) generateTsStruct(fields []*Field, mutability, typeName, kind string) {
	typeName = mutability + typeName + kind
	kind = strings.TrimSuffix(kind, "s")

	// first generate necessary array and map types
	for _, field := range fields {
		g.generateTsProxy(field, mutability)
	}

	g.printf("\nexport class %s extends wasmlib.ScMapID {\n", typeName)

	for _, field := range fields {
		varName := field.Name
		varID := "sc.idxMap[sc.Idx" + kind + capitalize(varName) + "]"
		if g.s.CoreContracts {
			varID = "wasmlib.Key32.fromString(sc." + kind + capitalize(varName) + ")"
		}
		varType := tsTypeIds[field.Type]
		if varType == "" {
			varType = tsTypeBytes
		}
		if field.Array {
			varType = g.generateTsArrayType(varType)
			arrayType := "ArrayOf" + mutability + field.Type
			g.printf("\n    %s(): sc.%s {\n", varName, arrayType)
			g.printf("        let arrID = wasmlib.getObjectID(this.mapID, %s, %s);\n", varID, varType)
			g.printf("        return new sc.%s(arrID)\n", arrayType)
			g.printf("    }\n")
			continue
		}
		if field.MapKey != "" {
			varType = tsTypeMap
			mapType := "Map" + field.MapKey + "To" + mutability + field.Type
			g.printf("\n    %s(): sc.%s {\n", varName, mapType)
			mapID := "this.mapID"
			if field.Alias != AliasThis {
				mapID = "mapID"
				g.printf("        let mapID = wasmlib.getObjectID(this.mapID, %s, %s);\n", varID, varType)
			}
			g.printf("        return new sc.%s(%s);\n", mapType, mapID)
			g.printf("    }\n")
			continue
		}

		proxyType := mutability + field.Type
		if field.TypeID == 0 {
			g.printf("\n    %s(): sc.%s {\n", varName, proxyType)
			g.printf("        return new sc.%s(this.mapID, %s);\n", proxyType, varID)
			g.printf("    }\n")
			continue
		}

		g.printf("\n    %s(): wasmlib.Sc%s {\n", varName, proxyType)
		g.printf("        return new wasmlib.Sc%s(this.mapID, %s);\n", proxyType, varID)
		g.printf("    }\n")
	}
	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsThunk(f *Func) {
	g.printf("\nfunction %sThunk(ctx: wasmlib.Sc%sContext): void {\n", f.FuncName, f.Kind)
	g.printf("    ctx.log(\"%s.%s\");\n", g.s.Name, f.FuncName)

	if f.Access != "" {
		g.generateTsThunkAccessCheck(f)
	}

	g.printf("    let f = new sc.%sContext();\n", f.Type)

	if len(f.Params) != 0 {
		g.printf("    f.params.mapID = wasmlib.OBJ_ID_PARAMS;\n")
	}

	if len(f.Results) != 0 {
		g.printf("    f.results.mapID = wasmlib.OBJ_ID_RESULTS;\n")
	}

	g.printf("    f.state.mapID = wasmlib.OBJ_ID_STATE;\n")

	for _, param := range f.Params {
		if !param.Optional {
			name := param.Name
			g.printf("    ctx.require(f.params.%s().exists(), \"missing mandatory %s\")\n", name, param.Name)
		}
	}

	g.printf("    sc.%s(ctx, f);\n", f.FuncName)
	g.printf("    ctx.log(\"%s.%s ok\");\n", g.s.Name, f.FuncName)
	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsThunkAccessCheck(f *Func) {
	grant := f.Access
	index := strings.Index(grant, "//")
	if index >= 0 {
		g.printf("    %s\n", grant[index:])
		grant = strings.TrimSpace(grant[:index])
	}
	switch grant {
	case AccessSelf:
		grant = "ctx.accountID()"
	case AccessChain:
		grant = "ctx.chainOwnerID()"
	case AccessCreator:
		grant = "ctx.contractCreator()"
	default:
		g.printf("    let access = ctx.state().getAgentID(wasmlib.Key32.fromString(\"%s\"));\n", grant)
		g.printf("    ctx.require(access.exists(), \"access not set: %s\");\n", grant)
		grant = "access.value()"
	}
	g.printf("    ctx.require(ctx.caller().equals(%s), \"no permission\");\n\n", grant)
}

func (g *TypeScriptGenerator) generateTsTypes() error {
	if len(g.s.Structs) == 0 {
		return nil
	}

	err := g.create(g.s.Folder + "types.ts")
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.println(tsImportWasmLib)

	for _, typeDef := range g.s.Structs {
		g.generateTsType(typeDef)
	}

	return nil
}

func (g *TypeScriptGenerator) generateTsType(typeDef *Struct) {
	nameLen, typeLen := calculatePadding(typeDef.Fields, tsTypes, false)

	g.printf("\nexport class %s {\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		fldName := pad(field.Name, nameLen)
		fldType := tsTypes[field.Type] + " = " + tsInits[field.Type] + ";"
		if field.Comment != "" {
			fldType = pad(fldType, typeLen)
		}
		g.printf("    %s: %s%s\n", fldName, fldType, field.Comment)
	}

	// write encoder and decoder for struct
	g.printf("\n    static fromBytes(bytes: u8[]): %s {\n", typeDef.Name)
	g.printf("        let decode = new wasmlib.BytesDecoder(bytes);\n")
	g.printf("        let data = new %s();\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		name := field.Name
		g.printf("        data.%s = decode.%s();\n", name, uncapitalize(field.Type))
	}
	g.printf("        decode.close();\n")
	g.printf("        return data;\n    }\n")

	g.printf("\n    bytes(): u8[] {\n")
	g.printf("        return new wasmlib.BytesEncoder().\n")
	for _, field := range typeDef.Fields {
		name := field.Name
		g.printf("            %s(this.%s).\n", uncapitalize(field.Type), name)
	}
	g.printf("            data();\n    }\n")

	g.printf("}\n")

	g.generateTsTypeProxy(typeDef, false)
	g.generateTsTypeProxy(typeDef, true)
}

func (g *TypeScriptGenerator) generateTsTypeProxy(typeDef *Struct, mutable bool) {
	typeName := PropImmutable + typeDef.Name
	if mutable {
		typeName = PropMutable + typeDef.Name
	}

	g.printf("\nexport class %s {\n", typeName)
	g.printf("    objID: i32;\n")
	g.printf("    keyID: wasmlib.Key32;\n")

	g.printf("\n    constructor(objID: i32, keyID: wasmlib.Key32) {\n")
	g.printf("        this.objID = objID;\n")
	g.printf("        this.keyID = keyID;\n")
	g.printf("    }\n")

	g.printf("\n    exists(): boolean {\n")
	g.printf("        return wasmlib.exists(this.objID, this.keyID, wasmlib.TYPE_BYTES);\n")
	g.printf("    }\n")

	if mutable {
		g.printf("\n    setValue(value: %s): void {\n", typeDef.Name)
		g.printf("        wasmlib.setBytes(this.objID, this.keyID, wasmlib.TYPE_BYTES, value.bytes());\n")
		g.printf("    }\n")
	}

	g.printf("\n    value(): %s {\n", typeDef.Name)
	g.printf("        return %s.fromBytes(wasmlib.getBytes(this.objID, this.keyID,wasmlib. TYPE_BYTES));\n", typeDef.Name)
	g.printf("    }\n")

	g.printf("}\n")
}

func (g *TypeScriptGenerator) generateTsTypeDefs() error {
	if len(g.s.Typedefs) == 0 {
		return nil
	}

	err := g.create(g.s.Folder + "typedefs.ts")
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.println(tsImportWasmLib)
	g.println(tsImportSelf)

	for _, subtype := range g.s.Typedefs {
		g.generateTsProxy(subtype, PropImmutable)
		g.generateTsProxy(subtype, PropMutable)
	}

	return nil
}
