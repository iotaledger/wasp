package apilib

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
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
	StateIndex uint32
	Timestamp  time.Time
	StateHash  *hashing.HashValue
	StateTxId  *valuetransaction.ID
	Requests   []*sctransaction.RequestId

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
	query.AddGeneralData()
	query.AddScalar(vmconst.VarNameOwnerAddress)
	query.AddScalar(vmconst.VarNameProgramHash)
	query.AddScalar(vmconst.VarNameDescription)
	query.AddScalar(vmconst.VarNameMinimumReward)
	addCustomQueries(query)

	res, err := QuerySCState(waspHost, query)
	if err != nil {
		return nil, nil, err
	}

	description, _ := res.Queries[vmconst.VarNameDescription].MustString()
	minReward, _ := res.Queries[vmconst.VarNameMinimumReward].MustInt64()

	return &SCStatus{
		StateIndex: res.StateIndex,
		Timestamp:  res.Timestamp.UTC(),
		StateHash:  res.StateHash,
		StateTxId:  res.StateTxId,
		Requests:   res.Requests,

		ProgramHash:   res.Queries[vmconst.VarNameProgramHash].MustHashValue(),
		Description:   description,
		OwnerAddress:  res.Queries[vmconst.VarNameOwnerAddress].MustAddress(),
		MinimumReward: minReward,
		SCAddress:     scAddress,
		Balance:       balance,
		FetchedAt:     time.Now().UTC(),
	}, res.Queries, nil
}

func fetchSCBalance(nodeClient nodeclient.NodeClient, scAddress *address.Address) (map[balance.Color]int64, error) {
	outs, err := nodeClient.GetAccountOutputs(scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}
