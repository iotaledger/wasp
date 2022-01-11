package tsclienttemplates

var serviceTs = map[string]string{
	// *******************************
	"service.ts": `
$#emit importWasmLib
$#if core else importEvents

$#each params constArg

$#each results constRes
$#each func funcStruct

///////////////////////////// $PkgName$+Service /////////////////////////////

export class $PkgName$+Service extends wasmclient.Service {

	public constructor(cl: wasmclient.ServiceClient) {
$#set eventHandlers events.eventHandlers
$#if core noEventHandlers
		super(cl, 0x$hscName, $eventHandlers);
	}
$#each func serviceFunction
}
`,
	// *******************************
	"noEventHandlers": `
$#set eventHandlers new Map()
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
	private args: wasmclient.Arguments = new wasmclient.Arguments();
`,
	// *******************************
	"funcArgSetter": `
$#if array funcArgSetterArray funcArgSetterBasic
`,
	// *******************************
	"funcArgSetterBasic": `
	
	public $fldName(v: $fldLangType): void {
		this.args.set(Arg$FldName, this.args.from$FldType(v));
	}
`,
	// *******************************
	"funcArgSetterArray": `
	
	public $fldName(a: $fldLangType[]): void {
		for (let i = 0; i < a.length; i++) {
			this.args.set(this.args.indexedKey(Arg$FldName, i), this.args.from$FldType(a[i]));
		}
		this.args.set(Arg$FldName, this.args.setInt32(a.length));
	}
`,
	// *******************************
	"funcPost": `
	
	public async post(): Promise<wasmclient.RequestID> {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
		return await super.post(0x$hFuncName, $args);
	}
`,
	// *******************************
	"viewCall": `

	public async call(): Promise<$FuncName$+Results> {
$#each mandatory mandatoryCheck
$#if param execWithArgs execNoArgs
        const res = new $FuncName$+Results();
		await this.callView("$funcName", $args, res);
		return res;
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

export class $FuncName$+Results extends wasmclient.Results {
$#each result callResultGetter
}
`,
	// *******************************
	"callResultGetter": `
$#if map callResultGetterMap callResultGetterBasic
`,
	// *******************************
	"callResultGetterMap": `

	$fldName(): Map<$fldKeyLangType, $fldLangType> {
		const res = new Map<$fldKeyLangType, $fldLangType>();
		this.forEach((key, val) => {
			res.set(this.to$fldMapKey(key), this.to$FldType(val));
		});
		return res;
	}
`,
	// *******************************
	"callResultGetterBasic": `
$#if mandatory else callResultOptional

	$fldName(): $fldLangType {
		return this.to$FldType(this.get(Res$FldName));
	}
`,
	// *******************************
	"callResultOptional": `
	
	$fldName$+Exists(): boolean {
		return this.exists(Res$FldName)
	}
`,
	// *******************************
	"serviceFunction": `

	public $funcName(): $FuncName$Kind {
		return new $FuncName$Kind(this);
	}
`,
}
