# Local Setup

This Directory contains a Docker-based setup to run your own Wasp development
setup.

## Usage

### Starting

Run `docker-compose pull` to fetch the dependencies.

Create dedicated volumes:
```
docker volume create --name hornet-nest-db
docker volume create --name wasp-db
```

Run `docker-compose up -d` to start the setup.

After startup, you should be able to see the wasp dashboard on:
http://localhost/wasp/dashboard/

### Stopping/Resuming

You can stop execution with `docker-compose down`.

### Removing data

After `docker compose down`:
```
docker volume rm wasp-db hornet-nest-db
```

You'll need to re-create the volumes to spin the setup up again.

## Ports

The nodes will then be reachable under these ports:

- Wasp:
  - API: <http://localhost:9090>
  - DASHBOARD: <http://localhost/wasp/dashboard>

- Hornet:
  - API: <http://localhost:14265>
  - Faucet: <http://localhost:8091>
  - Dashboard: <http://localhost:8081> (username: admin, password: admin)

## Wasp-cli setup

Download the wasp cli from the [releases page](https://github.com/iotaledger/wasp/releases)

To configure a new wasp-cli you can use the following commands:

```shell
wasp-cli init
wasp-cli set l1.apiaddress http://localhost:14265
wasp-cli set l1.faucetaddress http://localhost:8091
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
