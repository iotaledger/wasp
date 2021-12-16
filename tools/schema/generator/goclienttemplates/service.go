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

///////////////////////////// $PkgName$+Service /////////////////////////////

type $PkgName$+Service struct {
	client.Service
}

func New$PkgName$+Service(cl client.ServiceClient, chainID string) *$PkgName$+Service {
	s := &$PkgName$+Service{}
	s.Service.Init(cl, chainID, "$hscName", EventHandlers)
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

///////////////////////////// $funcName /////////////////////////////

type $FuncName$Kind struct {
	svc *client.Service
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
$#set exec f.svc.PostRequest
$#if param execWithArgs execNoArgs
	$exec
}
`,
	// *******************************
	"viewCall": `

func (f $FuncName$Kind) Call() $FuncName$+Results {
$#each param mandatoryCheck
$#set exec f.svc.CallView
$#if param execWithArgs execNoArgs
	return $FuncName$+Results { res: $exec }
}
$#if result resultStruct
`,
	// *******************************
	"mandatoryCheck": `
	f.args.Mandatory(Arg$FldName)
`,
	// *******************************
	"execWithArgs": `
$#set exec $exec("$funcName", &f.args)
`,
	// *******************************
	"execNoArgs": `
$#set exec $exec("$funcName", nil)
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
	return $FuncName$Kind{ svc: &s.Service }
}
`,
}
