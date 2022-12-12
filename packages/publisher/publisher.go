package publisher

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
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
	running          *atomic.Bool
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
		running:          &atomic.Bool{},
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
	//
	// Publish notifications on the processed requests.
	receipts, err := blocklog.RequestReceiptsFromBlock(blockApplied.block)
	if err != nil {
		p.log.Warnf("Unable to get receipts from a block: %v", err)
		return
	}
	for _, receipt := range receipts {
		p.publish("request_out",
			blockApplied.chainID.String(),
			receipt.Request.ID().String(),
			strconv.Itoa(int(stateIndex)),
			strconv.Itoa(len(receipts)),
		)
	}
	//
	// Publish notifications on the VM events / messages.
	events, err := blocklog.GetBlockEventsInternal(blockApplied.block.MutationsReader(), blockApplied.block.StateIndex())
	if err != nil {
		p.log.Warnf("Unable to get events from a block: %v", err)
		return
	}
	for _, event := range events {
		p.publish("vmmsg",
			blockApplied.chainID.String(),
			event,
		)
	}
	// TODO: publisher.Publish("state",
	// 	chainID.String(),
	// 	strconv.Itoa(int(stateOutput.GetStateIndex())),
	// 	strconv.Itoa(reqIDsLength),
	// 	isc.OID(stateOutput.ID()),
	// 	stateHash.String(),
	// )
	// TODO: publisher.Publish("rotate",
	// 	chainID.String(),
	// 	strconv.Itoa(int(stateOutput.GetStateIndex())),
	// 	isc.OID(stateOutput.ID()),
	// 	stateHash.String(),
	// )
	// TODO: publisher.Publish("dismissed_chain", c.chainID.String())
}

func (p *Publisher) publish(msgType string, parts ...string) {
	p.log.Debugf("Publishing %v: %v", msgType, strings.Join(parts, ", "))
	Event.Trigger(msgType, parts)
}
