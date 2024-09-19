package chainutil

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func SimulateRequest(
	ch chain.ChainCore,
	req isc.Request,
	estimateGas bool,
) (*blocklog.RequestReceipt, error) {
	anchor, err := ch.LatestAnchor(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, fmt.Errorf("could not get latest AliasOutput: %w", err)
	}
	res, err := runISCRequest(ch, anchor, time.Now(), req, estimateGas)
	if err != nil {
		return nil, err
	}
	return res.Receipt, nil
}
