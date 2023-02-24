package websocket

import (
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

type ISCEvent struct {
	Kind      publisher.ISCEventType `json:"kind"`
	Issuer    string                 `json:"issuer"`    // (isc.AgentID) nil means issued by the VM
	RequestID string                 `json:"requestID"` // (isc.RequestID)
	ChainID   string                 `json:"chainID"`   // (isc.ChainID)
	Payload   any                    `json:"payload"`
}

func MapISCEvent[T any](iscEvent *publisher.ISCEvent[T], mappedPayload any) *ISCEvent {
	issuer := iscEvent.Issuer.String()

	if issuer == "-" {
		// If the agentID is nil, it should be empty in the JSON result, not '-'
		issuer = ""
	}

	return &ISCEvent{
		Kind:      iscEvent.Kind,
		ChainID:   iscEvent.ChainID.String(),
		RequestID: iscEvent.RequestID.String(),
		Issuer:    issuer,
		Payload:   mappedPayload,
	}
}

type EventHandler struct {
	publisher             *publisher.Publisher
	publishEvent          *event.Event[*ISCEvent]
	subscriptionValidator *SubscriptionValidator

	// We need to store the closures to be able to Detach the hooks on shutdown.
	newBlockClosure       *event.Closure[*publisher.ISCEvent[*blocklog.BlockInfo]]
	requestReceiptClosure *event.Closure[*publisher.ISCEvent[*publisher.ReceiptWithError]]
	blockEventsClosure    *event.Closure[*publisher.ISCEvent[[]string]]
}

func NewEventHandler(pub *publisher.Publisher, publishEvent *event.Event[*ISCEvent], subscriptionValidator *SubscriptionValidator) *EventHandler {
	return &EventHandler{
		publisher:             pub,
		publishEvent:          publishEvent,
		subscriptionValidator: subscriptionValidator,
	}
}

func (p *EventHandler) attachToNewBlockEvent() {
	p.newBlockClosure = event.NewClosure(func(block *publisher.ISCEvent[*blocklog.BlockInfo]) {
		if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
			return
		}

		blockInfo := models.MapBlockInfoResponse(block.Payload)
		iscEvent := MapISCEvent(block, blockInfo)
		p.publishEvent.Trigger(iscEvent)
	})

	p.publisher.Events.NewBlock.Attach(p.newBlockClosure)
}

func (p *EventHandler) attachToRequestReceiptEvent() {
	p.requestReceiptClosure = event.NewClosure(func(block *publisher.ISCEvent[*publisher.ReceiptWithError]) {
		if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
			return
		}

		receipt := models.MapReceiptResponse(block.Payload.RequestReceipt, block.Payload.Error)
		iscEvent := MapISCEvent(block, receipt)
		p.publishEvent.Trigger(iscEvent)
	})

	p.publisher.Events.RequestReceipt.Attach(p.requestReceiptClosure)
}

func (p *EventHandler) attachToBlockEventsEvent() {
	p.blockEventsClosure = event.NewClosure(func(block *publisher.ISCEvent[[]string]) {
		if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
			return
		}

		iscEvent := MapISCEvent(block, block.Payload)
		p.publishEvent.Trigger(iscEvent)
	})

	p.publisher.Events.BlockEvents.Attach(p.blockEventsClosure)
}

func (p *EventHandler) AttachToEvents() {
	p.attachToNewBlockEvent()
	p.attachToRequestReceiptEvent()
	p.attachToBlockEventsEvent()
}

func (p *EventHandler) DetachEvents() {
	p.publisher.Events.NewBlock.Detach(p.newBlockClosure)
	p.publisher.Events.RequestReceipt.Detach(p.requestReceiptClosure)
	p.publisher.Events.BlockEvents.Detach(p.blockEventsClosure)
}
