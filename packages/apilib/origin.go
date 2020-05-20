package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
)

type NewOriginParams struct {
	Address      *address.Address
	OwnerAddress *address.Address
	Description  string
	ProgramHash  *hashing.HashValue
}

// content of the origin variable state. It does not linked to the origin transaction yet
func NewOriginBatchUncommitted(par NewOriginParams) state.Batch {
	stateUpd := state.NewStateUpdate(nil)
	vars := stateUpd.Variables()
	vars.Set("$address$", par.Address.String())
	vars.Set("$owneraddr$", par.OwnerAddress.String())
	vars.Set("$descr$", par.Description)
	vars.Set("$proghash$", par.ProgramHash.String())
	ret, err := state.NewBatch([]state.StateUpdate{stateUpd}, 0)
	if err != nil {
		panic(err)
	}
	return ret
}

// does not include color/origin tx hash
func OriginVariableStateHash(par NewOriginParams) *hashing.HashValue {
	batch := NewOriginBatchUncommitted(par)
	originState := state.NewVariableState(nil)
	if err := originState.Apply(batch); err != nil {
		panic(err)
	}
	return originState.Hash()
}

type NewOriginTransactionParams struct {
	NewOriginParams
	Input          valuetransaction.OutputID
	InputBalances  []*balance.Balance
	InputColor     balance.Color // default is ColorIOTA
	OwnerSigScheme signaturescheme.SignatureScheme
}

func NewOriginTransaction(par NewOriginTransactionParams) (*valuetransaction.Transaction, error) {
	reminderBalances := make([]*balance.Balance, 0, len(par.InputBalances))
	hasColor := false
	for _, bal := range par.InputBalances {
		if bal.Color() == par.InputColor {
			if bal.Value() > 1 {
				reminderBalances = append(reminderBalances, balance.New(par.InputColor, bal.Value()-1))
			}
			hasColor = true
		} else {
			reminderBalances = append(reminderBalances, bal)
		}
	}
	if !hasColor {
		return nil, fmt.Errorf("wrong inout color")
	}
	txb := sctransaction.NewTransactionBuilder()

	txb.AddInputs(par.Input)

	txb.AddBalanceToOutput(*par.Address, balance.New(balance.ColorNew, 1))
	// reminder outputs if any
	for _, remb := range reminderBalances {
		txb.AddBalanceToOutput(par.Input.Address(), remb)
	}
	var col balance.Color = balance.ColorNew
	txb.AddStateBlock(&col, 0)

	txb.SetVariableStateHash(OriginVariableStateHash(par.NewOriginParams))

	ret, err := txb.Finalize()
	if err != nil {
		panic(err)
	}

	return ret.Transaction.Sign(par.OwnerSigScheme), nil
}

func NewMetaData(par NewOriginParams, color *balance.Color, nodeLocations []string) *registry.SCMetaData {
	return &registry.SCMetaData{
		Address:       *par.Address,
		Color:         *color,
		OwnerAddress:  *par.OwnerAddress,
		Description:   par.Description,
		ProgramHash:   *par.ProgramHash,
		NodeLocations: nodeLocations,
	}
}
