package sbtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	debug  = false
	ScName = "testcore"
)

var HScName = isc.Hn(ScName)

func setupChain(t *testing.T, keyPairOriginator *cryptolib.KeyPair) (*solo.Solo, *solo.Chain) {
	env := solo.New(t, &solo.InitOptions{
		Debug:             debug,
		GasBurnLogEnabled: true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain, _ := env.NewChainExt(keyPairOriginator, 10_000, "chain1")
	err := chain.SendFromL1ToL2AccountBaseTokens(1000, iotaclient.FundsFromFaucetAmount/2, chain.OriginatorAgentID, chain.OriginatorPrivateKey)
	require.NoError(t, err)
	return env, chain
}

func setupDeployer(t *testing.T, ch *solo.Chain) (*cryptolib.KeyPair, isc.AgentID) {
	user, userAddr := ch.Env.NewKeyPairWithFunds()
	ch.Env.AssertL1BaseTokens(userAddr, iotaclient.FundsFromFaucetAmount)

	err := ch.DepositBaseTokensToL2(coin.Value(10*gas.LimitsDefault.MinGasPerRequest), user)
	require.NoError(t, err)

	req := solo.NewCallParams(root.FuncGrantDeployPermission.Message(isc.NewAddressAgentID(userAddr))).
		WithGasBudget(100_000)
	_, err = ch.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	return user, isc.NewAddressAgentID(userAddr)
}

func run2(t *testing.T, test func(*testing.T)) {
	t.Run(fmt.Sprintf("run CORE version of %s", t.Name()), func(t *testing.T) {
		test(t)
	})
}

func deployContract(chain *solo.Chain, user *cryptolib.KeyPair) error {
	return chain.DeployContract(user, ScName, sbtestsc.Contract.ProgramHash)
}

func setupTestSandboxSC(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair) isc.AgentID {
	err := deployContract(chain, user)
	require.NoError(t, err)

	deployed := isc.NewContractAgentID(chain.ChainID, HScName)
	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncDoNothing.Name).
		WithGasBudget(100_000)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)
	t.Logf("deployed test_sandbox'%s': %s", ScName, HScName)
	return deployed
}

func TestSetup1(t *testing.T) { run2(t, testSetup1) }
func testSetup1(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)
}

func TestSetup2(t *testing.T) { run2(t, testSetup2) }
func testSetup2(t *testing.T) {
	_, chain := setupChain(t, nil)
	user, _ := setupDeployer(t, chain)
	setupTestSandboxSC(t, chain, user)
}
