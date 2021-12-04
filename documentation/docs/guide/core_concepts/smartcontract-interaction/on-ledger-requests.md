---
keywords:
- Smart Contracts
- requests
- on-ledger
- tangle
- fees
- fallback address
- timelock
description: Requests to the smart contract as transactions on the Tangle are called on-ledger requests.
image: /img/tutorial/send_request.png
---

# On-ledger Requests

Requests to the smart contract as transactions on the Tangle are called `on-ledger` requests. As they are simply posted by a transaction to the Tangle, they can be sent to any chain deployed on the Tangle without accessing a Wasp node. The IOTA Smart Contracts committee is actively monitoring the Tangle and will be aware of the request.

[![Generic process of posting an on-ledger request to the smart contract](/img/tutorial/send_request.png)](/img/tutorial/send_request.png)

## Fees

Fees charged by the IOTA Smart Contracts chain will be taken from this transaction, so be sure to send enough IOTAs to cover the fee, or the request will be rejected.

## Processing Order

Requests are stored in the mempool and selected on a [random batch selection](../consensus.md). Because of this, the order of execution is not guaranteed. If a user needs a guaranteed sequence of processing, they should wait for a request to be processed before sending the following one.

## Fallback and Timelock Options

By leveraging the `ExtendedLockedOutput`, it is possible to send requests with time constraints.

- **fallback**: By providing a `deadline` timestamp and a `fallback address` it is possible to post a request which, if not processed by the `deadline`, will not be picked-up by IOTA Smart Contracts, and the funds will be spendable by the `fallback address`.
- **timelock**: By providing a `timelock` timestamp it is possible to post a request which will not be picked up until the `timelock` timestamp is passed.