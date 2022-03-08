// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// privtangle is a cluster of HORNET nodes started for testing purposes.
package privtangle

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	iotagob "github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/xerrors"
)

const (
	nodePortOffsetRestAPI = iota
	nodePortOffsetPeering
	nodePortOffsetDashboard
	nodePortOffsetProfiling
	nodePortOffsetPrometheus
	nodePortOffsetFaucet
	nodePortOffsetMQTT
)

type PrivTangle struct {
	CooKeyPair1   *cryptolib.KeyPair
	CooKeyPair2   *cryptolib.KeyPair
	FaucetKeyPair *cryptolib.KeyPair
	NetworkID     string
	SnapshotInit  string
	ConfigFile    string
	BaseDir       string
	BasePort      int
	NodeCount     int
	NodeKeyPairs  []*cryptolib.KeyPair
	NodeCommands  []*exec.Cmd
	NodeStdouts   []io.ReadCloser
	NodeStderrs   []io.ReadCloser
	ctx           context.Context
	t             *testing.T
}

func Start(ctx context.Context, baseDir string, basePort, nodeCount int, t *testing.T) *PrivTangle {
	pt := PrivTangle{
		CooKeyPair1:   cryptolib.NewKeyPair(),
		CooKeyPair2:   cryptolib.NewKeyPair(),
		FaucetKeyPair: cryptolib.NewKeyPair(),
		NetworkID:     "private_tangle_wasp_cluster",
		SnapshotInit:  "snapshot.init",
		ConfigFile:    "config.json",
		BaseDir:       baseDir,
		BasePort:      basePort,
		NodeCount:     nodeCount,
		NodeKeyPairs:  make([]*cryptolib.KeyPair, nodeCount),
		NodeCommands:  make([]*exec.Cmd, nodeCount),
		NodeStdouts:   make([]io.ReadCloser, nodeCount),
		NodeStderrs:   make([]io.ReadCloser, nodeCount),
		ctx:           ctx,
		t:             t,
	}
	for i := range pt.NodeKeyPairs {
		pt.NodeKeyPairs[i] = cryptolib.NewKeyPair()
	}
	pt.logf("Starting in baseDir=%s with basePort=%d, nodeCount=%d ...", baseDir, basePort, nodeCount)

	if err := os.MkdirAll(pt.BaseDir, 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", pt.BaseDir, err))
	}

	pt.generateSnapshot()

	for i := range pt.NodeKeyPairs {
		pt.startNode(i)
		time.Sleep(500 * time.Millisecond) // TODO: Remove?
	}
	pt.logf("Starting... Done, all nodes started.")

	time.Sleep(500 * time.Millisecond) // Just to decrease noise in the logs.
	pt.WaitAllAlive()
	pt.logf("Starting... Done, all nodes alive.")

	return &pt
}

func (pt *PrivTangle) generateSnapshot() {
	if err := os.RemoveAll(filepath.Join(pt.BaseDir, pt.SnapshotInit)); err != nil {
		panic(xerrors.Errorf("Unable to remove old snapshot: %w", err))
	}
	snapGenArgs := []string{
		"tool", "snap-gen",
		fmt.Sprintf("--networkID=%s", pt.NetworkID),
		fmt.Sprintf("--mintAddress=%s", pt.FaucetKeyPair.GetPublicKey().AsEd25519Address().String()[2:]), // Dropping 0x from HEX.
		fmt.Sprintf("--outputPath=%s", filepath.Join(pt.BaseDir, pt.SnapshotInit)),
	}
	snapGen := exec.CommandContext(pt.ctx, "hornet", snapGenArgs...)
	if snapGenOut, err := snapGen.Output(); err != nil {
		panic(xerrors.Errorf("Unable to run snap-gen %s ==> %s: %w", snapGen.String(), snapGenOut, err))
	}
}

