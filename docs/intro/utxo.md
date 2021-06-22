## UTXO ledger and digital assets

_(we use `tokens` and `digital assets` as synonyms)_

It is not our goal here to describe Value Tangle in detail, however, in order to follow other articles more easily 
we shortly introduce here main concepts of `UTXO ledger`.

The Goshimmer's Value Tangle implements _UTXO ledger_ with _colored balances_. 

UTXO stands for `Unspent TransaXion Output`. 
_Colored balances_ means each token in the ledger has _color_: a 32-byte code. 
The default color is _iota-color_ which corresponds to normal iotas. In the genesis of the IOTA ledger
all tokens have _iota-color_. Number of all tokens on the ledger is constant, no matter _color_.

The _UTXO ledger_ contains unspent outputs (UTXOs) rather than just addresses and balances of tokens, 
like in the current IOTA 1.0 and some ether ledgers, like Ethereum. 
Each unspent output of the transaction has the following form: 
```
Address:{colorCode1: balance1, colorCode2: balance2, ...}. 
```

Here _colorCodeN_ is the _color_ of the token, and the _balanceM_ is a number of tokens of that _color_ in the output. 

Each UTXO is always contained in some transaction, i.e. it is an _output of some transaction_. 
So, each output in the ledger is always booked together with the ID (the hash) of its containing transaction:
```
TxID:Address:{colorCode1: balance1, colorCode2: balance2, ...}
```

The _address balance_ is a collection of UTXOs with the same address. 
So, compared to the account-based ledger, the address balance in UTXO ledger becomes a rather complicated thing: 
it contains a 2D collection of outputs: each output contains the transaction ID of the output 
and token balances for each color. 

We transfer tokens by spending UTXOs, i.e. by posting a transaction with those outputs. 
The spending transaction takes UTXOs as inputs and produces new UTXOs. 
It must contain valid signatures corresponding to addresses of its inputs. 

#### Value transaction validation rules
Let’s say C is the non-iota color with positive balance in outputs. 
The general validation rules of value transactions are the following:

1. Number of tokens of all colors in inputs and outputs must be equal (number of tokens in the ledger is constant)
2. Balance of tokens with color C in outputs must be equal or less than the number of C tokens in inputs. 
In the _strictly less_ case we say some _tokens of color C are destroyed (uncolored)_.
3. Some balances in outputs can contain special color code `new-color`. This code 
means some tokens in the input acquire the color of the _hash of the containing transaction_. 
In the new UTXOs they will be booked with the hash of that transaction as a color code. 
It is called _minting of the new digital assets (colored balance, colored supply, tokens)_
4. The number of tokens with `iota-color` in the outputs can be smaller or larger than iotas in inputs, 
provided condition (1) is satisfied. 

ISCP heavily relies the logic of the tokens in the UTXO ledger. In documentation we will describe exactly how. 
Meanwhile, to simplify our thinking, we can represent 2D address balance as its 1-dimensional 
"projection to the color axis": a collection of sub balances per color, like this:
```
     color1: balance1
     color2: balance2
     ...
``` 

You can find the smart contract address balance displayed in this form in the dashboard of the _DonateWithFeedback_ 
as well as other PoC smart contracts.

