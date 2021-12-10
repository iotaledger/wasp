---
description: Smart contracts can exchange assets between themselves on the same chain and also between different chains, as well as with addresses on the UTXO Ledger.
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
# Account Balances

:::note

The example code can be found in the [Wasp repository](https://github.com/iotaledger/wasp/tree/develop/documentation/tutorial-examples).

:::
Each chain in the _IOTA Smart Contracts_ is a separate ledger, different from the UTXO ledger.
Multiple chains add another dimension on top of the UTXO Ledger. Smart contracts
can exchange assets between themselves on the same chain and also between different chains, as well as with
addresses on the UTXO Ledger. We will skip explaining the whole picture for the time
being and will concentrate on one specific use case.

Imagine you have a wallet, the private key (with the address), and some tokens on
that address on the UTXO Ledger, the Tangle. The use case is about sending those tokens to a smart contract on a chain
and receiving these tokens back to the address.

Here we explore the concept of `on-chain account`(a.k.a. just `account`). On the UTXO
Ledger, the private key is represented by the address (the hash of the public
key). That address holds balances of colored tokens. Those tokens are
_controlled_ by the private key.

In IOTA Smart Contracts, we extend the concept of _address_ with the concept of `account`. An
`account` contains colored tokens just like an `address`. The `account` is
located on some chain, and it is controlled by the same private key as the
associated address. So, an address can control tokens on the UTXO Ledger
(Level 1 or `L1`) and on each chain (Level 2 or `L2`).

So, the chain essentially is a custodian of the tokens deposited in its `accounts`.

The following test demonstrates how a wallet can deposit tokens in a chain
account and then withdraw them.

```go
func TestTutorial5(t *testing.T) {
 env := solo.New(t, false, false, seed)
 chain := env.NewChain(nil, "ex5")
 // create a wallet with 1000000 iotas.
 // the wallet has address and it is globally identified
 // through a universal identifier: the agent ID
 userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(5))
 userAgentID := iscp.NewAgentID(userAddress, 0)

 env.AssertAddressBalance(userAddress, colored.IOTA, solo.Saldo)
 chain.AssertAccountBalance(userAgentID, colored.IOTA, 0) // empty on-chain

 t.Logf("Address of the userWallet is: %s", userAddress.Base58())
 numIotas := env.GetAddressBalance(userAddress, colored.IOTA)
 t.Logf("balance of the userWallet is: %d iota", numIotas)
 env.AssertAddressBalance(userAddress, colored.IOTA, solo.Saldo)

 // send 42 iotas from wallet to own account on-chain, controlled by the same wallet
 req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).WithIotas(42)
 _, err := chain.PostRequestSync(req, userWallet)
 require.NoError(t, err)

 // check address balance: must be 42 iotas less
 env.AssertAddressBalance(userAddress, colored.IOTA, solo.Saldo-42)
 // check the on-chain account. Must contain 42 iotas
 chain.AssertAccountBalance(userAgentID, colored.IOTA, 42)

 // withdraw all iotas back to the sender
 req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).WithIotas(1)
 _, err = chain.PostRequestSync(req, userWallet)
 require.NoError(t, err)

 // we are back to initial situation: IOTA is fee-less!
 env.AssertAddressBalance(userAddress, colored.IOTA, solo.Saldo)
 chain.AssertAccountBalance(userAgentID, colored.IOTA, 0) // empty
}
```

The example above creates a chain, then creates a wallet with `solo.Saldo` iotas (1 Mi) and
sends (deposits) 42 iotas to the corresponding on-chain account by posting
a `deposit` request to the `accounts` core contract on that chain. That account
will now contain 42 iotas. The address on the UTXO Ledger will contain 42 iotas
less, of course.

In the next step, the same wallet (`userWallet`) will withdraw all 42 iotas back
to the address by sending a `withdraw` request to the `accounts` contract on
the same chain.

If the same request would be posted from another user wallet (another private
key), the `withdraw` request would fail. Try it! Only the owner of the address
can move those funds from the on-chain account.
