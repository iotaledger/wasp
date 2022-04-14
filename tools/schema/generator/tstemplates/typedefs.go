// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var typedefsTs = map[string]string{
	// *******************************
	"typedefs.ts": `
$#emit importWasmTypes
$#emit importSc
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
export class $mut$FldName extends $proxy {
}
`,
	// *******************************
	"typedefProxyArray": `
$#set proxy ArrayOf$mut$FldType
$#if exist else typedefProxyArrayNew
`,
	// *******************************
	"typedefProxyArrayNew": `

export class $proxy extends wasmtypes.ScProxy {
$#if mut typedefProxyArrayMut

	length(): u32 {
		return this.proxy.length();
	}
$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayMut": `
$#if basetype typedefProxyArrayAppendBaseType typedefProxyArrayAppendOtherType

	clear(): void {
		this.proxy.clearArray();
	}
`,
	// *******************************
	"typedefProxyArrayAppendBaseType": `

	append$FldType(): wasmtypes.Sc$mut$FldType {
		return new wasmtypes.Sc$mut$FldType(this.proxy.append());
	}
`,
	// *******************************
	"typedefProxyArrayAppendOtherType": `

	append$FldType(): sc.$mut$FldType {
		return new sc.$mut$FldType(this.proxy.append());
	}
`,
	// *******************************
	"typedefProxyArrayNewBaseType": `

	get$FldType(index: u32): wasmtypes.Sc$mut$FldType {
		return new wasmtypes.Sc$mut$FldType(this.proxy.index(index));
	}
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `

	get$FldType(index: u32): sc.$mut$FldType {
		return new sc.$mut$FldType(this.proxy.index(index));
	}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$FldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

export class $proxy extends wasmtypes.ScProxy {
$#if mut typedefProxyMapMut
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapMut": `

	clear(): void {
		this.proxy.clearMap();
	}
`,
	// *******************************
	"typedefProxyMapNewBaseType": `

	get$FldType(key: $fldKeyLangType): wasmtypes.Sc$mut$FldType {
		return new wasmtypes.Sc$mut$FldType(this.proxy.key(wasmtypes.$fldMapKey$+ToBytes(key)));
	}
`,
	// *******************************
	"typedefProxyMapNewOtherType": `

	get$FldType(key: $fldKeyLangType): sc.$mut$FldType {
		return new sc.$mut$FldType(this.proxy.key(wasmtypes.$fldMapKey$+ToBytes(key)));
	}
`,
}
