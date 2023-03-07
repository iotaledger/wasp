# Local Setup

This Directory contains a Docker-based setup to run your own Wasp development
setup.

## Usage

### Starting

Run `docker-compose pull` to fetch the dependencies.

Run `docker-compose up` to start the setup.

After startup, you should be able to see the wasp dashboard on:
http://localhost/wasp/dashboard/

### Stopping

Press `Ctrl-C` to shut down the setup, but don't press it twice to force it.
Otherwise, you can corrupt the Hornet database.

You can also shut down the setup with `docker-compose down` in a new terminal.

### Reset

Run `docker-compose down --volumes` to shut down the nodes and to remove all
databases.

### Re-build

If you made changes to the Wasp code and want to use it inside the setup, you can re-build the Wasp image using `build_container.sh` or `build_container.cmd`.

## Ports

The nodes will then be reachable under these ports:

- Wasp:
  - API: <http://localhost:9090>

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
http://localhost:9090/chain/<CHAIN ID (tst1...)>/evm/jsonrpc
ChainID: 1074
```
