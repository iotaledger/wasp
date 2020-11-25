(in progress)
## Core types
All core types used in the Wasp code are defined in the `coretypes` [package](https://github.com/iotaledger/wasp/tree/chain-refactor/packages/coretypes).

## Chain ID

ISCP allows to run multiple _blockchains_, called _smart contract chains_, _contract chains_ or just 
_chains_ on the Tangle in parallel.

Each chain has two properties of it:
- chain address (represented by _address.Address_ type from Goshimmer, 33 bytes long)
- chain color (represented by _balance.Color_ type from Goshimmer, 32 bytes long)

Both _chain address_ and _chain color_ uniquely identify chain, however _chain address_ is transient because chains can 
be moved from address to address. The _chain color_ identifier is and ultimate identified of the chain for its
lifetime. 

Each chain is identified on the ISCP by _chain id_. It is represented by `coretypes.ChainID` [type](https://github.com/iotaledger/wasp/blob/8981a717b53a9c6790ae8204ba77a65a1d5add04/packages/coretypes/chainid.go#L13).

In the current implementation of the Wasp `coretypes.ChainID` is just a synonym of _chain address_. In the future _chain color_ will be
used as _chain ID_.

## Hashed names
The hashed vales of string identifiers (`hname`) are used in several places of sandbox interface as type `coretypes.Hame`.
The type is alias of the `uint32`.

The `coretypes.Hn(string) coretypes.Hname` function takes any string, hashes it with `blake2b` and takes 
first 4 bytes. If all 4 bytes are _0x00_ or _0xFF_, the function takes next 4 bytes.
 
The 4 bytes taken as `uint32` value (little endian) is the `hname`, the hashed value of the string. 

## Smart contract ID
Each chain hosts multiple smart contracts. Each smart contract has `name`, a string value, assigned by the programmer.
 
The `hname` of the _name_ of the smart contract uniquely identifies it within a particular chain. 
Note that _0x00000000_ and _0xFFFFFFFF_ are reserved values of `hname`.

The global identifier of the smart contract contract is represented by the type `coretypes.ContractID`. 
The _contract ID_ is concatenation of the _chain ID_ and _hname_ of the contract:
```
<contract ID> = <chain ID> || <contract hname>
```
The data type `coretypes.ContractID` is a 37 byte long value.

The user-friendly representation of the contract ID is string representation of _chain ID_ and contract's _hname_
separated by `::`, for example `2AxoLpidnriXtSif5NnXSWdt28fUb6VwVRjULdDoe6pZVw::2752361992`
 
## Agent ID
Iotas and colored tokens on the Tangle are owned by `addresses`. Only entity which owns corresponding
private key is able to spend from the _address_.

In ISCP we introduce an ownership of the iotas and colored tokens by a smart contract. 
Ownership of the tokens by some _contract ID_ means only the smart contract represented by the _contract ID_ 
can spend those tokens, i.e. only smart contract program can move them to another location. 

So on ISCP we have two kinds of location where tokens are securely stored: ordinary _addresses_ and _contract IDs_.

ISCP enables trustless transfers of tokens by the owner to any of those locations: be it _address_ or _contract ID_. 

The _agent ID_ encompasses both concepts into one polymorphic data type. 
It is represented by the `coretypes.AgentID` type (37 bytes long value),
From the value of _agent ID_ one can recognize, does it represent ordinary _address_ or _contract ID_.  
(if reserved `hname` value _0x00_ is in the _contract hname_ to represent _address_, otherwise it is a _contract ID_).

User friendly representations of both types of _agent ID_ are prefixed by `A-` (for addresses) and `C-` (for contracts), like this:
```
A-26yEvt3imFtcvK4NpgXK3499Rw2man3LTuvK2Mg4Rp8reZ
C-mZdSYhXd4F5qQGgELK8JvzUYovcNmPoVHxW1p4LF4gxT::3870572162
```

## Colored balances
_Colored balances_ is a type (interface) `coretypes.ColoredBalances` which represents a map of color values and their balances.
The interface is backed by the type `map[balance.Color]int64`.
It is guaranteed, that map does not contain colores with 0 balances.
For example:
```
IOTA: 100000
mZdSYhXd4F5qQGgELK8JvzUYovcNmPoVHxW1p4LF4gxT: 1
```

## On-chain accounts
ISCP introduces a concept of **on-chain account**. Each chain maintains a list of pairs: `<agent ID>: <colored balance>`.
Each pair is an account with its colored balances.

**Any _agent ID_ on the ISCP network may have an account on any chain**. 
In other words, any smart contract and any ordinary address on the network can have account on 
any chain with _colored balances_.

The ISCP ensures the tokens on the account on any chain may be moved to another location only but the entity, 
represented by the corresponding _agent ID_: be it _address_ or _contract ID_. 
The system require cryptographically secure authorisation to move funds between on-chain accounts.

It also means, for example:
- any entity may move its tokens seamlessly from address on the tangle to the account 
controlled by the same address on any chain
- anyone can send tokens to the account of any smart contract on any chain.
- _address_ can any time withdraw it tokens from chain to the address on the Tangle
- _smart contract_ may keep its funds on its native chain or on any other chain.

## How on-chain accounts work
Each chain, when deployed on the Tangle, contains several built (aka _core_) smart contracts on it. 
One of those _core_ contracts is `accountsc` contract. It handles the whole account machinery for each chain.

All funds, belonging to the smart contracts on some chain are contained in the _chain address_ on base layer (IOTA ledger)
of that particular chain. The `accountsc` smart contract maintains accounts for each _agent ID_ 
**in the state of the chain**, i.e. on L2.
 
Funds are moved to and from any on-chain accounts by calling `accountsc` functions on that chain.
The `accountsc` smart contract does all the ledger accounting and guarantees security.

In each call, the smart contract securely knows _agent ID_ of the originator of the call, the **caller**. 
The originator may be ordinary wallet or it can be another smart contract. 
The _caller ID_  is used by `accountsc` for checking the authorisation of the call: 
for example call to `withdraw` function will only be authorised
if called from the _agent ID_ (for example _address), which is equal to the owner of the account.

There are two main functions of the `accountsc` contract (among others):
- `deposit`. It allows by the caller (address wallet or smart contract) to deposit own funds on the chain
- `withdraw`. It allows to take back own funds from the on-chain account. For example for the caller 
from the wallet owned by ordinary address it means sending funds from on-chain 
account back to ordinary address.

By sending requests to the `accountsc` contract on a chain the sender (for example address) 
is in full control on its on-chain funds. Nobody else can move those funds because state of the chain
can be modified only by the smart contract under the consensus of the chain's committee of validators.

For the smart contract there are two functions implemented in the `Sandbox` interface 
(behind scenes it results in direct calls or on-tangle requests to the `accountsc` smart contract:
- `TransferToAddress` allows smart contract sending its funds to any address on the Tangle
- `TransferCrossChain` allows smart contract sending its funds to any on-chain account on any chain. 
  
## How secure are the on-chain accounts?

On-chain accounts are as secure as the chain they are residing on.
  
## Node fees 

The collection of node fees uses on-chain accounts follows this logic:
- if fees are enabled, they are accrued to the on-chain account of `ChainOwnerID`, the _agent ID_ which is controlled
by the owner of the chain. The owner of the chain can be an ordinary address on the Tangle or it can be a smart contract
on this or another chain.
- if fees are disabled the request tokens (1 mandatory token contained in each request) 
are always accrued to the on-chain account controlled by the sender of the request. 
The requester may withdraw it back at any time. If never withdrawn and deposited separately, the account will contain
number of iotas equal to the numebr of requests sent by that requester.

The on-chain accounts saves a lot of TPS and makes the whole system very flexible.



