package vmcontext

import "github.com/iotaledger/wasp/packages/coretypes"

type accountsWrapper struct {
	vmctx *VMContext
}

func (a *accountsWrapper) BalancesImmutable(agentID coretypes.AgentID) (coretypes.ColoredBalancesImmutable, bool) {
	panic("implement me")
}

func (a *accountsWrapper) BalancesSpendable(agentID coretypes.AgentID) (coretypes.ColoredBalancesSpendable, bool) {
	panic("implement me")
}

func (a *accountsWrapper) Balances(agentID coretypes.AgentID) (coretypes.ColoredBalancesMutable, bool) {
	panic("implement me")
}

func (a *accountsWrapper) Iterate(f func(agentID coretypes.AgentID) bool) {
	panic("implement me")
}

func (a *accountsWrapper) IterateDeterministic(f func(id coretypes.AgentID) bool) {
	panic("implement me")
}

func (a *accountsWrapper) Create(agentID coretypes.AgentID) bool {
	panic("implement me")
}

func (a *accountsWrapper) Incoming() coretypes.ColoredBalancesSpendable {
	panic("implement me")
}
