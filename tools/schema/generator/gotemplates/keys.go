package gotemplates

var keysGo = map[string]string{
	// *******************************
	"keys.go": `
$#emit goHeader

const (
$#set constPrefix Param
$#each params constFieldIdx

$#set constPrefix Result
$#each results constFieldIdx

$#set constPrefix State
$#each state constFieldIdx
)

const keyMapLen = $maxIndex

var keyMap = [keyMapLen]wasmlib.Key{
$#set constPrefix Param
$#each params constFieldKey
$#set constPrefix Result
$#each results constFieldKey
$#set constPrefix State
$#each state constFieldKey
}

var idxMap [keyMapLen]wasmlib.Key32
`,
	// *******************************
	"constFieldIdx": `
	Idx$constPrefix$FldName$fldPad = $fldIndex
`,
	// *******************************
	"constFieldKey": `
	$constPrefix$FldName,
`,
}
