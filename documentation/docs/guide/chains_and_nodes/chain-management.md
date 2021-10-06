---
keywords:
- ISCP
- Smart Contracts
- Chain
- Management
- Grafana
description:  How to manage a chain using the Grafana dashboard, a client to receive published events, logging and validators.
image: /img/logo/WASP_logo_dark.png
---
# Chain Management

## Monitoring

Each Wasp node provides a [Grafana dashboard](./running-a-node.md###grafana) where it's possible to visualize metrics gathered by the node.
// TODO expand on this

The chain state can be viewed via the [dashboard](./running-a-node.md###Dashboard) (by default on port `:7000`).

Setting up a client to receive [published events](./running-a-node.md###Publisher) can also be a good way to have insights on what's happening in the system.

Lastly, each Wasp node will produce a log file (`wasp.log`) where the behaviour of a node can be investigated.

## Managing Chain Configuration and Validators

Managing the chain configuration and committee of validators can be done by interacting with the [Governance contract](../core_concepts/core_contracts/governance.md).
