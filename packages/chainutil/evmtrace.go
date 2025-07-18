package chainutil

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
)

func EVMTrace(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	blockTime time.Time,
	entropy hashing.HashValue,
	iscRequestsInBlock []isc.Request,
	enforceGasBurned []vm.EnforceGasBurned,
	tracer *tracers.Tracer,
) error {
	_, err := runISCTask(
		anchor,
		l1Params,
		store,
		processors,
		log,
		blockTime,
		entropy,
		iscRequestsInBlock,
		enforceGasBurned,
		false,
		tracer,
	)
	return err
}
