package clienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `

type Address = string;
type AgentID = string;
type Bool = boolean;
type Bytes = Uint8Array;
type ChainID = string;
type Color = string;
type Hash = string;
type Hname = string;
type Int8 = number;
type Int16 = number;
type Int32 = number;
type Int64 = bigint;
type RequestID = string;
type String = string;
type Uint8 = number;
type Uint16 = number;
type Uint32 = number;
type Uint64 = bigint;

export class $Package$+Events {
$#each events eventConst
}
$#each events eventInterface

  private handleVmMessage(message: string[]): void {
    const messageHandlers: MessageHandlers = {
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
	"eventConst": `
  public static readonly EVENT_$EVT_NAME: string = '$package.$evtName';
`,
	// *******************************
	"eventInterface": `

export interface Event$EvtName {
  timestamp: Int32;
$#each event eventInterfaceField
}
`,
	// *******************************
	"eventInterfaceField": `
  $fldName: $FldType;
`,
	// *******************************
	"eventHandler": `
      EVENT_$EVT_NAME: (index) => {
        const evt: Event$EvtName = {
          timestamp: message[index + 1],
$#each event eventHandlerField
        };
        this.emitter.emit(EVENT_$EVT_NAME, evt);
      },
`,
	// *******************************
	"eventHandlerField": `
          $fldName: message[index + 2],
`,
}
