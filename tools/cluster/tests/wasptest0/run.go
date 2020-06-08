package wasptest0

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/cluster"
)

func Run(clu *cluster.Cluster) error {
	fmt.Printf("-----------------------------------------------------------------\n")
	fmt.Printf("           Test 0: bootup records and activation of committees  \n")
	fmt.Printf("-----------------------------------------------------------------\n")

	failed := false
	for _, sc := range clu.SmartContractConfig {
		fmt.Printf("[cluster] creating bootup record for smart contract descr: '%s' addr: %s program hash: %s\n",
			sc.Description, sc.Address, util.Short(sc.ProgramHash))

		if err := putScData(&sc, clu); err != nil {
			fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("failed to create bootup records")
	}

	for _, sc := range clu.SmartContractConfig {
		if err := activate(&sc, clu); err != nil {
			fmt.Printf("[cluster] adctivate: addr = %s: %v\n", sc.Address, err)
		}
	}

	return nil
}

func putScData(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	addr, err := address.FromBase58(sc.Address)
	if err != nil {
		return err
	}

	color, err := util.ColorFromString(sc.Color)
	if err != nil {
		return err
	}

	committeeNodes := make([]string, len(sc.CommitteeNodes))
	for i, hi := range sc.CommitteeNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return fmt.Errorf("wrong committee node index")
		}
		committeeNodes[i] = fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].PeeringPort)
	}

	accessNodes := make([]string, len(sc.AccessNodes))
	for i, hi := range sc.AccessNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return fmt.Errorf("wrong access node index")
		}
		accessNodes[i] = fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].PeeringPort)
	}

	var failed bool
	for _, host := range committeeNodes {
		err = waspapi.PutSCData(host, registry.BootupData{
			Address:        addr,
			Color:          color,
			CommitteeNodes: committeeNodes,
			AccessNodes:    accessNodes,
		})
		if err != nil {
			fmt.Printf("[cluster] apilib.PutSCData returned: %v\n", err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("failed to send bootup data to some commitee nodes")
	}
	return nil
}

func activate(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	committeeNodes := make([]string, len(sc.CommitteeNodes))
	for i, hi := range sc.CommitteeNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return fmt.Errorf("wrong committee node index")
		}
		committeeNodes[i] = fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].PeeringPort)
	}
	failed := false
	for _, host := range committeeNodes {
		err := waspapi.ActivateSC(host, sc.Address)
		if err != nil {
			fmt.Printf("[cluster] apilib.ActivateSC returned: %v\n", err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("failed to activate some commitee nodes")
	}
	return nil
}
