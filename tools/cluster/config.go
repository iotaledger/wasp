package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type ClusterConfig struct {
	Nodes     []*WaspNodeConfig `json:"nodes"`
	Goshimmer *struct {
		ApiPort  int  `json:"api_port"`
		Provided bool `json:"provided"`
		cmd      *exec.Cmd
	} `json:"goshimmer"`
}

func (c *ClusterConfig) AllWaspNodes() []int {
	r := make([]int, len(c.Nodes))
	for i := 0; i < len(r); i++ {
		r[i] = i
	}
	return r
}

func (c *ClusterConfig) goshimmerApiHost() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Goshimmer.ApiPort)
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
	x := os.Getenv("GOSHIMMER_PROVIDED")
	if x != "" {
		config.Goshimmer.Provided = true
	}
	return config, nil
}
