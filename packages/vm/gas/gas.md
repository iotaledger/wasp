# Gas Costs

Current gas costs are still experimental and will change.

| Instruction                        | Cost      | Description                                                          |
| ---------------------------------- | --------- | -------------------------------------------------------------------- |
| CallTargetNotFound                 | 10        | gas burned when a call target doesn't exist                          |
| GetContext                         | 10        | get context data                                                     |
| GetCallerData                      | 10        | get caller data                                                      |
| GetStateAnchorInfo                 | 10        | get state anchor data                                                |
| GetBalance                         | 20        | get balance of account on the chain                                  |
| GetCoinInfo                        | 10        | get coin info                                                        |
| CallContract                       | 10        | call a SC target                                                     |
| EmitEvent                          | 1*B       | emit event (B = number of bytes)                                     |
| GetAllowance                       | 10        | get allowance                                                        |
| TransferAllowance                  | 10        | transfer allowance                                                   |
| SendL1Request                      | 200*N     | send a L1 transaction (N = number of issued txs in the current call) |
| Storage                            | 55*B      | storage (B = number of bytes)                                        |
| ReadFromState                      | 1*(B/100) | read from state (B = number of bytes, adjusted in the call)          |
| UtilsHashingBlake2b                | 50        | blake2b hash function (B = number of bytes)                          |
| UtilsHashingSha3                   | 80        | sha3 hash function (B = number of bytes)                             |
| UtilsHashingHname                  | 50        | get hname from string (hash function, B = number of bytes)           |
| UtilsHexEncode                     | 50*B      | encode data into hex (B = number of bytes)                           |
| UtilsHexDecode                     | 5*B       | decode data from hex (B = number of bytes)                           |
| UtilsED25519ValidSig               | 200       | validates a ed25517 signature                                        |
| UtilsED25519AddrFromPubKey         | 50        | get ed25517 address from public key                                  |
| UtilsBLSValidSignature             | 2000      | validates a bls signature valid (to be deprecated)                   |
| UtilsBLSAddrFromPubKey             | 50        | get bls address from public key (to be deprecated)                   |
| UtilsBLSAggregateBLS               | 400*B     | bls aggregate (to be deprecated, B = number of bytes)                |
| EVM                                | 1*B       | Burn gas from EVM                                                    |
