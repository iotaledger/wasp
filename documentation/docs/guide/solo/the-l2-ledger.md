---
description: Smart contracts can exchange assets between themselves on the same chain and also between different chains, as well as with addresses on the L1 ledger.
image: /img/logo/WASP_logo_dark.png
keywords:
- testing
- solo
- account
- address
- wallet
- balances
- ledger
---
# The L2 Ledger

Each chain in IOTA Smart Contracts contains its own L2 ledger, independent from the L1 ledger.
Smart contracts can exchange assets between themselves on the same chain and also between different chains, as well as with addresses on the L1 Ledger.

Let's imagine that we have a wallet with some tokens on the L1 ledger, and we want to send those tokens to a smart contract on a chain, and later receive these tokens back on L1.

On the L1 ledger, our wallet's private key is represented by an address, which holds some tokens.
Those tokens are _controlled_ by the private key.

In IOTA Smart Contracts the L2 ledger is a collection of _on-chain accounts_ (sometimes also called just _accounts_).
Each L2 account is controlled by the same private key as its associated address, and can hold tokens on the chain's ledger, just like an address can hold tokens on L1.
This way, the chain is essentially a custodian of the tokens deposited on its accounts.

The following test demonstrates how a wallet can deposit tokens in a chain
account and then withdraw them.
Note that the math is made somewhat more complex by the gas fees and storage deposit.
We could ignore those but we include them here to show exactly how they are handled.

```go
func TestTutorialAccounts(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "ch1")

	// create a wallet with some base tokens on L1:
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	// the wallet can we identified on L2 by an AgentID:
	userAgentID := isc.NewAgentID(userAddress)
	// for now our on-chain account is empty:
	chain.AssertL2BaseTokens(userAgentID, 0)

	// send 1 Mi from the L1 wallet to own account on-chain, controlled by the same wallet
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		AddBaseTokens(1 * isc.Million)

	// estimate the gas fee and storage deposit
	gas1, gasFee1, err := chain.EstimateGasOnLedger(req, userWallet, true)
	require.NoError(t, err)
	storageDeposit1, err := chain.EstimateNeededStorageDeposit(req, userWallet)
	require.NoError(t, err)
	require.Zero(t, storageDeposit1) // since 1 Mi is enough

	// send the deposit request
	req.WithGasBudget(gas1).
		AddBaseTokens(gasFee1) // including base tokens for gas fee
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// our L1 balance is 1 Mi + gas fee short
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-1*isc.Million-gasFee1)
	// our L2 balance is 1 Mi
	chain.AssertL2BaseTokens(userAgentID, 1*isc.Million)
	// (the gas fee went to the chain's private account)

	// withdraw all base tokens back to L1
	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		WithAllowance(isc.NewAllowanceBaseTokens(1 * isc.Million))

	// estimate the gas fee and storage deposit
	gas2, gasFee2, err := chain.EstimateGasOnLedger(req, userWallet, true)
	require.NoError(t, err)
	storageDeposit2, err := chain.EstimateNeededStorageDeposit(req, userWallet)
	require.NoError(t, err)

	// send the withdraw request
	req.WithGasBudget(gas2).
		AddBaseTokens(gasFee2 + storageDeposit2). // including base tokens for gas fee and storage
		AddAllowanceBaseTokens(storageDeposit2)   // and withdrawing the storage as well
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// we are back to the initial situation, having been charged some gas fees
	// in the process:
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-gasFee1-gasFee2)
	chain.AssertL2BaseTokens(userAgentID, 0)
}
```

The example above creates a chain and a wallet with `utxodb.FundsFromFaucetAmount` base tokens on L1.
Then it sends 1 Mi to the corresponding on-chain account by posting a `deposit` request to the `accounts` core contract on the chain.
Later, it sends a `withdraw` request to the `accounts` core contract, in order to get the tokens back to L1.

Both requests are affected by the gas fees and the storage deposit.
In some cases it is possible to ignore these amounts, if they are negligible compared to the amounts being transferred.
In our case, however, we want to be very precise.
Let's inspect the first request to see what is going on.

1. First we create a request to deposit the funds, with `solo.NewCallParams`.
2. Since we want to deposit 1 Mi, we call `AddBaseTokens(1 * isc.Million)`. This
  instructs Solo to take that amount from our L1 balance and add it to the
  transaction (this is only possible for on-ledger requests).
3. Once the request is executed by the chain, it will be charged some gas fee.
  We use `chain.EstimateGasOnLedger` before actually sending the request.
4. On-ledger requests also require some storage deposit. We use
  `EstimateNeededStorageDeposit` for this, and then realize that the 1 Mi
  already included is enough for the storage deposit, so no need to add more.
5. We adjust the request with the gas budget and the gas fee with `WithGasBudget` and `AddBaseTokens` respectively.
6. Finally, we send the on-ledger request with `PostRequestSync`.
7. The chain picks up the request. Any attached base tokens (1 Mi + gas fee) are automatically credited to the sender's L2 account.
8. The chain executes the request. The gas fee is deducted from the sender's L2
   account.
9. We have exactly 1 Mi on our L2 balance.

The process for the `withdraw` request is similar, with two main differences:

* We need to ensure that the L1 transaction contains enough funds for the storage deposit (because it is larger than the gas fee). These tokens are automatically deposited on our L2 account, and we immediately withdraw them back.
* We use `AddAllowanceBaseTokens` to set the _allowance_ of our request. The allowance specifies the maximum amount of tokens that the smart contract is allowed to debit from the sender's L2 account.

Note that if we posted the same `deposit` request from another user wallet (another private key), it would fail.
Try it! Only the owner of the address can move those funds from the on-chain account.

Also try removing the `AddAllowanceBaseTokens` call: it will fail because a smart contract is not allowed to deduct funds from the sender's L2 balance unless explicitly authorized by the allowance.
