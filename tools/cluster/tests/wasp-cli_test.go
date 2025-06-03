package tests

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

func TestWaspAuth(t *testing.T) {
	w := newWaspCLITest(t, waspClusterOpts{
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			configParams.AuthScheme = "jwt"
			return configParams
		},
	})
	_, err := w.Run("chain", "list", "--node=0", "--node=0")
	require.Error(t, err)
	out := w.MustRun("auth", "login", "--node=0", "-u=wasp", "-p=wasp")
	require.Equal(t, "Successfully authenticated", out[1])
	out = w.MustRun("chain", "list", "--node=0", "--node=0")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func TestZeroGasFee(t *testing.T) {
	w := newWaspCLITest(t)
	const chainName = "chain1"
	committee, quorum := w.ArgCommitteeConfig(0)

	// test chain deploy command
	w.MustRun("request-funds")
	w.MustRun("request-funds", "--address-index=1")
	w.MustRun("chain", "deploy", "--chain="+chainName, committee, quorum, "--block-keep-amount=123", "--node=0")
	w.ActivateChainOnAllNodes(chainName, 0)

	w.MustRun("address")
	alternativeAddress := getAddress(w.MustRun("address"))
	w.MustRun("chain", "deposit", alternativeAddress, "base|2000000", "--node=0")
	w.MustRun("chain", "balance", alternativeAddress, "--node=0")
	outs, err := w.Run("chain", "info", "--node=0", "--node=0")
	require.NoError(t, err)
	require.Contains(t, outs, "Gas fee: gas units * (100/1)")
	_, err = w.Run("chain", "disable-feepolicy", "--node=0")
	require.NoError(t, err)
	outs, err = w.Run("chain", "info", "--node=0", "--node=0")
	require.NoError(t, err)
	require.Contains(t, outs, "Gas fee: gas units * (0/0)")

	t.Run("send arbitrary EVM tx without funds", func(t *testing.T) {
		ethPvtKey, _ := newEthereumAccount()
		sendDummyEVMTx(t, w, ethPvtKey)
	})

	t.Run("deposit directly to EVM", func(t *testing.T) {
		alternativeAddress := getAddress(w.MustRun("address", "--address-index=1"))
		w.MustRun("send-funds", "-s", alternativeAddress, "base|1000000")
		outs := w.MustRun("balance", "--address-index=1")
		_, eth := newEthereumAccount()
		w.MustRun("chain", "deposit", eth.String(), "base|1000000", "--node=0", "--address-index=1")
		outs = w.MustRun("chain", "balance", eth.String(), "--node=0")
		checkL2Balance(t, outs, 1000000)
	})
}

func checkL1Balance(t *testing.T, out []string, expected int) {
	t.Helper()
	r := regexp.MustCompile(`- 0x[0]{0,63}2+::iota::IOTA: (\d+)`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't check balance")
	}
	amount, err := strconv.Atoi(r[1])
	require.NoError(t, err)
	require.EqualValues(t, expected, amount)
}

func checkL2Balance(t *testing.T, out []string, expected int) {
	t.Helper()
	r := regexp.MustCompile(`.*(?i:base)\s*(?i:tokens)?:*\s*(\d+).*`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't check balance")
	}
	amount, err := strconv.Atoi(r[1])
	require.NoError(t, err)
	require.EqualValues(t, expected, amount)
}

func getAddress(out []string) string {
	r := regexp.MustCompile(`.*Address:\s+(\w*).*`).FindStringSubmatch(strings.Join(out, ""))
	if r == nil {
		panic("couldn't get address")
	}
	return r[1]
}

func TestWaspCLISendFunds(t *testing.T) { // passed
	w := newWaspCLITest(t)

	alternativeAddress := getAddress(w.MustRun("address", "--address-index=1"))

	w.MustRun("request-funds")
	w.MustRun("send-funds", alternativeAddress, "base|1000")
	outs := w.MustRun("balance", "--address-index=1")
	checkL1Balance(t, outs, 1000)
}

