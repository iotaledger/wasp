package services

import (
	"errors"

	"github.com/iotaledger/wasp/packages/chain/chainutil"

	"github.com/iotaledger/hive.go/core/logger"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type VMService struct {
	logger *logger.Logger

	chainsProvider chains.Provider
}

func NewVMService(log *logger.Logger, chainsProvider chains.Provider) interfaces.VM {
	return &VMService{
		logger: log,

		chainsProvider: chainsProvider,
	}
}

func (v *VMService) CallViewByChainID(chainID *isc.ChainID, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	chain := v.chainsProvider().Get(chainID)

	if chain == nil {
		return nil, errors.New("chain not found")
	}

	return v.CallView(chain, contractName, functionName, params)
}

func (v *VMService) CallView(chain chainpkg.Chain, contractName, functionName isc.Hname, params dict.Dict) (dict.Dict, error) {
	return chainutil.CallView(chain, contractName, functionName, params)
}
