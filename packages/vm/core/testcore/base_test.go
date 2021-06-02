package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test1(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.AssertTotalIotas(1)
	chain.AssertOwnersIotas(1)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-1)
	env.WaitPublisher()
}

func TestNoContractPost(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams("dummyContract", "dummyEP").WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
}

func TestNoContractView(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	_, err := chain.CallView("dummyContract", "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
}

func TestNoEPPost(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, "dummyEP").WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
}

func TestNoEPView(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	_, err := chain.CallView(root.Interface.Name, "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
}

func TestOkCall(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0)
	req.WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
}

func TestNoTokens(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	env.WaitPublisher()
}
