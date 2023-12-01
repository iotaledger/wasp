package wasmclient

import (
	"context"
	"fmt"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SubscriptionCommand struct {
	Command string `json:"command"`
	Topic   string `json:"topic"`
}

type Event struct {
	ChainID    wasmtypes.ScChainID
	ContractID wasmtypes.ScHname `json:"contractID"`
	Payload    []byte            `json:"payload"`
	Timestamp  uint64            `json:"timestamp"`
	Topic      string            `json:"topic"`
}

func (e *Event) Bytes() []byte {
	enc := wasmtypes.NewWasmEncoder()
	wasmtypes.HnameEncode(enc, e.ContractID)
	wasmtypes.StringEncode(enc, e.Topic)
	wasmtypes.Uint64Encode(enc, e.Timestamp)
	enc.FixedBytes(e.Payload, uint32(len(e.Payload)))
	return enc.Buf()
}

type ISCEvent struct {
	Kind      string   `json:"kind"`
	Issuer    string   `json:"issuer"`    // (isc.AgentID) nil means issued by the VM
	RequestID string   `json:"requestID"` // (isc.RequestID)
	ChainID   string   `json:"chainID"`   // (isc.ChainID)
	Payload   []*Event `json:"payload"`
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
		evt := ISCEvent{}
		err := wsjson.Read(ctx, ws, &evt)
		if err != nil {
			return
		}
		items := evt.Payload
		for _, item := range items {
			event := NewContractEvent(evt.ChainID, item.Bytes())
			for _, h := range *eventHandlers {
				h.ProcessEvent(&event)
			}
		}
	}
}

func NewContractEvent(chainID string, eventData []byte) Event {
	dec := wasmtypes.NewWasmDecoder(eventData)
	hContract := wasmtypes.HnameDecode(dec)
	topic := wasmtypes.StringDecode(dec)
	payload := dec.FixedBytes(dec.Length())
	event := Event{
		ChainID:    wasmtypes.ChainIDFromString(chainID),
		ContractID: hContract,
		Payload:    payload,
		Timestamp:  wasmtypes.Uint64FromBytes(payload[:wasmtypes.ScUint64Length]),
		Topic:      topic,
	}
	return event
}

func (h WasmClientEvents) ProcessEvent(event *Event) {
	if event.ContractID != h.contractID || event.ChainID != h.chainID {
		return
	}
	fmt.Printf("%s %s %s\n", event.ChainID.String(), event.ContractID.String(), event.Topic)
	dec := wasmtypes.NewWasmDecoder(event.Payload)
	h.handler.CallHandler(event.Topic, dec)
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
	msg := SubscriptionCommand{
		Command: "subscribe",
		Topic:   topic,
	}
	err := wsjson.Write(ctx, ws, msg)
	if err != nil {
		return err
	}
	return wsjson.Read(ctx, ws, &msg)
}
