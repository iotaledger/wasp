# Run the cluster.

Build an image in `iotaledger/isc-private`:

```
docker build -t isc .
```

Then start the cluster

```bash
docker container prune
docker compose up
```

# Add nodes and login

If nodes were not added to the wasp-cli yet

```bash
wasp-cli wasp add wasp1 http://localhost:9091
wasp-cli wasp add wasp2 http://localhost:9092
wasp-cli wasp add wasp3 http://localhost:9093
wasp-cli wasp add wasp4 http://localhost:9094
```

Then login to each of the nodes:

```bash
wasp-cli login --node wasp1 # wasp:wasp
wasp-cli login --node wasp2
wasp-cli login --node wasp3
wasp-cli login --node wasp4
```


# Peering...

Get the node pub keys:

```bash
karolis@karolis-2020:~/temp/isc-20250212/committee$ wasp-cli --node=wasp1 peering list-trusted
----  ------                                                              ----------    -------
Name  PubKey                                                              PeeringURL    Trusted
----  ------                                                              ----------    -------
me    0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d  0.0.0.0:4000  true
karolis@karolis-2020:~/temp/isc-20250212/committee$ wasp-cli --node=wasp2 peering list-trusted
----  ------                                                              ----------    -------
Name  PubKey                                                              PeeringURL    Trusted
----  ------                                                              ----------    -------
me    0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2  0.0.0.0:4000  true
karolis@karolis-2020:~/temp/isc-20250212/committee$ wasp-cli --node=wasp3 peering list-trusted
----  ------                                                              ----------    -------
Name  PubKey                                                              PeeringURL    Trusted
----  ------                                                              ----------    -------
me    0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e  0.0.0.0:4000  true
karolis@karolis-2020:~/temp/isc-20250212/committee$ wasp-cli --node=wasp4 peering list-trusted
----  ------                                                              ----------    -------
Name  PubKey                                                              PeeringURL    Trusted
----  ------                                                              ----------    -------
me    0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4  0.0.0.0:4000  true

# ------------
# 1 0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d
# 2 0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2
# 3 0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e
# 4 0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4
```

```bash
wasp-cli --node=wasp2 peering trust wasp1 0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d 172.20.0.1:4000
wasp-cli --node=wasp3 peering trust wasp1 0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d 172.20.0.1:4000
wasp-cli --node=wasp4 peering trust wasp1 0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d 172.20.0.1:4000

wasp-cli --node=wasp1 peering trust wasp2 0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2 172.20.0.2:4000
wasp-cli --node=wasp3 peering trust wasp2 0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2 172.20.0.2:4000
wasp-cli --node=wasp4 peering trust wasp2 0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2 172.20.0.2:4000

wasp-cli --node=wasp1 peering trust wasp3 0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e 172.20.0.3:4000
wasp-cli --node=wasp2 peering trust wasp3 0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e 172.20.0.3:4000
wasp-cli --node=wasp4 peering trust wasp3 0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e 172.20.0.3:4000

wasp-cli --node=wasp1 peering trust wasp4 0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4 172.20.0.4:4000
wasp-cli --node=wasp2 peering trust wasp4 0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4 172.20.0.4:4000
wasp-cli --node=wasp3 peering trust wasp4 0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4 172.20.0.4:4000
```


# Works, but no need.

```bash
wasp-cli chain rundkg --node wasp1 --peers me,wasp2,wasp3,wasp4
```

```
# Address: 0x62e0bb46f943c196114e76d088deac0d22a1662dd3574c7151d1f584c0e38ee3
```


# Setup address to the L1.

Add l1 section in `~/.wasp-cli/wasp-cli.json`
```
  "l1": {
    "apiaddress": "https://api.iota-rebased-alphanet.iota.cafe",
    "faucetaddress": "https://faucet.iota-rebased-alphanet.iota.cafe/gas",
  }
```

# Get the coins.

```bash
wasp-cli request-funds
```

```
# Request funds for address 0x8a9a88968aa56ba1addceda4fa827d8c8450b4b9938b777af8636d9db96a37ac success
```

# Deploy the chain.

```bash
wasp-cli chain deploy --node=wasp1 --chain chain1 --peers=wasp2,wasp3,wasp4
```

