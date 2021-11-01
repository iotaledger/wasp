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
	allowDeadCode      = "#![allow(dead_code)]"
	allowUnusedImports = "#![allow(unused_imports)]"
	useConsts          = "use crate::consts::*;"
	useCrate           = "use crate::*;"
	useKeys            = "use crate::keys::*;"
	useParams          = "use crate::params::*;"
	useResults         = "use crate::results::*;"
	useState           = "use crate::state::*;"
	useStdPtr          = "use std::ptr;"
	useTypeDefs        = "use crate::typedefs::*;"
	useTypes           = "use crate::types::*;"
	useWasmLib         = "use wasmlib::*;"
	useWasmLibHost     = "use wasmlib::host::*;"
)

var rustFuncRegexp = regexp.MustCompile(`^pub fn (\w+).+$`)

var rustTypes = StringMap{
	"Address":   "ScAddress",
	"AgentID":   "ScAgentID",
	"ChainID":   "ScChainID",
	"Color":     "ScColor",
	"Hash":      "ScHash",
	"Hname":     "ScHname",
	"Int16":     "i16",
	"Int32":     "i32",
	"Int64":     "i64",
	"RequestID": "ScRequestID",
	"String":    "String",
}

var rustKeyTypes = StringMap{
	"Address":   "&ScAddress",
	"AgentID":   "&ScAgentID",
	"ChainID":   "&ScChainID",
	"Color":     "&ScColor",
	"Hash":      "&ScHash",
	"Hname":     "&ScHname",
	"Int16":     "??TODO",
	"Int32":     "i32",
	"Int64":     "??TODO",
	"RequestID": "&ScRequestID",
	"String":    "&str",
}

var rustKeys = StringMap{
	"Address":   "key",
	"AgentID":   "key",
	"ChainID":   "key",
	"Color":     "key",
	"Hash":      "key",
	"Hname":     "key",
	"Int16":     "??TODO",
	"Int32":     "Key32(int32)",
	"Int64":     "??TODO",
	"RequestID": "key",
	"String":    "key",
}

var rustTypeIds = StringMap{
	"Address":   "TYPE_ADDRESS",
	"AgentID":   "TYPE_AGENT_ID",
	"ChainID":   "TYPE_CHAIN_ID",
	"Color":     "TYPE_COLOR",
	"Hash":      "TYPE_HASH",
	"Hname":     "TYPE_HNAME",
	"Int16":     "TYPE_INT16",
	"Int32":     "TYPE_INT32",
	"Int64":     "TYPE_INT64",
	"RequestID": "TYPE_REQUEST_ID",
	"String":    "TYPE_STRING",
}

const (
	rustTypeBytes = "TYPE_BYTES"
	rustTypeMap   = "TYPE_MAP"
)

type RustGenerator struct {
	Generator
}

func NewRustGenerator(s *Schema) *RustGenerator {
	return &RustGenerator{Generator{s: s}}
}

func (g *RustGenerator) flushRustConsts(crateOnly bool) {
	if len(g.s.ConstNames) == 0 {
		return
	}

	crate := ""
	if crateOnly {
		crate = "(crate)"
	}
	g.println()
	g.s.flushConsts(func(name string, value string, padLen int) {
		g.printf("pub%s const %s %s;\n", crate, pad(name+":", padLen+1), value)
	})
}

func (g *RustGenerator) GenerateRust() error {
	g.s.NewTypes = make(map[string]bool)

	err := g.generateRustConsts()
	if err != nil {
		return err
	}
	err = g.generateRustTypes()
	if err != nil {
		return err
	}
	err = g.generateRustTypeDefs()
	if err != nil {
		return err
	}
	err = g.generateRustParams()
	if err != nil {
		return err
	}
	err = g.generateRustResults()
	if err != nil {
		return err
	}
	err = g.generateRustContract()
	if err != nil {
		return err
	}

	if !g.s.CoreContracts {
		err = g.generateRustKeys()
		if err != nil {
			return err
		}
		err = g.generateRustState()
		if err != nil {
			return err
		}
		err = g.generateRustLib()
		if err != nil {
			return err
		}
		err = g.generateRustFuncs()
		if err != nil {
			return err
		}

		// rust-specific stuff
		return g.generateRustCargo()
	}

	return g.generateRustMod()
}

