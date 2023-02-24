package publisher

import (
	"context"
	"sync"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type Events struct {
	BlockEvents    *event.Event[*ISCEvent[[]string]]
	NewBlock       *event.Event[*ISCEvent[*blocklog.BlockInfo]]
	RequestReceipt *event.Event[*ISCEvent[*ReceiptWithError]]

	Published *event.Event[*ISCEvent[any]]
}

////////////////////////////////////////////////////////////////////////////////
// Publisher

type Publisher struct {
	blockAppliedPipe pipe.Pipe[*blockApplied]
	mutex            *sync.RWMutex
	log              *logger.Logger
	Events           *Events
}

var _ chain.ChainListener = &Publisher{}

type blockApplied struct {
	chainID isc.ChainID
	block   state.Block
}

func New(log *logger.Logger) *Publisher {
	p := &Publisher{
		blockAppliedPipe: pipe.NewInfinitePipe[*blockApplied](),
		mutex:            &sync.RWMutex{},
		log:              log,
		Events: &Events{
			NewBlock:       event.New[*ISCEvent[*blocklog.BlockInfo]](),
			RequestReceipt: event.New[*ISCEvent[*ReceiptWithError]](),
			BlockEvents:    event.New[*ISCEvent[[]string]](),
			Published:      event.New[*ISCEvent[any]](),
		},
	}

	return p
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) BlockApplied(chainID isc.ChainID, block state.Block) {
	p.blockAppliedPipe.In() <- &blockApplied{chainID: chainID, block: block}
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
