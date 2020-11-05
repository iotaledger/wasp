package cluster

import (
	"fmt"
	"os/exec"

	"github.com/iotaledger/wasp/client"
)

type WaspNodeConfig struct {
	ApiPort       int `json:"api_port"`
	PeeringPort   int `json:"peering_port"`
	NanomsgPort   int `json:"nanomsg_port"`
	DashboardPort int `json:"dashboard_port"`
	cmd           *exec.Cmd
}

func (w *WaspNodeConfig) ApiHost() string {
	return fmt.Sprintf("127.0.0.1:%d", w.ApiPort)
}

func (w *WaspNodeConfig) PeeringHost() string {
	return fmt.Sprintf("127.0.0.1:%d", w.PeeringPort)
}

func (w *WaspNodeConfig) NanomsgHost() string {
	return fmt.Sprintf("127.0.0.1:%d", w.NanomsgPort)
}

func (w *WaspNodeConfig) IsUp() bool {
	return w.cmd != nil
}

func (w *WaspNodeConfig) Client() *client.WaspClient {
	return client.NewWaspClient(w.ApiHost())
}
