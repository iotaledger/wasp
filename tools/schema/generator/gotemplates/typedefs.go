package gotemplates

var typedefsGo = map[string]string{
	// *******************************
	"typedefs.go": `
$#emit goPackage

$#emit importWasmTypes
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

type $mut$FldName = $proxy
`,
	// *******************************
	"typedefProxyArray": `
$#set proxy ArrayOf$mut$FldType
$#if exist else typedefProxyArrayNew
`,
	// *******************************
	"typedefProxyArrayNew": `

type $proxy struct {
	proxy wasmtypes.Proxy
}
$#if mut typedefProxyArrayMut

func (a $proxy) Length() uint32 {
	return a.proxy.Length()
}
$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayMut": `
$#if basetype typedefProxyArrayAppendBaseType typedefProxyArrayAppendOtherType

func (a $proxy) Clear() {
	a.proxy.ClearArray()
}
`,
	// *******************************
	"typedefProxyArrayAppendBaseType": `

func (a $proxy) Append$FldType() wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(a.proxy.Append())
}
`,
	// *******************************
	"typedefProxyArrayAppendOtherType": `

func (a $proxy) Append$FldType() $mut$FldType {
	return $mut$FldType{proxy: a.proxy.Append()}
}
`,
	// *******************************
	"typedefProxyArrayNewBaseType": `

func (a $proxy) Get$FldType(index uint32) wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(a.proxy.Index(index))
}
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `

func (a $proxy) Get$FldType(index uint32) $mut$FldType {
	return $mut$FldType{proxy: a.proxy.Index(index)}
}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$fldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

type $proxy struct {
	proxy wasmtypes.Proxy
}
$#if mut typedefProxyMapMut
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapMut": `

func (m $proxy) Clear() {
	m.proxy.ClearMap()
}
`,
	// *******************************
	"typedefProxyMapNewBaseType": `

func (m $proxy) Get$FldType(key $fldKeyLangType) wasmtypes.Sc$mut$FldType {
	return wasmtypes.NewSc$mut$FldType(m.proxy.Key($fldKeyToBytes))
}
`,
	// *******************************
	"typedefProxyMapNewOtherType": `

func (m $proxy) Get$FldType(key $fldKeyLangType) $mut$FldType {
	return $mut$FldType{proxy: m.proxy.Key($fldKeyToBytes)}
}
`,
}
