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
	"github.com/iotaledger/wasp/packages/variables"
)

const (
	VarNameOwnerAddress  = "$owneraddr$"
	VarNameProgramHash   = "$proghash$"
	VarNameMinimumReward = "$minreward$"
)

type NewOriginParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	ProgramHash          hashing.HashValue
	Variables            variables.Variables
}

type NewOriginTransactionParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	Input                valuetransaction.OutputID
	InputBalances        []*balance.Balance
	InputColor           balance.Color // default is ColorIOTA
	StateHash            hashing.HashValue
}

// content of the origin variable state. It is not linked to the origin transaction yet
func NewOriginBatch(par NewOriginParams) state.Batch {
	stateUpd := state.NewStateUpdate(nil)

	stateUpd.Mutations().AddAll(par.Variables.Mutations())
	stateUpd.Mutations().AddAll(variables.FromMap(map[string][]byte{
		VarNameOwnerAddress: par.OwnerSignatureScheme.Address().Bytes(),
		VarNameProgramHash:  par.ProgramHash.Bytes(),
	}).Mutations())

	ret, err := state.NewBatch([]state.StateUpdate{stateUpd})
	if err != nil {
		panic(err)
	}

	return ret
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
		StateHash:  par.StateHash,
		Timestamp:  0, // <<<< to have deterministic origin tx hash
	})

	ret, err := txb.Finalize()
	if err != nil {
		panic(err)
	}
	ret.Transaction.Sign(par.OwnerSignatureScheme)
	return ret, nil
}
