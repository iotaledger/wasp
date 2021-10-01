---
keywords:
- ISCP
- Smart Contracts
- Rust
- JavaScript
- TypeScript
- WASM
description: An example game project with frontend and contract 
image: /img/logo/WASP_logo_dark.png
---

# Fair Roulette

The fair roulette is an example reference implementation demonstrating the development, setup and interaction with a smart contract.

## Introduction

For this example, we have created a very simple betting game in which players can bet on a number of a certain range. 

The goal of the game is to bet on the right number, to win a share of the placed funds. This is being done in rounds. 

A round is running for a certain amount of time. In our example: 60 seconds. In this timeframe, incoming bets will be added to a list of bets. After 60 seconds have passed, a winning number gets randomly generated and all players who made the right guess will receive a share of the pot.

If no round is active when a bet gets placed, the round gets initiated immediately.

## Mandatory setup

The mandatory setup consists out of:

* 1 GoShimmer node >= 0.7.5v
* 1 Beta Wasp node
* 1 static file server for the frontend

## Technicallity

Before we dive into the contents of the project, lets get an overview of what is actually required for this game.

### Fundamentals

Wasp is part of the IOTA ecosystem which enables the execution of smart contracts. These contracts run logic and are allowed to do state (change) requests towards the Tangle. To be able to store state, a respective GoShimmer node is required. It receives state change requests and if valid - saves them onto the Tangle. 

There are two ways to interact with smart contracts.

#### On Ledger requests

OnLedger requests are sent to GoShimmer nodes. Wasp periodically asks GoShimmer nodes for new OnLedger requests and handles them accordingly. These messages are validated through the network and take some time to be processed. 

#### Off Ledger requests

OffLedger requests are directly sent to Wasp nodes and do not require validation through GoShimmer nodes. They are therefore faster. However, they require an initial deposit of funds to a chain account as this account will initiate required OnLedger requests on the behalf of the desired contract or player.

> In our example we use OnLedger requests to initiate a betting request, a method to invoke OffLedger requests is implemented inside the frontend to make use of.

[.. cleverly link some docs we hopefully have about this stuff :D]

#### Funds

As these requests do cost some fees and to be able to actually bet with real token, the player requires a source of funds.  

Considering that the game runs on a testnet, funds can be requested from GoShimmer faucets inside the network[link faucet request]. 

After aquiring some funds, they reside inside an address which is handled by a wallet.

For this PoC, we have implemented our own very narrowed down wallet that runs inside the browser itself, mostly hidden from the player. 

In the future, we want to provide a solution that enables the use of Firefly or MetaMask(status?) as a secure external wallet.

#### Conclusion

To interact with a smart contract, we require a Wasp node which hosts the contract, a GoShimmer node to interact with the tangle, funds from a GoShimmer faucet, and a client that invokes the contract by either an On or Off Ledger request. 


### Implementation

The PoC consists out of two projects residing in `contracts/rust/fairroulette`.

One is the smart contract itself. Its boilerplate was generated using the new Schema tool[link] which is shipped with this beta release. 
The contract logic is written in Rust but the same implementation can be archived interchangebly with Golang.

The second project is an interactive frontend written in TypeScript, made reactive with the light Svelte framework. 
This frontend sends OnLedger requests to place bets towrds the fair roulette smart contract and makes use of the GoShimmer faucet to request funds.


### The smart contract 

> https://iscp.docs.iota.org/docs/tutorial/05/

As the smart contract is the only actor that is allowed to modify state in the context of the game, it got delegated a few tasks such as:

 * Validating and accepting placed bets
 * Starting and ending a betting round
 * Generation of a **random** winning number 
 * Sending payouts to the winners
 * Emitting status updates through the event system

Any incoming bet will be validated. This includes the amount of token which have been bet and also the number on which the player bet on. 
A number over 8 or under 1 will be rejected.

If the bet is valid and no round is active, the round state will be changed to `1` marking an active round. 
The bet will be the first of a list of bets. 

A delayed function call will be activated which executes **after 60 seconds**. 

This function is the payout function which generates a random winning number and pays out the winners of the round. 
The round state will be set to `0` indicating the end of the round.   

If the round is already active, the bet will be appended to the list of bets and await processing. 

State changes such as the `round started` or `round ended` but also placed bets and the payout of the winners are published as events. 
Events are published as messages through a public websocket.

#### Dependencies

