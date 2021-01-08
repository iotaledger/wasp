# Exploring IOTA Smart Contracts

## Purpose
The document is an introductory documentation and tutorial of the IOTA Smart Contract 
platform (ISCP) for developers. 

The level of the document is technical. The target audience is software engineers who want 
to understand ISCP and the direction it is taking, to develop their own dApps 
and/or contribute to the development of the ISCP and the Wasp node itself. 

The approach in this tutorial is an introduction to main concepts through writing
tests to example s,mart contracts. 
For this, we use testing package codenamed [_Solo_](../../packages/solo/readme.md) in all examples in the tutorial.

The knowledge of Go programming and basics of Rust programming is a prerequisite. 
For detailed introduction into the ISCP architecture [follow the link](https://docs.google.com/document/d/1zNJZMdetCzwiBC85K6gWbnzgdT1RXuZCLsTclKdrVqc/edit?usp=sharing).

## The _Solo_ package
Solo is a Go package for writing tests for IOTA smart contracts. 
It allows the deployment of ISCP chains and smart contracts, provides tools to interact with the smart contracts, 
manipulate tokens and ledger accounts in an environment that is almost 
identical to the distributed multi-chain environment of the ISCP. 

Smart contracts are notoriously isolated from the outside world. 
The effect of the user interaction with a smart contract is normally only observed in its state change. 
The approach in this tutorial is to explain all main concepts of ISCP development through 
loading smart contracts into the _Solo_ tests, invoking its functions and examining state changes.

ISCP is currently in active development, so things change and are less than perfect. 
In the current stage the ISCP software is experimental. 
We expect feedback from the community about hands-on experience. 
We also expect contribution to the development of ISCP itself, including Rust/Wasm development environment 
or, possibly, alternative VM implementations. 

_Solo_ is not a toy environment. It allows developers to develop and test real smart contracts and 
entire cross-chain protocols before deploying them on the distributed network.

Please follow [the link](install.md) for installation instructions.

## First example
The following is an example of a Solo test. 
It deploys a new chain and invokes a function in the `root` contract. 

The `root` contract always exists on any chain, 
so for this example there is no need to deploy a new contract.
The test writes to the testing log the main parameters of the chain, lists names and IDs of all four core contracts.

```go
func TestSolo1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex1")

	chainInfo, coreContracts := chain.GetInfo()   // calls view root::GetInfo
	require.EqualValues(t, 4, len(coreContracts)) // 4 core contracts deployed by default

	t.Logf("chainID: %s", chainInfo.ChainID)
	t.Logf("chain owner ID: %s", chainInfo.ChainOwnerID)
	for hname, rec := range coreContracts {
		t.Logf("    Core contract '%s': %s", rec.Name, coretypes.NewContractID(chain.ChainID, hname))
	}
}
```

The output of the test will be something like this:

```
=== RUN   TestSolo1
34:37.415	INFO	TestSolo1	solo/solo.go:153	deploying new chain 'ex1'
34:37.419	INFO	TestSolo1.ex1	vmcontext/runreq.go:177	eventlog -> '[req] [0]Ei4d6oUbcgSPnmpTupeLaTNoNf1hRu8ZfZfmw2KFKzZm: Ok'
34:37.420	INFO	TestSolo1.ex1	solo/run.go:75	state transition #0 --> #1. Requests in the block: 1. Posted: 0
34:37.420	INFO	TestSolo1	solo/clock.go:44	ClockStep: logical clock advanced by 1ms
34:37.420	INFO	TestSolo1.ex1	solo/solo.go:233	chain 'ex1' deployed. Chain ID: aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3
34:37.420	INFO	TestSolo1.ex1	solo/req.go:145	callView: root::getChainInfo
    solo_test.go:18: chainID: aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3
    solo_test.go:19: chain owner ID: A/UrYEv4Yh7WU1M29cKq73tb2CUx8EYXfJt6JZn5srw19U
    solo_test.go:21:     Core contract 'accounts': aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3::3c4b5e02
    solo_test.go:21:     Core contract 'blob': aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3::fd91bc63
    solo_test.go:21:     Core contract 'root': aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3::cebf5908
    solo_test.go:21:     Core contract 'eventlog': aEbE2vX6jrGhQ3AKHCPmQmn2qa11CpCRzaEgtVJRAje3::661aa7d8
--- PASS: TestSolo1 (0.01s)
```
The four core contracts listed (`root`, `accounts`, `blob`, `eventlog`) 
are automatically deployed on each new chain. You can see them listed in the test log together with their `contract IDs`.
 
The log message `state transition #0 --> #1` means that the chain has transited its state from block 
index 0 (the origin index of the empty state) to block index 1. 
The state #0 is the empty origin state, the #1 always contains all core smart contracts as well as other 
initialization data of the chain.

The `chainID` and `chain owner ID` are respectively ID of the deployed chain `YuyEwXdT9btMmJiHPmykR91hYSJVwRD9ciq5TZBXaffo`
 and the address of the wallet (the private key) which deployed that chain `A/Yk85765qdrwheQ4udj6RihxtPxudTSWF9qYe4NsAfp6K` 
 (with the prefix `/A` to indicate the address).
 
 ## Tokens and the Value Tangle
 The Pollen release of the Goshimmer node implements the _Value Tangle_, 
 a distributed ledger of tokens. We won't go into the detail of the Value Tangle. The introdution 
 of it can be found [here](../intro/utxo.md). Here we have to know that Value Tangle contains
 balances of colored tokens locked in addresses, like this: 
 ```
Address: Yk85765qdrwheQ4udj6RihxtPxudTSWF9qYe4NsAfp6K
    IOTA: 1000
    Red: 15
    Green: 200
```
where `IOTA` is the color code of IOTA tokens and `Red` and `Green` are other color codes. 
Tokens can only be moved on the _Value Tangle_ by the private of the corresponding address. We will use `private key`, 
`signature scheme` and `wallet` as synonyms in this tutorial.  

The `Solo` environment implements in-memory Value Tangle ledger to the finest details. 
You can only move tokens by creating and submitting valid and signed transaction in Solo. 
You also can create new wallets on the Value Tangle, request iotas from the faucet.

The following code shows how to do that:
```go
func TestSolo2(t *testing.T) {
	env := solo.New(t, false, false)
	userWallet := env.NewSignatureSchemeWithFunds()   // create new wallet with 1337 iotas
	userAddress := userWallet.Address()
	t.Logf("Address of the userWallet is: %s", userAddress)
	numIotas := env.GetUtxodbBalance(userAddress, balance.ColorIOTA)  // how many iotas contains the address
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, balance.ColorIOTA, 1337) // assert the address has 1337 iotas
}
```
 The output:
```
=== RUN   TestSolo2
    solo_test.go:29: Address of the userWallet is: WUwewZS3JFtEUtsfR5HcUANzyADv8pSmK7j6SuayNDRv
    solo_test.go:31: balance of the userWallet is: 1337 iota
--- PASS: TestSolo2 (0.00s)
```
 
The Value Tangle token ledger is shared among all chains deployed on the global environment `env`
of the test. It serves as a medium for transactions between smart contracts on different chains. 
It makes it possible on _Solo_ to test transacting between chains: the on-tangle messaging.
 
 Note that in the test above we didn’t deploy any chains: the Value Tangle exists in the `env` variable, 
outside of any chains.

# TODO (in progress)