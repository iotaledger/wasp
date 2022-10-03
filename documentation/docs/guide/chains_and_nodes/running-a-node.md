---
description: How to run a node. Requirements, configuration parameters, dashboard configuration, and tests.
image: /img/logo/WASP_logo_dark.png
keywords:

- Smart Contracts
- Running a node
- Go-lang
- Hornet
- Requirements
- Configuration
- Dashboard
- Grafana
- Prometheus

---

# Running a Node

## Requirements

`wasp` and `wasp-cli` binaries [installed](./installing-wasp) in $PATH.

### Hardware

- **Cores**: At least 2 cores (most modern processors will suffice)
- **RAM**: 4GB

### Software

- Access to a [Hornet](https://github.com/iotaledger/hornet) node (with mqtt enabled) for
  production operation.

## Configuration

You can configure your node/s using the [`config.json`](https://github.com/iotaledger/wasp/blob/master/config.json).
The minimum needed configuration to run a wasp node is a L1 connection. For more advanced configuration
see [Node Configuration](./node-config).

You can set L1 access of the node following the instructions below.

### Hornet

Wasp requires a Hornet node to communicate with the L1 Tangle.

You can use any [publicly available node](https://wiki.iota.org/wasp/guide/chains_and_nodes/testnet),
or [set up your own node](https://wiki.iota.org/hornet/getting_started),
or [create a private tangle](https://wiki.iota.org/hornet/how_tos/private_tangle).

### Hornet Connection Settings

`l1.apiAddress` specifies the Hornet API address (default port: `14265`)

`li.faucetAddress` specifies the Hornet faucet address (default port: `8091`)

## Running the Node

After you have tweaked `config.json` to your liking, you can start a Wasp node by executing `wasp` and providing the path to your `config.json` file with `-c`. Not providing this flag will cause your `config.json` file to be ignored and wasp will be started with default configurations.

```shell
mkdir wasp-node
cp config.json wasp-node
cd wasp-node
#<edit config.json as desired>
wasp -c config.json 
```

You can verify that your node is running by opening the dashboard with a web browser
at [`127.0.0.1:7000`](http://127.0.0.1:7000) (default URL).

Repeat this process to launch as many nodes as you want for your committee.

### Accessing Your Node From a Remote Machine

To access the Wasp node from outside its local network, you must add your public IP to the `webpi.adminWhitelist`. You
can add it to your config file or run the node with the `webapi.adminWhitelist` flag.

```shell
wasp --webapi.adminWhitelist=127.0.0.1,YOUR_IP
```

## Video Tutorial

<iframe
width="560"
height="315"
src="https://www.youtube.com/embed/eV2AoV3QPC4"
title="Wasp Node Setup"
frameborder="0"
allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
allowfullscreen
/>
