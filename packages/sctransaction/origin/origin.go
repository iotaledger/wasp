package origin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

type NewOriginTransactionParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	AllInputs            map[valuetransaction.OutputID][]*balance.Balance
	ProgramHash          hashing.HashValue
	InputColor           balance.Color // default is ColorIOTA
}

func NewOriginTransaction(par NewOriginTransactionParams) (*sctransaction.Transaction, error) {
	txb, err := txbuilder.NewFromOutputBalances(par.AllInputs)

	originState := state.NewVirtualState(nil)
	if err := originState.ApplyBatch(state.MustNewOriginBatch(nil)); err != nil {
		return nil, err
	}
	stateHash := originState.Hash()
	if err := txb.AddOriginStateBlock(&stateHash, &par.Address); err != nil {
		return nil, err
	}
	initRequest := sctransaction.NewRequestBlock(par.Address, vmconst.RequestCodeInit)
	initRequest.Args().SetAddress(vmconst.VarNameOwnerAddress, par.OwnerSignatureScheme.Address())
	if par.ProgramHash != *hashing.NilHash {
		initRequest.Args().SetHashValue(vmconst.VarNameProgramHash, par.ProgramHash)
	}
	if err := txb.AddRequestBlock(initRequest); err != nil {
		return nil, err
	}

	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.OwnerSignatureScheme)
	return tx, nil
}
