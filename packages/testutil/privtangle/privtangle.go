// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// privtangle is a cluster of HORNET nodes started for testing purposes.
package privtangle

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"

	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/packages/util"
)

// ===== Wasp dependencies ===== // DO NOT DELETE THIS LINE! It is needed for `make deps-versions` command
// requires hornet, and inx plugins binaries to be in PATH
// https://github.com/iotaledger/hornet (v2.0.0-rc.4)
// https://github.com/iotaledger/inx-indexer (v1.0.0-rc.3)
// https://github.com/iotaledger/inx-coordinator (v1.0.0-rc.3)
// https://github.com/iotaledger/inx-faucet (v1.0.0-rc.1) (requires `git submodule update --init --recursive` before building )
// ============================= // DO NOT DELETE THIS LINE! It is needed for `make deps-versions` command

type LogFunc func(format string, args ...interface{})

type PrivTangle struct {
	CooKeyPair1   *cryptolib.KeyPair
	CooKeyPair2   *cryptolib.KeyPair
	FaucetKeyPair *cryptolib.KeyPair
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
		SnapshotInit:  "snapshot.init",
		ConfigFile:    "config.json",
		BaseDir:       baseDir,
		BasePort:      basePort,
		NodeCount:     nodeCount,
		NodeKeyPairs:  make([]*cryptolib.KeyPair, nodeCount),
		NodeCommands:  make([]*exec.Cmd, nodeCount),
		ctx:           ctx,
		logfunc:       logfunc,
	}
	for i := range pt.NodeKeyPairs {
		pt.NodeKeyPairs[i] = cryptolib.NewKeyPair()
	}
	pt.logf("Starting in baseDir=%s with basePort=%d, nodeCount=%d ...", baseDir, basePort, nodeCount)

	if err := os.MkdirAll(pt.BaseDir, 0o755); err != nil {
		panic(fmt.Errorf("unable to create dir %v: %w", pt.BaseDir, err))
	}

	pt.generateSnapshot()
	pt.StartServers(true)

	return &pt
}

func (pt *PrivTangle) StartServers(deleteExisting bool) {
	ts := time.Now()
	pt.logf("Starting all HORNET nodes...")
	for i := range pt.NodeKeyPairs {
		pt.startNode(i, deleteExisting)
	}
	pt.logf("Starting all HORNET nodes... done! took: %v", time.Since(ts).Truncate(time.Millisecond))

	pt.waitAllReady(20 * time.Second)
	pt.startCoordinator(0, deleteExisting)
	pt.waitAllHealthy(20 * time.Second)
	pt.waitAllReturnTips(20 * time.Second)
	for i := range pt.NodeKeyPairs {
		pt.startIndexer(i)
	}
	pt.waitInxPluginsIndexer(20 * time.Second)
	pt.startFaucet(0) // faucet needs to be started after the indexer, otherwise it will take 1 milestone for the faucet get the correct balance
	pt.waitInxPluginsFaucet(20 * time.Second)
}

func (pt *PrivTangle) generateSnapshot() {
	snapshotPath := filepath.Join(pt.BaseDir, pt.SnapshotInit)
	if err := os.RemoveAll(snapshotPath); err != nil {
		panic(fmt.Errorf("unable to remove old snapshot: %w", err))
	}

	jsonConfigPath := filepath.Join(pt.BaseDir, "protocol_parameters.json")
	if err := os.WriteFile(jsonConfigPath, []byte(protocolParameters), 0o600); err != nil {
		panic(fmt.Errorf("unable to create %s: %w", pt.ConfigFile, err))
	}

	snapGenArgs := []string{
		"tool", "snap-gen",
		fmt.Sprintf("--protocolParametersPath=%s", jsonConfigPath),
		fmt.Sprintf("--mintAddress=%s", pt.FaucetKeyPair.GetPublicKey().AsEd25519Address().String()[2:]), // Dropping 0x from HEX.
		fmt.Sprintf("--outputPath=%s", snapshotPath),
	}
	snapGen := exec.CommandContext(pt.ctx, "hornet", snapGenArgs...)
	if snapGenOut, err := snapGen.Output(); err != nil {
		panic(fmt.Errorf("unable to run snap-gen %s ==> %s: %w", snapGen.String(), snapGenOut, err))
	}
}