func (pt *PrivTangle) startNode(i int) {
	env := []string{}
	plugins := "MQTT,Debug,Prometheus,Indexer"
	if i == 0 {
		plugins = "Coordinator,MQTT,Debug,Prometheus,Faucet,Indexer"
		env = append(env,
			fmt.Sprintf("COO_PRV_KEYS=%s,%s",
				hex.EncodeToString(pt.CooKeyPair1.GetPrivateKey().AsBytes()),
				hex.EncodeToString(pt.CooKeyPair2.GetPrivateKey().AsBytes()),
			),
			fmt.Sprintf("FAUCET_PRV_KEY=%s",
				hex.EncodeToString(pt.FaucetKeyPair.GetPrivateKey().AsBytes()),
			),
		)
	}
	nodePath := filepath.Join(pt.BaseDir, fmt.Sprintf("node-%d", i))
	nodePathDB := "db"               // Relative from nodePath.
	nodeP2PStore := "p2pStore"       // Relative from nodePath.
	nodePathSnapshots := "snapshots" // Relative from nodePath.
	nodePathSnapFull := fmt.Sprintf("%s/full_snapshot.bin", nodePathSnapshots)
	nodePathSnapDelta := fmt.Sprintf("%s/delta_snapshot.bin", nodePathSnapshots)

	if err := os.RemoveAll(nodePath); err != nil {
		panic(xerrors.Errorf("Unable to delete dir %v: %w", nodePath, err))
	}
	if err := os.MkdirAll(nodePath, 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", nodePath, err))
	}
	if err := os.MkdirAll(filepath.Join(nodePath, nodePathDB), 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", nodePathDB, err))
	}
	if err := os.MkdirAll(filepath.Join(nodePath, nodePathSnapshots), 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", nodePathSnapshots, err))
	}
	if err := os.WriteFile(filepath.Join(nodePath, pt.ConfigFile), []byte(pt.configFileContent()), 0o600); err != nil {
		panic(xerrors.Errorf("Unable to create %s: %w", pt.ConfigFile, err))
	}

	snapContents, err := os.ReadFile(filepath.Join(pt.BaseDir, pt.SnapshotInit))
	if err != nil {
		panic(xerrors.Errorf("Unable to read initial snapshot : %w", err))
	}
	if err := os.WriteFile(filepath.Join(nodePath, nodePathSnapFull), snapContents, 0o600); err != nil {
		panic(xerrors.Errorf("Unable to copy the initial snapshot : %w", err))
	}

	args := []string{
		"-c", pt.ConfigFile,
		fmt.Sprintf("--protocol.networkID=%s", pt.NetworkID),
		fmt.Sprintf("--restAPI.bindAddress=0.0.0.0:%d", pt.NodePortRestAPI(i)),
		fmt.Sprintf("--dashboard.bindAddress=localhost:%d", pt.NodePortDashboard(i)),
		fmt.Sprintf("--db.path=%s", nodePathDB),
		fmt.Sprintf("--node.disablePlugins=%s", "Autopeering"),
		fmt.Sprintf("--node.enablePlugins=%s", plugins),
		fmt.Sprintf("--snapshots.fullPath=%s", nodePathSnapFull),
		fmt.Sprintf("--snapshots.deltaPath=%s", nodePathSnapDelta),
		fmt.Sprintf("--p2p.bindMultiAddresses=/ip4/127.0.0.1/tcp/%d", pt.NodePortPeering(i)),
		fmt.Sprintf("--profiling.bindAddress=127.0.0.1:%d", pt.NodePortProfiling(i)),
		fmt.Sprintf("--prometheus.bindAddress=localhost:%d", pt.NodePortPrometheus(i)),
		fmt.Sprintf("--prometheus.fileServiceDiscovery.target=localhost:%d", pt.NodePortPrometheus(i)),
		fmt.Sprintf("--faucet.website.bindAddress=localhost:%d", pt.NodePortFaucet(i)),
		fmt.Sprintf("--mqtt.bindAddress=localhost:%d", pt.NodePortMQTT(i)),
		fmt.Sprintf("--p2p.db.path=%s", nodeP2PStore),
		fmt.Sprintf("--p2p.identityPrivateKey=%s", hex.EncodeToString(pt.NodeKeyPairs[i].GetPrivateKey().AsBytes())),
		fmt.Sprintf("--p2p.peers=%s", strings.Join(pt.NodeMultiAddrsWoIndex(i), ",")),
	}
	if i == 0 {
		args = append(args,
			"--cooBootstrap",
			"--cooStartIndex", "0",
		)
	}
	hornetCmd := exec.CommandContext(pt.ctx, "hornet", args...)
	hornetCmd.Env = os.Environ()
	hornetCmd.Env = append(hornetCmd.Env, env...)
	hornetCmd.Dir = nodePath
	stdout, err := hornetCmd.StdoutPipe()
	if err != nil {
		panic(xerrors.Errorf("Unable to get stdout for HORNET[%d]: %w", i, err))
	}
	stderr, err := hornetCmd.StderrPipe()
	if err != nil {
		panic(xerrors.Errorf("Unable to get stdout for HORNET[%d]: %w", i, err))
	}
	pt.NodeCommands[i] = hornetCmd
	pt.NodeStdouts[i] = stdout
	pt.NodeStderrs[i] = stderr
	if err := hornetCmd.Start(); err != nil {
		panic(xerrors.Errorf("Cannot start hornet node[%d]: %w", i, err))
	}
}

