package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

var seed = cryptolib.NewSeedFromBytes([]byte("long long seed for determinism...............1"))

func TestTutorial1(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex1")

	// calls view root::GetChainInfo
	chainID, chainOwnerID, coreContracts := chain.GetInfo()
	// assert all core contracts deployed (default)
	require.EqualValues(t, len(corecontracts.All), len(coreContracts))

	t.Logf("chain ID: %s", chainID.String())
	t.Logf("chain owner ID: %s", chainOwnerID.String())
	for hname, rec := range coreContracts {
		t.Logf("    Core contract '%s': %s", rec.Name, iscp.NewAgentID(chainID.AsAddress(), hname))
	}
}

func TestTutorial2(t *testing.T) {
	env := solo.New(t, false, false, seed)
	_, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	t.Logf("address of the userWallet is: %s", userAddress.Base58())
	numIotas := env.L1NativeTokens(userAddress, iscp.IotaTokenID) // how many iotas the address contains
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressNativeTokenBalance(userAddress, iscp.IotaTokenID, solo.Saldo)
}

func TestTutorial3(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex3")
	// deploy the contract on chain
	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	// call contract to store string
	req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!").AddIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// call the contract to extract value of the 'paramString' and check
	res, err := chain.CallView("example1", "getString")
	require.NoError(t, err)
	par := kvdecoder.New(res, chain.Log())
	returnedString := par.MustGetString("paramString")
	require.EqualValues(t, "Hello, world!", returnedString)
}

func TestTutorial4(t *testing.T) {
	env := solo.New(t, false, false, seed)

	chain := env.NewChain(nil, "ex4")
	// deploy the contract on chain
	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	// call contract incorrectly (omit 'paramString')
	req := solo.NewCallParams("example1", "storeString").AddIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestTutorial5(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex5")
	// create a wallet with 1000000 iotas.
	// the wallet has address and it is globally identified
	// through a universal identifier: the agent ID
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(5))
	userAgentID := iscp.NewAgentID(userAddress, 0)

	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0) // empty on-chain

	t.Logf("Address of the userWallet is: %s", userAddress.Base58())
	numIotas := env.L1NativeTokens(userAddress, colored.IOTA)
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)

	// send 42 iotas from wallet to own account on-chain, controlled by the same wallet
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).AddIotas(42)
	_, err := chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// check address balance: must be 42 iotas less
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo-42)
	// check the on-chain account. Must contain 42 iotas
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 42)

	// withdraw all iotas back to the sender
	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).AddIotas(1)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// we are back to initial situation: IOTA is fee-less!
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0) // empty
}

func TestTutorial6(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex6")

	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("example1"))

	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(5))
	userAgentID := iscp.NewAgentID(userAddress, 0)

	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 0) // empty on-chain
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0)     // empty on-chain

	req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!").AddIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 42)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0)
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo-42)
}

func TestTutorial7(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex7")

	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("example1"))

	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(5))
	userAgentID := iscp.NewAgentID(userAddress, 0)

	// we start with these balances on address and on chain
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 0) // empty
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0)     // empty

	// missing parameter, request will panic
	req := solo.NewCallParams("example1", "storeString").AddIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.Error(t, err)

	// assert balances didn't change on address and on chain
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 0) // still empty
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0)     // still empty
}

// test withdrawIota method
func TestTutorial8(t *testing.T) {
	// create deterministic solo environment
	env := solo.New(t, false, false, seed)
	// deploy new chain
	chain := env.NewChain(nil, "ex8")

	// create a user's wallet (private key) with address and request iotas from the faucet.
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(5))
	userAgentID := iscp.NewAgentID(userAddress, 0)
	t.Logf("userAgentID: %s", userAgentID)

	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 0) // empty on-chain

	// the chain owner (default) send a request to the root contract to grant right to deploy
	// contract on the chain to the use
	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name, root.ParamDeployer, userAgentID).AddIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// user deploys wasm smart contract on the chain under the name "example1"
	// the wasm binary is in the file
	err = chain.DeployWasmContract(userWallet, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn("example1"))

	// the deployment of the smart contract required 1 requests to the root contract:
	// - to submit binary to the on-chain "blob" registry
	// - to deploy contract from the blob
	// Two tokens were taken from the user account to form requests and then were
	// deposited to the user's account on the chain
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo-2)
	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 0) // empty on-chain
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 1)

	// user send a "storeString" request to the smart contract. It attaches 42 iotas to the request
	// It also takes 1 iota for the request token
	// Result is 42 iotas moved to the smart contract's account
	req = solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!").
		AddIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 42)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 1)
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo-44)

	// user withdraws all iotas from the smart contract back
	// Out of 42 iotas 41 iota is coming back to the user's address, 1 iotas
	// is accrued to the user on chain
	req = solo.NewCallParams("example1", "withdrawIota")
	req.AddIotas(1)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertL2NativeTokens(contractAgentID, colored.IOTA, 0)
	chain.AssertL2NativeTokens(userAgentID, colored.IOTA, 1)
	env.AssertAddressNativeTokenBalance(userAddress, colored.IOTA, solo.Saldo-44+42)
}
