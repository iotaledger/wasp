# Local Setup

This Directory contains a Docker-based setup to run your own Wasp development setup.

## Usage

### Starting

Run `docker-compose up` to start the setup.

### Stopping

Press `Ctrl-C` to shut down the setup, but don't press it twice to force it. Otherwise, you can corrupt the Hornet database.

You can also shut down the setup with `docker-compose down` in a new terminal.

### Reset

Run `docker-compose down --volumes` to shut down the nodes and to remove all databases.

### Recreation

If you made changes to the Wasp code and want to use it inside the setup, you need to recreate the Wasp image.

Run `docker-compose build`

## Ports

The nodes will then be reachable under these ports:

- Wasp:
  - API: <http://localhost:9090>

- Hornet:
  - API: <http://localhost:14265>
  - Faucet: <http://localhost:8091>
  - Dashboard: <http://localhost:8081> (username: admin, password: admin)

## Wasp-cli setup

To configure a new wasp-cli you can use the following commands:

:::note

You can either use a wasp-cli installed on your system, or use the one built-in to the wasp docker container by doing: `docker exec  wasp /app/wasp-cli init`

:::

```shell
wasp-cli init
wasp-cli set l1.apiaddress http://localhost:14265
wasp-cli set l1.faucetaddress http://localhost:8091
wasp-cli wasp add 0 http://localhost:9090
```

To create a chain:

```shell
wasp-cli request-funds
wasp-cli chain deploy --nodes=0 --quorum=1 --chain=testchain --description="Test Chain"
```

After a chain has been created, the EVM JSON-RPC can be accessed via:

```
http://localhost:9090/chain/<CHAIN ID (tst1...)>/evm/jsonrpc
ChainID: 1074
```
