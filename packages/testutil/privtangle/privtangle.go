// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// privtangle is a cluster of HORNET nodes started for testing purposes.
package privtangle

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/xerrors"
)

// requires hornet, inx-mqtt and inx-indexer binaries to be in PATH
// https://github.com/gohornet/hornet
// https://github.com/gohornet/inx-mqtt
// https://github.com/gohornet/inx-indexer

type LogFunc func(format string, args ...interface{})

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
	ctx           context.Context
	logfunc       LogFunc
}

func Start(ctx context.Context, baseDir string, basePort, nodeCount int, logfunc LogFunc) *PrivTangle {
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
		ctx:           ctx,
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

	pt.WaitAllAlive()
	pt.logf("Starting... Done, all nodes alive.")

	for i := range pt.NodeKeyPairs {
		pt.startIndexer(i)
		pt.startMqtt(i)
	}
	pt.WaitInxPlugins()

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
	plugins := "INX,Debug,Prometheus"
	if i == 0 {
		plugins = "Coordinator,INX,Debug,Prometheus,Faucet"
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
		fmt.Sprintf("--inx.bindAddress=localhost:%d", pt.NodePortINX(i)),
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
	// kill hornet cmd if the go test process is killed
	util.TerminateCmdWhenTestStops(hornetCmd)

	hornetCmd.Env = os.Environ()
	hornetCmd.Env = append(hornetCmd.Env, env...)
	hornetCmd.Dir = nodePath
	pt.NodeCommands[i] = hornetCmd

	writeOutputToFiles(nodePath, hornetCmd)

	if err := hornetCmd.Start(); err != nil {
		panic(xerrors.Errorf("Cannot start hornet node[%d]: %w", i, err))
	}
}

func (pt *PrivTangle) startIndexer(i int) {
	indexerPath := filepath.Join(pt.BaseDir, fmt.Sprintf("node-%d", i), "inx-indexer")

	if err := os.MkdirAll(indexerPath, 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", indexerPath, err))
	}

	if err := os.WriteFile(filepath.Join(indexerPath, pt.ConfigFile), []byte(pt.configFileContent()), 0o600); err != nil {
		panic(xerrors.Errorf("Unable to create %s: %w", pt.ConfigFile, err))
	}

	args := []string{
		fmt.Sprintf("--inx.address=0.0.0.0:%d", pt.NodePortINX(i)),
		fmt.Sprintf("--indexer.bindAddress=0.0.0.0:%d", pt.NodePortIndexer(i)),
	}

	indexerCmd := exec.CommandContext(pt.ctx, "inx-indexer", args...)
	// kill indexer cmd if the go test process is killed
	util.TerminateCmdWhenTestStops(indexerCmd)

	indexerCmd.Dir = indexerPath

	writeOutputToFiles(indexerPath, indexerCmd)

	if err := indexerCmd.Start(); err != nil {
		panic(xerrors.Errorf("Cannot start indexer [%d]: %w", i, err))
	}
}

func (pt *PrivTangle) startMqtt(i int) {
	indexerPath := filepath.Join(pt.BaseDir, fmt.Sprintf("node-%d", i), "inx-mqtt")

	if err := os.MkdirAll(indexerPath, 0o755); err != nil {
		panic(xerrors.Errorf("Unable to create dir %v: %w", indexerPath, err))
	}

	if err := os.WriteFile(filepath.Join(indexerPath, pt.ConfigFile), []byte(pt.configFileContent()), 0o600); err != nil {
		panic(xerrors.Errorf("Unable to create %s: %w", pt.ConfigFile, err))
	}

	args := []string{
		fmt.Sprintf("--inx.address=0.0.0.0:%d", pt.NodePortINX(i)),
		fmt.Sprintf("--mqtt.bindAddress=localhost:%d", pt.NodePortMQTT(i)),
		fmt.Sprintf("--mqtt.wsPort=%d", pt.NodePortMQTTWebSocket(i)),
	}

	mqttCmd := exec.CommandContext(pt.ctx, "inx-mqtt", args...)
	// kill indexer cmd if the go test process is killed
	util.TerminateCmdWhenTestStops(mqttCmd)

	mqttCmd.Dir = indexerPath

	writeOutputToFiles(indexerPath, mqttCmd)

	if err := mqttCmd.Start(); err != nil {
		panic(xerrors.Errorf("Cannot start indexer [%d]: %w", i, err))
	}
}

func (pt *PrivTangle) Stop() {
	pt.logf("Stopping...")
	for i, c := range pt.NodeCommands {
		if err := c.Process.Signal(os.Interrupt); err != nil {
			panic(xerrors.Errorf("Unable to send INT signal to Hornet node [%d]: %w", i, err))
		}
	}
	for i, c := range pt.NodeCommands {
		if err := c.Wait(); err != nil {
			panic(xerrors.Errorf("Failed while waiting for a HORNET node [%d]: %w", i, err))
		}
		if !c.ProcessState.Success() {
			panic(xerrors.Errorf("Hornet node [%d] failed: %w", i, c.ProcessState.String()))
		}
	}
	pt.logf("Stopping... Done")
}

func (pt *PrivTangle) nodeClient(i int) *nodeclient.Client {
	return nodeclient.New(fmt.Sprintf("http://localhost:%d", pt.NodePortRestAPI(i)))
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
			_, err := pt.nodeClient(i).Tips(pt.ctx)
			if err != nil {
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
			ok, err := pt.nodeClient(i).Health(pt.ctx)
			if err != nil {
				pt.logf("Failed to check Node[%d] health: %v", i, err)
			}
			if err != nil || !ok {
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

func (pt *PrivTangle) WaitInxPlugins() {
	for {
		allOK := true
		for i := range pt.NodeCommands {
			// indexer
			_, err := pt.nodeClient(i).Indexer(pt.ctx)
			if err != nil {
				allOK = false
				continue
			}
			// mqtt
			_, err = pt.nodeClient(i).EventAPI(pt.ctx)
			if err != nil {
				allOK = false
				continue
			}
		}
		if allOK {
			return
		}
		pt.logf("Waiting to all nodes INX plugings to startup.")
		time.Sleep(100 * time.Millisecond)
	}
}

func (pt *PrivTangle) NodeMultiAddr(i int) string {
	stdPrivKey := pt.NodeKeyPairs[i].GetPrivateKey().AsStdKey()
	lppPrivKey, _, err := crypto.KeyPairFromStdKey(&stdPrivKey)
	if err != nil {
		panic(xerrors.Errorf("Unable to convert privKey to the standard priv key."))
	}
	tmpNode, err := libp2p.New(libp2p.Identity(lppPrivKey))
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
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetRestAPI
}

func (pt *PrivTangle) NodePortPeering(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetPeering
}

func (pt *PrivTangle) NodePortDashboard(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetDashboard
}

func (pt *PrivTangle) NodePortProfiling(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetProfiling
}

func (pt *PrivTangle) NodePortPrometheus(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetPrometheus
}

func (pt *PrivTangle) NodePortFaucet(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetFaucet
}

func (pt *PrivTangle) NodePortMQTT(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetMQTT
}

func (pt *PrivTangle) NodePortMQTTWebSocket(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetMQTTWebSocket
}

func (pt *PrivTangle) NodePortIndexer(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetIndexer
}

func (pt *PrivTangle) NodePortINX(i int) int {
	return pt.BasePort + i*10 + privtangledefaults.NodePortOffsetINX
}

func (pt *PrivTangle) logf(msg string, args ...interface{}) {
	if pt.logfunc != nil {
		pt.logfunc("HORNET Cluster: "+msg, args...)
	}
}

func (pt *PrivTangle) L1Config(i ...int) nodeconn.L1Config {
	nodeIndex := 0
	if len(i) > 0 {
		nodeIndex = i[0]
	}
	return nodeconn.L1Config{
		Hostname:   "localhost",
		APIPort:    pt.NodePortRestAPI(nodeIndex),
		FaucetPort: pt.NodePortFaucet(nodeIndex),
		FaucetKey:  pt.FaucetKeyPair,
	}
}

func writeOutputToFiles(path string, cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(xerrors.Errorf("Unable to get stdout for HORNET [path: %s]: %w", path, err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(xerrors.Errorf("Unable to get stdout for HORNET[path: %s]: %w", path, err))
	}
	outFilePath := filepath.Join(path, "log.txt")
	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	errFilePath := filepath.Join(path, "errors.txt")
	errFile, err := os.Create(errFilePath)
	if err != nil {
		panic(err)
	}
	go scanLog(
		stderr,
		func(line string) {
			_, err := errFile.WriteString(fmt.Sprintln(line))
			if err != nil {
				panic(xerrors.Errorf("error writing to file %s: %w", errFilePath, err))
			}
		},
	)
	go scanLog(
		stdout,
		func(line string) {
			_, err := outFile.WriteString(fmt.Sprintln(line))
			if err != nil {
				panic(xerrors.Errorf("error writing to file %s: %w", outFilePath, err))
			}
		},
	)
}

func scanLog(reader io.Reader, hooks ...func(string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		for _, hook := range hooks {
			hook(line)
		}
	}
}
