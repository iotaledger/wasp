package chainutil

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
)

func EVMTraceTransaction(
	ch chain.ChainCore,
	aliasOutput *isc.AliasOutputWithID,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	txIndex uint64,
	tracer tracers.Tracer,
) error {
	_, err := runISCTask(
		ch,
		aliasOutput,
		blockTime,
		iscRequestsInBlock,
		false,
		&isc.EVMTracer{
			Tracer:  tracer,
			TxIndex: txIndex,
		},
	)
	return err
}
