package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func GetAccountBalance(ch chain.ChainCore, agentID isc.AgentID) (isc.CoinBalances, error) {
	ret, err := CallView(mustLatestState(ch), ch, accounts.ViewBalance.Message(&agentID))
	if err != nil {
		return nil, err
	}

	return accounts.ViewBalance.DecodeOutput(ret)
}

func mustLatestState(ch chain.ChainCore) state.State {
	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		panic(err)
	}
	return latestState
}
