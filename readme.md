![Wasp logo](WASP_logo_dark.png)
# Wasp: a node for IOTA Smart Contracts


The _IOTA Smart Contract Protocol_ (ISCP in short) is a layer 2 protocol on top of the core Value Tangle
protocol run by Goshimmer nodes. 

_Wasp_ node implements ISCP. To run smart contract you will need a network of _Wasp_ nodes.
Each _Wasp_ connects to _Goshimmer_ node to have access to the Value Tangle, 
the layer which implements the ledger of tokens.   

The repository represents code which can only be used in testing configurations, it is not ready for the commercial use.

### Run Goshimmer with WaspConn

To run a Wasp node you need a Goshimmer instance with the [WaspConn](https://github.com/iotaledger/goshimmer/tree/wasp/dapps/waspconn) plugin in it. 
This version of Goshimmer is located in the
 [_wasp_ branch of the Goshimmer repository](https://github.com/iotaledger/goshimmer/tree/wasp) 

First you have to clone and compile the Goshimmer version from the `wasp` branch. 

`git clone -b wasp https://github.com/iotaledger/goshimmer.git goshimmer-wasp`
`cd goshimmer-wasp`    
`go install`    

The Goshimmer instance in the Pollen network must be configured and started according to the 
[Goshimmer instructions](https://github.com/iotaledger/goshimmer/wiki/Setup-up-a-GoShimmer-node-(Joining-the-pollen-testnet)).  

The only difference between standard Goshimmer (the `develop` branch) and the `wasp` branch is `WaspConn` plugin.
It allows any number of Wasp nodes to connect to the Goshimmer instance running it.

By default, WaspConn plugin will be listening for Wasp connections on the port `5000`. 
To change this setting include the following section in the Goshimmer's `config.json` file:

```
  "waspconn": {
    "port": 12345
  }
```
   
### Run Wasp node

Note that you need multiple Wasp nodes to form committees for smart contracts. 
Many Wasp nodes can be connected to one Goshimmer instance.
    
Clone `develop` repository of the Wasp:
    
`git clone -b develop https://github.com/iotaledger/wasp.git wasp`
`cd wasp`
`go install`

`cd directory-with-config-file`
`wasp`

The `config.json` must be present in the directory where you run a Wasp node. 
Most of the setting are self-explanatory in the following example below or are the same as in the Goshimmer configuration.
Other are explained below.

An example of the `config.json` file for a Wasp instance:
```
{
  "database": {
    "inMemory": false,
    "directory": "waspdb"
  },
  "logger": {
    "level": "debug",
    "disableCaller": false,
    "disableStacktrace": true,
    "encoding": "console",
    "outputPaths": [
      "stdout",
      "wasp.log"
    ],
    "disableEvents": true
  },
  "webapi": {
    "bindAddress": "127.0.0.1:9090"
  },
  "dashboard": {
  	"auth": {
	  "scheme": "basic",
	  "username": "wasp",
	  "password": "wasp"
	},
    "bindAddress": "127.0.0.1:7000"
  },
  "peering":{
    "port": 4000,
    "netid": "wasphost:4000"
  },
  "nodeconn": {
    "address": "goshimmer-host:5000"
  },
  "nanomsg":{
    "port": 5550
  }
}
``` 

#### Peering settings
Wasp nodes connects to other Wasp peers to form committees. There's exactly one TCP connection between two Wasp nodes 
participating in the same committee(s). The node is using `peering.port` setting to specify the port for peering.

`peering.netid` must have form `host:port` where the `port` is equal to the setting of `peering.port`.
The `host` in the `peering.netid` must resolve to the machine where the node is running. 
The `netid` is used as an id of the node in the committee setting when deploying the smart contract: only `netid` 
can be used in the list of commitee nodes, not any equivalent for of the network location.

#### Goshimmer connection settings

`nodeconn.address` specifies the Goshimmer instance and port (exposed by the `WaspConn` plugin), where Wasp node connects. 

#### Publisher port

`nanomsg.port` specifies port for the Nanomsg even publisher. Wasp node publish important events happening 
in smart contracts, such as state transitions, incoming and processed requests and similar.  
Any Nanomsg client can subscribe to those messages.  
