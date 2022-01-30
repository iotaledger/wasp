package tstemplates

var keysTs = map[string]string{
	// *******************************
	"keys.ts": `

$#each params constFieldIdx
`,
	// *******************************
	"constFieldIdx": `
export const IdxParam$FldName$fldPad = $fldIndex;
`,
}
