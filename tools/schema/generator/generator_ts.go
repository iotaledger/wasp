// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
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

func (s *Schema) GenerateTs() error {
	s.NewTypes = make(map[string]bool)

	err := s.generateTsConsts()
	if err != nil {
		return err
	}
	err = s.generateTsTypes()
	if err != nil {
		return err
	}
	err = s.generateTsTypeDefs()
	if err != nil {
		return err
	}
	err = s.generateTsParams()
	if err != nil {
		return err
	}
	err = s.generateTsResults()
	if err != nil {
		return err
	}
	err = s.generateTsContract()
	if err != nil {
		return err
	}

	if !s.CoreContracts {
		err = s.generateTsKeys()
		if err != nil {
			return err
		}
		err = s.generateTsState()
		if err != nil {
			return err
		}
		err = s.generateTsLib()
		if err != nil {
			return err
		}
		err = s.generateTsFuncs()
		if err != nil {
			return err
		}
	}

	// typescript-specific stuff
	err = s.generateTsConfig()
	if err != nil {
		return err
	}
	return s.generateTsIndex()
}

func (s *Schema) generateTsArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if s.CoreContracts {
		return "wasmlib.TYPE_ARRAY16|" + varType
	}
	return "wasmlib.TYPE_ARRAY|" + varType
}

func (s *Schema) generateTsConsts() error {
	file, err := os.Create(s.Folder + "consts.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)

	scName := s.Name
	if s.CoreContracts {
		// remove 'core' prefix
		scName = scName[4:]
	}
	s.appendConst("ScName", "\""+scName+"\"")
	if s.Description != "" {
		s.appendConst("ScDescription", "\""+s.Description+"\"")
	}
	hName := iscp.Hn(scName)
	hNameType := "new wasmlib.ScHname"
	s.appendConst("HScName", hNameType+"(0x"+hName.String()+")")
	s.flushTsConsts(file)

	s.generateTsConstsFields(file, s.Params, "Param")
	s.generateTsConstsFields(file, s.Results, "Result")
	s.generateTsConstsFields(file, s.StateVars, "State")

	if len(s.Funcs) != 0 {
		for _, f := range s.Funcs {
			constName := capitalize(f.FuncName)
			s.appendConst(constName, "\""+f.String+"\"")
		}
		s.flushTsConsts(file)

		for _, f := range s.Funcs {
			constHname := "H" + capitalize(f.FuncName)
			s.appendConst(constHname, hNameType+"(0x"+f.Hname.String()+")")
		}
		s.flushTsConsts(file)
	}

	return nil
}

func (s *Schema) generateTsConstsFields(file *os.File, fields []*Field, prefix string) {
	if len(fields) != 0 {
		for _, field := range fields {
			if field.Alias == AliasThis {
				continue
			}
			name := prefix + capitalize(field.Name)
			value := "\"" + field.Alias + "\""
			s.appendConst(name, value)
		}
		s.flushTsConsts(file)
	}
}

func (s *Schema) generateTsContract() error {
	file, err := os.Create(s.Folder + "contract.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	for _, f := range s.Funcs {
		kind := f.Kind
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
		}
		fmt.Fprintf(file, "\nexport class %sCall {\n", f.Type)
		fmt.Fprintf(file, "    func: wasmlib.Sc%s = new wasmlib.Sc%s(sc.HScName, sc.H%s%s);\n", kind, kind, f.Kind, f.Type)
		if len(f.Params) != 0 {
			fmt.Fprintf(file, "    params: sc.Mutable%sParams = new sc.Mutable%sParams();\n", f.Type, f.Type)
		}
		if len(f.Results) != 0 {
			fmt.Fprintf(file, "    results: sc.Immutable%sResults = new sc.Immutable%sResults();\n", f.Type, f.Type)
		}
		fmt.Fprintf(file, "}\n")

		if !s.CoreContracts {
			mutability := PropMutable
			if f.Kind == KindView {
				mutability = PropImmutable
			}
			fmt.Fprintf(file, "\nexport class %sContext {\n", f.Type)
			if len(f.Params) != 0 {
				fmt.Fprintf(file, "    params: sc.Immutable%sParams = new sc.Immutable%sParams();\n", f.Type, f.Type)
			}
			if len(f.Results) != 0 {
				fmt.Fprintf(file, "    results: sc.Mutable%sResults = new sc.Mutable%sResults();\n", f.Type, f.Type)
			}
			fmt.Fprintf(file, "    state: sc.%s%sState = new sc.%s%sState();\n", mutability, s.FullName, mutability, s.FullName)
			fmt.Fprintf(file, "}\n")
		}
	}

	s.generateTsContractFuncs(file)
	return nil
}

