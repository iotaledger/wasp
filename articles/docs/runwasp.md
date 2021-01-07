# How to run a Wasp node on Pollen

Here we describe step by step instructions how to run Wasp nodes on the Pollen network.

## Run Goshimmer with WaspConn

To run a Wasp node you need a Goshimmer instance with the
[WaspConn](https://github.com/iotaledger/goshimmer/tree/wasp/dapps/waspconn)
plugin. This version of Goshimmer is located in the
[`wasp` branch of the Goshimmer repository](https://github.com/iotaledger/goshimmer/tree/wasp).

The only difference between standard Goshimmer (the `develop` branch) and the
`wasp` branch is the `WaspConn` plugin, which accepts connections from Wasp
nodes.

First, clone and compile the Goshimmer version from the `wasp` branch.

```
$ git clone -b wasp https://github.com/iotaledger/goshimmer.git
$ cd goshimmer
$ go install  
```

Next, follow [these instructions](https://github.com/iotaledger/goshimmer/wiki/Setup-up-a-GoShimmer-node-(Joining-the-pollen-testnet))
to configure and run the goshimmer node connected to the Pollen network.

Note: by default the WaspConn plugin will be listening for Wasp connections on port `5000`.
To change this setting include the following section in `config.json`:

```
"waspconn": {
  "port": 12345
}
```

## Run Wasp

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

Clone the `develop` branch:

```
$ git clone -b develop https://github.com/iotaledger/wasp.git
```

Compile and install Wasp binaries:

```
$ cd wasp
$ go install ./...
```

Create an empty directory, copy the [`config.json`](https://github.com/iotaledger/wasp/blob/develop/config.json)
file, and change it as needed:

```
$ mkdir wasp-instance
$ cp config.json wasp-instance
$ <edit wasp-instance/config.json>
```

Finally, start the Wasp node:

```
$ cd wasp-instance
$ wasp
```

### Wasp settings

Below we explain some settings in `config.json` you may need to adjust. You may
need to adjust ports especially if you plan to run several nodes in the same
host.

#### Peering

Wasp nodes connect to other Wasp peers to form committees. There is exactly one
TCP connection between two Wasp nodes participating in the same committee. Each
node uses the `peering.port` setting to specify the port for peering.

`peering.netid` must have the form `host:port`, with `port` equal to
`peering.port`, and where `host` must resolve to the machine where the node is
running, and must be reachable by other nodes in the committee. Each node in a
committee must have a unique `netid`.

#### Goshimmer connection settings

`nodeconn.address` specifies the Goshimmer host and port (exposed by the `WaspConn` plugin) to
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