* [wasm-pack](https://rustwasm.github.io/docs/wasm-pack/quickstart.html)

#### Building

```
cd contracts/rust/fairroulette
wasm-pack build 
```

### The frontend

The frontend has two main tasks. 

One is to visualize the contracts state. This includes a list of all placed bets, if a round is currently active and how long its still going. Any payouts will be shown as well, including a fancy animation in case the player itself has won. 

Furthermore the player can see his current available funds, his seed and his current address.

:::danger
As always, the seed is the key to your funds. As this is a PoC we display the seed for demonstrational purposes. 
**Never** share your seed with anyone under any circurmstance. 
:::

The second task is to enable the player to request funds and to participate in the game by placing bets. This is being done by showing the player a list of eight numbers, a selection of the amount of funds to bet, and a placing bet button. 

As faucet requests require minimal proof of work, the calculation happens inside a webworker to prevent a freezing of the browser ui.

To provide the frontend with the required events, it subscribes to the public websocket of Wasp to receive changes of state. 

Such events look like this:

`vmmsg kUUCRkhjD4EGAxv3kM77ay3KMbo51UhaEoCr14TeVHc5 df79d138: fairroulette.bet.placed 12sYqEZ5GM1BnqkZ88yJgPH3CdD9wKqfgGKY1j8FYDSZb3ao5wu 531819 2` 

This event displays a placed bet from the address `12sYqEZ5GM1BnqkZ88yJgPH3CdD9wKqfgGKY1j8FYDSZb3ao5wu`. This address bet `531819i` on the number `2`. This event originated from the smart contract id `df79d138`.

However, there is a bit more to the idea than to just subscribe to a websocket and "do requests".

### The communication layer

On and Off Ledger requests have a predefined structure. They need to get encoded in strict way and include a list of transactions provided by GoShimmer. Furthermore, they need to get signed by the client using the private key originating from a seed.  

Wasp is using the [ExtendedLockedOutput](https://goshimmer.docs.iota.org/docs/protocol_specification/components/advanced_outputs) message type, which enables certain additional properties such as: 

* fallback address and fallback timeout
* unlockable by AliasUnlockBlock (if address is of AliasAddress type)
* introducing a timelock (execution after deadline)
* data payload for arbitrary metadata (size limits apply)

This data payload is required to act on smart contracts as it contains: 

* the smart contract id to be selected
* the function id to be executed
* a list of arguments to be injected into the function 

As we don't expect developers of contracts and frontends to write their own implementation, we have separated the communication layer into two parts:

* The fairroulette_service
* The wasp_client

#### The wasp client

The wasp client is an example implementation of the communication protocol.

It provides:

* a basic wallet functionality
* hashing algorithms
* a webworker to provide proof of work 
* construction of On/Off Ledger requests
* construction of smart contract arguments and payload
* generation of seeds including their private keys and addresses
* serialization of data into binary messages
* deserialization of smart contract state

This wasp_client can be seen as a soon to be external library. For now this is a PoC client library shipped with the project. In the future however, we want to provide a library you can simply include into your project.

#### The fairroulette service

This service is meant to be a high level implementation of the actual app. In other words: It's the service that app or frontend developers would concentrate on. 

It does not construct message types, nor does it interact with GoShimmer directly. Besides of subscribing to the websocket event system of Wasp, it does not interact directly with Wasp either. Any such communication is handled by the `wasp_client`.

The fairroulette service is a mere wrapper around smart contract invocation calls. It accesses smart contract state through the `wasp_client` and does minimal decoding of data. 

Lets take a look into three parts of this service to make this more clear.

##### placeBetOnLedger

The [placeBetOnLedger](https://github.com/boxfish-studio/wasp/blob/feat/roulette_poc_ui/contracts/rust/fairroulette/frontend/src/lib/fairroulette_client/fair_roulette_service.ts#L144) function is responsible to send an On Ledger bet requests. It constructs a simple IOnLedger object  containing:

* the smart contract id: `fairroulette` 
* the function to invoke: `placeBet` 
* an argument: `-number` 
    * this is the number the player would bet on => the winning number  

Furthermore, this transaction requires an address to send the request to and also a variable amount of funds over `0i`.

:::note
For Wasp, the `chainId` is the address to use.  
:::

> Go into detail about the rest in a separate layer and hope we have documentation about some parts of this already.
> https://iscp.docs.iota.org/docs/misc/coretypes/, https://iscp.docs.iota.org/docs/misc/invoking/



#### Dependencies

* NodeJS >= 14
* NPM

#### Configuration

The frontend requires a config file to be created. The template can be copied from `contracts/rust/fairroulette/frontend/config.dev.sample.js` and renamed to `config.dev.js` inside the same folder.

Make sure to update the config values accordingly to your personal setup.
The `chainId` is the chainId which gets defined after deploying a chain (link #Deployment). 

`waspWebSocketUrl`, `waspApiUrl`, `goShimmerApiUrl`: Are dependent on the location of your Wasp and GoShimmer node. Make sure to keep the path of the `waspWebSocketUrl` (`/chain/%chainId/ws`)  at the end. 

`seed` can be either `undefined` or a predefined 44 length seed. If `seed` is set to `undefined` a new seed will be generated as soon a user opens the page. A predefined seed will be set for all users. This can be useful for development purposes. 

#### Building

```
cd contracts/rust/fairroulette/frontend
npm install
npm run build_worker
npm run dev
```

`npm run dev` will run a development server that exposes the transpiled frontend on `http://localhost:5000`.

## Deployment

> https://iscp.docs.iota.org/docs/misc/deploy/
> 

If the smart contract was built successfully, let's try to deploy the binary of the contract.

As deployments cost some fees, funds are required.

To validate the amount of funds run

`./wasp-cli balance`

If the output looks like this:

```
Address index 0
  Address: 17kfxps7zrxSi17sMGVwQt7NQwmdNYXjQXxkME2azX5Wb
  Balance:
    IOTA: 4525518
    ------
    Total: 4525518
```

It's possible to head directly to the deposit and deployment, otherwise request funds from the faucet first. `./wasp-cli request-funds`. 

---

Before we are able to deploy the contract, it's required to deploy a chain. 

In our example we deploy a chain to a single Wasp node, which equates into a committee of `0`

```
./wasp-cli chain deploy --committee=0 --quorum=1 --chain=roulette_chain --description="The fair roulette chain"
```

If this command was successful, we can deploy the contract binary afterwords. This deployment requires funds to be deposited to the **chain**. This is archived by running: 

`./wasp-cli chain deposit IOTA:10000`

Now it's possible to deploy the binary contract by executing:

`./wasp-cli chain deploy-contract wasmtime fairroulette "fairroulette"  contracts/rust/fairroulette/pkg/fairroulette_bg.wasm`