func TestWaspCLIDeposit(t *testing.T) { // passed
	w := newWaspCLITest(t)

	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("request-funds")
	w.MustRun("request-funds", "--address-index=1")
	outs := w.MustRun("balance")
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)

	// fund an alternative address to deposit from (so we can test the fees,
	// since --address-index=0 is the chain admin / default payoutAddress)
	alternativeAddress := getAddress(w.MustRun("address", "--address-index=1"))
	w.MustRun("send-funds", "-s", alternativeAddress, "base|10000000", "--address-index=1")

	outs = w.MustRun("balance")
	minFee := gas.DefaultFeePolicy().MinFee(nil, parameters.BaseTokenDecimals)

	outs, err := w.Run("chain", "info", "--node=0", "--node=0")
	require.NoError(t, err)
	t.Run("deposit directly to EVM", func(t *testing.T) {
		_, eth := newEthereumAccount()
		w.MustRun("chain", "deposit", "base|1000000", "--node=0")
		outs := w.MustRun("chain", "deposit", eth.String(), "base|10000", "--node=0", "--print-receipt")
		outs = w.MustRun("chain", "balance", eth.String(), "--node=0")
		checkL2Balance(t, outs, 10000)
	})

	t.Run("deposit to own account, then to EVM", func(t *testing.T) {
		w.MustRun("chain", "deposit", "base|1000000", "--node=0", "--address-index=1")
		outs = w.MustRun("chain", "balance", "--node=0", "--address-index=1")
		checkL2Balance(t, outs, 1000000-int(minFee))
		_, eth := newEthereumAccount()
		outs = w.MustRun("chain", "deposit", eth.String(), "base|1000000", "--node=0", "--address-index=1", "--print-receipt")
		re := regexp.MustCompile(`Gas fee charged:\s*(\d+)`)
		var l2GasFee int64
		for _, line := range outs {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				l2GasFee, err = strconv.ParseInt(matches[1], 10, 64)
				require.NoError(t, err)
			}
		}
		outs = w.MustRun("chain", "balance", eth.String(), "--node=0")
		checkL2Balance(t, outs, 1000000) // fee will be taken from the sender on-chain balance
		outs = w.MustRun("chain", "balance", "--node=0", "--address-index=1")
		checkL2Balance(t, outs, 1000000-int(minFee)-int(l2GasFee))
	})

	// t.Run("mint and deposit native tokens to an ethereum account", func(t *testing.T) {
	// 	_, eth := newEthereumAccount()
	// 	// create foundry
	// 	tokenScheme := codec.Encode[TokenScheme](&iotago.SimpleTokenScheme{
	// 		MintedTokens:  big.NewInt(0),
	// 		MeltedTokens:  big.NewInt(0),
	// 		MaximumSupply: big.NewInt(1000),
	// 	})

	// 	out := w.PostRequestGetReceipt(
	// 		"accounts", accounts.FuncNativeTokenCreate.Name,
	// 		"string", accounts.ParamTokenScheme, "bytes", iotago.EncodeHex(tokenScheme),
	// 		"-l", "base|1000000",
	// 		"-t", "base|100000000",
	// 		"string", accounts.ParamTokenName, "string", "TEST",
	// 		"string", accounts.ParamTokenTickerSymbol, "string", "TS",
	// 		"string", accounts.ParamTokenDecimals, "uint8", "8",
	// 		"--node=0",
	// 	)
	// 	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// 	// mint 2 native tokens
	// 	foundrySN := "1"
	// 	out = w.PostRequestGetReceipt(
	// 		"accounts", accounts.FuncNativeTokenModifySupply.Name,
	// 		"string", accounts.ParamFoundrySN, "uint32", foundrySN,
	// 		"string", accounts.ParamSupplyDeltaAbs, "bigint", "2",
	// 		"string", accounts.ParamDestroyTokens, "bool", "false",
	// 		"-l", "base|1000000",
	// 		"--off-ledger",
	// 		"--node=0",
	// 	)
	// 	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// 	out = w.MustRun("chain", "balance", "--node=0")
	// 	tokenID := ""
	// 	for _, line := range out {
	// 		if strings.Contains(line, "0x") {
	// 			tokenID = strings.Split(line, " ")[0]
	// 		}
	// 	}

	// 	// withdraw this token to the wasp-cli L1 address
	// 	out = w.PostRequestGetReceipt(
	// 		"accounts", accounts.FuncWithdraw.Name,
	// 		"-l", fmt.Sprintf("base|1000000, %s:2", tokenID),
	// 		"--off-ledger",
	// 		"--node=0",
	// 	)
	// 	require.Regexp(t, `.*Error: \(empty\).*`, strings.Join(out, ""))

	// 	// deposit the native token to the chain (to an ethereum account)
	// 	w.MustRun(
	// 		"chain", "deposit", eth.String(),
	// 		fmt.Sprintf("%s:1", tokenID),
	// 		"--adjust-storage-deposit",
	// 		"--node=0",
	// 	)
	// 	out = w.MustRun("chain", "balance", eth.String(), "--node=0")
	// 	require.Contains(t, strings.Join(out, ""), tokenID)

	// 	// deposit the native token to the chain (to the cli account)
	// 	w.MustRun(
	// 		"chain", "deposit",
	// 		fmt.Sprintf("%s:1", tokenID),
	// 		"--adjust-storage-deposit",
	// 		"--node=0",
	// 	)
	// 	out = w.MustRun("chain", "balance", "--node=0")
	// 	require.Contains(t, strings.Join(out, ""), tokenID)
	// 	// no token balance on L1
	// 	out = w.MustRun("balance")
	// 	require.NotContains(t, strings.Join(out, ""), tokenID)
	// })
}

