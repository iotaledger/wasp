package chainutil

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
)

func EVMTrace(
	ch chain.ChainCore,
	aliasOutput *isc.AliasOutputWithID,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	txIndex *uint64,
	blockNumber *uint64,
	tracer *tracers.Tracer,
) error {
	_, err := runISCTask(
		ch,
		aliasOutput,
		blockTime,
		iscRequestsInBlock,
		false,
		&isc.EVMTracer{
			Tracer:      tracer,
			TxIndex:     txIndex,
			BlockNumber: blockNumber,
		},
	)
	return err
}
