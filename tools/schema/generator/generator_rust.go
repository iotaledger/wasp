// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/tools/schema/generator/rstemplates"
)

const (
	allowUnusedImports = "#![allow(unused_imports)]"
	useCrate           = "use crate::*;"
	useStructs         = "use crate::structs::*;"
	useTypeDefs        = "use crate::typedefs::*;"
	useWasmLib         = "use wasmlib::*;"
	useWasmLibHost     = "use wasmlib::host::*;"
)

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
	rustTypeMap = "TYPE_MAP"
)

type RustGenerator struct {
	GenBase
}

func NewRustGenerator() *RustGenerator {
	g := &RustGenerator{}
	g.extension = ".rs"
	g.funcRegexp = regexp.MustCompile(`^pub fn (\w+).+$`)
	g.language = "Rust"
	g.rootFolder = "src"
	g.gen = g
	return g
}

func (g *RustGenerator) init(s *Schema) {
	g.GenBase.init(s)
	for _, template := range rstemplates.RsTemplates {
		g.addTemplates(template)
	}
	g.emitters["accessCheck"] = emitterRsAccessCheck
}

func (g *RustGenerator) crateOrWasmLib(withContract, withHost bool) string {
	if g.s.CoreContracts {
		retVal := useCrate
		if withContract {
			retVal += "\nuse crate::" + g.s.Name + "::*;"
		}
		if withHost {
			retVal += "\nuse crate::host::*;"
		}
		return retVal
	}
	retVal := useWasmLib
	if withHost {
		retVal += "\n" + useWasmLibHost
	}
	return retVal
}

func (g *RustGenerator) funcName(f *Func) string {
	return snake(f.FuncName)
}

func (g *RustGenerator) generateArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if g.s.CoreContracts {
		return "TYPE_ARRAY16 | " + varType
	}
	return "TYPE_ARRAY | " + varType
}

func (g *RustGenerator) generateFuncSignature(f *Func) {
	switch f.FuncName {
	case SpecialFuncInit:
		g.printf("\npub fn %s(ctx: &Sc%sContext, f: &%sContext) {\n", g.funcName(f), f.Kind, capitalize(f.Type))
		g.printf("    if f.params.owner().exists() {\n")
		g.printf("        f.state.owner().set_value(&f.params.owner().value());\n")
		g.printf("        return;\n")
		g.printf("    }\n")
		g.printf("    f.state.owner().set_value(&ctx.contract_creator());\n")
	case SpecialFuncSetOwner:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, f: &%sContext) {\n", g.funcName(f), f.Kind, capitalize(f.Type))
		g.printf("    f.state.owner().set_value(&f.params.owner().value());\n")
	case SpecialViewGetOwner:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, f: &%sContext) {\n", g.funcName(f), f.Kind, capitalize(f.Type))
		g.printf("    f.results.owner().set_value(&f.state.owner().value());\n")
	default:
		g.printf("\npub fn %s(_ctx: &Sc%sContext, _f: &%sContext) {\n", g.funcName(f), f.Kind, capitalize(f.Type))
	}
	g.printf("}\n")
}

func (g *RustGenerator) generateLanguageSpecificFiles() error {
	if g.s.CoreContracts {
		return g.createSourceFile("mod", g.writeSpecialMod)
	}
	return g.writeSpecialCargoToml()
}

func (g *RustGenerator) generateModLines(format string) {
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
			g.printf(format, "structs")
		}
		if len(g.s.Typedefs) != 0 {
			g.printf(format, "typedefs")
		}
	}
}

func (g *RustGenerator) generateProxyArray(field *Field, mutability, arrayType, proxyType string) {
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
		g.generateProxyArrayNewType(field, proxyType)
		return
	}

	// array of predefined type
	g.printf("\n    pub fn get_%s(&self, index: i32) -> Sc%s {\n", snake(field.Type), proxyType)
	g.printf("        Sc%s::new(self.obj_id, Key32(index))\n", proxyType)
	g.printf("    }\n")
}

