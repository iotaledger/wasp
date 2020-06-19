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

The purpose of the integration test `TestSend10Requests0Sec` is to test how nodes are syncing 
and requests are processed in difficult conditions and at hight TPS. 
Testing program does not wait for nodes to finish bootup nor to sync, so requests may reach the Wasp nodes
in any unprepared situation, messages may be lost, nodes may go out of sync (left behind the majority).
All 10 messages are sent without any delays, essentially all at once. 
Normally, in real situation Tangle would have to confirm each 
request transaction before it will reach the Wasp.

The test goes through the following steps:

- starts 1 Goshimmer node and 4 Wasp nodes in the background
- imports distributed private BLS keys of 3 test smart contracts to all 4 Wasp nodes.  
(Wasp nodes can generate distributed keys itself but for testing purpose and determinism 
we import pre-generated keys as well as other SC data from file `keys.json`)
- creates bootup records for all 3 test smart contracts in 4 Wasp nodes
- creates **origin transaction** for the first smart contract and sends it to the Value Tangle (Goshimeer)
- activates first test smart contract on Wasp nodes. At this point smart contract becomes active and ready to 
accept requests. The first testing smart contract runs empty program (which does not update the state variables). 
However, state transitions occur and each has own timestamp.
- sends 10 smart contract requests to active test smart contract. 
Each request is wrapped in separate value transaction and is sent to Goshimmer 
(to send requests to smart contract Wasp node is not needed).
- All 4 Wasp nodes in the committee process all 10 requests and posts resulting transaction(s)
to the Value Tangle (Goshimmer). Depending on how it goes, all 10 requests may be processed in 1 **batch**. 
In this case 1 state transaction will be sent to Goshimmer. 
But sometimes it may happen that 10 requests will be split in several batches and in this case 
the test will result in several state transitions and several state transactions sent to Goshimmer. 
 
To run it:  

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

