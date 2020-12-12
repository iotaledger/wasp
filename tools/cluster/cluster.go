package cluster

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/nodeclient/goshimmer"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/util"
)

type Cluster struct {
	Config     *ClusterConfig
	ConfigPath string // where the cluster configuration is stored - read only
	DataPath   string // where the cluster's volatile data lives
	Started    bool
	NodeClient nodeclient.NodeClient
	// reading publisher's output
	messagesCh   chan *subscribe.HostMessage
	stopReading  chan bool
	expectations map[string]int
	topics       []string
	counters     map[string]map[string]int
	testName     string
}

func New(configPath string, dataPath string) (*Cluster, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fmt.Printf("[cluster] current working directory is %s\n", wd)
	fmt.Printf("[cluster] config part is %s\n", configPath)
	fmt.Printf("[cluster] data part is %s\n", dataPath)

	config, err := readConfig(configPath)
	if err != nil {
		return nil, err
	}

	var nodeClient nodeclient.NodeClient
	if config.Goshimmer.Provided {
		nodeClient = goshimmer.NewGoshimmerClient(config.goshimmerApiHost())
	} else {
		nodeClient = testutil.NewGoshimmerUtxodbClient(config.goshimmerApiHost())
	}

	return &Cluster{
		Config:     config,
		ConfigPath: configPath,
		DataPath:   dataPath,
		NodeClient: nodeClient,
	}, nil
}

func (clu *Cluster) DeployDefaultChain() (*Chain, error) {
	committee := clu.Config.AllWaspNodes()
	quorum := uint16(len(committee) * 3 / 4)
	return clu.DeployChain("Default chain", committee, quorum, []int{})
}

func (clu *Cluster) DeployChain(description string, committeeNodes []int, quorum uint16, accessNodes []int) (*Chain, error) {
	ownerSeed := seed.NewSeed()

	chain := &Chain{
		Description:    description,
		OriginatorSeed: ownerSeed,
		CommitteeNodes: committeeNodes,
		AccessNodes:    accessNodes,
		Quorum:         quorum,
		Cluster:        clu,
	}

	err := clu.NodeClient.RequestFunds(chain.OriginatorAddress())
	if err != nil {
		return nil, err
	}

	chainid, addr, color, err := waspapi.DeployChain(waspapi.CreateChainParams{
		Node:                  clu.NodeClient,
		CommitteeApiHosts:     clu.WaspHosts(committeeNodes, (*WaspNodeConfig).ApiHost),
		CommitteePeeringHosts: clu.WaspHosts(committeeNodes, (*WaspNodeConfig).PeeringHost),
		AccessNodes:           clu.WaspHosts(accessNodes, (*WaspNodeConfig).PeeringHost),
		N:                     uint16(len(committeeNodes)),
		T:                     quorum,
		OriginatorSigScheme:   chain.OriginatorSigScheme(),
		Description:           description,
		Textout:               os.Stdout,
		Prefix:                "[cluster] ",
	})
	if err != nil {
		return nil, err
	}

	chain.Address = *addr
	chain.ChainID = *chainid
	chain.Color = *color

	return chain, nil
}

func (cluster *Cluster) IsGoshimmerUp() bool {
	return cluster.Config.Goshimmer.cmd != nil
}

func (cluster *Cluster) MultiClient() *multiclient.MultiClient {
	return multiclient.New(cluster.ApiHosts())
}

func (cluster *Cluster) WaspClient(nodeIndex int) *client.WaspClient {
	return cluster.Config.Nodes[nodeIndex].Client()
}

func (cluster *Cluster) JoinConfigPath(s string) string {
	return path.Join(cluster.ConfigPath, s)
}

func (cluster *Cluster) GoshimmerSnapshotBinPath() string {
	return cluster.JoinConfigPath("snapshot.bin")
}
func (cluster *Cluster) GoshimmerConfigTemplatePath() string {
	return cluster.JoinConfigPath("goshimmer-config-template.json")
}

func (cluster *Cluster) WaspConfigTemplatePath() string {
	return cluster.JoinConfigPath("wasp-config-template.json")
}

func (cluster *Cluster) ConfigKeysPath() string {
	return cluster.JoinConfigPath("keys.json")
}

