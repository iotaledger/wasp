
## General scheme of invoking the entry point
The entry point is a function of the smart contract. 
By “invocation” we mean calling the function from a particular environment. 
The effect of invoking an entry point depends on the calling environment and who is calling. 
Sending a request is an example of invocation of an entry point.

In general, each call of the entry point is similar to the object method call in 
the object-oriented paradigm. The invocations consists of:

* `Hname` of the target contract, the called object 
* `Hname` of the entry point function name, the name of the method
* Parameters: a collection of key/value pairs, the (named) parameters of the call
* Transferred tokens: a collection of color/balance pairs, a special type of parameter of the call

The invocation always returns a result, a collection of key/value pairs or error, if the call fails.
So, the following is a generic structure of the call to the entry point:
```
res = target_contract.function(parameters, transfer)
```
where `res` is a map (a dictionary) of key/value pairs, containing the result (possibly empty) or an error code.

There are several ways to invoke the entry point of the smart contract: request, call and view call.

* _request_ can be sent to the target contract from the “outside”: a wallet  
or from another smart contract, on the same or on another chain. 
The Solo environment is the “outside” (see [Invoking smart contract](06.md)). 
The _request_ itself is a transaction, it contains parameters and attached tokens. 
The tokens are transferred to the smart contract together with parameters and become part of its account. 
Requests can only be invoked (sent) with a private key which controls those tokens. 
Sending a request to a view will trigger an error and fallback actions.
In the Solo environment requests are stored on the UTXODB ledger and handled by the Solo environment, 
backlog is collected and requests are forwarded to the corresponding target chains.

* any _entry point_ can be _called_ from another smart contract on the same chain. 
It is just like a usual function call in a programming language. 
In this case both caller and target smart contracts assume the same state of the chain, 
i.e. the call is always synchronous.  

* The _views_ can be called from anywhere, including from outside, 
for example from a web API which fetches the smart contract data for display.