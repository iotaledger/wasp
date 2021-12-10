package clienttemplates

var serviceTs = map[string]string{
	// *******************************
	"service.ts": `
import * as wasmlib from "./wasmlib"
import * as events from "./events"
$#each func serviceResults

export class $PkgName$+Service extends wasmlib.Service {

  constructor(client: BasicClient, chainId: string) {
    super(client, chainId, 0x$hscName);
  }
$#each func serviceFunction
}
`,
	// *******************************
	"serviceResults": `
$#if result serviceHasResults
`,
	// *******************************
	"serviceHasResults": `

export interface $FuncName$+Result {
$#each result serviceResult
}
`,
	// *******************************
	"serviceResult": `
	$fldName: wasmlib.$FldType;
`,
	// *******************************
	"serviceFunction": `
$#set sep $empty
$#set params $empty
$#each param serviceParam
$#emit service$Kind
`,
	// *******************************
	"serviceParam": `
$#set params $params$sep$fldName: wasmlib.$FldType
$#set sep , 
`,
	// *******************************
	"serviceFunc": `

	public async $funcName($params): Promise<void> {
		const args: wasmlib.Argument[] = [
$#each param serviceFuncParam
		];
    	await this.postRequest(0x$funcHname, args);
	}
`,
	// *******************************
	"serviceFuncParam": `
				{ key: '$fldName', value: $fldName, },
`,
	// *******************************
	"serviceView": `

	public async $funcName($params): Promise<$FuncName$+Result> {
		const args: wasmlib.Argument[] = [
$#each param serviceFuncParam
		];
		const response = await this.callView(0x$funcHname, args);
        let result: $FuncName$+Result = {};

$#each result serviceViewResult

		return result;
	}
`,
	// *******************************
	"serviceViewResult": `
		let $fldName = response['$fldName'];
		result.$fldName = $fldDefault;
		if ($fldName) {
			result.$fldName = $fldName.$resConvert($fldName)$resConvert2;
		}
`,
}
