package tests

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/require"
)

const file = "inccounter_bg.wasm"

const srcFile = "wasm/" + file

func TestWaspCLINoChains(t *testing.T) {
	w := newWaspCLITest(t)

	w.Run("init")
	if !*goShimmerUseProvidedNode {
		w.Run("set", "goshimmer.faucetPoWTarget", "0")
	}
	w.Run("request-funds")

	out := w.Run("address")

	ownerAddr := regexp.MustCompile(`(?m)Address:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	require.NotEmpty(t, ownerAddr)

	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func TestWaspCLI1Chain(t *testing.T) {
	w := newWaspCLITest(t)

	w.Run("init")
	w.Run("request-funds")

	alias := "chain1"

	committee, quorum := w.CommitteeConfig()

	// test chain deploy command
	w.Run("chain", "deploy", "--chain="+alias, committee, quorum)

	// test chain info command
	out := w.Run("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
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
	require.Regexp(t, `(?m)IOTA\s+1$`, out[3])

	// same test, this time calling the view function manually
	out = w.Run("chain", "call-view", "accounts", "balance", "string", "a", "agentid", agentID)
	out = w.Pipe(out, "decode", "color", "int")
	require.Regexp(t, `(?m)IOTA:\s+1$`, out[0])

	// test the chainlog
	out = w.Run("chain", "events", "root")
	require.Len(t, out, 1)
}

func TestWaspCLIContract(t *testing.T) {
	w := newWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "IOTA:10")

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

func findRequestIDInOutput(out []string) string {
	for _, line := range out {
		m := regexp.MustCompile(`(?m)#\d+ \(check result with: wasp-cli chain request (\w+)\)$`).FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		return m[1]
	}
	return ""
}

func TestWaspCLIBlockLog(t *testing.T) {
	w := newWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	out := w.Run("chain", "deposit", "IOTA:100")
	reqID := findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.Run("chain", "block")
	require.Equal(t, "Block index: 2", out[0])
	found := false
	for _, line := range out {
		if strings.Contains(line, reqID) {
			found = true
			break
		}
	}
	require.True(t, found)

	out = w.Run("chain", "block", "2")
	require.Equal(t, "Block index: 2", out[0])

	out = w.Run("chain", "request", reqID)
	found = false
	for _, line := range out {
		if strings.Contains(line, "ErrorStr: (empty)") {
			found = true
			break
		}
	}
	require.True(t, found)

	// try an unsuccessful request (missing params)
	out = w.Run("chain", "post-request", "root", "deployContract", "string", "foo", "string", "bar")
	reqID = findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.Run("chain", "request", reqID)

	found = false
	for _, line := range out {
		if strings.Contains(line, "ErrorStr: ") {
			found = true
			require.Regexp(t, `cannot decode`, line)
			break
		}
	}
	require.True(t, found)

	found = false
	for _, line := range out {
		if strings.Contains(line, "foo") {
			found = true
			require.Contains(t, line, hex.EncodeToString([]byte("bar")))
			break
		}
	}
	require.True(t, found)
}

func TestWaspCLIBlobContract(t *testing.T) {
	w := newWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "IOTA:10")

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

	blobHash := regexp.MustCompile(`(?m)([[:alnum:]]+)\s`).FindStringSubmatch(out[4])[1]
	require.NotEmpty(t, blobHash)
	t.Logf("Blob hash: %s", blobHash)

	// test chain show-blob command
	out = w.Run("chain", "show-blob", blobHash)
	out = w.Pipe(out, "decode", "string", blob.VarFieldProgramDescription, "string")
	require.Contains(t, out[0], description)
}

func TestWaspCLIMint(t *testing.T) {
	w := newWaspCLITest(t)

	w.Run("init")
	w.Run("request-funds")

	out := w.Run("mint", "1000")
	colorb58 := regexp.MustCompile(`(?m)Minted 1000 tokens of color ([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	color, err := ledgerstate.ColorFromBase58EncodedString(colorb58)
	require.NoError(t, err)

	outs, err := w.Cluster.GoshimmerClient().GetConfirmedOutputs(w.Address())
	require.NoError(t, err)
	found := false
	for _, out := range outs {
		if v, ok := out.Balances().Get(color); ok {
			require.EqualValues(t, 1000, v)
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestWaspCLIBalance(t *testing.T) {
	w := newWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	w.Run("mint", "1000")

	out := w.Run("balance")

	bals := map[string]uint64{}
	var mintedColor string
	for _, line := range out {
		m := regexp.MustCompile(`(?m)(\w+):\s+(\d+)$`).FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		if m[1] == "Total" {
			continue
		}
		v, err := strconv.Atoi(m[2])
		require.NoError(t, err)
		bals[m[1]] = uint64(v)
		if m[1] != "IOTA" {
			mintedColor = m[1]
		}
	}
	t.Logf("%+v", bals)
	require.Equal(t, 2, len(bals))
	require.EqualValues(t, utxodb.RequestFundsAmount-1000, bals["IOTA"])
	require.EqualValues(t, 1000, bals[mintedColor])
}

func TestWaspCLIRejoinChain(t *testing.T) {
	w := newWaspCLITest(t)

	w.Run("init")
	w.Run("request-funds")

	alias := "chain1"

	committee, quorum := w.CommitteeConfig()

	// test chain deploy command
	w.Run("chain", "deploy", "--chain="+alias, committee, quorum)

	// test chain info command
	out := w.Run("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
	require.NotEmpty(t, chainID)
	t.Logf("Chain ID: %s", chainID)

	// test chain list command
	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	// deactivate chain and check that the chain was deactivated
	w.Run("chain", "deactivate")
	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	chOut := strings.Fields(out[4])
	active, _ := strconv.ParseBool(chOut[1])
	require.False(t, active)

	// activate chain and check that it was activated
	w.Run("chain", "activate")
	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	chOut = strings.Fields(out[4])
	active, _ = strconv.ParseBool(chOut[1])
	require.True(t, active)
}
