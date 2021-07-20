package webapiutil

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/plugins/chains"
)

func GetChain(chainID *iscp.ChainID) chain.Chain {
	return chains.AllChains().Get(chainID)
}

func GetAccountBalance(ch chain.ChainCore, agentID *iscp.AgentID) (map[ledgerstate.Color]uint64, error) {
	params := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	})
	ret, err := CallView(ch, accounts.Contract.Hname(), accounts.FuncViewBalance.Hname(), params)
	if err != nil {
		return nil, err
	}
	return accounts.DecodeBalances(ret)
}
