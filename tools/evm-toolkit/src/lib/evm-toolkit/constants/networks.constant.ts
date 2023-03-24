import type { INetwork } from '../interfaces';

export const NETWORKS: INetwork[] = [
  {
    "id": 0,
    "text": "Shimmer Beta Network",
    "apiEndpoint": "https://api.hornet.sc.testnet.shimmer.network",
    "faucetEndpoint": "https://faucet.testnet.shimmer.network",
    "chainAddress": "rms1prwgvvw472spqusqeufvlmp8xdpyxtrnmvt26jnuk6sxdcq2hk8scku26h7",
    "chainID": 1071,
    "multicallAddress": "0x259BeDa0D834ac60BCe35a27184F68e4A92b8382"
  },
  {
    "id": 1,
    "text": "Custom Network",
    "apiEndpoint": "http://localhost:14265",
    "faucetEndpoint": "http://localhost:8091",
    "chainAddress": "tst1pz0kc4urqxf4vh6c556gnr6uwmkwz5ye98vucakgafyneed4gft322s3gf0",
    "chainID": 1071,
    "multicallAddress": "0x259BeDa0D834ac60BCe35a27184F68e4A92b8382"
  }
]