func findRequestIDInOutput(out []string) string {
	for _, line := range out {
		m := regexp.MustCompile(`Request ID:\s*(0x[a-f0-9]+)`).FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		return m[1]
	}
	return ""
}

func TestWaspCLIBlockLog(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)

	w.MustRun("request-funds")
	out := w.MustRun("chain", "deposit", "base|100", "--node=0")
	reqID := findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "block", "--node=0")
	require.Equal(t, "Block index: 1", out[0])
	found := false
	for _, line := range out {
		if strings.Contains(line, reqID) {
			found = true
			break
		}
	}
	require.True(t, found)

	out = w.MustRun("chain", "block", "1", "--node=0")
	require.Equal(t, "Block index: 1", out[0])

	out = w.MustRun("chain", "request", reqID, "--node=0")
	t.Log(out)
	found = false
	for _, line := range out {
		if strings.Contains(line, "Request found in block 1") {
			found = true
			break
		}
	}
	require.True(t, found)

	// try an unsuccessful request (missing params)
	out = w.MustRun("chain", "post-request", "-s", "root", "deployContract", "string", "foo", "string", "bar", "--node=0")
	reqID = findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "request", reqID, "--node=0")
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
			require.Contains(t, line, cryptolib.EncodeHex([]byte("bar")))
			break
		}
	}
	require.True(t, found)
}

func TestWaspCLILongParam(t *testing.T) {
	w := newWaspCLITest(t)

	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)
	w.MustRun("chain", "deposit", "base|1000000", "--node=0")

	veryLongTokenName := strings.Repeat("A", 100_000)

	errMsg := "slice length is too long"
	defer func() {
		if r := recover(); r != nil {
			errStr := fmt.Sprintf("%s", r)
			if !strings.Contains(errStr, errMsg) {
				t.FailNow()
			}
		}
	}()

	w.CreateL2NativeToken(isc.SimpleTokenScheme{
		MaximumSupply: big.NewInt(1000000),
		MeltedTokens:  big.NewInt(0),
		MintedTokens:  big.NewInt(0),
	}, veryLongTokenName, "TST", 8)

	// The code should not reach here. CreateL2NativeToken should panic as the args are too long.
	// This is caught by the deferred recover.
	t.FailNow()
}

func TestWaspCLITrustListImport(t *testing.T) { // passed
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
			configParams.MetricsPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})

	// set cluster2/node0 to trust all nodes from cluster 1
	for _, nodeIndex := range w.Cluster.Config.AllNodes() {
		peeringInfoOutput := w.MustRun("peering", "info", fmt.Sprintf("--node=%d", nodeIndex))
		pubKey := regexp.MustCompile(`PubKey:\s+([[:alnum:]]+)$`).FindStringSubmatch(peeringInfoOutput[0])[1]
		peeringURL := regexp.MustCompile(`PeeringURL:\s+(.+)$`).FindStringSubmatch(peeringInfoOutput[1])[1]
		w2.MustRun("peering", "trust", fmt.Sprintf("x%d", nodeIndex), pubKey, peeringURL, "--node=0")
	}

	// import the trust from cluster2/node0 to cluster2/node1
	trustedFile0, err := os.CreateTemp("", "tmp-trusted-peers.*.json")
	require.NoError(t, err)
	defer os.Remove(trustedFile0.Name())
	w2.MustRun("peering", "export-trusted", "--node=0", "--peers=x0,x1,x2,x3", "-o="+trustedFile0.Name())
	w2.MustRun("peering", "import-trusted", trustedFile0.Name(), "--node=1")

	// export the trusted nodes from cluster2/node1 and assert the expected result
	trustedFile1, err := os.CreateTemp("", "tmp-trusted-peers.*.json")
	require.NoError(t, err)
	defer os.Remove(trustedFile1.Name())
	w2.MustRun("peering", "export-trusted", "--peers=x0,x1,x2,x3", "--node=1", "-o="+trustedFile1.Name())

	trustedBytes0, err := io.ReadAll(trustedFile0)
	require.NoError(t, err)
	trustedBytes1, err := io.ReadAll(trustedFile1)
	require.NoError(t, err)

	var trustedList0 []apiclient.PeeringNodeIdentityResponse
	require.NoError(t, json.Unmarshal(trustedBytes0, &trustedList0))

	var trustedList1 []apiclient.PeeringNodeIdentityResponse
	require.NoError(t, json.Unmarshal(trustedBytes1, &trustedList1))

	require.Equal(t, len(trustedList0), len(trustedList1))

	for _, trustedPeer := range trustedList0 {
		require.True(t,
			lo.ContainsBy(trustedList1, func(tp apiclient.PeeringNodeIdentityResponse) bool {
				return tp.PeeringURL == trustedPeer.PeeringURL && tp.PublicKey == trustedPeer.PublicKey && tp.IsTrusted == trustedPeer.IsTrusted
			}),
		)
	}
}

