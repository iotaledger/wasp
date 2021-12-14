package tsclienttemplates

var serviceTs = map[string]string{
	// *******************************
	"service.ts": `
import * as client from "wasmlib/client"
import * as events from "./events"

$#each params constArg

$#each results constRes
$#each func funcStruct

export class $PkgName$+Service extends client.Service {

	constructor(client: client.ServiceClient, chainId: string) {
		super(client, chainId, "$hscName", events.eventHandlers);
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

export class $FuncName$Kind {
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
	
	post(): void {
$#each param mandatoryCheck
$#set exec Post
$#if param execWithArgs execNoArgs
	//TODO Do$exec
	}
`,
	// *******************************
	"viewCall": `

	call(): $FuncName$+Results {
$#each param mandatoryCheck
$#set exec Call
$#if param execWithArgs execNoArgs
    	//TODO Do$exec instead of new client.Results()
		return new $FuncName$+Results(new client.Results());
	}
`,
	// *******************************
	"mandatoryCheck": `
		this.args.mandatory(Arg$FldName);
`,
	// *******************************
	"execWithArgs": `
$#set exec $exec(this.args)
`,
	// *******************************
	"execNoArgs": `
$#set exec $exec(null)
`,
	// *******************************
	"resultStruct": `

export class $FuncName$+Results {
	res: client.Results;

	constructor(res: client.Results) { this.res = res; }
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
    	return new $FuncName$Kind();
	}
`,
}
