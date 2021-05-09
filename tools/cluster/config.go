package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type GoshimmerConfig struct {
	TxStreamPort int
	ApiPort      int
	Provided     bool
}

type WaspConfig struct {
	NumNodes int

	// node ports are calculated as these values + node index
	FirstApiPort       int
	FirstPeeringPort   int
	FirstNanomsgPort   int
	FirstDashboardPort int
}

type ClusterConfig struct {
	Wasp      WaspConfig
	Goshimmer GoshimmerConfig
}

func DefaultConfig() *ClusterConfig {
	return &ClusterConfig{
		Wasp: WaspConfig{
			NumNodes:           4,
			FirstApiPort:       9090,
			FirstPeeringPort:   4000,
			FirstNanomsgPort:   5550,
			FirstDashboardPort: 7000,
		},
		Goshimmer: GoshimmerConfig{
			TxStreamPort: 5000,
			ApiPort:      8080,
			Provided:     false,
		},
	}
}

func ConfigExists(dataPath string) (bool, error) {
	return fileExists(configPath(dataPath))
}

func LoadConfig(dataPath string) (*ClusterConfig, error) {
	b, err := ioutil.ReadFile(configPath(dataPath))
	if err != nil {
		return nil, err
	}
	var c ClusterConfig
	err = json.Unmarshal(b, &c)
	return &c, err
}

func (c *ClusterConfig) Save(dataPath string) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configPath(dataPath), b, 0644)
}

func configPath(dataPath string) string {
	return path.Join(dataPath, "cluster.json")
}

func (c *ClusterConfig) goshimmerApiHost() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Goshimmer.ApiPort)
}

func (c *ClusterConfig) waspHosts(nodeIndexes []int, getHost func(i int) string) []string {
	hosts := make([]string, 0)
	for _, i := range nodeIndexes {
		if i < 0 || i > c.Wasp.NumNodes-1 {
			panic(fmt.Sprintf("Node index out of bounds in smart contract configuration: %d", i))
		}
		hosts = append(hosts, getHost(i))
	}
	return hosts
}

func (c *ClusterConfig) AllNodes() []int {
	nodes := make([]int, c.Wasp.NumNodes)
	for i := 0; i < c.Wasp.NumNodes; i++ {
		nodes[i] = i
	}
	return nodes
}

func (c *ClusterConfig) ApiHosts(nodeIndexes ...[]int) []string {
	nodes := c.AllNodes()
	if len(nodeIndexes) == 1 {
		nodes = nodeIndexes[0]
	}
	return c.waspHosts(nodes, func(i int) string { return c.ApiHost(i) })
}

func (c *ClusterConfig) ApiHost(nodeIndex int) string {
	return fmt.Sprintf("127.0.0.1:%d", c.ApiPort(nodeIndex))
}

func (c *ClusterConfig) ApiPort(nodeIndex int) int {
	return c.Wasp.FirstApiPort + nodeIndex
}

func (c *ClusterConfig) PeeringHosts(nodeIndexes ...[]int) []string {
	nodes := c.AllNodes()
	if len(nodeIndexes) == 1 {
		nodes = nodeIndexes[0]
	}
	return c.waspHosts(nodes, func(i int) string { return c.PeeringHost(i) })
}

func (c *ClusterConfig) NeighborsString() string {
	ret := make([]string, c.Wasp.NumNodes)
	for i := range ret {
		ret[i] = "\"" + c.PeeringHost(i) + "\""
	}
	return strings.Join(ret, ",")
}

func (c *ClusterConfig) PeeringHost(nodeIndex int) string {
	return fmt.Sprintf("127.0.0.1:%d", c.PeeringPort(nodeIndex))
}

func (c *ClusterConfig) PeeringPort(nodeIndex int) int {
	return c.Wasp.FirstPeeringPort + nodeIndex
}

func (c *ClusterConfig) NanomsgHosts(nodeIndexes ...[]int) []string {
	nodes := c.AllNodes()
	if len(nodeIndexes) == 1 {
		nodes = nodeIndexes[0]
	}
	return c.waspHosts(nodes, func(i int) string { return c.NanomsgHost(i) })
}

func (c *ClusterConfig) NanomsgHost(nodeIndex int) string {
	return fmt.Sprintf("127.0.0.1:%d", c.NanomsgPort(nodeIndex))
}

func (c *ClusterConfig) NanomsgPort(nodeIndex int) int {
	return c.Wasp.FirstNanomsgPort + nodeIndex
}

func (c *ClusterConfig) DashboardPort(nodeIndex int) int {
	return c.Wasp.FirstDashboardPort + nodeIndex
}

func (c *ClusterConfig) WaspConfigTemplateParams(i int) *templates.WaspConfigParams {
	return &templates.WaspConfigParams{
		ApiPort:       c.ApiPort(i),
		DashboardPort: c.DashboardPort(i),
		PeeringPort:   c.PeeringPort(i),
		NanomsgPort:   c.NanomsgPort(i),
		Neighbors:     c.NeighborsString(),
	}
}
