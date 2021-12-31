package tsclienttemplates

var serviceTs = map[string]string{
	// *******************************
	"service.ts": `
import * as wasmclient from "wasmclient"
import * as events from "./events"

$#each params constArg

$#each results constRes
$#each func funcStruct

///////////////////////////// $PkgName$+Service /////////////////////////////

export class $PkgName$+Service extends wasmclient.Service {

	constructor(cl: wasmclient.ServiceClient, chainID: string) {
		super(cl, chainID, 0x$hscName, events.eventHandlers);
	}
$#each func serviceFunction
}
`,
	// *******************************
	"constArg": `
const Arg$FldName = "$fldAlias";
`,
	// *******************************
	"constRes": `
const Res$FldName = "$fldAlias";
`,
	// *******************************
	"funcStruct": `

///////////////////////////// $funcName /////////////////////////////

export class $FuncName$Kind extends wasmclient.Client$Kind {
$#if param funcArgsMember
$#each param funcArgSetter
$#if func funcPost viewCall
}
$#if result resultStruct
`,
	// *******************************
	"funcArgsMember": `
	args: wasmclient.Arguments = new wasmclient.Arguments();
`,
	// *******************************
	"funcArgSetter": `
$#if array funcArgSetterArray funcArgSetterBasic
`,
	// *******************************
	"funcArgSetterBasic": `
	
	$fldName(v: $fldLangType): void {
		this.args.set$FldType(Arg$FldName, v);
	}
`,
	// *******************************
	"funcArgSetterArray": `
	
	$fldName(a: $fldLangType[]): void {
		for (let i = 0; i < a.length; i++) {
			this.args.set$FldType(this.args.indexedKey(Arg$FldName, i), a[i]);
		}
		this.args.setInt32(Arg$FldName, a.length);
	}
`,
	// *******************************
	"funcPost": `
	
	public post(): wasmclient.RequestID {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
		return super.post(0x$hFuncName, $args);
	}
`,
	// *******************************
	"viewCall": `

	public call(): $FuncName$+Results {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
		super.call("$funcName", $args);
		return new $FuncName$+Results(this.results());
	}
`,
	// *******************************
	"mandatoryCheck": `
		this.args.mandatory(Arg$FldName);
`,
	// *******************************
	"execWithArgs": `
$#set args this.args
`,
	// *******************************
	"execNoArgs": `
$#set args null
`,
	// *******************************
	"resultStruct": `

export class $FuncName$+Results extends wasmclient.ViewResults {
$#each result callResultGetter
}
`,
	// *******************************
	"callResultGetter": `
$#if mandatory else callResultOptional

	$fldName(): $fldLangType {
		return this.res.get$FldType(Res$FldName);
	}
`,
	// *******************************
	"callResultOptional": `
	
	$fldName$+Exists(): boolean {
		return this.res.exists(Res$FldName)
	}
`,
	// *******************************
	"serviceResultExtract": `
		let $fldName = result["$fldName"];
		if ($fldName) {
			this.$fldName = $fldName.$resConvert;
		}
`,
	// *******************************
	"serviceResult": `
	$fldName: wasmclient.$FldType;
`,
	// *******************************
	"serviceFunction": `

	public $funcName(): $FuncName$Kind {
    	return new $FuncName$Kind(this);
	}
`,
}
