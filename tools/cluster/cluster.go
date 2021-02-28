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
	"strings"
	"text/template"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/tangle"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/client/level1/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type Cluster struct {
	Name    string
	Config  *ClusterConfig
	Started bool

	goshimmerCmd *exec.Cmd
	waspCmds     []*exec.Cmd
}

func New(name string, config *ClusterConfig) *Cluster {
	return &Cluster{
		Name:         name,
		Config:       config,
		goshimmerCmd: nil,
		waspCmds:     make([]*exec.Cmd, config.Wasp.NumNodes),
	}
}

func (clu *Cluster) Level1Client() level1.Level1Client {
	if clu.Config.Goshimmer.Provided {
		return goshimmer.NewGoshimmerClient(clu.Config.goshimmerApiHost())
	}
	return testutil.NewGoshimmerUtxodbClient(clu.Config.goshimmerApiHost())
}

func (clu *Cluster) DeployDefaultChain() (*Chain, error) {
	committee := clu.Config.AllNodes()
	minQuorum := len(committee)/2 + 1
	quorum := len(committee) * 3 / 4
	if quorum < minQuorum {
		quorum = minQuorum
	}
	return clu.DeployChain("Default chain", committee, uint16(quorum))
}

func (clu *Cluster) DeployChain(description string, committeeNodes []int, quorum uint16) (*Chain, error) {
	ownerSeed := seed.NewSeed()

	chain := &Chain{
		Description:    description,
		OriginatorSeed: ownerSeed,
		CommitteeNodes: committeeNodes,
		Quorum:         quorum,
		Cluster:        clu,
	}

	err := clu.Level1Client().RequestFunds(chain.OriginatorAddress())
	if err != nil {
		return nil, err
	}

	chainid, addr, color, err := waspapi.DeployChain(waspapi.CreateChainParams{
		Node:                  clu.Level1Client(),
		CommitteeApiHosts:     chain.ApiHosts(),
		CommitteePeeringHosts: chain.PeeringHosts(),
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
	return cluster.goshimmerCmd != nil
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

	if !cluster.Config.Goshimmer.Provided {
		err = initNodeConfig(
			goshimmerDataPath(dataPath),
			path.Join(templatesPath, "goshimmer-config-template.json"),
			templates.GoshimmerConfig,
			cluster.Config.GoshimmerConfigTemplateParams(),
		)
		if err != nil {
			return err
		}
		err = cluster.copySnapshotBin(
			path.Join(templatesPath, "snapshot.bin"),
			path.Join(goshimmerDataPath(dataPath), "snapshot.bin"),
		)
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

func (cluster *Cluster) copySnapshotBin(srcFilename string, dstFilename string) error {
	exists, err := fileExists(srcFilename)
	if err != nil {
		return err
	}
	if !exists {
		// generate a snapshot from scratch
		const genesisBalance = 1000000000
		seed := ed25519.NewSeed()
		genesisAddr := address.FromED25519PubKey(seed.KeyPair(0).PublicKey)
		snapshot := tangle.Snapshot{
			transaction.GenesisID: {
				genesisAddr: {
					balance.New(balance.ColorIOTA, genesisBalance),
				},
			},
		}

		f, err := os.Create(dstFilename)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = snapshot.WriteTo(f)
		if err != nil {
			return err
		}
		return nil
	} else {
		snapshotSrc, err := os.Open(srcFilename)
		if err != nil {
			return err
		}
		defer snapshotSrc.Close()

		snapshotDest, err := os.Create(dstFilename)
		if err != nil {
			return err
		}
		defer snapshotDest.Close()

		_, err = io.Copy(snapshotDest, snapshotSrc)
		return err
	}
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

	initOk := make(chan bool, cluster.Config.Wasp.NumNodes)

	if !cluster.Config.Goshimmer.Provided {
		cmd, err := cluster.startServer("goshimmer", goshimmerDataPath(dataPath), "goshimmer", initOk, "WebAPI started")
		if err != nil {
			return err
		}
		cluster.goshimmerCmd = cmd

		select {
		case <-initOk:
		case <-time.After(10 * time.Second):
			return fmt.Errorf("Timeout starting goshimmer node\n")
		}
		fmt.Printf("[cluster] started goshimmer node\n")
	}

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
	url := cluster.Config.goshimmerApiHost()
	fmt.Printf("[cluster] Sending shutdown to goshimmer at %s\n", url)
	err := nodeapi.Shutdown(url)
	if err != nil {
		fmt.Println(err)
	}
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
	waitCmd(&cluster.goshimmerCmd)
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

func (cluster *Cluster) PostTransaction(tx *sctransaction.TransactionEssence) error {
	fmt.Printf("[cluster] posting request tx: %s\n", tx.ID().String())
	err := cluster.Level1Client().PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.Transaction.String(), err)
		return err
	}
	fmt.Printf("[cluster] request tx confirmed: %s\n", tx.ID().String())
	return nil
}

func (cluster *Cluster) VerifyAddressBalances(addr *address.Address, totalExpected int64, expect map[balance.Color]int64, comment ...string) bool {
	allOuts, err := cluster.Level1Client().GetConfirmedAccountOutputs(addr)
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
	if model.IsHTTPNotFound(err) {
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
	for _, host := range cluster.Config.ApiHosts(cluster.ActiveNodes()) {
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
