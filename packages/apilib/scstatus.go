package apilib

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type SCStatus struct {
	ProgramHash   *hashing.HashValue
	Description   string
	OwnerAddress  *address.Address
	SCAddress     *address.Address
	Balance       map[balance.Color]int64
	MinimumReward int64
	FetchedAt     time.Time
}

func FetchSCStatus(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, addCustomQueries func(query *stateapi.QueryRequest)) (*SCStatus, map[kv.Key]*stateapi.QueryResult, error) {
	balance, err := fetchSCBalance(nodeClient, scAddress)
	if err != nil {
		return nil, nil, err
	}

	query := stateapi.NewQueryRequest(scAddress)
	query.AddScalar(vmconst.VarNameOwnerAddress)
	query.AddScalar(vmconst.VarNameProgramHash)
	query.AddScalar(vmconst.VarNameDescription)
	query.AddScalar(vmconst.VarNameMinimumReward)
	addCustomQueries(query)

	results, err := QuerySCState(waspHost, query)
	if err != nil {
		return nil, nil, err
	}

	description, _ := results[vmconst.VarNameDescription].MustString()
	minReward, _ := results[vmconst.VarNameMinimumReward].MustInt64()

	return &SCStatus{
		ProgramHash:   results[vmconst.VarNameProgramHash].MustHashValue(),
		Description:   description,
		OwnerAddress:  results[vmconst.VarNameOwnerAddress].MustAddress(),
		MinimumReward: minReward,
		SCAddress:     scAddress,
		Balance:       balance,
		FetchedAt:     time.Now().UTC(),
	}, results, nil
}

func fetchSCBalance(nodeClient nodeclient.NodeClient, scAddress *address.Address) (map[balance.Color]int64, error) {
	outs, err := nodeClient.GetAccountOutputs(scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}
