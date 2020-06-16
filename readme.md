![Wasp logo](WASP_logo_dark.png)
# Wasp: a node for IOTA Smart Contracts


The _IOTA Smart Contract Protocol_ (ISCP) is a layer 2 protocol on top of the core Value Tangle
protocol, run by Goshimmer nodes. _Wasp_ connects to Goshimmer node to have access to the Value Tangle.   

The repository represents code which can only be used in testing configurations.
Until the first release, the software is not ready to run as a real node.
    
Wasp tests run different scenarios on with one Goshimmer node with mocked Value Tangle
and 4 Wasp nodes.
To ensure determinism, all private keys are imported from file, not generated.
    
To run tests, do the following:

- clone `wasp` branch of Goshimmer and compile Goshimmer:

`git clone -b wasp https://github.com/iotaledger/goshimmer.git goshimmer`

`cd goshimmer`    

`go install`    
    
- clone `develop` repository of the Wasp:
    
`git clone -b develop https://github.com/iotaledger/wasp.git wasp`

`cd wasp`

`go install`

- run tests:

`cd ./tools/cluster/tests/wasptest`

The following are integration tests to test basic functionality:
 
`go test -run TestPutBootupRecords` put bootup records for 3 smart contacts

`go test -run TestActivate1SC` activate 1 smart contact

`go test -run TestActivate3SC` activate 3 smart contacts

`go test -run TestCreateOrigin` create origin state

`go test -run TestSend5Requests1Sec` process 5 requests, 1 per second

`go test -run TestSend10Requests0Sec` process 10 requests, no waiting 
