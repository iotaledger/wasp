# Core types

All core types used in the Wasp code are defined in the
[`IOTA Smart Contracts`](https://github.com/iotaledger/wasp/tree/master/packages/iscp)
package.


## Chain ID

IOTA Smart Contracts allows running multiple blockchains, called _smart contract chains_, _contract chains_ or just
_chains_ on the Tangle in parallel.

A chain is defined by two properties:

- chain *address* (type `address.Address` from GoShimmer, 33 bytes long)
- chain *color* (type `balance.Color` from GoShimmer, 32 bytes long)

Both address and color uniquely identify a chain. However, the chain address is
transient because chains can be moved from address to address. The chain color
is an ultimate identifier of the chain for its lifetime.

Each chain is identified on the IOTA Smart Contracts by its _chain ID_, represented by the
[`iscp.ChainID`](https://github.com/iotaledger/wasp/blob/master/packages/iscp/chainid/chainid.go)
type. In the current implementation `iscp.ChainID` is just a synonym of
the chain address. In the future, the chain color will be used as chain ID.


## Hashed names

The hashed values of string identifiers (called _hname_ for short) are used in
several places of the sandbox interface as type `iscp.Hame`.  The type is
alias of `uint32`.

The function `iscp.Hn(string) iscp.Hname` is used to compute the
hname of an identifier, by hashing the string with `blake2b` and returning the
first 4 bytes, cast to uint32 (little endian). (If all 4 bytes are _0x00_ or
_0xFF_, the function takes the next 4 bytes, since `0x00000000` and
`0xFFFFFFFF` are reserved values.)


## Smart contract ID

Each chain hosts multiple smart contracts. Each smart contract instance has a
_name_, a string value, assigned by the chain owner at deployment time.  The
_hname_ of the _name_ uniquely identifies the contract within a particular
chain.

The global identifier of the smart contract is represented by the type
`iscp.ContractID`.  The _contract ID_ is concatenation of the _chain ID_
and _hname_ of the contract (resulting in a 37 byte long value):

```
<contract ID> = <chain ID> || <contract hname>
```

The user-friendly representation of the contract ID is `<chain ID (base58)>::<hname (hex)>`.
For example: `2AxoLpidnriXtSif5NnXSWdt28fUb6VwVRjULdDoe6pZVw::cebf5908`.


## Agent ID

In the IOTA Tangle, iotas and colored tokens are owned by an address. Only the
entity owning the corresponding private key is able to spend from the address.

In IOTA Smart Contracts, iotas and colored tokens can be owned either by an address or by a
smart contract. In the latter case, only the contract represented by the
contract ID can spend those tokens, i.e. only the contract program can move
them to another location.

In order for contracts to be able to manipulate tokens, the tokens have to be
transferred to the chain address. The chain keeps a ledger to know who is the
actual owner of the tokens (address or contract), and allows the actual owner
to withdraw at any time (more on this [here](./accounts.md)).

An _agent ID_ (`iscp.AgentID` type, 37 bytes) represents an owner of
tokens in the internal chain ledger. It is a union type that can be either an
IOTA address (when the last 4 bytes are 0x00) or an IOTA Smart Contracts contract ID.

User-friendly representations of both types of _agent ID_ are prefixed by `A/`
(for addresses) and `C/` (for contracts), like this:

```
A/26yEvt3imFtcvK4NpgXK3499Rw2man3LTuvK2Mg4Rp8reZ
C/mZdSYhXd4F5qQGgELK8JvzUYovcNmPoVHxW1p4LF4gxT::cebf5908
```


## Colored balances

The interface `iscp.ColoredBalances` represents a map of color values and
their balances. The implementation is backed by the type `map[balance.Color]int64`.
It is guaranteed that the map does not contain colors with balance 0.

Example:

```
IOTA: 100000
mZdSYhXd4F5qQGgELK8JvzUYovcNmPoVHxW1p4LF4gxT: 1
```
