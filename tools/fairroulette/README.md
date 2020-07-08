# FairRoulette example Smart Contract tools

Steps:

1. Install `goshimmer` command: `go install` in the Goshimmer directory

2. Install the `wasp` command: `go install`

3. Install the Wasp tools: `go install ./tools/...`

4. Start the FairRoulette cluster in a console:

```
$ cd tools/fairroulette/cluster
$ waspt init
$ waspt start
```

5. In another console, initialize the FairRoulette SC:

```
$ fr-admin init
Initialized FairRoulette smart contract
SC Address: mUbfBM...
```

Copy the generated SC address.

6. Initialize a wallet: `wallet init`.

This creates `wallet.json` with a new seed for addresses.

7. Transfer some IOTAs to your wallet: `wallet transfer 1 10000`.

`1` is the `utxodb` address index used as source for the funds.

`10000` is the amount of IOTAs to transfer.

8. Query your balance:

```
$ wallet balance
Index 0
  Address: WKos8N...
  Balances:
    10000 IOTA
```

9. Make a bet: `fr-client -sc <sc-address> bet 2 100`

Use the SC address copied in step 5.

`2` is the color to bet.

`100` is the amount of IOTAs to bet.

10. Query the SC state:

```
$ fr-client -sc <sc-address> status
FairRoulette Smart Contract State:
bets: 1
locked bets: 0
last winning color: 0
play period (s): 0
```
