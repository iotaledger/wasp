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
)

// PublishedEvent contains the information about the published message.
type PublishedEvent struct {
	MsgType string
	ChainID isc.ChainID
	Parts   []string
}

type Events struct {
	BlockApplied *event.Event[*BlockApplied]
}

////////////////////////////////////////////////////////////////////////////////
// Publisher

type Publisher struct {
	blockAppliedPipe pipe.Pipe[*BlockApplied]
	mutex            *sync.RWMutex
	log              *logger.Logger
	Events           *Events
}

var _ chain.ChainListener = &Publisher{}

type BlockApplied struct {
	ChainID isc.ChainID
	Block   state.Block
}

func New(log *logger.Logger) *Publisher {
	p := &Publisher{
		blockAppliedPipe: pipe.NewInfinitePipe[*BlockApplied](),
		mutex:            &sync.RWMutex{},
		log:              log,
		Events: &Events{
			BlockApplied: event.New[*BlockApplied](),
		},
	}

	return p
}

// Implements the chain.ChainListener interface.
// NOTE: Do not Block the caller!
func (p *Publisher) BlockApplied(chainID isc.ChainID, block state.Block) {
	p.blockAppliedPipe.In() <- &BlockApplied{ChainID: chainID, Block: block}
}

// Implements the chain.ChainListener interface.
// NOTE: Do not Block the caller!
func (p *Publisher) AccessNodesUpdated(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	// We don't need this event.
}

// Implements the chain.ChainListener interface.
// NOTE: Do not Block the caller!
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

			p.Events.BlockApplied.Trigger(blockAppliedUntyped)
		case <-ctx.Done():
			return
		}
	}
}
