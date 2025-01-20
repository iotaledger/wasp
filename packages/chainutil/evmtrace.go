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
	tracer *tracers.Tracer,
) error {
	_, err := runISCTask(
		ch,
		aliasOutput,
		blockTime,
		iscRequestsInBlock,
		false,
		tracer,
	)
	return err
}
