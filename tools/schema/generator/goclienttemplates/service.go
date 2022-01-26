package goclienttemplates

var serviceGo = map[string]string{
	// *******************************
	"service.go": `
package $package$+client
$#if funcs emitContract
`,
	// *******************************
	"emitContract": `

import "github.com/iotaledger/wasp/wasmvm/wasmlib/go/wasmclient"

const (
$#each params constArg

$#each results constRes
)
$#each func funcStruct

///////////////////////////// $PkgName$+Service /////////////////////////////

type $PkgName$+Service struct {
	wasmclient.Service
}

func New$PkgName$+Service(cl *wasmclient.ServiceClient, chainID string) (*$PkgName$+Service, error) {
	s := &$PkgName$+Service{}
	err := s.Service.Init(cl, chainID, 0x$hscName)
	return s, err
}
$#if events newEventHandler
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
	"newEventHandler": `

func (s *$PkgName$+Service) NewEventHandler() *$PkgName$+Events {
	return &$PkgName$+Events{}
}
`,
	// *******************************
	"funcStruct": `

///////////////////////////// $funcName /////////////////////////////

type $FuncName$Kind struct {
	wasmclient.Client$Kind
$#if param funcArgsMember
}
$#each param funcArgSetter
$#if func funcPost viewCall
`,
	// *******************************
	"funcArgsMember": `
	args wasmclient.Arguments
`,
	// *******************************
	"funcArgSetter": `
$#if array funcArgSetterArray funcArgSetterBasic
`,
	// *******************************
	"funcArgSetterBasic": `

func (f *$FuncName$Kind) $FldName(v $fldLangType) {
	f.args.Set(Arg$FldName, f.args.From$FldType(v))
}
`,
	// *******************************
	"funcArgSetterArray": `

func (f *$FuncName$Kind) $FldName(a []$fldLangType) {
	for i, v := range a {
		f.args.Set(f.args.IndexedKey(Arg$FldName, i), f.args.From$FldType(v))
	}
	f.args.Set(Arg$FldName, f.args.SetInt32(int32(len(a))))
}
`,
	// *******************************
	"funcPost": `

func (f *$FuncName$Kind) Post() wasmclient.Request {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
	return f.ClientFunc.Post(0x$hFuncName, $args)
}
`,
	// *******************************
	"viewCall": `

func (f *$FuncName$Kind) Call() $FuncName$+Results {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
	f.ClientView.Call("$funcName", $args)
	return $FuncName$+Results{res: f.Results()}
}
$#if result resultStruct
`,
	// *******************************
	"mandatoryCheck": `
	f.args.Mandatory(Arg$FldName)
`,
	// *******************************
	"execWithArgs": `
$#set args &f.args
`,
	// *******************************
	"execNoArgs": `
$#set args nil
`,
	// *******************************
	"resultStruct": `

type $FuncName$+Results struct {
	res wasmclient.Results
}
$#each result callResultGetter
`,
	// *******************************
	"callResultGetter": `
$#if map callResultGetterMap callResultGetter2
`,
	// *******************************
	"callResultGetter2": `
$#if basetype callResultGetterBasic callResultGetterStruct
`,
	// *******************************
	"callResultGetterMap": `

func (r *$FuncName$+Results) $FldName() map[$fldKeyLangType]$fldLangType {
    res := make(map[$fldKeyLangType]$fldLangType)
    r.res.ForEach(func(key []byte, val []byte) {
        res[r.res.To$fldMapKey(key)] = r.res.To$FldType(val)
    })
	return res
}
`,
	// *******************************
	"callResultGetterBasic": `
$#if mandatory else callResultOptional

func (r *$FuncName$+Results) $FldName() $fldLangType {
	return r.res.To$FldType(r.res.Get(Res$FldName))
}
`,
	// *******************************
	"callResultGetterStruct": `

func (r *$FuncName$+Results) $FldName() *$FldType {
	return New$FldType$+FromBytes(r.res.Get(Res$FldName))
}
`,
	// *******************************
	"callResultOptional": `

func (r *$FuncName$+Results) $FldName$+Exists() bool {
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
	return $FuncName$Kind{Client$Kind: s.AsClient$Kind()}
}
`,
}
