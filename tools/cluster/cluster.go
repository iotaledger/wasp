package cluster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/apilib"
)

type ClusterConfig struct {
	Nodes []struct {
		BindAddress string `json:"bindAddress"`
		PeeringPort int    `json:"PeeringPort"`
	} `json:"nodes"`
}

type Cluster struct {
	Config     *ClusterConfig
	ConfigPath string // where the cluster configuration is stored - read only
	DataPath   string // where the cluster's volatile data lives
	cmds       []*exec.Cmd
}

func readConfig(configPath string) *ClusterConfig {
	data, err := ioutil.ReadFile(path.Join(configPath, "cluster.json"))
	if err != nil {
		panic(err)
	}

	config := &ClusterConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func New(configPath string, dataPath string) *Cluster {
	config := readConfig(configPath)
	return &Cluster{
		Config:     config,
		ConfigPath: configPath,
		DataPath:   dataPath,
		cmds:       make([]*exec.Cmd, 0),
	}
}

func (cluster *Cluster) Path(s string) string {
	return path.Join(cluster.ConfigPath, s)
}

func (cluster *Cluster) ConfigTemplatePath() string {
	return cluster.Path("wasp-config-template.json")
}

func (cluster *Cluster) WaspNodePath(i int) string {
	return path.Join(cluster.DataPath, strconv.Itoa(i))
}

// Init creates in DataPath a directory with config.json for each node
func (cluster *Cluster) Init() *Cluster {
	if _, err := os.Stat(cluster.DataPath); err == nil {
		fmt.Printf("%s directory exists. Delete it first.\n", cluster.DataPath)
		os.Exit(1)
	} else if !os.IsNotExist(err) {
		panic(err)
	}

	configTmpl, err := template.ParseFiles(cluster.ConfigTemplatePath())
	if err != nil {
		panic(err)
	}

	for i, nodeConfig := range cluster.Config.Nodes {
		nodePath := cluster.WaspNodePath(i)
		fmt.Printf("Initializing node configuration at %s\n", nodePath)

		err := os.MkdirAll(nodePath, os.ModePerm)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(path.Join(nodePath, "config.json"))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = configTmpl.Execute(f, &nodeConfig)
		if err != nil {
			panic(err)
		}
	}

	return cluster
}

func logNode(i int, scanner *bufio.Scanner, initString string, initOk chan bool) {
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if !found && strings.Contains(line, initString) {
			initOk <- true
			found = true
		}
		fmt.Printf("[wasp %d] %s\n", i, line)
	}
}

// Start launches all wasp nodes in the cluster, each running in its own directory
func (cluster *Cluster) Start() {
	initOk := make(chan bool, len(cluster.Config.Nodes))

	for i, _ := range cluster.Config.Nodes {
		cmd := exec.Command("wasp")
		cmd.Dir = cluster.WaspNodePath(i)
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(pipe)
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		cluster.cmds = append(cluster.cmds, cmd)

		go logNode(i, scanner, "WebAPI started", initOk)
	}

	for i := 0; i < len(cluster.Config.Nodes); i++ {
		<-initOk
	}
}

// Stop sends an interrupt signal to all nodes and waits for them to exit
func (cluster *Cluster) Stop() {
	for _, node := range cluster.Config.Nodes {
		fmt.Printf("Sending shutdown to %s\n", node.BindAddress)
		err := apilib.Shutdown(node.BindAddress)
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

func (cluster *Cluster) Hosts() []string {
	hosts := make([]string, 0)
	for _, node := range cluster.Config.Nodes {
		hosts = append(hosts, node.BindAddress)
	}
	return hosts
}

func (cluster *Cluster) GenerateNewDistributedKeySet(quorum int) *address.Address {
	addr, err := apilib.GenerateNewDistributedKeySet(cluster.Hosts(), uint16(len(cluster.Config.Nodes)), uint16(quorum))
	if err != nil {
		panic(err)
	}
	return addr
}
