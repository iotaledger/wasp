package tests

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const file = "inccounter_bg.wasm"

const srcFile = "wasm/" + file

func TestWaspCLINoChains(t *testing.T) {
	w := newWaspCLITest(t)

	out := w.Run("address")

	ownerAddr := regexp.MustCompile(`(?m)Address:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	require.NotEmpty(t, ownerAddr)

	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func TestWaspCLI1Chain(t *testing.T) {
	w := newWaspCLITest(t)

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
	require.Regexp(t, `Total \d+ contracts in chain .{64}`, out[0])

	// test chain list-accounts command
	out = w.Run("chain", "list-accounts")
	require.Contains(t, out[0], "Total 1 account(s)")
	agentID := strings.TrimSpace(out[4])
	require.NotEmpty(t, agentID)
	t.Logf("Agent ID: %s", agentID)

	// test chain balance command
	out = w.Run("chain", "balance", agentID)
	// check that the chain balance of owner is > 0
	r := regexp.MustCompile(`(?m)base\s+(\d+)$`).FindStringSubmatch(out[len(out)-1])
	require.Len(t, r, 2)
	bal, err := strconv.ParseInt(r[1], 10, 64)
	require.NoError(t, err)
	require.Positive(t, bal)

	// same test, this time calling the view function manually
	out = w.Run("chain", "call-view", "accounts", "balance", "string", "a", "agentid", agentID)
	out = w.Pipe(out, "decode", "bytes", "bigint")

	r = regexp.MustCompile(`(?m):\s+(\d+)$`).FindStringSubmatch(out[0])
	bal2, err := strconv.ParseInt(r[1], 10, 64)
	require.NoError(t, err)
	require.EqualValues(t, bal, bal2)

	// test the chainlog
	out = w.Run("chain", "events", "root")
	require.Len(t, out, 1)
}

func checkBalance(t *testing.T, out []string, expected int) {
	amount := 0
	r := regexp.MustCompile(`(?i).*base( tokens:)?\s+(\d+).*`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't check balance")
	}
	amount, err := strconv.Atoi(r[2])
	require.NoError(t, err)
	require.GreaterOrEqual(t, amount, expected)
}

func getAddress(out []string) string {
	r := regexp.MustCompile(`.*Address:\s+(\w*).*`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't get address")
	}
	return r[1]
}

func TestWaspCLISendFunds(t *testing.T) {
	w := newWaspCLITest(t)

	alternativeAddress := getAddress(w.Run("address", "--address-index=1"))

	w.Run("send-funds", "-s", alternativeAddress, "base:1000000")
	checkBalance(t, w.Run("balance", "--address-index=1"), 1000000)
}

func TestWaspCLIDeposit(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	t.Run("deposit to own account", func(t *testing.T) {
		w.Run("chain", "deposit", "base:1000000")
		checkBalance(t, w.Run("chain", "balance"), 1000000)
	})

	t.Run("deposit to ethereum account", func(t *testing.T) {
		_, eth := newEthereumAccount()
		w.Run("chain", "deposit", eth.String(), "base:1000000")
		checkBalance(t, w.Run("chain", "balance", eth.String()), 1000000-100) //-100 for the fee
	})
}

func TestWaspCLIContract(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "base:10000000")

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
	w.Run("chain", "post-request", "-s", name, "increment")
	checkCounter(43)

	// include a funds transfer
	w.Run("chain", "post-request", "-s", name, "increment", "--transfer=base:10000000")
	checkCounter(44)

	// test off-ledger request
	w.Run("chain", "post-request", "-s", name, "increment", "--off-ledger")
	checkCounter(45)

	// include an allowance transfer
	w.Run("chain", "post-request", "-s", name, "increment", "--transfer=base:10000000", "--allowance=base:10000000")
	checkCounter(46)
}

func findRequestIDInOutput(out []string) string {
	for _, line := range out {
		m := regexp.MustCompile(`(?m)#\d+ \(check result with: wasp-cli chain request ([-\w]+)\)$`).FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		return m[1]
	}
	return ""
}

func TestWaspCLIBlockLog(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	out := w.Run("chain", "deposit", "base:100")
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
	t.Log(out)
	found = false
	for _, line := range out {
		if strings.Contains(line, "Error: (empty)") {
			found = true
			break
		}
	}
	require.True(t, found)

	// try an unsuccessful request (missing params)
	out = w.Run("chain", "post-request", "-s", "root", "deployContract", "string", "foo", "string", "bar")
	reqID = findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.Run("chain", "request", reqID)

	found = false
	for _, line := range out {
		if strings.Contains(line, "Error: ") {
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
			require.Contains(t, line, hexutil.Encode([]byte("bar")))
			break
		}
	}
	require.True(t, found)
}

func TestWaspCLIBlobContract(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.Run("chain", "deposit", "base:10")

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

func TestWaspCLIRejoinChain(t *testing.T) {
	w := newWaspCLITest(t)

	// make sure deploying with a bad quorum breaks
	require.Panics(
		t,
		func() {
			w.Run("chain", "deploy", "--chain=chain1", "--committee=0,1,2,3,4,5", "--quorum=4")
		})

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
