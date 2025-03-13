package chainutil

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func EVMTrace(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
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
		false,
		tracer,
	)
	return err
}
