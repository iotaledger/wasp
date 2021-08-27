# Chain Management

## Monitoring

Each wasp node provides a [Grafana dashboard](./run-node.md###grafana) where its possible to visualize metrics gathered by the node.
// TODO expand on this

The chain state can be viewed via the [dashboard](./run-node.md###Dashboard) (by default on port `:7000`).

Setting up a client to receive [published events](./run-node.md###Publisher) can also be a good way to have insights on what's happening in the system.

Lastly, each wasp node will produce a log file (`wasp.log`) where the behaviour of a node can be investigated.

## Managing chain configuration and validators

Managing the chain configuration and committe of validators can be done by interacting with the [Governance contract](../../contract_core/governance.md).
