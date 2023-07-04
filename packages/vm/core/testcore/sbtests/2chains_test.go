package sbtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

// TODO deposit fee needs to be constant, this test is using a placeholder value that will need to be changed

// test case:
// 2 chains
// SC deployed on chain 2
// funds are deposited by some user on chain 1, on behalf of SC
// SC tries to withdraw those funds from chain 1 to chain 2
func Test2Chains(t *testing.T) { run2(t, test2Chains) }

func test2Chains(t *testing.T, w bool) {
	corecontracts.PrintWellKnownHnames()

	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain1 := env.NewChain()
	chain2, _ := env.NewChainExt(nil, 0, "chain2")
	err := chain2.DepositAssetsToL2(isc.NewAssetsBaseTokens(5*isc.Million), nil)
	require.NoError(t, err)
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	_ = setupTestSandboxSC(t, chain1, nil, w)
	contractAgentID2 := setupTestSandboxSC(t, chain2, nil, w)

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAgentID(userAddress)
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	chain1TotalBaseTokens := chain1.L2TotalBaseTokens()
	chain2TotalBaseTokens := chain2.L2TotalBaseTokens()

	chain1.WaitForRequestsMark()
	chain2.WaitForRequestsMark()

	// send base tokens to contractAgentID2 (that is an entity of chain2) on chain1
	const baseTokensToSend = 11 * isc.Million
	const baseTokensCreditedToScOnChain1 = 10 * isc.Million
	_, err = chain1.PostRequestSync(solo.NewCallParams(
		accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		accounts.ParamAgentID, contractAgentID2,
	).
		AddBaseTokens(baseTokensToSend).
		AddAllowanceBaseTokens(baseTokensCreditedToScOnChain1).
		WithGasBudget(math.MaxUint64),
		userWallet)
	require.NoError(t, err)

	chain1TransferAllowanceReceipt := chain1.LastReceipt()
	chain1TransferAllowanceGas := chain1TransferAllowanceReceipt.GasFeeCharged

	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-baseTokensToSend)
	chain1.AssertL2BaseTokens(userAgentID, baseTokensToSend-baseTokensCreditedToScOnChain1-chain1TransferAllowanceGas)
	chain1.AssertL2BaseTokens(contractAgentID2, baseTokensCreditedToScOnChain1)
	chain1.AssertL2TotalBaseTokens(chain1TotalBaseTokens + baseTokensToSend)
	chain1TotalBaseTokens += baseTokensToSend

	chain2.AssertL2BaseTokens(userAgentID, 0)
	chain2.AssertL2BaseTokens(contractAgentID2, 0)
	chain2.AssertL2TotalBaseTokens(chain2TotalBaseTokens)

	fmt.Println("---------------chain1---------------")
	fmt.Println(chain1.DumpAccounts())
	fmt.Println("---------------chain2---------------")
	fmt.Println(chain2.DumpAccounts())
	fmt.Println("------------------------------------")

	// make chain2 send a call to chain1 to withdraw base tokens
	baseTokensToWithdrawFromChain1 := baseTokensCreditedToScOnChain1

	// actual gas fee is less, but always rounded up to this minimum amount
	const gasFee = wasmlib.MinGasFee
	const storageDeposit = wasmlib.StorageDeposit

	// NOTE: make sure you READ THE DOCS for accounts.transferAccountToChain()
	// to understand fully how to call it and why.

	// reqAllowance is the allowance provided to chain2.testcore.withdrawFromChain(),
	// which needs to be enough to cover any storage deposit along the way and to pay
	// the gas fees for the chain2.accounts.transferAccountToChain() request and the
	// chain1.accounts.transferAllowanceTo() request.
	// note that the storage deposit will be returned in the end
	reqAllowance := storageDeposit + gasFee + gasFee

	// also cover gas fee for `FuncWithdrawFromChain` on chain2
	assetsBaseTokens := reqAllowance + isc.Million

	_, err = chain2.PostRequestSync(solo.NewCallParams(ScName, sbtestsc.FuncWithdrawFromChain.Name,
		sbtestsc.ParamChainID, chain1.ChainID,
		sbtestsc.ParamBaseTokens, baseTokensToWithdrawFromChain1).
		AddBaseTokens(assetsBaseTokens).
		WithAllowance(isc.NewAssetsBaseTokens(reqAllowance)).
		WithGasBudget(isc.Million),
		userWallet)
	require.NoError(t, err)
	chain2WithdrawFromChainReceipt := chain2.LastReceipt()
	chain2WithdrawFromChainGas := chain2WithdrawFromChainReceipt.GasFeeCharged

	require.True(t, chain1.WaitForRequestsThrough(2, 10*time.Second))
	require.True(t, chain2.WaitForRequestsThrough(2, 10*time.Second))

	chain2TransferAllowanceReceipt := chain2.LastReceipt()
	chain2TransferAllowanceGas := chain2TransferAllowanceReceipt.GasFeeCharged
	chain2TransferAllowanceTarget := chain2TransferAllowanceReceipt.DeserializedRequest().CallTarget()
	require.Equal(t, chain2TransferAllowanceTarget.Contract, accounts.Contract.Hname())
	require.Equal(t, chain2TransferAllowanceTarget.EntryPoint, accounts.FuncTransferAllowanceTo.Hname())
	require.Nil(t, chain2TransferAllowanceReceipt.Error)

	chain1TransferAccountToChainReceipt := chain1.LastReceipt()
	chain1TransferAccountToChainGas := chain1TransferAccountToChainReceipt.GasFeeCharged
	chain1TransferAccountToChainTarget := chain1TransferAccountToChainReceipt.DeserializedRequest().CallTarget()
	require.Equal(t, chain1TransferAccountToChainTarget.Contract, accounts.Contract.Hname())
	require.Equal(t, chain1TransferAccountToChainTarget.EntryPoint, accounts.FuncTransferAccountToChain.Hname())
	require.Nil(t, chain1TransferAccountToChainReceipt.Error)

	fmt.Println("---------------chain1---------------")
	fmt.Println(chain1.DumpAccounts())
	fmt.Println("---------------chain2---------------")
	fmt.Println(chain2.DumpAccounts())
	fmt.Println("------------------------------------")

	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-baseTokensToSend-assetsBaseTokens)

	chain1.AssertL2BaseTokens(userAgentID, baseTokensToSend-baseTokensCreditedToScOnChain1-chain1TransferAllowanceGas)
	chain1.AssertL2BaseTokens(contractAgentID2, 0) // emptied the account
	chain1.AssertL2TotalBaseTokens(chain1TotalBaseTokens + chain1TransferAllowanceGas + chain1TransferAccountToChainGas - chain2TransferAllowanceGas - baseTokensToWithdrawFromChain1)

	chain2.AssertL2BaseTokens(userAgentID, assetsBaseTokens-reqAllowance-chain2WithdrawFromChainGas)
	chain2.AssertL2BaseTokens(contractAgentID2, baseTokensToWithdrawFromChain1+storageDeposit)
	chain2.AssertL2TotalBaseTokens(chain2TotalBaseTokens + assetsBaseTokens + baseTokensCreditedToScOnChain1 + chain2TransferAllowanceGas - chain1TransferAllowanceGas - chain1TransferAccountToChainGas)
}
