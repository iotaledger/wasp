package services

import (
	"errors"

	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type VMService struct {
	chainsProvider chains.Provider
}

func NewVMService(chainsProvider chains.Provider) interfaces.VMService {
	return &VMService{
		chainsProvider: chainsProvider,
	}
}

func (v *VMService) ParseReceipt(chain chainpkg.Chain, receipt *blocklog.RequestReceipt) (*isc.Receipt, *isc.VMError, error) {
	resolvedReceiptErr, err := chainutil.ResolveError(chain, receipt.Error)
	if err != nil {
		return nil, nil, err
	}

	iscReceipt := receipt.ToISCReceipt(resolvedReceiptErr)

	return iscReceipt, resolvedReceiptErr, nil
}

func (v *VMService) GetReceipt(chainID isc.ChainID, requestID isc.RequestID) (*isc.Receipt, *isc.VMError, error) {
	ch := v.chainsProvider().Get(chainID)
	if ch == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	blocklog := corecontracts.NewBlockLog(v)
	receipt, err := blocklog.GetRequestReceipt(chainID, requestID)
	if err != nil {
		return nil, nil, err
	}

	return v.ParseReceipt(ch, receipt)
}

func (v *VMService) CallViewByChainID(chainID isc.ChainID, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	ch := v.chainsProvider().Get(chainID)
	if ch == nil {
		return nil, errors.New("chain not found")
	}

	// TODO: should blockIndex be an optional parameter of this endpoint?
	latestState, err := ch.LatestState(chainpkg.ActiveOrCommittedState)
	if err != nil {
		return nil, errors.New("error getting latest chain state")
	}
	return v.CallView(latestState, ch, contractName, functionName, params)
}

func (v *VMService) CallView(chainState state.State, chain chainpkg.Chain, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	return chainutil.CallView(chainState, chain, contractName, functionName, params)
}
