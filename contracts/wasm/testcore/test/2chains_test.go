package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
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

		const withdrawalAmount = 7 * isc.Million
		const gasBudgetForDeposit = 200

		// We need to set up an account for chain2.testcore on chain1 with
		// enough tokens to cover the future 'withdrawalAmount'.
		// We do this by sending enough L1 tokens from 'user' as part of a
		// request that executes the chain1.accounts.TransferAllowanceTo()
		// function. The L1 tokens will be deposited in the L2 account of 'user',
		// and that account will be used as the source account for the allowance
		// that will be transferred to the chain2.testcore account, plus it has
		// to cover any gas fees that the TransferAllowanceTo() call will use up.
		const transferAmount = withdrawalAmount + gasBudgetForDeposit
		xfer := coreaccounts.ScFuncs.TransferAllowanceTo(ctxAcc1.Sign(user))
		xfer.Params.AgentID().SetValue(testcore2.ScAgentID())
		xfer.Func.TransferBaseTokens(transferAmount).
			AllowanceBaseTokens(withdrawalAmount).Post()
		require.NoError(t, ctxAcc1.Err)

		// verify the L1 balance change for 'user'
		userL1 -= transferAmount
		require.Equal(t, userL1, user.Balance())

		// The actual chain1.accounts.TransferAllowanceTo() gas fee will be credited chain1.Common
		bal1.Chain += ctxAcc1.GasFee
		// The 'user' account ends up with the remainder after 'withdrawalAmount'
		// and the actual gas fee have both been deducted from 'transferAmount'
		bal1.Add(user, transferAmount-withdrawalAmount-ctxAcc1.GasFee)
		// 'withdrawalAmount' should have been deposited in the chain1.testcore account
		bal1.Add(testcore2, withdrawalAmount)
		// verify these changes against the actual chain1 account balances
		bal1.VerifyBalances(t)
		// verify that no changes were made to the chain2 account balances
		bal2.VerifyBalances(t)

		// Now that the tokens are available for withdrawal, we will invoke
		// chain2.testcore.WithdrawFromChain(). This function will in turn
		// tell chain2.accounts.Withdraw() to withdraw the required tokens
		// from the chain2.testcore account on chain1 and deposit them into
		// the chain2.testcore account on chain2.

		// here is the (simplified) equivalent code that will be executed by
		// testcore.WithdrawFromChain():
		//
		// targetChain := f.Params.ChainID().Value()
		// withdrawal := f.Params.BaseTokensWithdrawal().Value()
		// transfer := wasmlib.NewScTransferFromBalances(ctx.Allowance())
		// ctx.TransferAllowed(ctx.AccountID(), transfer)
		// withdraw := coreaccounts.ScFuncs.Withdraw(ctx)
		// withdraw.Func.Transfer(transfer).AllowanceBaseTokens(withdrawal).PostToChain(targetChain)
		//
		// Note that this function will post a request to chain1.accounts.Withdraw().
		// Therefore, it will need to make sure it provides enough storage deposit
		// for the request to be placed on the Tangle. This storage deposit in turn
		// will be deposited in the chain2.testcore account on chain1. Any gas fees
		// for the Withdraw() execution will come out if this account as well.

		// WasmLib automatically assumes a minimum of 'WasmStorageDeposit' for storage
		// deposit when posting a request from within a WasmLib contract, so make sure
		// that this amount is available to the testcore.WithdrawFromChain() request.
		const storageDepositForWithdraw = wasmhost.WasmStorageDeposit
		f := testcore.ScFuncs.WithdrawFromChain(ctx2)
		f.Params.ChainID().SetValue(ctx1.CurrentChainID())
		f.Params.BaseTokensWithdrawal().SetValue(withdrawalAmount)
		f.Func.AllowanceBaseTokens(storageDepositForWithdraw).Post()
		require.NoError(t, ctx2.Err)

		// Note that the accounts.Withdraw() request determines where to withdraw the
		// tokens by looking at the caller. When the caller is an ordinary user, the
		// tokens are withdrawn to that user's associated L1 account. But a contract
		// does not have a L1 address associated with it. Therefore, withdrawal will
		// be done to the contract's L2 account instead. This only makes sense when
		// withdrawing between different chains. Because withdrawal happens *from* the
		// caller's L2 account. So having a contract withdraw from its own L2 account
		// on its own chain is essentially a no-op because it would withdraw *from*
		// its own L2 account *into* its own L2 account.
		// Withdrawing from a contract's L2 account between chains will work, but there
		// are a few issues to take into account.
		// The accounts.Withdraw() function will need to deposit the tokens into the
		// calling contract's L2 account. This it can only do by invoking the same
		// accounts.TransferAllowanceTo() that we used above to deposit tokens into
		// the chain1.coretest account. The Withdraw() function on the other chain will
		// first transfer the allowed tokens into its own account on that chain, and then
		// use a TransferAllowanceTo() request to send them from there to the contract's
		// L2 account on the other chain.
		// This in turn requires a request to be sent to the calling contract's chain by
		// the accounts contract. The tokens that get sent by the accounts contract will
		// be deposited into an account for that contract on the other chain. The amount
		// withdrawn should therefore be >= the necessary storage deposit for that chain,
		// or the request will fail. In addition, the request will need to cover any gas
		// fees required to execute accounts.TransferAllowanceTo() on that chain, and this
		// fee will be taken out of the accounts contract's L2 account on that chain.
		// The issue is that we don't know what these fees will be, so as a temp solution
		// the accounts contract will reserve 'ConstDepositFeeTmp' tokens for this purpose.
		// This value has been set pretty high (1M tokens) to always be able to cover the
		// gas fee. But this means that the remainder, which can be pretty high when gas
		// fees are low, stays locked in the L2 account of the accounts contract on that
		// chain and will be irretrievably lost.
		// So to sum it up, it IS possible for a contract to call accounts.Withdraw(),
		// but it will come at a price of 'ConstDepositFeeTmp' tokens.

		// - chain1.accounts.TransferAllowanceTo() request by 'user'
		// - chain1.accounts.Withdraw() request by chain2.coretest.WithdrawFromChain()
		require.True(t, ctx1.WaitForPendingRequests(2))

		// - chain2.coretest.WithdrawFromChain() request by chain2 originator
		// - chain2.accounts.TransferAllowanceTo() request by chain1.Withdrawal()
		require.True(t, ctx2.WaitForPendingRequests(2))

		// update context with latest gas fees, since context does not
		// know that chain1.accounts.Withdraw() was triggered by chain2
		ctxAcc1.UpdateGasFees()

		// The chain1.accounts.Withdraw() gas fee will be credited to chain1.Common
		bal1.Chain += ctxAcc1.GasFee
		// chain2.testcore account will pay the gas, receive 'storageDepositForWithdraw',
		// and be debited by 'withdrawalAmount'
		// TODO maybe withdraw 'storageDepositForWithdraw' as part of Withdraw() request?
		bal1.Add(testcore2, storageDepositForWithdraw-ctxAcc1.GasFee-withdrawalAmount)
		// verify these changes against the actual chain1 account balances
		bal1.VerifyBalances(t)

		// we'll need to know the actual gas fees for both requests on chain2
		receipts := ctx2.Chain.GetRequestReceiptsForBlockRange(0, 0)
		prevReceipt := receipts[len(receipts)-2]
		lastReceipt := receipts[len(receipts)-1]

		// The gas fees will be credited to chain1.Common
		bal2.Chain += prevReceipt.GasFeeCharged + lastReceipt.GasFeeCharged
		// chain2.originator initiated coretest.WithdrawFromChain()
		// deduct gas fee and 'storageDepositForWithdraw'
		// TODO why is wasmhost.WasmStorageDeposit added to Originator?
		bal2.Originator += wasmhost.WasmStorageDeposit - prevReceipt.GasFeeCharged - storageDepositForWithdraw
		// chain1.accounts account is stuck with the locked tokens
		bal2.Add(accounts1, accounts.ConstDepositFeeTmp-lastReceipt.GasFeeCharged)
		// chain2.testcore account receives the withdrawn tokens
		bal2.Account += withdrawalAmount - accounts.ConstDepositFeeTmp
		// verify these changes against the actual chain2 account balances
		bal2.VerifyBalances(t)
	})
}
