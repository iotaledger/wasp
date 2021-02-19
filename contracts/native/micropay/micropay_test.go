package micropay

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBasics(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)
}

func TestSubmitPk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)
}

func TestOpenChannelFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer := env.NewSignatureSchemeWithFunds()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.Error(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 0)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-1)
}

func TestOpenChannelOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-600-2)
}

func TestOpenChannelTwice(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-600-2)

	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600+600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-600-600-3)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeInt64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600+600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamLastOrd))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestRevokeWarrant(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-600-2)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeInt64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	req = solo.NewCallParams("micropay", FuncRevokeWarrant,
		ParamServiceAddress, providerAddr,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	env.AdvanceClockBy(30 * time.Minute)

	ret, err = chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err = codec.DecodeInt64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600, warrant)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.True(t, exists)

	env.AdvanceClockBy(31 * time.Minute)
	chain.WaitForEmptyBacklog()

	ret, err = chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.False(t, exists)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	_, exists, err = codec.DecodeInt64(ret.MustGet(ParamLastOrd))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestPayment(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerPubKey := env.NewSignatureSchemeWithFundsAndPubKey()
	payerAddr := payer.Address()
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	provider := env.NewSignatureSchemeWithFunds()
	providerAddr := provider.Address()
	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithTransfer(balance.ColorIOTA, 600)
	_, err = chain.PostRequest(req, payer)
	require.NoError(t, err)

	cID := coretypes.NewContractID(chain.ChainID, coretypes.Hn("micropay"))
	cAgentID := coretypes.NewAgentIDFromContractID(cID)
	chain.AssertAccountBalance(cAgentID, balance.ColorIOTA, 600)
	env.AssertAddressBalance(payerAddr, balance.ColorIOTA, solo.Saldo-600-2)

	res, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	var ok bool
	var w int64
	require.NoError(t, err)
	w, ok, err = codec.DecodeInt64(res.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 600, w)

	_, ok, err = codec.DecodeInt64(res.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, ok)

	pay1 := NewPayment(uint32(time.Now().Unix()), 42, providerAddr, payer).Bytes()
	time.Sleep(1 * time.Second)
	last := uint32(time.Now().Unix())
	pay2 := NewPayment(last, 41, providerAddr, payer).Bytes()
	par := dict.New()
	par.Set(ParamPayerAddress, codec.EncodeAddress(payerAddr))
	arr := collections.NewArray(par, ParamPayments)
	_ = arr.Push(pay1)
	_ = arr.Push(pay2)
	req = solo.NewCallParamsFromDic("micropay", FuncSettle, par)
	_, err = chain.PostRequest(req, provider)
	require.NoError(t, err)

	env.AssertAddressBalance(providerAddr, balance.ColorIOTA, solo.Saldo+42+41-1)

	res, err = chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeInt64(res.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600-42-41, warrant)

	_, exists, err = codec.DecodeInt64(res.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	lastOrd, exists, err := codec.DecodeInt64(res.MustGet(ParamLastOrd))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, last, lastOrd)
}
