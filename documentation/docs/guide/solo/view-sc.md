---
description: The basic principle of calling a view is similar to sending a request to the smart contract. The essential difference is that calling a view does not constitute an asynchronous transaction, it is just a direct synchronous call to the view entry point function, exposed by the smart contract.
image: /img/tutorial/call_view.png
keywords:
- testing
- solo
- views
- call
- synchronous
- entry points
---
# Calling a View

The following snippet shows how to call the view entry point `getString` of the
smart contract `example1` without parameters:

```go
res, err := chain.CallView("example1", "getString")
```

The call returns both a collection of key/value pairs `res` and an error result
`err` in the typical Go fashion.

[![Calling a view process](/img/tutorial/call_view.png)](/img/tutorial/call_view.png)

The basic principle of calling a view is similar to sending a request to the
smart contract. The essential difference is that calling a view does not
constitute an asynchronous transaction, it is just a direct synchronous
call to the view entry point function, exposed by the smart contract.

Therefore, calling a view doesn't involve any token transfers. Sending a
request (a transaction) to a view entry point will result in an exception. It
will return all attached tokens to the sender (minus fees, if any).

Views are used to retrieve information about the state of the smart contract,
for example to display the information on a website. Certain _Solo_ methods such
as `chain.GetInfo`, `chain.GetFeeInfo` and `chain.GetTotalAssets` call views of
the core smart contracts behind the scenes to retrieve the information about the
chain or a specific smart contract.

## Decoding Results Returned by _PostRequestSync_ and _CallView_

The following is a specific technicality of the Go environment of _Solo_.

The result returned by the call to an entry point from the _Solo_ environment
is in the form of key/value pairs, the `dict.Dict` type. It is an alias of `map[string][]byte`.
The [dict.Dict](https://github.com/iotaledger/wasp/blob/master/packages/kv/dict/dict.go)
package implements the `kv.KVStore` interface and provides a lot of useful
functions to handle this form of key/value storage.

Since view calls are synchronous, in normal smart contracts operations one can only retrieve
results returned by view calls. Sending a request to a smart
contract is normally an asynchronous operation, and the caller cannot retrieve
the result. However, in the _Solo_ environment, the call to `PostRequestSync` is
synchronous, and the caller can inspect the result: this is a convenient
difference between the mocked _Solo_ environment, and the distributed UTXO
Ledger used by Wasp nodes. It can be used to make assertions about the results
of a call in the tests.

In the `TestTutorial3` example `res` is a dictionary where keys and values are binary slices.
The following fragment creates object `par` which offer all kinds of useful function to take 
and decode key/value pairs in the `kv.KVStoreReader` interface (the logger is used to log the 
data conversion and any errors that may occur). The second statement takes the value of the key `paramString` from the key/value store and attempts
to decode it as a `string`. It panics if the key does not exist, or the data conversion fails.

```go
 par := kvdecoder.New(res, chain.Log)
 returnedString := par.MustGetString("paramString")
```