func (cluster *Cluster) WaspNodeDataPath(i int) string {
	return path.Join(cluster.DataPath, fmt.Sprintf("wasp%d", i))
}

func (cluster *Cluster) GoshimmerDataPath() string {
	return path.Join(cluster.DataPath, "goshimmer")
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

// Init creates in DataPath a directory with config.json for each node
func (cluster *Cluster) Init(resetDataPath bool, name string) error {
	cluster.testName = name
	exists, err := fileExists(cluster.DataPath)
	if err != nil {
		return err
	}
	if exists {
		if !resetDataPath {
			return fmt.Errorf("%s directory exists", cluster.DataPath)
		}
		err = os.RemoveAll(cluster.DataPath)
		if err != nil {
			return err
		}
	}

	if !cluster.Config.Goshimmer.Provided {
		err = cluster.initDataPath(
			cluster.GoshimmerDataPath(),
			cluster.GoshimmerConfigTemplatePath(),
			cluster.Config.Goshimmer,
		)
		if err != nil {
			return err
		}
		err = cluster.copySnapshotBin()
		if err != nil {
			return err
		}
	}
	for i, waspParams := range cluster.Config.Nodes {
		err = cluster.initDataPath(
			cluster.WaspNodeDataPath(i),
			cluster.WaspConfigTemplatePath(),
			waspParams,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) initDataPath(dataPath string, configTemplatePath string, params interface{}) error {
	configTmpl, err := template.ParseFiles(configTemplatePath)
	if err != nil {
		return err
	}

	fmt.Printf("Initializing %s\n", dataPath)

	err = os.MkdirAll(dataPath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(dataPath, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	return configTmpl.Execute(f, params)
}

func (cluster *Cluster) copySnapshotBin() error {
	snapshotSrc, err := os.Open(cluster.GoshimmerSnapshotBinPath())
	if err != nil {
		return err
	}
	defer snapshotSrc.Close()

	snapshotDest, err := os.Create(path.Join(cluster.GoshimmerDataPath(), "snapshot.bin"))
	if err != nil {
		return err
	}
	defer snapshotDest.Close()

	_, err = io.Copy(snapshotDest, snapshotSrc)
	return err
}

// Start launches all wasp nodes in the cluster, each running in its own directory
func (cluster *Cluster) Start() error {
	exists, err := fileExists(cluster.DataPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Data path %s does not exist", cluster.DataPath)
	}

	err = cluster.start()
	if err != nil {
		return err
	}

	cluster.Started = true
	return nil
}

func (cluster *Cluster) start() error {
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", len(cluster.Config.Nodes))

	initOk := make(chan bool, len(cluster.Config.Nodes))

	if !cluster.Config.Goshimmer.Provided {
		cmd, err := cluster.startServer("goshimmer", cluster.GoshimmerDataPath(), "goshimmer", initOk, "WebAPI started")
		if err != nil {
			return err
		}
		cluster.Config.Goshimmer.cmd = cmd

		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return fmt.Errorf("Timeout starting goshimmer node\n")
		}
		fmt.Printf("[cluster] started goshimmer node\n")
	}

	for i, _ := range cluster.Config.Nodes {
		cmd, err := cluster.startServer("wasp", cluster.WaspNodeDataPath(i), fmt.Sprintf("wasp %d", i), initOk, "nanomsg publisher is running")
		if err != nil {
			return err
		}
		cluster.Config.Nodes[i].cmd = cmd
	}

	for i := 0; i < len(cluster.Config.Nodes); i++ {
		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return fmt.Errorf("Timeout starting wasp nodes\n")
		}
	}
	fmt.Printf("[cluster] started %d Wasp nodes\n", len(cluster.Config.Nodes))
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
	url := cluster.Config.goshimmerApiHost()
	fmt.Printf("[cluster] Sending shutdown to goshimmer at %s\n", url)
	err := nodeapi.Shutdown(url)
	if err != nil {
		fmt.Println(err)
	}
}

func (cluster *Cluster) stopNode(nodeIndex int) {
	node := cluster.Config.Nodes[nodeIndex]
	if !node.IsUp() {
		return
	}
	fmt.Printf("[cluster] Sending shutdown to wasp node at %s\n", node.ApiHost())
	err := node.Client().Shutdown()
	if err != nil {
		fmt.Println(err)
	}
}

func (cluster *Cluster) StopNode(nodeIndex int) {
	cluster.stopNode(nodeIndex)
	waitCmd(&cluster.Config.Nodes[nodeIndex].cmd)
	fmt.Printf("[cluster] Node %s has been shut down\n", cluster.Config.Nodes[nodeIndex].ApiHost())
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (cluster *Cluster) Stop() {
	cluster.stopGoshimmer()
	for i := range cluster.Config.Nodes {
		cluster.stopNode(i)
	}

	cluster.Wait()
}

func (cluster *Cluster) Wait() {
	waitCmd(&cluster.Config.Goshimmer.cmd)
	for _, node := range cluster.Config.Nodes {
		waitCmd(&node.cmd)
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

func (cluster *Cluster) ActiveApiHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		if !node.IsUp() {
			continue
		}
		url := node.ApiHost()
		hosts = append(hosts, url)
	}
	return hosts
}

func (cluster *Cluster) ApiHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		url := node.ApiHost()
		hosts = append(hosts, url)
	}
	return hosts
}

func (cluster *Cluster) PeeringHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		url := node.PeeringHost()
		hosts = append(hosts, url)
	}
	return hosts
}

