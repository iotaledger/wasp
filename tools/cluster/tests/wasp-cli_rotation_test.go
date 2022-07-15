package tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIExternalRotation(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	out := w.Run("chain", "deploy", "--chain=chain1", committee, quorum)
	chainID := regexp.MustCompile(`(.*)ChainID:\s*([a-zA-Z0-9_]*),`).FindStringSubmatch(out[len(out)-1])[2]

	vmtype := vmtypes.WasmTime
	name := "inccounter"
	w.CopyFile(srcFile)

	// test chain deploy-contract command
	w.Run("chain", "deploy-contract", vmtype, name, "inccounter SC", file,
		"string", "counter", "int64", "42",
	)

	checkCounter := func(wTest *WaspCLITest, n int) {
		// test chain call-view command
		out = wTest.Run("chain", "call-view", name, "getCounter")
		out = wTest.Pipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	checkCounter(w, 42)

	// init maintenance
	out = w.PostRequestGetReceipt("governance", "startMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// stop the initial cluster
	w.Cluster.Stop()

	// start a new wasp cluster
	w2 := newWaspCLITest(t, waspClusterOpts{
		dirName: "wasp-cluster-new-gov",
	})
	// run DKG on the new cluster, obtain the new state controller address
	out = w2.Run("chain", "rundkg")
	newStateControllerAddr := regexp.MustCompile(`(.*):\s*([a-zA-Z0-9_]*)$`).FindStringSubmatch(out[0])[2]

	w2.Run("chain", "add", "chain1", chainID)
	w2.Run("set", "chain", "chain1")

	// issue a governance rotatation via CLI
	out = w.Run("chain", "rotate", newStateControllerAddr)
	println(out)

	// activate the chain on the new nodes
	w2.Run("chain", "activate")

	// stop maintenance
	out = w2.PostRequestGetReceipt("governance", "stopMaintenance")
	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// chain still works

	w2.Run("chain", "post-request", name, "increment")
	checkCounter(w2, 43)
}
