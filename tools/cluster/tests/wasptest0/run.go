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

	for _, sc := range clu.SmartContractConfig {
		fmt.Printf("[cluster] creating bootup record for smart contract descr: '%s' addr: %s program hash: %s\n",
			sc.Description, sc.Address, util.Short(sc.ProgramHash))

		if err := putScData(&sc, clu); err != nil {
			fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)

			return fmt.Errorf("failed to create bootup records")
		}
	}
	//
	for _, sc := range clu.SmartContractConfig {
		if err := activate(&sc, clu); err != nil {
			fmt.Printf("[cluster] activate %s: %v\n", sc.Address, err)

			return fmt.Errorf("failed to activate smart contracts")
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

	committeePeerNodes, accessPeerNodes, allNodesApi, err := getUrls(sc, clu)
	if err != nil {
		return err
	}

	var failed bool
	for _, host := range allNodesApi {
		err = waspapi.PutSCData(host, registry.BootupData{
			Address:        addr,
			Color:          color,
			CommitteeNodes: committeePeerNodes,
			AccessNodes:    accessPeerNodes,
		})
		if err != nil {
			fmt.Printf("[cluster] apilib.PutSCData returned for host %s: %v\n", host, err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("failed to send bootup data to some commitee nodes")
	}
	return nil
}

func activate(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	_, _, allNodesApi, err := getUrls(sc, clu)
	if err != nil {
		return err
	}
	failed := false
	for _, host := range allNodesApi {
		err := waspapi.ActivateSC(host, sc.Address)
		if err != nil {
			fmt.Printf("[cluster] apilib.ActivateSC returned: %v\n", err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("failed to activate smart contracts on some nodes")
	}
	return nil
}

func getUrls(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) ([]string, []string, []string, error) {
	allNodesApi := make([]string, 0, len(sc.CommitteeNodes)+len(sc.AccessNodes))
	committeePeerNodes := make([]string, len(sc.CommitteeNodes))
	for i, hi := range sc.CommitteeNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return nil, nil, nil, fmt.Errorf("wrong committee node index")
		}
		committeePeerNodes[i] = fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].PeeringPort)
		allNodesApi = append(allNodesApi, fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].ApiPort))
	}

	accessPeerNodes := make([]string, len(sc.AccessNodes))
	for i, hi := range sc.AccessNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return nil, nil, nil, fmt.Errorf("wrong access node index")
		}
		accessPeerNodes[i] = fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].PeeringPort)
		allNodesApi = append(allNodesApi, fmt.Sprintf("%s:%d", clu.Config.Nodes[hi].NetAddress, clu.Config.Nodes[hi].ApiPort))
	}
	return committeePeerNodes, accessPeerNodes, allNodesApi, nil
}
