package wasptest

import (
	"fmt"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/cluster"
)

func Activate1SC(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	fmt.Printf("-----------------------------     Test 1: activation of 1 smart contract  \n")

	if err := activate(sc, clu); err != nil {
		return fmt.Errorf("activate %s: %v\n", sc.Address, err)
	}

	return nil
}

func Activate3SC(clu *cluster.Cluster) error {
	fmt.Printf("------------------------------   Test 2: activation of 3 smart contract  \n")

	for _, sc := range clu.SmartContractConfig {
		if err := activate(&sc, clu); err != nil {
			return fmt.Errorf("activate %s: %v\n", sc.Address, err)
		}
	}

	return nil
}

func activate(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	allNodesApi := clu.WaspHosts(sc.AllNodes(), (*cluster.WaspNodeConfig).ApiHost)
	for _, host := range allNodesApi {
		err := waspapi.ActivateSC(host, sc.Address)
		if err != nil {
			return fmt.Errorf("apilib.ActivateSC returned: %v\n", err)
		}
	}
	return nil
}
