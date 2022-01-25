package tstemplates

var typedefsTs = map[string]string{
	// *******************************
	"typedefs.ts": `
$#emit tsImports
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

export class $mut$FldName extends $proxy {
};
`,
	// *******************************
	"typedefProxyArray": `
$#set proxy ArrayOf$mut$FldType
$#if exist else typedefProxyArrayNew
`,
	// *******************************
	"typedefProxyArrayNew": `

export class $proxy {
	objID: i32;

    constructor(objID: i32) {
        this.objID = objID;
    }
$#if mut typedefProxyArrayClear

    length(): u32 {
        return wasmlib.getLength(this.objID);
    }
$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayClear": `

    clear(): void {
        wasmlib.clear(this.objID);
    }
`,
	// *******************************
	"typedefProxyArrayNewBaseType": `

    get$FldType(index: u32): wasmlib.Sc$mut$FldType {
        return new wasmlib.Sc$mut$FldType(this.objID, new wasmlib.Key32(index as i32));
    }
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `
$#if typedef typedefProxyArrayNewOtherTypeTypeDef typedefProxyArrayNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyArrayNewOtherTypeTypeDef": `
$#emit setVarType

	get$OldType(index: u32): sc.$mut$OldType {
		let subID = wasmlib.getObjectID(this.objID, new wasmlib.Key32(index as i32), $varType);
		return new sc.$mut$OldType(subID);
	}
`,
	// *******************************
	"typedefProxyArrayNewOtherTypeStruct": `

	get$FldType(index: u32): sc.$mut$FldType {
		return new sc.$mut$FldType(this.objID, new wasmlib.Key32(index as i32));
	}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$fldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

export class $proxy {
	objID: i32;

    constructor(objID: i32) {
        this.objID = objID;
    }
$#if mut typedefProxyMapClear
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapClear": `

    clear(): void {
        wasmlib.clear(this.objID);
    }
`,
	// *******************************
	"typedefProxyMapNewBaseType": `

    get$FldType(key: $fldKeyLangType): wasmlib.Sc$mut$FldType {
        return new wasmlib.Sc$mut$FldType(this.objID, $fldKeyToKey32);
    }
`,
	// *******************************
	"typedefProxyMapNewOtherType": `
$#if typedef typedefProxyMapNewOtherTypeTypeDef typedefProxyMapNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyMapNewOtherTypeTypeDef": `
$#emit setVarType

    get$OldType(key: $oldKeyLangType): sc.$mut$OldType {
        let subID = wasmlib.getObjectID(this.objID, $oldKeyToKey32, $varType);
        return new sc.$mut$OldType(subID);
    }
`,
	// *******************************
	"typedefProxyMapNewOtherTypeStruct": `

    get$FldType(key: $fldKeyLangType): sc.$mut$FldType {
        return new sc.$mut$FldType(this.objID, $fldKeyToKey32);
    }
`,
}
