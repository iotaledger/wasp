_Obsolete. To be adjusted to the new version_

----

# Handling tagged tokens with _TokenRegistry_ and _FairAuction_

Tagged tokens, (a.k.a. colored tokens) is a powerful feature of the [UTXO ledger](../utxo.md).
Two PoC smart contracts, _TokenRegistry_ and _FairAuction_ demonstrates how smart contract can manipulate them and 
extend their functions. The demo dashboards can be found [here](http://waspdev01.iota.cafe:10000).

## _TokenRegistry_

Any wallet can mint tagged tokens by creating transaction output with `new-color`. The _minting_
means "tagging" the iota tokens with color code, "painting" iota tokens.

The tagged tokens then can be moved around and sent to addresses just like any other token. 
The tag of it, the `color`, is a 32 byte long code which doesn't say much about the token.

The _TokenRegistry_ is a simple smart contract which keeps a registry of tagged tokens with their metadata.
It also keeps cryptographical proof when, how many and which address minted the tagged tokens. 

The program Go code with program hash `8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p` of the smart contract
can be found [in repository](https://github.com/iotaledger/wasp/blob/master/packages/vm/examples/tokenregistry/impl.go).

The following command mints 3 tokens into the same address of the wallet.
 
`wasp-cli fa mint "My first 3 coins" 3`

With `fa` tag `wasp-cli` derives we are dealing with the `TokenRegistry`.
The  command creates a transaction which contains wor things in it:
- "minting" itself, assigning the new tag to 3 iotas asn sending them to the same address
- the `mint` request to the _TokenRegistry_.

It is an atomic operation: either both token minting and sending the request is successful, or none of it.

The program will output something like this:
```
using Goshimmer host waspdev03.iota.cafe:8080
Minted 3 tokens of color Fu5UEJ2SiJZdTUbS72NkQ9uyQPQtqMDJW4SodDdr5G1w into address WoCZCHKr1AsbDXXyPQn5JDs7jkbQ2LNDrADXGXTqB7pW.
Metadata of the supply: 'test coin for an auction'
Metadata was sent to TokenRegistry SC at bX3H7Sfh8Ez3g5ygevbUnchTKm6NjTtD3MsHQKGuMTgC
```

The minted tokens goes right into the same address which created the `mint` request.

The request itself is picked by the SC and result in record in the state of it. It contains:
- color of the minted tokens
- minted supply, 3 in this case
- address, the originator of the new supply
- description text
- any optional attached metadata

The record cannot be faked, and it is immutable in the smart contract. 
So, if a supply was minted with the _TokenRegistry_,
the metadata contained in it can be trusted by anyone, because it is a cryptographical proof it is correct.

## _FairAuction_

What can we do with tagged tokens? They actually can represent real world assets, like shares, 
collectible tokens or any other use case. The powerful thing is those tokens can be moved around the ledger
between addresses using standard IOTA wallet: there's no need for any smart contracts.

But if we want to sell it on the market, we have to entrust someone who is running that market and send it the
token(s) we want to sell. Example of would be an exchange.

Here we provide alternative: a simple demo smart contract _FairAuction_ which implements a simple 
distributed marketplace where you can sell and buy tagged tokens for iotas in an auction.

In order to sell one of 3 tokens we minted above we can use the following command:

`wasp-cli fa start-auction "my first auction" Fu5UEJ2SiJZdTUbS72NkQ9uyQPQtqMDJW4SodDdr5G1w 1 100 60` 

The command will send request to the smart contract instance and it will start auction for 1 token of
color `Fu5UEJ2SiJZdTUbS72NkQ9uyQPQtqMDJW4SodDdr5G1w` with minimum bid of 100 iotas and duration of 60 minutes.

With the `start-auction` request the tagged token will be sent to the smart contract address. 

Smart contract starts an auction by self-posting `close-auction` request which is **time-locked** for the duration
of the auction, i.e. in the example above for 60 minutes. 
The committee nodes will automatically run that request after 60 minutes as a result of consensus between nodes that 
auction closing time is over.  

Next 60 minutes anyone can place bids in the auction with commands like this:

`wasp-cli fa place-bid Fu5UEJ2SiJZdTUbS72NkQ9uyQPQtqMDJW4SodDdr5G1w 120`

Upon expiration of the duration of the auction, the time-locked request will be unlocked and 
auction will be closed by selecting first bidder with the highest bid as a winner. 
It is all programmed in the [code of the smart contract](https://github.com/iotaledger/wasp/blob/develop/packages/vm/examples/fairauction/impl.go).

The smart contract will transfer tokens for sale to the winner and will return all placed sums to non-winning bidders. 
It is all codded in the smart contract and it can't be overriden by any parties which do not have quorum. 
Hence _FairAuction_.

   
  

