package wasmclient

import (
	"context"
	"fmt"
	"strings"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	websocketservice "github.com/iotaledger/wasp/packages/webapi/websocket"
	"github.com/iotaledger/wasp/packages/webapi/websocket/commands"
)

type ContractEvent struct {
	ChainID    wasmtypes.ScChainID
	ContractID wasmtypes.ScHname
	Data       string
}

type WasmClientEvents struct {
	chainID    wasmtypes.ScChainID
	contractID wasmtypes.ScHname
	handler    wasmlib.IEventHandlers
}

func startEventLoop(url string, eventDone chan bool, eventHandlers *[]*WasmClientEvents) error {
	ctx := context.Background()
	ws, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return err
	}
	err = subscribe(ctx, ws, "chains")
	if err != nil {
		return err
	}
	err = subscribe(ctx, ws, "block_events")
	if err != nil {
		return err
	}

	go func() {
		<-eventDone
		_ = ws.Close(websocket.StatusNormalClosure, "intentional close")
	}()

	go eventLoop(ctx, ws, eventHandlers)

	return nil
}

func eventLoop(ctx context.Context, ws *websocket.Conn, eventHandlers *[]*WasmClientEvents) {
	for {
		evt := websocketservice.ISCEvent{}
		err := wsjson.Read(ctx, ws, &evt)
		if err != nil {
			return
		}
		items := evt.Payload.([]interface{})
		for _, item := range items {
			parts := strings.Split(item.(string), ": ")
			event := ContractEvent{
				ChainID:    wasmtypes.ChainIDFromString(evt.ChainID),
				ContractID: wasmtypes.HnameFromString(parts[0]),
				Data:       parts[1],
			}
			for _, h := range *eventHandlers {
				h.ProcessEvent(&event)
			}
		}
	}
}

func (h WasmClientEvents) ProcessEvent(event *ContractEvent) {
	if event.ContractID != h.contractID || event.ChainID != h.chainID {
		return
	}
	sep := strings.Index(event.Data, "|")
	if sep < 0 {
		return
	}
	topic := event.Data[:sep]
	fmt.Printf("%s %s %s\n", event.ChainID.String(), event.ContractID.String(), topic)
	buf := wasmtypes.HexDecode(event.Data[sep+1:])
	dec := wasmtypes.NewWasmDecoder(buf)
	h.handler.CallHandler(topic, dec)
}

func RemoveHandler(eventHandlers []*WasmClientEvents, eventsID uint32) []*WasmClientEvents {
	eh := eventHandlers[:0]
	for _, h := range eventHandlers {
		if h.handler.ID() != eventsID {
			eh = append(eh, h)
		}
	}
	return eh
}

func subscribe(ctx context.Context, ws *websocket.Conn, topic string) error {
	msg := commands.SubscriptionCommand{
		BaseCommand: commands.BaseCommand{
			Command: commands.CommandSubscribe,
		},
		Topic: topic,
	}
	err := wsjson.Write(ctx, ws, msg)
	if err != nil {
		return err
	}
	return wsjson.Read(ctx, ws, &msg)
}
