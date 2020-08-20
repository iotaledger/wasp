# Demo cluster

This cluster starts with no configured smart contracts, and is used to test the
Wasp Client tool.

Steps:

1. Start a Goshimmer network. Eg, using the `docker-network-waspconn` tool available in the
   Goshimmer repository (`wasp` branch):

```
cd <goshimmer>/tools/docker-network-waspconn
./run.sh 2
```

2. Install the `wasp`, `wasp-client` and `waspt` commands:

```
go install . ./tools/wasp-client ./tools/cluster/waspt
```

3. Start the Wasp cluster in a console:

```
$ cd tools/cluster/demo
$ waspt init
$ GOSHIMMER_PROVIDED=1 waspt start
```
