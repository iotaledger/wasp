# wasp-cluster

`wasp-cluster` is a tool that allows to easily run a cluster of Wasp nodes
along with a Goshimmer node (needed for Wasp), all in a single host, to
experiment with smart contracts in a controlled environment.

Note: `wasp-cluster` is intended for testing purposes, and is *not* the
recommended way to run a cluster in a production environment.

## Before

1. Run `go install ./...` to compile and install Wasp binaries. This installs:

   * `wasp` (the Wasp server)
   * `wasp-cli` (CLI client for the Wasp node)
   * `wasp-cluster` (this tool)

2. Run `go install` in the Goshimmer repository (`wasp` branch) to install the
   `goshimmer` command.

## Initialize the cluster configuration

```
wasp-cluster init my-cluster
```

This will create a directory named `my-cluster`, containing the cluster
configuration file (`cluster.json`) and one subdirectory for each node.

```
my-cluster/
├── cluster.json
├── goshimmer
│   ├── config.json
│   └── snapshot.bin
├── wasp0
│   └── config.json
├── wasp1
│   └── config.json
├── wasp2
│   └── config.json
└── wasp3
    └── config.json
```

By default the cluster contains 4 Wasp nodes. You can change this with the
`-n` parameter. E.g.: `wasp-cluster init my-cluster -n 2`.

If you need to change the default configuration of the nodes, you can do so
now, by editing the `config.json` files.

Note: by default `wasp-cluster` configures all nodes to store the database in
main memory: all data will be lost when the cluster is stopped (remember that
this tool is used primarily for testing). If you need a persistent database,
change the `inMemory` setting in all `config.json` files.

## Start the cluster

```
cd my-cluster
wasp-cluster start
```

(Alternatively: `wasp-cluster start my-cluster`.)

When done using the cluster, press `Ctrl-C` to stop it.

## Connecting to an existing Goshimmer network

By default, the cluster includes a single Goshimmer node configured in such a
way that the ledger can be operated without the need for consensus. This is
good for quick tests, but is far from how Goshimmer works in a production
environment.

To connect the cluster to a more realistic environment (e.g. to be able to
persist the ledger), you can use the `docker-network-waspconn` tool available
in the Goshimmer repository (`wasp` branch) in order to operate a cluster of
Goshimmer nodes.

Example steps:

1. Start a Goshimmer network of 2 nodes:

```
cd <goshimmer>/tools/docker-network-waspconn
./run.sh 2
```

2. In another console, initialize a cluster of 4 Wasp nodes (`-n 4`) with no Goshimmer node (`-g`).

```
wasp-cluster init my-cluster -n 4 -g
```

3. Start the Wasp cluster:

```
$ cd my-cluster
$ wasp-cluster start
```

## Running a disposable cluster

If you just need to do a quick test, you can run a disposable cluster of nodes
with the default configuration with the command:

```
wasp-cluster start -d
```

No need to call `init` first; this command will automatically initialize the
cluster configuration in a temporary directory, which will be removed when the
cluster is stopped.
