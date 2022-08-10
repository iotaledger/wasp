package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type WaspConfig struct {
	NumNodes int

	// node ports are calculated as these values + node index
	FirstAPIPort       int
	FirstPeeringPort   int
	FirstNanomsgPort   int
	FirstDashboardPort int
	FirstProfilingPort int
	FirstMetricsPort   int
}

func (w *WaspConfig) WaspConfigTemplateParams(i int) templates.WaspConfigParams {
	return templates.WaspConfigParams{
		APIPort:                      w.FirstAPIPort + i,
		DashboardPort:                w.FirstDashboardPort + i,
		PeeringPort:                  w.FirstPeeringPort + i,
		NanomsgPort:                  w.FirstNanomsgPort + i,
		ProfilingPort:                w.FirstProfilingPort + i,
		MetricsPort:                  w.FirstMetricsPort + i,
		OffledgerBroadcastUpToNPeers: 10,
	}
}

type ClusterConfig struct {
	Wasp []templates.WaspConfigParams
	L1   nodeconn.L1Config
}

func DefaultWaspConfig() WaspConfig {
	return WaspConfig{
		NumNodes:           4,
		FirstAPIPort:       9090,
		FirstPeeringPort:   4000,
		FirstNanomsgPort:   5550,
		FirstDashboardPort: 7000,
		FirstProfilingPort: 1060,
		FirstMetricsPort:   2112,
	}
}

func ConfigExists(dataPath string) (bool, error) {
	return fileExists(configPath(dataPath))
}

func NewConfig(waspConfig WaspConfig, l1Config nodeconn.L1Config, modifyConfig ...templates.ModifyNodesConfigFn) *ClusterConfig {
	nodesConfigs := make([]templates.WaspConfigParams, waspConfig.NumNodes)
	for i := 0; i < waspConfig.NumNodes; i++ {
		// generate template from waspconfigs
		nodesConfigs[i] = waspConfig.WaspConfigTemplateParams(i)
		// set L1 part of the template
		nodesConfigs[i].L1APIAddress = l1Config.APIAddress
		nodesConfigs[i].L1UseRemotePow = l1Config.UseRemotePoW
		// modify the template if needed
		if len(modifyConfig) > 0 && modifyConfig[0] != nil {
			nodesConfigs[i] = modifyConfig[0](i, nodesConfigs[i])
		}
	}

	return &ClusterConfig{
		Wasp: nodesConfigs,
		L1:   l1Config,
	}
}

func LoadConfig(dataPath string) (*ClusterConfig, error) {
	b, err := os.ReadFile(configPath(dataPath))
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
	return os.WriteFile(configPath(dataPath), b, 0o600)
}

func configPath(dataPath string) string {
	return path.Join(dataPath, "cluster.json")
}

func (c *ClusterConfig) SetOwnerAddress(address string) {
	for i := range c.Wasp {
		c.Wasp[i].OwnerAddress = address
	}
}

func (c *ClusterConfig) waspHosts(nodeIndexes []int, getHost func(i int) string) []string {
	hosts := make([]string, 0)
	for _, i := range nodeIndexes {
		if i < 0 || i > len(c.Wasp)-1 {
			panic(fmt.Sprintf("Node index out of bounds in smart contract configuration: %d/%d", i, len(c.Wasp)))
		}
		hosts = append(hosts, getHost(i))
	}
	return hosts
}

func (c *ClusterConfig) AllNodes() []int {
	nodes := make([]int, len(c.Wasp))
	for i := 0; i < len(c.Wasp); i++ {
		nodes[i] = i
	}
	return nodes
}

func (c *ClusterConfig) APIHosts(nodeIndexes ...[]int) []string {
	nodes := c.AllNodes()
	if len(nodeIndexes) == 1 {
		nodes = nodeIndexes[0]
	}
	return c.waspHosts(nodes, func(i int) string { return c.APIHost(i) })
}

func (c *ClusterConfig) APIHost(nodeIndex int) string {
	return fmt.Sprintf("127.0.0.1:%d", c.APIPort(nodeIndex))
}

func (c *ClusterConfig) APIPort(nodeIndex int) int {
	return c.Wasp[nodeIndex].APIPort
}

func (c *ClusterConfig) PeeringHosts(nodeIndexes ...[]int) []string {
	nodes := c.AllNodes()
	if len(nodeIndexes) == 1 {
		nodes = nodeIndexes[0]
	}
	return c.waspHosts(nodes, func(i int) string { return c.PeeringHost(i) })
}

func (c *ClusterConfig) PeeringHost(nodeIndex int) string {
	return fmt.Sprintf("127.0.0.1:%d", c.PeeringPort(nodeIndex))
}

func (c *ClusterConfig) PeeringPort(nodeIndex int) int {
	return c.Wasp[nodeIndex].PeeringPort
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
	return c.Wasp[nodeIndex].NanomsgPort
}

func (c *ClusterConfig) DashboardPort(nodeIndex int) int {
	return c.Wasp[nodeIndex].DashboardPort
}

func (c *ClusterConfig) L1APIAddress(nodeIndex int) string {
	return c.L1.APIAddress
}

func (c *ClusterConfig) ProfilingPort(nodeIndex int) int {
	return c.Wasp[nodeIndex].ProfilingPort
}

func (c *ClusterConfig) PrometheusPort(nodeIndex int) int {
	return c.Wasp[nodeIndex].MetricsPort
}
