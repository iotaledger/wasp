# UTXO Ledger and Digital Assets

(we use `tokens` and `digital assets` as synonyms)

It is not our goal here to describe the UTXO Ledger in detail, however, in order
to follow other articles more easily we will introduce the main concepts of
the `UTXO Ledger` here.

Goshimmer implements a _UTXO Ledger_ with _colored balances_.

UTXO stands for `Unspent Transaction (TX) Output`. _Colored balances_ means that
tokens in the ledger have a 32-byte attribute called a _color_. The default
color is _ColorIOTA_ which corresponds to normal iotas. In the genesis of the
IOTA ledger all tokens were assigned _ColorIOTA_. The number of all tokens on
the ledger is constant, no matter the _color_.

The _UTXO Ledger_ contains unspent transaction outputs (UTXOs) rather than just
addresses and balances of tokens, like in Bitcoin and unlike like Ethereum.
Each unspent output of the transaction has the following form:

```
Address: {colorCode1: balance1, colorCode2: balance2, ...}. 
```

Here _colorCodeN_ is the _color_ of the token, and _balanceM_ is the number of
tokens of that _color_ in the output.

Each UTXO is always contained in some transaction, i.e. it is an _output of some
transaction_. So, each output in the ledger is always booked together with the
ID (the hash) of its containing transaction:

```
TxID: Address: {colorCode1: balance1, colorCode2: balance2, ...}
```

The _address balance_ is a collection of UTXOs with the same address. So,
compared to the account-based ledger, the address balance in an UTXO ledger
becomes a rather complicated thing: it contains a 2D collection of outputs: each
output contains the transaction ID of the output and token balances for each
color.

We transfer tokens by spending UTXOs, i.e. by posting a transaction with those
outputs. The spending transaction takes UTXOs as inputs and produces new UTXOs.
It must contain valid signatures corresponding to the addresses of its inputs.

#### Value transaction validation rules

Let’s say C is a non-iota color with positive balance in outputs. The general
validation rules of value transactions are the following:

1. The number of tokens of all colors in inputs and outputs in a transaction
   must be equal (the number of tokens in the ledger is constant)
2. The total of tokens with color C in outputs must be equal or less than the
   total of C tokens in inputs. In the _strictly less_ case we say some _tokens
   of color C are destroyed (uncolored)_.
3. Some balances in outputs can contain the special color code `new-color`. This
   code means that some tokens in the input acquire the color of the _hash of
   the containing transaction_. In the new UTXOs they will be booked with the
   hash of that transaction as their color attribute. This is called
   _minting_ of new digital assets (colored balance, colored supply, tokens)
4. The number of tokens with `ColorIOTA` in the outputs can be smaller or
   larger than iotas in inputs, provided condition (1) is satisfied.

The IOTA Smart Contracts relies heavily on the logic of the tokens in the UTXO Ledger. In later
documentation we will describe exactly how. Meanwhile, to simplify our thinking,
we can represent a 2D address balance as its 1-dimensional "projection to the
color axis": a collection of sub balances per color, like this:

```
color1: balance1
color2: balance2
...
``` 

You can find the smart contract address balance displayed in this form in the
dashboard of the _DonateWithFeedback_ smart contract example as well as in other
PoC smart contracts.
