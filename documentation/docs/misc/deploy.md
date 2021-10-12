# Using `wasp-cli` to deploy a chain and a contract

Once you have one or more Wasp nodes you can use the
[`wasp-cli`](https://github.com/iotaledger/wasp/tree/master/tools/wasp-cli) tool to interact with it.  Here is
an example set of commands that will deploy one chain and the example
[`inccounter`](https://github.com/iotaledger/wasp/tree/master/contracts/rust/inccounter/src)
contract to the chain.

---

First we need to tell `wasp-cli` the location of the Goshimmer node and the
committee of Wasp nodes:

```
$ wasp-cli set goshimmer.api 127.0.0.1:8080

$ wasp-cli set wasp.0.api 127.0.0.1:9090
$ wasp-cli set wasp.0.nanomsg 127.0.0.1:5550
$ wasp-cli set wasp.0.peering 127.0.0.1:4000

$ wasp-cli set wasp.1.api 127.0.0.1:9091
$ wasp-cli set wasp.1.nanomsg 127.0.0.1:5551
$ wasp-cli set wasp.1.peering 127.0.0.1:4001

...
```

The configuration is stored in `wasp-cli.json`, you can also edit the file by hand
instead of running the commands.

---

Next, we initialize a seed and request some funds from the faucet (we need at
least one token for each transaction; which can be [redeemed](./accounts.md) later).

```
$ wasp-cli init
$ wasp-cli request-funds
```

---

Now we can deploy a chain:

```
$ wasp-cli chain deploy --committee=0,1,2,3 --quorum=3 --chain=mychain --description="My chain"
```

The indices in `--committee=0,1,2,3` will correspond to `wasp.0`, `wasp.1`,
etc in `wasp-cli.json`.

The `--chain=mychain` sets up an alias for the chain. From now on all chain
commands will be targeted to this chain.

You can check that the chain was properly deployed in the Wasp node dashboard
(e.g. `127.0.0.1:7000`). Note that the chain was deployed with some [core contracts](../guide/core_concepts/core_contracts/overview.md)

---

We can now deploy a Wasm contract to our chain:

```
$ wasp-cli chain deploy-contract wasmtime inccounter "inccounter SC" tools/cluster/tests/wasm/inccounter_bg.wasm
```

The `inccounter_bg.wasm` file is a precompiled Wasm contract included as an
example.

Check again in the dashboard that the `inccounter` contract is listed in the chain.

---

We can interact with a contract by calling its exposed functions and views.

For instance, the
[`inccounter`](https://github.com/iotaledger/wasp/tree/master/contracts/rust/inccounter/src)
contract exposes the `increment` function, which simply increments a counter
stored in the state. Also we have the `getCounter` view that returns the
current value of the counter.

Let's call the `getCounter` view:

```
$ wasp-cli chain call-view inccounter getCounter | wasp-cli decode string counter int
counter: 0
```

Note: the part after `|` is necessary because the return value is encoded and
we need to know the _schema_ in order to decode it. The schema definition is in
its early stages and will likely change in the future.

Now, let's call the `increment` function:

```
$ wasp-cli chain post-request inccounter increment
```

After the request has been processed by the committee we should get a new
counter value after calling `getCounter`:

```
$ wasp-cli chain call-view inccounter getCounter | wasp-cli decode string counter int
counter: 1
```
