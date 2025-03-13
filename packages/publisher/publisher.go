package publisher

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

type Events struct {
	BlockEvents    *event.Event1[*ISCEvent[[]*isc.Event]]
	NewBlock       *event.Event1[*ISCEvent[*BlockWithTrieRoot]]
	RequestReceipt *event.Event1[*ISCEvent[*ReceiptWithError]]

	Published *event.Event1[*ISCEvent[any]]
}

////////////////////////////////////////////////////////////////////////////////
// Publisher

type Publisher struct {
	blockAppliedPipe pipe.Pipe[*blockApplied]
	mutex            *sync.RWMutex
	log              log.Logger
	Events           *Events
}

// commented out to avoid importing chain package here
// var _ chain.ChainListener = &Publisher{}

type blockApplied struct {
	chainID     isc.ChainID
	block       state.Block
	latestState kv.KVStoreReader
}

func New(log log.Logger) *Publisher {
	p := &Publisher{
		blockAppliedPipe: pipe.NewInfinitePipe[*blockApplied](),
		mutex:            &sync.RWMutex{},
		log:              log,
		Events: &Events{
			NewBlock:       event.New1[*ISCEvent[*BlockWithTrieRoot]](),
			RequestReceipt: event.New1[*ISCEvent[*ReceiptWithError]](),
			BlockEvents:    event.New1[*ISCEvent[[]*isc.Event]](),
			Published:      event.New1[*ISCEvent[any]](),
		},
	}

	return p
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) BlockApplied(chainID isc.ChainID, block state.Block, latestState kv.KVStoreReader) {
	p.blockAppliedPipe.In() <- &blockApplied{chainID: chainID, block: block, latestState: latestState}
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) AccessNodesUpdated(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	// We don't need this event.
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) ServerNodesUpdated(chainID isc.ChainID, serverNodes []*cryptolib.PublicKey) {
	// We don't need this event.
}

// This is called by the component to run this.
func (p *Publisher) Run(ctx context.Context) {
	blockAppliedPipeOutCh := p.blockAppliedPipe.Out()
	for {
		select {
		case blockAppliedUntyped, ok := <-blockAppliedPipeOutCh:
			if !ok {
				blockAppliedPipeOutCh = nil
				continue
			}

			PublishBlockEvents(blockAppliedUntyped, p.Events, p.log)
		case <-ctx.Done():
			return
		}
	}
}
