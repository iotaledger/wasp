---
description:  How to run a node. Requirements, configuration parameters, dashboard configuration and tests.
image: /img/logo/WASP_logo_dark.png
keywords:
- Smart Contracts
- Running a node
- Go-lang
- GoShimmer
- Requirements
- Configuration
- Dashboard
- Grafana
- Prometheus
---

# Running a Node

In the following section, you can find information on how to use Wasp by cloning the repository and building the application.
If you prefer, you can also configure a node [using a docker image (standalone)](docker_standalone.md) or [docker (preconfigured)](docker_preconfigured.md) (official images will be provided in the future).

## Requirements

### Hardware

- **Cores**: At least 2 cores (most modern processors will suffice)
- **RAM**: 4GB

### Software

- [Go 1.16](https://golang.org/doc/install)
- [RocksDB](https://github.com/facebook/rocksdb/blob/master/INSTALL.md)
- Access to a [GoShimmer](https://github.com/iotaledger/goshimmer) node for
  production operation.  

:::warning

GoShimmer is a developing prototype, so some things are prone to break. For a smooth development experience, you should use the GoShimmer code at [this commit](https://github.com/iotaledger/goshimmer/commit/25c827e8326a).

:::

:::info note 

The Wasp node requires the Goshimmer node to have the
[TXStream](https://github.com/iotaledger/goshimmer/tree/master/plugins/txstream)
plugin enabled. Being an experimental plugin, it is currently disabled by default and can
be enabled via configuration.

:::
## Download Wasp

You can get the source code of the latest Wasp version from the [official repository](https://github.com/iotaledger/wasp).

```shell
git clone https://github.com/iotaledger/wasp
```

## Compile

You can build and install both `wasp` and `wasp-cli` by running:

### Linux/macOS

```shell
make install
```

### macOS arm64 (M1 Apple Silicon)

[`wasmtime-go`](https://github.com/bytecodealliance/wasmtime-go) hasn't supported macOS on arm64 yet, so you should build your own wasmtime library. You can follow the README in `wasmtime-go` to build the library.
Once a wasmtime library is built, then you can run the following commands.

```shell
go mod edit -replace=github.com/bytecodealliance/wasmtime-go=<wasmtime-go path>
make install
```

### Microsoft Windows

```shell
make install-windows
```

### Add to Path

It is recommendable to add `wasp` and `wasp-cli` to your PATH. Please follow the instructions specific to your OS to do this.

#### Microsoft Windows Installation Errors

If the `make install-windows` command tells you it cannot find `gcc` you will need to
install [MinGW-w64](https://sourceforge.net/projects/mingw-w64/).Make sure
to select *x86_64* architecture instead of the preselected *i686*
architecture during the installation process. After the installation make sure to
add the following folder to your PATH variable:

```
C:\Program Files\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin
```

## Test

### Run All Tests

You can run integration and unit test together with the following command:

```shell
go test -tags rocksdb,builtin_static -timeout 20m ./...
```
Keep in mind that this process may take several minutes.

### Run Unit Tests

You can run the unit tests without running integration tests with the following command:

```shell
go test -tags rocksdb,builtin_static -short ./...
```

This will take significantly less time than [running all tests](#run-all-tests).

## Configuration

You can configure your node/s using the [`config.json`](https://github.com/iotaledger/wasp/blob/master/config.json)
configuration file.  If you plan to run several nodes in the same host, you will need to adjust the port configuration.

### Peering

Wasp nodes connect to other Wasp peers to form committees. There is exactly one
TCP connection between two Wasp nodes participating in the same committee. Each
node uses the `peering.port` setting to specify the port that will be used for peering.

`peering.netid` must have the form `host:port`, with a `port` value equal to
`peering.port`, and where `host` must resolve to the machine where the node is
running and be reachable by other nodes in the committee. Each node in a
committee must have a unique `netid`.

### Goshimmer Connection Settings

`nodeconn.address` specifies the Goshimmer host and port (exposed by the
[TXStream](https://github.com/iotaledger/goshimmer/tree/master/plugins/txstream) plugin) to
connect to. You can find more information about the Goshimmer node in the [Goshimmer Provider section](#goshimmer-provider).

### Publisher

`nanomsg.port` specifies the port for the [Nanomsg](https://nanomsg.org/) event publisher. Wasp nodes
publish important events happening in smart contracts, such as state
transitions, incoming and processed requests, and similar. Any Nanomsg client
can subscribe to these messages.

<details>
  <summary>More Information on Wasp and Nanomsg</summary>
  <div>
  
  Each Wasp node publishes important events via a [Nanomsg](https://nanomsg.org/) message stream
  (just like ZMQ is used in IRI). Possibly, in the future, [ZMQ](https://zeromq.org/) and [MQTT](https://mqtt.org/) publishers will be supported too.

  Any Nanomsg client can subscribe to the message stream. In Go, you can use the
  `packages/subscribe` package provided in Wasp for this.

  The Publisher port can be configured in `config.json` with the `nanomsg.port`
  setting.

  The Message format is simply a string consisting of a space-separated list of tokens, and the first token
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

`prometheus.bindAddress` specifies the bind address/port for the prometheus server, where it's possible to get multiple system metrics.

By default, Prometheus is disabled and should be enabled by setting `prometheus.enabled` to `true`.

### Grafana

Grafana provides a dashboard to visualize system metrics. It can use the prometheus metrics as a data source.

## Goshimmer Provider

For the Wasp node to communicate with the L1 (Tangle/Goshimmer Network), it needs access to a Goshimmer node with the TXStream plugin enabled. You can use any [publicly available node](https://wiki.iota.org/wasp/guide/chains_and_nodes/testnet), or [set up your own node](https://wiki.iota.org/goshimmer/tutorials/setup/).

:::info note

By default, the TXStream plugin will be listening for Wasp connections on port `5000`.

:::

### Default Configuration Changes

If you are using the default GoShimmer docker image you will need to [edit your GoShimmer docker-compose.yml](https://wiki.iota.org/goshimmer/tutorials/setup#define-the-docker-composeyml) file to include the following settings.

#### Add the TXStream plugin port

```yml
ports:
    [...]
      # TXStream plugin
      - "0.0.0.0:5000:5000/tcp"
    [...]
```

#### Enable TXStream and Faucet Plugins

```yml
[...]
--node.enablePlugins=remotelog,networkdelay,spammer,prometheus,txstream,faucet
[...]
```

## Running the Node

After you have tweaked `config.json` to your liking, you can start a Wasp node by executing `wasp` in the same directory
as shown in the following snippet.

```shell
mkdir wasp-node
cp config.json wasp-node
cd wasp-node
#<edit config.json as desired>
wasp
```

You can verify that your node is running by opening the dashboard with a web browser at [`127.0.0.1:7000`](http://127.0.0.1:7000) (default url).

Repeat this process to launch as many nodes as you want for your committee.

### Accessing Your Node From a Remote Machine 

If you want to access the Wasp node from outside its local network, you will need to add your public IP to the `webpi.adminWhitelist`. You can do so by adding it to your config file, or running the node with the `webapi.adminWhitelist` flag.

```shell
wasp --webapi.adminWhitelist=127.0.0.1,YOUR_IP
```

## Video Tutorial

<iframe width="560" height="315" src="https://www.youtube.com/embed/eV2AoV3QPC4" title="Wasp Node Setup" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
