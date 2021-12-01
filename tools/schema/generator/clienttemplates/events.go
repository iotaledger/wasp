package clienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
import * as wasmlib from "./wasmlib"
import * as service from "./service"

$#each events eventInterface

export interface $Package$+Events {
$#each events eventSignature
}

export function handleVmMessage(message: string[]): void {
    const messageHandlers: wasmlib.MessageHandlers = {
$#each events eventHandler
    };

    const topicIndex = 3;
    const topic = message[topicIndex];

    if (typeof messageHandlers[topic] != 'undefined') {
      messageHandlers[topic](topicIndex);
    }
}
`,
	// *******************************
	"eventSignature": `
	$package$+_$evtName: (event: Event$EvtName) => void;
`,
	// *******************************
	"eventInterface": `

export interface Event$EvtName {
  timestamp: wasmlib.Int32;
$#each event eventInterfaceField
}
`,
	// *******************************
	"eventInterfaceField": `
  $fldName: wasmlib.$FldType;
`,
	// *******************************
	"eventHandler": `
		'$package.$evtName': (index) => {
			const evt: Event$EvtName = {
				timestamp: Number(message[++index]),
$#each event eventHandlerField
			};
			this.emitter.emit('$package$+_$evtName', evt);
		},
`,
	// *******************************
	"eventHandlerField": `
				$fldName: $msgConvert,
`,
}
