package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/cluster"
)

func PutBootupRecords(clu *cluster.Cluster) error {
	fmt.Printf("------------------------- Test 0: bootup records  \n")

	for _, sc := range clu.SmartContractConfig {
		fmt.Printf("[cluster] creating bootup record for smart contract addr: %s\n", sc.Address)

		if err := putScData(&sc, clu); err != nil {
			fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)

			return fmt.Errorf("failed to create bootup records")
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

	committeePeerNodes := clu.WaspHosts(sc.CommitteeNodes, (*cluster.WaspNodeConfig).PeeringHost)
	accessPeerNodes := clu.WaspHosts(sc.AccessNodes, (*cluster.WaspNodeConfig).PeeringHost)
	allNodesApi := clu.WaspHosts(sc.AllNodes(), (*cluster.WaspNodeConfig).ApiHost)

	var failed bool
	for _, host := range allNodesApi {

		err = waspapi.PutSCData(host, registry.BootupData{
			Address:        addr,
			Color:          color,
			OwnerAddress:   utxodb.GetAddress(sc.OwnerIndexUtxodb),
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
