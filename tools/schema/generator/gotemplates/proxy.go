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

$#if fldComment _fldComment
func (s $TypeName) $FldName() ArrayOf$mut$FldType {
	return ArrayOf$mut$FldType{proxy: s.proxy.Root($Kind$FldName)}
}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `

$#if fldComment _fldComment
func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	//nolint:gosimple
	return Map$FldMapKey$+To$mut$FldType{proxy: s.proxy}
}
`,
	// *******************************
	"proxyMapOther": `

$#if fldComment _fldComment
func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	return Map$FldMapKey$+To$mut$FldType{proxy: s.proxy.Root($Kind$FldName)}
}
`,
	// *******************************
	"proxyBaseType": `

$#if fldComment _fldComment
func (s $TypeName) $FldName() wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(s.proxy.Root($Kind$FldName))
}
`,
	// *******************************
	"proxyOtherType": `

$#if fldComment _fldComment
func (s $TypeName) $FldName() $mut$FldType {
	return $mut$FldType{proxy: s.proxy.Root($Kind$FldName)}
}
`,
}
