---
keywords:
- Smart Contracts
- Chain
- Management
- Grafana
description:  How to manage a chain using the Grafana dashboard, a client to receive published events, logging and validators.
image: /img/logo/WASP_logo_dark.png
---
# Chain Management

## Monitoring

Each Wasp node provides a [Grafana dashboard](./running-a-node.md###grafana) where you can visualize metrics gathered by the node.

<!--
// TODO expand on this
-->

You can view the chain state using the [dashboard](./running-a-node.md###Dashboard) (by default on port `:7000`).

You can also set up a client to receive [published events](./running-a-node.md###Publisher) to have insights on what's happening in the system.

Lastly, each Wasp node will produce a log file (`wasp.log`) where the behaviour of a node can be investigated.

## Managing Chain Configuration and Validators

You can manage the chain configuration and committee of validators by interacting with the [Governance contract](../core_concepts/core_contracts/governance.md).
