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
	"runtime"
	"strconv"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/multiclient"
	"github.com/iotaledger/wasp/v2/packages/apilib"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/evm/evmlogger"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

type Cluster struct {
	Name              string
	Config            *ClusterConfig
	Started           bool
	DataPath          string
	OriginatorKeyPair *cryptolib.KeyPair
	l1                clients.L1Client
	l1ParamsFetcher   parameters.L1ParamsFetcher
	waspCmds          []*waspCmd
	t                 *testing.T
	log               log.Logger
}

type waspCmd struct {
	cmd        *exec.Cmd
	logScanner sync.WaitGroup
}

func New(name string, config *ClusterConfig, dataPath string, t *testing.T, log log.Logger, l1PacakgeID *iotago.PackageID) *Cluster {
	if log == nil {
		if t == nil {
			panic("one of t or log must be set")
		}
		log = testlogger.NewLogger(t)
	}
	evmlogger.Init(log)

	config.setValidatorAddressIfNotSet() // privtangle prefix
	for i := range config.Wasp {
		config.Wasp[i].PackageID = l1PacakgeID
	}
	client := config.L1Client()
	return &Cluster{
		Name:              name,
		Config:            config,
		OriginatorKeyPair: cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes(testcommon.TestSeed)), // TODO temporary use a fixed account with a lot of tokens
		waspCmds:          make([]*waspCmd, len(config.Wasp)),
		t:                 t,
		log:               log,
		l1:                client,
		l1ParamsFetcher:   parameters.NewL1ParamsFetcher(client.IotaClient(), log),
		DataPath:          dataPath,
	}
}

func (clu *Cluster) Logf(format string, args ...any) {
	if clu.t != nil {
		clu.t.Logf(format, args...)
		return
	}
	clu.log.LogInfof(format, args...)
}

func (clu *Cluster) NewKeyPairWithFunds() (*cryptolib.KeyPair, *cryptolib.Address, error) {
	key, addr := testkey.GenKeyAddr()
	err := clu.RequestFunds(addr)
	return key, addr, err
}

func (clu *Cluster) RequestFunds(addr *cryptolib.Address) error {
	return clu.l1.RequestFunds(context.Background(), *addr)
}

func (clu *Cluster) L1Client() clients.L1Client {
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
			//nolint:bodyclose // false positive
			nodes[ni]).NodeAPI.TrustPeer(context.Background()).PeeringTrustRequest(peerInfo).Execute(); err != nil {
			return err
		}
	}
	return nil
}

func (clu *Cluster) Login() ([]string, error) {
	allNodes := clu.Config.AllNodes()
	jwtTokens := make([]string, len(allNodes))
	for ni := range allNodes {
		res, _, err := clu.WaspClient(allNodes[ni]).AuthAPI.Authenticate(context.Background()).
			LoginRequest(*apiclient.NewLoginRequest("wasp", "wasp")).
			Execute() //nolint:bodyclose // false positive
		if err != nil {
			return nil, err
		}
		jwtTokens[ni] = "Bearer " + res.Jwt
	}
	return jwtTokens, nil
}

