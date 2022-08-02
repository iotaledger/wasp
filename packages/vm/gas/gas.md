# Gas Costs

Current gas costs are still experimental and will change.

| Instruction                        | Cost      | Description                                                          |
| ---------------------------------- | --------- | -------------------------------------------------------------------- |
| CallTargetNotFound                 | 10        | gas burned when a call target doesn't exist                          |
| GetContext                         | 10        | get context data                                                     |
| GetCallerData                      | 10        | get caller data                                                      |
| GetStateAnchorInfo                 | 10        | get state anchor data                                                |
| GetBalance                         | 20        | get balance of account on the chain                                  |
| BurnCodeGetNFTData                 | 10        | get data about the NFT (issuer/metadata)                             |
| CallContract                       | 10        | call a target (another SC in the same chain)                         |
| EmitEventFixed                     | 10        | emit event                                                           |
| GetAllowance                       | 10        | get allowance                                                        |
| TransferAllowance                  | 10        | transfer allowance                                                   |
| BurnCodeEstimateStorageDepositCost | 5         | estimate the storage deposit cost of a L1 request to be sent         |
| SendL1Request                      | 200*N     | send a L1 transaction (N = number of issued txs in the current call) |
| DeployContract                     | 10        | deploy a contract                                                    |
| Storage                            | 1*B       | storage (B = number of bytes)                                        |
| ReadFromState                      | 1*(B/100) | read from state (B = number of bytes, adjusted in the call)          |
| Wasm                               | X         | wasm code execution (X = gas returnted by WASM VM)                   |
| UtilsHashingBlake2b                | 5*B       | blake2b hash function (B = number of bytes)                          |
| UtilsHashingSha3                   | 8*B       | sha3 hash function (B = number of bytes)                             |
| UtilsHashingHname                  | 5*B       | get hname from string (hash function, B = number of bytes)           |
| UtilsBase58Encode                  | 50*B      | encode data into base58 (B = number of bytes)                        |
| UtilsBase58Decode                  | 5*B       | decode data from base58 (B = number of bytes)                        |
| UtilsED25519ValidSig               | 200       | validates a ed25517 signature                                        |
| UtilsED25519AddrFromPubKey         | 50        | get ed25517 address from public key                                  |
| UtilsBLSValidSignature             | 2000      | validates a bls signature valid (to be deprecated)                   |
| UtilsBLSAddrFromPubKey             | 50        | get bls address from public key (to be deprecated)                   |
| UtilsBLSAggregateBLS               | 400*B     | bls aggregate (to be deprecated, B = number of bytes)                |
