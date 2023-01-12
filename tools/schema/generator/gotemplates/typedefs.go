// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var typedefsGo = map[string]string{
	// *******************************
	"typedefs.go": `
package $package

$#emit importWasmTypes
$#each typedef typedefProxy
`,
	// *******************************
	"typedefProxy": `
$#set mut Immutable
$#if array typedefProxyArray
$#if array typedefProxyAlias
$#if map typedefProxyMap
$#if map typedefProxyAlias
$#set mut Mutable
$#if array typedefProxyArray
$#if array typedefProxyAlias
$#if map typedefProxyMap
$#if map typedefProxyAlias
`,
	// *******************************
	"typedefProxyAlias": `

$#each fldComment _typedefComment
type $mut$FldName = $proxy
`,
	// *******************************
	"typedefProxyArray": `
$#set proxy ArrayOf$mut$FldType
$#if exist else typedefProxyArrayNew
`,
	// *******************************
	"typedefProxyArrayNew": `

type $proxy struct {
	Proxy wasmtypes.Proxy
}
$#if mut typedefProxyArrayMut

func (a $proxy) Length() uint32 {
	return a.Proxy.Length()
}
$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayMut": `
$#if basetype typedefProxyArrayAppendBaseType typedefProxyArrayAppendOtherType

func (a $proxy) Clear() {
	a.Proxy.ClearArray()
}
`,
	// *******************************
	"typedefProxyArrayAppendBaseType": `

func (a $proxy) Append$FldType() wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(a.Proxy.Append())
}
`,
	// *******************************
	"typedefProxyArrayAppendOtherType": `

func (a $proxy) Append$FldType() $mut$FldType {
	return $mut$FldType{Proxy: a.Proxy.Append()}
}
`,
	// *******************************
	"typedefProxyArrayNewBaseType": `

func (a $proxy) Get$FldType(index uint32) wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(a.Proxy.Index(index))
}
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `

func (a $proxy) Get$FldType(index uint32) $mut$FldType {
	return $mut$FldType{Proxy: a.Proxy.Index(index)}
}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$FldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

type $proxy struct {
	Proxy wasmtypes.Proxy
}
$#if mut typedefProxyMapMut
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapMut": `

func (m $proxy) Clear() {
	m.Proxy.ClearMap()
}
`,
	// *******************************
	"typedefProxyMapNewBaseType": `

func (m $proxy) Get$FldType(key $fldKeyLangType) wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(m.Proxy.Key(wasmtypes.$FldMapKey$+ToBytes(key)))
}
`,
	// *******************************
	"typedefProxyMapNewOtherType": `

func (m $proxy) Get$FldType(key $fldKeyLangType) $mut$FldType {
	return $mut$FldType{Proxy: m.Proxy.Key(wasmtypes.$FldMapKey$+ToBytes(key))}
}
`,
}
