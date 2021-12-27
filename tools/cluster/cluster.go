// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/tools/cluster/mocknode"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	"golang.org/x/xerrors"
)

type Cluster struct {
	Name     string
	Config   *ClusterConfig
	Started  bool
	DataPath string

	goshimmer *mocknode.MockNode
	waspCmds  []*exec.Cmd
}

func New(name string, config *ClusterConfig) *Cluster {
	return &Cluster{
		Name:     name,
		Config:   config,
		waspCmds: make([]*exec.Cmd, config.Wasp.NumNodes),
	}
}

func (clu *Cluster) NewKeyPairWithFunds() (*ed25519.KeyPair, ledgerstate.Address, error) {
	key, addr := testkey.GenKeyAddr()
	err := clu.GoshimmerClient().RequestFunds(addr)
	return key, addr, err
}

func (clu *Cluster) GoshimmerClient() *goshimmer.Client {
	return goshimmer.NewClient(clu.Config.goshimmerAPIHost(), clu.Config.Goshimmer.FaucetPoWTarget)
}

func (clu *Cluster) TrustAll() error {
	allNodes := clu.Config.AllNodes()
	allPeers := make([]*model.PeeringTrustedNode, len(allNodes))
	for ni := range allNodes {
		var err error
		if allPeers[ni], err = clu.WaspClient(allNodes[ni]).GetPeeringSelf(); err != nil {
			return err
		}
	}
	for ni := range allNodes {
		for pi := range allPeers {
			var err error
			if _, err = clu.WaspClient(allNodes[ni]).PostPeeringTrusted(allPeers[pi].PubKey, allPeers[pi].NetID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (clu *Cluster) DeployDefaultChain() (*Chain, error) {
	committee := clu.Config.AllNodes()
	minQuorum := len(committee)/2 + 1
	quorum := len(committee) * 3 / 4
	if quorum < minQuorum {
		quorum = minQuorum
	}
	return clu.DeployChainWithDKG("Default chain", committee, committee, uint16(quorum))
}

func (clu *Cluster) InitDKG(committeeNodeCount int) ([]int, ledgerstate.Address, error) {
	cmt := util.MakeRange(0, committeeNodeCount)
	quorum := uint16((2*len(cmt))/3 + 1)

	address, err := clu.RunDKG(cmt, quorum)

	return cmt, address, err
}

func (clu *Cluster) RunDKG(committeeNodes []int, threshold uint16, timeout ...time.Duration) (ledgerstate.Address, error) {
	if threshold == 0 {
		threshold = (uint16(len(committeeNodes))*2)/3 + 1
	}
	apiHosts := clu.Config.APIHosts(committeeNodes)

	peerPubKeys := make([]string, 0)
	for _, i := range committeeNodes {
		peeringNodeInfo, err := clu.WaspClient(i).GetPeeringSelf()
		if err != nil {
			return nil, err
		}
		peerPubKeys = append(peerPubKeys, peeringNodeInfo.PubKey)
	}

	dkgInitiatorIndex := uint16(rand.Intn(len(apiHosts)))
	return apilib.RunDKG(apiHosts, peerPubKeys, threshold, dkgInitiatorIndex, timeout...)
}

func (clu *Cluster) DeployChainWithDKG(description string, allPeers, committeeNodes []int, quorum uint16) (*Chain, error) {
	stateAddr, err := clu.RunDKG(committeeNodes, quorum)
	if err != nil {
		return nil, err
	}
	return clu.DeployChain(description, allPeers, committeeNodes, quorum, stateAddr)
}

func (clu *Cluster) DeployChain(description string, allPeers, committeeNodes []int, quorum uint16, stateAddr ledgerstate.Address) (*Chain, error) {
	ownerSeed := seed.NewSeed()

	if len(allPeers) == 0 {
		allPeers = clu.Config.AllNodes()
	}

	chain := &Chain{
		Description:    description,
		OriginatorSeed: ownerSeed,
		AllPeers:       allPeers,
		CommitteeNodes: committeeNodes,
		Quorum:         quorum,
		Cluster:        clu,
	}

	address := chain.OriginatorAddress()

	err := clu.GoshimmerClient().RequestFunds(address)
	if err != nil {
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	committeePubKeys := make([]string, 0)
	for _, i := range chain.CommitteeNodes {
		peeringNode, err := clu.WaspClient(i).GetPeeringSelf()
		if err != nil {
			return nil, err
		}
		committeePubKeys = append(committeePubKeys, peeringNode.PubKey)
	}

	chainid, err := apilib.DeployChain(apilib.CreateChainParams{
		Node:              clu.GoshimmerClient(),
		CommitteeAPIHosts: chain.CommitteeAPIHosts(),
		CommitteePubKeys:  committeePubKeys,
		N:                 uint16(len(committeeNodes)),
		T:                 quorum,
		OriginatorKeyPair: chain.OriginatorKeyPair(),
		Description:       description,
		Textout:           os.Stdout,
		Prefix:            "[cluster] ",
	}, stateAddr)
	if err != nil {
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	chain.StateAddress = stateAddr
	chain.ChainID = chainid

	return chain, nil
}

func (clu *Cluster) IsGoshimmerUp() bool {
	return clu.goshimmer != nil
}

func (clu *Cluster) IsNodeUp(i int) bool {
	return clu.waspCmds[i] != nil
}

func (clu *Cluster) MultiClient() *multiclient.MultiClient {
	return multiclient.New(clu.Config.APIHosts())
}

func (clu *Cluster) WaspClient(nodeIndex int) *client.WaspClient {
	return client.NewWaspClient(clu.Config.APIHost(nodeIndex))
}

func waspNodeDataPath(dataPath string, i int) string {
	return path.Join(dataPath, fmt.Sprintf("wasp%d", i))
}

func fileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

type ModifyNodesConfigFn = func(nodeIndex int, configParams *templates.WaspConfigParams) *templates.WaspConfigParams

// InitDataPath initializes the cluster data directory (cluster.json + one subdirectory
// for each node).
func (clu *Cluster) InitDataPath(templatesPath, dataPath string, removeExisting bool, modifyConfig ModifyNodesConfigFn) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if exists {
		if !removeExisting {
			return xerrors.Errorf("%s directory exists", dataPath)
		}
		err = os.RemoveAll(dataPath)
		if err != nil {
			return err
		}
	}

	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		err = initNodeConfig(
			waspNodeDataPath(dataPath, i),
			path.Join(templatesPath, "wasp-config-template.json"),
			templates.WaspConfig,
			clu.Config.WaspConfigTemplateParams(i),
			i,
			modifyConfig,
		)
		if err != nil {
			return err
		}
	}

	clu.DataPath = dataPath

	return clu.Config.Save(dataPath)
}

func initNodeConfig(nodePath, configTemplatePath, defaultTemplate string, params *templates.WaspConfigParams, nodeIndex int, modifyConfig ModifyNodesConfigFn) error {
	exists, err := fileExists(configTemplatePath)
	if err != nil {
		return err
	}
	var configTmpl *template.Template
	if !exists {
		configTmpl, err = template.New("config").Parse(defaultTemplate)
	} else {
		configTmpl, err = template.ParseFiles(configTemplatePath)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Initializing %s\n", nodePath)

	err = os.MkdirAll(nodePath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(nodePath, "config.json"))
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	if modifyConfig != nil {
		params = modifyConfig(nodeIndex, params)
	}

	return configTmpl.Execute(f, params)
}

// Start launches all wasp nodes in the cluster, each running in its own directory
func (clu *Cluster) Start(dataPath string) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if !exists {
		return xerrors.Errorf("Data path %s does not exist", dataPath)
	}

	if err := clu.start(dataPath); err != nil {
		return err
	}

	if err := clu.TrustAll(); err != nil {
		return err
	}

	clu.Started = true
	return nil
}

func (clu *Cluster) start(dataPath string) error {
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", clu.Config.Wasp.NumNodes)

	if !clu.Config.Goshimmer.UseProvidedNode {
		clu.goshimmer = mocknode.Start(
			fmt.Sprintf(":%d", clu.Config.Goshimmer.TxStreamPort),
			fmt.Sprintf(":%d", clu.Config.Goshimmer.APIPort),
		)
		fmt.Printf("[cluster] started goshimmer node\n")
	}

	initOk := make(chan bool, clu.Config.Wasp.NumNodes)

	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		cmd, err := clu.startServer("wasp", waspNodeDataPath(dataPath, i), i, initOk)
		if err != nil {
			return err
		}
		clu.waspCmds[i] = cmd
	}

	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return xerrors.Errorf("Timeout starting wasp nodes\n")
		}
	}
	fmt.Printf("[cluster] started %d Wasp nodes\n", clu.Config.Wasp.NumNodes)
	return nil
}

func (clu *Cluster) KillNode(nodeIndex int) error {
	if nodeIndex >= len(clu.waspCmds) {
		return xerrors.Errorf("[cluster] Wasp node with index %d not found", nodeIndex)
	}

	process := clu.waspCmds[nodeIndex]

	if process != nil {
		err := process.Process.Kill()

		if err == nil {
			clu.waspCmds[nodeIndex] = nil
		}

		return err
	}

	return nil
}

func (clu *Cluster) RestartNode(nodeIndex int) error {
	if nodeIndex >= len(clu.waspCmds) {
		return xerrors.Errorf("[cluster] Wasp node with index %d not found", nodeIndex)
	}

	initOk := make(chan bool, 1)

	cmd, err := clu.startServer("wasp", waspNodeDataPath(clu.DataPath, nodeIndex), nodeIndex, initOk)
	if err != nil {
		return err
	}

	select {
	case <-initOk:
	case <-time.After(10 * time.Second):
		return xerrors.Errorf("Timeout starting wasp nodes\n")
	}

	clu.waspCmds[nodeIndex] = cmd

	return err
}

func (clu *Cluster) startServer(command, cwd string, nodeIndex int, initOk chan<- bool) (*exec.Cmd, error) {
	name := fmt.Sprintf("wasp %d", nodeIndex)
	cmd := exec.Command(command)
	cmd.Dir = cwd
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go scanLog(
		stderrPipe,
		func(line string) { fmt.Printf("[!%s] %s\n", name, line) },
	)
	go scanLog(
		stdoutPipe,
		func(line string) { fmt.Printf("[ %s] %s\n", name, line) },
	)
	go clu.waitForAPIReady(initOk, nodeIndex)

	return cmd, nil
}

const pollAPIInterval = 500 * time.Millisecond

// waits until API for a given node is ready
func (clu *Cluster) waitForAPIReady(initOk chan<- bool, nodeIndex int) {
	infoEndpointURL := fmt.Sprintf("http://localhost:%s%s", strconv.Itoa(clu.Config.APIPort(nodeIndex)), routes.Info())

	ticker := time.NewTicker(pollAPIInterval)
	go func() {
		for {
			<-ticker.C
			rsp, err := http.Get(infoEndpointURL) //nolint:gosec,noctx
			if err != nil {
				fmt.Printf("Error Polling node %d API ready status: %v\n", nodeIndex, err)
				continue
			}
			fmt.Printf("Polling node %d API ready status: %s %s\n", nodeIndex, infoEndpointURL, rsp.Status)
			//goland:noinspection GoUnhandledErrorResult
			rsp.Body.Close()
			if err == nil && rsp.StatusCode != 404 {
				initOk <- true
				ticker.Stop()
				return
			}
		}
	}()
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

func (clu *Cluster) stopGoshimmer() {
	if !clu.IsGoshimmerUp() {
		return
	}
	fmt.Printf("[cluster] Stopping Goshimmer MockNode\n")
	clu.goshimmer.Stop()
}

func (clu *Cluster) stopNode(nodeIndex int) {
	if !clu.IsNodeUp(nodeIndex) {
		return
	}
	fmt.Printf("[cluster] Sending shutdown to wasp node %d\n", nodeIndex)
	err := clu.WaspClient(nodeIndex).Shutdown()
	if err != nil {
		fmt.Println(err)
	}
}

func (clu *Cluster) StopNode(nodeIndex int) {
	clu.stopNode(nodeIndex)
	waitCmd(&clu.waspCmds[nodeIndex])
	fmt.Printf("[cluster] Node %d has been shut down\n", nodeIndex)
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (clu *Cluster) Stop() {
	clu.stopGoshimmer()
	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		clu.stopNode(i)
	}
	clu.Wait()
}

func (clu *Cluster) Wait() {
	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		waitCmd(&clu.waspCmds[i])
	}
}

func waitCmd(cmd **exec.Cmd) {
	if *cmd == nil {
		return
	}
	err := (*cmd).Wait()
	*cmd = nil
	if err != nil {
		fmt.Println(err)
	}
}

func (clu *Cluster) ActiveNodes() []int {
	nodes := make([]int, 0)
	for i := 0; i < clu.Config.Wasp.NumNodes; i++ {
		if !clu.IsNodeUp(i) {
			continue
		}
		nodes = append(nodes, i)
	}
	return nodes
}

func (clu *Cluster) StartMessageCounter(expectations map[string]int) (*MessageCounter, error) {
	return NewMessageCounter(clu, clu.Config.AllNodes(), expectations)
}

func (clu *Cluster) PostTransaction(tx *ledgerstate.Transaction) error {
	fmt.Printf("[cluster] posting request tx: %s\n", tx.ID().String())
	err := clu.GoshimmerClient().PostTransaction(tx)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.String(), err)
		return err
	}
	if err = clu.GoshimmerClient().WaitForConfirmation(tx.ID()); err != nil {
		fmt.Printf("[cluster] posting tx: %v\n", err)
		return err
	}
	fmt.Printf("[cluster] request tx confirmed: %s\n", tx.ID().String())
	return nil
}

