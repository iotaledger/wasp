package tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

const file = "inccounter_bg.wasm"

const srcFile = "wasm/" + file

func TestWaspCLINoChains(t *testing.T) {
	w := testutil.NewWaspCLITest(t)

	w.Run("init")
	w.Run("request-funds")

	out := w.Run("address")

	ownerAddr := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1] //nolint:gocritic
	require.NotEmpty(t, ownerAddr)

	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func TestWaspCLI1Chain(t *testing.T) {
	w := testutil.NewWaspCLITest(t)

	w.Run("init")
	w.Run("request-funds")

	out := w.Run("address")
	ownerAddr := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1] //nolint:gocritic
	require.NotEmpty(t, ownerAddr)
	t.Logf("Owner address: %s", ownerAddr)

	alias := "chain1"

	committee, quorum := w.CommitteeConfig()

	// test chain deploy command
	w.Run("chain", "deploy", "--chain="+alias, committee, quorum)

	// test chain info command
	out = w.Run("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1] //nolint:gocritic
	require.NotEmpty(t, chainID)
	t.Logf("Chain ID: %s", chainID)

	// test chain list command
	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	// unnecessary, since it is the latest deployed chain
	w.Run("set", "chain", alias)

	// test chain list-contracts command
	out = w.Run("chain", "list-contracts")
	require.Regexp(t, `Total \d+ contracts`, out[0])

	// test chain list-accounts command
	out = w.Run("chain", "list-accounts")
	require.Contains(t, out[0], "Total 1 account(s)")
	agentID := strings.TrimSpace(out[4])
	require.NotEmpty(t, agentID)
	t.Logf("Agent ID: %s", agentID)

	// test chain balance command
	out = w.Run("chain", "balance", agentID)
	// check that the chain balance of owner is 1 IOTA
	require.Regexp(t, "(?m)IOTA[[:space:]]+1$", out[3])

	// same test, this time calling the view function manually
	out = w.Run("chain", "call-view", "accounts", "balance", "string", "a", "agentid", agentID)
	out = w.Pipe(out, "decode", "color", "int")
	require.Regexp(t, "(?m)IOTA:[[:space:]]+1$", out[0])

	// test the chainlog
	out = w.Run("chain", "log", "root")
	require.Len(t, out, 1)
}

func TestWaspCLIContract(t *testing.T) {
	w := testutil.NewWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

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
		require.Regexp(t, fmt.Sprintf("(?m)counter:[[:space:]]+%d$", n), out[0])
	}

	checkCounter(42)

	// test chain post-request command
	w.Run("chain", "post-request", name, "increment")
	checkCounter(43)

	// include a funds transfer
	w.Run("chain", "post-request", name, "increment", "--transfer=IOTA:10")
	checkCounter(44)

	// test off-ledger request
	w.Run("chain", "post-request", name, "increment", "--off-ledger")
	checkCounter(45)
}

func TestWaspCLIBlobContract(t *testing.T) {
	w := testutil.NewWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// test chain list-blobs command
	out := w.Run("chain", "list-blobs")
	require.Contains(t, out[0], "Total 0 blob(s)")

	vmtype := vmtypes.WasmTime
	description := "inccounter SC"
	w.CopyFile(srcFile)

	// test chain store-blob command
	w.Run(
		"chain", "store-blob",
		"string", blob.VarFieldProgramBinary, "file", file,
		"string", blob.VarFieldVMType, "string", vmtype,
		"string", blob.VarFieldProgramDescription, "string", description,
	)

	out = w.Run("chain", "list-blobs")
	require.Contains(t, out[0], "Total 1 blob(s)")

	blobHash := regexp.MustCompile(`(?m)([[:alnum:]]+)[[:space:]]`).FindStringSubmatch(out[4])[1] //nolint:gocritic
	require.NotEmpty(t, blobHash)
	t.Logf("Blob hash: %s", blobHash)

	// test chain show-blob command
	out = w.Run("chain", "show-blob", blobHash)
	out = w.Pipe(out, "decode", "string", blob.VarFieldProgramDescription, "string")
	require.Contains(t, out[0], description)
}

func TestWaspCLIBlobRegistry(t *testing.T) {
	w := testutil.NewWaspCLITest(t)

	// test that `blob has` returns false
	out := w.Run("blob", "has", hashing.RandomHash(nil).String())
	require.Contains(t, out[0], "false")

	// test `blob put` command
	w.CopyFile(srcFile)
	out = w.Run("blob", "put", file)
	blobHash := regexp.MustCompile(`(?m)Hash: ([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
	require.NotEmpty(t, blobHash)
	t.Logf("Blob hash: %s", blobHash)

	// test that `blob has` returns true
	out = w.Run("blob", "has", blobHash)
	require.Contains(t, out[0], "true")
}
