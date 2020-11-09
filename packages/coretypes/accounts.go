package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

// ColoredBalancesImmutable read only
type ColoredBalancesImmutable interface {
	Balance(color balance.Color) (int64, bool)
	Iterate(func(color balance.Color, balance int64) bool)
	IterateDeterministic(func(color balance.Color, balance int64) bool)
}

// ColoredBalancesSpendable can be spent, can't be added
type ColoredBalancesSpendable interface {
	ColoredBalancesImmutable
	// Spend transfers tokens from receiver to target
	Spend(target ColoredBalancesMutable, color balance.Color, bal int64) bool
	SpendAll(target ColoredBalancesMutable)
}

// ColoredBalancesMutable all functions
type ColoredBalancesMutable interface {
	ColoredBalancesSpendable
	Add(color balance.Color, bal int64) bool
}

type ColoredAccountsImmutable interface {
	BalancesImmutable(agentID AgentID) (ColoredBalancesImmutable, bool)
	Iterate(func(agentID AgentID) bool)
	IterateDeterministic(func(id AgentID) bool)
}

// ColoredAccounts interface to a collection of account.
type ColoredAccounts interface {
	ColoredAccountsImmutable
	BalancesSpendable(agentID AgentID) (ColoredBalancesSpendable, bool)
	Balances(agentID AgentID) (ColoredBalancesMutable, bool)
	Create(agentID AgentID) bool
}