func TestWaspCLICantPeerWithSelf(t *testing.T) { // passed
	w := newWaspCLITest(t, waspClusterOpts{
		nNodes: 1,
	})

	peeringInfoOutput := w.MustRun("peering", "info")
	pubKey := regexp.MustCompile(`PubKey:\s+([[:alnum:]]+)$`).FindStringSubmatch(peeringInfoOutput[0])[1]

	require.Panics(
		t,
		func() {
			w.MustRun("peering", "trust", "self", pubKey, "0.0.0.0:4000")
		})
}

func TestWaspCLIListTrustDistrust(t *testing.T) { // passed
	w := newWaspCLITest(t)
	out := w.MustRun("peering", "list-trusted", "--node=0")
	// one of the entries starts with "1", meaning node 0 trusts node 1
	containsNode1 := func(output []string) bool {
		for _, line := range output {
			if strings.HasPrefix(line, "1") {
				return true
			}
		}
		return false
	}
	require.True(t, containsNode1(out))

	// distrust node 1
	w.MustRun("peering", "distrust", "1", "--node=0")

	// 1 is not included anymore in the trusted list
	out = w.MustRun("peering", "list-trusted", "--node=0")
	// one of the entries starts with "1", meaning node 0 trusts node 1
	require.False(t, containsNode1(out))
}

func TestWaspCLIMintNativeToken(t *testing.T) {
	t.Skip("TODO MintNativeToken")
	w := newWaspCLITest(t)

	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)
	w.MustRun("chain", "deposit", "base|100000000", "--node=0")

	out := w.MustRun(
		"chain", "create-native-token",
		"--max-supply=1000000",
		"--melted-tokens=0",
		"--minted-tokens=0",
		"--allowance=base|1000000",
		"--token-name=TEST",
		"--token-decimals=8",
		"--token-symbol=TS",
		"--node=0",
		"-o",
	)

	reqID := findRequestIDInOutput(out)
	require.NotEmpty(t, reqID)

	out = w.MustRun("chain", "request", reqID, "--node=0")
	require.Contains(t, strings.Join(out, "\n"), "Error: (empty)")
}

func sendDummyEVMTx(t *testing.T, w *WaspCLITest, ethPvtKey *ecdsa.PrivateKey) *types.Transaction {
	gasPrice := gas.DefaultFeePolicy().DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
	jsonRPCClient := NewEVMJSONRPClient(t, w.Cluster, 0)
	tx, err := types.SignTx(
		types.NewTransaction(0, common.Address{}, big.NewInt(123), 100000, gasPrice, []byte{}),
		EVMSigner(),
		ethPvtKey,
	)
	require.NoError(t, err)
	err = jsonRPCClient.SendTransaction(context.Background(), tx)
	require.NoError(t, err)
	return tx
}

func TestEVMISCReceipt(t *testing.T) {
	w := newWaspCLITest(t)
	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)
	ethPvtKey, _ := newEthereumAccount()
	w.MustRun("chain", "deposit", "base|100000000", "--node=0")
	// send some arbitrary EVM tx
	tx := sendDummyEVMTx(t, w, ethPvtKey)
	out := w.MustRun("chain", "request", tx.Hash().Hex(), "--node=0")
	require.Contains(t, out[0], "Request found in block")
}

func TestChangeGovernanceController(t *testing.T) {
	t.Skip("Cluster tests currently disabled")

	t.Fatalf("Implement gov controller change")

	w := newWaspCLITest(t)
	committee, quorum := w.ArgCommitteeConfig(0)
	w.MustRun("chain", "deploy", "--chain=chain1", committee, quorum, "--node=0")
	w.ActivateChainOnAllNodes("chain1", 0)

	// create the new controller
	_, newGovControllerAddr := testkey.GenKeyAddr()
	// change gov controller
	w.MustRun("chain", "change-gov-controller", newGovControllerAddr.String(), "--chain=chain1")

	t.Fatalf("Implement gov controller change")
}
