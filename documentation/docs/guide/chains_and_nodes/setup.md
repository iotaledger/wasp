# Setting up a chain

Note: it is possible to run a "committee" composed of a single Wasp node, and
this may be fine for testing purposes. However, in normal operation the idea is
to have multiple Wasp nodes in order to run the smart contracts in a
distributed fashion. If you want to run a committee of several nodes on the
same machine, ensure that each Wasp instance runs in separate directory with
its own `config.json` and database. Ports and other settings must be adjusted
accordingly.

Also, for testing purposes, all Wasp nodes can be connected to the same
Goshimmer instance.  In normal operation, it is recommended for each Wasp node
to connect to a different Goshimmer instance.

## Requirements

- a number of many wasp nodes running with access to the same L1 network (pollen).
- wasp-cli configured to interact with the wasp nodes

## Trust Setup

After starting all the `wasp` nodes, one should make them trust each other.
Operators of the nodes should do that manually. That's their responsibility to
accept trusted nodes only.

The operator can read its node's public key and NetID by running `wasp-cli peering info`, e.g.:

```shell
$ wasp-cli peering info
PubKey: 8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM
NetID:  127.0.0.1:4000
```

PubKey and NetID should be provided to other node operators.
They can use this info to trust your node and accept communications with it.
That's done by invoking `wasp-cli peering trust <PubKey> <NetID>`, e.g.:

```shell
$ wasp-cli peering list-trusted
$ wasp-cli peering trust 8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM 127.0.0.1:4000
$ wasp-cli peering list-trusted
------                                        -----
PubKey                                        NetID
------                                        -----
8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM  127.0.0.1:4000
```

All the nodes in a committee must trust each other to run the chain.

## Starting the ChainId

### Requesting test funds

(If you're using a seed tha that already holds fund, you can skip this step)

```shell
wasp-cli request-funds
```

### Deploy the ISCP chain

```shell
wasp-cli chain deploy --committee=0,1,2,3 --quorum=3 --chain=mychain --description="My chain"
```

The indices in `--committee=0,1,2,3` will correspond to `wasp.0`, `wasp.1`,etc in `wasp-cli.json`.

The `--chain=mychain` sets up an alias for the chain. From now on all chain commands will be targeted to this chain.

## Testing if it works

You can check that the chain was properly deployed in the Wasp node dashboard
(e.g. `127.0.0.1:7000`). Note that the chain was deployed with some [core contracts](../core_concepts/core-contracts.md).

### Deploying a wasm contract

It's now possible deploy a Wasm contract to the chain:

```shell
wasp-cli chain deploy-contract wasmtimevm inccounter "inccounter SC" tools/cluster/tests/wasm/inccounter_bg.wasm
```

The `inccounter_bg.wasm` file is a precompiled Wasm contract included in the Wasp repo as an example.

Check again in the dashboard that the `inccounter` contract is listed in the chain.

### Interacting with a Smart contract

We can interact with a contract by calling its exposed functions and views.

For instance, the
[`inccounter`](https://github.com/iotaledger/wasp/tree/master/contracts/rust/inccounter/src)
contract exposes the `increment` function, which simply increments a counter
stored in the state. Also we have the `getCounter` view that returns the
current value of the counter.

Let's call the `getCounter` view:

```shell
$ wasp-cli chain call-view inccounter getCounter | wasp-cli decode string counter int
counter: 0
```

Note: the part after `|` is necessary because the return value is encoded and
we need to know the _schema_ in order to decode it. The schema definition is in
its early stages and will likely change in the future.

Now, let's call the `increment` function:

```shell
wasp-cli chain post-request inccounter increment
```

After the request has been processed by the committee we should get a new
counter value after calling `getCounter`:

```shell
$ wasp-cli chain call-view inccounter getCounter | wasp-cli decode string counter int
counter: 1
```