func (pt *PrivTangle) startNode(i int, deleteExisting bool) {
	env := []string{}
	nodePath := filepath.Join(pt.BaseDir, fmt.Sprintf("node-%d", i))
	nodePathDB := "db"               // Relative from nodePath.
	nodeP2PStore := "p2pStore"       // Relative from nodePath.
	nodePathSnapshots := "snapshots" // Relative from nodePath.
	nodePathSnapFull := fmt.Sprintf("%s/full_snapshot.bin", nodePathSnapshots)
	nodePathSnapDelta := fmt.Sprintf("%s/delta_snapshot.bin", nodePathSnapshots)

	if deleteExisting {
		if err := os.RemoveAll(nodePath); err != nil {
			panic(fmt.Errorf("unable to delete dir %v: %w", nodePath, err))
		}
		if err := os.MkdirAll(nodePath, 0o755); err != nil {
			panic(fmt.Errorf("unable to create dir %v: %w", nodePath, err))
		}
		if err := os.MkdirAll(filepath.Join(nodePath, nodePathDB), 0o755); err != nil {
			panic(fmt.Errorf("unable to create dir %v: %w", nodePathDB, err))
		}
		if err := os.MkdirAll(filepath.Join(nodePath, nodePathSnapshots), 0o755); err != nil {
			panic(fmt.Errorf("unable to create dir %v: %w", nodePathSnapshots, err))
		}
		if err := os.WriteFile(filepath.Join(nodePath, pt.ConfigFile), []byte(pt.configFileContent()), 0o600); err != nil {
			panic(fmt.Errorf("unable to create %s: %w", pt.ConfigFile, err))
		}

		snapContents, err := os.ReadFile(filepath.Join(pt.BaseDir, pt.SnapshotInit))
		if err != nil {
			panic(fmt.Errorf("unable to read initial snapshot : %w", err))
		}
		if err := os.WriteFile(filepath.Join(nodePath, nodePathSnapFull), snapContents, 0o600); err != nil {
			panic(fmt.Errorf("unable to copy the initial snapshot : %w", err))
		}
	}

	args := []string{
		"-c", pt.ConfigFile,
		fmt.Sprintf("--restAPI.bindAddress=0.0.0.0:%d", pt.NodePortRestAPI(i)),
		fmt.Sprintf("--db.path=%s", nodePathDB),
		fmt.Sprintf("--inx.enabled=%s", "true"),
		fmt.Sprintf("--snapshots.fullPath=%s", nodePathSnapFull),
		fmt.Sprintf("--snapshots.deltaPath=%s", nodePathSnapDelta),
		fmt.Sprintf("--p2p.bindMultiAddresses=/ip4/127.0.0.1/tcp/%d", pt.NodePortPeering(i)),
		fmt.Sprintf("--profiling.bindAddress=127.0.0.1:%d", pt.NodePortProfiling(i)),
		fmt.Sprintf("--prometheus.bindAddress=localhost:%d", pt.NodePortPrometheus(i)),
		fmt.Sprintf("--prometheus.fileServiceDiscovery.target=localhost:%d", pt.NodePortPrometheus(i)),
		fmt.Sprintf("--inx.bindAddress=localhost:%d", pt.NodePortINX(i)),
		fmt.Sprintf("--p2p.db.path=%s", nodeP2PStore),
		// nodes almost start at the same time in the clustertests,
		// causing them to try to connect to each other at the same time, which ends up in "duplicated stream" errors.
		// we can only fix that by a reconnect.
		"--p2p.reconnectInterval=2s",
		fmt.Sprintf("--p2p.identityPrivateKey=%s", hex.EncodeToString(pt.NodeKeyPairs[i].GetPrivateKey().AsBytes())),
		fmt.Sprintf("--p2p.peers=%s", strings.Join(pt.NodeMultiAddrsWoIndex(i), ",")),
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
		panic(fmt.Errorf("cannot start hornet node[%d]: %w", i, err))
	}
}

func (pt *PrivTangle) startCoordinator(i int, deleteExisting bool) *exec.Cmd {
	env := []string{
		fmt.Sprintf("COO_PRV_KEYS=%s,%s",
			hex.EncodeToString(pt.CooKeyPair1.GetPrivateKey().AsBytes()),
			hex.EncodeToString(pt.CooKeyPair2.GetPrivateKey().AsBytes()),
		),
	}
	args := []string{
		"--cooStartIndex=0",
		"--coordinator.interval=100ms",
		"--coordinator.debugFakeMilestoneTimestamps=true",
		// no need to keep a backup of the coo milestones
		"--coordinator.blockBackups.enabled=false",
		fmt.Sprintf("--inx.address=0.0.0.0:%d", pt.NodePortINX(i)),
	}
	if deleteExisting {
		args = append(args, "--cooBootstrap")
	}
	return pt.startINXPlugin(i, "inx-coordinator", args, env)
}