func (g *RustGenerator) generateRustArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if g.s.CoreContracts {
		return "TYPE_ARRAY16 | " + varType
	}
	return "TYPE_ARRAY | " + varType
}

func (g *RustGenerator) generateRustCargo() error {
	_, err := os.Stat("Cargo.toml")
	if err == nil {
		// already exists
		return nil
	}

	err = g.create("Cargo.toml")
	if err != nil {
		return err
	}
	defer g.close()

	g.printf("[package]\n")
	g.printf("name = \"%s\"\n", g.s.Name)
	g.printf("description = \"%s\"\n", g.s.Description)
	g.printf("license = \"Apache-2.0\"\n")
	g.printf("version = \"0.1.0\"\n")
	g.printf("authors = [\"John Doe <john@doe.org>\"]\n")
	g.printf("edition = \"2018\"\n")
	g.printf("repository = \"https://%s\"\n", ModuleName)
	g.printf("\n[lib]\n")
	g.printf("crate-type = [\"cdylib\", \"rlib\"]\n")
	g.printf("\n[features]\n")
	g.printf("default = [\"console_error_panic_hook\"]\n")
	g.printf("\n[dependencies]\n")
	g.printf("wasmlib = { git = \"https://github.com/iotaledger/wasp\", branch = \"develop\" }\n")
	g.printf("console_error_panic_hook = { version = \"0.1.6\", optional = true }\n")
	g.printf("wee_alloc = { version = \"0.4.5\", optional = true }\n")
	g.printf("\n[dev-dependencies]\n")
	g.printf("wasm-bindgen-test = \"0.3.13\"\n")

	return nil
}

func (g *RustGenerator) generateRustConsts() error {
	err := g.create(g.s.Folder + "consts.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println()
	g.println(g.s.crateOrWasmLib(false, false))

	scName := g.s.Name
	if g.s.CoreContracts {
		// remove 'core' prefix
		scName = scName[4:]
	}
	g.s.appendConst("SC_NAME", "&str = \""+scName+"\"")
	if g.s.Description != "" {
		g.s.appendConst("SC_DESCRIPTION", "&str = \""+g.s.Description+"\"")
	}
	hName := iscp.Hn(scName)
	g.s.appendConst("HSC_NAME", "ScHname = ScHname(0x"+hName.String()+")")
	g.flushRustConsts(false)

	g.generateRustConstsFields(g.s.Params, "PARAM_")
	g.generateRustConstsFields(g.s.Results, "RESULT_")
	g.generateRustConstsFields(g.s.StateVars, "STATE_")

	if len(g.s.Funcs) != 0 {
		for _, f := range g.s.Funcs {
			constName := upper(snake(f.FuncName))
			g.s.appendConst(constName, "&str = \""+f.String+"\"")
		}
		g.flushRustConsts(g.s.CoreContracts)

		for _, f := range g.s.Funcs {
			constHname := "H" + upper(snake(f.FuncName))
			g.s.appendConst(constHname, "ScHname = ScHname(0x"+f.Hname.String()+")")
		}
		g.flushRustConsts(g.s.CoreContracts)
	}

	g.formatter(true)
	return nil
}

func (g *RustGenerator) generateRustConstsFields(fields []*Field, prefix string) {
	if len(fields) != 0 {
		for _, field := range fields {
			if field.Alias == AliasThis {
				continue
			}
			name := prefix + upper(snake(field.Name))
			value := "&str = \"" + field.Alias + "\""
			g.s.appendConst(name, value)
		}
		g.flushRustConsts(g.s.CoreContracts)
	}
}

func (g *RustGenerator) generateRustContract() error {
	err := g.create(g.s.Folder + "contract.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println()
	g.println(useStdPtr)
	g.println()
	g.println(g.s.crateOrWasmLib(true, false))
	if !g.s.CoreContracts {
		g.println()
		g.println(useConsts)
		g.println(useParams)
		g.println(useResults)
	}

	for _, f := range g.s.Funcs {
		nameLen := f.nameLen(4) + 1
		kind := f.Kind
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
		}
		g.printf("\npub struct %sCall {\n", f.Type)
		g.printf("    pub %s Sc%s,\n", pad("func:", nameLen), kind)
		if len(f.Params) != 0 {
			g.printf("    pub %s Mutable%sParams,\n", pad("params:", nameLen), f.Type)
		}
		if len(f.Results) != 0 {
			g.printf("    pub results: Immutable%sResults,\n", f.Type)
		}
		g.printf("}\n")
	}

	g.generateRustContractFuncs()
	g.formatter(true)
	return nil
}

func (g *RustGenerator) generateRustContractFuncs() {
	g.println("\npub struct ScFuncs {")
	g.println("}")
	g.println("\nimpl ScFuncs {")

	for _, f := range g.s.Funcs {
		nameLen := f.nameLen(4) + 1
		funcName := snake(f.FuncName)
		constName := upper(funcName)
		letMut := ""
		if len(f.Params) != 0 || len(f.Results) != 0 {
			letMut = "let mut f = "
		}
		kind := f.Kind
		if f.Type == InitFunc {
			kind = f.Type + f.Kind
		}
		g.printf("    pub fn %s(_ctx: & dyn Sc%sCallContext) -> %sCall {\n", funcName[5:], f.Kind, f.Type)
		g.printf("        %s%sCall {\n", letMut, f.Type)
		g.printf("            %s Sc%s::new(HSC_NAME, H%s),\n", pad("func:", nameLen), kind, constName)
		paramsID := "ptr::null_mut()"
		if len(f.Params) != 0 {
			paramsID = "&mut f.params.id"
			g.printf("            %s Mutable%sParams { id: 0 },\n", pad("params:", nameLen), f.Type)
		}
		resultsID := "ptr::null_mut()"
		if len(f.Results) != 0 {
			resultsID = "&mut f.results.id"
			g.printf("            results: Immutable%sResults { id: 0 },\n", f.Type)
		}
		g.printf("        }")
		if len(f.Params) != 0 || len(f.Results) != 0 {
			g.printf(";\n")
			g.printf("        f.func.set_ptrs(%s, %s);\n", paramsID, resultsID)
			g.printf("        f")
		}
		g.printf("\n    }\n")
	}
	g.printf("}\n")
}

func (g *RustGenerator) generateRustFuncs() error {
	scFileName := g.s.Folder + g.s.Name + ".rs"
	err := g.open(scFileName)
	if err != nil {
		// generate initial code file
		return g.generateRustFuncsNew(scFileName)
	}

	// append missing function signatures to existing code file

	// scan existing file for signatures
	lines, existing, err := g.scanExistingCode(rustFuncRegexp)
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
		name := snake(f.FuncName)
		if existing[name] == "" {
			g.generateRustFuncSignature(f)
		}
	}

	return os.Remove(scOriginal)
}

