package wasmsolo

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type SoloBalances struct {
	Account    uint64
	accounts   map[string]uint64
	agents     []*SoloAgent
	Chain      uint64
	ctx        *SoloContext
	Originator uint64
}

// NewSoloBalances takes a snapshot of all balances necessary to track token
// movements easily. It will track L2 Originator, Chain, snd SC Account balances
// Additional agents can be specified as extra accounts
// This is typically called from SoloContext.Balances() before a call to the SC.
// After the call, update the balances with the expected token movements and then
// call SoloBalances.VerifyBalances() to see if they match the actual balances.
func NewSoloBalances(ctx *SoloContext, agents ...*SoloAgent) *SoloBalances {
	bal := &SoloBalances{
		ctx:        ctx,
		Account:    ctx.Balance(ctx.Account()),
		Chain:      ctx.Balance(ctx.ChainAccount()),
		Originator: ctx.Balance(ctx.Originator()),
		agents:     agents,
		accounts:   make(map[string]uint64),
	}
	for _, agent := range agents {
		bal.accounts[agent.AgentID().String()] = ctx.Balance(agent)
	}
	bal.dumpBalances()
	return bal
}

// dumpBalances prints all known accounts, both L2 and L1, in debug mode.
// It uses the L2 ledger to enumerate the known accounts.
func (bal *SoloBalances) dumpBalances() {
	if !SoloDebug {
		return
	}
	ctx := bal.ctx
	accs := ctx.Chain.L2Accounts()
	for _, agent := range bal.agents {
		agentID := agent.AgentID()
		if !contains(accs, agentID) {
			accs = append(accs, agentID)
		}
	}
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].String() < accs[j].String()
	})
	txt := "ACCOUNTS:"
	for _, acc := range accs {
		l2 := ctx.Chain.L2Assets(acc)
		l1 := ctx.Chain.Env.L1Assets(acc.Address())
		txt += fmt.Sprintf("\n%s\n\tL2: %10d", acc.String(), l2.Iotas)
		if acc.Hname() == 0 {
			txt += fmt.Sprintf(",\tL1: %10d", l1.Iotas)
		}
		for _, token := range l2.Tokens {
			txt += fmt.Sprintf("\n\tL2: %10d", token.Amount)
			tokTxt := ",\t           "
			if acc.Hname() == 0 {
				for i := range l1.Tokens {
					if *l1.Tokens[i] == *token {
						l1.Tokens = append(l1.Tokens[:i], l1.Tokens[i+1:]...)
						tokTxt = fmt.Sprintf(",\tL1: %10d", l1.Iotas)
						break
					}
				}
			}
			txt += fmt.Sprintf("%s,\t%s", tokTxt, token.ID.String())
		}
		for _, token := range l1.Tokens {
			txt += fmt.Sprintf("\n\tL2: %10d,\tL1: %10d,\t%s", 0, l1.Iotas, token.ID.String())
		}
	}
	receipt := ctx.Chain.LastReceipt()

	fmt.Printf("%s\nGas: %d, fee %d (from last receipt)\n", txt, receipt.GasBurned, receipt.GasFeeCharged)
}

func (bal *SoloBalances) Add(agent *SoloAgent, balance uint64) {
	bal.accounts[agent.AgentID().String()] += balance
}

func (bal *SoloBalances) VerifyBalances(t *testing.T) {
	bal.dumpBalances()
	ctx := bal.ctx
	require.EqualValues(t, bal.Account, ctx.Balance(ctx.Account()))
	require.EqualValues(t, bal.Chain, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, bal.Originator, ctx.Balance(ctx.Originator()))
	for _, agent := range bal.agents {
		require.EqualValues(t, bal.accounts[agent.AgentID().String()], ctx.Balance(agent))
	}
}
