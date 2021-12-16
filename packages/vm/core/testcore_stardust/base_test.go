package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestLedgerConsistency(t *testing.T) {
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

	dustInfo := chain.GetTotalOnChainDustDeposit()
	totalIotasOnChain := chain.L2TotalIotas()
	totalSpent := dustInfo.Total() + totalIotasOnChain
	t.Logf("total on chain: dust deposit: %d, total iotas: %d, total sent: %d",
		dustInfo, totalIotasOnChain, totalSpent)
	env.AssertL1AddressIotas(chain.OriginatorAddress, solo.Saldo-totalSpent)

	outs, ids := env.L1Ledger().GetAliasOutputs(chain.ChainID.AsAddress())
	require.EqualValues(t, 1, len(outs))
	require.EqualValues(t, 1, len(ids))

	totalAssets := chain.L2TotalAssets()
	require.EqualValues(t, 0, len(totalAssets.Tokens))
	require.EqualValues(t, int(totalAssets.Iotas), int(outs[0].Amount+dustInfo.Total()))

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
