package sbtests

import (
	"fmt"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/stretchr/testify/require"
)

const (
	DEBUG           = false
	FORCE_SKIP_WASM = false
	FORCE_RUST_WASM = true
)

const (
	ScName           = "testcore"
	HScName          = iscp.Hname(0x370d33ad)
	WasmFileTestcore = "sbtestsc/testcore_bg.wasm"
	// WasmFileTestcore = "../../../../../contracts/wasm/testcore/go/pkg/testcore_go.wasm"
	// WasmFileTestcore = "../../../../../contracts/wasm/testcore/ts/pkg/testcore_ts.wasm"
)

func init() {
	if iscp.Hn(ScName) != HScName {
		panic("iscp.Hn(ScName) != HScName")
	}
}

func setupChain(t *testing.T, keyPairOriginator *cryptolib.KeyPair) (*solo.Solo, *solo.Chain) {
	// corecontracts.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{
		Debug:                 DEBUG,
		AutoAdjustDustDeposit: true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain, _, _ := env.NewChainExt(keyPairOriginator, 10_000, "ch1")
	err := chain.SendFromL1ToL2AccountIotas(1000, utxodb.FundsFromFaucetAmount/2, chain.OriginatorAgentID, chain.OriginatorPrivateKey)
	require.NoError(t, err)
	return env, chain
}

func setupDeployer(t *testing.T, ch *solo.Chain) (*cryptolib.KeyPair, iotago.Address, *iscp.AgentID) {
	user, userAddr := ch.Env.NewKeyPairWithFunds()
	ch.Env.AssertL1Iotas(userAddr, utxodb.FundsFromFaucetAmount)

	err := ch.DepositIotasToL2(10_000, user)
	require.NoError(t, err)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, iscp.NewAgentID(userAddr, 0)).WithGasBudget(100_000)
	_, err = ch.PostRequestSync(req.AddIotas(1), nil)
	require.NoError(t, err)
	return user, userAddr, iscp.NewAgentID(userAddr, 0)
}

func run2(t *testing.T, test func(*testing.T, bool), skipWasm ...bool) {
	t.Run(fmt.Sprintf("run CORE version of %s", t.Name()), func(t *testing.T) {
		test(t, false)
	})
	if FORCE_SKIP_WASM || (len(skipWasm) > 0 && skipWasm[0]) {
		t.Logf("skipped Wasm version of '%s'", t.Name())
		return
	}
	t.Run(fmt.Sprintf("run Wasm version of %s", t.Name()), func(t *testing.T) {
		test(t, true)
	})
}

func deployContract(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair, runWasm bool) error {
	if FORCE_SKIP_WASM || !runWasm {
		// run core version of testcore
		return chain.DeployContract(user, ScName, sbtestsc.Contract.ProgramHash)
	}

	if FORCE_RUST_WASM {
		// run Rust Wasm version of testcore
		return chain.DeployWasmContract(user, ScName, WasmFileTestcore)
	}

	// run non-Wasm go version of testcore
	wasmhost.GoWasmVM = func() wasmhost.WasmVM {
		return wasmhost.NewWasmGoVM(ScName, testcore.OnLoad)
	}
	hProg, err := chain.UploadWasm(user, []byte("go:"+ScName))
	if err != nil {
		return err
	}
	err = chain.DeployContract(user, ScName, hProg)
	wasmhost.GoWasmVM = nil
	return err
}

// WARNING: setupTestSandboxSC will fail if AutoAdjustDustDeposit is not enabled
func setupTestSandboxSC(t *testing.T, chain *solo.Chain, user *cryptolib.KeyPair, runWasm bool) *iscp.AgentID {
	err := deployContract(t, chain, user, runWasm)
	require.NoError(t, err)

	deployed := iscp.NewAgentID(chain.ChainID.AsAddress(), HScName)
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
	user, _, _ := setupDeployer(t, chain)
	setupTestSandboxSC(t, chain, user, w)
}