func (g *RustGenerator) generateRustFuncSignature(f *Func) {
	switch f.FuncName {
	case SpecialFuncInit:
		g.printf("\npub fn %s(ctx: &Sc%sContext, f: &%sContext) {\n", snake(f.FuncName), f.Kind, capitalize(f.Type))
		g.printf("    if f.params.owner().exists() {\n")
		g.printf("        f.state.owner().set_value(&f.params.owner().value());\n")
		g.printf("        return;\n")
		g.printf("    }\n")
		g.printf("    f.state.owner().set_value(&ctx.contract_creator());\n")
	case SpecialFuncSetOwner:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, f: &%sContext) {\n", snake(f.FuncName), f.Kind, capitalize(f.Type))
		g.printf("    f.state.owner().set_value(&f.params.owner().value());\n")
	case SpecialViewGetOwner:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, f: &%sContext) {\n", snake(f.FuncName), f.Kind, capitalize(f.Type))
		g.printf("    f.results.owner().set_value(&f.state.owner().value());\n")
	default:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, _f: &%sContext) {\n", snake(f.FuncName), f.Kind, capitalize(f.Type))
	}
	g.printf("}\n")
}

func (g *RustGenerator) generateRustFuncsNew(scFileName string) error {
	err := g.create(scFileName)
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(false))
	g.println(useWasmLib)
	g.println()
	g.println(useCrate)
	if len(g.s.Typedefs) != 0 {
		g.println(useTypeDefs)
	}
	if len(g.s.Structs) != 0 {
		g.println(useTypes)
	}

	for _, f := range g.s.Funcs {
		g.generateRustFuncSignature(f)
	}
	return nil
}

