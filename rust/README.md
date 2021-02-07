## WasmLib for Rust

The interface provided by `WasmLib` hides the underlying complexities of the
Iota Smart Contract Protocol (ISCP)

`WasmLib` allows developers to use Rust to create smart contracts for ISCP
that compile into Wasm and can run directly on Iota's ISCP-enabled Wasp nodes.

`WasmLib` treats the programming of smart contracts as simple access to a
key/value data storage where smart contract properties, request parameters,
and the smart contract state can be accessed in a universal, consistent way.

The _wasmlib_ folder provides the interface to the ISCP through _ScCallContext_
and _ScViewContext_.

The _contracts_ folder contains a number of example smart contracts that can be
used to learn how to use _ScCallContext_ and _ScViewContext_ properly. For more
information on how to go about creating your own smart contracts see the
README.md in this folder.

