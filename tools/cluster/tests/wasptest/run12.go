package wasptest

import (
	"fmt"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/tools/cluster"
	"time"
)

func Activate1SC(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	if err := activate(sc, clu); err != nil {
		return fmt.Errorf("activate %s: %v\n", sc.Address, err)
	}

	return nil
}

func ActivateAllSC(clu *cluster.Cluster) error {
	for _, sc := range clu.SmartContractConfig {
		if err := activate(&sc, clu); err != nil {
			return fmt.Errorf("activate %s: %v\n", sc.Address, err)
		}
	}

	return nil
}

func activate(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	allNodesApi := clu.WaspHosts(sc.AllNodes(), (*cluster.WaspNodeConfig).ApiHost)

	funs := make([]func() (interface{}, error), len(allNodesApi))
	for i, host := range allNodesApi {
		h := host
		funs[i] = func() (interface{}, error) {
			err := waspapi.ActivateSC(h, sc.Address)
			return nil, err
		}
	}

	resp, succ := multicall.MultiCall(funs, 500*time.Millisecond)
	if succ {
		return nil
	}
	for i := range resp {
		fmt.Printf("apilib.ActivateSC returned: %v\n", resp[i].Err)
	}
	//for _, host := range allNodesApi {
	//	err := waspapi.ActivateSC(host, sc.Address)
	//	if err != nil {
	//		return fmt.Errorf("apilib.ActivateSC returned: %v\n", err)
	//	}
	//}
	return fmt.Errorf("apilib.ActivateSC failed\n")
}
