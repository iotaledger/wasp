package test

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

//TODO update comments to reflect new way of token handling

func TestTutorial1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex1")

	chainID, chainOwnerID, coreContracts := chain.GetInfo()                      // calls view root::GetChainInfo
	require.EqualValues(t, len(core.AllCoreContractsByHash), len(coreContracts)) // all core contracts deployed by default

	t.Logf("chain ID: %s", chainID.String())
	t.Logf("chain owner ID: %s", chainOwnerID.String())
	for hname, rec := range coreContracts {
		t.Logf("    Core contract '%s': %s", rec.Name, coretypes.NewAgentID(chainID.AsAddress(), hname))
	}
}

func TestTutorial2(t *testing.T) {
	env := solo.New(t, false, false)
	_, userAddress := env.NewKeyPairWithFunds() // create new wallet with 1337 iotas
	t.Logf("Address of the userWallet is: %s", userAddress)
	numIotas := env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA) // how many iotas the address contains
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337) // assert the address has 1337 iotas
}

func TestTutorial3(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex3")
	// deploy the contract on chain
	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	// call contract to store string
	req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!")
	req.WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// call the contract to extract value of the 'paramString' and check
	res, err := chain.CallView("example1", "getString")
	require.NoError(t, err)
	returnedString, exists, err := codec.DecodeString(res.MustGet("paramString"))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, "Hello, world!", returnedString)
}

func TestTutorial4(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex4")
	// deploy the contract on chain
	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	// call contract incorrectly (omit 'paramString')
	req := solo.NewCallParams("example1", "storeString")
	req.WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestTutorial5(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex5")

	// create a wallet with 1337 iotas.
	// the wallet has address and it is globally identified
	// through a universal identifier: the agent ID
	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)

	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337) // 1337 on address
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)  // empty on-chain

	t.Logf("Address of the userWallet is: %s", userAddress)
	numIotas := env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337)

	// send 42 iotas from wallet to own account on-chain, controlled by the same wallet
	req := solo.NewCallParams(accounts.Name, accounts.FuncDeposit)
	req.WithIotas(42)
	_, err := chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// check address balance: must be 42 iotas less
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-42)
	// check the on-chain account. Must contain 42 iotas
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 42)

	// withdraw all iotas
	req = solo.NewCallParams(accounts.Name, accounts.FuncWithdraw)
	req.WithIotas(1)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// we are back to initial situation: IOTA is fee-less!
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-1)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0) // empty
}

func TestTutorial6(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex6")

	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("example1"))

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)

	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337)
	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 0) // empty on-chain
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)     // empty on-chain

	req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!")
	req.WithIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 42)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-42)
}

func TestTutorial7(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex7")

	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("example1"))

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)

	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337)
	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 0) // empty on-chain
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)     // empty on-chain

	// missing parameter, will panic
	req := solo.NewCallParams("example1", "storeString")
	req.WithIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.Error(t, err)

	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 0)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337)
}

// test withdrawIota method
func TestTutorial8(t *testing.T) {
	// create solo environment
	env := solo.New(t, false, false)
	// deploy new chain
	chain := env.NewChain(nil, "ex8")

	// create a user's wallet (private key) and request 1337 iotas from the faucet.
	// It corresponds to L1 address
	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := coretypes.NewAgentID(userAddress, 0)
	t.Logf("userAgentID: %s", userAgentID)

	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0) // empty on-chain

	// the chain owner (default) send a request to the root contract to grant right to deploy
	// contract on the chain to the use
	req := solo.NewCallParams(root.Interface.Name, root.FuncGrantDeployPermission, root.ParamDeployer, userAgentID)
	req.WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// user deploys wasm smart contract on the chain under the name "example1"
	// the wasm binary is in the file
	err = chain.DeployWasmContract(userWallet, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	contractAgentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), coretypes.Hn("example1"))

	// the deployment of the smart contract required 1 requests to the root contract:
	// - to submit binary to the on-chain "blob" registry
	// - to deploy contract from the blob
	// Two tokens were taken from the user account to form requests and then were
	// deposited to the user's account on the chain
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-2)
	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 0) // empty on-chain
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)

	// user send a "storeString" request to the smart contract. It attaches 42 iotas to the request
	// It also takes 1 iota for the request token
	// Result is 42 iotas moved to the smart contract's account
	req = solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!").
		WithIotas(42)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 42)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-44)

	// user withdraws all iotas from the smart contract back
	// Out of 42 iotas 41 iota is coming back to the user's address, 1 iotas
	// is accrued to the user on chain
	req = solo.NewCallParams("example1", "withdrawIota")
	req.WithIotas(1)
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	chain.AssertAccountBalance(contractAgentID, ledgerstate.ColorIOTA, 0)
	chain.AssertAccountBalance(userAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 1337-44+42)
}
