---
keywords:
- Smart Contracts
- requests
- off-ledger
- node
- nonce
- tangle
- API calls
description:  An Off-ledger request is not a transaction, but it contains the same information as an on-ledger request, and it is cryptographically signed. This kind of requests do not rely on the Tangle for confirmation, so they are much faster.
image: /img/logo/WASP_logo_dark.png
---

# Off-ledger Requests

You can send `off-ledger` requests by sending an API call to a WASP node, which to the state of the target chain, an `access node` (which can be a committee node, or not). Unlike [`on-leger` requests](on-ledger-requests.md), the `off-ledger` request is not a transaction, it just contains the same information as an on-ledger request, and it is cryptographically signed. This kind of requests do not rely on the Tangle for confirmation, so they are much faster.

## Nonce

In order to [prevent replay attacks](../../../rfc/prevent-mev.md),  off-ledger requests must include a special parameter, the `nonce`.
Nonces are account-bound; the current nonce for a given account can be obtained via the [`accounts`](../core_contracts/accounts.md) core contract `getAccountNonce` view.

:::info Important
It is highly recommended you use **strictly monotonic increasing** nonces in off-ledger requests (i.e. 1,2,3,4,5).
:::

## Using the WASP Web API

After you have constructed an Off-ledger request, you can send it to a Wasp node webapi `/request/<chain_id>` endpoint via POST with the request as the body binary, or as a base64 string (MIME-type must be defined accordingly).
