package origin

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
)

type NewOriginParams struct {
	Address      address.Address
	OwnerAddress address.Address
	ProgramHash  hashing.HashValue
}

type NewOriginTransactionParams struct {
	NewOriginParams
	Input          valuetransaction.OutputID
	InputBalances  []*balance.Balance
	InputColor     balance.Color // default is ColorIOTA
	OwnerSigScheme signaturescheme.SignatureScheme
}

// content of the origin variable state. It does not linked to the origin transaction yet
func NewOriginBatch(par NewOriginParams) state.Batch {
	stateUpd := state.NewStateUpdate(nil)
	vars := stateUpd.Variables()
	vars.Set("$address$", par.Address.String())
	vars.Set("$owneraddr$", par.OwnerAddress.String())
	vars.Set("$proghash$", par.ProgramHash.String())
	ret, err := state.NewBatch([]state.StateUpdate{stateUpd})
	if err != nil {
		panic(err)
	}
	return ret
}

// does not include color/origin tx hash
func originVariableStateHash(par NewOriginParams) *hashing.HashValue {
	batch := NewOriginBatch(par)
	originState := state.NewVariableState(nil)
	if err := originState.ApplyBatch(batch); err != nil {
		panic(err)
	}
	ret := originState.Hash()
	return &ret
}

func NewOriginTransaction(par NewOriginTransactionParams) (*sctransaction.Transaction, error) {
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
		return nil, fmt.Errorf("wrong input color")
	}
	txb := sctransaction.NewTransactionBuilder()

	if err := txb.AddInputs(par.Input); err != nil {
		return nil, err
	}

	txb.AddBalanceToOutput(par.Address, balance.New(balance.ColorNew, 1))
	// reminder outputs if any
	for _, remb := range reminderBalances {
		txb.AddBalanceToOutput(par.Input.Address(), remb)
	}
	txb.AddStateBlock(sctransaction.NewStateBlockParams{
		Color:      balance.ColorNew,
		StateIndex: 0,
		StateHash:  *originVariableStateHash(par.NewOriginParams),
		Timestamp:  0, // <<<< to have deterministic origin tx hash
	})

	ret, err := txb.Finalize()
	if err != nil {
		panic(err)
	}
	ret.Transaction.Sign(par.OwnerSigScheme)
	return ret, nil
}
