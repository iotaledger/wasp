## Smart Contracts

Note that most example smart contracts only show the concepts involved in
creating them but should not be taken as fully functional code yet.

Sample smart contracts:

- dividend

  Allows for setting up automatic dividend dispersal to a number of recipients
  according to distribution factors per recipient. Any tokens sent to this
  contract will automatically be divided accordingly over these recipients.

- donatewithfeedback

  Allows for donations and registers feedback associated with the donation. The
  contract owner can at any point decide to withdraw donated funds from the
  contract.

- fairauction

  Allows an auctioneer to auction a number of tokens. The contract owner takes a
  small fee. The contract guarantees that the tokens will be sent to the highest
  bidder, and that the losing bidders will be completely refunded. Everyone
  involved stakes their tokens, so there is no possibility for anyone to cheat.

- fairroulette

  A simple betting contract. Betters can bet on a random color and after a
  predefined time period the contract will automatically pay the total bet
  amount proportionally to the bet size of the winners.

- helloworld

  ISC version of the ubiquitous "Hello, world!" program.

- inccounter

  A simple test contract. All it does is increment a counter value. It is also
  used to test basic ISC capabilities, like persisting state, batching
  requests, and sending (time-locked) requests from a contract.

- testcore

  Helper smart contract to test the ISC core functionality.

- testwasmlib

  Helper smart contract to test the WasmLib functionality.

- tokenregistry

  Mints and registers colored tokens in a token registry.

### How to create your own Rust smart contracts

We provide _Schema Tool_ to help developers to develop their projects. See the [README](../../tools/schema/README.md) in _Schema Tool_ for more information.