func (cluster *Cluster) PublisherHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		url := node.NanomsgHost()
		hosts = append(hosts, url)
	}
	return hosts
}

func (cluster *Cluster) ActivePublisherHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		if !node.IsUp() {
			continue
		}
		url := node.NanomsgHost()
		hosts = append(hosts, url)
	}
	return hosts
}

func (cluster *Cluster) AllWaspNodes() []int {
	r := make([]int, 0)
	for i := range cluster.Config.Nodes {
		r = append(r, i)
	}
	return r
}

func (cluster *Cluster) WaspHosts(nodeIndexes []int, getHost func(w *WaspNodeConfig) string) []string {
	hosts := make([]string, 0)
	for _, i := range nodeIndexes {
		if i < 0 || i > len(cluster.Config.Nodes)-1 {
			panic(fmt.Sprintf("Node index out of bounds in smart contract configuration: %d", i))
		}
		hosts = append(hosts, getHost(cluster.Config.Nodes[i]))
	}
	return hosts
}

func (cluster *Cluster) ListenToMessages(expectations map[string]int) error {
	allNodesNanomsg := cluster.WaspHosts(cluster.AllWaspNodes(), (*WaspNodeConfig).NanomsgHost)

	cluster.expectations = expectations
	cluster.messagesCh = make(chan *subscribe.HostMessage, 1000)
	cluster.stopReading = make(chan bool)
	cluster.counters = make(map[string]map[string]int)

	cluster.topics = make([]string, 0)
	for t := range expectations {
		cluster.topics = append(cluster.topics, t)
	}
	sort.Strings(cluster.topics)

	for _, host := range allNodesNanomsg {
		cluster.counters[host] = make(map[string]int)
		for msgType := range expectations {
			cluster.counters[host][msgType] = 0
		}
	}

	return subscribe.SubscribeMultiOld(allNodesNanomsg, cluster.messagesCh, cluster.stopReading, cluster.topics...)
}

