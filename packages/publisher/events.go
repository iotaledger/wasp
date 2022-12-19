package publisher

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

const (
	ISCEventKindNewBlock      = "new_block"
	ISCEventKindReceipt       = "receipt" // issuer will be the request sender
	ISCEventKindSmartContract = "contract"
)

const ISCEventIssuerVM = "vm"

type ISCEvent struct {
	Kind      string
	Issuer    isc.AgentID // nil means issued by the VM
	RequestID isc.RequestID
	Content   string
}

// kind is not printed right now, because it is added when calling p.publish
func (e *ISCEvent) String() string {
	issuerStr := "vm"
	if e.Issuer != nil {
		issuerStr = e.Issuer.String()
	}
	return fmt.Sprintf("%s - %s", issuerStr, e.Content)
}

// EventsFromBlock extracts the events from a block, its returns a chan of ISCEvents so they can be filtered
func EventsFromBlock(block state.Block) (chan ISCEvent, chan error) {
	resultCh := make(chan ISCEvent)
	errCh := make(chan error)

	go func() {
		defer close(resultCh)
		defer close(errCh)

		//
		// Publish notifications about the state change (new block).
		blockIndex := block.StateIndex()
		blocklogStatePartition := subrealm.NewReadOnly(block.MutationsReader(), kv.Key(blocklog.Contract.Hname().Bytes()))
		blockInfo, err := blocklog.GetBlockInfo(blocklogStatePartition, blockIndex)
		if err != nil {
			errCh <- fmt.Errorf("Unable to get blockInfo for blockIndex %d: %w", blockIndex, err)
			return
		}
		resultCh <- ISCEvent{
			Kind:   ISCEventKindNewBlock,
			Issuer: nil,
			// TODO should probably be JSON? right now its just some printed strings
			// TODO the L1 commitment will be nil (on the blocklog), but at this point the L1 commitment has already been calculated, so we could potentially add it to blockInfo
			Content: blockInfo.String(),
		}

		//
		// Publish receipts of processed requests.
		receipts, err := blocklog.RequestReceiptsFromBlock(block)
		if err != nil {
			errCh <- fmt.Errorf("Unable to get receipts from a block: %w", err)
		} else {
			for _, receipt := range receipts {
				resultCh <- ISCEvent{
					Kind:      ISCEventKindReceipt,
					Issuer:    receipt.Request.SenderAccount(),
					Content:   receipt.String(),
					RequestID: receipt.Request.ID(),
				}
			}
		}

		//
		// Publish contract-issued events.
		events, err := blocklog.GetEventsByBlockIndex(blocklogStatePartition, blockIndex, blockInfo.TotalRequests)
		if err != nil {
			errCh <- fmt.Errorf("Unable to get events from a block: %w", err)
		} else {
			for _, event := range events {
				resultCh <- ISCEvent{
					Kind: ISCEventKindSmartContract,
					// TODO should be the contract Hname, but right now events are just stored as strings.
					// must be refactored so its possible to filter by "events from a contract"
					Issuer: nil,
					// TODO should be possible to filter by request ID (not possible with current events impl)
					// RequestID: event.RequestID,
					Content: event,
				}
			}
		}
	}()

	return resultCh, errCh
}
