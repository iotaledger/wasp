---
keywords:
- ISCP
- requests
- off-ledger
description:  Off-ledger Requests
image: /img/logo/WASP_logo_dark.png
---

# Off-ledger Requests

The alternative way of sending requests is so-called `off-ledger` requests. Its an API call to a Wasp node, which has access
to the state of the target chain, an `access node` (which can be a committee node, or not).
The `off-ledger` request is not a transaction, it just contains the same information as an on-ledger request and its
cryptographically signed. These kind of request don't rely on the Tangle for confirmation so they are much faster.

## Nonce

In order to [prevent replay attacks](../../../rfc/prevent-mev.md), it is required for off-ledger requests to include a special parameter, the `nonce`.
Nonces are account-bound, the current nonce for a given account can be obtained via the [`accounts`](../core_contracts/accounts.md) core contract `getAccountNonce` view.

:::info Important
It's highly recommended to use **strictly monotonic increasing** nonces in off-ledger requests (i.e. 1,2,3,4,5).
:::

## Using the WASP Web API

Off-ledger requests, after constructed can be sent a Wasp node webapi `/request/<chain_id>` endpoint via POST with the request as the body binary, or as a base64 string (MIME-type must be defined accordingly).