func (clu *Cluster) VerifyAddressBalances(addr ledgerstate.Address, totalExpected uint64, expect colored.Balances, comment ...string) bool {
	allOuts, err := clu.GoshimmerClient().GetConfirmedOutputs(addr)
	if err != nil {
		fmt.Printf("[cluster] GetConfirmedOutputs error: %v\n", err)
		return false
	}
	byColor, total := colored.OutputBalancesByColor(allOuts)
	dumpStr, assertionOk := dumpBalancesByColor(byColor, expect)

	var totalExpectedStr string
	if totalExpected == total {
		totalExpectedStr = fmt.Sprintf("(%d) OK", totalExpected)
	} else {
		totalExpectedStr = fmt.Sprintf("(%d) FAIL", totalExpected)
		assertionOk = false
	}
	cmt := ""
	if len(comment) > 0 {
		cmt = " (" + comment[0] + ")"
	}
	fmt.Printf("[cluster] Inputs of the address %s%s\n      Total tokens: %d %s\n%s\n",
		addr.Base58(), cmt, total, totalExpectedStr, dumpStr)

	if !assertionOk {
		fmt.Printf("[cluster] assertion on balances failed\n")
	}
	return assertionOk
}

func dumpBalancesByColor(actual, expect colored.Balances) (string, bool) {
	assertionOk := true
	lst := make([]colored.Color, 0, len(expect))
	for col := range expect {
		lst = append(lst, col)
	}
	colored.Sort(lst)
	ret := ""
	for _, col := range lst {
		act := actual[col]
		isOk := "OK"
		if act != expect[col] {
			assertionOk = false
			isOk = "FAIL"
		}
		ret += fmt.Sprintf("         %s: %d (%d)   %s\n", col, act, expect[col], isOk)
	}
	lst = lst[:0]
	for col := range actual {
		if _, ok := expect[col]; !ok {
			lst = append(lst, col)
		}
	}
	if len(lst) == 0 {
		return ret, assertionOk
	}
	colored.Sort(lst)
	ret += "      Unexpected colors in actual outputs:\n"
	for _, col := range lst {
		ret += fmt.Sprintf("         %s %d\n", col.String(), actual[col])
	}
	return ret, assertionOk
}
