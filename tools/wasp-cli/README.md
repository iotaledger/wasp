# Wasp Client tool

Wasp Client is a CLI tool for interacting with Goshimmer and Wasp, allowing to:

* Manipulate a Goshimmer wallet
* Interact with the FairRoulette smart contract

## Wallet

* Create a new wallet (creates `wasp-cli.json` which stores the seed):

```
wasp-cli init
```

* Show private key + public key + account address for index 0 (index optional, default 0):

```
wasp-cli address -i 0
```

* Query Goshimmer for account balance:

```
wasp-cli balance [-i index]
```

* Use Testnet Faucet to transfer some funds into the wallet address at index n:

```
wasp-cli request-funds [-i index]
```

## FairRoulette (a.k.a "fr") Smart Contract

Steps:

1. Create a new wallet for the owner account:

```
wasp-cli -c owner.json init
```

This will create the file `owner.json` with the admin user's wallet.

2. Transfer some funds to the owner address: `wasp-cli -c owner.json request-funds`.

3. Initialize the FairRoulette smart contract, and transfer some operating
   capital to it:

```
$ wasp-cli -c owner.json fr admin deploy
Initialized FairRoulette smart contract
SC Address: mUbfBM...
$ wasp-cli -c owner.json send-funds mUbfBM... IOTA 100
```

Copy the generated SC address. (It is also saved in `owner.json`)

4. Initialize a wallet for the client account:

```
wasp-cli init
```

This creates `wasp-cli.json` (can be changed with `-c`).

5. Transfer some funds to your wallet: `wasp-cli request-funds`.

6. Query your balance:

```
$ wasp-cli balance
Index 0
  Address: WKos8N...
  Balances:
    10000 IOTA
```

7. Configure the SC address in `wasp-cli.json` (obtained in step 7):

```
wasp-cli fr set address mUbfBM...
```

8. Make a bet: `wasp-cli fr bet 2 100`

`2` is the color to bet.

`100` is the amount of IOTAs to bet.

9. Query the SC state: `wasp-cli fr status`

Also try `wasp-cli dashboard`!

10. Change the play period (as admin): `wasp-cli fr admin set-period 10 -c owner.json`

## FairAuction (a.k.a "fa") Smart Contract

Steps:

1. Initialize wallets for the owner and client account, as needed (see the
   relevant steps in FairRoulette section.

2. Initialize the FairAuction smart contract, and transfer some operating
   capital to it:

```
$ wasp-cli -c owner.json fa admin deploy
Initialized FairAuction smart contract
SC Address: mUbfBM...
$ wasp-cli -c owner.json send-funds mUbfBM... IOTA 100
```

Copy the generated SC address. (It is also saved in `owner.json`)

3. Configure the SC address in `wasp-cli.json` (obtained in step 2):

```
wasp-cli fa set address mUbfBM...
```

4. Mint some tokens:

```
$ wasp-cli mint 10
Minted 10 tokens of color y72kGq...
```

Copy the color ID.

5. Start an auction for those tokens:

```
$ wasp-cli fa start-auction "My first auction" y72kGq... 10 100 1
```

Arguments are:  `start-auction <description> <color> <amount> <minimum-bid> <duration-in-minutes>`

6. Place a bid for an auction:

```
$ wasp-cli fa place-bid y72kGq... 110
```

7. Query the SC state: `wasp-cli fa status`

Also try `wasp-cli dashboard`!
