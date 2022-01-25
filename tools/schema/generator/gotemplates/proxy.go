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
	// TODO when will this be called, and if so, fix it
	"proxyOtherType": `
$#if typedef proxyTypeDef proxyStruct
`,
	// *******************************
	"proxyArray": `

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

func (s $TypeName) $FldName() Map$fldMapKey$+To$mut$FldType {
	//nolint:gosimple
	return Map$fldMapKey$+To$mut$FldType{proxy: s.proxy}
}
`,
	// *******************************
	"proxyMapOther": `

func (s $TypeName) $FldName() Map$fldMapKey$+To$mut$FldType {
	return Map$fldMapKey$+To$mut$FldType{proxy: s.proxy.Root($Kind$FldName)}
}
`,
	// *******************************
	"proxyBaseType": `

func (s $TypeName) $FldName() wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(s.proxy.Root($Kind$FldName))
}
`,
	// *******************************
	// TODO when will this be called, and if so, fix it
	"proxyTypeDef": `

func (s $TypeName) $OldName() $mut$OldType {
	subID := wasmlib.GetObjectID(s.id, $varID, $varType)
	return $mut$OldType{objID: subID}
}
`,
	// *******************************
	// TODO when will this be called, and if so, fix it
	"proxyStruct": `

func (s $TypeName) $FldName() $mut$FldType {
	return $mut$FldType{objID: s.id, keyID: $varID}
}
`,
}