func (pt *PrivTangle) Stop() {
	pt.logf("Stopping...")
	printOutput := func(i int) {
		stdoutData, _ := io.ReadAll(pt.NodeStdouts[i])
		stderrData, _ := io.ReadAll(pt.NodeStderrs[i])
		pt.logf("Node[%d] stdout=%s", i, stdoutData)
		pt.logf("Node[%d] stderr=%s", i, stderrData)
	}
	for i, c := range pt.NodeCommands {
		if err := c.Process.Signal(os.Interrupt); err != nil {
			printOutput(i)
			panic(xerrors.Errorf("Unable to send INT signal to Hornet node [%d]: %w", i, err))
		}
	}
	for i, c := range pt.NodeCommands {
		if err := c.Wait(); err != nil {
			printOutput(i)
			panic(xerrors.Errorf("Failed while waiting for a HORNET node [%d]: %w", i, err))
		}
		if !c.ProcessState.Success() {
			printOutput(i)
			panic(xerrors.Errorf("Hornet node [%d] failed: %w", i, c.ProcessState.String()))
		}
	}
	pt.logf("Stopping... Done")
}

func (pt *PrivTangle) NodeClient(i int) *nodeclient.Client {
	return nodeclient.New(
		fmt.Sprintf("http://localhost:%d", pt.NodePortRestAPI(i)),
		iotago.ZeroRentParas,
		nodeclient.WithIndexer(),
	)
}

// The health endpoint is not working for now.
// So we are using other endpoint for this check.
func (pt *PrivTangle) WaitAllAlive() {
	pt.waitAllHealthy()
	pt.waitAllReturnTips()
}

func (pt *PrivTangle) waitAllReturnTips() {
	for {
		allOK := true
		for i := range pt.NodeCommands {
			_, err := pt.NodeClient(i).Tips(pt.ctx)
			if err != nil && pt.t != nil {
				pt.logf("Node[%d] is not ready yet: %v", i, err)
			}
			if err != nil {
				allOK = false
			}
		}
		if allOK {
			break
		}
		pt.logf("Waiting to all nodes to startup.")
		time.Sleep(100 * time.Millisecond)
	}
}