func (g *RustGenerator) generateProxyArrayNewType(field *Field, proxyType string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := rustTypeMap
		if subtype.Array {
			varType = rustTypeIds[subtype.Type]
			if varType == "" {
				varType = "TYPE_BYTES"
			}
			varType = g.generateArrayType(varType)
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

func (g *RustGenerator) generateProxyMap(field *Field, mutability, mapType, proxyType string) {
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
		g.generateProxyMapNewType(field, proxyType, keyType, keyValue)
		return
	}

	// map of predefined type
	g.printf("\n    pub fn get_%s(&self, key: %s) -> Sc%s {\n", snake(field.Type), keyType, proxyType)
	g.printf("        Sc%s::new(self.obj_id, %s.get_key_id())\n", proxyType, keyValue)
	g.printf("    }\n")
}

func (g *RustGenerator) generateProxyMapNewType(field *Field, proxyType, keyType, keyValue string) {
	for _, subtype := range g.s.Typedefs {
		if subtype.Name != field.Type {
			continue
		}
		varType := rustTypeMap
		if subtype.Array {
			varType = rustTypeIds[subtype.Type]
			if varType == "" {
				varType = "TYPE_BYTES"
			}
			varType = g.generateArrayType(varType)
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

func (g *RustGenerator) generateProxyReference(field *Field, mutability, typeName string) {
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\npub type %s%s = %s;\n", mutability, field.Name, typeName)
	}
}

func (g *RustGenerator) writeConsts() {
	g.emit("consts.rs")
}

func (g *RustGenerator) writeContract() {
	g.emit("contract.rs")
}

func (g *RustGenerator) writeInitialFuncs() {
	g.println(useWasmLib)
	g.println()
	g.println(useCrate)
	if len(g.s.Structs) != 0 {
		g.println(useStructs)
	}
	if len(g.s.Typedefs) != 0 {
		g.println(useTypeDefs)
	}

	for _, f := range g.s.Funcs {
		g.generateFuncSignature(f)
	}
}

func (g *RustGenerator) writeKeys() {
	g.emit("keys.rs")
}

func (g *RustGenerator) writeLib() {
	g.emit("lib.rs")
}

func (g *RustGenerator) writeParams() {
	g.emit("params.rs")
}

func (g *RustGenerator) writeResults() {
	g.emit("results.rs")
}

func (g *RustGenerator) writeSpecialCargoToml() error {
	err := g.exists("Cargo.toml")
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

func (g *RustGenerator) writeSpecialMod() {
	g.println(allowUnusedImports)
	g.generateModLines("pub use %s::*;\n")
	g.generateModLines("pub mod %s;\n")
}

func (g *RustGenerator) writeState() {
	g.emit("state.rs")
}

func (g *RustGenerator) writeStructs() {
	g.emit("structs.rs")
}

func (g *RustGenerator) writeTypeDefs() {
	g.emit("typedefs.rs")
}

func emitterRsAccessCheck(g *GenBase) {
	if g.currentFunc.Access == "" {
		return
	}
	grant := g.currentFunc.Access
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
		g.keys["grant"] = grant
		g.emit("grantForKey")
		grant = "access.value()"
	}
	g.keys["grant"] = grant
	g.emit("grantRequire")
}

func (g *RustGenerator) setFieldKeys() {
	g.GenBase.setFieldKeys()

	field := g.currentField
	fldRef := "&"
	if field.Type == "Hname" || field.Type == "Int64" || field.Type == "Int32" || field.Type == "Int16" {
		fldRef = ""
	}
	g.keys["ref"] = fldRef

	fldTypeID := rustTypeIds[field.Type]
	if fldTypeID == "" {
		fldTypeID = "TYPE_BYTES"
	}
	g.keys["FldTypeID"] = fldTypeID
	g.keys["FldTypeKey"] = rustKeys[field.Type]
	g.keys["FldLangType"] = rustTypes[field.Type]
	g.keys["FldMapKeyLangType"] = rustKeyTypes[field.MapKey]
	g.keys["FldMapKeyKey"] = rustKeys[field.MapKey]

	// native core contracts use Array16 instead of our nested array type
	arrayTypeID := "TYPE_ARRAY"
	if g.s.CoreContracts {
		arrayTypeID = "TYPE_ARRAY16"
	}
	g.keys["ArrayTypeID"] = arrayTypeID
}

func (g *RustGenerator) setFuncKeys() {
	g.GenBase.setFuncKeys()

	paramsID := "ptr::null_mut()"
	if len(g.currentFunc.Params) != 0 {
		paramsID = "&mut f.params.id"
	}
	g.keys["paramsID"] = paramsID

	resultsID := "ptr::null_mut()"
	if len(g.currentFunc.Results) != 0 {
		resultsID = "&mut f.results.id"
	}
	g.keys["resultsID"] = resultsID

	initFunc := ""
	if g.currentFunc.Type == InitFunc {
		initFunc = InitFunc
	}
	g.keys["initFunc"] = initFunc
}
