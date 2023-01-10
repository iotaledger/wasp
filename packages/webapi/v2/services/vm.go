package services

import (
	"errors"

	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/packages/webapi/v2/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
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
	chain := v.chainsProvider().Get(chainID)
	if chain == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	blocklog := corecontracts.NewBlockLog(v)
	receipt, err := blocklog.GetRequestReceipt(chainID, requestID)
	if err != nil {
		return nil, nil, err
	}

	return v.ParseReceipt(chain, receipt)
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
	b := &viewcontext.BlockIndexOrTrieRoot{BlockIndex: blockIndex}
	return chainutil.CallView(chain, b, contractName, functionName, params)
}
