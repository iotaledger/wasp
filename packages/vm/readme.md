# VM abstraction

## Contents

* General
* The VM
  * Calling the VM
  * Result of running the VM
  * Structure of the VM
  * Deployment of the smart contract
  * VM type
  * Program binary and blob hash
* Processor and the sandbox interface
  * `Sandbox` interface
  * `SandboxView` interface
* Implementation of EVM. A Virtual Ethereum

## General

By _VM abstraction_ in ISCP we understand a collection of abstract interfaces which makes whole architecture of ISCP and Wasp
node _agnostic_ about what exactly kind of deterministic computation machinery is used to run smart contract programs.

In ISCP we distinguish two things:
* The VM (Virtual Machine)
* a VM plugin, a pluggable part of VM

The _VM_ itself is a deterministic executable, a "black box", which is used by the distributed part of the protocol,
to calculate outputs from inputs before, the come to _consensus_ among different VMs and finally submitting it to the ledger  
as the state update of the chain, the block.  
Naturally, results of calculations, the output, is fully defined by inputs.

The _VM plugin_ is a _processor_ which is called by _VM_. The _processor_ is known to the VM through [coretypes.VMProcessor](../coretypes/vmprocessor.go#L23) interface. Wasp node can have several different implementation
of _VMProcessor_ interface. The _VM_ has many processors. Usually, each processor represents one smart contract. Alternatively,  
one processor can represent entirely new plugged-in VM, such as EVM.

For details about implementation of _VMProcessor_ see below.

In Wasp node, the VM related code is mostly located in [wasp/packages/vm](../vm) directory.  
The globally defined data types and definitions are located in [wasp/packages/coretypes](../coretypes).

## The VM

### Calling the VM
The entry point to the VM in the Wasp codebase is the function [MustRunVMTaskAsync](runvm/runtask.go#L20) function.  
It is called each time the Wasp node needs to run computations.

The function _MustRunVMTaskAsync_ start a goroutine to run calculations. It takes as a parameter  
 [vm.VMTask](taskcontext.go#L19):
```go
type VMTask struct {
	Processors         *processors.ProcessorCache
	ChainInput         *ledgerstate.AliasOutput
	VirtualState       state.VirtualState 
	Requests           []coretypes.Request
	Timestamp          time.Time
	Entropy            hashing.HashValue
	ValidatorFeeTarget coretypes.AgentID
	Log                *logger.Logger
	// call when finished
	OnFinish func(callResult dict.Dict, callError error, vmError error)
	// result
	ResultTransaction *ledgerstate.TransactionEssence
	ResultBlock       state.Block
}
```
The most important input parameters are:
```
	Requests           []coretypes.Request
	VirtualState       state.VirtualState 
```
* _VirtualState_ represents current state of the chain. It is _virtual_ in a sense that all updates to that state produced
by smart contracts are accumulated in memory and only are written (committed) into the database upon confirmation of the block
* _Requests_ represents a batch of requests. Each _request_ carries call parameters
as well as attached tokens (digital assets). Each request is processed by a smart contract it is targeting to.
So, requests update the  _virtual state_ sequentially, in each step producing new virtual state.

### Result of running the VM
The result of running the task by the VM consists of:
* final state of the _VirtualState_
* _ResultBlock_, a sequence of mutations to the _Virtual state_.
* _ResultTransaction_, an _essence_ part (unsigned yet) of the _anchor transaction_ which will be sent to the
Tangle ledger for confirmation.

The _VirtualState_ at the output of the task is always equal to the _VirtualState_ at the output with applied all
mutations contained in the _block_.

The _VirtualState_ has _state hash_ (Merkle root or similar) which is deterministically calculated from  
the initial _VirtualState_ and the resulting _block_.

The hash of the resulting _VirtualState_ is contained in the _ResultTransaction_, therefore upon confirmation
of the transaction on the Tangle ledger the virtual state is immutably anchored.

### Structure of the VM

The VM wraps many _processors_. On top of all processors, the VM wrapper implements  
fee logic and call between processors. In the generic (native) structure on _processor_  
repesents one smart contract. However, a processor may implemenent any deterministic calculations as long as  
it conforms to the _VMProcessor_ and other related interfaces.

![](VM.png)

 Significant part of the VM logic is implemented as _core smart contracts_. The core smart contracts also expose  
 core logic of each ISCP chain to outside users: the core smart contracts can be called by requests just like any other  
 smart contact.

 The implementation of core smart contracts is hardcoded into the Wasp. Implementations of all core contract as well  
 as their unit test can be found in [wasp/packages/vm/code](./core).

 Except core smart contracts, all other processors are plugged into the VM dynamically, hence _VM plugins_.

 All processors are alike: core contracts are attached to the VM just like any other VM plugin.

### Deployment of the smart contract
 The process of plugging a new smart contract (processor) into the VM is called _deployment_. The deployment  
 is handled by the _root_ contract by sending it the _deployContract_ request. As a result of the request,  
 new smart contract (a processor, VM plugin) is deployed on the chain. The registry of deployed smart contracts  
 is maintained by the _root_ contract as a part of the chain's state.

 The _deployContract_ request takes two parameters :
* _VM type_ parameter defines interpreter of the smart contract binary code
* _blob hash_ parameter is a hash of the binary which is loaded into the interpreter to create a _processor_.

### VM type
All _VM types_ are statically predefined in the Wasp node. It means, to implement a new type of VM plugin, you will need
to modify the Wasp node by adding a new VM type. However, the _VM type_ is part of VM abstraction therefore adding
new VM type is seemless.

The new VM Type is introduced to the rest of the VM abstraction logic through the call to the function  
[processors.RegisterVMType](processors/factory.go#L20).

The call to `processors.RegisterVMType` takes name of the new VM type and the constructor, a function which creates  
new `coretypes.Processor` object from the binary data of the program.

The following VM types are pre-defined in the current release of the Wasp:
* `builtinvm` represents core contracts
* `examplevm` represents example contracts which conforms to the native interface and are hardcoded before run
* `wasmtimevm` represents Wasmtime WebAssembly interpreter and native `Rust/Wasm` environment to create smart contracts.

To implement new types of interpreters, for example other languages or VM plugins based on EVM, new _VMTypes_  
must be implemented into the Wasp.

### Program binary and blob hash
To dynamically deploy smart contracts on the chain we need code of it in some binary format and dynamical linking of it to  
be able to call from VM. The very idea is to make the binary executable code of the smart contracts immutable,  
which means it must be part of the chain's state.

For example, `WebAssembly` smart contracts produced by the Rust/Wasm environment provided together with the Wasp,  
 are represented by `wasm` binaries. Other VM types may take different formats to represent it executable code.

To deploy a `wasmtimevm` smart contract on the chain, first we need to upload to the chain the corresponding `wasm` binary.  
All `wasm` binaries (as well as any other chunks of data) are kept in the registry handled by the `blob` core contact.  
To upload a blob to the chain you must send a request to the `blob`. Each blob on the chain has it hash and  
it is referenced by it.

The smart contract deployment takes VM type and binary blob hash as parameters. It makes the smart contract  
deployment process completely independent on the VM types and binary data formats of executables.

The only thing which is needed is to implement constructor function for the VM type and register it with `processors.RegisterVMType`.

## Processor and the sandbox interface

The processor implements one smart contract or any other deterministic executable, such as EVM.

In native and `wasmtimevm` implementations one processor represents one smart contract. It gives full power
to the smart contracts on the ISCP chain, such as manipulate native IOTA assets, call other smart contracts (processors)
on the same chain and send request to other ISCP chains.

Each processor implements a simple `coretypes.VMProcessor` [interface](../coretypes/vmprocessor.go#L15) and related  
`coretypes.VMProcessorEntryPoint` interface.

```go
type VMProcessor interface {
	GetEntryPoint(code Hname) (VMProcessorEntryPoint, bool)
	GetDefaultEntryPoint() VMProcessorEntryPoint 
	GetDescription() string
}

type VMProcessorEntryPoint interface {
	Call(ctx interface{}) (dict.Dict, error)
	IsView() bool
}
```

The smart contract is "plugged" into the VM with this interface.

### Entry points
A processor (smart contract) is a collection of callable entry points.

Each entry point is identified in the processor with its `hname` (hashed name), a 4 byte value, normally first 4 bytes
of `blake2b` hash of the smart contract's name of function signature.  
For more information see [coretypes.Hname](../coretypes/hname.go#L19).

Function `GetEntryPoint` returns entry point object with existence flag.

`GetDefaultEntryPoint` must always return default entry point. It will be called each time when entry point
with given `hname` is not found.

`VMProcessorEntryPoint` interface allows to call an entry point and passes it a context: a sandbox interface handler.  
The call returns a dictionary of result values, a collection of key/value pairs and, optionally, error code.

There are two types of entry points: _full entry points_ and _view entry points_.

* _full entry point_ only accepts call parameters of `coretypes.Sandbox` interface type. It provides full
access to the state for the smart contract so that the smart contract could modify it.
* _view entry point_ only accepts call parameters of `coretypes.SandboxView` interface type. It provides
limited _read-only_ access to the state.

The type of entry point is recognized by `IsView()` function. Using `Call()` with the wrong context type will result
panic in the VM.

Each new VM type has to provide its own processor and entry point implementations. The essential part is access to the  
chain state inside the entry point call. All the access is provided by the [Sandbox](../coretypes/sandbox.go)  
or [SandboxView](../coretypes/sandboxview.go) interfaces.

### Sandbox interface

The `coretypes.Sandbox` interface implements a number of functions which can be used by the processor's implementatio  
(smart contract). Here we will comment some of them:

* `Params()` returns a dictionary (key/value pairs) of the call parameters
* `State()` returns access to the `VirtualState` in the context of the call: a collection of key/value pairs.

 All key/value pair storages, like virtual state and call parameters are defined in the [kv package](../kv/kv.go#L19).  
 It defines different kinds of key/value pair collections and different encoding/decoding options.

* `Balances()` returns a collection color/value pairs: balances of colored tokens in the control of the
current smart contract
* `IncomingTransfer()` represent colored balances which are coming with this call. Those tokens are already part of `Balances()`
* `Call()` invokes entry point of another smart contract (processor) on the same chain.
Note that the call is agnostic about type of VM of the calling entry point.
* `Caller()` is a secure identification of the calling entity: and address or another smart contract
* `Send()` allows to send requests and funds to another chains, smart contracts or ordinary wallets.
* `Utils()` implements a number of utility function which may be called on the VM host. For example hashing or cryptography

### SandboxView interface
The view entry points are called from outside, for example by web server to query state if the chain and  
smart contracts. By intention those entry points cannot modify the state of the chain.

To view entry points the `SandboxView` interface is passed as a parameter. The `SandboxView` implements limited access  
to the state, fo example it doesn't have concept of `IncomingTransfer` and possibility of `Send()` tokens.  
The `State()` interface provided read-only access to the `VirtualState`.

The logic of the VM ensures that full entry points can call all other entry points, while view entry points can only call  
other view entry points.

## Implementation of EVM. A Virtual Ethereum

The ISCP team is pursuing a plan to implement EVM as one builtin smart contract which would run a forked version
of EVM. The EVM would be able to access key/value store through `State()` interface of the `Sandbox()`. This way EVM  
would run in and isolated environments and Solidity code won't be able to access and manipulate
native IOTA assets, hence Virtual Ethereum. To open EVM to access all spectrum of ISCP functions
would be the next step.

The Virtual Ethereum project is in the phase of definition therefore it is open for all kind of sugestions
of architectural decisions with the final goal in mind: to be able to run native EVM code (binary compatibility) as  
a VM on the ISCP chain.

The external interfaces of Virtual Ethereum would be wrapped into the native transactions and calls of IOTA and ISCP.


