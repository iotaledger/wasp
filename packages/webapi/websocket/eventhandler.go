package websocket

import (
	"context"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/event"
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
	publishEvent          *event.Event1[*ISCEvent]
	subscriptionValidator *SubscriptionValidator
}

func NewEventHandler(pub *publisher.Publisher, publishEvent *event.Event1[*ISCEvent], subscriptionValidator *SubscriptionValidator) *EventHandler {
	return &EventHandler{
		publisher:             pub,
		publishEvent:          publishEvent,
		subscriptionValidator: subscriptionValidator,
	}
}

func (p *EventHandler) AttachToEvents() context.CancelFunc {
	return lo.Batch(

		p.publisher.Events.NewBlock.Hook(func(block *publisher.ISCEvent[*blocklog.BlockInfo]) {
			if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
				return
			}

			blockInfo := models.MapBlockInfoResponse(block.Payload)
			iscEvent := MapISCEvent(block, blockInfo)
			p.publishEvent.Trigger(iscEvent)
		}).Unhook,

		p.publisher.Events.RequestReceipt.Hook(func(block *publisher.ISCEvent[*publisher.ReceiptWithError]) {
			if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
				return
			}

			receipt := models.MapReceiptResponse(block.Payload.RequestReceipt, block.Payload.Error)
			iscEvent := MapISCEvent(block, receipt)
			p.publishEvent.Trigger(iscEvent)
		}).Unhook,

		p.publisher.Events.BlockEvents.Hook(func(block *publisher.ISCEvent[[]string]) {
			if !p.subscriptionValidator.shouldProcessEvent(block.ChainID.String(), block.Kind) {
				return
			}

			iscEvent := MapISCEvent(block, block.Payload)
			p.publishEvent.Trigger(iscEvent)
		}).Unhook,
	)
}
