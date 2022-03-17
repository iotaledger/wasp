package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

type L1Config struct {
	Hostname   string
	APIPort    int
	FaucetPort int
	FaucetKey  *cryptolib.KeyPair
	NetworkID  string
}

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

type ClusterConfig struct {
	Wasp WaspConfig
	L1   L1Config
}

func DefaultConfig() *ClusterConfig {
	return &ClusterConfig{
		Wasp: WaspConfig{
			NumNodes:           4,
			FirstAPIPort:       9090,
			FirstPeeringPort:   4000,
			FirstNanomsgPort:   5550,
			FirstDashboardPort: 7000,
			FirstProfilingPort: 6060,
			FirstMetricsPort:   2112,
		},
		L1: L1Config{
			Hostname:   "127.0.0.1",
			APIPort:    8080,
			FaucetPort: 8091,
		},
	}
}

func ConfigExists(dataPath string) (bool, error) {
	return fileExists(configPath(dataPath))
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

func (c *ClusterConfig) goshimmerAPIHost() string {
	return fmt.Sprintf("%s:%d", c.L1.Hostname, c.L1.APIPort)
}

func (c *ClusterConfig) waspHosts(nodeIndexes []int, getHost func(i int) string) []string {
	hosts := make([]string, 0)
	for _, i := range nodeIndexes {
		if i < 0 || i > c.Wasp.NumNodes-1 {
			panic(fmt.Sprintf("Node index out of bounds in smart contract configuration: %d/%d", i, c.Wasp.NumNodes))
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
	return c.Wasp.FirstAPIPort + nodeIndex
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

func (c *ClusterConfig) L1Host(nodeIndex int) string {
	return c.L1.Hostname
}

func (c *ClusterConfig) L1Port(nodeIndex int) int {
	return c.L1.APIPort
}

func (c *ClusterConfig) ProfilingPort(nodeIndex int) int {
	return c.Wasp.FirstProfilingPort + nodeIndex
}

func (c *ClusterConfig) PrometheusPort(nodeIndex int) int {
	return c.Wasp.FirstMetricsPort + nodeIndex
}

func (c *ClusterConfig) WaspConfigTemplateParams(i int, ownerAddress iotago.Address) *templates.WaspConfigParams {
	return &templates.WaspConfigParams{
		APIPort:                      c.APIPort(i),
		DashboardPort:                c.DashboardPort(i),
		PeeringPort:                  c.PeeringPort(i),
		NanomsgPort:                  c.NanomsgPort(i),
		ProfilingPort:                c.ProfilingPort(i),
		L1Host:                       c.L1Host(i),
		L1Port:                       c.L1Port(i),
		MetricsPort:                  c.PrometheusPort(i),
		OwnerAddress:                 ownerAddress.Bech32(iscp.NetworkPrefix),
		OffledgerBroadcastUpToNPeers: 10,
	}
}
