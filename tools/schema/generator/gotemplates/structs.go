package gotemplates

var structsGo = map[string]string{
	// *******************************
	"structs.go": `
$#emit goHeader
$#each structs structType
`,
	// *******************************
	"structType": `

type $StrName struct {
$#each struct structField
}

func New$StrName$+FromBytes(bytes []byte) *$StrName {
	decode := wasmlib.NewBytesDecoder(bytes)
	data := &$StrName$+{}
$#each struct structDecode
	decode.Close()
	return data
}

func (o *$StrName) Bytes() []byte {
	return wasmlib.NewBytesEncoder().
$#each struct structEncode
		Data()
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
	data.$FldName$fldPad = decode.$FldType()
`,
	// *******************************
	"structEncode": `
		$FldType(o.$FldName).
`,
	// *******************************
	"structMethods": `

type $mut$StrName struct {
	objID int32
	keyID wasmlib.Key32
}
$#if mut structMethodDelete

func (o $mut$StrName) Exists() bool {
	return wasmlib.Exists(o.objID, o.keyID, wasmlib.TYPE_BYTES)
}
$#if mut structMethodSetValue

func (o $mut$StrName) Value() *$StrName {
	return New$StrName$+FromBytes(wasmlib.GetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES))
}
`,
	// *******************************
	"structMethodDelete": `

func (o $mut$StrName) Delete() {
	wasmlib.DelKey(o.objID, o.keyID, wasmlib.TYPE_BYTES)
}
`,
	// *******************************
	"structMethodSetValue": `

func (o $mut$StrName) SetValue(value *$StrName) {
	wasmlib.SetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES, value.Bytes())
}
`,
}
