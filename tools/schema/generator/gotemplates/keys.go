package gotemplates

var keysGo = map[string]string{
	// *******************************
	"keys.go": `
$#emit goHeader

const (
$#each params constFieldIdx
$#each results constFieldIdx
$#each state constFieldIdx
)

const keyMapLen = $maxIndex

var keyMap = [keyMapLen]wasmlib.Key{
$#each params constFieldKey
$#each results constFieldKey
$#each state constFieldKey
}

var idxMap [keyMapLen]wasmlib.Key32
`,
	// *******************************
	"constFieldIdx": `
	Idx$constPrefix$FldName = $fldIndex
`,
	// *******************************
	"constFieldKey": `
	$constPrefix$FldName,
`,
}
