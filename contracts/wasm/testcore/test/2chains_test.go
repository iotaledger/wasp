package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func Test2Chains(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		// set up chain1 and chain2, deploy testcore contract on both,
		// and set up contract contexts for the testcore contract and
		// accounts core contract on both chains

		chain1 := wasmsolo.StartChain(t, "chain1")
		chain1.CheckAccountLedger()

		chain2 := wasmsolo.StartChain(t, "chain2", chain1.Env)
		chain2.CheckAccountLedger()

		ctx1 := deployTestCoreOnChain(t, w, chain1, nil)
		require.NoError(t, ctx1.Err)
		ctx2 := deployTestCoreOnChain(t, w, chain2, nil)
		require.NoError(t, ctx2.Err)

		ctxAcc1 := ctx1.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnDispatch)
		require.NoError(t, ctxAcc1.Err)
		ctxAcc2 := ctx2.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnDispatch)
		require.NoError(t, ctxAcc2.Err)

		testcore1 := ctx1.Account()
		testcore2 := ctx2.Account()
		accounts1 := ctxAcc1.Account()
		accounts2 := ctxAcc2.Account()

		// create a user with enough tokens to deposit as the withdrawal target
		user := wasmsolo.NewSoloAgent(chain1.Env, "user")
		userL1 := user.Balance()

		// snapshot the current balances on both chains
		bal1 := ctx1.Balances(user, testcore2, accounts2)
		bal2 := ctx2.Balances(user, testcore1, accounts1)

		// mark current state of each chain
		ctx1.WaitForPendingRequestsMark()
		ctx2.WaitForPendingRequestsMark()

		// We need to set up an account for chain2.testcore on chain1 with enough tokens
		// to cover the future 'withdrawalAmount'.
		// We do this by sending enough L1 tokens from 'user' as part of a request that
		// executes the chain1.accounts.TransferAllowanceTo() function. The L1 tokens
		// will be deposited in the L2 account of 'user', and that account will be used
		// as the source account for the allowance that will be transferred to the
		// chain2.testcore account. In addition, it has to cover any gas fees that the
		// accounts.TransferAllowanceTo() request will use up.
		const withdrawalAmount = 7 * isc.Million
		const transferAmount = withdrawalAmount + wasmlib.MinGasFee
		xfer := coreaccounts.ScFuncs.TransferAllowanceTo(ctxAcc1.Sign(user))
		xfer.Params.AgentID().SetValue(testcore2.ScAgentID())
		xfer.Func.TransferBaseTokens(transferAmount).
			AllowanceBaseTokens(withdrawalAmount).Post()
		require.NoError(t, ctxAcc1.Err)

		// verify the L1 balance change for 'user'
		userL1 -= transferAmount
		require.Equal(t, userL1, user.Balance())

		// The actual chain1.accounts.TransferAllowanceTo() gas fee will be credited to the common account because its balance is lower than the minimum expected
		bal1.UpdateFeeBalances(ctxAcc1.GasFee)
		// The 'user' account ends up with the remainder after both 'withdrawalAmount'
		// and the actual gas fee have been deducted from 'transferAmount' (zero)
		bal1.Add(user, transferAmount-withdrawalAmount-ctxAcc1.GasFee)
		// 'withdrawalAmount' should have been deposited in the chain2.testcore account
		bal1.Add(testcore2, withdrawalAmount)
		// verify these changes against the actual chain1 account balances
		bal1.VerifyBalances(t)
		// verify that no changes were made to the chain2 account balances yet
		bal2.VerifyBalances(t)

		// Now that the tokens are available for withdrawal, we will invoke
		// chain2.testcore.WithdrawFromChain(). This function will in turn
		// tell chain2.accounts.TransferAccountToChain() to transfer the
		// required tokens from the chain2.testcore account on chain1 to
		// the chain2.testcore account on chain2.

		// here is the (simplified) equivalent code that will be executed by
		// testcore.WithdrawFromChain():
		//
		// targetChain := f.Params.ChainID().Value()
		// withdrawal := f.Params.BaseTokens().Value()
		// transfer := wasmlib.ScTransferFromBalances(ctx.Allowance())
		// ctx.TransferAllowed(ctx.AccountID(), transfer)
		// const gasFee = wasmlib.MinGasFee
		// const gasReserve = wasmlib.MinGasFee
		// const storageDeposit = wasmlib.StorageDeposit
		// xfer := coreaccounts.ScFuncs.TransferAccountToChain(ctx)
		// xfer.Params.GasReserve().SetValue(gasReserve)
		// xfer.Func.TransferBaseTokens(storageDeposit + gasFee + gasReserve).
		//		AllowanceBaseTokens(withdrawal + storageDeposit + gasReserve).
		//		PostToChain(targetChain)
		//
		// Note that this function will post to chain1.accounts.TransferAccountToChain().
		// Therefore, it will need to make sure it provides enough storage deposit for
		// the request. This storage deposit in turn will be deposited into the
		// chain2.testcore account on chain1. Any gas fees for the TransferAccountToChain()
		// execution will come out if this account as well.

		// WasmLib automatically assumes a minimum of 'wasmlib.StorageDeposit' for storage
		// deposit when posting a request from within a WasmLib contract, so make sure
		// that this amount is available to the testcore.WithdrawFromChain() request.
		// It also needs to provide gas to the accounts.transferAccountToChain() and
		// accounts.transferAccountToChain() requests and the testcore.WithdrawFromChain()
		// request itself.
		// Note that the latter may be a Wasm contract, so we'll provide a cool million
		// extra base tokens to draw from so that the gas fees will be covered for sure.

		// NOTE: make sure you READ THE DOCS for accounts.transferAccountToChain()
		// to understand fully how to call it and why.

		// allowance for accounts.transferAccountToChain(): SD + GAS1 + GAS2
		xferDeposit := wasmlib.StorageDeposit
		const gasFeeTransferAccountToChain = 10 * wasmlib.MinGasFee
		const gasReserve = 10 * wasmlib.MinGasFee
		const gasWithdrawFromChain = 10 * wasmlib.MinGasFee
		xferAllowance := xferDeposit + gasReserve + gasFeeTransferAccountToChain
		f := testcore.ScFuncs.WithdrawFromChain(ctx2.Sign(user))
		f.Params.ChainID().SetValue(ctx1.CurrentChainID())
		f.Params.BaseTokens().SetValue(withdrawalAmount)
		f.Params.GasReserve().SetValue(gasReserve)
		f.Params.GasReserveTransferAccountToChain().SetValue(gasFeeTransferAccountToChain)
		f.Func.TransferBaseTokens(xferAllowance + gasWithdrawFromChain).
			AllowanceBaseTokens(xferAllowance).Post()
		require.NoError(t, ctx2.Err)

		// - chain1.accounts.TransferAllowanceTo() request by 'user'
		// - chain1.accounts.TransferAccountToChain() request by chain2.coretest.WithdrawFromChain()
		require.True(t, ctx1.WaitForPendingRequests(2))

		// - chain2.coretest.WithdrawFromChain() request by chain2 originator
		// - chain2.accounts.TransferAllowanceTo() request by chain1.TransferAccountToChain()
		require.True(t, ctx2.WaitForPendingRequests(2))

		// update context with latest gas fees, since context does not know
		// that chain1.accounts.TransferAllowanceTo() was triggered by chain2
		ctxAcc1.UpdateGasFees()

		// we'll need to know the actual gas fees for both requests on chain2
		receipts := ctx2.Chain.GetRequestReceiptsForBlockRange(0, 0)
		withdrawalReceipt := receipts[len(receipts)-2]
		transferReceipt := receipts[len(receipts)-1]

		// chain2.testcore account will be credited with SD+GAS1+GAS2, pay actual GAS1,
		// and be debited by SD+GAS2+'withdrawalAmount'
		bal1.UpdateFeeBalances(ctxAcc1.GasFee)
		bal1.Add(testcore2, xferDeposit+gasWithdrawFromChain+gasWithdrawFromChain-ctxAcc1.GasFee-xferDeposit-gasReserve-withdrawalAmount)
		// verify these changes against the actual chain1 account balances
		bal1.VerifyBalances(t)

		userL1 -= xferAllowance + gasWithdrawFromChain
		require.Equal(t, userL1, user.Balance())

		// The gas fees will be credited to chain1.Originator
		bal2.UpdateFeeBalances(withdrawalReceipt.GasFeeCharged)
		bal2.UpdateFeeBalances(transferReceipt.GasFeeCharged)
		// deduct coretest.WithdrawFromChain() gas fee from user's cool million
		bal2.Add(user, gasWithdrawFromChain-withdrawalReceipt.GasFeeCharged)
		// chain2.accounts1 will be credited with SD+GAS2+'withdrawalAmount', pay actual GAS2,
		// and be debited by SD+'withdrawalAmount', leaving zero
		bal2.Add(accounts1, xferDeposit+gasReserve+withdrawalAmount-transferReceipt.GasFeeCharged-xferDeposit-withdrawalAmount)
		// chain2.testcore account receives the withdrawn tokens and storage deposit
		bal2.Account += withdrawalAmount + xferDeposit
		// verify these changes against the actual chain2 account balances
		bal2.VerifyBalances(t)
	})
}
