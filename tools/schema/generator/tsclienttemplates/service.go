package tsclienttemplates

var serviceTs = map[string]string{
	// *******************************
	"service.ts": `
import * as client from "wasmlib/client"
import * as events from "./events"

$#each params constArg

$#each results constRes
$#each func funcStruct

///////////////////////////// $PkgName$+Service /////////////////////////////

export class $PkgName$+Service extends client.Service {

	constructor(cl: client.ServiceClient, chainID: string) {
		super(cl, chainID, "$hscName", events.eventHandlers);
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

export class $FuncName$Kind extends client.FuncObject {
$#if param funcArgsMember
$#each param funcArgSetter
$#if func funcPost viewCall
}
$#if result resultStruct
`,
	// *******************************
	"funcArgsMember": `
	args: client.Arguments = new client.Arguments();
`,
	// *******************************
	"funcArgSetter": `
	
	$fldName(v: $fldLangType): void {
		this.args.set$FldType(Arg$FldName, v);
	}
`,
	// *******************************
	"funcPost": `
	
	public async post(): Promise<void> {
$#each param mandatoryCheck
$#set exec this.svc.postRequest
$#if param execWithArgs execNoArgs
		$exec;
	}
`,
	// *******************************
	"viewCall": `

	public async call(): Promise<$FuncName$+Results> {
$#each param mandatoryCheck
$#set exec this.svc.callView
$#if param execWithArgs execNoArgs
		return new $FuncName$+Results($exec);
	}
`,
	// *******************************
	"mandatoryCheck": `
		this.args.mandatory(Arg$FldName);
`,
	// *******************************
	"execWithArgs": `
$#set exec $exec("$funcName", this.args)
`,
	// *******************************
	"execNoArgs": `
$#set exec $exec("$funcName", null)
`,
	// *******************************
	"resultStruct": `

export class $FuncName$+Results extends client.ViewResults {
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
	$fldName: client.$FldType;
`,
	// *******************************
	"serviceFunction": `

	public $funcName(): $FuncName$Kind {
    	return new $FuncName$Kind(this);
	}
`,
}
