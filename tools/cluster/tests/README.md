# Cluster Tests

To run cluster tests, ensure you have installed the necessary dependencies in [INX](#inx).

Privtangle is used to build the L1 network for running the cluster tests. For more information about privtangle, please check [privtangle.go](packages/testutil/privtangle/privtangle.go).

After executing the cluster tool, the cluster tool will start wasp nodes for running the tests. The number of wasp nodes have been started depends on each tests. See [Troubleshooting](#troubleshooting) for the information of checking test logs.

## INX

INX dependencies are necessary to run cluster tests. This includes

* [hornet](https://github.com/iotaledger/hornet) (v2.0.0-rc.4)
  Use scripts under `scripts` folder to install.
* [inx-indexer](https://github.com/iotaledger/inx-indexer) (v1.0.0-rc.3)
* [inx-coordinator](https://github.com/iotaledger/inx-coordinator) (v1.0.0-rc.3)
* [inx-faucet](https://github.com/iotaledger/inx-faucet) (v1.0.0-rc.1)
  Require `git submodule update --init --recursive` before building.

See [privtangle.go](packages/testutil/privtangle/privtangle.go) you can get more information.

## Troubleshooting

Sometimes hornet, wasp or inx may not be successfully terminated in the last run. Therefore the ports are still occupied. In this situation, timeout panic may happen (if you set the `-timeout` when executing go test), when privtangle is still waiting hornet's response of healthy. The message could be `privtangle.go:527: HORNET Cluster: Waiting for all HORNET nodes to become healthy...`.

To solve the problem, simply using `pkill` to kill the previous instances.

```bash
pkill -9 "hornet|wasp|inx"
```

The logs of privtangle are stored in the temporary folder created by `os.TempDir()` which `$TMPDIR` in UNIX system.
Go to `$TMPDIR/privtangle`, you can see the logs for different nodes.
The exact location will be printed in log message if a privtangle is enabled.

An example print out is

```
wasp/tools/cluster/tests/privtangle.go:527: HORNET Cluster: Starting in baseDir=/var/folders/fj/99whzry17dxfk7hyv99md3740000gn/T/privtangle with basePort=16500, nodeCount=2 ...
```

Here `baseDir` is the location of logs.

So we can see the log files would be in the following file structure.

```
$TMPDIR
├── privtangle
│   └── ... (logs)
├── wasp-cluster
│   └── ... (logs)
└── ... (other folders for test logs)
```
