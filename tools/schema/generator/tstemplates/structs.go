package tstemplates

var structsTs = map[string]string{
	// *******************************
	"structs.ts": `
$#emit importWasmLib
$#each structs structType
`,
	// *******************************
	"structType": `

export class $StrName {
$#each struct structField

    static fromBytes(bytes: u8[]): $StrName {
        let decode = new wasmlib.BytesDecoder(bytes);
        let data = new $StrName();
$#each struct structDecode
        decode.close();
        return data;
    }

    bytes(): u8[] {
        return new wasmlib.BytesEncoder().
$#each struct structEncode
            data();
    }
}
$#set mut Immutable
$#emit structMethods
$#set mut Mutable
$#emit structMethods
`,
	// *******************************
	"structField": `
    $fldName$fldPad : $fldLangType = $fldTypeInit; $fldComment
`,
	// *******************************
	"structDecode": `
        data.$fldName$fldPad = decode.$fldType();
`,
	// *******************************
	"structEncode": `
		    $fldType(this.$fldName).
`,
	// *******************************
	"structMethods": `

export class $mut$StrName {
    objID: i32;
    keyID: wasmlib.Key32;

    constructor(objID: i32, keyID: wasmlib.Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }
$#if mut structMethodDelete

    exists(): boolean {
        return wasmlib.exists(this.objID, this.keyID, wasmlib.TYPE_BYTES);
    }
$#if mut structMethodSetValue

    value(): $StrName {
        return $StrName.fromBytes(wasmlib.getBytes(this.objID, this.keyID, wasmlib.TYPE_BYTES));
    }
}
`,
	// *******************************
	"structMethodDelete": `

    delete(): void {
        wasmlib.delKey(this.objID, this.keyID, wasmlib.TYPE_BYTES);
    }
`,
	// *******************************
	"structMethodSetValue": `

    setValue(value: $StrName): void {
        wasmlib.setBytes(this.objID, this.keyID, wasmlib.TYPE_BYTES, value.bytes());
    }
`,
}
