// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var proxyGo = map[string]string{
	// *******************************
	"proxyContainers": `
$#if array typedefProxyArray
$#if map typedefProxyMap
`,
	// *******************************
	"proxyMethods": `

$#each fldComment _fldComment
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"proxyMethods2": `
$#if map proxyMap proxyMethods3
`,
	// *******************************
	"proxyMethods3": `
$#if basetype proxyBaseType proxyOtherType
`,
	// *******************************
	"proxyArray": `
func (s $TypeName) $FldName() ArrayOf$mut$FldType {
	return ArrayOf$mut$FldType{Proxy: s.Proxy.Root($Kind$FldName)}
}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `
func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	return Map$FldMapKey$+To$mut$FldType(s)
}
`,
	// *******************************
	"proxyMapOther": `
func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	return Map$FldMapKey$+To$mut$FldType{Proxy: s.Proxy.Root($Kind$FldName)}
}
`,
	// *******************************
	"proxyBaseType": `
func (s $TypeName) $FldName() wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(s.Proxy.Root($Kind$FldName))
}
`,
	// *******************************
	"proxyOtherType": `
func (s $TypeName) $FldName() $mut$FldType {
	return $mut$FldType{Proxy: s.Proxy.Root($Kind$FldName)}
}
`,
}
