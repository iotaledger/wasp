package chainutil

import (
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func SimulateRequest(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log *logger.Logger,
	req isc.Request,
) (*blocklog.RequestReceipt, error) {
	res, err := runISCRequest(
		anchor,
		l1Params,
		store,
		processors,
		log,
		time.Now(),
		req,
	)
	if err != nil {
		return nil, err
	}
	return res.Receipt, nil
}
