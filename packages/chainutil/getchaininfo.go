package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func GetAccountBalance(ch chain.ChainCore, agentID isc.AgentID) (*isc.FungibleTokens, error) {
	params := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	})
	ret, err := CallView(mustLatestState(ch), ch, accounts.Contract.Hname(), accounts.ViewBalance.Hname(), params)
	if err != nil {
		return nil, err
	}
	return isc.FungibleTokensFromDict(ret)
}