func (pt *PrivTangle) waitAllHealthy() {
	for {
		allOK := true
		for i := range pt.NodeCommands {
			ok, err := pt.NodeClient(i).Health(pt.ctx)
			if err != nil && pt.t != nil {
				pt.logf("Failed to check Node[%d] health: %v", i, err)
			}
			if err != nil || !ok {
				allOK = false
			}
		}
		if allOK {
			break
		}
		if pt.t != nil {
			pt.logf("Waiting to all nodes to startup.")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (pt *PrivTangle) NodeMultiAddr(i int) string {
	stdPrivKey := pt.NodeKeyPairs[i].GetPrivateKey().AsStdKey()
	lppPrivKey, _, err := crypto.KeyPairFromStdKey(&stdPrivKey)
	if err != nil {
		panic(xerrors.Errorf("Unable to convert privKey to the standard priv key."))
	}
	tmpNode, err := libp2p.New(context.Background(), libp2p.Identity(lppPrivKey))
	if err != nil {
		panic(xerrors.Errorf("Unable to create temporary p2p node: %v", err))
	}
	peerIdentity := tmpNode.ID().String()
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", pt.NodePortPeering(i), peerIdentity)
}

func (pt *PrivTangle) NodeMultiAddrsWoIndex(x int) []string {
	acc := make([]string, 0)
	for i := range pt.NodeKeyPairs {
		if i == x {
			continue
		}
		acc = append(acc, pt.NodeMultiAddr(i))
	}
	return acc
}

func (pt *PrivTangle) NodePortRestAPI(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetRestAPI
}

func (pt *PrivTangle) NodePortPeering(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetPeering
}

func (pt *PrivTangle) NodePortDashboard(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetDashboard
}

func (pt *PrivTangle) NodePortProfiling(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetProfiling
}

func (pt *PrivTangle) NodePortPrometheus(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetPrometheus
}

func (pt *PrivTangle) NodePortFaucet(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetFaucet
}

func (pt *PrivTangle) NodePortMQTT(i int) int {
	return pt.BasePort + i*10 + nodePortOffsetMQTT
}

func (pt *PrivTangle) logf(msg string, args ...interface{}) {
	if pt.t != nil {
		pt.t.Logf("HORNET Cluster: "+msg, args...)
	}
}

// PostFaucetRequest makes a faucet request.
// It is here mostly as an example. Simple value TX is processed faster, and should be used in tests instead.
// Example:
//
//    PostFaucetRequest(context.Background(), cryptolib.Ed25519AddressFromPubKey(myKeyPair.PublicKey), iotago.PrefixTestnet)
//
func (pt *PrivTangle) PostFaucetRequest(ctx context.Context, recipientAddr iotago.Address, netPrefix iotago.NetworkPrefix) error {
	faucetReq := fmt.Sprintf("{\"address\":%q}", recipientAddr.Bech32(netPrefix))
	faucetURL := fmt.Sprintf("http://localhost:%d/api/plugins/faucet/v1/enqueue", pt.NodePortFaucet(0))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", faucetURL, bytes.NewReader([]byte(faucetReq)))
	if err != nil {
		return xerrors.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return xerrors.Errorf("unable to call faucet: %w", err)
	}
	if res.StatusCode == 202 {
		return nil
	}
	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return xerrors.Errorf("faucet status=%v, unable to read response body: %w", res.Status, err)
	}
	return xerrors.Errorf("faucet call failed, responPrivateKeyse status=%v, body=%v", res.Status, resBody)
}

// PostSimpleValueTX submits a simple value transfer TX.
// Can be used instead of the faucet API if the genesis key is known.
func (pt *PrivTangle) PostSimpleValueTX(
	ctx context.Context,
	nc *nodeclient.Client,
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) (*iotago.Message, error) {
	tx, err := pt.MakeSimpleValueTX(ctx, nc, sender, recipientAddr, amount)
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	//
	// Build a message and post it.
	txMsg, err := iotagob.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = nc.SubmitMessage(ctx, txMsg)
	if err != nil {
		return nil, xerrors.Errorf("failed to submit a tx message: %w", err)
	}
	return txMsg, nil
}

func (pt *PrivTangle) MakeSimpleValueTX(
	ctx context.Context,
	nc *nodeclient.Client,
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) (*iotago.Transaction, error) {
	senderAddr := sender.GetPublicKey().AsEd25519Address()
	senderOuts, err := pt.OutputMap(ctx, nc, senderAddr)
	if err != nil {
		return nil, xerrors.Errorf("failed to get address outputs: %w", err)
	}
	txBuilder := iotagob.NewTransactionBuilder(
		iotago.NetworkIDFromString(pt.NetworkID),
	)
	inputSum := uint64(0)
	for i, o := range senderOuts {
		if inputSum >= amount {
			break
		}
		oid := i
		out := o
		txBuilder = txBuilder.AddInput(&iotagob.ToBeSignedUTXOInput{
			Address:  senderAddr,
			OutputID: oid,
			Output:   out,
		})
		inputSum += out.Deposit()
	}
	if inputSum < amount {
		return nil, xerrors.Errorf("not enough funds, have=%v, need=%v", inputSum, amount)
	}
	txBuilder = txBuilder.AddOutput(&iotago.BasicOutput{
		Amount:     amount,
		Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: recipientAddr}},
	})
	if inputSum > amount {
		txBuilder = txBuilder.AddOutput(&iotago.BasicOutput{
			Amount:     inputSum - amount,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: senderAddr}},
		})
	}
	tx, err := txBuilder.Build(
		iotago.ZeroRentParas,
		sender.AsAddressSigner(),
	)
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	return tx, nil
}

func (pt *PrivTangle) OutputMap(ctx context.Context, node0 *nodeclient.Client, myAddress *iotago.Ed25519Address) (map[iotago.OutputID]iotago.Output, error) {
	res, err := node0.Indexer().Outputs(ctx, &nodeclient.OutputsQuery{
		AddressBech32: myAddress.Bech32(iotago.PrefixTestnet),
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to query address outputs: %w", err)
	}
	result := make(map[iotago.OutputID]iotago.Output)
	for res.Next() {
		outs, err := res.Outputs()
		if err != nil {
			return nil, xerrors.Errorf("failed to fetch address outputs: %w", err)
		}
		oids := res.Response.Items.MustOutputIDs()
		for i, o := range outs {
			result[oids[i]] = o
		}
	}
	return result, nil
}
