package scclient

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/client/statequery"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
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

func (sc *SCClient) FetchSCStatus(addCustomQueries func(query *statequery.Request)) (*SCStatus, *statequery.Results, error) {
	balance, err := sc.FetchBalance()
	if err != nil {
		return nil, nil, err
	}

	query := statequery.NewRequest()
	query.AddGeneralData()
	query.AddScalar(vmconst.VarNameOwnerAddress)
	query.AddScalar(vmconst.VarNameProgramHash)
	query.AddScalar(vmconst.VarNameDescription)
	query.AddScalar(vmconst.VarNameMinimumReward)
	addCustomQueries(query)

	res, err := sc.WaspClient.StateQuery(sc.Address, query)
	if err != nil {
		return nil, nil, err
	}

	description, _ := res.Get(vmconst.VarNameDescription).MustString()
	minReward, _ := res.Get(vmconst.VarNameMinimumReward).MustInt64()

	return &SCStatus{
		StateIndex: res.StateIndex,
		Timestamp:  res.Timestamp.UTC(),
		StateHash:  res.StateHash,
		StateTxId:  res.StateTxId.ID(),
		Requests:   res.Requests,

		ProgramHash:   res.Get(vmconst.VarNameProgramHash).MustHashValue(),
		Description:   description,
		OwnerAddress:  res.Get(vmconst.VarNameOwnerAddress).MustAddress(),
		MinimumReward: minReward,
		SCAddress:     sc.Address,
		Balance:       balance,
		FetchedAt:     time.Now().UTC(),
	}, res, nil
}

func (sc *SCClient) FetchBalance() (map[balance.Color]int64, error) {
	outs, err := sc.NodeClient.GetConfirmedAccountOutputs(sc.Address)
	if err != nil {
		return nil, err
	}
	ret, _ := txutil.OutputBalancesByColor(outs)
	return ret, nil
}