func (g *RustGenerator) generateRustKeys() error {
	err := g.create(g.s.Folder + "keys.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println()
	g.println(useWasmLib)
	g.println()
	g.println(useCrate)

	g.s.KeyID = 0
	g.generateRustKeysIndexes(g.s.Params, "PARAM_")
	g.generateRustKeysIndexes(g.s.Results, "RESULT_")
	g.generateRustKeysIndexes(g.s.StateVars, "STATE_")
	g.flushRustConsts(true)

	size := g.s.KeyID
	g.printf("\npub const KEY_MAP_LEN: usize = %d;\n", size)
	g.printf("\npub const KEY_MAP: [&str; KEY_MAP_LEN] = [\n")
	g.generateRustKeysArray(g.s.Params, "PARAM_")
	g.generateRustKeysArray(g.s.Results, "RESULT_")
	g.generateRustKeysArray(g.s.StateVars, "STATE_")
	g.printf("];\n")

	g.printf("\npub static mut IDX_MAP: [Key32; KEY_MAP_LEN] = [Key32(0); KEY_MAP_LEN];\n")

	g.printf("\npub fn idx_map(idx: usize) -> Key32 {\n")
	g.printf("    unsafe {\n")
	g.printf("        IDX_MAP[idx]\n")
	g.printf("    }\n")
	g.printf("}\n")

	g.formatter(true)
	return nil
}

func (g *RustGenerator) generateRustKeysArray(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := prefix + upper(snake(field.Name))
		g.printf("    %s,\n", name)
		g.s.KeyID++
	}
}

func (g *RustGenerator) generateRustKeysIndexes(fields []*Field, prefix string) {
	for _, field := range fields {
		if field.Alias == AliasThis {
			continue
		}
		name := "IDX_" + prefix + upper(snake(field.Name))
		field.KeyID = g.s.KeyID
		value := "usize = " + strconv.Itoa(field.KeyID)
		g.s.KeyID++
		g.s.appendConst(name, value)
	}
}

