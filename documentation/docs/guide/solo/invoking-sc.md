---
description: In Solo, you can post an on-ledger request with a single call to PostRequestSync.  Alternatively, you can post an off-ledger request by using chain.PostRequestOffLedger instead of PostRequestSync.
image: /img/tutorial/send_request.png
keywords:
- testing
- PostRequestSync
- PostRequestOffLedger
- send
- requests
- post
- solo
- on-ledger
- off-ledger
---
# Invoking Smart Contracts

Invoking a Smart Contract's entry point in Solo looks like the following:

```go
  req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!").WithIotas(1)
  _, err = chain.PostRequestSync(req, nil)
```

In the example `TestTutorial3` we invoke the `storeString` entry point of the
`example1` smart contract by posting it as a request. The parameter
named `paramString` is passed with the string value "Hello, world!". The _Solo_
test code itself is separate from the chain where the smart contract is invoked and is not executed "on-chain".

`NewCallParams` creates a call object named `req` which wraps all call
parameters into a single object. This is syntactic sugar just for convenience.
In this case, the call object only wraps the target smart contract name, target
entry point name, and one named parameter `paramString`. In other cases it could
contain many parameters.

`WithIotas` attaches a single iota with the request. All `on-ledger` requests
are implemented as value transactions with additional data, and therefore we
need to transfer at least a single token for the request to be valid.

`PostRequestSync` sends the request to the chain. Let’s describe in detail what 
is going on here.

[![Generic process of posting an on-ledger request to the smart contract](/img/tutorial/send_request.png)](/img/tutorial/send_request.png)

The diagram above depicts the generic process of posting an `on-ledger` request to the smart
contract. The same picture is valid for the _Solo_ environment and for any other
requester which sends an `on-ledger` request to the smart contract, for example, the IOTA Smart Contracts
wallet or another chain.

Posting the request always consists of the steps below. Note that in Solo all 7
steps are carried out by the single call to `PostRequestSync`.

1. Creating the smart contract transaction which wraps the request with metadata
   and moves tokens. Each request transaction is a value transaction, it always
   moves at least one token. Therefore, each request transaction must be signed
   by the private key of the owner of the tokens: the requester. That securely
   identifies each requester in IOTA Smart Contracts. In Solo, the transaction is signed by the
   private key provided in the second parameter of the `PostRequestSync`
   call (see below).
2. Posting the request transaction to the Tangle and confirming it. In _Solo_ it
   is just adding the transaction to the `UTXODB ledger`, the emulated UTXO
   Ledger, so it is confirmed immediately and synchronously. The confirmed
   transaction on the ledger becomes part of the backlog of requests to the
   chain. In the real UTXO Ledger the sender would have to wait until the ledger
   confirms the transaction.
3. The chain picks the request from the backlog and runs the request on the VM.
4. The VM calls the target entry point of the smart contract program. The
   program updates the state.
5. The VM produces a state update transaction (the `anchor`).
6. The chain signs the transaction with its private key. In the _Solo_
   environment is the `ChainSigScheme` property of the chain. In the real
   Wasp environment it is the threshold signature of the committee of validator nodes.
7. The chain posts the resulting transaction to the Tangle and, after confirmation, solidifies the corresponding state. In the _Solo_ environment it adds
the transaction to the UTXODB ledger.

The following lines in the log correspond to step 7:

```log
54:43.809 INFO TestTutorial3.ex3 vmcontext/runreq.go:311 eventlog -> '[req] [0]CHvU6BUDgt9MZJTxsYMZ1p1veg591mvwKGQBJd2KYdaB: Ok'
54:43.809 INFO TestTutorial3 solo/clock.go:35 AdvanceClockBy: logical clock advanced by 2ns
54:43.809 INFO TestTutorial3.ex3.m mempool/mempool.go:119 OUT MEMPOOL [0]CHvU6BUDgt9MZJTxsYMZ1p1veg591mvwKGQBJd2KYdaB
54:43.809 INFO TestTutorial3.ex3 solo/run.go:86 state transition #2 --> #3. Requests in the block: 1. Outputs: 1
```

The chain adds a record about any successfully processed request
`[0]CHvU6BUDgt9MZJTxsYMZ1p1veg591mvwKGQBJd2KYdaB` to the immutable on-chain log.

The statement `_, err = chain.PostRequestSync(req, nil)` in the Solo test uses `nil`
for the default signature scheme of the requester. The `OriginatorSigScheme`,
the one which deployed the chain, is used as the default requester. In the
_Solo_ environment you can create other identities for requesters (“wallets”)
with `NewKeyPairWithFunds`.

## Off-ledger Requests

Alternatively, in the example above, we could send an off-ledger request by using `chain.PostRequestOffLedger` instead of `PostRequestSync`.
However, to be able to submit off-ledger request, the account sending the request must deposit funds to the chain beforehand.

```go
  wallet, address := env.NewKeyPairWithFunds()
  AgentID := iscp.NewAgentID(ownerAddr, 0)

  // deposit into the account
  req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).WithIotas(100)
  _, err := chain.PostRequestSync(req, owner)
  require.NoError(t, err)

  //no .WithIotas() needed, fees will be deducted from the sender on-chain account
  req := solo.NewCallParams("example1", "storeString", "paramString", "Hello, world!")
  _, err = chain.PostRequestOffLedger(req, wallet)

```
