## The `eventlog` contract

The `eventlog` contract holds immutable on-chain log of events.

Each event is emitted by `Event()` sandbox call from the smart contract. 
This is the only way to create `eventlog` records and publish events. 
VM core logic emits event records when deploying a contract, settling the request, confirming a block (state transition)
and others. 

An event contains arbitrary data, typically a string. 

Emitting an event means the following actions:
* recording the event data into the `eventlog` core contract under the emitting contract's `hname` and 
timestamped with the current timestamp of the contracts.
* logging the event on the node with sandbox's `Log().Info()` 
* sending the event over the `nanomsg` publisher to subscribers of the node events 
(in the future other publishers, like `zmq` and `mqtt`) will be supported

### Entry points
The `eventlog` core contract does not contain any entry points which modify its state.

The only way to modify `eventlog` state is to add an event record from the smart contract by calling 
sandbox method `Event()`. 

### Views
* **getNumRecords** returns total number of records recorded by a smart contract with particultal `hname` (parameter)

* **getRecords** query log records according to filter criteria specified in parameters. 
The records are returned in descending order of timestamps, i.e. latest first.  The filter parameters:
    * `hname` of the contract. Mandatory
    * `from timestamp` timestamp in Unix nanoseconds. Default is 0
    * `to timestamp` timestamp in Unix nanosecods. Default is `now`
    * `max records` maximum number of records to return. Default is 50   