```
NOTE: Adding this node as a committee member.
DKG successful
Address: 0x548301349f252162743491298215b91bd11eedb40a0fc6016f9187550f89e0cb
* committee size = 4
* quorum = 3
* members: 0xb43c7cf7c0fc6111273eb607ef2af26bc2a140e28d9b90b8d389bc806b3071e2 (wasp2)
0xc7ec30be24e24960370349b6efb95e04fc81193a116a523fe2594f9ea236073e (wasp3)
0x624a84ab6d22daee77658291b75ad0077a95be15d05b27664ad76b7624e3d4f4 (wasp4)
0x6621a1a10bfccc77f3eb9a386fa1c15e5c56350baf9cef94a3adf2ce2667ec7d ()

Creating new chain
* Owner address:    0x8a9a88968aa56ba1addceda4fa827d8c8450b4b9938b777af8636d9db96a37ac
* State controller: 0x548301349f252162743491298215b91bd11eedb40a0fc6016f9187550f89e0cb
* committee size = 5
* quorum = 0
Chain has been created successfully on the Tangle.
* ChainID: 0x104ce66c316e441c111da693eed06736417b04e022f219ddbc62346bc7341fd0
* State address: 0x548301349f252162743491298215b91bd11eedb40a0fc6016f9187550f89e0cb
* committee size = 5
* quorum = 0
Make sure to activate the chain on all committee nodes
Chain: 0x104ce66c316e441c111da693eed06736417b04e022f219ddbc62346bc7341fd0 (chain1)
Activated
```

# Activate it

```bash
wasp-cli chain list --node=wasp1
wasp-cli chain list --node=wasp2
wasp-cli chain list --node=wasp3
wasp-cli chain list --node=wasp4
```

```bash
wasp-cli chain activate --node=wasp2 --chain=chain1
wasp-cli chain activate --node=wasp3 --chain=chain1
wasp-cli chain activate --node=wasp4 --chain=chain1
# Fails with timeout, but actually works.
```

# Spam it

Run in different shells.

```
./spam.sh 0
./spam.sh 1
./spam.sh 2
./spam.sh 3
./spam.sh 4
./spam.sh 5
./spam.sh 6
```


# Rotation

```bash
wasp-cli chain list --node=wasp1
wasp-cli chain info --chain=chain1 --node=wasp1
# Chain ID: 0xad8cdc361e53abd681f9c6272197d4dac1ef6dee0becdcca4bcd4db34a700824
# EVM Chain ID: 1074
# Active: true
# State address: 0x94e510ee673b689a90e5ebf91be971d7867eaa2cfbd4fff01d314c37d7e7057f
#
# Committee nodes: 4
# ------                                                              ----------       -----  ---------  ------  ---------
# PubKey                                                              PeeringURL       Alive  Committee  Access  AccessAPI
# ------                                                              ----------       -----  ---------  ------  ---------
# 0x4e084179431dbee94cfd8be35124527280bd52b619247a4c68dc1bcdb5b67bda  0.0.0.0:4000     true   true       false
# 0x63ad49c3d5a0c66d8beedabffb2ac67aa78d6565420a558f0c520e85ae8ac254  172.20.0.2:4000  false  true       false
# 0x3b1ba3b3476c404dcab8dc71cfa58ab0f3acda8e25fb78fe97d2681d12e41078  172.20.0.3:4000  false  true       false
# 0x64d7daa288af64a84961b9b350a3e500bd37487aff1ff06c4f654a86e7307f88  172.20.0.4:4000  false  true       false
# ...
```

Create the new committee:

```bash
wasp-cli chain rundkg --node wasp1 --peers me,wasp2,wasp3,wasp4
# DKG successful
# Address: 0x919e37c9623b02253629b45d9ca5d6d12b6f0cf927ffa414422f4912d60249bd
# * committee size = 4
# * quorum = 3
# * members: 0x4e084179431dbee94cfd8be35124527280bd52b619247a4c68dc1bcdb5b67bda (me)
# 0x63ad49c3d5a0c66d8beedabffb2ac67aa78d6565420a558f0c520e85ae8ac254 (wasp2)
# 0x3b1ba3b3476c404dcab8dc71cfa58ab0f3acda8e25fb78fe97d2681d12e41078 (wasp3)
# 0x64d7daa288af64a84961b9b350a3e500bd37487aff1ff06c4f654a86e7307f88 (wasp4)
```

```bash
wasp-cli chain rotate 0x919e37c9623b02253629b45d9ca5d6d12b6f0cf927ffa414422f4912d60249bd --node=wasp1 --chain=chain1
wasp-cli chain rotate 0x919e37c9623b02253629b45d9ca5d6d12b6f0cf927ffa414422f4912d60249bd --node=wasp2 --chain=chain1
wasp-cli chain rotate 0x919e37c9623b02253629b45d9ca5d6d12b6f0cf927ffa414422f4912d60249bd --node=wasp3 --chain=chain1
wasp-cli chain rotate 0x919e37c9623b02253629b45d9ca5d6d12b6f0cf927ffa414422f4912d60249bd --node=wasp4 --chain=chain1
```

NOTE: Then post at least 1 request to the chain.
It is required to trigger the consensus, if no ongoing traffic exist.
And eventually check, if the committee has changed:

```bash
wasp-cli chain info --chain=chain1 --node=wasp1
```
