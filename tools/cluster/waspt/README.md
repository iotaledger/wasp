# Waspt command

## Before

1. Run `go install` to compile and install the `wasp` command. Whenever
   you make a change in the wasp code you need to re-run this command.

2. Same thing goes for the `goshimmer` command: run `go install` in the
   goshimmer repository.

## Configuring a cluster

1. Create the configuration directory for the cluster (e.g. `my-cluster`)
2. Copy the `cluster.json`, `wasp-config-template.json` and
   `goshimmer-config-template.json` files from
   `tools/cluster/example`, and modify them with the new parameters.

## Initialize the cluster

	cd my-cluster
	waspt init

This creates `my-cluster/cluster-data` directory, which contains each
node's configuration and database.

## Create the SC addresses

	waspt gendksets

This creates an address and key shares for each node in the committee
for each smart contract.

The newly created keys are exported to `my-cluster/keys.json`.
Later, if the `cluster-data` directory (which contains the DBs) is
deleted, the `gendksets` step can be skipped, since the nodes will
automatically import keys from `keys.json`.

## Start the cluster

	waspt start
