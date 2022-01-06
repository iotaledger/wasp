package sbtests

import (
	"fmt"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

const (
	DEBUG            = false
	ScName           = "testcore"
	HScName          = iscp.Hname(0x370d33ad)
	WasmFileTestcore = "sbtestsc/testcore_bg.wasm"
)

func setupChain(t *testing.T, keyPairOriginator *cryptolib.KeyPair) (*solo.Solo, *solo.Chain) {
	core.PrintWellKnownHnames()
	env := solo.New(t, DEBUG, false).WithNativeContract(sbtestsc.Processor)
	chain := env.NewChain(keyPairOriginator, "ch1")
	return env, chain
}

func setupDeployer(t *testing.T, chain *solo.Chain) (*cryptolib.KeyPair, iotago.Address, *iscp.AgentID) {
	user, userAddr := chain.Env.NewKeyPairWithFunds()
	chain.Env.AssertAddressIotas(userAddr, solo.Saldo)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, iscp.NewAgentID(userAddr, 0))
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)
	return user, userAddr, iscp.NewAgentID(userAddr, 0)
}

func run2(t *testing.T, test func(*testing.T, bool), skipWasm ...bool) {
	t.Run(fmt.Sprintf("run CORE version of %s", t.Name()), func(t *testing.T) {
		test(t, false)
	})
	if len(skipWasm) == 0 || !skipWasm[0] {
		t.Run(fmt.Sprintf("run Wasm version of %s", t.Name()), func(t *testing.T) {
			test(t, true)
		})
	} else {
		t.Logf("skipped Wasm version of '%s'", t.Name())
	}
}

func setupTestSandboxSC(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair, runWasm bool) (*iscp.AgentID, uint64) {
	var err error
	var extraToken uint64
	if runWasm {
		err = chain.DeployWasmContract(user, ScName, WasmFileTestcore)
		extraToken = 1
	} else {
		err = chain.DeployContract(user, ScName, sbtestsc.Contract.ProgramHash)
		extraToken = 0
	}
	require.NoError(t, err)

	deployed := iscp.NewAgentID(chain.ChainID.AsAddress(), HScName)
	req := solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name)
	_, err = chain.PostRequestSync(req.AddAssetsIotas(1), user)
	require.NoError(t, err)
	t.Logf("deployed test_sandbox'%s': %s", ScName, HScName)
	return deployed, extraToken
}

func TestSetup1(t *testing.T) { run2(t, testSetup1) }
func testSetup1(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)
}

func TestSetup2(t *testing.T) { run2(t, testSetup2) }
func testSetup2(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	user, _, _ := setupDeployer(t, chain)
	setupTestSandboxSC(t, chain, user, w)
}
