_Obsolete. To be adjusted to the new version_

----

# Main concepts with _DonateWithFeedback_
We will explain main concepts of IOTA smart contracts using deployed PoC smart contract called _DonateWithFeedback_.
We wrote this article as a part of the _PoC release of the Wasp_. 
The [demo dashboard](http://waspdev01.iota.cafe:10000) 
of the instance of the _DonateWithFeedback_ SC and _wwallet_, 
a smart contract wallet, can be used as a visual aid for the text below. 
 
## Introduction 
_DonateWithFeedback_ is the simplest of all 3 hard coded smart contracts, included into the PoC release of the Wasp. 
You may see it as a “Hello, world” example. We will go along all main concepts of the IOTA Smart Contracts using 
it as an example.

_DonateWithFeedback_ is a dApp which performs a very common function: it handles an IOTA donation 
address and user’s feedback log for the owner of a web page.  

_(You may want to follow the Go [code of this smart contract](https://github.com/iotaledger/wasp/blob/master/packages/vm/examples/donatewithfeedback/dwfimpl/impl.go) 
which currently is embedded right into the Wasp. This approach is not typical for the smart contracts 
in general, however it is very convenient for testing the protocol. 
In the future all smart contract programs will be run on wasm VM)_   

The common practice is to display the donation address on a web page for anyone wanting to send us a donation, 
for example with their Trinity wallet. 
What if we also want each donor to attach some feedback/comment text to the donation? For example, 
“Here I send you 2 MIOTA because I like your site”? 
To support that possibility, we would need some kind of application, or an extension of the existing IOTA wallet 
with a database for comment messages behind and so on (of course, there is more than one way to do that, including 
embedding the text right into the transaction).

_DonateWithFeedback_ smart contract takes responsibility to handle both the donation 
account (the address) and the log of feedback messages. The log of feedback messages will be kept in the 
immutable storage of the smart contract’s ledger (the state). 
Some may say it is an overkill, and yes, it may be a bit artificial, but why not? 
Once the smart contract is fee-less anyway? Do not try this with Ethereum :smirk:.

## The state

Let’s open a [dashboard](http://waspdev01.iota.cafe:10000/dwf) of the demo instance of _DonateWithFeedback_. 
The dashboard displays the _state of the smart contract_ instance, i.e. what is contained in its tamper proof storage. 

The following are the main properties of any smart contract. 
- smart contract address (_SC address_)
- Program hash
- smart contract color (_Color_)
- description

It is displayed in the section _Smart contract details_ of the demo dashboard.
These values of the state normally do not change throughout the life of the smart contract.

The state of any smart contract consists of two parts:
- the _Balance_ part displays balances of digital assets contained in the smart contract account address.
 It is an **on-tangle** part of the state.
 
- The collection of _key-value pairs_ which can be interpreted as variables and their values. 
It is an **off-tangle** part of the state. In general, it can contain any data of arbitrary size.
In the _DonateWithFeedback_ smart contract the dashboard displays the generic data of the state in 
an SC-specific and user-friendly format: 
the statistics of donations and list of feedback messages stored in the log. 

(note that we use [Base58 encoding](https://en.bitcoinwiki.org/wiki/Base58) for binary data of fixed size, like 
transaction hashes and addresses displayed on the dashboard). 

## The address and program of smart contract

Naturally, we can have many instances of the _DonateWithFeedback_ smart contract. Each instance will correspond to 
a specific donation address and its owner. The demo dashboard is displaying the state of one particular instance. 

Two things define instance of any smart contract:

- _Program code_, represented by the 32-byte long _program hash_
- _Smart contract address_, an address on the Tangle

The _program hash_ is a globally unique identifier of the program. 
The program itself encodes algorithms of the smart contract in some language. In IOTA smart contracts we will use
WebAssembly (a.k.a. _Wasm_) as the binary format of the smart contract program. 
The hash of Wasm binary ensures the program code cannot be modified without changing the hash.

For the hard coded PoC smart contract such as _DonateWithFeedback_, the program hash is just 
an ID statically embedded into the Wasp code. 
For the builtin program of _DonateWithFeedback_ program hash is equal to `5ydEfDeAJZX6dh6Fy7tMoHcDeh42gENeqVDASGWuD64X`.
For a WebAssembly smart contract code it will be the hash of its binary code. 

The _smart contract address_ is a 33-byte long address as it is implemented in Goshimmer, 
for example `pxsUocho2dJQ8EX5PxHUHY7e4qfVmcu7K4dGRrrFrwaG`.

The SC address is a global unique identifier of the smart contract instance: 
the smart contract can always be located on the Tangle by its address. 
The address scheme we use in ISCP is a BLS address, a kind of multi-signature address. 
Tokens, contained in the address, can be moved only with the corresponding private keys. 

The smart contract program always “hides” behind the smart contract address, which 
is otherwise just a normal address. The _DonateWithFeedback_ program is attached to each instance of 
_DonateWithFeedback_ program. 

Unlike ordinary addresses on the Tangle, we can send _requests_ to the SC address, not just tokens. 
The attached program processes requests sent to it and updates the state of 
the smart contract. The state can be updated only if the processor which runs the program 
has access to the corresponding private keys.

Whenever we [deploy](deploy.md) a smart contract instance on the Tangle you have to specify its program hash, among other things. 
Wasp node will find the program code by the hash and will link the program with the instance. 

Just like the usual `Ed25519` address is generated from its private key, 
the smart contract address is generated during the deployment process of the smart contract. 
 
In _DonateWithFeedback_ we use the smart contract address as a donation address to display on the web page.
We use this address to send donations. This is just an address on the Tangle, 
so you can send iotas and other tokens to it using ordinary wallet. 

However, this is not what we want to achieve here. The important thing is you can send **requests** to the 
smart contract address, value transactions with metadata attached to it: 
_smart contract transactions_ (_SC transactions_). 
Specifically, you can send a `donate` request to the _DonateWithFeedback_ smart contract which encapsulates 
both donated iotas and the feedback message in a single transaction.

## Balance of the smart contract account

SC transactions are value transactions, i.e. they move tokens from addresses to addresses. 
It is because ISCP is a protocol on top of the Goshimmer's UTXO Ledger.
Here you can find a [short introduction](../utxo.md) to main concepts of it.

The smart contract account balance consists of all UTXOs contained in the smart contract address. 
On the demo dashboards it is displayed in the simplified 1-dimensional form by skipping the containing transaction 
IDs and summing up number of tokens for each color:
```
ColorCode1: balance1
ColorCode2: balance2
```

The smart contract balance is an _on-tangle_ part of the smart contract’s state. 
It means the UTXO Ledger takes care of the consistency and immutability of the smart contract balance.

Tokens of the smart contract balances can be moved only by the smart contract instance, 
which is the only entity which owns private keys. 
Actually, the tokens can be moved by the smart contract program of the instance as a result of processing the request.

For example, the account balance of _DonateWithFeedback_ smart contract instance contains all tokens sent 
to its address. Only the smart contract can move those tokens. 

If the owner of the site wants to withdraw donated iotas from the donation 
address, he or she will have to _nicely ask_ the smart contract to do it for them.

## Requests to the smart contract

“Nicely asking” the smart contract to do something means sending a request to its instance. 
A request to the smart contract always has the following structure:
```
Request code, Arguments --> SC address
```

Here `SC address` is the address of the target smart contract instance. 
`Request code` is analogous to a function/method name, `Arguments` are parameters. 
So, sending a request to a smart contract is equivalent to a method call in the object-oriented paradigm, 
except that caller doesn't wait for the result.

Upon receiving the request, the smart contract finds the function in the smart contract program
which implements the request code and, if the function exists, it executes it. 
The program function moves tokens and does all other things a smart contract program is supposed to do.

Some request codes may be marked as _protected requests_. These requests will only be processed by the smart 
contract if the sender of the request is the _owner_ of the smart contract (see [deployment](deploy.md)). 
Protected requests sent by other than owner address will be equivalent to _no operation_ 
and attached funds won't be returned. 

For the user of the smart contract, however, sending a request is not complicated at all, because it is usually 
is implemented as a feature the wallet.
 
For example, you may try the two types of requests implemented by the _DonateWithFeedback_ smart contract:
 
- The `donate` request donates iotas to the address and gives a feedback message. 
- The `withdraw` request allows the owner to take the iotas out from the smart contract account. 

Let’s say address of our donation smart contract is `pxsUocho2dJQ8EX5PxHUHY7e4qfVmcu7K4dGRrrFrwaG`.
Here are full sequence of commands to `wasp-cli` to donate some iotas to our project:

Initialize your demo wallet.
It will create the _wwallet.json_ file in the current directory with the private and public keys of the the address:

`wasp-cli init`

Request some demo iotas from the Goshimmer’s faucet: 

`wasp-cli request-funds`

Check balance of your newly created wallet:

`wasp-cli balance`

It will display something like:
```
  Address: YyVP3g9B7YfvHCEzLaeD4M3iqKSawyuf8NsXxggghFhP
  Balance:
    IOTA: 1337
    ------
    Total: 1337
``` 

Now let the wallet know the specific address of our demo _DonateWithFeedback_ 
smart contract instance (_dwf_ here stands for _DonateWithFeedback_, the command below associates 
it with the specific address):

`wasp-cli dwf set address pxsUocho2dJQ8EX5PxHUHY7e4qfVmcu7K4dGRrrFrwaG`

Now we can send a `donate` request to the demo `DonateWithFeedback` smart contract instance:

`wasp-cli dwf donate 42 “This is my first donation to the IOTA Smart Contract Project`  

In approximately 15-20 seconds the dashboard of _DonateWithFeedback_ will show you the new state 
of the smart contract: with 42 iotas more in the balance and your message in the log.

## Processing the request

What is going on when the we send the request like the one aboveto the smart contract address? 
Let's look at it step-by-step:

1. The `wasp-cli` command creates a value transaction containing the smart contract request. 
It takes 42 iotas from the address of the wallet, signs the transaction with the private key of the wallet. 
Then it posts it to the UTXO Ledger, i.e to Goshimmer node, running on the Pollen network. 
(there's a bit more of a token manipulation behind scenes but we skip it here)

2. Pollen network confirms the transaction containing the request. 
It takes some ~10 seconds. After confirmation the request it becomes part of the immutable backlog for 
the smart contract instance, represented by the target SC address.

3. Smart contract instance picks up the request and runs the smart contract program. 

4. The smart contract program computes a state update, which includes new outputs in the UTXO ledger and a 
record in the log of feedback messages. Then the smart contract produces another transaction, 
called _state or anchor transaction_ which carries all the information about the new state of the SC. 

5. The smart contract committee multi-signs the new state transaction. 

6. The smart contract posts the state transaction to the tangle and Pollen network confirms 
it in yet another ~10 seconds. This is how state transition occurs. 

So, all in all, it takes to produce and confirm 2 subsequent transactions to 
update the state of the smart contract, i.e. about 20 seconds in the Pollen network. 

It may not look very fast, however, each state update can contain hundreds of (batched) requests, 
so even one smart contract instance is able to process many requests per second on the average. 
We normally do not expect too many donations per 
second (although it would be nice!), however for other use cases TPS for one smart contract may be very important.

## The state: more details 
 
The _State details_ section of the _DonateWithFeedback_ dashboard shows values, which are changing with each donation.

The current state of any smart contract has:
- `state index`. Smart contract state is a chain of batches of state updates (think: blocks). 
Each batch has an index (think: block index). The origin state has index 0, it is incremented with each new batch 
of updates.
- `state hash` It is the result of hashing all state values since origin
- `anchor transaction` (or `state transaction`) which secures state on the tangle
- `batch of requests` (a block) which resulted in the current state from the previous. 
- the `timestamp` of the state. The timestamp of the state is a timestamp
 _consistent with the local clocks_ of all committee nodes (see below) and 
 all committee nodes have consensus on it. So, it is an _objective_ timestamp, not just some logical value.
 
As mentioned above, the _on-tangle_ part of the state of the smart contract is the balance of its account address. 
It contains tagged (colored) tokens. You can see it as the Balance in the _DonateWithFeedback_ dashboard.

The state of the smart contract also contains arbitrary key/value pairs which can be interpreted as variables with values. 
This part of the state is stored _off-tangle_, i.e. is not a part of transactions on the Pollen network 
but instead it stored in _smart contract ledger_, the distributed database run by 
the Wasp nodes for each smart contract. 

The log of feedback messages in the _DonateWithFeedback_ smart contract state is a collection of key/value pairs in 
the database. The size of the data contained in the smart contract state is limited only by the underlying database, 
i.e. is unlimited in most practical use cases. Think about Oracle with historical temperature data for months 
and years in its temper-proof state. 

Just like the balances are locked in the smart contract address, the whole state of the smart contract, 
including the _off-tangle_ part, is locked into the account address and can be updated only by the smart 
contract instances with its private keys and the program. 
We won’t go into details here, but the base principle is anchoring the hash of the off-tangle part of the 
state in the on-tangle transaction.

### Conclusion

The _DonateWithFeedback_ smart contract state contains the endless log of objectively time stamped donation 
feedback messages, also state variables such as _totalDonatedSum_ and _maxDonationSoFar_. 
Each donation not only brings iotas to the smart contract account, but also it adds records to 
the time stamped log of messages. The _time stamped log_ is an immutable append-only data structure 
in the state of the smart contract which can be queried by the time parameter, for example extracting records 
between two moments of time. 

In this article we went through the main concepts of IOTA smart contracts, such as _SC address_, 
_program hash_, _SC state_, _requests_ and _SC transaction_ using a simple yet typical 
smart contract _DonateWithFeedback_ as an example. 
 
 
   
