![Wasp logo](WASP_logo_dark.png)
# Wasp: a node for IOTA Smart Contracts


The _IOTA Smart Contract Protocol_ (ISCP) is a layer 2 protocol on top of the core Value Tangle
protocol, run by Goshimmer nodes. _Wasp_ connects to Goshimmer node to have access to the Value Tangle.   

The repository represents code which can only be used in testing configurations.
Until the first release, the software is not ready to run as a real node.
 
To run tests, do the following:

- clone `wasp` branch of Goshimmer and compile Goshimmer:

`git clone -b wasp https://github.com/iotaledger/goshimmer.git goshimmer`

`cd goshimmer`    

`go install`    
    
- clone `develop` repository of the Wasp:
    
`git clone -b develop https://github.com/iotaledger/wasp.git wasp`

`cd wasp`

`go install`

- run tests using standard Go testing infrastructure. 
All tests are configured to run on in-memory database and are using mocked Value Tangle on Goshimmer.

For example, the purpose of the integration test `TestSend10Requests0Sec` is to test how nodes synchronisation 
and requests processing in difficult conditions and at hight TPS. 
The testing program does not wait for nodes to finish bootup nor to sync, so requests may reach the Wasp nodes
in any unprepared situation, messages may be lost, nodes may be left behind by the majority of other nodes (go out of sync).
All 10 requests are sent without any delays, essentially all at once. 
In real situation Tangle would have to confirm each request transaction before it will reach the Wasp. 
So, sending all requests at once creates "high TPS" situation.

The test `TestSend10Requests0Sec` goes through the following steps:

- starts 1 Goshimmer node and 4 Wasp nodes in the background
- imports distributed private BLS keys of 3 testing smart contracts to all 4 Wasp nodes.  
Committees of Wasp nodes can generate distributed keys itself but for testing purpose and determinism 
we import pre-generated keys as well as other SC data from file `keys.json`.
- creates bootup records for all 3 test smart contracts in all 4 Wasp nodes
- creates **origin transaction** for one of testing smart contracts and sends it to the Value Tangle (Goshimeer)
- activates the testing smart contract on the committee of Wasp nodes. 
- at this point smart contract is active and ready to accept requests. The testing smart contract 
runs empty program. It does not update the state variables, however state transitions occur: 
colored smart contract token is moved, request tokens are created and destroyed. 
Each state has own timestamp, hash and token balance.
- sends 10 smart contract requests to the active testing smart contract. 
Each request is wrapped in separate value transaction and is sent to Goshimmer. To send requests to 
smart contract Wasp node is not needed.
- All 4 Wasp nodes in the committee process all 10 requests and posts resulting transaction(s)
to the Value Tangle (Goshimmer). Depending on how it goes, all 10 requests may be processed in 1 **batch**. 
In this case 1 state transaction will be sent to Goshimmer. 
Sometimes may happen that 10 requests will be split into several batches. In this case 
the test will result in several state transitions and several state transactions will be sent to Goshimmer. 
 
To run the test:  

`cd ./tools/cluster/tests/wasptest`

`go test -run TestSend10Requests0Sec` 

## Wasp Publisher messages

Wasp publishes important events via Nanomsg message stream (just like ZMQ is used in IRI. Possibly  in the future ZMQ and MQTT publishers will be supported too).

Anyone can subscribe to the Nanomsg output stream of the node. In Golang you can use `packages/subscribe` package provided in Wasp for this.
The publisher's output port can be configured in ```config.json``` like this:
```
  "nanomsg":{
    "port": 5550
  } 
```

Search for  "```publisher.Publish```" in the repo for exact places in the code where messages are published. 

Currently supported messages and formats (space separated list of strings):

|Message|Format|
|:--- |:--- |
|SC bootup record has been saved in the registry | ```bootuprec <SC address> <SC color>``` |
|SC committee has been activated|```active_committee <SC address>```|
|SC committee dismissed|```dismissed_commitee <SC address>```|
|A new SC request reached the node|```request_in <SC address> <request tx ID> <request block index>```|
|SC request has been processed (i.e. corresponding state update was confirmed)|```request_out <SC address> <request tx ID> <request block index> <state index> <seq number in the batch> <batch size>```|
|State transition (new state has been committed to DB)| ```state <SC address> <state index> <batch size> <state tx ID> <state hash> <timestamp>```|
|VM (processor) initialized succesfully|```vmready <SC address> <program hash>```|

## Pluggable VM abstraction
_(for experimenting. Not secure in general)_

Wasp implements VM abstraction to make it possible to use any VM or even language interpreter available on the market 
as smart contract VM processor. At least theoretically. Requirements for the VM processor:

- shoudn't be able to access anything on the host but the sandbox. Otherwise, security breach.

- must be deterministic. Otherwise, smart contract committee won't come to consensus on result.

To plug your VM into the Wasp code, follow the following steps:

- implement `vmtypes.Processor` and `vmtypes.EntryPoint` interfaces for your VM. The VM will be accessing 
runtime environment via `vmtypes.Sandbox` interface.

- implement your VM binary loader (constructor) and register it with `vmtype.RegisterVMType` function in your init code. 

Smart contract loader will be reading program metadata from the registry (see `registry.ProgramMetadata`). 
It will locate program binary (for example `.wasm` file) as specified in the `Location` of the metadata, load it from there 
and will use constructor function to create a VM processor instance from the loaded binary. 
Currently, `Location` is interpreted as `file://<file>` and Wasp will be looking for the binary file `<file>` in the `./wasm` directory.

 