func (clu *Cluster) TrustAll(jwtTokens ...string) error {
	allNodes := clu.Config.AllNodes()
	allPeers := make([]*apiclient.PeeringNodeIdentityResponse, len(allNodes))
	clients := make([]*apiclient.APIClient, len(allNodes))
	for ni := range allNodes {
		clients[ni] = clu.WaspClient(allNodes[ni])
		if jwtTokens != nil {
			clients[ni].GetConfig().AddDefaultHeader("Authorization", jwtTokens[ni])
		}
	}
	for ni := range allNodes {
		var err error
		//nolint:bodyclose // false positive
		if allPeers[ni], _, err = clients[ni].NodeAPI.GetPeeringIdentity(context.Background()).Execute(); err != nil {
			return err
		}
	}
	for ni := range allNodes {
		for pi := range allPeers {
			var err error
			if ni == pi {
				continue // dont trust self
			}
			if _, err = clients[ni].NodeAPI.TrustPeer(context.Background()).PeeringTrustRequest(
				apiclient.PeeringTrustRequest{
					Name:       fmt.Sprintf("%d", pi),
					PublicKey:  allPeers[pi].PublicKey,
					PeeringURL: allPeers[pi].PeeringURL,
				},
			).Execute(); err != nil { //nolint:bodyclose // false positive
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
	return clu.DeployChainWithDKG(committee, committee, uint16(quorum))
}

func (clu *Cluster) InitDKG(committeeNodeCount int) ([]int, *cryptolib.Address, error) {
	cmt := util.MakeRange(0, committeeNodeCount-1) // End is inclusive for some reason.
	quorum := uint16((2*len(cmt))/3 + 1)

	address, err := clu.RunDKG(cmt, quorum)

	return cmt, address, err
}

func (clu *Cluster) RunDKG(committeeNodes []int, threshold uint16, timeout ...time.Duration) (*cryptolib.Address, error) {
	if threshold == 0 {
		threshold = (uint16(len(committeeNodes))*2)/3 + 1
	}
	apiHosts := clu.Config.APIHosts(committeeNodes)

	peerPubKeys := make([]string, 0)
	for _, i := range committeeNodes {
		//nolint:bodyclose // false positive
		peeringNodeInfo, _, err := clu.WaspClient(i).NodeAPI.GetPeeringIdentity(context.Background()).Execute()
		if err != nil {
			return nil, err
		}

		peerPubKeys = append(peerPubKeys, peeringNodeInfo.PublicKey)
	}

	dkgInitiatorIndex := rand.Intn(len(apiHosts))
	client := clu.WaspClientFromHostName(apiHosts[dkgInitiatorIndex])

	return apilib.RunDKG(context.Background(), client, peerPubKeys, threshold, timeout...)
}

func (clu *Cluster) DeployChainWithDKG(allPeers, committeeNodes []int, quorum uint16, blockKeepAmount ...int32) (*Chain, error) {
	stateAddr, err := clu.RunDKG(committeeNodes, quorum)
	if err != nil {
		return nil, err
	}
	return clu.DeployChain(allPeers, committeeNodes, quorum, stateAddr, blockKeepAmount...)
}

func (clu *Cluster) DeployChain(allPeers, committeeNodes []int, quorum uint16, stateAddr *cryptolib.Address, blockKeepAmount ...int32) (*Chain, error) {
	if len(allPeers) == 0 {
		allPeers = clu.Config.AllNodes()
	}

	chain := &Chain{
		OriginatorKeyPair: clu.OriginatorKeyPair,
		AllPeers:          allPeers,
		CommitteeNodes:    committeeNodes,
		Quorum:            quorum,
		Cluster:           clu,
	}

	l1Client := clu.L1Client()
	address := chain.OriginatorAddress()

	err := clu.RequestFunds(address)
	if err != nil {
		return nil, fmt.Errorf("DeployChain: %w", err)
	}

	committeePubKeys := make([]string, len(chain.CommitteeNodes))
	for i, nodeIndex := range chain.CommitteeNodes {
		//nolint:bodyclose // false positive
		peeringNode, _, err2 := clu.WaspClient(nodeIndex).NodeAPI.GetPeeringIdentity(context.Background()).Execute()
		if err2 != nil {
			return nil, err2
		}

		committeePubKeys[i] = peeringNode.PublicKey
	}

	var blockKeepAmountVal int32
	if len(blockKeepAmount) > 0 {
		blockKeepAmountVal = blockKeepAmount[0]
	}
	encodedInitParams := origin.NewInitParams(
		isc.NewAddressAgentID(chain.OriginatorAddress()),
		1074,
		blockKeepAmountVal,
		false,
	).Encode()

	getCoinsRes, err := l1Client.GetCoins(
		context.Background(),
		iotaclient.GetCoinsRequest{Owner: address.AsIotaAddress()},
	)
	if err != nil {
		return nil, fmt.Errorf("cant get gas coin: %w", err)
	}

	var gascoin *iotajsonrpc.Coin
	for _, coin := range getCoinsRes.Data {
		// dont pick a too big coin object
		if coin.Balance.Uint64() < 3*iotaclient.FundsFromFaucetAmount &&
			iotaclient.FundsFromFaucetAmount <= coin.Balance.Uint64() {
			gascoin = coin
		}
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.TransferObject(stateAddr.AsIotaAddress(), gascoin.Ref())
	if err != nil {
		return nil, fmt.Errorf("cant transfer gas coin: %w", err)
	}
	pt := ptb.Finish()

	resTransferGasCoin, err := l1Client.SignAndExecuteTxWithRetry(
		context.Background(),
		cryptolib.SignerToIotaSigner(chain.OriginatorKeyPair),
		pt,
		nil,
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowInput:   true,
			ShowEffects: true,
		},
	)
	if err != nil || !resTransferGasCoin.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("can't transfer GasCoin, resTransferGasCoin.Effects.Data.IsSuccess(): %v: %w", resTransferGasCoin.Effects.Data.IsSuccess(), err)
	}
	fmt.Printf("chosen GasCoin %s", gascoin.String())

	l1Params, err := clu.l1ParamsFetcher.GetOrFetchLatest(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cant get L1Params: %w", err)
	}

	stateMetaData := *transaction.NewStateMetadata(
		allmigrations.DefaultScheme.LatestSchemaVersion(),
		origin.L1Commitment(
			allmigrations.DefaultScheme.LatestSchemaVersion(),
			encodedInitParams,
			*gascoin.CoinObjectID,
			0,
			l1Params,
		),
		gascoin.CoinObjectID,
		gas.DefaultFeePolicy(),
		encodedInitParams,
		0,
		"",
	)

	chainID, err := apilib.DeployChain(
		context.Background(),
		apilib.CreateChainParams{
			Layer1Client:      l1Client,
			CommitteeAPIHosts: chain.CommitteeAPIHosts(),
			Signer:            chain.OriginatorKeyPair,
			Textout:           os.Stdout,
			Prefix:            "[cluster] ",
			StateMetadata:     stateMetaData,
			PackageID:         clu.Config.ISCPackageID(),
		},
		stateAddr,
	)
	if err != nil {
		return nil, fmt.Errorf("DeployChain: %w", err)
	}

	// activate chain on nodes
	err = apilib.ActivateChainOnNodes(clu.WaspClientFromHostName, chain.CommitteeAPIHosts(), chainID)
	if err != nil {
		clu.t.Fatalf("activating chain %s.. FAILED: %v\n", chainID.String(), err)
	}
	fmt.Printf("activating chain %s.. OK.\n", chainID.String())

	// ---------- wait until the request is processed at least in all committee nodes
	{
		fmt.Printf("waiting until nodes receive the origin output..\n")

		retries := 30
		for {
			time.Sleep(200 * time.Millisecond)
			err = multiclient.New(clu.WaspClientFromHostName, chain.CommitteeAPIHosts()).Do(
				func(_ int, a *apiclient.APIClient) error {
					_, _, err2 := a.ChainsAPI.GetChainInfo(context.Background()).Execute() //nolint:bodyclose // false positive
					return err2
				})
			if err != nil {
				if retries > 0 {
					retries--
					continue
				}
				return nil, err
			}
			break
		}

		fmt.Printf("waiting until nodes receive the origin output.. DONE\n")
	}

	chain.StateAddress = stateAddr
	chain.ChainID = chainID

	// After a rotation other nodes can become access nodes,
	// so we make all of the nodes are possible access nodes.
	return chain, clu.addAllAccessNodes(chain, allPeers)
}

func (clu *Cluster) addAllAccessNodes(chain *Chain, accessNodes []int) error {
	//
	// Register all nodes as access nodes.
	addAccessNodesTxs := make([]*iotajsonrpc.IotaTransactionBlockResponse, len(accessNodes))
	for i, a := range accessNodes {
		tx, err := clu.addAccessNode(a, chain)
		if err != nil {
			return err
		}
		addAccessNodesTxs[i] = tx
		time.Sleep(100 * time.Millisecond) // give some time for the indexer to catch up, otherwise it might not find the user outputs...
	}

	peers := multiclient.New(clu.WaspClientFromHostName, chain.CommitteeAPIHosts())

	for _, tx := range addAccessNodesTxs {
		// ---------- wait until the requests are processed in all committee nodes

		if _, err := peers.WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chain.ChainID, tx, true, 30*time.Second); err != nil {
			return fmt.Errorf("WaitAddAccessNode: %w", err)
		}
	}

	pubKeys := governance.ChangeAccessNodeActions{}
	for _, a := range accessNodes {
		waspClient := clu.WaspClient(a)

		//nolint:bodyclose // false positive
		accessNodePeering, _, err := waspClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
		if err != nil {
			return err
		}

		accessNodePubKey, err := cryptolib.PublicKeyFromString(accessNodePeering.PublicKey)
		if err != nil {
			return err
		}

		pubKeys = append(pubKeys, governance.AcceptAccessNodeAction(accessNodePubKey))
	}
	scParams := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(iotaclient.DefaultGasBudget + 10),
		GasBudget: 2 * iotaclient.DefaultGasBudget,
	}

	govClient := chain.Client(clu.OriginatorKeyPair)

	tx, err := govClient.PostRequest(context.Background(), governance.FuncChangeAccessNodes.Message(pubKeys), scParams)
	if err != nil {
		return err
	}
	_, err = peers.WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chain.ChainID, tx, true, 30*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// addAccessNode introduces node at accessNodeIndex as an access node to the chain.
// This is done by activating the chain on the node and asking the governance contract
// to consider it as an access node.
func (clu *Cluster) addAccessNode(accessNodeIndex int, chain *Chain) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	waspClient := clu.WaspClient(accessNodeIndex)
	if err := apilib.ActivateChainOnNodes(clu.WaspClientFromHostName, clu.Config.APIHosts([]int{accessNodeIndex}), chain.ChainID); err != nil {
		return nil, err
	}

	validatorKeyPair := clu.Config.ValidatorKeyPair(accessNodeIndex)
	err := clu.RequestFunds(validatorKeyPair.Address())
	if err != nil {
		return nil, err
	}

	//nolint:bodyclose // false positive
	accessNodePeering, _, err := waspClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	if err != nil {
		return nil, err
	}

	accessNodePubKey, err := cryptolib.PublicKeyFromString(accessNodePeering.PublicKey)
	if err != nil {
		return nil, err
	}

	cert, _, err := waspClient.NodeAPI.OwnerCertificate(context.Background()).Execute() //nolint:bodyclose // false positive
	if err != nil {
		return nil, err
	}

	decodedCert, err := cryptolib.DecodeHex(cert.Certificate)
	if err != nil {
		return nil, err
	}

	accessAPI := clu.Config.APIHost(accessNodeIndex)
	forCommittee := false

	govClient := chain.Client(validatorKeyPair)
	params := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(1000),
		GasBudget: iotaclient.DefaultGasBudget,
	}
	tx, err := govClient.PostRequest(
		context.Background(),
		governance.FuncAddCandidateNode.Message(accessNodePubKey, decodedCert, accessAPI, forCommittee),
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call FuncAddCandidateNode: %w", err)
	}

	fmt.Printf("[cluster] Governance::AddCandidateNode, Posted TX, digest=%v, NodePubKey=%v, Certificate=%x, accessAPI=%v, forCommittee=%v\n",
		tx.Digest, accessNodePubKey, decodedCert, accessAPI, forCommittee)
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

	for i := range clu.Config.Wasp {
		err = initNodeConfig(
			clu.NodeDataPath(i),
			path.Join(templatesPath, "wasp-config-template.json"),
			waspConfigTemplate,
			&clu.Config.Wasp[i],
		)
		if err != nil {
			return err
		}
	}
	return clu.Config.Save(clu.DataPath)
}