func (cluster *Cluster) CollectMessages(duration time.Duration) {
	fmt.Printf("[cluster] collecting publisher's messages for %v\n", duration)

	deadline := time.Now().Add(duration)
	for {
		select {
		case msg := <-cluster.messagesCh:
			cluster.countMessage(msg)

		case <-time.After(500 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			break
		}
	}
}

func (cluster *Cluster) WaitUntilExpectationsMet() bool {
	fmt.Printf("[cluster] collecting publisher's messages\n")

	for {
		fail, pass, report := cluster.report()
		if fail {
			fmt.Printf("\n[cluster] Message expectations failed for '%s':\n%s\n", cluster.testName, report)
			return false
		}
		if pass {
			return true
		}

		select {
		case msg := <-cluster.messagesCh:
			cluster.countMessage(msg)
		case <-time.After(90 * time.Second):
			return cluster.Report()
		}
	}
}

func (cluster *Cluster) countMessage(msg *subscribe.HostMessage) {
	cluster.counters[msg.Sender][msg.Message[0]] += 1
}

func (cluster *Cluster) Report() bool {
	_, pass, report := cluster.report()
	fmt.Printf("\n[cluster] Message statistics for '%s':\n%s\n", cluster.testName, report)
	return pass
}

func (cluster *Cluster) report() (bool, bool, string) {
	fail := false
	pass := true
	report := ""
	for host, counters := range cluster.counters {
		report += fmt.Sprintf("Node: %s\n", host)
		for _, t := range cluster.topics {
			res := counters[t]
			exp := cluster.expectations[t]
			e := "-"
			f := ""
			if exp >= 0 {
				e = strconv.Itoa(exp)
				if res == exp {
					f = "ok"
				} else {
					f = "fail"
					pass = false
					if res > exp {
						// got more messages than expected, no need to keep running
						fail = true
					}
				}
			}
			report += fmt.Sprintf("          %s: %d (%s) %s\n", t, res, e, f)
		}
	}
	return fail, pass, report
}

func (cluster *Cluster) PostTransaction(tx *sctransaction.Transaction) error {
	fmt.Printf("[cluster] posting request tx: %s\n", tx.ID().String())
	err := cluster.NodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.Transaction.String(), err)
		return err
	}
	fmt.Printf("[cluster] request tx confirmed: %s\n", tx.ID().String())
	return nil
}

func (cluster *Cluster) VerifyAddressBalances(addr *address.Address, totalExpected int64, expect map[balance.Color]int64, comment ...string) bool {
	allOuts, err := cluster.NodeClient.GetConfirmedAccountOutputs(addr)
	if err != nil {
		fmt.Printf("[cluster] GetConfirmedAccountOutputs error: %v\n", err)
		return false
	}
	byColor, total := txutil.OutputBalancesByColor(allOuts)
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
	fmt.Printf("[cluster] Balances of the address %s%s\n      Total tokens: %d %s\n%s\n",
		addr.String(), cmt, total, totalExpectedStr, dumpStr)

	if !assertionOk {
		fmt.Printf("[cluster] assertion on balances failed\n")
	}
	return assertionOk
}

func verifySCStateVariables2(host string, addr *address.Address, expectedValues map[kv.Key]interface{}) bool {
	contractID := coretypes.NewContractID((coretypes.ChainID)(*addr), 0)
	actual, err := client.NewWaspClient(host).DumpSCState(&contractID)
	if client.IsNotFound(err) {
		fmt.Printf("              state does not exist: FAIL\n")
		return false
	}
	if err != nil {
		panic(err)
	}
	pass := true
	fmt.Printf("    host %s, state index #%d\n", host, actual.Index)
	for k, vexp := range expectedValues {
		vact, _ := actual.Variables.Get(k)
		if vact == nil {
			vact = []byte("N/A")
		}
		vres := "FAIL"
		if bytes.Equal(interface2bytes(vexp), vact) {
			vres = "OK"
		} else {
			pass = false
		}
		// TODO prettier output?
		var actualValue interface{}
		switch vexp.(type) {
		case string:
			actualValue = string(vact)
		case []byte:
			actualValue = vact
		default:
			if len(vact) == 8 {
				actualValue = util.MustUint64From8Bytes(vact)
			} else {
				actualValue = vact
			}
		}
		fmt.Printf("      '%s': %v (%v) -- %s\n", k, actualValue, vexp, vres)
	}
	return pass
}

func (cluster *Cluster) VerifySCStateVariables2(addr *address.Address, expectedValues map[kv.Key]interface{}) bool {
	fmt.Printf("verifying state variables for address %s\n", addr.String())
	pass := true
	for _, host := range cluster.ActiveApiHosts() {
		pass = pass && verifySCStateVariables2(host, addr, expectedValues)
	}
	return pass
}

func dumpBalancesByColor(actual, expect map[balance.Color]int64) (string, bool) {
	assertionOk := true
	lst := make([]balance.Color, 0, len(expect))
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
