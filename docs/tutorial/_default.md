# The `_default` contract

The `_default` contract is one of the [core contracts](coresc.md) on each ISCP
chain.

The function of the `_default` contract is to provide a fall-back target for any
request that cannot be handled by the chain or contract it was addressed to.

It provides no entry points but scoops up any tokens that were passed as part of
the request and returns them to the caller (minus fees).
