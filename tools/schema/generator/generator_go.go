// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/tools/schema/generator/gotemplates"
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

func (g *GoGenerator) init(s *Schema) {
	g.GenBase.init(s)
	for _, template := range gotemplates.GoTemplates {
		g.addTemplates(template)
	}
	g.emitters["accessCheck"] = emitterAccessCheck
	g.emitters["funcSignature"] = emitterFuncSignature
}

func (g *GoGenerator) funcName(f *Func) string {
	return f.FuncName
}

func (g *GoGenerator) generateLanguageSpecificFiles() error {
	if g.s.CoreContracts {
		return nil
	}
	return g.createSourceFile("../main", g.writeSpecialMain)
}

func (g *GoGenerator) generateProxyArray(field *Field, mutability, arrayType, proxyType string) {
	panic("generateProxyArray")
}

func (g *GoGenerator) generateProxyMap(field *Field, mutability, mapType, proxyType string) {
	panic("generateProxyMap")
}

func (g *GoGenerator) generateProxyReference(field *Field, mutability, typeName string) {
	panic("generateProxyReference")
}

func (g *GoGenerator) writeConsts() {
	g.emit("consts.go")
}

func (g *GoGenerator) writeContract() {
	g.emit("contract.go")
}

func (g *GoGenerator) writeInitialFuncs() {
	g.emit("funcs.go")
}

func (g *GoGenerator) writeKeys() {
	g.s.KeyID = 0
	g.emit("keys.go")
}

func (g *GoGenerator) writeLib() {
	g.emit("lib.go")
}

func (g *GoGenerator) writeParams() {
	g.emit("params.go")
}

func (g *GoGenerator) writeResults() {
	g.emit("results.go")
}

func (g *GoGenerator) writeSpecialMain() {
	g.emit("main.go")
}

func (g *GoGenerator) writeState() {
	g.emit("state.go")
}

func (g *GoGenerator) writeStructs() {
	g.emit("structs.go")
}

func (g *GoGenerator) writeTypeDefs() {
	g.emit("typedefs.go")

	//for _, subtype := range g.s.Typedefs {
	//	g.generateProxy(subtype, PropImmutable)
	//	g.generateProxy(subtype, PropMutable)
	//}
}

func emitterAccessCheck(g *GenBase) {
	if g.currentFunc.Access == "" {
		return
	}
	grant := g.currentFunc.Access
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
		g.keys["grant"] = grant
		g.emit("grantForKey")
		grant = "access.Value()"
	}
	g.keys["grant"] = grant
	g.emit("grantRequire")
}

func emitterFuncSignature(g *GenBase) {
	switch g.currentFunc.FuncName {
	case SpecialFuncInit:
	case SpecialFuncSetOwner:
	case SpecialViewGetOwner:
	default:
		return
	}
	g.emit(g.currentFunc.FuncName)
}
