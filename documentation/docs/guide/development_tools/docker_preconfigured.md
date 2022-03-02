---
description: How to run the preconfigured Docker setup.
image: /img/logo/WASP_logo_dark.png
keywords:
  - ISCP
  - Smart Contracts
  - Running a node
  - docker
  - image
  - build
  - configure
  - arguments
  - GoShimmer
---

# Preconfigured Development Docker setup

This page describes the usage of the preconfigured developer Docker setup.

## Introduction

To diminish the time spent on configuration and research, we have created a docker-compose setup that ships a pre-configured Wasp and GoShimmer (v0.7.7) node, that are connected with each other - ready to run out of the box.

## Running the setup

Checkout the project and start with docker-compose:

```shell
git clone https://github.com/iotaledger/wasp.git
cd tools/devnet
docker-compose up
```

Docker will build a lightly modified GoShimmer (v0.7.7) image and a Wasp image based on the contents of the checked out develop branch. If you do modifications inside the branch, docker-compose will include them into the Wasp image too.

## Usage

Wasp is configured to allow any connection coming from wasp-cli. This is fine for development purposes, but please make sure to not run it on a publicly available server, or to create matching firewall filter rules.

Besides this, everything should simply work as expected. Faucet requests will be handled accordingly, you will be able to deploy and run smart contracts. All useful ports such as:

- Wasp Dashboard (7000)
- Wasp API (9090)
- GoShimmer Dashboard (8081)
- GoShimmer API (8080)

are available to the local machine.

## Wasp-CLI configuration

As all ports are locally available, this `wasp-cli.json` configuration is to be used:

```json
{
  "goshimmer": {
    "api": "127.0.0.1:8080",
    "faucetpowtarget": -1
  },
  "wasp": {
    "0": {
      "api": "127.0.0.1:9090",
      "nanomsg": "127.0.0.1:5550",
      "peering": "127.0.0.1:4000"
    }
  }
}
```

Run `wasp-cli init` to generate a seed, and you are ready to go.

See [Configuring wasp-cli](/smart-contracts/guide/chains_and_nodes/wasp-cli) for further information.

## Notes about GoShimmer

As GoShimmer runs as a standalone node, it establishes no connection to other GoShimmer nodes. Running it in this way is unusual, but fine for development purposes. Warnings about Tangle Time not synced or similar can be ignored.

GoShimmer keeps the tangle tips inside memory only and will lose it after a restart. To recover these tips from the database, a fork was required and is to be found [here](https://github.com/lmoe/goshimmer). It is included in this package.
