package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

const file = "inccounter_bg.wasm"

const srcFile = "wasm/" + file

func TestWaspCLINoChains(t *testing.T) {
	w := newWaspCLITest(t)

	out := w.MustRun("address")

	ownerAddr := regexp.MustCompile(`(?m)Address:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	require.NotEmpty(t, ownerAddr)

	out = w.MustRun("chain", "list")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func TestWaspCLI1Chain(t *testing.T) {
	w := newWaspCLITest(t)

	alias := "chain1"

	committee, quorum := w.CommitteeConfigArgs()

	// test chain deploy command
	w.MustRun("chain", "deploy", "--chain="+alias, committee, quorum)

	// test chain info command
	out := w.MustRun("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
	require.NotEmpty(t, chainID)
	t.Logf("Chain ID: %s", chainID)

	// test chain list command
	out = w.MustRun("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	// unnecessary, since it is the latest deployed chain
	w.MustRun("set", "defaultchain", alias)

	// test chain list-contracts command
	out = w.MustRun("chain", "list-contracts")
	require.Regexp(t, `Total \d+ contracts in chain .{64}`, out[0])

	// test chain list-accounts command
	out = w.MustRun("chain", "list-accounts")
	require.Contains(t, out[0], "Total 1 account(s)")
	agentID := strings.TrimSpace(out[4])
	require.NotEmpty(t, agentID)
	t.Logf("Agent ID: %s", agentID)

	// test chain balance command
	out = w.MustRun("chain", "balance", agentID)
	// check that the chain balance of owner is > 0
	r := regexp.MustCompile(`(?m)base\s+(\d+)$`).FindStringSubmatch(out[len(out)-1])
	require.Len(t, r, 2)
	bal, err := strconv.ParseInt(r[1], 10, 64)
	require.NoError(t, err)
	require.Positive(t, bal)

	// same test, this time calling the view function manually
	out = w.MustRun("chain", "call-view", "accounts", "balance", "string", "a", "agentid", agentID)
	out = w.MustPipe(out, "decode", "bytes", "bigint")

	r = regexp.MustCompile(`(?m):\s+(\d+)$`).FindStringSubmatch(out[0])
	bal2, err := strconv.ParseInt(r[1], 10, 64)
	require.NoError(t, err)
	require.EqualValues(t, bal, bal2)

	// test the chainlog
	out = w.MustRun("chain", "events", "root")
	require.Len(t, out, 1)
}

func checkBalance(t *testing.T, out []string, expected int) {
	amount := 0
	// regex example: base tokens 1000000
	//				  -----  ------token  amount-----  ------base   1364700
	r := regexp.MustCompile(`.*(?i:base)\s*(?i:tokens)?:*\s*(\d+).*`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't check balance")
	}
	amount, err := strconv.Atoi(r[1])
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

	alternativeAddress := getAddress(w.MustRun("address", "--address-index=1"))

	w.MustRun("send-funds", "-s", alternativeAddress, "base:1000000")
	checkBalance(t, w.MustRun("balance", "--address-index=1"), 1000000)
}

func TestWaspCLIDeposit(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfigArgs()
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum)

	t.Run("deposit to own account", func(t *testing.T) {
		w.MustRun("chain", "deposit", "base:1000000")
		checkBalance(t, w.MustRun("chain", "balance"), 1000000)
	})

	t.Run("deposit to ethereum account", func(t *testing.T) {
		_, eth := newEthereumAccount()
		w.MustRun("chain", "deposit", eth.String(), "base:1000000")
		checkBalance(t, w.MustRun("chain", "balance", eth.String()), 1000000-100) //-100 for the fee
	})

	t.Run("mint and deposit native tokens to an ethereum account", func(t *testing.T) {
		_, eth := newEthereumAccount()
		// create foundry
		tokenScheme := codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
			MintedTokens:  big.NewInt(10), // TODO this is not 0, where do these tokens go to? they don't show up in the wasp-cli balances...
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: big.NewInt(1000),
		})
		out := w.PostRequestGetReceipt(
			"accounts", accounts.FuncFoundryCreateNew.Name,
			"string", accounts.ParamTokenScheme, "bytes", iotago.EncodeHex(tokenScheme), "-l", "base:1000000",
			"-t", "base:100000000",
		)
		require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

		// mint 2 native tokens
		// TODO we know its the first foundry to be created by the chain, so SN must be 1,
		// but there is NO WAY to obtain the SN, this is a design flaw that must be fixed.
		foundrySN := "1"
		out = w.PostRequestGetReceipt(
			"accounts", accounts.FuncFoundryModifySupply.Name,
			"string", accounts.ParamFoundrySN, "uint32", foundrySN,
			"string", accounts.ParamSupplyDeltaAbs, "bigint", "2",
			"string", accounts.ParamDestroyTokens, "bool", "false",
			"-l", "base:1000000",
			"--off-ledger",
		)
		require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

		// withdraw this token to the wasp-cli L1 address
		out = w.MustRun("chain", "balance")
		tokenID := ""
		for _, line := range out {
			if strings.Contains(line, "0x") {
				tokenID = strings.Split(line, " ")[0]
			}
		}

		out = w.PostRequestGetReceipt(
			"accounts", accounts.FuncWithdraw.Name,
			"-l", fmt.Sprintf("base:1000000, %s:2", tokenID),
			"--off-ledger",
		)
		require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

		// deposit the native token to the chain (to an ethereum account)
		w.MustRun(
			"chain", "deposit", eth.String(),
			fmt.Sprintf("%s:1", tokenID),
			"--adjust-storage-deposit",
		)
		out = w.MustRun("chain", "balance", eth.String())
		require.Contains(t, strings.Join(out, ""), tokenID)

		// deposit the native token to the chain (to the cli account)
		w.MustRun(
			"chain", "deposit",
			fmt.Sprintf("%s:1", tokenID),
			"--adjust-storage-deposit",
		)
		out = w.MustRun("chain", "balance")
		require.Contains(t, strings.Join(out, ""), tokenID)
		// no token balance on L1
		out = w.MustRun("balance")
		require.NotContains(t, strings.Join(out, ""), tokenID)
	})
}

func TestWaspCLIContract(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfigArgs()
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.MustRun("chain", "deposit", "base:10000000")

	vmtype := vmtypes.WasmTime
	name := "inccounter"
	description := "inccounter SC"
	w.CopyFile(srcFile)

	// test chain deploy-contract command
	w.MustRun("chain", "deploy-contract", vmtype, name, description, file,
		"string", "counter", "int64", "42",
	)

	out := w.MustRun("chain", "list-contracts")
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
		out = w.MustRun("chain", "call-view", name, "getCounter")
		out = w.MustPipe(out, "decode", "string", "counter", "int")
		require.Regexp(t, fmt.Sprintf(`(?m)counter:\s+%d$`, n), out[0])
	}

	checkCounter(42)

	// test chain post-request command
	w.MustRun("chain", "post-request", "-s", name, "increment")
	checkCounter(43)

	// include a funds transfer
	w.MustRun("chain", "post-request", "-s", name, "increment", "--transfer=base:10000000")
	checkCounter(44)

	// test off-ledger request
	w.MustRun("chain", "post-request", "-s", name, "increment", "--off-ledger")
	checkCounter(45)

	// include an allowance transfer
	w.MustRun("chain", "post-request", "-s", name, "increment", "--transfer=base:10000000", "--allowance=base:10000000")
	checkCounter(46)
}

func findRequestIDInOutput(out []string) string {
	for _, line := range out {
		m := regexp.MustCompile(`(?m)\(check result with: wasp-cli chain request ([-\w]+)\)$`).FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		return m[1]
	}
	return ""
}

func TestWaspCLIBlockLog(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfigArgs()
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum)

	out := w.MustRun("chain", "deposit", "base:100")
	reqID := findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "block")
	require.Equal(t, "Block index: 2", out[0])
	found := false
	for _, line := range out {
		if strings.Contains(line, reqID) {
			found = true
			break
		}
	}
	require.True(t, found)

	out = w.MustRun("chain", "block", "2")
	require.Equal(t, "Block index: 2", out[0])

	out = w.MustRun("chain", "request", reqID)
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
	out = w.MustRun("chain", "post-request", "-s", "root", "deployContract", "string", "foo", "string", "bar")
	reqID = findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "request", reqID)

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
			require.Contains(t, line, iotago.EncodeHex([]byte("bar")))
			break
		}
	}
	require.True(t, found)
}

func TestWaspCLIBlobContract(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfigArgs()
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum)

	// for running off-ledger requests
	w.MustRun("chain", "deposit", "base:10")

	// test chain list-blobs command
	out := w.MustRun("chain", "list-blobs")
	require.Contains(t, out[0], "Total 0 blob(s)")

	vmtype := vmtypes.WasmTime
	description := "inccounter SC"
	w.CopyFile(srcFile)

	// test chain store-blob command
	w.MustRun(
		"chain", "store-blob",
		"string", blob.VarFieldProgramBinary, "file", file,
		"string", blob.VarFieldVMType, "string", vmtype,
		"string", blob.VarFieldProgramDescription, "string", description,
	)

	out = w.MustRun("chain", "list-blobs")
	require.Contains(t, out[0], "Total 1 blob(s)")

	blobHash := regexp.MustCompile(`(?m)([[:alnum:]]+)\s`).FindStringSubmatch(out[4])[1]
	require.NotEmpty(t, blobHash)
	t.Logf("Blob hash: %s", blobHash)

	// test chain show-blob command
	out = w.MustRun("chain", "show-blob", blobHash)
	out = w.MustPipe(out, "decode", "string", blob.VarFieldProgramDescription, "string")
	require.Contains(t, out[0], description)
}

func TestWaspCLIRejoinChain(t *testing.T) {
	w := newWaspCLITest(t)

	// make sure deploying with a bad quorum breaks
	require.Panics(
		t,
		func() {
			w.MustRun("chain", "deploy", "--chain=chain1", "--nodes=0,1,2,3,4,5", "--quorum=4")
		})

	alias := "chain1"

	committee, quorum := w.CommitteeConfigArgs()

	// test chain deploy command
	w.MustRun("chain", "deploy", "--chain="+alias, committee, quorum)

	// test chain info command
	out := w.MustRun("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
	require.NotEmpty(t, chainID)
	t.Logf("Chain ID: %s", chainID)

	// test chain list command
	out = w.MustRun("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	// deactivate chain and check that the chain was deactivated
	w.MustRun("chain", "deactivate", w.AllNodesArg())
	out = w.MustRun("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	chOut := strings.Fields(out[4])
	active, _ := strconv.ParseBool(chOut[1])
	require.False(t, active)

	// activate chain and check that it was activated
	w.MustRun("chain", "activate", w.AllNodesArg())
	out = w.MustRun("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	chOut = strings.Fields(out[4])
	active, _ = strconv.ParseBool(chOut[1])
	require.True(t, active)
}

func TestWaspCLILongParam(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.CommitteeConfigArgs()
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum)

	// create foundry
	w.MustRun(
		"chain", "post-request", "accounts", accounts.FuncFoundryCreateNew.Name,
		"string", accounts.ParamTokenScheme, "bytes", "0x00d107000000000000000000000000000000000000000000000000000000000000d207000000000000000000000000000000000000000000000000000000000000d30700000000000000000000000000000000000000000000000000000000000000", "-l", "base:1000000",
		"-t", "base:1000000",
	)

	veryLongTokenName := strings.Repeat("A", 100_000)
	out := w.MustRun(
		"chain", "post-request", "-o", "evm", "registerERC20NativeToken",
		"string", "fs", "uint32", "1",
		"string", "n", "string", veryLongTokenName,
		"string", "t", "string", "test_symbol",
		"string", "d", "uint8", "1",
	)

	reqID := findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "request", reqID)
	require.Contains(t, strings.Join(out, "\n"), "too long")
}

func TestWaspCLITrustListImport(t *testing.T) {
	w := newWaspCLITest(t, waspClusterOpts{
		nNodes:  4,
		dirName: "wasp-cluster-initial",
	})

	w2 := newWaspCLITest(t, waspClusterOpts{
		nNodes:  2,
		dirName: "wasp-cluster-new-gov",
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.DashboardPort += 100
			configParams.MetricsPort += 100
			configParams.NanomsgPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})
	// set cluster2/node0 to trust all nodes from cluster 1

	for _, nodeIndex := range w.Cluster.Config.AllNodes() {
		// equivalent of "wasp-cli peer info"
		peerInfo, _, err := w.Cluster.WaspClient(nodeIndex).NodeApi.
			GetPeeringIdentity(context.Background()).
			Execute()
		require.NoError(t, err)

		w2.MustRun("peering", "trust", peerInfo.PublicKey, peerInfo.NetId, "--node=0")
		require.NoError(t, err)
	}

	// import the trust from cluster2/node0 to cluster2/node1
	trustedOut := w2.MustRun("peering", "list-trusted", "--node=0", "--json")
	// create temporary file to be consumed by the import command
	file, err := ioutil.TempFile("", "tmp-trusted-peers.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString(trustedOut[0])
	w2.MustRun("peering", "import-trusted", file.Name(), "--node=1")

	trustedOut2 := w2.MustRun("peering", "list-trusted", "--node=1", "--json")
	println(trustedOut2)

	var trustedList1 []apiclient.PeeringNodeIdentityResponse
	require.NoError(t, json.Unmarshal([]byte(trustedOut[0]), &trustedList1))

	var trustedList2 []apiclient.PeeringNodeIdentityResponse
	require.NoError(t, json.Unmarshal([]byte(trustedOut2[0]), &trustedList2))

	require.Equal(t, len(trustedList1), len(trustedList2))

	for _, trustedPeer := range trustedList1 {
		require.True(t,
			lo.ContainsBy(trustedList2, func(tp apiclient.PeeringNodeIdentityResponse) bool {
				return tp.NetId == trustedPeer.NetId && tp.PublicKey == trustedPeer.PublicKey && tp.IsTrusted == trustedPeer.IsTrusted
			}),
		)
	}
}
