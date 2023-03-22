package publisher

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
)

type ISCEventType string

const (
	ISCEventKindNewBlock    ISCEventType = "new_block"
	ISCEventKindReceipt     ISCEventType = "receipt" // issuer will be the request sender
	ISCEventKindBlockEvents ISCEventType = "block_events"
	ISCEventIssuerVM        ISCEventType = "vm"
)

type ISCEvent[T any] struct {
	Kind      ISCEventType  `json:"kind"`
	Issuer    isc.AgentID   `json:"issuer"`    // (AgentID) nil means issued by the VM
	RequestID isc.RequestID `json:"requestID"` // (isc.RequestID)
	ChainID   isc.ChainID   `json:"chainID"`   // (isc.ChainID)
	Payload   T             `json:"payload"`
}

// kind is not printed right now, because it is added when calling p.publish
func (e *ISCEvent[T]) String() string {
	issuerStr := "vm"
	if e.Issuer != nil {
		issuerStr = e.Issuer.String()
	}

	return fmt.Sprintf("%s | %s (%s)", e.ChainID, issuerStr, e.Kind)
}

type ReceiptWithError struct {
	RequestReceipt *isc.Receipt
	Error          *isc.VMError
}

func triggerEvent[T any](events *Events, event *event.Event1[*ISCEvent[T]], obj *ISCEvent[T]) {
	event.Trigger(obj)

	// To support Solo and other consumers, push each event into the Published event
	// It's basically a catch-all event for all publisher events.
	// Otherwise Solo and other consumers would have to subscribe to each event manually,
	// and we would have to make sure that each new event gets added there too.
	events.Published.Trigger(&ISCEvent[any]{
		Kind:      obj.Kind,
		Issuer:    obj.Issuer,
		RequestID: obj.RequestID,
		ChainID:   obj.ChainID,
		Payload:   obj.Payload,
	})
}

// PublishBlockEvents extracts the events from a block, its returns a chan of ISCEventType, so they can be filtered
func PublishBlockEvents(blockApplied *blockApplied, events *Events, log *logger.Logger) {
	block := blockApplied.block
	chainID := blockApplied.chainID
	//
	// Publish notifications about the state change (new block).
	blockIndex := block.StateIndex()
	blocklogStatePartition := subrealm.NewReadOnly(block.MutationsReader(), kv.Key(blocklog.Contract.Hname().Bytes()))
	blockInfo, err := blocklog.GetBlockInfo(blocklogStatePartition, blockIndex)
	if err != nil {
		log.Errorf("unable to get blockInfo for blockIndex %d: %v", blockIndex, err)
	}

	triggerEvent(events, events.NewBlock, &ISCEvent[*blocklog.BlockInfo]{
		Kind:   ISCEventKindNewBlock,
		Issuer: &isc.NilAgentID{},
		// TODO the L1 commitment will be nil (on the blocklog), but at this point the L1 commitment has already been calculated, so we could potentially add it to blockInfo
		Payload: blockInfo,
		ChainID: chainID,
	})

	//
	// Publish receipts of processed requests.
	receipts, err := blocklog.RequestReceiptsFromBlock(block)

	if err != nil {
		log.Errorf("unable to get receipts from a block: %v", err)
	} else {
		for index, receipt := range receipts {
			vmError, resolveError := errors.ResolveFromState(blocklogStatePartition, receipt.Error)
			if resolveError != nil {
				log.Errorf("Could not parse vmerror of receipt [%v]: %v", receipt.Request.ID(), resolveError)
			}

			// TODO: Validate logic here:
			receipt.BlockIndex = blockIndex
			receipt.RequestIndex = uint16(index)

			parsedReceipt := receipt.ToISCReceipt(vmError)

			triggerEvent(events, events.RequestReceipt, &ISCEvent[*ReceiptWithError]{
				Kind:      ISCEventKindReceipt,
				Issuer:    receipt.Request.SenderAccount(),
				Payload:   &ReceiptWithError{RequestReceipt: parsedReceipt, Error: vmError},
				RequestID: receipt.Request.ID(),
				ChainID:   chainID,
			})
		}
	}

	//
	// Publish contract-issued events.
	blockEvents, err := blocklog.GetEventsByBlockIndex(blocklogStatePartition, blockIndex, blockInfo.TotalRequests)
	if err != nil {
		log.Errorf("unable to get events from a block: %v", err)
	} else {
		triggerEvent(events, events.BlockEvents, &ISCEvent[[]string]{
			Kind: ISCEventKindBlockEvents,
			// TODO should be the contract Hname, but right now events are just stored as strings.
			// must be refactored so its possible to filter by "events from a contract"
			Issuer: &isc.NilAgentID{},
			// TODO should be possible to filter by request ID (not possible with current events impl)
			// RequestID: event.RequestID,
			Payload: blockEvents,
			ChainID: chainID,
		})
	}
}
