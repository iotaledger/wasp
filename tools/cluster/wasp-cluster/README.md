# wasp-cluster

`wasp-cluster` is a tool that allows to easily run a cluster of Wasp nodes
in a single host, to experiment with smart contracts in a controlled environment.

**Note:** `wasp-cluster` is intended for **testing purposes**, and is *not*
suitable for running a cluster in a production environment.

## Before

Make sure you have all needed binaries compiled and installed in the system
path:

* `wasp` (Wasp server)
* `wasp-cli` (CLI client for the Wasp node)
* `wasp-cluster` (this tool)

You can find instructions in
the [main README file](../../../readme.md#Prerequisites).

## Initialize the cluster configuration

```
wasp-cluster init my-cluster
```

This will create a directory named `my-cluster`, containing the cluster
configuration file (`cluster.json`) and one subdirectory for each node.

```
my-cluster/
├── cluster.json
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

If you need to change the default configuration of the nodes, you can do so now,
by editing the `config.json` files.

Note: by default `wasp-cluster` configures all nodes to store the database in
main memory: all data will be lost when the cluster is stopped (remember that
this tool is used primarily for testing). If you need a persistent database,
change the `db.engine` setting in all `config.json` files.

## Start the cluster

```
cd my-cluster
wasp-cluster start
```

(Alternatively: `wasp-cluster start my-cluster`.)

When done using the cluster, press `Ctrl-C` to stop it.

## Connecting to an existing Goshimmer network

By default, `wasp-cluster` provides a mock Goshimmer node to operate a
simulated ledger without the need for consensus. This is good
for quick tests, but is far from how the ledger works in a production
environment.

To connect the Wasp cluster to a more realistic environment (e.g. to be able to
persist the ledger), you can use the `docker-network` tool available
in the Goshimmer repository in order to start a cluster of Goshimmer nodes.

Example steps:

1. Edit `<goshimmer>/tools/docker-network/docker-compose.yml` adding `txstream`
   to all lines with `--node.enablePlugins=...`. (Just add `,txstream` at the
   end of the line.)

2. Start a Goshimmer network of 2 nodes:

```
cd <goshimmer>/tools/docker-network
./run.sh 2 0
```

3. In another console, initialize a cluster of 4 Wasp nodes (`-n 4`) with no
   mock Goshimmer node (`-g`).

```
wasp-cluster init my-cluster -n 4 -g
```

4. Start the Wasp cluster:

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
