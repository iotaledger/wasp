package tests

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIExternalRotation(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "iota:10")

	vmtype := vmtypes.WasmTime
	name := "inccounter"
	w.CopyFile(srcFile)

	// test chain deploy-contract command
	w.Run("chain", "deploy-contract", vmtype, name, "inccounter SC", file,
		"string", "counter", "int64", "42",
	)

	checkCounter := func(wTest *WaspCLITest, n int) {
		// test chain call-view command
		out := wTest.Run("chain", "call-view", name, "getCounter")
		out = wTest.Pipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	checkCounter(w, 42)

	// init maintenance

	// TODO

	// stop the initial cluster
	w.Cluster.Stop()

	// start a new wasp cluster
	w2 := newWaspCLITest(t, waspClusterOpts{
		dirName: "wasp-cluster-new-gov",
	})
	// run DKG on the new cluster, obtain the new state controller address
	out := w2.Run("chain", "rundkg")

	println(out)

	// issue a governance rotatation via CLI

	// activate the chain on the new nodes

	// stop maintenance

	// update cli api address to the new node

	// chain still works

	// w2.Run("chain", "post-request", name, "increment")
	// checkCounter(43)
}
