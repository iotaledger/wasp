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
	Published *event.Event[*PublishedEvent]
}

////////////////////////////////////////////////////////////////////////////////
// Publisher

type Publisher struct {
	blockAppliedPipe pipe.Pipe[*publisherBlockApplied]
	mutex            *sync.RWMutex
	log              *logger.Logger
	Events           *Events
}

var _ chain.ChainListener = &Publisher{}

type publisherBlockApplied struct {
	chainID isc.ChainID
	block   state.Block
}

func New(log *logger.Logger) *Publisher {
	p := &Publisher{
		blockAppliedPipe: pipe.NewInfinitePipe[*publisherBlockApplied](),
		mutex:            &sync.RWMutex{},
		log:              log,
		Events: &Events{
			Published: event.New[*PublishedEvent](),
		},
	}

	return p
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) BlockApplied(chainID isc.ChainID, block state.Block) {
	p.blockAppliedPipe.In() <- &publisherBlockApplied{chainID: chainID, block: block}
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
			p.handleBlockApplied(blockAppliedUntyped)
		case <-ctx.Done():
			return
		}
	}
}

func (p *Publisher) handleBlockApplied(blockApplied *publisherBlockApplied) {
	stateIndex := blockApplied.block.StateIndex()
	p.log.Debugf("BlockApplied, chainID=%v, stateIndex=%v", blockApplied.chainID.String(), stateIndex)
	PublishBlockEvents(blockApplied, p.publish, p.log)
}

func (p *Publisher) publish(e *ISCEvent) {
	p.log.Debugf("Publishing %v", e.String())

	p.Events.Published.Trigger(&PublishedEvent{
		MsgType: e.Kind,
		ChainID: e.ChainID,
		Parts:   []string{e.String()},
	})
}
