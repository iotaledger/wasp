// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type Cluster struct {
	Name             string
	Config           *ClusterConfig
	Started          bool
	DataPath         string
	ValidatorKeyPair *cryptolib.KeyPair // Default identity for validators, chain owners, etc.
	l1               l1connection.Client
	waspCmds         []*waspCmd
	t                *testing.T
	log              *logger.Logger
}

type waspCmd struct {
	cmd        *exec.Cmd
	logScanner sync.WaitGroup
}

func New(name string, config *ClusterConfig, dataPath string, t *testing.T, log *logger.Logger) *Cluster {
	if log == nil {
		if t == nil {
			panic("one of t or log must be set")
		}
		log = testlogger.NewLogger(t)
	}

	validatorKp := cryptolib.NewKeyPair()
	config.SetOwnerAddress(validatorKp.Address().Bech32(parameters.L1().Protocol.Bech32HRP))

	return &Cluster{
		Name:             name,
		Config:           config,
		ValidatorKeyPair: validatorKp,
		waspCmds:         make([]*waspCmd, len(config.Wasp)),
		t:                t,
		log:              log,
		l1:               l1connection.NewClient(config.L1, log),
		DataPath:         dataPath,
	}
}

func (clu *Cluster) Logf(format string, args ...any) {
	if clu.t != nil {
		clu.t.Logf(format, args...)
		return
	}
	clu.log.Infof(format, args...)
}

func (clu *Cluster) ValidatorAddress() iotago.Address {
	return clu.ValidatorKeyPair.Address()
}

func (clu *Cluster) NewKeyPairWithFunds() (*cryptolib.KeyPair, iotago.Address, error) {
	key, addr := testkey.GenKeyAddr()
	err := clu.RequestFunds(addr)
	return key, addr, err
}

func (clu *Cluster) RequestFunds(addr iotago.Address) error {
	return clu.l1.RequestFunds(addr)
}

func (clu *Cluster) L1Client() l1connection.Client {
	return clu.l1
}

func (clu *Cluster) AddTrustedNode(peerInfo apiclient.PeeringTrustRequest, onNodes ...[]int) error {
	nodes := clu.Config.AllNodes()
	if len(onNodes) > 0 {
		nodes = onNodes[0]
	}

	for ni := range nodes {
		var err error
		if _, err = clu.WaspClient(
			nodes[ni]).NodeApi.TrustPeer(context.Background()).PeeringTrustRequest(peerInfo).Execute(); err != nil {
			return err
		}
	}
	return nil
}

