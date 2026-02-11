# Local Setup

This Directory contains a Docker-based setup to run your own Wasp development
setup.

## Usage

### Starting

Run `docker compose pull` to fetch the dependencies.

Run `docker compose up -d` to start the setup.

After startup, you should be able to see the wasp dashboard on:
http://localhost/wasp/dashboard/

### Stopping/Resuming

You can stop execution with `docker compose down`.

### Removing data

Run:
```
docker compose down -v
```

This removes the setup's volumes. You'll need to spin the setup up again to re-create them.

## Ports

The nodes will then be reachable under these ports:

- Wasp:
  - API: <http://localhost:9090>
  - DASHBOARD: <http://localhost/wasp/dashboard>

- IOTA:
  - Node RPC: <http://localhost:9000>
  - Faucet: <http://localhost:9123/gas>
  - GraphQL: <http://localhost:9125>

## Wasp-cli setup

Download the wasp cli from the [releases page](https://github.com/iotaledger/wasp/releases)

To configure a new wasp-cli you can use the following commands:

```shell
wasp-cli init
wasp-cli set l1.apiaddress http://localhost:9125
wasp-cli set l1.faucetaddress http://localhost:9123/gas
wasp-cli wasp add 0 http://localhost:9090
```

To create a chain:

```shell
wasp-cli request-funds
wasp-cli chain deploy --chain=testchain
```

After a chain has been created, the EVM JSON-RPC can be accessed via:

```
http://localhost/wasp/api/v1/chains/<CHAIN ID (tst1...)>/evm
ChainID: 1074
```

### Re-build (wasp-devs only)

If you made changes to the Wasp code and want to use it inside the setup, you can re-build the Wasp image using `build_container.sh` or `build_container.cmd`.
