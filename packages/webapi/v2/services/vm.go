package services

import (
	"errors"

	"github.com/iotaledger/hive.go/core/logger"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type VMService struct {
	logger *logger.Logger

	chainsProvider chains.Provider
}

func NewVMService(log *logger.Logger, chainsProvider chains.Provider) interfaces.VMService {
	return &VMService{
		logger:         log,
		chainsProvider: chainsProvider,
	}
}

func (v *VMService) getReceipt(chainID isc.ChainID, requestID isc.RequestID) (*isc.Receipt, *isc.VMError, error) {
	chain := v.chainsProvider().Get(chainID)
	if chain == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	receipt, err := chain.GetRequestReceipt(requestID)
	if err != nil {
		return nil, nil, err
	}

	resolvedError, err := chain.ResolveError(receipt.Error)
	if err != nil {
		return nil, nil, err
	}

	receiptData := receipt.ToISCReceipt(resolvedError)

	return receiptData, resolvedError, nil
}

func (v *VMService) GetReceipt(chainID isc.ChainID, requestID isc.RequestID) (ret *isc.Receipt, vmError *isc.VMError, err error) {
	err = optimism.RetryOnStateInvalidated(func() (err error) {
		panicCatchErr := panicutil.CatchPanicReturnError(func() {
			ret, vmError, err = v.getReceipt(chainID, requestID)
		}, coreutil.ErrorStateInvalidated)
		if err != nil {
			return err
		}
		return panicCatchErr
	})
	return ret, vmError, err
}

func (v *VMService) CallViewByChainID(chainID isc.ChainID, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	chain := v.chainsProvider().Get(chainID)

	if chain == nil {
		return nil, errors.New("chain not found")
	}

	// TODO: should blockIndex be an optional parameter of this endpoint?
	blockIndex, err := chain.GetStateReader().LatestBlockIndex()
	if err != nil {
		return nil, errors.New("error getting latest chain block index")
	}

	return v.CallView(blockIndex, chain, contractName, functionName, params)
}

func (v *VMService) CallView(blockIndex uint32, chain chainpkg.Chain, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	return chainutil.CallView(blockIndex, chain, contractName, functionName, params)
}
