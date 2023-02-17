package sbtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcoreimpl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

const (
	debug         = false
	forceGoNoWasm = false
	forceSkipWasm = false
)

const (
	ScName           = "testcore"
	HScName          = isc.Hname(0x370d33ad)
	WasmFileTestcore = "sbtestsc/testcore_bg.wasm"
)

func init() {
	if isc.Hn(ScName) != HScName {
		panic("isc.Hn(ScName) != HScName")
	}
}

func setupChain(t *testing.T, keyPairOriginator *cryptolib.KeyPair) (*solo.Solo, *solo.Chain) {
	env := solo.New(t, &solo.InitOptions{
		Debug:                    debug,
		AutoAdjustStorageDeposit: true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain, _, _ := env.NewChainExt(keyPairOriginator, 10_000, "chain1")
	err := chain.SendFromL1ToL2AccountBaseTokens(1000, utxodb.FundsFromFaucetAmount/2, chain.OriginatorAgentID, chain.OriginatorPrivateKey)
	require.NoError(t, err)
	return env, chain
}

func setupDeployer(t *testing.T, ch *solo.Chain) (*cryptolib.KeyPair, isc.AgentID) {
	user, userAddr := ch.Env.NewKeyPairWithFunds()
	ch.Env.AssertL1BaseTokens(userAddr, utxodb.FundsFromFaucetAmount)

	err := ch.DepositBaseTokensToL2(10_000, user)
	require.NoError(t, err)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, isc.NewAgentID(userAddr)).WithGasBudget(100_000)
	_, err = ch.PostRequestSync(req.AddBaseTokens(1), nil)
	require.NoError(t, err)
	return user, isc.NewAgentID(userAddr)
}

func run2(t *testing.T, test func(*testing.T, bool), skipWasm ...bool) {
	t.Run(fmt.Sprintf("run CORE version of %s", t.Name()), func(t *testing.T) {
		test(t, false)
	})
	if forceSkipWasm || (len(skipWasm) > 0 && skipWasm[0]) {
		t.Logf("skipped Wasm version of '%s'", t.Name())
		return
	}
	t.Run(fmt.Sprintf("run Wasm version of %s", t.Name()), func(t *testing.T) {
		test(t, true)
	})
}

func deployContract(chain *solo.Chain, user *cryptolib.KeyPair, runWasm bool) error {
	if forceSkipWasm || !runWasm {
		// run core version of testcore
		return chain.DeployContract(user, ScName, sbtestsc.Contract.ProgramHash)
	}

	// enable this code to be able to debug using Go version of Wasm testcore SC
	if forceGoNoWasm {
		// run non-Wasm go version of testcore
		wasmhost.GoWasmVM = func() wasmhost.WasmVM {
			return wasmhost.NewWasmGoVM(ScName, testcoreimpl.OnDispatch)
		}
		hProg, err := chain.UploadWasm(user, []byte("go:"+ScName))
		if err != nil {
			return err
		}
		err = chain.DeployContract(user, ScName, hProg)
		wasmhost.GoWasmVM = nil
		return err
	}

	// run Rust Wasm version of testcore
	return chain.DeployWasmContract(user, ScName, WasmFileTestcore)
}

// WARNING: setupTestSandboxSC will fail if AutoAdjustStorageDeposit is not enabled
func setupTestSandboxSC(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair, runWasm bool) isc.AgentID {
	err := deployContract(chain, user, runWasm)
	require.NoError(t, err)

	deployed := isc.NewContractAgentID(chain.ChainID, HScName)
	req := solo.NewCallParams(ScName, sbtestsc.FuncDoNothing.Name).
		WithGasBudget(100_000)
	_, err = chain.PostRequestSync(req, user)
	require.NoError(t, err)
	t.Logf("deployed test_sandbox'%s': %s", ScName, HScName)
	return deployed
}

func TestSetup1(t *testing.T) { run2(t, testSetup1) }
func testSetup1(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)
}

func TestSetup2(t *testing.T) { run2(t, testSetup2) }
func testSetup2(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	user, _ := setupDeployer(t, chain)
	setupTestSandboxSC(t, chain, user, w)
}
