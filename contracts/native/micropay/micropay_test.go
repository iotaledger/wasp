package micropay

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestBasics(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)
}

func TestSubmitPk(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	pubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey.Name,
		ParamPublicKey, pubKey,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)
}

func TestOpenChannelFail(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncAddWarrant.Name, ParamServiceAddress, providerAddr).AddAssetsIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.Error(t, err)

	cAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("micropay"))
	chain.AssertL2AccountIotas(cAgentID, 0)
	env.AssertAddressIotas(payerAddr, solo.Saldo)
}

func TestOpenChannelOk(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey.Name,
		ParamPublicKey, payerPubKey,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant.Name, ParamServiceAddress, providerAddr).AddAssetsIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	cAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("micropay"))
	chain.AssertL2AccountIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)
}

func TestOpenChannelTwice(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey.Name,
		ParamPublicKey, payerPubKey,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	req = solo.NewCallParams("micropay", FuncAddWarrant.Name,
		ParamServiceAddress, providerAddr).
		AddAssetsIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	cAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("micropay"))
	chain.AssertL2AccountIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertL2AccountIotas(cAgentID, 600+600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-600-1)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, err := codec.DecodeUint64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.EqualValues(t, 600+600, int(warrant))

	require.False(t, ret.MustHas(ParamRevoked))

	require.False(t, ret.MustHas(ParamLastOrd))
}

func TestRevokeWarrant(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey.Name,
		ParamPublicKey, payerPubKey,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	_, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	cAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("micropay"))

	req = solo.NewCallParams("micropay", FuncAddWarrant.Name,
		ParamServiceAddress, providerAddr).
		AddAssetsIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertL2AccountIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	ret, err := chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, err := codec.DecodeUint64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.EqualValues(t, 600, warrant)

	require.False(t, ret.MustHas(ParamRevoked))

	req = solo.NewCallParams("micropay", FuncRevokeWarrant.Name,
		ParamServiceAddress, providerAddr,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	env.AdvanceClockBy(30 * time.Minute)

	ret, err = chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, err = codec.DecodeUint64(ret.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.EqualValues(t, 600, warrant)

	_, err = codec.DecodeInt64(ret.MustGet(ParamRevoked))
	require.NoError(t, err)

	env.AdvanceClockBy(31 * time.Minute)
	require.True(t, chain.WaitForRequestsThrough(6))

	ret, err = chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	require.False(t, ret.MustHas(ParamWarrant))

	require.False(t, ret.MustHas(ParamRevoked))

	require.False(t, ret.MustHas(ParamLastOrd))
}

func TestPayment(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "micropay", Contract.ProgramHash)
	require.NoError(t, err)

	payer, payerAddr := env.NewKeyPairWithFunds()
	payerPubKey := payer.PublicKey.Bytes()
	env.AssertAddressIotas(payerAddr, solo.Saldo)

	req := solo.NewCallParams("micropay", FuncPublicKey.Name,
		ParamPublicKey, payerPubKey,
	).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	provider, providerAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(providerAddr, solo.Saldo)

	cAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("micropay"))

	req = solo.NewCallParams("micropay", FuncAddWarrant.Name,
		ParamServiceAddress, providerAddr).
		AddAssetsIotas(600)
	_, err = chain.PostRequestSync(req, payer)
	require.NoError(t, err)

	chain.AssertL2AccountIotas(cAgentID, 600+1)
	env.AssertAddressIotas(payerAddr, solo.Saldo-600-1)

	res, err := chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	w, err := codec.DecodeUint64(res.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.EqualValues(t, 600, w)

	require.False(t, res.MustHas(ParamRevoked))

	pay1 := NewPayment(uint32(time.Now().Unix()), 42, providerAddr, payer).Bytes()
	time.Sleep(1 * time.Second)
	last := uint32(time.Now().Unix())
	pay2 := NewPayment(last, 41, providerAddr, payer).Bytes()
	par := dict.New()
	par.Set(ParamPayerAddress, codec.EncodeAddress(payerAddr))
	arr := collections.NewArray16(par, ParamPayments)
	_ = arr.Push(pay1)
	_ = arr.Push(pay2)
	req = solo.NewCallParamsFromDic("micropay", FuncSettle.Name, par).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, provider)
	require.NoError(t, err)

	env.AssertAddressIotas(providerAddr, solo.Saldo+42+41-1)

	res, err = chain.CallView("micropay", FuncGetChannelInfo.Name,
		ParamPayerAddress, payerAddr,
		ParamServiceAddress, providerAddr,
	)
	require.NoError(t, err)
	warrant, err := codec.DecodeInt64(res.MustGet(ParamWarrant))
	require.NoError(t, err)
	require.EqualValues(t, 600-42-41, warrant)

	require.False(t, res.MustHas(ParamRevoked))

	lastOrd, err := codec.DecodeInt64(res.MustGet(ParamLastOrd))
	require.NoError(t, err)
	require.EqualValues(t, last, lastOrd)
}
