package gotemplates

var structsGo = map[string]string{
	// *******************************
	"structs.go": `
$#emit goPackage

$#emit importWasmCodec
$#emit importWasmTypes
$#each structs structType
`,
	// *******************************
	"structType": `

type $StrName struct {
$#each struct structField
}

func New$StrName$+FromBytes(buf []byte) *$StrName {
	dec := wasmcodec.NewWasmDecoder(buf)
	data := &$StrName$+{}
$#each struct structDecode
	dec.Close()
	return data
}

func (o *$StrName) Bytes() []byte {
	enc := wasmcodec.NewWasmEncoder()
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
	data.$FldName$fldPad = wasmtypes.Decode$FldType(dec)
`,
	// *******************************
	"structEncode": `
		wasmtypes.Encode$FldType(enc, o.$FldName)
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
