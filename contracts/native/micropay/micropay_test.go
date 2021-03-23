package micropay

import (
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

	payer, payerAddr := env.NewKeyPairWithFunds()
	pubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, pubKey,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)
}

func TestOpenChannelFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncAddWarrant, ParamServiceAddress, providerAddr).WithIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.Error(t, err)

	cAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("micropay"))
	chain.AssertIotas(cAgentID, 0)
	env.AssertAddressIotas(payerAddr, solo.Saldo)
}

func TestOpenChannelOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant, ParamServiceAddress, providerAddr).WithIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	cAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("micropay"))
	chain.AssertIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)
}

func TestOpenChannelTwice(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	cAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("micropay"))
	chain.AssertIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertIotas(cAgentID, 600+600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-600-1)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeUint64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600+600, int(warrant))

	_, exists, err = codec.DecodeUint64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	_, exists, err = codec.DecodeUint64(ret.MustGet(ParamLastOrd))
	require.NoError(t, err)
	require.False(t, exists)
}

func TestRevokeWarrant(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Interface.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	cAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("micropay"))

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err := codec.DecodeUint64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 600, warrant)

	_, exists, err = codec.DecodeUint64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)
	require.False(t, exists)

	req = solo.NewCallParams("micropay", FuncRevokeWarrant,
		ParamServiceAddress, providerAddr,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	env.AdvanceClockBy(30 * time.Minute)

	ret, err = chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, exists, err = codec.DecodeUint64(ret.MustGet(ParamWarrant))
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

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey,
		ParamPublicKey, payerPubKey,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	provider, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	cAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("micropay"))

	req = solo.NewCallParams("micropay", FuncAddWarrant,
		ParamServiceAddress, providerAddr).
		WithIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	res, err := chain.CallView("micropay", FuncGetChannelInfo,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	var ok bool
	var w uint64
	require.NoError(t, err)
	w, ok, err = codec.DecodeUint64(res.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 600, w)

	_, ok, err = codec.DecodeUint64(res.MustGet(ParamRevoked))
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
	req = solo.NewCallParamsFromDic("micropay", FuncSettle, par).WithIotas(1)
	_, err = chain.PostRequestSync(req, provider)
	require.NoError(t, err)

	env.AssertAddressIotas(providerAddr, solo.Saldo+42+41-1)

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
