package gotemplates

var proxyGo = map[string]string{
	// *******************************
	"proxyContainers": `
$#if array typedefProxyArray
$#if map typedefProxyMap
`,
	// *******************************
	"proxyMethods": `
$#set varID wasmlib.KeyID($Kind$FldName)
$#if init setInitVarID
$#if core setCoreVarID
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"setInitVarID": `
$#set varID idxMap[Idx$Kind$FldName]
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
$#set varID wasmlib.KeyID($Kind$FldName)
`,
	// *******************************
	"proxyArray": `

func (s $TypeName) $FldName() ArrayOf$mut$FldType {
	arrID := wasmlib.GetObjectID(s.id, $varID, $arrayTypeID|$fldTypeID)
	return ArrayOf$mut$FldType{objID: arrID}
}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `

func (s $TypeName) $FldName() Map$fldMapKey$+To$mut$FldType {
	return Map$fldMapKey$+To$mut$FldType{objID: s.id}
}
`,
	// *******************************
	"proxyMapOther": `

func (s $TypeName) $FldName() Map$fldMapKey$+To$mut$FldType {
	mapID := wasmlib.GetObjectID(s.id, $varID, wasmlib.TYPE_MAP)
	return Map$fldMapKey$+To$mut$FldType{objID: mapID}
}
`,
	// *******************************
	"proxyBaseType": `

func (s $TypeName) $FldName() wasmlib.Sc$mut$FldType {
	return wasmlib.NewSc$mut$FldType(s.id, $varID)
}
`,
	// *******************************
	"proxyTypeDef": `
$#emit setVarType

func (s $TypeName) $OldName() $mut$OldType {
	subID := wasmlib.GetObjectID(s.id, $varID, $varType)
	return $mut$OldType{objID: subID}
}
`,
	// *******************************
	"proxyStruct": `

func (s $TypeName) $FldName() $mut$FldType {
	return $mut$FldType{objID: s.id, keyID: $varID}
}
`,
}
