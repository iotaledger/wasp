package chainutil

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func EVMTraceTransaction(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log *logger.Logger,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	txIndex *uint64,
	blockNumber *uint64,
	tracer *tracers.Tracer,
) error {
	_, err := runISCTask(
		anchor,
		l1Params,
		store,
		processors,
		log,
		blockTime,
		iscRequestsInBlock,
		&isc.EVMTracer{
			Tracer:      tracer,
			TxIndex:     txIndex,
			BlockNumber: blockNumber,
		},
	)
	return err
}
