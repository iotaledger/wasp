// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"
	"strings"

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
	g.GenBase.init(s)
	for _, template := range tstemplates.TsTemplates {
		g.addTemplates(template)
	}
	g.emitters["accessCheck"] = emitterTsAccessCheck
}

func (g *TypeScriptGenerator) funcName(f *Func) string {
	return f.FuncName
}

func (g *TypeScriptGenerator) generateLanguageSpecificFiles() error {
	err := g.createSourceFile("index", g.writeSpecialIndex)
	if err != nil {
		return err
	}
	return g.writeSpecialConfigJSON()
}

func (g *TypeScriptGenerator) writeInitialFuncs() {
	g.emit("funcs.ts")
}

func (g *TypeScriptGenerator) writeSpecialConfigJSON() error {
	err := g.exists(g.folder + "tsconfig.json")
	if err == nil {
		// already exists
		return nil
	}

	err = g.create(g.folder + "tsconfig.json")
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

func (g *TypeScriptGenerator) writeSpecialIndex() {
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
			g.println("export * from \"./structs\";")
		}
		if len(g.s.Typedefs) != 0 {
			g.println("export * from \"./typedefs\";")
		}
	}
}

func emitterTsAccessCheck(g *GenBase) {
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
		grant = "ctx.accountID()"
	case AccessChain:
		grant = "ctx.chainOwnerID()"
	case AccessCreator:
		grant = "ctx.contractCreator()"
	default:
		g.keys["grant"] = grant
		g.emit("grantForKey")
		grant = "access.value()"
	}
	g.keys["grant"] = grant
	g.emit("grantRequire")
}

func (g *TypeScriptGenerator) setFieldKeys() {
	g.GenBase.setFieldKeys()

	fldTypeID := tsTypeIds[g.currentField.Type]
	if fldTypeID == "" {
		fldTypeID = "wasmlib.TYPE_BYTES"
	}
	g.keys["FldTypeID"] = fldTypeID
	g.keys["FldTypeKey"] = tsKeys[g.currentField.Type]
	g.keys["FldTypeInit"] = tsInits[g.currentField.Type]
	g.keys["FldLangType"] = tsTypes[g.currentField.Type]
	g.keys["FldMapKeyLangType"] = tsTypes[g.currentField.MapKey]
	g.keys["FldMapKeyKey"] = tsKeys[g.currentField.MapKey]
}

func (g *TypeScriptGenerator) setFuncKeys() {
	g.GenBase.setFuncKeys()

	initFunc := ""
	initMap := ""
	if g.currentFunc.Type == InitFunc {
		initFunc = InitFunc
		initMap = ", keyMap[:], idxMap[:]"
	}
	g.keys["initFunc"] = initFunc
	g.keys["initMap"] = initMap
}
