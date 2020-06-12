package cluster

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/subscribe"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	nodeapi "github.com/iotaledger/wasp/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
)

type SmartContractFinalConfig struct {
	Address          string   `json:"address"`
	Color            string   `json:"color"`
	Description      string   `json:"description"`
	ProgramHash      string   `json:"program_hash"`
	CommitteeNodes   []int    `json:"committee_nodes"`
	AccessNodes      []int    `json:"access_nodes,omitempty"`
	OwnerIndexUtxodb int      `json:"owner_index_utxodb"`
	DKShares         []string `json:"dkshares"` // [node index]
}

type SmartContractInitData struct {
	Description    string `json:"description"`
	CommitteeNodes []int  `json:"committee_nodes"`
	AccessNodes    []int  `json:"access_nodes,omitempty"`
	Quorum         int    `json:"quorum"`
}

type WaspNodeConfig struct {
	NetAddress  string `json:"net_address"`
	ApiPort     int    `json:"api_port"`
	PeeringPort int    `json:"peering_port"`
	NanomsgPort int    `json:"nanomsg_port"`
}

type ClusterConfig struct {
	Nodes     []WaspNodeConfig `json:"nodes"`
	Goshimmer struct {
		BindAddress string `json:"bind_address"`
	} `json:"goshimmer"`
	SmartContracts []SmartContractInitData `json:"smart_contracts"`
}

type Cluster struct {
	Config              *ClusterConfig
	SmartContractConfig []SmartContractFinalConfig
	ConfigPath          string // where the cluster configuration is stored - read only
	DataPath            string // where the cluster's volatile data lives
	Started             bool
	cmds                []*exec.Cmd
	// reading publisher's output
	messagesCh  chan *subscribe.HostMessage
	stopReading chan bool
	counters    map[string]int
}

func (sc *SmartContractFinalConfig) AllNodes() []int {
	r := make([]int, 0)
	r = append(r, sc.CommitteeNodes...)
	r = append(r, sc.AccessNodes...)
	return r
}

func (w *WaspNodeConfig) ApiHost() string {
	return fmt.Sprintf("%s:%d", w.NetAddress, w.ApiPort)
}

func (w *WaspNodeConfig) PeeringHost() string {
	return fmt.Sprintf("%s:%d", w.NetAddress, w.PeeringPort)
}

func (w *WaspNodeConfig) NanomsgHost() string {
	return fmt.Sprintf("%s:%d", w.NetAddress, w.NanomsgPort)
}

func readConfig(configPath string) (*ClusterConfig, error) {
	data, err := ioutil.ReadFile(path.Join(configPath, "cluster.json"))
	if err != nil {
		return nil, err
	}

	config := &ClusterConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
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
	return &Cluster{
		Config:     config,
		ConfigPath: configPath,
		DataPath:   dataPath,
		cmds:       make([]*exec.Cmd, 0),
	}, nil
}

