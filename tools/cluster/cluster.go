package cluster

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/cluster/mocknode"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type Cluster struct {
	Name    string
	Config  *ClusterConfig
	Started bool

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

func (clu *Cluster) GoshimmerClient() *goshimmer.Client {
	return goshimmer.NewClient(clu.Config.goshimmerApiHost(), clu.Config.FaucetPoWTarget)
}

func (clu *Cluster) DeployDefaultChain() (*Chain, error) {
	committee := clu.Config.AllNodes()
	minQuorum := len(committee)/2 + 1
	quorum := len(committee) * 3 / 4
	if quorum < minQuorum {
		quorum = minQuorum
	}
	return clu.DeployChainWithDKG("Default chain", committee, uint16(quorum))
}

func (clu *Cluster) RunDKG(committeeNodes []int, threshold uint16, timeout ...time.Duration) (ledgerstate.Address, error) {
	apiHosts := clu.Config.ApiHosts(committeeNodes)
	peeringHosts := clu.Config.PeeringHosts(committeeNodes)
	dkgInitiatorIndex := uint16(rand.Intn(len(apiHosts)))
	return apilib.RunDKG(apiHosts, peeringHosts, threshold, dkgInitiatorIndex, timeout...)
}

func (clu *Cluster) DeployChainWithDKG(description string, committeeNodes []int, quorum uint16) (*Chain, error) {
	stateAddr, err := clu.RunDKG(committeeNodes, quorum)
	if err != nil {
		return nil, err
	}

	return clu.DeployChain(description, committeeNodes, quorum, stateAddr)
}

func (clu *Cluster) DeployChain(description string, committeeNodes []int, quorum uint16, stateAddr ledgerstate.Address) (*Chain, error) {
	ownerSeed := seed.NewSeed()

	chain := &Chain{
		Description:    description,
		OriginatorSeed: ownerSeed,
		CommitteeNodes: committeeNodes,
		Quorum:         quorum,
		Cluster:        clu,
	}
	err := clu.GoshimmerClient().RequestFunds(chain.OriginatorAddress())
	if err != nil {
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	chainid, err := apilib.DeployChain(apilib.CreateChainParams{
		Node:                  clu.GoshimmerClient(),
		CommitteeApiHosts:     chain.ApiHosts(),
		CommitteePeeringHosts: chain.PeeringHosts(),
		N:                     uint16(len(committeeNodes)),
		T:                     quorum,
		OriginatorKeyPair:     chain.OriginatorKeyPair(),
		Description:           description,
		Textout:               os.Stdout,
		Prefix:                "[cluster] ",
	}, stateAddr)
	if err != nil {
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	chain.StateAddress = stateAddr
	chain.ChainID = *chainid

	return chain, nil
}

func (cluster *Cluster) IsGoshimmerUp() bool {
	return cluster.goshimmer != nil
}

func (cluster *Cluster) IsNodeUp(i int) bool {
	return cluster.waspCmds[i] != nil
}

func (cluster *Cluster) MultiClient() *multiclient.MultiClient {
	return multiclient.New(cluster.Config.ApiHosts())
}

func (cluster *Cluster) WaspClient(nodeIndex int) *client.WaspClient {
	return client.NewWaspClient(cluster.Config.ApiHost(nodeIndex))
}

func waspNodeDataPath(dataPath string, i int) string {
	return path.Join(dataPath, fmt.Sprintf("wasp%d", i))
}

func goshimmerDataPath(dataPath string) string {
	return path.Join(dataPath, "goshimmer")
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
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
func (cluster *Cluster) InitDataPath(templatesPath string, dataPath string, removeExisting bool) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if exists {
		if !removeExisting {
			return fmt.Errorf("%s directory exists", dataPath)
		}
		err = os.RemoveAll(dataPath)
		if err != nil {
			return err
		}
	}

	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		err = initNodeConfig(
			waspNodeDataPath(dataPath, i),
			path.Join(templatesPath, "wasp-config-template.json"),
			templates.WaspConfig,
			cluster.Config.WaspConfigTemplateParams(i),
		)
		if err != nil {
			return err
		}
	}
	return cluster.Config.Save(dataPath)
}

func initNodeConfig(nodePath string, configTemplatePath string, defaultTemplate string, params interface{}) error {
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
	defer f.Close()

	return configTmpl.Execute(f, params)
}

// Start launches all wasp nodes in the cluster, each running in its own directory
func (cluster *Cluster) Start(dataPath string) error {
	exists, err := fileExists(dataPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Data path %s does not exist", dataPath)
	}

	err = cluster.start(dataPath)
	if err != nil {
		return err
	}

	cluster.Started = true
	return nil
}

func (cluster *Cluster) start(dataPath string) error {
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", cluster.Config.Wasp.NumNodes)

	if !cluster.Config.Goshimmer.Provided {
		cluster.goshimmer = mocknode.Start(
			fmt.Sprintf(":%d", cluster.Config.Goshimmer.TxStreamPort),
			fmt.Sprintf(":%d", cluster.Config.Goshimmer.ApiPort),
		)
		fmt.Printf("[cluster] started goshimmer node\n")
	}

	initOk := make(chan bool, cluster.Config.Wasp.NumNodes)

	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		cmd, err := cluster.startServer("wasp", waspNodeDataPath(dataPath, i), fmt.Sprintf("wasp %d", i), initOk, "nanomsg publisher is running")
		if err != nil {
			return err
		}
		cluster.waspCmds[i] = cmd
	}

	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return fmt.Errorf("Timeout starting wasp nodes\n")
		}
	}
	fmt.Printf("[cluster] started %d Wasp nodes\n", cluster.Config.Wasp.NumNodes)
	return nil
}

