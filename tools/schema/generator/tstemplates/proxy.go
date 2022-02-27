// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var proxyTs = map[string]string{
	// *******************************
	"proxyContainers": `
$#if array typedefProxyArray
$#if map typedefProxyMap
`,
	// *******************************
	"proxyMethods": `
$#if separator newline
$#set separator $true
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
	$fldName(): sc.ArrayOf$mut$FldType {
		return new sc.ArrayOf$mut$FldType(this.proxy.root(sc.$Kind$FldName));
	}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `
	$fldName(): sc.Map$FldMapKey$+To$mut$FldType {
		return new sc.Map$FldMapKey$+To$mut$FldType(this.proxy);
	}
`,
	// *******************************
	"proxyMapOther": `
	$fldName(): sc.Map$FldMapKey$+To$mut$FldType {
		return new sc.Map$FldMapKey$+To$mut$FldType(this.proxy.root(sc.$Kind$FldName));
	}
`,
	// *******************************
	"proxyBaseType": `
	$fldName(): wasmtypes.Sc$mut$FldType {
		return new wasmtypes.Sc$mut$FldType(this.proxy.root(sc.$Kind$FldName));
	}
`,
	// *******************************
	"proxyOtherType": `
	$fldName(): sc.$mut$FldType {
		return new sc.$mut$FldType(this.proxy.root(sc.$Kind$FldName));
	}
`,
}
