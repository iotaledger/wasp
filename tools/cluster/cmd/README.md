    # Cluster command

## Before

1. Run `go install` to compile and install the `wasp` command. Whenever
   you make a change in the wasp code you need to re-run this command.

## Configuring a cluster

1. Create the configuration directory for the cluster (e.g. `my-cluster`)
1. Copy the `cluster.json` and `wasp-config-template.json` files from
   `tools/cluster/example`, and modify them with the new parameters.

## Initialize the cluster

1. `cd my-cluster`
1. `go run <root>/tools/cluster/cmd/main.go init`

This creates `my-cluster/cluster-data` directory, which contains each
node's configuration and database.

## Create the SC addresses

1. `go run <root>/tools/cluster/cmd/main.go gendksets`

This creates an address and key shares for each node in the committee
for each smart contract.

The newly created keys are exported to `my-cluster/keys.json`.
Later, if the `cluster-data` directory (which contains the DBs) is
deleted, the `gendksets` step can be skipped, since the nodes will
automatically import keys from `keys.json`.

## Create the origin TX for the SCs

1. `go run <root>/tools/cluster/cmd/main.go origintx`

