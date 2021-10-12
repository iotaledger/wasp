---
keywords:
- ISCP
- Smart Contracts
- Running a node
- Go-lang
- GoShimmer
- Requirements
- Configuration
- Dashborad
- Grafana
- Prometheus
description:  How to run a node. Requirements, configuration parameters, dashboard configuration and tests.
image: /img/logo/WASP_logo_dark.png
---

# Running a Node

In the following section we describe how to use Wasp by cloning the repository and building the application.
If you prefer, you can also configure a node [using a docker image](../../misc/docker.md) (official images will be provided in the future).

## Requirements

### Hardware

- **Cores**: At least 2 cores (most modern processors will suffice)
- **RAM**: 4GB

### Software

- [Go 1.16](https://golang.org/doc/install)
- [RocksDB](https://github.com/facebook/rocksdb/blob/master/INSTALL.md)
- Access to a [GoShimmer](https://github.com/iotaledger/goshimmer) node for
  production operation

:::info note 

The Wasp node requires the Goshimmer node to have the
[TXStream](https://github.com/iotaledger/goshimmer/tree/master/plugins/txstream)
plugin enabled. Being an experimental plugin, it is currently disabled by default and can
be enabled via configuration.

:::

### Microsoft Windows Installation Errors

If the go install command is telling you it cannot find gcc you will need to
install [MinGW-w64](https://sourceforge.net/projects/mingw-w64/). When you do
make sure to select *x86_64* architecture instead of the preselected *i686*
architecture. After the installation make sure to add this folder to your PATH variable:

```
C:\Program Files\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin
```

## Compile

- Build the `wasp` binary (Wasp node): `go build -tags rocksdb`
- Build the `wasp-cli` binary (CLI client): `go build -tags rocksdb ./tools/wasp-cli`

Alternatively, you can build and install everything with `go install -tags rocksdb ./...`

On Windows you will need to use `go install -tags rocksdb -buildmode=exe ./...` instead.

## Test

- Run all tests (including integration tests which may take several minutes): `go test -tags rocksdb -timeout 20m ./...`
- Run only unit tests: `go test -tags rocksdb -short ./...`

:::info Note

Integration tests require the `wasp` and `wasp-cli` commands
in the system path (i.e. you need to run `go install ./...` before running
tests).

:::


## Configuration

Below we explain some settings in `config.json` you may need to adjust. You
will need to adjust ports especially if you plan to run several nodes in the
same host.

### Peering

Wasp nodes connect to other Wasp peers to form committees. There is exactly one
TCP connection between two Wasp nodes participating in the same committee. Each
node uses the `peering.port` setting to specify the port that will be used for peering.

`peering.netid` must have the form `host:port`, with `port` equal to
`peering.port`, and where `host` must resolve to the machine where the node is
running, and must be reachable by other nodes in the committee. Each node in a
committee must have a unique `netid`.

### Goshimmer Connection Settings

`nodeconn.address` specifies the Goshimmer host and port (exposed by the
[TXStream](https://github.com/iotaledger/goshimmer/tree/master/plugins/txstream) plugin) to
connect to. You can find more information about the Goshimmer node in the [Goshimmer Provider section](#goshimmer-provider).

### Publisher

`nanomsg.port` specifies the port for the [Nanomsg](https://nanomsg.org/) event publisher. Wasp nodes
publish important events happening in smart contracts, such as state
transitions, incoming and processed requests and similar. Any Nanomsg client
can subscribe to these messages.

<details>
  <summary>More Information on Wasp and Nanomsg</summary>
  <div>
  
  Each Wasp node publishes important events via a [Nanomsg](https://nanomsg.org/) message stream
  (just like ZMQ is used in IRI). Possibly, in the future, [ZMQ](https://zeromq.org/) and [MQTT](https://mqtt.org/) publishers will be supported too.

  Any Nanomsg client can subscribe to the message stream. In Go you can use the
  `packages/subscribe` package provided in Wasp for this.

  The Publisher port can be configured in `config.json` with the `nanomsg.port`
  setting.

  The Message format is simply a string consisting of a space separated list of tokens; and the first token
  is the message type. Below is a list of all message types published by Wasp (you can search for
  `publisher.Publish` in the code to see the exact places where each message is published).

  |Message|Format|
  |:--- |:--- |
  |Chain record has been saved in the registry | `chainrec <chain ID> <color>` |
  |Chain committee has been activated|`active_committee <chain ID>`|
  |Chain committee dismissed|`dismissed_committee <chain ID>`|
  |A new SC request reached the node|`request_in <chain ID> <request tx ID> <request block index>`|
  |SC request has been processed (i.e. corresponding state update was confirmed)|`request_out <chain ID> <request tx ID> <request block index> <state index> <seq number in the block> <block size>`|
  |State transition (new state has been committed to DB)| `state <chain ID> <state index> <block size> <state tx ID> <state hash> <timestamp>`|
  |Event generated by a SC|`vmmsg <chain ID> <contract hname> ...`|

  </div>
</details>

### Web API

`webapi.bindAddress` specifies the bind address/port for the Web API, used by
`wasp-cli` and other clients to interact with the Wasp node.

### Dashboard

`dashboard.bindAddress` specifies the bind address/port for the node dashboard,
which can be accessed with a web browser.

### Prometheus

`prometheus.bindAddress` specifies the bind address/port for the prometheus server, where its possible to get multiple system metrics.
By default Prometheus is disabled and should be enabled by setting `prometheus.enabled` to `true`.

### Grafana

Grafana provides a dashboard to visualize system metrics, it can use the prometheus metrics as a data source.

## Goshimmer Provider

For the Wasp node to communicate with the L1 (Tangle/Goshimmer Network), it needs access to a Goshimmer node with the TXStream plugin enabled. You can use any publicly available node, or [set up your own node](https://wiki.iota.org/goshimmer/tutorials/setup/).

:::info note

By default, the TXStream plugin will be listening for Wasp connections on port `5000`.
To change this setting you can add the argument `--txstream.port: 12345`.

:::

## Running the Node

After `config.json` is tweaked as necessary you can simply start a Wasp node by executing `wasp` on the same directory.

```shell
mkdir wasp-node
cp config.json wasp-node
cd wasp-node
#<edit config.json as desired>
wasp
```

You can verify that your node is running by opening the dashboard with a web browser at `127.0.0.1:7000` (default url).

Repeat this process to launch as many nodes as you want for your committee.


## Exposing a Port

You can expose a port by 