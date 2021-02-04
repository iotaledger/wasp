package test

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/examples/micropay"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBasics(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)
}

func TestSubmitPk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	req := solo.NewCallParams("micropay", micropay.FuncPublicKey,
		micropay.ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)
}

func TestOpenChannelFail(t *testing.T) {
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

	req := solo.NewCallParams("micropay", micropay.FuncAddWarrant,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.Error(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 0)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-1)
}

func TestOpenChannelOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	req := solo.NewCallParams("micropay", micropay.FuncPublicKey,
		micropay.ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, 1337)

	req = solo.NewCallParams("micropay", micropay.FuncAddWarrant,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-2)
}

func TestOpenChannelTwice(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	req := solo.NewCallParams("micropay", micropay.FuncPublicKey,
		micropay.ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, 1337)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)

	req = solo.NewCallParams("micropay", micropay.FuncAddWarrant,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-2)

	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600+600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-600-3)

	ret, err := chain.CallView("micropay", micropay.FuncGetChannelInfo,
		micropay.ParamPayerAddress, payerAddr,
		micropay.ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeInt64(ret.MustGet(micropay.ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600+600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestRevokeWarrant(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", micropay.Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337)

	req := solo.NewCallParams("micropay", micropay.FuncPublicKey,
		micropay.ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, 1337)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)

	req = solo.NewCallParams("micropay", micropay.FuncAddWarrant,
		micropay.ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, 1337-600-2)

	ret, err := chain.CallView("micropay", micropay.FuncGetChannelInfo,
		micropay.ParamPayerAddress, payerAddr,
		micropay.ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeInt64(ret.MustGet(micropay.ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	req = solo.NewCallParams("micropay", micropay.FuncRevokeWarrant,
		micropay.ParamServiceAddress, providerAddr,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	env.AdvanceClockBy(30 * time.Minute)

	ret, err = chain.CallView("micropay", micropay.FuncGetChannelInfo,
		micropay.ParamPayerAddress, payerAddr,
		micropay.ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamRevoked))
	require.NoError(t, err)
	require.True(t, exists)

	env.AdvanceClockBy(31 * time.Minute)
	chain.WaitForEmptyBacklog()

	ret, err = chain.CallView("micropay", micropay.FuncGetChannelInfo,
		micropay.ParamPayerAddress, payerAddr,
		micropay.ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	_, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamWarrant))
	require.NoError(t, err)
	require.False(t, exists)

	_, exists, err = codec.DecodeInt64(ret.MustGet(micropay.ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

}
