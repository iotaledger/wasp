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
$#set varID wasmlib.Key32.fromString(sc.$Kind$FldName)
$#if init setInitVarID
$#if core setCoreVarID
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"setInitVarID": `
$#set varID sc.idxMap[sc.Idx$Kind$FldName]
`,
	// *******************************
	"proxyMethods2": `
$#if map proxyMap proxyMethods3
`,
	// *******************************
	"proxyMethods3": `
$#if basetype proxyBaseType proxyMethods4
`,
	// *******************************
	"proxyMethods4": `
$#if typedef proxyTypeDef proxyStruct
`,
	// *******************************
	"setCoreVarID": `
$#set varID wasmlib.Key32.fromString(sc.$Kind$FldName)
`,
	// *******************************
	"proxyArray": `
    $fldName(): sc.ArrayOf$mut$FldType {
		let arrID = wasmlib.getObjectID(this.mapID, $varID, $arrayTypeID|$fldTypeID);
		return new sc.ArrayOf$mut$FldType(arrID);
	}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `
    $fldName(): sc.Map$fldMapKey$+To$mut$FldType {
		return new sc.Map$fldMapKey$+To$mut$FldType(this.mapID);
	}
`,
	// *******************************
	"proxyMapOther": `
    $fldName(): sc.Map$fldMapKey$+To$mut$FldType {
		let mapID = wasmlib.getObjectID(this.mapID, $varID, wasmlib.TYPE_MAP);
		return new sc.Map$fldMapKey$+To$mut$FldType(mapID);
	}
`,
	// *******************************
	"proxyBaseType": `
    $fldName(): wasmlib.Sc$mut$FldType {
		return new wasmlib.Sc$mut$FldType(this.mapID, $varID);
	}
`,
	// *******************************
	"proxyTypeDef": `
$#emit setVarType
    $oldName(): sc.$mut$OldType {
		let subID = wasmlib.getObjectID(this.mapID, $varID, $varType);
		return new sc.$mut$OldType(subID);
	}
`,
	// *******************************
	"proxyStruct": `
    $fldName(): sc.$mut$FldType {
		return new sc.$mut$FldType(this.mapID, $varID);
	}
`,
}
