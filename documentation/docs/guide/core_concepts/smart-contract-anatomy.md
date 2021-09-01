# Anatomy of a Smart Contract

Smart contracts are programs, immutably stored in the chain.

The logical structure of an ISCP smart contract is independent of the VM type we
use, be it a _Wasm_ smart contract or any other VM type.

![Smart Contract Structure](/img/tutorial/SC-structure.png)

## Identifying a Smart Contract

Each smart contract on the chain is identified by its name hashed into 4 bytes
and interpreted as `uint32` value: the `hname`.

For example, the `hname` of the root contract is `0xcebf5908`, the unique identifier of the
`root` contract in every chain. The exception is `_default` contract which always has hname `0x00000000`.

Each smart contract instance has a program with a collection of entry points and
a state. An entry point is a function of the program through which the program
can be invoked.

There are several ways to invoke an entry point: a request, a call and a view
call, depending on the type of the entry point.

The smart contract program can access its state and account through an interface
layer called the _Sandbox_.

## State

The smart contract state is its data, with each update stored on the chain. The
state can only be modified by the smart contract program itself. There are two
parts of the state:

- A collection of key/value pairs called the `data state`. Each key and value
  are byte arrays of arbitrary size (there are practical limits set by the
  database, of course). The value of the key/value pair is always retrieved by
  its key.
- A collection of `color: balance` pairs called the `account`. The account
  represents the balances of tokens of specific colors controlled by the smart
  contract. Receiving and spending tokens into/from the account means changing
  the account's balances.

Only the smart contract program can change its data state and spend from its
account. Tokens can be sent to the smart contract account by any other agent on
the ledger, be it a wallet with an address or another smart contract.

See [Accounts](../core_concepts/accounts/how-accounts-work.md) for more info on sending and receiving tokens.

## Entry points

There are two types of entry points:

- _Full entry points_ or just _entry points_. These functions can modify
  (mutate) the state of the smart contract.
- _View entry points_ or _views_. These are read-only functions. They are used
  only to retrieve the information from the smart contract state. They can’t
  modify the state, i.e. are read-only calls.

## Execution Results

After a request to a Smart Contract is executed (a call to a "full entry point"),
There will be a `receipt` added to the [`BlockLog`](../core_concepts/core_contracts/blocklog.md)
detailing the execution results of said request: whether it was successful, the block it was
included in, and other information. Any events dispatched by the Smart Contract in context of
this execution will also be added to the BlockLog.

## Error Handling

When a smart contract execution is interrupted for some reason (exception), or it produces an
error (missing parameter, or other inconsistency), the funds will be refunded to the caller,
except the fees. Any error that resulted from the SC execution can be viewed on the contract
`receipt` (present in the [`BlockLog`](../core_concepts/core_contracts/blocklog.md)).

<!-- 
// TODO this should be moved to the RUST SC example docs

The `example1` program has three entry points:

- `storeString` a full entry point. It first checks if parameter
  called `paramString` exist. If so, it stores the string value of the parameter
  into the state variable `storedString`. If parameter `paramString` is missing,
  the program panics.

- `getString` is a view entry point that returns the value of the
  variable `storedString`.

- `withdrawIota` is a full entry point that checks if the caller is and address
  and if the caller is equal to the creator of smart contract. If not, it
  panics. If it passes the validation, the program sends all the iotas contained
  in the smart contract's account to the caller.

Note that in the `example1` the Rust functions associated with full entry points
take a parameter of type `ScFuncContext`. It gives full (read-write) access to
the state. In contrast, `getString` is a view entry point and its associated
function parameter has type `ScViewContext`. A view is not allowed to mutate 
the state. -->
