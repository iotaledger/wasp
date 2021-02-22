// +build ignore

package chainclient

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
)

type SCStatus struct {
	StateIndex uint32
	Timestamp  time.Time
	StateHash  *hashing.HashValue
	StateTxId  valuetransaction.ID
	Requests   []*coretypes.RequestID

	ProgramHash   *hashing.HashValue
	Description   string
	OwnerAddress  address.Address
	SCAddress     address.Address
	Balance       map[balance.Color]int64
	MinimumReward int64
	FetchedAt     time.Time
}

func (c *Client) FetchSCStatus(addCustomQueries func(query *statequery.Request)) (*SCStatus, *statequery.Results, error) {
	balance, err := c.FetchBalance()
	if err != nil {
		return nil, nil, err
	}

	query := statequery.NewRequest()
	query.AddGeneralData()
	query.AddScalar(vmconst.VarNameOwnerAddress)
	query.AddScalar(vmconst.VarNameProgramData)
	query.AddScalar(vmconst.VarNameDescription)
	query.AddScalar(vmconst.VarNameMinimumReward)
	addCustomQueries(query)

	res, err := c.WaspClient.StateQuery(&c.ChainID, query)
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

		ProgramHash:   &res.Get(vmconst.VarNameProgramData).MustHashValue(),
		Description:   description,
		OwnerAddress:  res.Get(vmconst.VarNameOwnerAddress).MustAddress(),
		MinimumReward: minReward,
		SCAddress:     (address.Address)(c.ChainID),
		Balance:       balance,
		FetchedAt:     time.Now().UTC(),
	}, res, nil
}

func (c *Client) FetchBalance() (map[balance.Color]int64, error) {
	addr := (address.Address)(c.ChainID)
	outs, err := c.Level1Client.GetConfirmedAccountOutputs(&addr)
	if err != nil {
		return nil, err
	}
	ret, _ := txutil.OutputBalancesByColor(outs)
	return ret, nil
}
