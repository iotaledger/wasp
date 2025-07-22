package chainutil

import (
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
)

func SimulateRequest(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	req isc.Request,
	estimateGasMode bool,
) (*blocklog.RequestReceipt, error) {
	res, err := runISCRequest(
		anchor,
		l1Params,
		store,
		processors,
		log,
		time.Now(),
		hashing.PseudoRandomHash(nil),
		req,
		estimateGasMode,
	)
	if err != nil {
		return nil, err
	}
	return res.Receipt, nil
}
