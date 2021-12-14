package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	defer chain.Log.Sync()
	chain.AssertControlAddresses()
	chain.AssertTotalIotas(212)
	chain.AssertCommonAccountIotas(0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-1)
	env.WaitPublisher()
	chain.AssertControlAddresses()
}

func TestNoContractPost(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams("dummyContract", "dummyEP")
	_, err := chain.PostRequestSync(req.WithIotas(2), nil)
	require.NoError(t, err)
	env.WaitPublisher()
	chain.AssertControlAddresses()
	chain.AssertControlAddresses()
}

func TestNoContractView(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.AssertControlAddresses()

	_, err := chain.CallView("dummyContract", "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
	chain.AssertControlAddresses()
}

func TestNoEPPost(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.AssertControlAddresses()

	req := solo.NewCallParams(root.Contract.Name, "dummyEP")
	_, err := chain.PostRequestSync(req.WithIotas(2), nil)
	require.NoError(t, err)
	env.WaitPublisher()
	chain.AssertControlAddresses()
}

func TestNoEPView(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.AssertControlAddresses()

	_, err := chain.CallView(root.Contract.Name, "dummyEP")
	require.Error(t, err)
	env.WaitPublisher()
	chain.AssertControlAddresses()
}

func TestOkCall(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name)
	req.WithIotas(2)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	chain.AssertControlAddresses()
	env.WaitPublisher()
	chain.AssertControlAddresses()
}

func TestNoTokens(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	chain.AssertControlAddresses()

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	chain.AssertControlAddresses()
	env.WaitPublisher()
	chain.AssertControlAddresses()
}
