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

func (g *GoGenerator) generateArrayType(varType string) string {
	// native core contracts use Array16 instead of our nested array type
	if g.s.CoreContracts {
		return "wasmlib.TYPE_ARRAY16|" + varType
	}
	return "wasmlib.TYPE_ARRAY|" + varType
}

func (g *GoGenerator) generateLanguageSpecificFiles() error {
	if g.s.CoreContracts {
		return nil
	}
	return g.createSourceFile("../main", g.writeSpecialMain)
}

func (g *GoGenerator) generateProxyArray(field *Field, mutability, arrayType, proxyType string) {
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
		g.generateProxyArrayNewType(field, proxyType, arrayType)
		return
	}

	// array of predefined type
	g.printf("\nfunc (a %s) Get%s(index int32) wasmlib.Sc%s {\n", arrayType, field.Type, proxyType)
	g.printf("\treturn wasmlib.NewSc%s(a.objID, wasmlib.Key32(index))\n", proxyType)
	g.printf("}\n")
}

func (g *GoGenerator) generateProxyArrayNewType(field *Field, proxyType, arrayType string) {
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
			varType = g.generateArrayType(varType)
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

func (g *GoGenerator) generateProxyMap(field *Field, mutability, mapType, proxyType string) {
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
		g.generateProxyMapNewType(field, proxyType, mapType, keyType, keyValue)
		return
	}

	// map of predefined type
	g.printf("\nfunc (m %s) Get%s(key %s) wasmlib.Sc%s {\n", mapType, field.Type, keyType, proxyType)
	g.printf("\treturn wasmlib.NewSc%s(m.objID, %s.KeyID())\n", proxyType, keyValue)
	g.printf("}\n")
}

func (g *GoGenerator) generateProxyMapNewType(field *Field, proxyType, mapType, keyType, keyValue string) {
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
			varType = g.generateArrayType(varType)
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

func (g *GoGenerator) generateProxyReference(field *Field, mutability, typeName string) {
	if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
		g.printf("\ntype %s%s = %s\n", mutability, field.Name, typeName)
	}
}

func (g *GoGenerator) generateProxyStruct(fields []*Field, mutability, typeName, kind string) {
	typeName = mutability + typeName + kind
	kind = strings.TrimSuffix(kind, "s")

	// first generate necessary array and map types
	for _, field := range fields {
		g.generateProxy(field, mutability)
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
			varType = g.generateArrayType(varType)
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

func (g *GoGenerator) generateStruct(typeDef *Struct) {
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

	g.generateStructProxy(typeDef, false)
	g.generateStructProxy(typeDef, true)
}

func (g *GoGenerator) generateStructProxy(typeDef *Struct, mutable bool) {
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
	g.emit("goHeader")

	for _, f := range g.s.Funcs {
		if len(f.Params) == 0 {
			continue
		}
		g.generateProxyStruct(f.Params, PropImmutable, f.Type, "Params")
		g.generateProxyStruct(f.Params, PropMutable, f.Type, "Params")
	}
}

func (g *GoGenerator) writeResults() {
	g.emit("goHeader")

	for _, f := range g.s.Funcs {
		if len(f.Results) == 0 {
			continue
		}
		g.generateProxyStruct(f.Results, PropImmutable, f.Type, "Results")
		g.generateProxyStruct(f.Results, PropMutable, f.Type, "Results")
	}
}

func (g *GoGenerator) writeSpecialMain() {
	g.keys["module"] = ModuleName + strings.Replace(ModuleCwd[len(ModulePath):], "\\", "/", -1)
	g.emit("main.go")
}

func (g *GoGenerator) writeState() {
	g.emit("goPackage")
	if len(g.s.StateVars) != 0 {
		g.emit("importWasmLib")
	}

	g.generateProxyStruct(g.s.StateVars, PropImmutable, g.s.FullName, "State")
	g.generateProxyStruct(g.s.StateVars, PropMutable, g.s.FullName, "State")
}

func (g *GoGenerator) writeStructs() {
	g.emit("goHeader")

	for _, typeDef := range g.s.Structs {
		g.generateStruct(typeDef)
	}
}

func (g *GoGenerator) writeTypeDefs() {
	g.emit("goHeader")

	for _, subtype := range g.s.Typedefs {
		g.generateProxy(subtype, PropImmutable)
		g.generateProxy(subtype, PropMutable)
	}
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