func (cluster *Cluster) startServer(command string, cwd string, name string, initOk chan<- bool, initOkMsg string) (*exec.Cmd, error) {
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
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go scanLog(
		stderrPipe,
		func(line string) { fmt.Printf("[!%s] %s\n", name, line) },
	)
	go scanLog(
		stdoutPipe,
		func(line string) { fmt.Printf("[ %s] %s\n", name, line) },
		waitFor(initOkMsg, initOk),
	)

	return cmd, nil
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

func waitFor(msg string, initOk chan<- bool) func(line string) {
	found := false
	return func(line string) {
		if found {
			return
		}
		if strings.Contains(line, msg) {
			initOk <- true
			found = true
		}
	}
}

func (cluster *Cluster) stopGoshimmer() {
	if !cluster.IsGoshimmerUp() {
		return
	}
	fmt.Printf("[cluster] Stopping Goshimmer MockNode\n")
	cluster.goshimmer.Stop()
}

func (cluster *Cluster) stopNode(nodeIndex int) {
	if !cluster.IsNodeUp(nodeIndex) {
		return
	}
	fmt.Printf("[cluster] Sending shutdown to wasp node %d\n", nodeIndex)
	err := cluster.WaspClient(nodeIndex).Shutdown()
	if err != nil {
		fmt.Println(err)
	}
}

func (cluster *Cluster) StopNode(nodeIndex int) {
	cluster.stopNode(nodeIndex)
	waitCmd(&cluster.waspCmds[nodeIndex])
	fmt.Printf("[cluster] Node %d has been shut down\n", nodeIndex)
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (cluster *Cluster) Stop() {
	cluster.stopGoshimmer()
	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		cluster.stopNode(i)
	}
	cluster.Wait()
}

func (cluster *Cluster) Wait() {
	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		waitCmd(&cluster.waspCmds[i])
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

func (cluster *Cluster) ActiveNodes() []int {
	nodes := make([]int, 0)
	for i := 0; i < cluster.Config.Wasp.NumNodes; i++ {
		if !cluster.IsNodeUp(i) {
			continue
		}
		nodes = append(nodes, i)
	}
	return nodes
}

func (cluster *Cluster) StartMessageCounter(expectations map[string]int) (*MessageCounter, error) {
	return NewMessageCounter(cluster, cluster.Config.AllNodes(), expectations)
}

func (cluster *Cluster) PostTransaction(tx *ledgerstate.Transaction) error {
	fmt.Printf("[cluster] posting request tx: %s\n", tx.ID().String())
	err := cluster.GoshimmerClient().PostTransaction(tx)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.String(), err)
		return err
	}
	if err = cluster.GoshimmerClient().WaitForConfirmation(tx.ID()); err != nil {
		fmt.Printf("[cluster] posting tx: %v\n", err)
		return err
	}
	fmt.Printf("[cluster] request tx confirmed: %s\n", tx.ID().String())
	return nil
}

func (cluster *Cluster) VerifyAddressBalances(addr ledgerstate.Address, totalExpected uint64, expect map[ledgerstate.Color]uint64, comment ...string) bool {
	allOuts, err := cluster.GoshimmerClient().GetConfirmedOutputs(addr)
	if err != nil {
		fmt.Printf("[cluster] GetConfirmedOutputs error: %v\n", err)
		return false
	}
	byColor, total := util.OutputBalancesByColor(allOuts)
	dumpStr, assertionOk := dumpBalancesByColor(byColor, expect)

	totalExpectedStr := "(-)"
	if totalExpected >= 0 {
		if totalExpected == total {
			totalExpectedStr = fmt.Sprintf("(%d) OK", totalExpected)
		} else {
			totalExpectedStr = fmt.Sprintf("(%d) FAIL", totalExpected)
			assertionOk = false
		}
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

func dumpBalancesByColor(actual, expect map[ledgerstate.Color]uint64) (string, bool) {
	assertionOk := true
	lst := make([]ledgerstate.Color, 0, len(expect))
	for col := range expect {
		lst = append(lst, col)
	}
	sort.Slice(lst, func(i, j int) bool {
		return bytes.Compare(lst[i][:], lst[j][:]) < 0
	})
	ret := ""
	for _, col := range lst {
		act, _ := actual[col]
		isOk := "OK"
		if act != expect[col] {
			assertionOk = false
			isOk = "FAIL"
		}
		ret += fmt.Sprintf("         %s: %d (%d)   %s\n", col.String(), act, expect[col], isOk)
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
	sort.Slice(lst, func(i, j int) bool {
		return bytes.Compare(lst[i][:], lst[j][:]) < 0
	})
	ret += "      Unexpected colors in actual outputs:\n"
	for _, col := range lst {
		ret += fmt.Sprintf("         %s %d\n", col.String(), actual[col])
	}
	return ret, assertionOk
}

func interface2bytes(v interface{}) []byte {
	var ret []byte
	switch vt := v.(type) {
	case int:
		ret = util.Uint64To8Bytes(uint64(vt))
	case int16:
		ret = util.Uint64To8Bytes(uint64(vt))
	case int32:
		ret = util.Uint64To8Bytes(uint64(vt))
	case int64:
		ret = util.Uint64To8Bytes(uint64(vt))
	case uint:
		ret = util.Uint64To8Bytes(uint64(vt))
	case uint16:
		ret = util.Uint64To8Bytes(uint64(vt))
	case uint32:
		ret = util.Uint64To8Bytes(uint64(vt))
	case uint64:
		ret = util.Uint64To8Bytes(uint64(vt))
	case []byte:
		ret = vt
	case string:
		ret = []byte(vt)
	default:
		panic("unexpected type")
	}
	return ret
}
