package goclienttemplates

var funcsGo = map[string]string{
	// *******************************
	"funcs.go": `
package $package

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/client"
)
$#each events funcSignature
`,
	// *******************************
	"funcSignature": `

func On$PkgName$EvtName(event *Event$EvtName) {
}
`,
}
