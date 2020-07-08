# FairRoulette example Smart Contract client tool

Steps:

1. Install `goshimmer` command: `go install` in the Goshimmer directory

2. Install the `wasp` command: `go install`

3. Install the `fairroulette` and `waspt` tools:

```
go install ./tools/waspt
go install ./tools/fairroulette
```

4. Start the Wasp cluster in a console:

```
$ cd tools/fairroulette/cluster
$ waspt init
$ waspt start
```

5. In another console, create a new wallet for the owner account:

```
fairroulette -c owner.json wallet init
```

This will create the file `owner.json` with the admin user's wallet.

6. Transfer some funds to the owner address: `fairroulette -c owner.json wallet transfer 1 10000`.

`1` is the `utxodb` address index used as source for the funds.

`10000` is the amount of IOTAs to transfer.

7. Initialize the FairRoulette smart contract:

```
$ fairroulette -c owner.json admin init
Initialized FairRoulette smart contract
SC Address: mUbfBM...
```

Copy the generated SC address. (It is also saved in `owner.json`)

8. Initialize a wallet for the client account:

```
fairroulette wallet init
```

This creates `fairroulette.json` (can be changed with `-c`).

9. Transfer some funds to your wallet: `fairroulette wallet transfer 1 10000`.

10. Query your balance:

```
$ fairroulette wallet balance
Index 0
  Address: WKos8N...
  Balances:
    10000 IOTA
```

11. Configure the SC address (obtained in step 7): `fairroulette set-address mUbfBM...`

12. Make a bet: `fairroulette bet 2 100`

`2` is the color to bet.

`100` is the amount of IOTAs to bet.

13. Query the SC state:

```
$ fairroulette status
FairRoulette Smart Contract status:
bets: 1
locked bets: 0
last winning color: 0
play period (s): 0
```
