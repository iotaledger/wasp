package wasptest

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/cluster"
)

func PutBootupRecords(clu *cluster.Cluster) (map[string]*balance.Color, error) {
	fmt.Printf("------------------------- Test 0: bootup records  \n")

	requested := make(map[address.Address]bool)

	colors := make(map[string]*balance.Color)
	for _, sc := range clu.SmartContractConfig {
		fmt.Printf("[cluster] creating bootup record for smart contract addr: %s\n", sc.Address)

		ownerAddr := sc.OwnerAddress()
		_, ok := requested[ownerAddr]
		if !ok {
			err := testutil.RequestFunds(clu.Config.GoshimmerApiHost(), ownerAddr)
			if err != nil {
				fmt.Printf("[cluster] Could not request funds: %v\n", err)
				return nil, fmt.Errorf("Could not request funds")
			}
			requested[ownerAddr] = true
		}

		color, err := putScData(&sc, clu)
		if err != nil {
			fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)
			return nil, fmt.Errorf("failed to create bootup records")
		}
		colors[sc.Address] = color
	}
	return colors, nil
}

func putScData(sc *cluster.SmartContractFinalConfig, clu *cluster.Cluster) (*balance.Color, error) {
	addr, err := address.FromBase58(sc.Address)
	if err != nil {
		return nil, err
	}

	origTx, err := sc.CreateOrigin(clu.Config.GoshimmerApiHost())
	if err != nil {
		return nil, err
	}

	color := balance.Color(origTx.ID())

	committeePeerNodes := clu.WaspHosts(sc.CommitteeNodes, (*cluster.WaspNodeConfig).PeeringHost)
	accessPeerNodes := clu.WaspHosts(sc.AccessNodes, (*cluster.WaspNodeConfig).PeeringHost)
	allNodesApi := clu.WaspHosts(sc.AllNodes(), (*cluster.WaspNodeConfig).ApiHost)

	var failed bool
	for _, host := range allNodesApi {

		err = waspapi.PutSCData(host, registry.BootupData{
			Address:        addr,
			Color:          color,
			OwnerAddress:   sc.OwnerAddress(),
			CommitteeNodes: committeePeerNodes,
			AccessNodes:    accessPeerNodes,
		})
		if err != nil {
			fmt.Printf("[cluster] apilib.PutSCData returned for host %s: %v\n", host, err)
			failed = true
		}
	}
	if failed {
		return nil, fmt.Errorf("failed to send bootup data to some commitee nodes")
	}
	return &color, nil
}
