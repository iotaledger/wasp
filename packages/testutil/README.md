# Testutil

## L1starter

Responsible for providing a L1 connection to integration tests.

This can either be:

* A local iota node which will start automatically by using Docker. 
  * Each tested package will receive its own iota node container, isolating each tested package properly.
* An external node such as the Alphanet

By default, the L1starter will always start a local node, this can be configured with a `.testconfig` file in the projects root.

```json
{
  "IS_LOCAL": false,
  "API_URL": "https://api.iota-rebased-alphanet.iota.cafe",
  "FAUCET_URL": "https://faucet.iota-rebased-alphanet.iota.cafe/gas"
}
```

Set `IS_LOCAL` to false, to provide an external node.

Set `IS_LOCAL` to true, to use a local node instance. 