package goclienttemplates

var funcsGo = map[string]string{
	// *******************************
	"funcs.go": `
$#emit clientHeader
$#each events funcSignature
`,
	// *******************************
	"funcSignature": `

func On$PkgName$EvtName(event *Event$EvtName) {
}
`,
}
