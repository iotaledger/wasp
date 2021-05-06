## The `eventlog` contract

The `eventlog` contract is one of the [core contracts](coresc.md) on each ISCP
chain.

The function of the `eventlog` contract is to record an immutable on-chain log
of events.

Each event is emitted by the `Event()` sandbox call from the smart contract.
This is the only way to create `eventlog` records and publish events. VM core
logic emits event records when deploying a contract, settling the request,
confirming a block (state transition), and on other occasions.

An event contains arbitrary data, typically a string.

Emitting an event means the following actions:

* Recording the event data into the `eventlog` core contract under the emitting
  contract's `hname` and timestamped with the current timestamp of the contract.
* Logging the event on the node with sandbox's `Log().Info()`
* Sending the event over the `nanomsg` publisher to subscribers of the node
  events (in the future other publishers, like `zmq` and `mqtt`) will be
  supported

### Entry points

The `eventlog` core contract does not contain any entry points which modify its
state.

The only way to modify the `eventlog` state is by adding an event record from
another smart contract by calling sandbox method `Event()`.

### Views

* **getNumRecords** - Returns the total number of records recorded by the smart
  contract with specified `hname` (parameter)

* **getRecords** - Queries log records according to filter criteria specified in
  the parameters. The records are returned in descending order of timestamp,
  i.e. the latest is returned first. The filter parameters:
    * `hname` of a contract. Mandatory
    * `from timestamp` timestamp in Unix nanoseconds. Default is 0
    * `to timestamp` timestamp in Unix nanoseconds. Default is `now`
    * `max records` maximum number of records to return. Default is 50   
