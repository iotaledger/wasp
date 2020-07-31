# FairRoulette example Smart Contract client tool

Steps:

1. Start a Goshimmer network from the Goshimmer directory:

```
cd <goshimmer>/tools/docker-network
./run.sh 2
```

2. Install the `wasp`, `fairroulette` and `waspt` commands:

```
go install . ./tools/fairroulette ./tools/cluster/waspt
```

3. Start the Wasp cluster in a console:

```
$ cd tools/fairroulette/cluster
$ waspt init
$ waspt start
```

4. In another console, create a new wallet for the owner account:

```
fairroulette -c owner.json wallet init
```

This will create the file `owner.json` with the admin user's wallet.

5. Transfer some funds to the owner address: `fairroulette -c owner.json wallet request-funds`.

6. Initialize the FairRoulette smart contract:

```
$ fairroulette -c owner.json admin init
Initialized FairRoulette smart contract
SC Address: mUbfBM...
```

Copy the generated SC address. (It is also saved in `owner.json`)

7. Initialize a wallet for the client account:

```
fairroulette wallet init
```

This creates `fairroulette.json` (can be changed with `-c`).

8. Transfer some funds to your wallet: `fairroulette wallet request-funds`.

9. Query your balance:

```
$ fairroulette wallet balance
Index 0
  Address: WKos8N...
  Balances:
    10000 IOTA
```

10. Configure the SC address in `fairroulette.json` (obtained in step 7):

```
fairroulette set address mUbfBM...
```

11. Make a bet: `fairroulette bet 2 100`

`2` is the color to bet.

`100` is the amount of IOTAs to bet.

12. Query the SC state: `fairroulette status`

Also try `fairroulette dashboard`!

13. Change the play period (as admin): `fairroulette admin set-period 10 -c owner.json`
