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

A round is running for a certain amount of time. In our example: 60 seconds. In this timeframe, incoming bets will be added to a list of bets. After 60 seconds have passed, a winning number gets randomly generated and all players who made the right guess will receive a share of the pot depending on the amount of funds.

If no round is active when a bet gets placed, the round gets initiated immediately.

## Mandatory setup

The mandatory setup consists out of:

* 1 GoShimmer node >= 0.7.5v ([25c827e8326a](https://github.com/iotaledger/goshimmer/commit/25c827e8326a))
* 1 Beta Wasp node
* 1 Static file server (nginx, Apache, fasthttp)

## Technicality

Before we dive into the contents of the project, let's introduce important fundamentals first.

### Fundamentals

Wasp is part of the IOTA ecosystem which enables the execution of smart contracts. These contracts run logic and are allowed to do state (change) requests towards the Tangle. To be able to store state, a respective GoShimmer node is required. It receives state change requests and if valid - saves them onto the Tangle. 

There are two ways to interact with smart contracts.

#### On Ledger requests

See: [On-ledger Requests](/docs/guide/core_concepts/smartcontract-interaction/on-ledger-requests/)

On-ledger requests are sent to GoShimmer nodes. Wasp periodically asks GoShimmer nodes for new On-ledger requests and handles them accordingly. These messages are validated through the network and take some time to be processed. 

#### Off Ledger requests

See: [Off-ledger Requests](/docs/guide/core_concepts/smartcontract-interaction/off-ledger-requests/)

Off-ledger requests are directly sent to Wasp nodes and do not require validation through GoShimmer nodes. They are therefore faster. However, they require an initial deposit of funds to a chain account as this account will initiate required On-ledger requests on the behalf of the desired contract or player.

:::note
We use On-ledger requests in our example to initiate a betting request. A method to invoke Off-ledger requests is implemented inside the frontend to make use of.

See: [placeBetOffLedger](https://github.com/iotaledger/wasp/blob/roulette_poc/contracts/rust/fairroulette/frontend/src/lib/fairroulette_client/fair_roulette_service.ts#L133)
:::
#### Funds

As these requests do cost some fees and to be able to actually bet with real token, the player requires a source of funds.  

Considering that the game runs on a testnet, funds can be requested from GoShimmer faucets inside the network. 

See: [How to Obtain Tokens From the Faucet](https://goshimmer.docs.iota.org/docs/tutorials/obtain_tokens)

After acquiring some funds, they reside inside an address which is handled by a wallet.

For this PoC, we have implemented a very narrowed down wallet that runs inside the browser itself, mostly hidden from the player. 

In the future, we want to provide a solution that enables the use of Firefly or MetaMask as a secure external wallet.

#### Conclusion

To interact with a smart contract, we require a Wasp node which hosts the contract, a GoShimmer node to interact with the tangle, funds from a GoShimmer faucet, and a client that invokes the contract by either an On or Off Ledger request. In our example, the Frontend acts as the client.


### Implementation

The PoC consists out of two projects residing in `contracts/rust/fairroulette`.

One is the smart contract itself. Its boilerplate was generated using the new [Schema tool](/docs/guide/schema/intro) which is shipped with this beta release. 
The contract logic is written in Rust, but the same implementation can be archived interchangeably with Golang which is demonstrated in the root folder and `./src`.

The second project is an interactive frontend written in TypeScript, made reactive with the light Svelte framework and to be found in the subfolder `./frontend`.
This frontend sends On-ledger requests to place bets towards the fair roulette smart contract and makes use of the GoShimmer faucet to request funds.


### The Smart Contract 

See: [Structure of the smart contract](https://iscp.docs.iota.org/docs/tutorial/05)

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
Events are published as messages through a public web socket.

#### Dependencies

* [wasm-pack](https://rustwasm.github.io/docs/wasm-pack/quickstart.html)

#### Building

```
cd contracts/rust/fairroulette
wasm-pack build 
```

### The Frontend

The frontend has two main tasks. 

One is to visualize the contracts state. This includes a list of all placed bets, if a round is currently active and how long it's still going. Any payouts will be shown as well, including a fancy animation in case the player itself has won. 

Furthermore, the player can see his current available funds, his seed and his current address.

:::danger
As always, the seed is the key to your funds. We display the seed for demonstration purposes in this PoC. 
**Never** share your seed with anyone under any circumstance. 
:::

The second task is to enable the player to request funds and to participate in the game by placing bets. This is being done by showing the player a list of eight numbers, a selection of the amount of funds to bet, and a placing bet button. 

As faucet requests require minimal proof of work, the calculation happens inside a web worker to prevent a freezing of the browser UI.

To provide the frontend with the required events, it subscribes to the public web socket of Wasp to receive changes of state. 

Such events look like this:

`vmmsg kUUCRkhjD4EGAxv3kM77ay3KMbo51UhaEoCr14TeVHc5 df79d138: fairroulette.bet.placed 2sYqEZ5GM1BnqkZ88yJgPH3CdD9wKqfgGKY1j8FYDSZb3ao5wu 531819 2` 

This event displays a placed bet from the address `12sYqEZ5GM1BnqkZ88yJgPH3CdD9wKqfgGKY1j8FYDSZb3ao5wu`, a bet of `531819i` on the number `2`. Originating from the smart contract ID `df79d138`.

However, there is a bit more to the idea than to just subscribe to a web socket and "do requests".

### The Communication Layer

On and Off Ledger requests have a predefined structure. They need to get encoded in strict way and include a list of transactions provided by Shimmer. Furthermore, they need to get signed by the client using the private key originating from a seed.  

Wasp is using the [ExtendedLockedOutput](https://goshimmer.docs.iota.org/docs/protocol_specification/components/advanced_outputs) message type, which enables certain additional properties such as: 

* Fallback address and fallback timeout
* Unlockable by AliasUnlockBlock (if address is of Misaddress type)
* Introducing a time lock (execution after deadline)
* Data payload for arbitrary metadata (size limits apply)

This data payload is required to act on smart contracts as it contains: 

* the smart contract ID to be selected
* the function ID to be executed
* a list of arguments to be passed into the function 

As we don't expect developers of contracts and frontends to write their own implementation, we have separated the communication layer into two parts:

* The fairroulette_service
* The wasp_client

#### The Wasp Client

The wasp client is an example implementation of the communication protocol.

It provides:

* A basic wallet functionality
* Hashing algorithms
* A web worker to provide proof of work 
* Construction of On/Off Ledger requests
* Construction of smart contract arguments and payload
* Generation of seeds including their private keys and addresses
* Serialization of data into binary messages
* Deserialization of smart contract state

This wasp_client can be seen as a soon-to-be external library. For now this is a PoC client library shipped with the project. In the future however, we want to provide a library you can simply include into your project.

#### The Fairroulette Service

This service is meant to be a high level implementation of the actual app. In other words: It's the service that app or frontend developers would concentrate on. 

It does not construct message types, nor does it interact with GoShimmer directly. Besides of subscribing to the web socket event system of Wasp, it does not interact directly with Wasp either. Any such communication is handled by the `wasp_client`.

The fairroulette service is a mere wrapper around smart contract invocation calls. It accesses smart contract state through the `wasp_client` and does minimal decoding of data. 

Let's take a look into three parts of this service to make this more clear.

##### PlaceBetOnLedger

The [placeBetOnLedger](https://github.com/boxfish-studio/wasp/blob/feat/roulette_poc_ui/contracts/rust/fairroulette/frontend/src/lib/fairroulette_client/fair_roulette_service.ts#L144) function is responsible for sending On-Ledger bet requests. It constructs a simple IOnLedger object containing:

* The smart contract ID: `fairroulette` 
* The function to invoke: `placeBet` 
* An argument: `-number` 
    * this is the number the player would bet on => the winning number  

Furthermore, this transaction requires an address to send the request to and also a variable amount of funds over `0i`.

:::note
For Wasp, the address to send funds to is the chainId.
:::

See: [CoreTypes](https://iscp.docs.iota.org/docs/misc/coretypes/) and [Invoking](https://iscp.docs.iota.org/docs/misc/invoking/)


##### CallView 

The [callView](https://github.com/boxfish-studio/wasp/blob/feat/roulette_poc_ui/contracts/rust/fairroulette/frontend/src/lib/fairroulette_client/fair_roulette_service.ts#L160) function is responsible for calling smart contract view functions. 

See: [Calling a view](https://iscp.docs.iota.org/docs/guide/solo/view-sc/) 

To give access to the smart contracts state, view functions can be used to return selected parts of the state. 

In our use case, we poll the state of the contract at the initial page load of the frontend. 
State changes that happen afterwords are published through the websocket event system.

To give an example on how to build and call such functions take a look into:

* Frontend: [getRoundStatus](https://github.com/boxfish-studio/wasp/blob/feat/roulette_poc_ui/contracts/rust/fairroulette/frontend/src/lib/fairroulette_client/fair_roulette_service.ts#L176) 

* Smart Contract: [view_round_status](https://github.com/boxfish-studio/wasp/blob/feat/roulette_poc_ui/contracts/rust/fairroulette/src/fairroulette.rs#L289)

Since the returned data of views are encoded in Base64, the frontend needs to decode this by using simple `Buffer` methods. 
The `view_round_status` view returns an `UInt16` which has a state of either `0` or `1`. 

This means to get a proper value from a view call, use `readUInt16LE` to decode the matching value.





#### Dependencies

* [NodeJS >= 14](https://nodejs.org/en/download/)
* [NPM](https://www.npmjs.com/)

#### Configuration

The frontend requires a config file to be created. The template can be copied from `contracts/rust/fairroulette/frontend/config.dev.sample.js` and renamed to `config.dev.js` inside the same folder.

Make sure to update the config values accordingly to your personal setup.
The `chainId` is the chainId which gets defined after deploying a chain (link #Deployment). 

`waspWebSocketUrl`, `waspApiUrl`, `goShimmerApiUrl`: Are dependent on the location of your Wasp and GoShimmer node. Make sure to keep the path of the `waspWeb SocketUrl` (`/chain/%chainId/ws`) at the end. 

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


Follow the [Deployment](https://iscp.docs.iota.org/docs/misc/deploy/) documentation until you reach the `deploy-contract` command.

The deployment of a contract requires funds to be deposited to the **chain**. 
This is archived by executing: 

`./wasp-cli chain deposit IOTA:10000`

To be able to deploy the contract make sure to [Build](#building) it first.

Now deploy the contract with a wasmtime configuration.

`./wasp-cli chain deploy-contract wasmtime fairroulette "fairroulette"  contracts/rust/fairroulette/pkg/fairroulette_bg.wasm`
