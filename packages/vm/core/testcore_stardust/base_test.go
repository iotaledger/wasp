package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	genesisAddr := env.L1Ledger().GenesisAddress()
	assets := env.L1AddressBalances(genesisAddr)
	require.EqualValues(t, env.L1Ledger().Supply(), assets.Iotas)

	chain := env.NewChain(nil, "chain1")
	defer chain.Log.Sync()
	env.WaitPublisher()
	chain.AssertControlAddresses()
	chain.AssertTotalIotas(212)
	t.Logf("originator address iotas: %d (spent %d)",
		env.L1IotaBalance(chain.OriginatorAddress), solo.Saldo-env.L1IotaBalance(chain.OriginatorAddress))

	nativeTokenIDs := chain.GetOnChainTokenIDs()
	require.EqualValues(t, 0, len(nativeTokenIDs))

	totalDustDeposit := chain.GetTotalOnChainDustDeposit()
	totalIotasOnChain := chain.L2TotalIotas()
	totalSpent := totalDustDeposit + totalIotasOnChain
	t.Logf("total on chain: dust deposit: %d, total iotas: %d, total sent: %d",
		totalDustDeposit, totalIotasOnChain, totalSpent)
	env.AssertL1AddressIotas(chain.OriginatorAddress, solo.Saldo-totalSpent)
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