func (g *RustGenerator) generateRustLib() error {
	err := g.create(g.s.Folder + "lib.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println(allowUnusedImports)
	g.println()
	g.printf("use %s::*;\n", g.s.Name)
	g.println(useWasmLib)
	g.println(useWasmLibHost)
	g.println()
	g.println(useConsts)
	g.println(useKeys)
	g.println(useParams)
	g.println(useResults)
	g.println(useState)
	g.println()

	g.println("mod consts;")
	g.println("mod contract;")
	g.println("mod keys;")
	g.println("mod params;")
	g.println("mod results;")
	g.println("mod state;")
	if len(g.s.Typedefs) != 0 {
		g.println("mod typedefs;")
	}
	if len(g.s.Structs) != 0 {
		g.println("mod types;")
	}
	g.printf("mod %s;\n", g.s.Name)

	g.println("\n#[no_mangle]")
	g.println("fn on_load() {")
	if len(g.s.Funcs) != 0 {
		g.printf("    let exports = ScExports::new();\n")
	}
	for _, f := range g.s.Funcs {
		name := snake(f.FuncName)
		g.printf("    exports.add_%s(%s, %s_thunk);\n", lower(f.Kind), upper(name), name)
	}

	g.printf("\n    unsafe {\n")
	g.printf("        for i in 0..KEY_MAP_LEN {\n")
	g.printf("            IDX_MAP[i] = get_key_id_from_string(KEY_MAP[i]);\n")
	g.printf("        }\n")
	g.printf("    }\n")

	g.printf("}\n")

	// generate parameter structs and thunks to set up and check parameters
	for _, f := range g.s.Funcs {
		g.generateRustThunk(f)
	}

	g.formatter(true)
	return nil
}

func (g *RustGenerator) generateRustMod() error {
	err := g.create(g.s.Folder + "mod.rs")
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.println(allowUnusedImports)
	err = g.generateRustModLines("pub use %s::*;\n")
	if err != nil {
		return err
	}
	return g.generateRustModLines("pub mod %s;\n")
}

func (g *RustGenerator) generateRustModLines(format string) error {
	g.println()

	if !g.s.CoreContracts {
		g.printf(format, g.s.Name)
		g.println()
	}

	g.printf(format, "consts")
	g.printf(format, "contract")
	if !g.s.CoreContracts {
		g.printf(format, "keys")
		g.printf(format, "lib")
	}
	if len(g.s.Params) != 0 {
		g.printf(format, "params")
	}
	if len(g.s.Results) != 0 {
		g.printf(format, "results")
	}
	if !g.s.CoreContracts {
		g.printf(format, "state")
		if len(g.s.Structs) != 0 {
			g.printf(format, "types")
		}
		if len(g.s.Typedefs) != 0 {
			g.printf(format, "typedefs")
		}
	}
	return nil
}

func (g *RustGenerator) generateRustParams() error {
	err := g.create(g.s.Folder + "params.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(allowDeadCode)
	g.println(allowUnusedImports)
	g.println()
	g.println(g.s.crateOrWasmLib(true, true))
	if !g.s.CoreContracts {
		g.println()
		g.println(useCrate)
		g.println(useKeys)
	}

	for _, f := range g.s.Funcs {
		params := make([]*Field, 0, len(f.Params))
		for _, param := range f.Params {
			if param.Alias != "@" {
				params = append(params, param)
			}
		}
		if len(params) == 0 {
			continue
		}
		g.generateRustStruct(params, PropImmutable, f.Type, "Params")
		g.generateRustStruct(params, PropMutable, f.Type, "Params")
	}
	return nil
}

func (g *RustGenerator) generateRustProxy(field *Field, mutability string) {
	if field.Array {
		g.generateRustProxyArray(field, mutability)
		return
	}

	if field.MapKey != "" {
		g.generateRustProxyMap(field, mutability)
	}
}

func (g *RustGenerator) generateRustProxyArray(field *Field, mutability string) {
	proxyType := mutability + field.Type
	arrayType := "ArrayOf" + proxyType
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\npub type %s%s = %s;\n", mutability, field.Name, arrayType)
	}
	if g.s.NewTypes[arrayType] {
		// already generated this array
		return
	}
	g.s.NewTypes[arrayType] = true

	g.printf("\npub struct %s {\n", arrayType)
	g.printf("    pub(crate) obj_id: i32,\n")
	g.printf("}\n")

	g.printf("\nimpl %s {", arrayType)
	defer g.printf("}\n")

	if mutability == PropMutable {
		g.printf("\n    pub fn clear(&self) {\n")
		g.printf("        clear(self.obj_id);\n")
		g.printf("    }\n")
	}

	g.printf("\n    pub fn length(&self) -> i32 {\n")
	g.printf("        get_length(self.obj_id)\n")
	g.printf("    }\n")

	if field.TypeID == 0 {
		g.generateRustProxyArrayNewType(field, proxyType)
		return
	}

	// array of predefined type
	g.printf("\n    pub fn get_%s(&self, index: i32) -> Sc%s {\n", snake(field.Type), proxyType)
	g.printf("        Sc%s::new(self.obj_id, Key32(index))\n", proxyType)
	g.printf("    }\n")
}

func (g *RustGenerator) generateRustProxyArrayNewType(field *Field, proxyType string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := rustTypeMap
		if subtype.Array {
			varType = rustTypeIds[subtype.Type]
			if varType == "" {
				varType = rustTypeBytes
			}
			varType = g.generateRustArrayType(varType)
		}
		g.printf("\n    pub fn get_%s(&self, index: i32) -> %s {\n", snake(field.Type), proxyType)
		g.printf("        let sub_id = get_object_id(self.obj_id, Key32(index), %s)\n", varType)
		g.printf("        %s { obj_id: sub_id }\n", proxyType)
		g.printf("    }\n")
		return
	}

	g.printf("\n    pub fn get_%s(&self, index: i32) -> %s {\n", snake(field.Type), proxyType)
	g.printf("        %s { obj_id: self.obj_id, key_id: Key32(index) }\n", proxyType)
	g.printf("    }\n")
}

