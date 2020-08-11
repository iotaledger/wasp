package wasptest

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/cluster"
)

// Puts bootup records into the nodes. Also requests funds from the nodeClient for owners.
func PutBootupRecord(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) (*balance.Color, error) {
	requested := make(map[address.Address]bool)

	fmt.Printf("[cluster] creating bootup record for smart contract addr: %s\n", sc.Address)

	ownerAddr := sc.OwnerAddress()
	_, ok := requested[ownerAddr]
	if !ok {
		err := clu.NodeClient.RequestFunds(&ownerAddr)
		if err != nil {
			fmt.Printf("[cluster] Could not request funds: %v\n", err)
			return nil, fmt.Errorf("Could not request funds: %v", err)
		}
		requested[ownerAddr] = true
	}

	color, err := putScData(clu, sc)
	if err != nil {
		fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)
		return nil, fmt.Errorf("failed to create bootup records: %v", err)
	}
	return color, nil
}

func putScData(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) (*balance.Color, error) {
	addr, err := address.FromBase58(sc.Address)
	if err != nil {
		return nil, err
	}

	origTx, err := sc.CreateOrigin(clu.NodeClient)
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
