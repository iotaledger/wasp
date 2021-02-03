package test

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/examples/micropay"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasics(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)
}

func TestOpenChannel(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer := env.NewSignatureSchemeWithFunds()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, 1337)

	req := solo.NewCallParams("micropay", micropay.FuncOpenChannel,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-1)
}

func TestOpenChannelTwice(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer := env.NewSignatureSchemeWithFunds()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, 1337)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)

	req := solo.NewCallParams("micropay", micropay.FuncOpenChannel,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-1)

	_, err = chain.PostRequest(req, payer)
	require.Error(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-2)
}
