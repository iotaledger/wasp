package tstemplates

var keysTs = map[string]string{
	// *******************************
	"keys.ts": `
$#emit tsImports

$#set constPrefix Param
$#each params constFieldIdx

$#set constPrefix Result
$#each results constFieldIdx

$#set constPrefix State
$#each state constFieldIdx

export let keyMap: string[] = [
$#set constPrefix Param
$#each params constFieldKey
$#set constPrefix Result
$#each results constFieldKey
$#set constPrefix State
$#each state constFieldKey
];

export let idxMap: wasmlib.Key32[] = new Array(keyMap.length);
`,
	// *******************************
	"constFieldIdx": `
export const Idx$constPrefix$FldName$fldPad = $fldIndex;
`,
	// *******************************
	"constFieldKey": `
	sc.$constPrefix$FldName,
`,
}
