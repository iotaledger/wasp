package wasptest0

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/apilib"
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
			fmt.Printf("putScData: %v\n", err)
		}
	}
	return nil
}

func putScData(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) error {
	addr, err := address.FromBase58(sc.Address)
	if err != nil {
		return err
	}

	for _, hi := range sc.CommitteeNodes {
		if hi < 0 || hi >= len(clu.Config.Nodes) {
			return fmt.Errorf("wrong node index")
		}
		host := clu.Config.Nodes[hi].BindAddress
		err = apilib.PutSCData(host, registry.BootupData{
			Address: addr,
			//CommitteeNodes: sc.CommitteeNodes,
		})
	}

	//err := apilib.Put
	return nil
}
