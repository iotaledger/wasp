package origin

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type NewOriginTransactionParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	ProgramHash          hashing.HashValue
	Input                valuetransaction.OutputID
	InputBalances        []*balance.Balance
	InputColor           balance.Color // default is ColorIOTA
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

	// adding 2 iotas: one for SC token, another for the init request
	txb.AddBalanceToOutput(par.Address, balance.New(balance.ColorNew, 2))
	// reminder outputs if any
	for _, remb := range reminderBalances {
		txb.AddBalanceToOutput(par.Input.Address(), remb)
	}

	originState := state.NewVirtualState(nil)
	if err := originState.ApplyBatch(state.MustNewOriginBatch(nil)); err != nil {
		return nil, err
	}

	txb.AddStateBlock(sctransaction.NewStateBlockParams{
		Color:      balance.ColorNew,
		StateIndex: 0,
		StateHash:  originState.Hash(), // hash of the origin state does not depend on color
		Timestamp:  0,                  // <<<< to have deterministic origin tx hash
	})

	// add init request
	initRequest := sctransaction.NewRequestBlock(par.Address, vmconst.RequestCodeInit)
	initRequest.Args().SetAddress(vmconst.VarNameOwnerAddress, par.OwnerSignatureScheme.Address())
	if par.ProgramHash != *hashing.NilHash {
		initRequest.Args().SetHashValue(vmconst.VarNameProgramHash, par.ProgramHash)
	}
	txb.AddRequestBlock(initRequest)

	ret, err := txb.Finalize()
	if err != nil {
		panic(err)
	}
	ret.Transaction.Sign(par.OwnerSignatureScheme)
	return ret, nil
}