func (cluster *Cluster) readKeysConfig() ([]SmartContractFinalConfig, error) {
	data, err := ioutil.ReadFile(cluster.ConfigKeysPath())
	if err != nil {
		return nil, err
	}

	config := make([]SmartContractFinalConfig, 0)
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (cluster *Cluster) JoinConfigPath(s string) string {
	return path.Join(cluster.ConfigPath, s)
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
func (cluster *Cluster) Init(resetDataPath bool) error {
	exists, err := fileExists(cluster.DataPath)
	if err != nil {
		return err
	}
	if exists {
		if !resetDataPath {
			return errors.New(fmt.Sprintf("%s directory exists", cluster.DataPath))
		}
		err = os.RemoveAll(cluster.DataPath)
		if err != nil {
			return err
		}
	}

	err = cluster.initDataPath(
		cluster.GoshimmerDataPath(),
		cluster.GoshimmerConfigTemplatePath(),
		cluster.Config.Goshimmer,
	)
	if err != nil {
		return err
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

// Start launches all wasp nodes in the cluster, each running in its own directory
func (cluster *Cluster) Start() error {
	exists, err := fileExists(cluster.DataPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("Data path %s does not exist", cluster.DataPath))
	}

	err = cluster.start()
	if err != nil {
		return err
	}

	keysExist, err := cluster.readKeysAndData()
	if err != nil {
		return err
	}
	if keysExist {
		err = cluster.importKeys()
		if err != nil {
			return err
		}

	} else {
		fmt.Printf("[cluster] keys.json does not exist\n")
	}
	cluster.Started = true
	return nil
}

func (cluster *Cluster) start() error {
	fmt.Printf("[cluster] starting %d Wasp nodes...\n", len(cluster.Config.Nodes))

	initOk := make(chan bool, len(cluster.Config.Nodes))

	err := cluster.startServer("goshimmer", cluster.GoshimmerDataPath(), "goshimmer", initOk, "WebAPI started")
	if err != nil {
		return err
	}

	select {
	case <-initOk:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Timeout starting goshimmer node\n")
	}
	fmt.Printf("[cluster] started goshimmer node\n")

	for i, _ := range cluster.Config.Nodes {
		err = cluster.startServer("wasp", cluster.WaspNodeDataPath(i), fmt.Sprintf("wasp %d", i), initOk, "nanomsg publisher is running")
		if err != nil {
			return err
		}
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

func (cluster *Cluster) startServer(command string, cwd string, name string, initOk chan<- bool, initOkMsg string) error {
	cmd := exec.Command(command)
	cmd.Dir = cwd
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	cluster.cmds = append(cluster.cmds, cmd)

	go scanLog(
		stderrPipe,
		func(line string) { fmt.Printf("[!%s] %s\n", name, line) },
	)
	go scanLog(
		stdoutPipe,
		func(line string) { fmt.Printf("[ %s] %s\n", name, line) },
		waitFor(initOkMsg, initOk),
	)

	return nil
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

func (cluster *Cluster) readKeysAndData() (bool, error) {
	exists, err := fileExists(cluster.ConfigKeysPath())
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	fmt.Printf("[cluster] loading keys and smart contract data from %s\n", cluster.ConfigKeysPath())
	cluster.SmartContractConfig, err = cluster.readKeysConfig()
	return true, nil
}

func (cluster *Cluster) importKeys() error {
	for _, scKeys := range cluster.SmartContractConfig {
		fmt.Printf("[cluster] Importing DKShares for address %s...\n", scKeys.Address)
		for nodeIndex, dks := range scKeys.DKShares {
			url := fmt.Sprintf("%s:%d", cluster.Config.Nodes[nodeIndex].NetAddress, cluster.Config.Nodes[nodeIndex].ApiPort)
			err := waspapi.ImportDKShare(url, dks)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (cluster *Cluster) Stop() {
	fmt.Printf("[cluster] Sending shutdown to goshimmer at %s\n", cluster.Config.Goshimmer.BindAddress)
	err := nodeapi.Shutdown(cluster.Config.Goshimmer.BindAddress)
	if err != nil {
		fmt.Println(err)
	}

	for _, node := range cluster.Config.Nodes {
		url := fmt.Sprintf("%s:%d", node.NetAddress, node.ApiPort)
		fmt.Printf("[cluster] Sending shutdown to wasp node at %s\n", url)
		err := waspapi.Shutdown(url)
		if err != nil {
			fmt.Println(err)
		}
	}
	cluster.Wait()
}

// Wait blocks until all nodes exit
func (cluster *Cluster) Wait() {
	for _, cmd := range cluster.cmds {
		err := cmd.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (cluster *Cluster) ApiHosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		url := fmt.Sprintf("%s:%d", node.NetAddress, node.ApiPort)
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
		hosts = append(hosts, getHost(&cluster.Config.Nodes[i]))
	}
	return hosts
}

func (cluster *Cluster) ListenToMessages(topics ...string) error {
	allNodesNanomsg := cluster.WaspHosts(cluster.AllWaspNodes(), (*WaspNodeConfig).NanomsgHost)

	fmt.Printf("[cluster] will be listening on topics %+v from %+v\n", topics, allNodesNanomsg)

	cluster.messagesCh = make(chan *subscribe.HostMessage, 1000)
	cluster.stopReading = make(chan bool)
	cluster.counters = make(map[string]int)

	for _, host := range allNodesNanomsg {
		for _, t := range topics {
			cluster.counters[t+"--"+host] = 0
		}
	}

	return subscribe.SubscribeMulti(allNodesNanomsg, cluster.messagesCh, cluster.stopReading, topics...)
}

func (cluster *Cluster) CountMessages(duration time.Duration) ([]*subscribe.HostMessage, map[string]int) {
	fmt.Printf("[cluster] counting publisher's messages for %v\n", duration)

	all := make([]*subscribe.HostMessage, 0)

	deadline := time.Now().Add(duration)
	for {
		select {
		case msg := <-cluster.messagesCh:
			all = append(all, msg)
			cluster.counters[msg.Message[0]+"--"+msg.Sender] += 1

		case <-time.After(500 * time.Millisecond):

		}
		if time.Now().After(deadline) {
			break
		}
	}
	return all, cluster.counters
}
