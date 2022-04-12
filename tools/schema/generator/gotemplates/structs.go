// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var structsGo = map[string]string{
	// *******************************
	"structs.go": `
$#emit goPackage

$#emit importWasmTypes
$#each structs structType
`,
	// *******************************
	"structType": `

$#if structComment _structComment
type $StrName struct {
$#each struct structField
}

func New$StrName$+FromBytes(buf []byte) *$StrName {
	dec := wasmtypes.NewWasmDecoder(buf)
	data := &$StrName$+{}
$#each struct structDecode
	dec.Close()
	return data
}

func (o *$StrName) Bytes() []byte {
	enc := wasmtypes.NewWasmEncoder()
$#each struct structEncode
	return enc.Buf()
}
$#set mut Immutable
$#emit structMethods
$#set mut Mutable
$#emit structMethods
`,
	// *******************************
	"structField": `
	$FldName$fldPad $fldLangType $fldComment
`,
	// *******************************
	"structDecode": `
	data.$FldName$fldPad = wasmtypes.$FldType$+Decode(dec)
`,
	// *******************************
	"structEncode": `
		wasmtypes.$FldType$+Encode(enc, o.$FldName)
`,
	// *******************************
	"structMethods": `

type $mut$StrName struct {
	proxy wasmtypes.Proxy
}
$#if mut structMethodDelete

func (o $mut$StrName) Exists() bool {
	return o.proxy.Exists()
}
$#if mut structMethodSetValue

func (o $mut$StrName) Value() *$StrName {
	return New$StrName$+FromBytes(o.proxy.Get())
}
`,
	// *******************************
	"structMethodDelete": `

func (o $mut$StrName) Delete() {
	o.proxy.Delete()
}
`,
	// *******************************
	"structMethodSetValue": `

func (o $mut$StrName) SetValue(value *$StrName) {
	o.proxy.Set(value.Bytes())
}
`,
}
