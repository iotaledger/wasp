package goclienttemplates

var serviceGo = map[string]string{
	// *******************************
	"service.go": `
package $package

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/client"
)

const (
$#each params constArg

$#each results constRes
)
$#each func funcStruct

type $PkgName$+Service struct {
	client.Service
}

func New$PkgName$+Service(client client.ServiceClient, chainId string) *$PkgName$+Service {
	s := &$PkgName$+Service{}
	s.Service.Init(client, chainId, "$hscName", EventHandlers)
	return s
}
$#each func serviceFunction
`,
	// *******************************
	"constArg": `
	Arg$FldName = "$fldAlias"
`,
	// *******************************
	"constRes": `
	Res$FldName = "$fldAlias"
`,
	// *******************************
	"funcStruct": `

type $FuncName$Kind struct {
$#if param funcArgsMember
}
$#each param funcArgSetter
$#if func funcPost viewCall
`,
	// *******************************
	"funcArgsMember": `
	args client.Arguments
`,
	// *******************************
	"funcArgSetter": `

func (f $FuncName$Kind) $FldName(v $fldLangType) {
	f.args.Set$FldType(Arg$FldName, v)
}
`,
	// *******************************
	"funcPost": `

func (f $FuncName$Kind) Post() {
$#each param mandatoryCheck
$#set exec Post
$#if param execWithArgs execNoArgs
	//TODO Do$exec
}
`,
	// *******************************
	"viewCall": `

func (f $FuncName$Kind) Call() $FuncName$+Results {
$#each param mandatoryCheck
$#set exec Call
$#if param execWithArgs execNoArgs
    //TODO Do$exec instead of client.NewResults()
	return $FuncName$+Results { res: client.NewResults() }
}
$#if result resultStruct
`,
	// *******************************
	"mandatoryCheck": `
	f.args.Mandatory(Arg$FldName)
`,
	// *******************************
	"execWithArgs": `
$#set exec $exec(f.args)
`,
	// *******************************
	"execNoArgs": `
$#set exec $exec(nil)
`,
	// *******************************
	"resultStruct": `

type $FuncName$+Results struct {
	res client.Results
}
$#each result callResultGetter
`,
	// *******************************
	"callResultGetter": `
$#if mandatory else callResultOptional

func (r $FuncName$+Results) $FldName() $fldLangType {
	return r.res.Get$FldType(Res$FldName)
}
`,
	// *******************************
	"callResultOptional": `

func (r $FuncName$+Results) $FldName$+Exists() bool {
	return r.res.Exists(Res$FldName)
}
`,
	// *******************************
	"serviceResultExtract": `
	if buf, ok := result["$fldName"]; ok {
		r.$FldName = buf.$resConvert
	}
`,
	// *******************************
	"serviceResult": `
	$FldName $fldLangType
`,
	// *******************************
	"serviceFunction": `

func (s *$PkgName$+Service) $FuncName() $FuncName$Kind {
	return $FuncName$Kind{}
}
`,
}