func (g *RustGenerator) generateRustProxyMap(field *Field, mutability string) {
	proxyType := mutability + field.Type
	mapType := "Map" + field.MapKey + "To" + proxyType
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\npub type %s%s = %s;\n", mutability, field.Name, mapType)
	}
	if g.s.NewTypes[mapType] {
		// already generated this map
		return
	}
	g.s.NewTypes[mapType] = true

	keyType := rustKeyTypes[field.MapKey]
	keyValue := rustKeys[field.MapKey]

	g.printf("\npub struct %s {\n", mapType)
	g.printf("    pub(crate) obj_id: i32,\n")
	g.printf("}\n")

	g.printf("\nimpl %s {", mapType)
	defer g.printf("}\n")

	if mutability == PropMutable {
		g.printf("\n    pub fn clear(&self) {\n")
		g.printf("        clear(self.obj_id)\n")
		g.printf("    }\n")
	}

	if field.TypeID == 0 {
		g.generateRustProxyMapNewType(field, proxyType, keyType, keyValue)
		return
	}

	// map of predefined type
	g.printf("\n    pub fn get_%s(&self, key: %s) -> Sc%s {\n", snake(field.Type), keyType, proxyType)
	g.printf("        Sc%s::new(self.obj_id, %s.get_key_id())\n", proxyType, keyValue)
	g.printf("    }\n")
}

func (g *RustGenerator) generateRustProxyMapNewType(field *Field, proxyType, keyType, keyValue string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := rustTypeMap
		if subtype.Array {
			varType = rustTypeIds[subtype.Type]
			if varType == "" {
				varType = rustTypeBytes
			}
			varType = g.generateRustArrayType(varType)
		}
		g.printf("\n    pub fn get_%s(&self, key: %s) -> %s {\n", snake(field.Type), keyType, proxyType)
		g.printf("        let sub_id = get_object_id(self.obj_id, %s.get_key_id(), %s);\n", keyValue, varType)
		g.printf("        %s { obj_id: sub_id }\n", proxyType)
		g.printf("    }\n")
		return
	}

	g.printf("\n    pub fn get_%s(&self, key: %s) -> %s {\n", snake(field.Type), keyType, proxyType)
	g.printf("        %s { obj_id: self.obj_id, key_id: %s.get_key_id() }\n", proxyType, keyValue)
	g.printf("    }\n")
}

