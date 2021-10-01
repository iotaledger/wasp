---
keywords:
- ISCP
- requests
- on-ledger
description:  On-ledger Requests
image: /img/logo/WASP_logo_dark.png
---

# On-ledger Requests

In this chapter we will describe sending requests to the smart contract as transactions on the Tangle.
It is so called `on-ledger` requests and can be sent to any chain deployed on the Tangle without accessing Wasp node,
just posting transaction to the Tangle. The ISCP committe is actively monitoring the Tangle and will be aware of the request.

![Generic process of posting an on-ledger request to the smart contract](/img/tutorial/send_request.png)

## Fees

Fees charged by the ISCP chain will be taken from this transaction, so be sure to send enough IOTAs to cover the fee, or the request will be rejected.

## Processing order

Requests are stored in the mempool and selected on a [random batch selection](../consensus.md), because of this order of execution is not guaranteed. If a user needs guaranteed sequence of processing, they should wait for a request to be processed before sending the following one.

## Fallback and Timelock options

By leveraging the `ExtendedLockedOutput`, it is possible to send requests with time constraints.

- fallback: by provinding a `deadline` timestamp and a `fallback address` its possible to post a request which: if not processed by `deadline`, it will not be picked-up by ISCP, and the funds will be spendable by the `fallback address`
- timelock: by providing a `timelock` timestamp its possible to post a request which won't be picked up until the the `timelock` timestamp is passed.