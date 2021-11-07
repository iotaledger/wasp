// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"

	"github.com/iotaledger/wasp/tools/schema/generator/tstemplates"
)

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

type TypeScriptGenerator struct {
	GenBase
}

func NewTypeScriptGenerator() *TypeScriptGenerator {
	g := &TypeScriptGenerator{}
	g.extension = ".ts"
	g.funcRegexp = regexp.MustCompile(`^export function (\w+).+$`)
	g.language = "TypeScript"
	g.rootFolder = "ts"
	g.gen = g
	return g
}

func (g *TypeScriptGenerator) init(s *Schema) {
	g.GenBase.init(s, tstemplates.TsTemplates)
}

func (g *TypeScriptGenerator) generateLanguageSpecificFiles() error {
	err := g.createSourceFile("index")
	if err != nil {
		return err
	}

	tsconfig := "tsconfig.json"
	err = g.exists(g.folder + tsconfig)
	if err == nil {
		// already exists
		return nil
	}

	err = g.create(g.folder + tsconfig)
	if err != nil {
		return err
	}
	defer g.close()

	g.emit(tsconfig)
	return nil
}

func (g *TypeScriptGenerator) setFieldKeys(pad bool) {
	g.GenBase.setFieldKeys(pad)

	fldTypeID := tsTypeIds[g.currentField.Type]
	if fldTypeID == "" {
		fldTypeID = "wasmlib.TYPE_BYTES"
	}
	g.keys["fldTypeID"] = fldTypeID
	g.keys["fldTypeInit"] = tsInits[g.currentField.Type]
	g.keys["fldLangType"] = tsTypes[g.currentField.Type]
	g.keys["fldMapKeyLangType"] = tsTypes[g.currentField.MapKey]
	g.keys["fldMapKeyKey"] = tsKeys[g.currentField.MapKey]
}