func (pt *PrivTangle) startFaucet(i int) *exec.Cmd {
	env := []string{
		fmt.Sprintf("FAUCET_PRV_KEY=%s",
			hex.EncodeToString(pt.FaucetKeyPair.GetPrivateKey().AsBytes()),
		),
	}
	args := []string{
		fmt.Sprintf("--inx.address=0.0.0.0:%d", pt.NodePortINX(i)),
		fmt.Sprintf("--faucet.bindAddress=localhost:%d", pt.NodePortFaucet(i)),
		"--faucet.rateLimit.enabled=false",
	}
	return pt.startINXPlugin(i, "inx-faucet", args, env)
}

func (pt *PrivTangle) startIndexer(i int) *exec.Cmd {
	args := []string{
		fmt.Sprintf("--inx.address=0.0.0.0:%d", pt.NodePortINX(i)),
		fmt.Sprintf("--restAPI.bindAddress=0.0.0.0:%d", pt.NodePortIndexer(i)),
	}
	return pt.startINXPlugin(i, "inx-indexer", args, nil)
}

func (pt *PrivTangle) startINXPlugin(i int, plugin string, args, env []string) *exec.Cmd {
	path := filepath.Join(pt.BaseDir, fmt.Sprintf("node-%d", i), plugin)
	if err := os.MkdirAll(path, 0o755); err != nil {
		panic(fmt.Errorf("unable to create dir %v: %w", path, err))
	}

	cmd := exec.CommandContext(pt.ctx, plugin, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)
	cmd.Dir = path
	writeOutputToFiles(path, cmd)

	// kill cmd if the go test process is killed
	util.TerminateCmdWhenTestStops(cmd)

	if err := cmd.Start(); err != nil {
		panic(fmt.Errorf("cannot start %s [%d]: %w", plugin, i, err))
	}
	return cmd
}

func (pt *PrivTangle) Stop() {
	pt.logf("Stopping...")
	for i, c := range pt.NodeCommands {
		if err := c.Process.Signal(os.Interrupt); err != nil {
			panic(fmt.Errorf("unable to send INT signal to Hornet node [%d]: %w", i, err))
		}
	}
	for i, c := range pt.NodeCommands {
		if err := c.Wait(); err != nil {
			panic(fmt.Errorf("failed while waiting for a HORNET node [%d]: %w", i, err))
		}
		if !c.ProcessState.Success() {
			panic(fmt.Errorf("Hornet node [%d] failed: %s", i, c.ProcessState.String()))
		}
	}
	pt.logf("Stopping... Done")
}

func (pt *PrivTangle) nodeClient(i int) *nodeclient.Client {
	return nodeclient.New(fmt.Sprintf("http://localhost:%d", pt.NodePortRestAPI(i)))
}

