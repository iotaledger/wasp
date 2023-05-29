package common

import (
	"errors"

	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func ParseReceipt(chain chainpkg.Chain, receipt *blocklog.RequestReceipt) (*isc.Receipt, error) {
	resolvedReceiptErr, err := chainutil.ResolveError(chain, receipt.Error)
	if err != nil {
		return nil, err
	}

	iscReceipt := receipt.ToISCReceipt(resolvedReceiptErr)

	return iscReceipt, nil
}

func CallView(ch chainpkg.Chain, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	// TODO: should blockIndex be an optional parameter of this endpoint?
	latestState, err := ch.LatestState(chainpkg.ActiveOrCommittedState)
	if err != nil {
		return nil, errors.New("error getting latest chain state")
	}
	return chainutil.CallView(latestState, ch, contractName, functionName, params)
}

func EstimateGas(ch chainpkg.Chain, req isc.Request) (*isc.Receipt, error) {
	rec, err := chainutil.SimulateRequest(ch, req)
	if err != nil {
		return nil, err
	}
	parsedRec, err := ParseReceipt(ch, rec)
	return parsedRec, err
}
