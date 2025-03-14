package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var (
	ScName  = sbtestsc.Contract.Name
	HScName = sbtestsc.Contract.Hname()
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func setupChain(t *testing.T, keyPairOriginator *cryptolib.KeyPair) (*solo.Solo, *solo.Chain) {
	env := solo.New(t, &solo.InitOptions{
		Debug:             false,
		GasBurnLogEnabled: true,
	})
	chain, _ := env.NewChainExt(keyPairOriginator, 10_000, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	err := chain.SendFromL1ToL2AccountBaseTokens(1000, iotaclient.FundsFromFaucetAmount/10, chain.OwnerAgentID(), chain.OwnerPrivateKey)
	require.NoError(t, err)
	return env, chain
}

func setupDeployer(t *testing.T, ch *solo.Chain) (*cryptolib.KeyPair, isc.AgentID) {
	user, userAddr := ch.Env.NewKeyPairWithFunds()
	ch.Env.AssertL1BaseTokens(userAddr, iotaclient.FundsFromFaucetAmount)

	err := ch.DepositBaseTokensToL2(coin.Value(10*gas.LimitsDefault.MinGasPerRequest), user)
	require.NoError(t, err)

	return user, isc.NewAddressAgentID(userAddr)
}

func setupTestSandboxSC(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair) isc.AgentID {
	deployed := isc.NewContractAgentID(chain.ChainID, HScName)
	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncDoNothing.Name).
		WithGasBudget(100_000)
	_, err := chain.PostRequestSync(req, user)
	require.NoError(t, err)
	t.Logf("deployed test_sandbox'%s': %s", ScName, HScName)
	return deployed
}

func TestSetup1(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)
}

func TestSetup2(t *testing.T) {
	_, chain := setupChain(t, nil)
	user, _ := setupDeployer(t, chain)
	setupTestSandboxSC(t, chain, user)
}
