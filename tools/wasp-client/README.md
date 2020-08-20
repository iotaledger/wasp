# Wasp Client tool

Wasp Client is a CLI tool for interacting with Goshimmer and Wasp, allowing to:

* Manipulate a Goshimmer wallet
* Interact with the FairRoulette smart contract

## Wallet

* Create a new wallet (creates `wasp-client.json` which stores the seed):

```
wasp-client wallet init
```

* Show private key + public key + account address for index 0 (index optional, default 0):

```
wasp-client wallet address -i 0
```

* Query Goshimmer for account balance:

```
wasp-client wallet balance [-i index]
```

* Use Testnet Faucet to transfer some funds into the wallet address at index n:

```
wasp-client wallet request-funds [-i index]
```

## FairRoulette (a.k.a "fr") Smart Contract

Steps:

1. Create a new wallet for the owner account:

```
wasp-client -c owner.json wallet init
```

This will create the file `owner.json` with the admin user's wallet.

2. Transfer some funds to the owner address: `wasp-client -c owner.json wallet request-funds`.

3. Initialize the FairRoulette smart contract:

```
$ wasp-client -c owner.json fr admin init
Initialized FairRoulette smart contract
SC Address: mUbfBM...
```

Copy the generated SC address. (It is also saved in `owner.json`)

4. Initialize a wallet for the client account:

```
wasp-client wallet init
```

This creates `wasp-client.json` (can be changed with `-c`).

5. Transfer some funds to your wallet: `wasp-client wallet request-funds`.

6. Query your balance:

```
$ wasp-client wallet balance
Index 0
  Address: WKos8N...
  Balances:
    10000 IOTA
```

7. Configure the SC address in `wasp-client.json` (obtained in step 7):

```
wasp-client fr set address mUbfBM...
```

8. Make a bet: `wasp-client fr bet 2 100`

`2` is the color to bet.

`100` is the amount of IOTAs to bet.

9. Query the SC state: `wasp-client fr status`

Also try `wasp-client dashboard`!

10. Change the play period (as admin): `wasp-client fr admin set-period 10 -c owner.json`

## FairAuction (a.k.a "fa") Smart Contract

Steps:

1. Initialize wallets for the owner and client account, as needed (see the
   relevant steps in FairRoulette section.

2. Initialize the FairAuction smart contract:

```
$ wasp-client -c owner.json fa admin init
Initialized FairAuction smart contract
SC Address: mUbfBM...
```

Copy the generated SC address. (It is also saved in `owner.json`)

3. Configure the SC address in `wasp-client.json` (obtained in step 2):

```
wasp-client fa set address mUbfBM...
```

4. Mint some tokens:

```
$ wasp-client wallet mint 10
Minted 10 tokens of color y72kGq...
```

Copy the color ID.

5. Start an auction for those tokens:

```
$ wasp-client fa start-auction "My first auction" y72kGq... 10 100 1
```

Arguments are:  `start-auction <description> <color> <amount> <minumum-bid> <duraion-in-minutes>`

6. Place a bid for an auction:

```
$ wasp-client fa place-bid y72kGq... 110
```

7. Query the SC state: `wasp-client fa status`

Also try `wasp-client dashboard`!