func (g *RustGenerator) generateRustResults() error {
	err := g.create(g.s.Folder + "results.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(allowDeadCode)
	g.println(allowUnusedImports)
	g.println()
	g.println(g.s.crateOrWasmLib(true, true))
	if !g.s.CoreContracts {
		g.println()
		g.println(useCrate)
		g.println(useKeys)
		if len(g.s.Structs) != 0 {
			g.println(useTypes)
		}
	}

	for _, f := range g.s.Funcs {
		if len(f.Results) == 0 {
			continue
		}
		g.generateRustStruct(f.Results, PropImmutable, f.Type, "Results")
		g.generateRustStruct(f.Results, PropMutable, f.Type, "Results")
	}
	return nil
}

func (g *RustGenerator) generateRustState() error {
	err := g.create(g.s.Folder + "state.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.println(allowDeadCode)
	g.println(allowUnusedImports)
	g.println()
	g.println(useWasmLib)
	g.println(useWasmLibHost)
	g.println()
	g.println(useCrate)
	g.println(useKeys)
	if len(g.s.Typedefs) != 0 {
		g.println(useTypeDefs)
	}
	if len(g.s.Structs) != 0 {
		g.println(useTypes)
	}

	g.generateRustStruct(g.s.StateVars, PropImmutable, g.s.FullName, "State")
	g.generateRustStruct(g.s.StateVars, PropMutable, g.s.FullName, "State")
	return nil
}

func (g *RustGenerator) generateRustStruct(fields []*Field, mutability, typeName, kind string) {
	typeName = mutability + typeName + kind
	kind = strings.TrimSuffix(kind, "s")
	kind = upper(kind) + "_"

	// first generate necessary array and map types
	for _, field := range fields {
		g.generateRustProxy(field, mutability)
	}

	g.printf("\n#[derive(Clone, Copy)]\n")
	g.printf("pub struct %s {\n", typeName)
	g.printf("    pub(crate) id: i32,\n")
	g.printf("}\n")

	if len(fields) != 0 {
		g.printf("\nimpl %s {", typeName)
		defer g.printf("}\n")
	}

	for _, field := range fields {
		varName := snake(field.Name)
		varID := "idx_map(IDX_" + kind + upper(varName) + ")"
		if g.s.CoreContracts {
			varID = kind + upper(varName) + ".get_key_id()"
		}
		varType := rustTypeIds[field.Type]
		if varType == "" {
			varType = rustTypeBytes
		}
		if field.Array {
			varType = g.generateRustArrayType(varType)
			arrayType := "ArrayOf" + mutability + field.Type
			g.printf("\n    pub fn %s(&self) -> %s {\n", varName, arrayType)
			g.printf("        let arr_id = get_object_id(self.id, %s, %s);\n", varID, varType)
			g.printf("        %s { obj_id: arr_id }\n", arrayType)
			g.printf("    }\n")
			continue
		}
		if field.MapKey != "" {
			varType = rustTypeMap
			mapType := "Map" + field.MapKey + "To" + mutability + field.Type
			g.printf("\n    pub fn %s(&self) -> %s {\n", varName, mapType)
			mapID := "self.id"
			if field.Alias != AliasThis {
				mapID = "map_id"
				g.printf("        let map_id = get_object_id(self.id, %s, %s);\n", varID, varType)
			}
			g.printf("        %s { obj_id: %s }\n", mapType, mapID)
			g.printf("    }\n")
			continue
		}

		proxyType := mutability + field.Type
		if field.TypeID == 0 {
			g.printf("\n    pub fn %s(&self) -> %s {\n", varName, proxyType)
			g.printf("        %s { obj_id: self.id, key_id: %s }\n", proxyType, varID)
			g.printf("    }\n")
			continue
		}

		g.printf("\n    pub fn %s(&self) -> Sc%s {\n", varName, proxyType)
		g.printf("        Sc%s::new(self.id, %s)\n", proxyType, varID)
		g.printf("    }\n")
	}
}

func (g *RustGenerator) generateRustThunk(f *Func) {
	nameLen := f.nameLen(5) + 1
	mutability := PropMutable
	if f.Kind == KindView {
		mutability = PropImmutable
	}
	g.printf("\npub struct %sContext {\n", f.Type)
	if len(f.Params) != 0 {
		g.printf("    %s Immutable%sParams,\n", pad("params:", nameLen), f.Type)
	}
	if len(f.Results) != 0 {
		g.printf("    results: Mutable%sResults,\n", f.Type)
	}
	g.printf("    %s %s%sState,\n", pad("state:", nameLen), mutability, g.s.FullName)
	g.printf("}\n")

	g.printf("\nfn %s_thunk(ctx: &Sc%sContext) {\n", snake(f.FuncName), f.Kind)
	g.printf("    ctx.log(\"%s.%s\");\n", g.s.Name, f.FuncName)

	if f.Access != "" {
		g.generateRustThunkAccessCheck(f)
	}

	g.printf("    let f = %sContext {\n", f.Type)

	if len(f.Params) != 0 {
		g.printf("        params: Immutable%sParams {\n", f.Type)
		g.printf("            id: OBJ_ID_PARAMS,\n")
		g.printf("        },\n")
	}

	if len(f.Results) != 0 {
		g.printf("        results: Mutable%sResults {\n", f.Type)
		g.printf("            id: OBJ_ID_RESULTS,\n")
		g.printf("        },\n")
	}

	g.printf("        state: %s%sState {\n", mutability, g.s.FullName)
	g.printf("            id: OBJ_ID_STATE,\n")
	g.printf("        },\n")

	g.printf("    };\n")

	for _, param := range f.Params {
		if !param.Optional {
			name := snake(param.Name)
			g.printf("    ctx.require(f.params.%s().exists(), \"missing mandatory %s\");\n", name, param.Name)
		}
	}

	g.printf("    %s(ctx, &f);\n", snake(f.FuncName))
	g.printf("    ctx.log(\"%s.%s ok\");\n", g.s.Name, f.FuncName)
	g.printf("}\n")
}

func (g *RustGenerator) generateRustThunkAccessCheck(f *Func) {
	grant := f.Access
	index := strings.Index(grant, "//")
	if index >= 0 {
		g.printf("    %s\n", grant[index:])
		grant = strings.TrimSpace(grant[:index])
	}
	switch grant {
	case AccessSelf:
		grant = "ctx.account_id()"
	case AccessChain:
		grant = "ctx.chain_owner_id()"
	case AccessCreator:
		grant = "ctx.contract_creator()"
	default:
		g.printf("    let access = ctx.state().get_agent_id(\"%s\");\n", grant)
		g.printf("    ctx.require(access.exists(), \"access not set: %s\");\n", grant)
		grant = "access.value()"
	}
	g.printf("    ctx.require(ctx.caller() == %s, \"no permission\");\n\n", grant)
}

func (g *RustGenerator) generateRustTypes() error {
	if len(g.s.Structs) == 0 {
		return nil
	}

	err := g.create(g.s.Folder + "types.rs")
	if err != nil {
		return err
	}
	defer g.close()

	// write file header
	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println()
	g.println(useWasmLib)
	g.println(useWasmLibHost)

	// write structs
	for _, typeDef := range g.s.Structs {
		g.generateRustType(typeDef)
	}

	g.formatter(true)
	return nil
}

func (g *RustGenerator) generateRustType(typeDef *Struct) {
	nameLen, typeLen := calculatePadding(typeDef.Fields, rustTypes, true)

	g.printf("\npub struct %s {\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		fldName := pad(snake(field.Name)+":", nameLen+1)
		fldType := rustTypes[field.Type] + ","
		if field.Comment != "" {
			fldType = pad(fldType, typeLen+1)
		}
		g.printf("    pub %s %s%s\n", fldName, fldType, field.Comment)
	}
	g.printf("}\n")

	// write encoder and decoder for struct
	g.printf("\nimpl %s {", typeDef.Name)

	g.printf("\n    pub fn from_bytes(bytes: &[u8]) -> %s {\n", typeDef.Name)
	g.printf("        let mut decode = BytesDecoder::new(bytes);\n")
	g.printf("        %s {\n", typeDef.Name)
	for _, field := range typeDef.Fields {
		name := snake(field.Name)
		g.printf("            %s: decode.%s(),\n", name, snake(field.Type))
	}
	g.printf("        }\n")
	g.printf("    }\n")

	g.printf("\n    pub fn to_bytes(&self) -> Vec<u8> {\n")
	g.printf("        let mut encode = BytesEncoder::new();\n")
	for _, field := range typeDef.Fields {
		name := snake(field.Name)
		ref := "&"
		if field.Type == "Hname" || field.Type == "Int64" || field.Type == "Int32" || field.Type == "Int16" {
			ref = ""
		}
		g.printf("        encode.%s(%sself.%s);\n", snake(field.Type), ref, name)
	}
	g.printf("        return encode.data();\n")
	g.printf("    }\n")
	g.printf("}\n")

	g.generateRustTypeProxy(typeDef, false)
	g.generateRustTypeProxy(typeDef, true)
}

func (g *RustGenerator) generateRustTypeProxy(typeDef *Struct, mutable bool) {
	typeName := PropImmutable + typeDef.Name
	if mutable {
		typeName = PropMutable + typeDef.Name
	}

	g.printf("\npub struct %s {\n", typeName)
	g.printf("    pub(crate) obj_id: i32,\n")
	g.printf("    pub(crate) key_id: Key32,\n")
	g.printf("}\n")

	g.printf("\nimpl %s {", typeName)

	g.printf("\n    pub fn exists(&self) -> bool {\n")
	g.printf("        exists(self.obj_id, self.key_id, TYPE_BYTES)\n")
	g.printf("    }\n")

	if mutable {
		g.printf("\n    pub fn set_value(&self, value: &%s) {\n", typeDef.Name)
		g.printf("        set_bytes(self.obj_id, self.key_id, TYPE_BYTES, &value.to_bytes());\n")
		g.printf("    }\n")
	}

	g.printf("\n    pub fn value(&self) -> %s {\n", typeDef.Name)
	g.printf("        %s::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))\n", typeDef.Name)
	g.printf("    }\n")

	g.printf("}\n")
}

func (g *RustGenerator) generateRustTypeDefs() error {
	if len(g.s.Typedefs) == 0 {
		return nil
	}

	err := g.create(g.s.Folder + "typedefs.rs")
	if err != nil {
		return err
	}
	defer g.close()

	g.println(copyright(true))
	g.formatter(false)
	g.println(allowDeadCode)
	g.println()
	g.println(useWasmLib)
	g.println(useWasmLibHost)
	if len(g.s.Structs) != 0 {
		g.println()
		g.println(useTypes)
	}

	for _, subtype := range g.s.Typedefs {
		g.generateRustProxy(subtype, PropImmutable)
		g.generateRustProxy(subtype, PropMutable)
	}

	g.formatter(true)
	return nil
}
