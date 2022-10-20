// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var structsTs = map[string]string{
	// *******************************
	"structs.ts": `
$#emit importWasmTypes
$#each structs structType
`,
	// *******************************
	"structType": `

$#each structComment _structComment
export class $StrName {
$#each struct structField

    static fromBytes(buf: u8[]): $StrName {
        const dec = new wasmtypes.WasmDecoder(buf);
        const data = new $StrName();
$#each struct structDecode
        dec.close();
        return data;
    }

    bytes(): u8[] {
        const enc = new wasmtypes.WasmEncoder();
$#each struct structEncode
        return enc.buf();
    }
}
$#set mut Immutable
$#emit structMethods
$#set mut Mutable
$#emit structMethods
`,
	// *******************************
	"structField": `
$#each fldComment _structFieldComment
    $fldName$fldPad : $fldLangType = $fldTypeInit;
`,
	// *******************************
	"structDecode": `
        data.$fldName$fldPad = wasmtypes.$fldType$+Decode(dec);
`,
	// *******************************
	"structEncode": `
        wasmtypes.$fldType$+Encode(enc, this.$fldName);
`,
	// *******************************
	"structMethods": `

export class $mut$StrName extends wasmtypes.ScProxy {
$#if mut structMethodDelete

    exists(): bool {
        return this.proxy.exists();
    }
$#if mut structMethodSetValue

    value(): $StrName {
        return $StrName.fromBytes(this.proxy.get());
    }
}
`,
	// *******************************
	"structMethodDelete": `

    delete(): void {
        this.proxy.delete();
    }
`,
	// *******************************
	"structMethodSetValue": `

    setValue(value: $StrName): void {
        this.proxy.set(value.bytes());
    }
`,
}
