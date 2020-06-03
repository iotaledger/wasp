package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/util"
)

func CreateOriginData(par *origin.NewOriginParams, dscr string, nodeLocations []string) (*sctransaction.Transaction, *registry.SCMetaData) {
	allOuts := utxodb.GetAddressOutputs(par.OwnerAddress)            // non deterministic
	outs := util.SelectMinimumOutputs(allOuts, balance.ColorIOTA, 1) // must be deterministic!
	if len(outs) == 0 {
		panic("inconsistency: not enough outputs for 1 iota!")
	}
	// select first and the only
	var input valuetransaction.OutputID
	var inputBalances []*balance.Balance

	for oid, v := range outs {
		input = oid
		inputBalances = v
		break
	}

	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		NewOriginParams: *par,
		Input:           input,
		InputBalances:   inputBalances,
		InputColor:      balance.ColorIOTA,
		OwnerSigScheme:  utxodb.GetSigScheme(par.OwnerAddress),
	})
	if err != nil {
		panic(err)
	}
	if nodeLocations == nil {
		return originTx, nil
	}
	scdata := &registry.SCMetaData{
		Address:       par.Address,
		Color:         balance.Color(originTx.ID()),
		OwnerAddress:  par.OwnerAddress,
		ProgramHash:   par.ProgramHash,
		Description:   dscr,
		NodeLocations: nodeLocations,
	}
	return originTx, scdata
}
