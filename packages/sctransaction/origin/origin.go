package origin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"strings"
)

type NewOriginTransactionParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	AllInputs            map[valuetransaction.OutputID][]*balance.Balance
	ProgramHash          hashing.HashValue
	Description          string
	InputColor           balance.Color // default is ColorIOTA
}

func NewOriginTransaction(par NewOriginTransactionParams) (*sctransaction.Transaction, error) {
	txb, err := txbuilder.NewFromOutputBalances(par.AllInputs)

	originState := state.NewVirtualState(nil, &par.Address)
	if err := originState.ApplyBatch(state.MustNewOriginBatch(nil)); err != nil {
		return nil, err
	}
	stateHash := originState.Hash()
	if err := txb.CreateOriginStateBlock(&stateHash, &par.Address); err != nil {
		return nil, err
	}

	initRequest := sctransaction.NewRequestBlock(par.Address, vmconst.RequestCodeInit)
	args := kv.NewMap()
	ownerAddress := par.OwnerSignatureScheme.Address()
	args.Codec().SetAddress(vmconst.VarNameOwnerAddress, &ownerAddress)
	if par.ProgramHash != *hashing.NilHash {
		args.Codec().SetHashValue(vmconst.VarNameProgramHash, &par.ProgramHash)
	}
	s := strings.TrimSpace(par.Description)
	if len(s) > 0 {
		args.Codec().SetString(vmconst.VarNameDescription, s)
	}

	initRequest.SetArgs(args)

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
