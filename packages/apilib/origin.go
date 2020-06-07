package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// CreateOriginData creates origin transaction and origin batch, It asks for inputs from goshimmer node
// origin transaction approves origin state. Origin batch is linked to the origin transaction
func CreateOriginData(nodeurl string, par origin.NewOriginParams) (*sctransaction.Transaction, state.Batch, error) {
	ownerAddress := par.OwnerSignatureScheme.Address()
	// get outputs from goshimmer
	allOuts, err := nodeapi.GetAccountOutputs(nodeurl, &ownerAddress)
	if err != nil {
		return nil, nil, err
	}

	outs := util.SelectOutputsForAmount(allOuts, balance.ColorIOTA, 1) // must be deterministic!
	if len(outs) == 0 {
		panic("inconsistency: not enough outputs for 1 iota?!")
	}
	// select first and the only
	var input valuetransaction.OutputID
	var inputBalances []*balance.Balance

	for oid, v := range outs {
		input = oid
		inputBalances = v
		break
	}
	originBatch := origin.NewOriginBatch(par)

	// calculate state hash
	originState := state.NewVariableState(nil)
	if err := originState.ApplyBatch(originBatch); err != nil {
		return nil, nil, err
	}

	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              par.Address,
		OwnerSignatureScheme: par.OwnerSignatureScheme,
		Input:                input,
		InputBalances:        inputBalances,
		InputColor:           balance.ColorIOTA,
		StateHash:            originState.Hash(),
	})
	if err != nil {
		return nil, nil, err
	}
	originBatch.WithStateTransaction(originTx.ID())

	return originTx, originBatch, nil
}
