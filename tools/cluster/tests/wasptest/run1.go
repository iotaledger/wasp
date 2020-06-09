package wasptest

import (
	"fmt"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/cluster"
)

func Run1(clu *cluster.Cluster) error {
	fmt.Printf("-----------------------------     Test 1: activation of 1 smart contract  \n")

	sc := &clu.SmartContractConfig[0]
	if err := activate(sc, clu); err != nil {
		return fmt.Errorf("activate %s: %v\n", sc.Address, err)
	}

	return nil
}

func Run2(clu *cluster.Cluster) error {
	fmt.Printf("------------------------------   Test 2: activation of 3 smart contract  \n")

	for _, sc := range clu.SmartContractConfig {
		if err := activate(&sc, clu); err != nil {
			return fmt.Errorf("activate %s: %v\n", sc.Address, err)
		}
	}

	return nil
}

func activate(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	_, _, allNodesApi, err := getUrls(sc, clu)
	if err != nil {
		return err
	}
	for _, host := range allNodesApi {
		err := waspapi.ActivateSC(host, sc.Address)
		if err != nil {
			return fmt.Errorf("apilib.ActivateSC returned: %v\n", err)
		}
	}
	return nil
}