func (pt *PrivTangle) waitAllReady(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(pt.ctx, timeout)
	defer cancel()

	ts := time.Now()
	pt.logf("Waiting for all HORNET nodes to start...")

	for {
		select {
		case <-ctx.Done():
			panic("nodes didn't start in time")

		default:
			allOK := true
			for i := range pt.NodeCommands {
				_, err := pt.nodeClient(i).Info(pt.ctx)
				if err != nil {
					allOK = false
				}
			}
			if allOK {
				pt.logf("Waiting for all HORNET nodes to start... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (pt *PrivTangle) waitAllHealthy(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(pt.ctx, timeout)
	defer cancel()

	ts := time.Now()
	pt.logf("Waiting for all HORNET nodes to become healthy...")

	for {
		select {
		case <-ctx.Done():
			panic("nodes didn't become healthy in time")

		default:
			allOK := true
			for i := range pt.NodeCommands {
				ok, err := pt.nodeClient(i).Health(pt.ctx)
				if err != nil || !ok {
					allOK = false
				}
			}
			if allOK {
				pt.logf("Waiting for all HORNET nodes to become healthy... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (pt *PrivTangle) waitAllReturnTips(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(pt.ctx, timeout)
	defer cancel()

	ts := time.Now()
	pt.logf("Waiting for all HORNET nodes to return tips...")

	for {
		select {
		case <-ctx.Done():
			panic("nodes didn't return tips in time")

		default:
			allOK := true
			for i := range pt.NodeCommands {
				_, err := pt.nodeClient(i).Tips(pt.ctx)
				if err != nil {
					allOK = false
				}
			}
			if allOK {
				pt.logf("Waiting for all HORNET nodes to return tips... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (pt *PrivTangle) waitInxPluginsIndexer(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(pt.ctx, timeout)
	defer cancel()

	ts := time.Now()
	pt.logf("Waiting for all nodes INX Indexer plugins to start...")

	for {
		select {
		case <-ctx.Done():
			panic("indexer didn't start in time")

		default:
			allOK := true
			for i := range pt.NodeCommands {
				_, err := pt.nodeClient(i).Indexer(pt.ctx)
				if err != nil {
					allOK = false
					continue
				}
			}
			if allOK {
				pt.logf("Waiting for all nodes INX Indexer plugins to start... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (pt *PrivTangle) waitInxPluginsFaucet(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(pt.ctx, timeout)
	defer cancel()

	pt.logf("Waiting for the faucet to start...")

	for {
		select {
		case <-ctx.Done():
			panic("faucet didn't start in time")

		default:
			err := pt.queryFaucetInfo()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return
		}
	}
}

type FaucetInfoResponse struct {
	Balance uint64 `json:"balance"`
}

func (pt *PrivTangle) queryFaucetInfo() error {
	faucetURL := fmt.Sprintf("http://localhost:%d/api/info", pt.NodePortFaucet(0))
	httpReq, err := http.NewRequestWithContext(pt.ctx, http.MethodGet, faucetURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unable to call faucet info endpoint: %w", err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error querying faucet info endpoint: HTTP %d, %s", res.StatusCode, resBody)
	}
	var parsedResp FaucetInfoResponse
	err = json.Unmarshal(resBody, &parsedResp)
	if err != nil {
		return err
	}
	if parsedResp.Balance == 0 {
		return errors.New("faucet has 0 balance")
	}
	return nil
}

func (pt *PrivTangle) NodeMultiAddr(i int) string {
	stdPrivKey := pt.NodeKeyPairs[i].GetPrivateKey().AsStdKey()
	lppPrivKey, _, err := crypto.KeyPairFromStdKey(&stdPrivKey)
	if err != nil {
		panic(errors.New("unable to convert privKey to the standard priv key"))
	}
	tmpNode, err := libp2p.New(libp2p.Identity(lppPrivKey))
	if err != nil {
		panic(fmt.Errorf("unable to create temporary p2p node: %w", err))
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
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetRestAPI
}

func (pt *PrivTangle) NodePortPeering(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetPeering
}

func (pt *PrivTangle) NodePortDashboard(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetDashboard
}

func (pt *PrivTangle) NodePortProfiling(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetProfiling
}

func (pt *PrivTangle) NodePortPrometheus(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetPrometheus
}

func (pt *PrivTangle) NodePortFaucet(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetFaucet
}

func (pt *PrivTangle) NodePortCoordinator(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetCoordinator
}

func (pt *PrivTangle) NodePortIndexer(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetIndexer
}

func (pt *PrivTangle) NodePortINX(i int) int {
	return pt.BasePort + i*100 + privtangledefaults.NodePortOffsetINX
}

func (pt *PrivTangle) logf(msg string, args ...interface{}) {
	if pt.logfunc != nil {
		pt.logfunc("HORNET Cluster: "+msg, args...)
	}
}

func (pt *PrivTangle) L1Config(i ...int) l1connection.Config {
	nodeIndex := 0
	if len(i) > 0 {
		nodeIndex = i[0]
	}
	return l1connection.Config{
		APIAddress:    fmt.Sprintf("http://localhost:%d", pt.NodePortRestAPI(nodeIndex)),
		INXAddress:    fmt.Sprintf("localhost:%d", pt.NodePortINX(nodeIndex)),
		FaucetAddress: fmt.Sprintf("http://localhost:%d", pt.NodePortFaucet(nodeIndex)),
		FaucetKey:     pt.FaucetKeyPair,
		UseRemotePoW:  false,
	}
}

func writeOutputToFiles(path string, cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Errorf("unable to get stdout for HORNET [path: %s]: %w", path, err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Errorf("unable to get stdout for HORNET[path: %s]: %w", path, err))
	}
	outFilePath := filepath.Join(path, "stdout.log")
	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	errFilePath := filepath.Join(path, "stderr.log")
	errFile, err := os.Create(errFilePath)
	if err != nil {
		panic(err)
	}
	go scanLog(
		stderr,
		func(line string) {
			_, err := errFile.WriteString(fmt.Sprintln(line))
			if err != nil {
				panic(fmt.Errorf("error writing to file %s: %w", errFilePath, err))
			}
		},
	)
	go scanLog(
		stdout,
		func(line string) {
			_, err := outFile.WriteString(fmt.Sprintln(line))
			if err != nil {
				panic(fmt.Errorf("error writing to file %s: %w", outFilePath, err))
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
