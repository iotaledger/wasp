package gotemplates

var proxyGo = map[string]string{
	// *******************************
	"proxyContainers": `
$#if array typedefProxyArray
$#if map typedefProxyMap
`,
	// *******************************
	"proxyMethods": `
$#set varID idxMap[Idx$Kind$FldName]
$#if core setCoreVarID
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"proxyMethods2": `
$#if map proxyMap proxyMethods3
`,
	// *******************************
	"proxyMethods3": `
$#if basetype proxyBaseType proxyNewType
`,
	// *******************************
	"setCoreVarID": `
$#set varID $Kind$FldName.KeyID()
`,
	// *******************************
	"proxyArray": `

func (s $TypeName) $FldName() ArrayOf$mut$FldType {
	arrID := wasmlib.GetObjectID(s.id, $varID, $ArrayTypeID|$FldTypeID)
	return ArrayOf$mut$FldType{objID: arrID}
}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `

func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	return Map$FldMapKey$+To$mut$FldType{objID: s.id}
}
`,
	// *******************************
	"proxyMapOther": `55544444.0

func (s $TypeName) $FldName() Map$FldMapKey$+To$mut$FldType {
	mapID := wasmlib.GetObjectID(s.id, $varID, wasmlib.TYPE_MAP)
	return Map$FldMapKey$+To$mut$FldType{objID: mapID}
}
`,
	// *******************************
	"proxyBaseType": `

func (s $TypeName) $FldName() wasmlib.Sc$mut$FldType {
	return wasmlib.NewSc$mut$FldType(s.id, $varID)
}
`,
	// *******************************
	"proxyNewType": `

func (s $TypeName) $FldName() $mut$FldType {
	return $mut$FldType{objID: s.id, keyID: $varID}
}
`,
}
