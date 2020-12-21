// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package 'solo' is a development tool for writing unit tests for IOTA Smart Contracts (ISCP).
//
// A development tool
//
// The package is intended for developers of smart contracts as well as contributors to the development
// of the ISCP and the Wasp node itself.
//
// Normally, the smart contract is developed and tested in the
// 'solo' environment before trying it out on the network of Wasp nodes.
// Running and testing the smart contract on 'solo' does not require to run the Wasp
// nodes nor committee of nodes: just ordinary 'go test' environment.
//
// Native environment
//
// 'solo' shares the same code of Virtual Machine with the Wasp node. This guarantees that smart contract programs
// can later be deployed on chains which are run by the network of Wasp nodes without any modifications.
//
// The 'solo' environment uses in-memory UTXO ledger to validate and store transactions. The UTXODB
// mocks Goshimmer UTXO ledger, it uses same value transaction structure, colored tokens, signature
// schemes as well as transaction and signature validation as in Value Tangle of Goshimmer (Pollen release).
// The only difference with the Value Tangle is that UTXODB provides full synchronicity of ledger updates.
//
// The virtual state (key/value database) in 'solo' is an in-memory database. It provides exactly the same
// interface of access to it as the database of the Wasp node.
//
// Writing smart contracts
//
// The smart contracts are usually written in Rust using Rust libraries provided
// in the 'wasplib' repository at https://github.com/iotaledger/wasplib.
// Rust code is compiled into the WebAssembly (Wasm) binary.
// The Wasm binary is uploaded by 'solo' onto the chain and then loaded into the VM
// and executed.
//
// Another option to write and run ISCP smart contracts is to use the native Go environment
// of the Wasp node and 'Sandbox' interface, provided by the Wasp for the VM: the "hardcoded" mode. The latter approach is not normally used to develop apps,
// however is used for the 4 builtin contracts which constitutes the core of the ISCP chains.
// The approach to write "hardcoded" smart contracts may also be very useful for
// the development and debugging of the smart contract logic in IDE such as GoLand, before writing it as
// a Rust/Wasm smart contract.
//
// Example test
//
// The following example deploys chain and retrieves basic info from the deployed chain.
// It is expected 4 core contracts deployed on it by default and the test prints them.
//	func TestExample1(t *testing.T){
//		glb := solo.New(t, false, false)
//		chain := glb.NewChain(nil, "exampleChain")
//
//		chainID, ownerID, coreContracts := chain.GetInfo()  // calls view root::GetInfo
//
//		require.EqualValues(t, 4, len(coreContracts))  // 4 core contracts deployed by default
//		t.Logf("chainID: %s", chainID)
//		t.Logf("chain owner ID: %s", ownerID)
//		for _, rec := range coreContracts{
//			t.Logf("    Contract: %s", rec.Name)
//		}
//	}
// will produce the following output:
//  	=== RUN   TestExample1
//	12:24:16.779	INFO	TestExample1	solo/solo.go:144	deploying new chain 'exampleChain'
//	12:24:16.783	INFO	TestExample1.exampleChain	solo/run.go:75	state transition #0 --> #1. Requests in the block: 1. Posted: 0
//	12:24:16.783	INFO	TestExample1	solo/clock.go:44	ClockStep: logical clock advanced by 1ms
//	12:24:16.783	INFO	TestExample1.exampleChain	solo/solo.go:220	chain 'exampleChain' deployed. Chain ID: JG1UBRzEesdkTnLJBc7oK9HTm98ioNekdKLJ1f2ScK5Q
//	12:24:16.783	INFO	TestExample1.exampleChain	solo/req.go:136	callView: root::getChainInfo
//		example1_test.go:17: chainID: JG1UBRzEesdkTnLJBc7oK9HTm98ioNekdKLJ1f2ScK5Q
//		example1_test.go:18: chain owner ID: A/SHRk4KG8ruMB7GwoqWVsaeHreQAKFCUc45bBCHwnrk5h
//		example1_test.go:20:     Contract: accounts
//		example1_test.go:20:     Contract: chainlog
//		example1_test.go:20:     Contract: root
//		example1_test.go:20:     Contract: blob
//	--- PASS: TestExample1 (0.00s)
//PASS
package solo
