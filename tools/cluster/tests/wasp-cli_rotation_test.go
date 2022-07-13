package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIRotation(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "iota:10")

	vmtype := vmtypes.WasmTime
	name := "inccounter"
	description := "inccounter SC"
	w.CopyFile(srcFile)

	// test chain deploy-contract command
	w.Run("chain", "deploy-contract", vmtype, name, description, file,
		"string", "counter", "int64", "42",
	)

	out := w.Run("chain", "list-contracts")
	found := false
	for _, s := range out {
		if strings.Contains(s, name) {
			found = true
			break
		}
	}
	require.True(t, found)

	checkCounter := func(n int) {
		// test chain call-view command
		out = w.Run("chain", "call-view", name, "getCounter")
		out = w.Pipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	checkCounter(42)

	// init maintenance

	// start a new wasp node

	// issue a governance rotatation via CLI

	// stop maintenance

	// update cli api address to the new node

	// chain still works

	w.Run("chain", "post-request", name, "increment")
	checkCounter(43)
}