func (clu *Cluster) TrustAll() error {
	allNodes := clu.Config.AllNodes()
	allPeers := make([]*apiclient.PeeringNodeIdentityResponse, len(allNodes))
	for ni := range allNodes {
		var err error
		if allPeers[ni], _, err = clu.WaspClient(allNodes[ni]).NodeApi.GetPeeringIdentity(context.Background()).Execute(); err != nil {
			return err
		}
	}
	for ni := range allNodes {
		for pi := range allPeers {
			var err error
			if _, err = clu.WaspClient(
				allNodes[ni]).NodeApi.TrustPeer(context.Background()).PeeringTrustRequest(apiclient.PeeringTrustRequest{
				PublicKey: allPeers[pi].PublicKey,
				NetId:     allPeers[pi].NetId,
			}).Execute(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (clu *Cluster) DeployDefaultChain() (*Chain, error) {
	committee := clu.Config.AllNodes()
	maxFaulty := (len(committee) - 1) / 3
	minQuorum := len(committee) - maxFaulty
	quorum := len(committee) * 3 / 4
	if quorum < minQuorum {
		quorum = minQuorum
	}
	return clu.DeployChainWithDKG("Default chain", committee, committee, uint16(quorum))
}

func (clu *Cluster) InitDKG(committeeNodeCount int) ([]int, iotago.Address, error) {
	cmt := util.MakeRange(0, committeeNodeCount-1) // End is inclusive for some reason.
	quorum := uint16((2*len(cmt))/3 + 1)

	address, err := clu.RunDKG(cmt, quorum)

	return cmt, address, err
}

func (clu *Cluster) RunDKG(committeeNodes []int, threshold uint16, timeout ...time.Duration) (iotago.Address, error) {
	if threshold == 0 {
		threshold = (uint16(len(committeeNodes))*2)/3 + 1
	}
	apiHosts := clu.Config.APIHosts(committeeNodes)

	peerPubKeys := make([]string, 0)
	for _, i := range committeeNodes {
		peeringNodeInfo, _, err := clu.WaspClient(i).NodeApi.GetPeeringIdentity(context.Background()).Execute()
		if err != nil {
			return nil, err
		}

		peerPubKeys = append(peerPubKeys, peeringNodeInfo.PublicKey)
	}

	dkgInitiatorIndex := uint16(rand.Intn(len(apiHosts)))

	return apilib.RunDKG(clu.WaspClientFromHostName, apiHosts, peerPubKeys, threshold, dkgInitiatorIndex, timeout...)
}

func (clu *Cluster) DeployChainWithDKG(description string, allPeers, committeeNodes []int, quorum uint16) (*Chain, error) {
	stateAddr, err := clu.RunDKG(committeeNodes, quorum)
	if err != nil {
		return nil, err
	}
	return clu.DeployChain(description, allPeers, committeeNodes, quorum, stateAddr)
}

func (clu *Cluster) DeployChain(description string, allPeers, committeeNodes []int, quorum uint16, stateAddr iotago.Address) (*Chain, error) {
	if len(allPeers) == 0 {
		allPeers = clu.Config.AllNodes()
	}

	chain := &Chain{
		Description:       description,
		OriginatorKeyPair: clu.ValidatorKeyPair,
		AllPeers:          allPeers,
		CommitteeNodes:    committeeNodes,
		Quorum:            quorum,
		Cluster:           clu,
	}

	address := chain.OriginatorAddress()

	err := clu.RequestFunds(address)
	if err != nil {
		return nil, fmt.Errorf("DeployChain: %w", err)
	}

	committeePubKeys := make([]string, len(chain.CommitteeNodes))
	for i, nodeIndex := range chain.CommitteeNodes {
		peeringNode, _, err := clu.WaspClient(nodeIndex).NodeApi.GetPeeringIdentity(context.Background()).Execute()
		if err != nil {
			return nil, err
		}

		committeePubKeys[i] = peeringNode.PublicKey
	}

	chainID, err := apilib.DeployChain(clu.WaspClientFromHostName, apilib.CreateChainParams{
		Layer1Client:      clu.L1Client(),
		CommitteeAPIHosts: chain.CommitteeAPIHosts(),
		CommitteePubKeys:  committeePubKeys,
		N:                 uint16(len(committeeNodes)),
		T:                 quorum,
		OriginatorKeyPair: chain.OriginatorKeyPair,
		Description:       description,
		Textout:           os.Stdout,
		Prefix:            "[cluster] ",
	}, stateAddr, stateAddr)
	if err != nil {
		return nil, fmt.Errorf("DeployChain: %w", err)
	}

	chain.StateAddress = stateAddr
	chain.ChainID = chainID

	// After a rotation other nodes can become access nodes,
	// so we make all of the nodes possible access nodes.
	return chain, clu.addAllAccessNodes(chain, allPeers)
}

func (clu *Cluster) addAllAccessNodes(chain *Chain, accessNodes []int) error {
	//
	// Register all nodes as access nodes.
	addAccessNodesRequests := make([]*iotago.Transaction, len(accessNodes))
	for i, a := range accessNodes {
		tx, err := clu.AddAccessNode(a, chain)
		if err != nil {
			return err
		}
		addAccessNodesRequests[i] = tx
	}

	peers := multiclient.New(clu.WaspClientFromHostName, chain.CommitteeAPIHosts()) //.WithLogFunc(clu.t.Logf)

	for _, tx := range addAccessNodesRequests {
		// ---------- wait until the requests are processed in all committee nodes

		if _, err := peers.WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, tx, 30*time.Second); err != nil {
			return fmt.Errorf("WaitAddAccessNode: %w", err)
		}
	}

	scArgs := governance.NewChangeAccessNodesRequest()
	for _, a := range accessNodes {
		waspClient := clu.WaspClient(a)

		accessNodePeering, _, err := waspClient.NodeApi.GetPeeringIdentity(context.Background()).Execute()
		if err != nil {
			return err
		}

		accessNodePubKey, err := cryptolib.NewPublicKeyFromString(accessNodePeering.PublicKey)
		if err != nil {
			return err
		}
		scArgs.Accept(accessNodePubKey)
	}
	scParams := chainclient.
		NewPostRequestParams(scArgs.AsDict()).
		WithBaseTokens(1 * isc.Million)
	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	tx, err := govClient.PostRequest(governance.FuncChangeAccessNodes.Name, *scParams)
	if err != nil {
		return err
	}
	_, err = peers.WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, tx, 30*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// AddAccessNode introduces node at accessNodeIndex as an access node to the chain.
// This is done by activating the chain on the node and asking the governance contract
// to consider it as an access node.
func (clu *Cluster) AddAccessNode(accessNodeIndex int, chain *Chain) (*iotago.Transaction, error) {
	waspClient := clu.WaspClient(accessNodeIndex)
	if err := apilib.ActivateChainOnNodes(clu.WaspClientFromHostName, clu.Config.APIHosts([]int{accessNodeIndex}), chain.ChainID); err != nil {
		return nil, err
	}

	accessNodePeering, _, err := waspClient.NodeApi.GetPeeringIdentity(context.Background()).Execute()
	if err != nil {
		return nil, err
	}

	accessNodePubKey, err := cryptolib.NewPublicKeyFromString(accessNodePeering.PublicKey)
	if err != nil {
		return nil, err
	}

	cert, _, err := waspClient.NodeApi.SetNodeOwner(context.Background()).NodeOwnerCertificateRequest(apiclient.NodeOwnerCertificateRequest{
		PublicKey:    accessNodePubKey.String(),
		OwnerAddress: chain.OriginatorAddress().String(),
	}).Execute()

	if err != nil {
		return nil, err
	}

	decodedCert, err := iotago.DecodeHex(cert.Certificate)
	if err != nil {
		return nil, err
	}

	scArgs := governance.AccessNodeInfo{
		NodePubKey:    accessNodePubKey.AsBytes(),
		ValidatorAddr: isc.BytesFromAddress(chain.OriginatorAddress()),
		Certificate:   decodedCert,
		ForCommittee:  false,
		AccessAPI:     clu.Config.APIHost(accessNodeIndex),
	}
	scParams := chainclient.
		NewPostRequestParams(scArgs.ToAddCandidateNodeParams()).
		WithBaseTokens(1000)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	tx, err := govClient.PostRequest(governance.FuncAddCandidateNode.Name, *scParams)
	if err != nil {
		return nil, err
	}

	txID, err := tx.ID()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[cluster] Governance::AddCandidateNode, Posted TX, id=%v, args=%+v\n", txID, scArgs)
	return tx, nil
}

func (clu *Cluster) IsNodeUp(i int) bool {
	return clu.waspCmds[i] != nil && clu.waspCmds[i].cmd != nil
}

func (clu *Cluster) MultiClient() *multiclient.MultiClient {
	return multiclient.New(clu.WaspClientFromHostName, clu.Config.APIHosts()) //.WithLogFunc(clu.t.Logf)
}

func (clu *Cluster) WaspClientFromHostName(hostName string) *apiclient.APIClient {
	client, err := apiextensions.WaspAPIClientByHostName(hostName)

	if err != nil {
		panic(err.Error())
	}

	return client
}

func (clu *Cluster) WaspClient(nodeIndex ...int) *apiclient.APIClient {
	idx := 0
	if len(nodeIndex) == 1 {
		idx = nodeIndex[0]
	}

	return clu.WaspClientFromHostName(clu.Config.APIHost(idx))
}

func (clu *Cluster) NodeDataPath(i int) string {
	return path.Join(clu.DataPath, fmt.Sprintf("wasp%d", i))
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

// InitDataPath initializes the cluster data directory (cluster.json + one subdirectory
// for each node).
func (clu *Cluster) InitDataPath(templatesPath string, removeExisting bool) error {
	exists, err := fileExists(clu.DataPath)
	if err != nil {
		return err
	}
	if exists {
		if !removeExisting {
			return fmt.Errorf("%s directory exists", clu.DataPath)
		}
		err = os.RemoveAll(clu.DataPath)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(clu.Config.Wasp); i++ {
		err = initNodeConfig(
			clu.NodeDataPath(i),
			path.Join(templatesPath, "wasp-config-template.json"),
			templates.WaspConfig,
			&clu.Config.Wasp[i],
		)
		if err != nil {
			return err
		}
	}
	return clu.Config.Save(clu.DataPath)
}

func initNodeConfig(nodePath, configTemplatePath, defaultTemplate string, params *templates.WaspConfigParams) error {
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

	return configTmpl.Execute(f, params)
}

// Start launches all wasp nodes in the cluster, each running in its own directory
func (clu *Cluster) Start(dataPath string) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("data path %s does not exist", dataPath)
	}

	if err := clu.start(); err != nil {
		return err
	}

	if err := clu.TrustAll(); err != nil {
		return err
	}

	clu.Started = true
	return nil
}

func (clu *Cluster) start() error {
	start := time.Now()
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", len(clu.Config.Wasp))

	initOk := make(chan bool, len(clu.Config.Wasp))

	for i := 0; i < len(clu.Config.Wasp); i++ {
		err := clu.startWaspNode(i, initOk)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(clu.Config.Wasp); i++ {
		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return errors.New("timeout starting wasp nodes")
		}
	}
	fmt.Printf("[cluster] started %d Wasp nodes in %v\n", len(clu.Config.Wasp), time.Since(start))
	return nil
}

func (clu *Cluster) KillNodeProcess(nodeIndex int) error {
	if nodeIndex >= len(clu.waspCmds) {
		return fmt.Errorf("[cluster] Wasp node with index %d not found", nodeIndex)
	}

	wcmd := clu.waspCmds[nodeIndex]
	if wcmd == nil {
		return nil
	}

	if err := wcmd.cmd.Process.Kill(); err != nil {
		return err
	}

	clu.waspCmds[nodeIndex] = nil
	return nil
}

func (clu *Cluster) RestartNodes(nodeIndex ...int) error {
	for _, ni := range nodeIndex {
		if !lo.Contains(clu.AllNodes(), ni) {
			panic(fmt.Errorf("unexpected node index specified for a restart: %v", ni))
		}
	}

	// send stop commands
	for _, i := range nodeIndex {
		clu.stopNode(i)
	}

	// wait until all nodes are stopped
	exitedWaitGroup := sync.WaitGroup{}
	exitedWaitGroup.Add(len(nodeIndex))

	exited := make(chan struct{})
	go func() {
		exitedWaitGroup.Wait()
		close(exited)
	}()

	for _, i := range nodeIndex {
		go func(cmd *waspCmd) {
			waitCmd(cmd)
			// mark process as finished
			exitedWaitGroup.Done()
		}(clu.waspCmds[i])
	}

	select {
	case <-time.After(20 * time.Second):
		return errors.New("[cluster] Wasp nodes did not shutdown in time")
	case <-exited:
		// nodes stopped
	}

	// start nodes
	initOk := make(chan bool, len(nodeIndex))
	okCount := 0
	for _, i := range nodeIndex {
		err := clu.startWaspNode(i, initOk)
		if err != nil {
			return err
		}
		select {
		case <-initOk:
			okCount++
			if okCount == len(nodeIndex) {
				return nil
			}
		case <-time.After(5 * time.Second):
			return errors.New("timeout starting wasp nodes")
		}
	}
	return nil
}

func (clu *Cluster) startWaspNode(nodeIndex int, initOk chan<- bool) error {
	wcmd := &waspCmd{}

	cmd := exec.Command("wasp", "-c", "config.json")
	cmd.Dir = clu.NodeDataPath(nodeIndex)

	// force the wasp processes to close if the cluster tests time out
	if clu.t != nil {
		util.TerminateCmdWhenTestStops(cmd)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	name := fmt.Sprintf("wasp %d", nodeIndex)
	go scanLog(stderrPipe, &wcmd.logScanner, fmt.Sprintf("!%s", name))
	go scanLog(stdoutPipe, &wcmd.logScanner, fmt.Sprintf(" %s", name))

	nodeAPIURL := fmt.Sprintf("http://localhost:%s", strconv.Itoa(clu.Config.APIPort(nodeIndex)))
	go waitForAPIReady(initOk, nodeAPIURL)

	wcmd.cmd = cmd
	clu.waspCmds[nodeIndex] = wcmd
	return nil
}

const pollAPIInterval = 500 * time.Millisecond

// waits until API for a given WASP node is ready
func waitForAPIReady(initOk chan<- bool, apiURL string) {
	infoEndpointURL := fmt.Sprintf("%s%s", apiURL, "/node/version")

	ticker := time.NewTicker(pollAPIInterval)
	go func() {
		for {
			<-ticker.C

			rsp, err := http.Get(infoEndpointURL) //nolint:gosec,noctx
			if err != nil {
				fmt.Printf("Error Polling node API %s ready status: %v\n", apiURL, err)
				continue
			}
			fmt.Printf("Polling node API %s ready status: %s %s\n", apiURL, infoEndpointURL, rsp.Status)
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

func scanLog(reader io.Reader, wg *sync.WaitGroup, tag string) {
	wg.Add(1)
	defer wg.Done()

	// unlike bufio.Scanner, bufio.Reader supports reading lines of unlimited size
	br := bufio.NewReader(reader)
	isLineStart := true
	for {
		line, isPrefix, err := br.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[cluster] error reading output for %s: %v\n", tag, err)
			}
			break
		}

		if isLineStart {
			fmt.Printf("[%s] ", tag)
		}
		fmt.Printf("%s", line)
		if !isPrefix {
			fmt.Println()
		}

		isLineStart = !isPrefix // for next iteration
	}
}

func (clu *Cluster) stopNode(nodeIndex int) {
	if !clu.IsNodeUp(nodeIndex) {
		return
	}
	fmt.Printf("[cluster] Sending shutdown to wasp node %d\n", nodeIndex)

	err := clu.KillNodeProcess(nodeIndex)
	if err != nil {
		fmt.Println(err)
	}
}

func (clu *Cluster) StopNode(nodeIndex int) {
	clu.stopNode(nodeIndex)
	waitCmd(clu.waspCmds[nodeIndex])
	clu.waspCmds[nodeIndex] = nil
	fmt.Printf("[cluster] Node %d has been shut down\n", nodeIndex)
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (clu *Cluster) Stop() {
	for i := 0; i < len(clu.Config.Wasp); i++ {
		clu.stopNode(i)
	}
	clu.Wait()
}

func (clu *Cluster) Wait() {
	for i := 0; i < len(clu.Config.Wasp); i++ {
		waitCmd(clu.waspCmds[i])
		clu.waspCmds[i] = nil
	}
}

func waitCmd(wcmd *waspCmd) {
	if wcmd == nil {
		return
	}
	wcmd.logScanner.Wait()
	if err := wcmd.cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}

func (clu *Cluster) AllNodes() []int {
	nodes := make([]int, 0)
	for i := 0; i < len(clu.Config.Wasp); i++ {
		nodes = append(nodes, i)
	}
	return nodes
}

func (clu *Cluster) ActiveNodes() []int {
	nodes := make([]int, 0)
	for _, i := range clu.AllNodes() {
		if !clu.IsNodeUp(i) {
			continue
		}
		nodes = append(nodes, i)
	}
	return nodes
}

func (clu *Cluster) PostTransaction(tx *iotago.Transaction) error {
	_, err := clu.l1.PostTxAndWaitUntilConfirmation(tx)
	return err
}

func (clu *Cluster) AddressBalances(addr iotago.Address) *isc.Assets {
	// get funds controlled by addr
	outputMap, err := clu.l1.OutputMap(addr)
	if err != nil {
		fmt.Printf("[cluster] GetConfirmedOutputs error: %v\n", err)
		return nil
	}
	balance := isc.NewEmptyAssets()
	for _, out := range outputMap {
		balance.Add(transaction.AssetsFromOutput(out))
	}

	// if the address is an alias output, we also need to fetch the output itself and add that balance
	if aliasAddr, ok := addr.(*iotago.AliasAddress); ok {
		_, aliasOutput, err := clu.l1.GetAliasOutput(aliasAddr.AliasID())
		if err != nil {
			fmt.Printf("[cluster] GetAliasOutput error: %v\n", err)
			return nil
		}
		balance.Add(transaction.AssetsFromOutput(aliasOutput))
	}
	return balance
}

func (clu *Cluster) L1BaseTokens(addr iotago.Address) uint64 {
	tokens := clu.AddressBalances(addr)
	return tokens.BaseTokens
}

func (clu *Cluster) AssertAddressBalances(addr iotago.Address, expected *isc.Assets) bool {
	return clu.AddressBalances(addr).Equals(expected)
}

func (clu *Cluster) GetOutputs(addr iotago.Address) (map[iotago.OutputID]iotago.Output, error) {
	return clu.l1.OutputMap(addr)
}