func initNodeConfig(nodePath, configTemplatePath, defaultTemplate string, params *WaspConfigParams) error {
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

// StartAndTrustAll launches all wasp nodes in the cluster, each running in its own directory
func (clu *Cluster) StartAndTrustAll(dataPath string) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("data path %s does not exist", dataPath)
	}

	if err = clu.Start(); err != nil {
		return err
	}

	var jwtTokens []string
	if clu.Config.Wasp[0].AuthScheme == "jwt" {
		if jwtTokens, err = clu.Login(); err != nil {
			return err
		}
	}

	if err := clu.TrustAll(jwtTokens...); err != nil {
		return err
	}

	clu.Started = true
	return nil
}

func (clu *Cluster) Start() error {
	start := time.Now()
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", len(clu.Config.Wasp))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initOk := make(chan bool, len(clu.Config.Wasp))
	for i := 0; i < len(clu.Config.Wasp); i++ {
		err := clu.startWaspNode(ctx, i, initOk)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(clu.Config.Wasp); i++ {
		select {
		case <-initOk:
		case <-time.After(30 * time.Second):
			return errors.New("timeout starting wasp nodes")
		}
	}

	fmt.Printf("[cluster] started %d Wasp nodes in %v\n", len(clu.Config.Wasp), time.Since(start))
	return nil
}

func (clu *Cluster) KillNodeProcess(nodeIndex int, gracefully bool) error {
	if nodeIndex >= len(clu.waspCmds) {
		return fmt.Errorf("[cluster] Wasp node with index %d not found", nodeIndex)
	}

	wcmd := clu.waspCmds[nodeIndex]
	if wcmd == nil {
		return nil
	}

	if gracefully && runtime.GOOS != util.WindowsOS {
		if err := wcmd.cmd.Process.Signal(os.Interrupt); err != nil {
			return err
		}
		if _, err := wcmd.cmd.Process.Wait(); err != nil {
			return err
		}
	} else {
		if err := wcmd.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	clu.waspCmds[nodeIndex] = nil
	return nil
}

func (clu *Cluster) RestartNodes(keepDB bool, nodeIndexes ...int) error {
	for _, ni := range nodeIndexes {
		if !lo.Contains(clu.AllNodes(), ni) {
			panic(fmt.Errorf("unexpected node index specified for a restart: %v", ni))
		}
	}

	// send stop commands
	for _, i := range nodeIndexes {
		clu.stopNode(i)
		if !keepDB {
			dbPath := clu.NodeDataPath(i) + "/waspdb/chains/data/"
			clu.log.LogInfof("Deleting DB from %v", dbPath)
			if err := os.RemoveAll(dbPath); err != nil {
				return fmt.Errorf("cannot remove the node=%v DB at %v: %w", i, dbPath, err)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// restart nodes
	initOk := make(chan bool, len(nodeIndexes))
	for _, i := range nodeIndexes {
		err := clu.startWaspNode(ctx, i, initOk)
		if err != nil {
			return err
		}
	}

	for range nodeIndexes {
		select {
		case <-initOk:
		case <-time.After(60 * time.Second):
			return errors.New("timeout restarting wasp nodes")
		}
	}

	return nil
}

func (clu *Cluster) startWaspNode(ctx context.Context, nodeIndex int, initOk chan<- bool) error {
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
	go waitForAPIReady(ctx, initOk, nodeAPIURL)

	wcmd.cmd = cmd
	clu.waspCmds[nodeIndex] = wcmd
	return nil
}

const pollAPIInterval = 500 * time.Millisecond

// waits until API for a given WASP node is ready
func waitForAPIReady(ctx context.Context, initOk chan<- bool, apiURL string) {
	waspHealthEndpointURL := fmt.Sprintf("%s%s", apiURL, "/health")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				rsp, err := http.Get(waspHealthEndpointURL) //nolint:gosec,noctx
				if err != nil {
					time.Sleep(pollAPIInterval)
					continue
				}
				_ = rsp.Body.Close()

				if rsp.StatusCode != http.StatusOK {
					time.Sleep(pollAPIInterval)
					continue
				}

				initOk <- true
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

	err := clu.KillNodeProcess(nodeIndex, true)
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

func (clu *Cluster) AddressBalances(addr *cryptolib.Address) *isc.Assets {
	// get funds controlled by addr

	balances, err := clu.l1.GetAllBalances(context.Background(), addr.AsIotaAddress())
	if err != nil {
		clu.log.LogPanicf("[cluster] failed to GetAllBalances for address[%v]", addr.String())
		return nil
	}

	balance := isc.NewEmptyAssets()
	for _, out := range balances {
		if coin.BaseTokenType.MatchesStringType(out.CoinType.String()) {
			balance.SetBaseTokens(coin.Value(out.TotalBalance.Uint64()))
		} else {
			balance.AddCoin(coin.MustTypeFromString(out.CoinType.String()), coin.Value(out.TotalBalance.Uint64()))
		}
	}

	return balance
}

func (clu *Cluster) L1BaseTokens(addr *cryptolib.Address) coin.Value {
	tokens := clu.AddressBalances(addr)
	return tokens.BaseTokens()
}

func (clu *Cluster) AssertAddressBalances(addr *cryptolib.Address, expected *isc.Assets) bool {
	return clu.AddressBalances(addr).Equals(expected)
}
