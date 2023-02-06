This page contains requirements to create a proper API documentation and client generation.

# uints (16, 32)

UInts are unsupported by the Swagger standard. It only knows signed integers and floats.

Usually openapi generators have defined the rule, that if the documented property contains a min value of at least 0, it is treated as an uint.
This allows generation of clients with proper uint typing. 

Therefore, all Uints of any size need to have a min(0) or min(1) documented. See: `models/chain.go` => `ChainInfoResponse`: `MaxBlobSize`

## uints in path parameters 

Paths like `/accounts/account/:id` are mostly documented with `.AddParamPath`. It automatically gets the proper type and documents it.

Except for uints. If the query path requires uints, it is required to use `.AddParamPathNested` instead. Those properties require a `min(0)` or `min(1)`.

See: `controllers/corecontracts/controller.go` at route `chains/:chainID/core/blocklog/blocks/:blockIndex` => `getBlockInfo`: `blockIndex`. 

Those properties need to be named the same way as the parameters in the route. The linter will complain about unused properties. 
Therefore, a `//nolint:unused` is required in these cases.

## uint 64

UInt64s are unsupported in JavaScript when consumed via JSON. Therefore, all uint64s are serialized as strings by the server. 

The documentation should point out, that these strings are to be treated as uint64s. 

See: `models/core_accounts.go` => `AccountNonceResponse`: `Nonce`