func (s *Schema) generateTsContractFuncs(file *os.File) {
	fmt.Fprint(file, "\nexport class ScFuncs {\n")
	for _, f := range s.Funcs {
		fmt.Fprintf(file, "\n    static %s(ctx: wasmlib.Sc%sCallContext): %sCall {\n", uncapitalize(f.Type), f.Kind, f.Type)
		fmt.Fprintf(file, "        let f = new %sCall();\n", f.Type)

		paramsID := "null"
		if len(f.Params) != 0 {
			paramsID = "f.params"
		}
		resultsID := "null"
		if len(f.Results) != 0 {
			resultsID = "f.results"
		}
		if len(f.Params) != 0 || len(f.Results) != 0 {
			fmt.Fprintf(file, "        f.func.setPtrs(%s, %s);\n", paramsID, resultsID)
		}
		fmt.Fprintf(file, "        return f;\n")
		fmt.Fprintf(file, "    }\n")
	}
	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsFuncs() error {
	scFileName := s.Folder + s.Name + ".ts"
	file, err := os.Open(scFileName)
	if err != nil {
		// generate initial code file
		return s.generateTsFuncsNew(scFileName)
	}

	// append missing function signatures to existing code file

	lines, existing, err := s.scanExistingCode(file, tsFuncRegexp)
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := s.Folder + s.Name + ".bak"
	err = os.Rename(scFileName, scOriginal)
	if err != nil {
		return err
	}
	file, err = os.Create(scFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// make copy of file
	for _, line := range lines {
		fmt.Fprintln(file, line)
	}

	// append any new funcs
	for _, f := range s.Funcs {
		if existing[f.FuncName] == "" {
			s.generateTsFuncSignature(file, f)
		}
	}

	return os.Remove(scOriginal)
}

func (s *Schema) generateTsFuncSignature(file *os.File, f *Func) {
	fmt.Fprintf(file, "\nexport function %s(ctx: wasmlib.Sc%sContext, f: sc.%sContext): void {\n", f.FuncName, f.Kind, f.Type)
	switch f.FuncName {
	case SpecialFuncInit:
		fmt.Fprintf(file, "    if (f.params.owner().exists()) {\n")
		fmt.Fprintf(file, "        f.state.owner().setValue(f.params.owner().value());\n")
		fmt.Fprintf(file, "        return;\n")
		fmt.Fprintf(file, "    }\n")
		fmt.Fprintf(file, "    f.state.owner().setValue(ctx.contractCreator());\n")
	case SpecialFuncSetOwner:
		fmt.Fprintf(file, "    f.state.owner().setValue(f.params.owner().value());\n")
	case SpecialViewGetOwner:
		fmt.Fprintf(file, "    f.results.owner().setValue(f.state.owner().value());\n")
	default:
	}
	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsFuncsNew(scFileName string) error {
	file, err := os.Create(scFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(false))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	for _, f := range s.Funcs {
		s.generateTsFuncSignature(file, f)
	}
	return nil
}

func (s *Schema) generateTsKeys() error {
	file, err := os.Create(s.Folder + "keys.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	s.KeyID = 0
	s.generateTsKeysIndexes(s.Params, "Param")
	s.generateTsKeysIndexes(s.Results, "Result")
	s.generateTsKeysIndexes(s.StateVars, "State")
	s.flushTsConsts(file)

	fmt.Fprintf(file, "\nexport let keyMap: string[] = [\n")
	s.generateTsKeysArray(file, s.Params, "Param")
	s.generateTsKeysArray(file, s.Results, "Result")
	s.generateTsKeysArray(file, s.StateVars, "State")
	fmt.Fprintf(file, "];\n")
	fmt.Fprintf(file, "\nexport let idxMap: wasmlib.Key32[] = new Array(keyMap.length);\n")
	return nil
}

func (s *Schema) generateTsKeysArray(file *os.File, fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := prefix + capitalize(field.Name)
		fmt.Fprintf(file, "    sc.%s,\n", name)
		s.KeyID++
	}
}

func (s *Schema) generateTsKeysIndexes(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := "Idx" + prefix + capitalize(field.Name)
		field.KeyID = s.KeyID
		value := strconv.Itoa(field.KeyID)
		s.KeyID++
		s.appendConst(name, value)
	}
}

func (s *Schema) generateTsLib() error {
	file, err := os.Create(s.Folder + "lib.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	fmt.Fprintf(file, "\nexport function on_call(index: i32): void {\n")
	fmt.Fprintf(file, "    return wasmlib.onCall(index);\n")
	fmt.Fprintf(file, "}\n")

	fmt.Fprintf(file, "\nexport function on_load(): void {\n")
	fmt.Fprintf(file, "    let exports = new wasmlib.ScExports();\n")
	for _, f := range s.Funcs {
		constName := capitalize(f.FuncName)
		fmt.Fprintf(file, "    exports.add%s(sc.%s, %sThunk);\n", f.Kind, constName, f.FuncName)
	}

	fmt.Fprintf(file, "\n    for (let i = 0; i < sc.keyMap.length; i++) {\n")
	fmt.Fprintf(file, "        sc.idxMap[i] = wasmlib.Key32.fromString(sc.keyMap[i]);\n")
	fmt.Fprintf(file, "    }\n")

	fmt.Fprintf(file, "}\n")

	// generate parameter structs and thunks to set up and check parameters
	for _, f := range s.Funcs {
		s.generateTsThunk(file, f)
	}
	return nil
}

func (s *Schema) generateTsProxy(file *os.File, field *Field, mutability string) {
	if field.Array {
		s.generateTsProxyArray(file, field, mutability)
		arrayType := "ArrayOf" + mutability + field.Type
		s.generateTsProxyReference(file, field, mutability, arrayType)
		return
	}

	if field.MapKey != "" {
		s.generateTsProxyMap(file, field, mutability)
		mapType := "Map" + field.MapKey + "To" + mutability + field.Type
		s.generateTsProxyReference(file, field, mutability, mapType)
	}
}

func (s *Schema) generateTsProxyArray(file *os.File, field *Field, mutability string) {
	proxyType := mutability + field.Type
	arrayType := "ArrayOf" + proxyType
	if s.NewTypes[arrayType] {
		// already generated this array
		return
	}
	s.NewTypes[arrayType] = true

	fmt.Fprintf(file, "\nexport class %s {\n", arrayType)
	fmt.Fprintf(file, "    objID: i32;\n")

	fmt.Fprintf(file, "\n    constructor(objID: i32) {\n")
	fmt.Fprintf(file, "        this.objID = objID;\n")
	fmt.Fprintf(file, "    }\n")

	if mutability == PropMutable {
		fmt.Fprintf(file, "\n    clear(): void {\n")
		fmt.Fprintf(file, "        wasmlib.clear(this.objID);\n")
		fmt.Fprintf(file, "    }\n")
	}

	fmt.Fprintf(file, "\n    length(): i32 {\n")
	fmt.Fprintf(file, "        return wasmlib.getLength(this.objID);\n")
	fmt.Fprintf(file, "    }\n")

	if field.TypeID == 0 {
		s.generateTsProxyArrayNewType(file, field, proxyType)
		fmt.Fprintf(file, "}\n")
		return
	}

	// array of predefined type
	fmt.Fprintf(file, "\n    get%s(index: i32): wasmlib.Sc%s {\n", field.Type, proxyType)
	fmt.Fprintf(file, "        return new wasmlib.Sc%s(this.objID, new wasmlib.Key32(index));\n", proxyType)
	fmt.Fprintf(file, "    }\n")

	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsProxyArrayNewType(file *os.File, field *Field, proxyType string) {
	for _, subtype := range s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := tsTypeMap
		if subtype.Array {
			varType = tsTypeIds[subtype.Type]
			if varType == "" {
				varType = tsTypeBytes
			}
			varType = s.generateTsArrayType(varType)
		}
		fmt.Fprintf(file, "\n    get%s(index: i32): sc.%s {\n", field.Type, proxyType)
		fmt.Fprintf(file, "        let subID = wasmlib.getObjectID(this.objID, new wasmlib.Key32(index), %s);\n", varType)
		fmt.Fprintf(file, "        return new sc.%s(subID);\n", proxyType)
		fmt.Fprintf(file, "    }\n")
		return
	}

	fmt.Fprintf(file, "\n    get%s(index: i32): sc.%s {\n", field.Type, proxyType)
	fmt.Fprintf(file, "        return new sc.%s(this.objID, new wasmlib.Key32(index));\n", proxyType)
	fmt.Fprintf(file, "    }\n")
}

func (s *Schema) generateTsProxyMap(file *os.File, field *Field, mutability string) {
	proxyType := mutability + field.Type
	mapType := "Map" + field.MapKey + "To" + proxyType
	if s.NewTypes[mapType] {
		// already generated this map
		return
	}
	s.NewTypes[mapType] = true

	keyType := tsTypes[field.MapKey]
	keyValue := tsKeys[field.MapKey]

	fmt.Fprintf(file, "\nexport class %s {\n", mapType)
	fmt.Fprintf(file, "    objID: i32;\n")

	fmt.Fprintf(file, "\n    constructor(objID: i32) {\n")
	fmt.Fprintf(file, "        this.objID = objID;\n")
	fmt.Fprintf(file, "    }\n")

	if mutability == PropMutable {
		fmt.Fprintf(file, "\n    clear(): void {\n")
		fmt.Fprintf(file, "        wasmlib.clear(this.objID)\n")
		fmt.Fprintf(file, "    }\n")
	}

	if field.TypeID == 0 {
		s.generateTsProxyMapNewType(file, field, proxyType, keyType, keyValue)
		fmt.Fprintf(file, "}\n")
		return
	}

	// map of predefined type
	fmt.Fprintf(file, "\n    get%s(key: %s): wasmlib.Sc%s {\n", field.Type, keyType, proxyType)
	fmt.Fprintf(file, "        return new wasmlib.Sc%s(this.objID, %s.getKeyID());\n", proxyType, keyValue)
	fmt.Fprintf(file, "    }\n")

	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsProxyMapNewType(file *os.File, field *Field, proxyType, keyType, keyValue string) {
	for _, subtype := range s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := tsTypeMap
		if subtype.Array {
			varType = tsTypeIds[subtype.Type]
			if varType == "" {
				varType = tsTypeBytes
			}
			varType = s.generateTsArrayType(varType)
		}
		fmt.Fprintf(file, "\n    get%s(key: %s): sc.%s {\n", field.Type, keyType, proxyType)
		fmt.Fprintf(file, "        let subID = wasmlib.getObjectID(this.objID, %s.getKeyID(), %s);\n", keyValue, varType)
		fmt.Fprintf(file, "        return new sc.%s(subID);\n", proxyType)
		fmt.Fprintf(file, "    }\n")
		return
	}

	fmt.Fprintf(file, "\n    get%s(key: %s): sc.%s {\n", field.Type, keyType, proxyType)
	fmt.Fprintf(file, "        return new sc.%s(this.objID, %s.getKeyID());\n", proxyType, keyValue)
	fmt.Fprintf(file, "    }\n")
}

func (s *Schema) generateTsParams() error {
	file, err := os.Create(s.Folder + "params.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	for _, f := range s.Funcs {
		if len(f.Params) == 0 {
			continue
		}
		s.generateTsStruct(file, f.Params, PropImmutable, f.Type, "Params")
		s.generateTsStruct(file, f.Params, PropMutable, f.Type, "Params")
	}

	return nil
}

func (s *Schema) generateTsResults() error {
	file, err := os.Create(s.Folder + "results.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	for _, f := range s.Funcs {
		if len(f.Results) == 0 {
			continue
		}
		s.generateTsStruct(file, f.Results, PropImmutable, f.Type, "Results")
		s.generateTsStruct(file, f.Results, PropMutable, f.Type, "Results")
	}
	return nil
}

func (s *Schema) generateTsState() error {
	file, err := os.Create(s.Folder + "state.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	s.generateTsStruct(file, s.StateVars, PropImmutable, s.FullName, "State")
	s.generateTsStruct(file, s.StateVars, PropMutable, s.FullName, "State")
	return nil
}

// TODO nested structs
func (s *Schema) generateTsStruct(file *os.File, fields []*Field, mutability, typeName, kind string) {
	typeName = mutability + typeName + kind
	kind = strings.TrimSuffix(kind, "s")

	// first generate necessary array and map types
	for _, field := range fields {
		s.generateTsProxy(file, field, mutability)
	}

	fmt.Fprintf(file, "\nexport class %s extends wasmlib.ScMapID {\n", typeName)

	for _, field := range fields {
		varName := field.Name
		varID := "sc.idxMap[sc.Idx" + kind + capitalize(varName) + "]"
		if s.CoreContracts {
			varID = "wasmlib.Key32.fromString(sc." + kind + capitalize(varName) + ")"
		}
		varType := tsTypeIds[field.Type]
		if varType == "" {
			varType = tsTypeBytes
		}
		if field.Array {
			varType = s.generateTsArrayType(varType)
			arrayType := "ArrayOf" + mutability + field.Type
			fmt.Fprintf(file, "\n    %s(): sc.%s {\n", varName, arrayType)
			fmt.Fprintf(file, "        let arrID = wasmlib.getObjectID(this.mapID, %s, %s);\n", varID, varType)
			fmt.Fprintf(file, "        return new sc.%s(arrID)\n", arrayType)
			fmt.Fprintf(file, "    }\n")
			continue
		}
		if field.MapKey != "" {
			varType = tsTypeMap
			mapType := "Map" + field.MapKey + "To" + mutability + field.Type
			fmt.Fprintf(file, "\n    %s(): sc.%s {\n", varName, mapType)
			mapID := "this.mapID"
			if field.Alias != AliasThis {
				mapID = "mapID"
				fmt.Fprintf(file, "        let mapID = wasmlib.getObjectID(this.mapID, %s, %s);\n", varID, varType)
			}
			fmt.Fprintf(file, "        return new sc.%s(%s);\n", mapType, mapID)
			fmt.Fprintf(file, "    }\n")
			continue
		}

		proxyType := mutability + field.Type
		if field.TypeID == 0 {
			fmt.Fprintf(file, "\n    %s(): sc.%s {\n", varName, proxyType)
			fmt.Fprintf(file, "        return new sc.%s(this.mapID, %s);\n", proxyType, varID)
			fmt.Fprintf(file, "    }\n")
			continue
		}

		fmt.Fprintf(file, "\n    %s(): wasmlib.Sc%s {\n", varName, proxyType)
		fmt.Fprintf(file, "        return new wasmlib.Sc%s(this.mapID, %s);\n", proxyType, varID)
		fmt.Fprintf(file, "    }\n")
	}
	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsTypeDefs() error {
	if len(s.Typedefs) == 0 {
		return nil
	}

	file, err := os.Create(s.Folder + "typedefs.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)
	fmt.Fprintln(file, tsImportSelf)

	for _, subtype := range s.Typedefs {
		s.generateTsProxy(file, subtype, PropImmutable)
		s.generateTsProxy(file, subtype, PropMutable)
	}

	return nil
}

func (s *Schema) generateTsThunk(file *os.File, f *Func) {
	fmt.Fprintf(file, "\nfunction %sThunk(ctx: wasmlib.Sc%sContext): void {\n", f.FuncName, f.Kind)
	fmt.Fprintf(file, "    ctx.log(\"%s.%s\");\n", s.Name, f.FuncName)

	if f.Access != "" {
		s.generateTsThunkAccessCheck(file, f)
	}

	fmt.Fprintf(file, "    let f = new sc.%sContext();\n", f.Type)

	if len(f.Params) != 0 {
		fmt.Fprintf(file, "    f.params.mapID = wasmlib.OBJ_ID_PARAMS;\n")
	}

	if len(f.Results) != 0 {
		fmt.Fprintf(file, "    f.results.mapID = wasmlib.OBJ_ID_RESULTS;\n")
	}

	fmt.Fprintf(file, "    f.state.mapID = wasmlib.OBJ_ID_STATE;\n")

	for _, param := range f.Params {
		if !param.Optional {
			name := param.Name
			fmt.Fprintf(file, "    ctx.require(f.params.%s().exists(), \"missing mandatory %s\")\n", name, param.Name)
		}
	}

	fmt.Fprintf(file, "    sc.%s(ctx, f);\n", f.FuncName)
	fmt.Fprintf(file, "    ctx.log(\"%s.%s ok\");\n", s.Name, f.FuncName)
	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsThunkAccessCheck(file *os.File, f *Func) {
	grant := f.Access
	index := strings.Index(grant, "//")
	if index >= 0 {
		fmt.Fprintf(file, "    %s\n", grant[index:])
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
		fmt.Fprintf(file, "    let access = ctx.state().getAgentID(wasmlib.Key32.fromString(\"%s\"));\n", grant)
		fmt.Fprintf(file, "    ctx.require(access.exists(), \"access not set: %s\");\n", grant)
		grant = "access.value()"
	}
	fmt.Fprintf(file, "    ctx.require(ctx.caller().equals(%s), \"no permission\");\n\n", grant)
}

func (s *Schema) generateTsTypes() error {
	if len(s.Structs) == 0 {
		return nil
	}

	file, err := os.Create(s.Folder + "types.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, copyright(true))
	fmt.Fprintln(file, tsImportWasmLib)

	for _, typeDef := range s.Structs {
		s.generateTsType(file, typeDef)
	}

	return nil
}

func (s *Schema) generateTsType(file *os.File, typeDef *Struct) {
	nameLen, typeLen := calculatePadding(typeDef.Fields, tsTypes, false)

	fmt.Fprintf(file, "\nexport class %s {\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		fldName := pad(field.Name, nameLen)
		fldType := tsTypes[field.Type] + " = " + tsInits[field.Type] + ";"
		if field.Comment != "" {
			fldType = pad(fldType, typeLen)
		}
		fmt.Fprintf(file, "    %s: %s%s\n", fldName, fldType, field.Comment)
	}

	// write encoder and decoder for struct
	fmt.Fprintf(file, "\n    static fromBytes(bytes: u8[]): %s {\n", typeDef.Name)
	fmt.Fprintf(file, "        let decode = new wasmlib.BytesDecoder(bytes);\n")
	fmt.Fprintf(file, "        let data = new %s();\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		name := field.Name
		fmt.Fprintf(file, "        data.%s = decode.%s();\n", name, uncapitalize(field.Type))
	}
	fmt.Fprintf(file, "        decode.close();\n")
	fmt.Fprintf(file, "        return data;\n    }\n")

	fmt.Fprintf(file, "\n    bytes(): u8[] {\n")
	fmt.Fprintf(file, "        return new wasmlib.BytesEncoder().\n")
	for _, field := range typeDef.Fields {
		name := field.Name
		fmt.Fprintf(file, "            %s(this.%s).\n", uncapitalize(field.Type), name)
	}
	fmt.Fprintf(file, "            data();\n    }\n")

	fmt.Fprintf(file, "}\n")

	s.generateTsTypeProxy(file, typeDef, false)
	s.generateTsTypeProxy(file, typeDef, true)
}

func (s *Schema) generateTsTypeProxy(file *os.File, typeDef *Struct, mutable bool) {
	typeName := PropImmutable + typeDef.Name
	if mutable {
		typeName = PropMutable + typeDef.Name
	}

	fmt.Fprintf(file, "\nexport class %s {\n", typeName)
	fmt.Fprintf(file, "    objID: i32;\n")
	fmt.Fprintf(file, "    keyID: wasmlib.Key32;\n")

	fmt.Fprintf(file, "\n    constructor(objID: i32, keyID: wasmlib.Key32) {\n")
	fmt.Fprintf(file, "        this.objID = objID;\n")
	fmt.Fprintf(file, "        this.keyID = keyID;\n")
	fmt.Fprintf(file, "    }\n")

	fmt.Fprintf(file, "\n    exists(): boolean {\n")
	fmt.Fprintf(file, "        return wasmlib.exists(this.objID, this.keyID, wasmlib.TYPE_BYTES);\n")
	fmt.Fprintf(file, "    }\n")

	if mutable {
		fmt.Fprintf(file, "\n    setValue(value: %s): void {\n", typeDef.Name)
		fmt.Fprintf(file, "        wasmlib.setBytes(this.objID, this.keyID, wasmlib.TYPE_BYTES, value.bytes());\n")
		fmt.Fprintf(file, "    }\n")
	}

	fmt.Fprintf(file, "\n    value(): %s {\n", typeDef.Name)
	fmt.Fprintf(file, "        return %s.fromBytes(wasmlib.getBytes(this.objID, this.keyID,wasmlib. TYPE_BYTES));\n", typeDef.Name)
	fmt.Fprintf(file, "    }\n")

	fmt.Fprintf(file, "}\n")
}

func (s *Schema) generateTsIndex() error {
	file, err := os.Create(s.Folder + "index.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, copyright(true))

	if !s.CoreContracts {
		fmt.Fprintf(file, "export * from \"./%s\";\n\n", s.Name)
	}

	fmt.Fprintln(file, "export * from \"./consts\";")
	fmt.Fprintln(file, "export * from \"./contract\";")
	if !s.CoreContracts {
		fmt.Fprintln(file, "export * from \"./keys\";")
		fmt.Fprintln(file, "export * from \"./lib\";")
	}
	if len(s.Params) != 0 {
		fmt.Fprintln(file, "export * from \"./params\";")
	}
	if len(s.Results) != 0 {
		fmt.Fprintln(file, "export * from \"./results\";")
	}
	if !s.CoreContracts {
		fmt.Fprintln(file, "export * from \"./state\";")
		if len(s.Structs) != 0 {
			fmt.Fprintln(file, "export * from \"./types\";")
		}
		if len(s.Typedefs) != 0 {
			fmt.Fprintln(file, "export * from \"./typedefs\";")
		}
	}
	return nil
}

func (s *Schema) flushTsConsts(file *os.File) {
	if len(s.ConstNames) == 0 {
		return
	}

	fmt.Fprintln(file)
	s.flushConsts(func(name string, value string, padLen int) {
		fmt.Fprintf(file, "export const %s = %s;\n", pad(name, padLen), value)
	})
}

func (s *Schema) generateTsConfig() error {
	file, err := os.Open(s.Folder + "tsconfig.json")
	if err == nil {
		// already exists
		file.Close()
		return nil
	}

	file, err = os.Create(s.Folder + "tsconfig.json")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"extends\": \"assemblyscript/std/assembly.json\",\n")
	fmt.Fprintf(file, "  \"include\": [\"./*.ts\"]\n")
	fmt.Fprintf(file, "}\n")

	return nil
}
