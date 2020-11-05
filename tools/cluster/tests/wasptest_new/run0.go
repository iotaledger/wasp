package wasptest

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/cluster"
)

// Puts chain records into the nodes. Also requests funds from the nodeClient for owners.
func PutChainRecord(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) (*balance.Color, error) {
	requested := make(map[address.Address]bool)

	fmt.Printf("[cluster] creating chain record for smart contract addr: %s\n", sc.Address)

	ownerAddr := sc.OwnerAddress()
	_, ok := requested[*ownerAddr]
	if !ok {
		err := clu.NodeClient.RequestFunds(ownerAddr)
		if err != nil {
			fmt.Printf("[cluster] Could not request funds: %v\n", err)
			return nil, fmt.Errorf("Could not request funds: %v", err)
		}
		requested[*ownerAddr] = true
	}

	color, err := putScData(clu, sc)
	if err != nil {
		fmt.Printf("[cluster] putScdata: addr = %s: %v\n", sc.Address, err)
		return nil, fmt.Errorf("failed to create chain records: %v", err)
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

	err = clu.MultiClient().PutChainRecord(&registry.ChainRecord{
		ChainID:        (coretypes.ChainID)(addr),
		Color:          color,
		OwnerAddress:   *sc.OwnerAddress(),
		CommitteeNodes: committeePeerNodes,
		AccessNodes:    accessPeerNodes,
	})

	if err != nil {
		fmt.Printf("[cluster] PutChainRecord returned: %v\n", err)
		return nil, fmt.Errorf("failed to send chain record to some commitee nodes")
	}
	return &color, nil
}
