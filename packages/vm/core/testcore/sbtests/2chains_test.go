package sbtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/gas"
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
	// chain owner deposit base tokens on chain2
	chain2BaseTokenOwnerDeposit := 5 * isc.Million
	err := chain2.DepositAssetsToL2(isc.NewAssetsBaseTokens(chain2BaseTokenOwnerDeposit), nil)
	require.NoError(t, err)
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	_ = setupTestSandboxSC(t, chain1, nil, w)
	contractAgentID2 := setupTestSandboxSC(t, chain2, nil, w)

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := isc.NewAgentID(userAddress)
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	fmt.Println("---------------chain1---------------")
	fmt.Println(chain1.DumpAccounts())
	fmt.Println("---------------chain2---------------")
	fmt.Println(chain2.DumpAccounts())
	fmt.Println("------------------------------------")

	chain1TotalBaseTokens := chain1.L2TotalBaseTokens()
	chain2TotalBaseTokens := chain2.L2TotalBaseTokens()

	chain1.WaitForRequestsMark()
	chain2.WaitForRequestsMark()

	// send base tokens to contractAgentID2 (that is an entity of chain2) on chain1
	const baseTokensCreditedToScOnChain1 = 10 * isc.Million
	creditBaseTokensToSend := baseTokensCreditedToScOnChain1 + gas.LimitsDefault.MinGasPerRequest
	_, err = chain1.PostRequestSync(solo.NewCallParams(
		accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		accounts.ParamAgentID, contractAgentID2,
	).
		AddBaseTokens(creditBaseTokensToSend).
		AddAllowanceBaseTokens(baseTokensCreditedToScOnChain1).
		WithMaxAffordableGasBudget(),
		userWallet)
	require.NoError(t, err)

	chain1TransferAllowanceReceipt := chain1.LastReceipt()
	chain1TransferAllowanceGas := chain1TransferAllowanceReceipt.GasFeeCharged

	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-creditBaseTokensToSend)
	chain1.AssertL2BaseTokens(userAgentID, creditBaseTokensToSend-baseTokensCreditedToScOnChain1-chain1TransferAllowanceGas)
	chain1.AssertL2BaseTokens(contractAgentID2, baseTokensCreditedToScOnChain1)
	chain1.AssertL2TotalBaseTokens(chain1TotalBaseTokens + creditBaseTokensToSend)
	chain1TotalBaseTokens += creditBaseTokensToSend

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

	gasFeeTransferAccountToChain := 10 * gas.LimitsDefault.MinGasPerRequest
	// gas reserve for the 'TransferAllowanceTo' func call in 'TransferAccountToChain' func call
	gasReserve := 10 * gas.LimitsDefault.MinGasPerRequest
	withdrawFeeGas := 10 * gas.LimitsDefault.MinGasPerRequest
	const storageDeposit = wasmlib.StorageDeposit

	// NOTE: make sure you READ THE DOCS for accounts.transferAccountToChain()
	// to understand fully how to call it and why.

	// withdrawReqAllowance is the allowance provided to chain2.testcore.withdrawFromChain(),
	// which needs to be enough to cover any storage deposit along the way and to pay
	// the gas fees for the chain2.accounts.transferAccountToChain() request and the
	// chain1.accounts.transferAllowanceTo() request.
	// note that the storage deposit will be returned in the end
	withdrawReqAllowance := storageDeposit + gasFeeTransferAccountToChain + gasReserve

	// also cover gas fee for `FuncWithdrawFromChain` on chain2
	withdrawBaseTokensToSend := withdrawReqAllowance + withdrawFeeGas

	_, err = chain2.PostRequestSync(solo.NewCallParams(
		ScName, sbtestsc.FuncWithdrawFromChain.Name,
		sbtestsc.ParamChainID, chain1.ChainID,
		sbtestsc.ParamBaseTokens, baseTokensToWithdrawFromChain1,
		sbtestsc.ParamGasReserve, gasReserve,
		sbtestsc.ParamGasReserveTransferAccountToChain, gasFeeTransferAccountToChain,
	).
		AddBaseTokens(withdrawBaseTokensToSend).
		WithAllowance(isc.NewAssetsBaseTokens(withdrawReqAllowance)).
		WithMaxAffordableGasBudget(),
		userWallet)
	require.NoError(t, err)
	chain2WithdrawFromChainReceipt := chain2.LastReceipt()
	chain2WithdrawFromChainGas := chain2WithdrawFromChainReceipt.GasFeeCharged
	chain2WithdrawFromChainTarget := chain2WithdrawFromChainReceipt.DeserializedRequest().CallTarget()
	require.Equal(t, sbtestsc.Contract.Hname(), chain2WithdrawFromChainTarget.Contract)
	require.Equal(t, sbtestsc.FuncWithdrawFromChain.Hname(), chain2WithdrawFromChainTarget.EntryPoint)
	require.Nil(t, chain2WithdrawFromChainReceipt.Error)

	// accounts.FuncTransferAllowanceTo()
	// accounts.FuncTransferAccountToChain()
	require.True(t, chain1.WaitForRequestsThrough(2, 10*time.Second))
	// testcore.FuncWithdrawFromChain()
	// accounts.FuncTransferAllowanceTo()
	require.True(t, chain2.WaitForRequestsThrough(2, 10*time.Second))

	chain2TransferAllowanceReceipt := chain2.LastReceipt()
	// chain2TransferAllowanceGas := chain2TransferAllowanceReceipt.GasFeeCharged
	chain2TransferAllowanceTarget := chain2TransferAllowanceReceipt.DeserializedRequest().CallTarget()
	require.Equal(t, accounts.Contract.Hname(), chain2TransferAllowanceTarget.Contract)
	require.Equal(t, accounts.FuncTransferAllowanceTo.Hname(), chain2TransferAllowanceTarget.EntryPoint)
	require.Nil(t, chain2TransferAllowanceReceipt.Error)

	chain1TransferAccountToChainReceipt := chain1.LastReceipt()
	chain1TransferAccountToChainGas := chain1TransferAccountToChainReceipt.GasFeeCharged
	chain1TransferAccountToChainTarget := chain1TransferAccountToChainReceipt.DeserializedRequest().CallTarget()
	require.Equal(t, accounts.Contract.Hname(), chain1TransferAccountToChainTarget.Contract)
	require.Equal(t, accounts.FuncTransferAccountToChain.Hname(), chain1TransferAccountToChainTarget.EntryPoint)
	require.Nil(t, chain1TransferAccountToChainReceipt.Error)

	fmt.Println("---------------chain1---------------")
	fmt.Println(chain1.DumpAccounts())
	fmt.Println("---------------chain2---------------")
	fmt.Println(chain2.DumpAccounts())
	fmt.Println("------------------------------------")

	// the 2 function call we did above are requests from L1
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-creditBaseTokensToSend-withdrawBaseTokensToSend)
	// on chain1 user only made the first transaction, so it is the same as its balance before 'WithdrawFromChain' function call
	chain1.AssertL2BaseTokens(userAgentID, creditBaseTokensToSend-baseTokensCreditedToScOnChain1-chain1TransferAllowanceGas)
	// gasFeeTransferAccountToChain is is used for paying the gas fee of the 'TransferAccountToChain' func call
	// in 'WithdrawFromChain' func call
	// gasReserve is used for paying the gas fee of the 'TransferAllowanceTo' func call in 'TransferAccountToChain' func call
	// So the token left in contractAgentID2 on chain1 is the unused gas fee
	chain1.AssertL2BaseTokens(contractAgentID2, gasFeeTransferAccountToChain-chain1TransferAccountToChainGas)
	// tokens in 'withdrawBaseTokensToSend' amount are moved with the request from L1 to L2
	// 'withdrawReqAllowance' is is the amount moved from chain1 to chain2 with the request
	// 'baseTokensToWithdrawFromChain1' is the amount we assigned to withdraw in 'WithdrawFromChain' func call
	chain1.AssertL2TotalBaseTokens(chain1TotalBaseTokens + (withdrawBaseTokensToSend - withdrawReqAllowance - baseTokensToWithdrawFromChain1))

	// tokens in 'withdrawBaseTokensToSend' amount are moved from L1 to L2 with the 'WithdrawFromChain' func call
	// token in 'withdrawReqAllowance' amount are withdrawn by contractAgentID2
	// and 'WithdrawFromChain' func call was sent by user on chain2, so its balance should deduct 'chain2WithdrawFromChainGas'
	chain2.AssertL2BaseTokens(userAgentID, withdrawBaseTokensToSend-withdrawReqAllowance-chain2WithdrawFromChainGas)
	chain2.AssertL2BaseTokens(contractAgentID2, baseTokensToWithdrawFromChain1+storageDeposit)
	chain2.AssertL2TotalBaseTokens(chain2TotalBaseTokens + baseTokensToWithdrawFromChain1 + withdrawReqAllowance)
}
