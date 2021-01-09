# Exploring IOTA Smart Contracts

## Purpose
The document is an introductory tutorial of the IOTA Smart Contract 
platform (ISCP) for developers. 

The level of the document is technical. The target audience is software engineers who want 
to understand ISCP and the direction it is taking. in order to develop their own dApps 
and/or contribute to the development of the ISCP and the Wasp node. 

The approach in this tutorial is an introduction to main concepts through writing
tests to example smart contracts. 
For this, we use Go testing package codenamed [_Solo_](../../packages/solo/readme.md) in all examples in the tutorial.

The knowledge of Go programming and basics of Rust programming is a prerequisite. 
For detailed introduction into the ISCP architecture [follow the link](https://docs.google.com/document/d/1zNJZMdetCzwiBC85K6gWbnzgdT1RXuZCLsTclKdrVqc/edit?usp=sharing).

## The _Solo_ package
Solo is a Go package for writing tests for IOTA smart contracts. 
It allows the deployment of ISCP chains and smart contracts, provides toolkit for interaction with smart contracts, 
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
The 4 core contracts listed in the log (`root`, `accounts`, `blob`, `eventlog`) 
are automatically deployed on each new chain. You can see them listed in the test log together with their `contract IDs`.
 
The log message `state transition #0 --> #1` means the state of the chain has changed from block 
index 0 (the origin index of the empty state) to block index 1. 
The state #0 is the empty origin state, the #1 always contains all core smart contracts deployed as well as other 
initialization data of the chain.

The `chainID` and `chain owner ID` are respectively ID of the deployed chain `YuyEwXdT9btMmJiHPmykR91hYSJVwRD9ciq5TZBXaffo`
and the address of the wallet (the private key) which deployed that chain `A/Yk85765qdrwheQ4udj6RihxtPxudTSWF9qYe4NsAfp6K` 
(with the prefix `/A` to indicate that an address is the chain owner, not a smart contract).
 
Next: [Tokens and the Value Tangle](chapter2.md)