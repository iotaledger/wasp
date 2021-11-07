// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"

	"github.com/iotaledger/wasp/tools/schema/generator/rstemplates"
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
	g.GenBase.init(s, rstemplates.RsTemplates)
}

func (g *RustGenerator) funcName(f *Func) string {
	return snake(f.FuncName)
}

func (g *RustGenerator) generateLanguageSpecificFiles() error {
	if g.s.CoreContracts {
		return g.createSourceFile("mod")
	}
	return g.writeSpecialCargoToml()
}

func (g *RustGenerator) writeInitialFuncs() {
	g.emit("funcs.rs")
}

func (g *RustGenerator) writeSpecialCargoToml() error {
	cargoToml := "Cargo.toml"
	err := g.exists(cargoToml)
	if err == nil {
		// already exists
		return nil
	}

	err = g.create(cargoToml)
	if err != nil {
		return err
	}
	defer g.close()

	g.emit(cargoToml)
	return nil
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
}

func (g *RustGenerator) setFuncKeys() {
	g.GenBase.setFuncKeys()

	initFunc := ""
	if g.currentFunc.Type == InitFunc {
		initFunc = InitFunc
	}
	g.keys["initFunc"] = initFunc
}
