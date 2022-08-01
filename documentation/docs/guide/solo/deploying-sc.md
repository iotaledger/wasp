---
description: Deploying Wasm smart contracts with Solo.
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
# Deploying Wasm Smart Contracts

:::note
For information about how to create Wasm smart contracts, refer to the [Wasm VM chapter](../wasm_vm/intro.mdx).
:::

The following examples will make use of the [`solotutorial` Rust/Wasm smart contract](https://github.com/iotaledger/wasp/tree/develop/documentation/tutorial-examples/src/solotutorial.rs).
In order to test the smart contract using Solo, we first need to deploy it:

```go
func TestTutorialDeploySC(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "example")
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)
}
```

That's it! This will work as long as the `solotutorial_bg.wasm` file is in the same directory as the Go test code.

The first parameters to `NewChain` and `DeployWasmContract` is the key pair of the chain originator and the deployer of the smart contract, respectively.
Conveniently, we can pass `nil` to use a default wallet, which can be accessed as `chain.OriginatorPrivateKey`.

The second parameter to `DeployWasmContract` (`"solotutorial"`), is the name assigned to the smart contract instance.
Smart contract instance names must be unique across each chain.

:::note
In the example above we enabled the `AutoAdjustDustDeposit` option.
This is necessary in order to automatically adjust all sent L1 transactions to include the storage deposit if necessary (provided that the sender owns the funds).
It it is possible to disable the option and have manual control of the storage deposit, but in that case the deployment of the smart contract will have to be done "by hand".
In most cases it is recommended to leave it enabled.
:::
