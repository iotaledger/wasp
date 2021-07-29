package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	defer chain.Log.Sync()
	chain.CheckControlAddresses()
	chain.AssertTotalIotas(1)
	chain.AssertCommonAccountIotas(1)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-1)
	env.WaitPublisher()
	chain.CheckControlAddresses()
}

func TestNoContractPost(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams("dummyContract", "dummyEP").WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
	chain.CheckControlAddresses()
	chain.CheckControlAddresses()
}

func TestNoContractView(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.CheckControlAddresses()

	_, err := chain.CallView("dummyContract", "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
	chain.CheckControlAddresses()
}

func TestNoEPPost(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.CheckControlAddresses()

	req := solo.NewCallParams(root.Contract.Name, "dummyEP").WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
	chain.CheckControlAddresses()
}

func TestNoEPView(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.CheckControlAddresses()

	_, err := chain.CallView(root.Contract.Name, "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
	chain.CheckControlAddresses()
}

func TestOkCall(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetChainConfig.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0)
	req.WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	chain.CheckControlAddresses()
	env.WaitPublisher()
	chain.CheckControlAddresses()
}

func TestNoTokens(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.CheckControlAddresses()

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetChainConfig.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	chain.CheckControlAddresses()
	env.WaitPublisher()
	chain.CheckControlAddresses()
}
