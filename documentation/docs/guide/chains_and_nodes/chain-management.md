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

Each Wasp node provides a [Grafana dashboard](./running-a-node.md#grafana) where you can visualize metrics gathered by the node.

You can view the chain state using the [dashboard](./running-a-node.md#dashboard) (by default on port `:7000`).

You can also set up a client to receive [published events](./running-a-node.md#publisher) to have insights on what's happening in the system.

Lastly, each Wasp node will produce a log file (`wasp.log`) where the behaviour of a node can be investigated.

## Managing Chain Configuration and Validators

You can manage the chain configuration and committee of validators by interacting with the [Governance contract](../core_concepts/core_contracts/governance.md).

Administrative tasks can only be performed by the "Chain Owner".

### Changing chain ownership

In order to change the chain ownership, there must be a call to `delegateChainOwnership` by the current owner, specifying the agentID of the next owner, then the next owner must call `claimChainOwnership` to finalize the process.

### Changing Access nodes

For new access nodes to join the network they need to:

- be added as a trusted peer to at least 1 of the existing nodes
- be added by the administrator to the list of access nodes by calling `changeAccessNodes` (there is a helper for this in wasp-cli: `wasp-cli chain change-access-nodes`).

After this, new nodes should be able to sync the state and execute view queries (call view entrypoints).

To remove an access node, a call to `changeAccessNodes` should be enough.

### Changing the set of validators

This can be done in different ways, depending who the [governor address](https://wiki.iota.org/introduction/develop/explanations/ledger/alias) of the alias output of the chain is.

- If the chain governor address is the chain itself, or another chain, the rotation can be performed by calling `rotateStateController` after adding the next state controller via `addAllowedStateControllerAddress`.
- If the chain governor address is an regular user wallet, the rotation transaction can be issued by using wasp-cli: `wasp-cli chain rotate <new controller address>`
