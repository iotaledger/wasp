package publisher

import (
	"context"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

var Event = events.NewEvent(func(handler interface{}, params ...interface{}) {
	callback := handler.(func(msgType string, parts []string))
	msgType := params[0].(string)
	parts := params[1].([]string)
	callback(msgType, parts)
})

////////////////////////////////////////////////////////////////////////////////
// Publisher

type Publisher struct {
	blockAppliedPipe pipe.Pipe
	mutex            *sync.RWMutex
	log              *logger.Logger
}

var _ chain.ChainListener = &Publisher{}

type publisherBlockApplied struct {
	chainID isc.ChainID
	block   state.Block
}

func NewPublisher(log *logger.Logger) *Publisher {
	p := &Publisher{
		blockAppliedPipe: pipe.NewDefaultInfinitePipe(),
		mutex:            &sync.RWMutex{},
		log:              log,
	}
	return p
}

// Implements the chain.ChainListener interface.
// NOTE: Do not block the caller!
func (p *Publisher) BlockApplied(chainID isc.ChainID, block state.Block) {
	p.blockAppliedPipe.In() <- &publisherBlockApplied{chainID: chainID, block: block}
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
			p.handleBlockApplied(blockAppliedUntyped.(*publisherBlockApplied))
		case <-ctx.Done():
			return
		}
	}
}

func (p *Publisher) handleBlockApplied(blockApplied *publisherBlockApplied) {
	stateIndex := blockApplied.block.StateIndex()
	p.log.Debugf("BlockApplied, chainID=%v, stateIndex=%v", blockApplied.chainID.String(), stateIndex)

	eventsCh, errCh := EventsFromBlock(blockApplied.block)

	for {
		select {
		case event, ok := <-eventsCh:
			if !ok {
				return
			}
			p.publish(event.Kind,
				blockApplied.chainID.String(),
				event.String(),
			)
		case err, ok := <-errCh:
			if !ok {
				return
			}
			if err != nil {
				p.log.Error(err.Error())
			}
		}
	}
}

func (p *Publisher) publish(msgType string, parts ...string) {
	p.log.Debugf("Publishing %v: %v", msgType, strings.Join(parts, ", "))
	Event.Trigger(msgType, parts)
}
