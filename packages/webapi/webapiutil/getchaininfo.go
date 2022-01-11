package webapiutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func GetAccountBalance(ch chain.Chain, agentID *iscp.AgentID) (*iscp.Assets, error) {
	params := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	})
	ret, err := CallView(ch, accounts.Contract.Hname(), accounts.FuncViewBalance.Hname(), params)
	if err != nil {
		return nil, err
	}
	return iscp.AssetsFromDict(ret)
}
