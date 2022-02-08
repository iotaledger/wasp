# How to run a Wasp node

Here we describe step-by-step instructions how to run Wasp nodes on a Goshimmer network.

You will need the `wasp` and `wasp-cli` commands installed in the system, and
access to a Goshimmer node.

## Step 1: Compile & install Wasp

The `wasp` and `wasp-cli` commands can be installed from this repository:

```
$ git clone https://github.com/iotaledger/wasp.git
$ cd wasp
$ make install
```

### macOS arm64 (M1 Apple Silicon)

[`wasmtime-go`](https://github.com/bytecodealliance/wasmtime-go) hasn't supported macOS on arm64 yet, so you should build your own wasmtime library. You can follow the README in `wasmtime-go` to build the library.
Once a wasmtime library is built, then you can run the following commands.

```bash
$ git clone https://github.com/iotaledger/wasp.git
$ cd wasp
$ go mod edit -replace=github.com/bytecodealliance/wasmtime-go=<wasmtime-go path>
$ make install
```

## Step 2: Run Goshimmer

### Option 1: follow the official docs

Please find detailed instructions on how to run a Goshimmer node
[here](https://goshimmer.docs.iota.org/docs/tutorials/setup/).  This option
uses the official Goshimmer Docker image, so you will need Docker and Docker
Compose installed in your system.

:::info Important
The only change necessary from those instructions is to add some
plugins to `node.enablePlugins`:

- `txstream`: used by Wasp nodes to connect to the Goshimmer node
- `faucet`: required by the `wasp-cli request-funds` command
:::

### Option 2: compile & run Goshimmer

Alternatively, you can compile the `goshimmer` command and run it without
Docker.

```
$ git clone https://github.com/iotaledger/goshimmer.git
$ cd goshimmer
$ go install -tags rocksdb,builtin_static
```

Then, create an empty working directory for Goshimmer, and download the
`snapshot.bin` file needed for bootstrap:

```
$ mkdir goshimmer-node
$ cd goshimmer-node
$ wget -O snapshot.bin https://dbfiles-goshimmer.s3.eu-central-1.amazonaws.com/snapshots/nectar/snapshot-latest.bin
```

Start the GoShimmer node:

```
$ goshimmer \
        --skip-config=true \
        --analysis.client.serverAddress=ressims.iota.cafe:21888 \
        --autopeering.port=14626 \
        --dashboard.bindAddress=0.0.0.0:8081 \
        --gossip.port=14666 \
        --webapi.bindAddress=0.0.0.0:8080 \
        --profiling.bindAddress=0.0.0.0:6061 \
        --networkdelay.originPublicKey=9DB3j9cWYSuEEtkvanrzqkzCQMdH1FGv3TawJdVbDxkd \
        --fpc.bindAddress=0.0.0.0:10895 \
        --prometheus.bindAddress=0.0.0.0:9311 \
        --autopeering.entryNodes=2PV5487xMw5rasGBXXWeqSi4hLz7r19YBt8Y1TGAsQbj@ressims.iota.cafe:15626,5EDH4uY78EA6wrBkHHAVBWBMDt7EcksRq6pjzipoW15B@entryshimmer.tanglebay.com:14646 \
        --node.disablePlugins= \
        --node.enablePlugins=remotelog,networkdelay,spammer,prometheus,faucet,txstream \
        --faucet.seed=7R1itJx5hVuo9w9hjg5cwKFmek4HMSoBDgJZN8hKGxih \
        --logger.level=info \
        --logger.disableEvents=false \
        --logger.remotelog.serverAddress=ressims.iota.cafe:5213 \
        --drng.pollen.instanceId=1 \
        --drng.pollen.threshold=3 \
        --drng.pollen.committeeMembers=AheLpbhRs1XZsRF8t8VBwuyQh9mqPHXQvthV5rsHytDG,FZ28bSTidszUBn8TTCAT9X1nVMwFNnoYBmZ1xfafez2z,GT3UxryW4rA9RN9ojnMGmZgE2wP7psagQxgVdA4B9L1P,4pB5boPvvk2o5MbMySDhqsmC2CtUdXyotPPEpb7YQPD7,64wCsTZpmKjRVHtBKXiFojw7uw3GszumfvC4kHdWsHga \
        --drng.xteam.instanceId=1339 \
        --drng.xteam.threshold=4 \
        --drng.xteam.committeeMembers=GUdTwLDb6t6vZ7X5XzEnjFNDEVPteU7tVQ9nzKLfPjdo,68vNzBFE9HpmWLb2x4599AUUQNuimuhwn3XahTZZYUHt,Dc9n3JxYecaX3gpxVnWb4jS3KVz1K1SgSK1KpV1dzqT1,75g6r4tqGZhrgpDYZyZxVje1Qo54ezFYkCw94ELTLhPs,CN1XLXLHT9hv7fy3qNhpgNMD6uoHFkHtaNNKyNVCKybf,7SmttyqrKMkLo5NPYaiFoHs8LE6s7oCoWCQaZhui8m16,CypSmrHpTe3WQmCw54KP91F5gTmrQEL7EmTX38YStFXx
```

:::note
Argument values are adapted from the [official
instructions](https://goshimmer.docs.iota.org/docs/tutorials/setup/). You may
need to adjust them if they are outdated.
:::

:::tip
by default the TXStream plugin will be listening for Wasp connections on port `5000`.
To change this setting you can add the argument `--txstream.port: 12345`.
:::

## Run Wasp

:::note
It is possible to run a "committee" composed of a single Wasp node, and
this may be fine for testing purposes. However, in normal operation the idea is
to have multiple Wasp nodes in order to run the smart contracts in a
distributed fashion. If you want to run a committee of several nodes on the
same machine, ensure that each Wasp instance runs in separate directory with
its own `config.json` and database. Ports and other settings must be adjusted
accordingly.

Also, for testing purposes, all Wasp nodes can be connected to the same
Goshimmer instance.  In normal operation, it is recommended for each Wasp node
to connect to a different Goshimmer instance.
:::

Create an empty working directory for the Wasp node, copy the
[`config.json`](https://github.com/iotaledger/wasp/blob/master/config.json)
file, and change it as needed:

```
$ mkdir wasp-node
$ cp config.json wasp-node
$ <edit wasp-node/config.json>
```

Finally, start the Wasp node:

```
$ cd wasp-node
$ wasp
```

You can check that your node is running by opening the dashboard with a web
browser at `127.0.0.1:7000`.

Repeat this process to launch as many nodes as you want for your committee.

After starting all the `wasp` nodes, one should make them trust each other.
Operators of the nodes should do that manually. That's their responsibility to
accept trusted nodes only.

The operator can read its node's public key and NetID by running `wasp-cli peering info`, e.g.:

```
$ wasp-cli peering info
PubKey: 8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM
NetID:  127.0.0.1:4000
```

PubKey and NetID should be provided to other node operators.
They can use this info to trust your node and accept communications with it.
That's done by invoking `wasp-cli peering trust <PubKey> <NetID>`, e.g.:

```
$ wasp-cli peering list-trusted
$ wasp-cli peering trust 8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM 127.0.0.1:4000
$ wasp-cli peering list-trusted
------                                        -----
PubKey                                        NetID
------                                        -----
8oQ9xHWvfnShRxB22avvjbMyAumZ7EXKujuthqrzapNM  127.0.0.1:4000
```

All the nodes in a committee must trust each other to run the chain.

That's it!


### Wasp settings

Below we explain some settings in `config.json` you may need to adjust. You
will need to adjust ports especially if you plan to run several nodes in the
same host.

#### Peering

Wasp nodes connect to other Wasp peers to form committees. There is exactly one
TCP connection between two Wasp nodes participating in the same committee. Each
node uses the `peering.port` setting to specify the port for peering.

`peering.netid` must have the form `host:port`, with `port` equal to
`peering.port`, and where `host` must resolve to the machine where the node is
running, and must be reachable by other nodes in the committee. Each node in a
committee must have a unique `netid`.

#### Goshimmer connection settings

`nodeconn.address` specifies the Goshimmer host and port (exposed by the TXStream plugin) to
connect to.

#### Publisher

`nanomsg.port` specifies the port for the Nanomsg event publisher. Wasp nodes
publish important events happening in smart contracts, such as state
transitions, incoming and processed requests and similar.  Any Nanomsg client
can subscribe to these messages. More about the Publisher [here](./publisher.md).

#### Web API

`webapi.bindAddress` specifies the bind address/port for the Web API, used by
`wasp-cli` and other clients to interact with the Wasp node.

#### Dashboard

`dashboard.bindAddress` specifies the bind address/port for the node dashboard,
which can be accessed with a web browser.

## Now what?

Now that you have one or more Wasp nodes you can use the
[`wasp-cli`](https://github.com/iotaledger/wasp/tree/master/tools/wasp-cli) tool to [deploy a chain and smart
contracts](./deploy.md).
