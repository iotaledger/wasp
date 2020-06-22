package origin

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
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
	// need 2 tokens: one for SC token, another for init request
	tokens := int64(2)
	outs := util.SelectOutputsForAmount(par.AllInputs, balance.ColorIOTA, tokens) // must be deterministic!
	if len(outs) == 0 {
		return nil, errors.New("inconsistency: not enough outputs for 2 tokens")
	}

	fmt.Printf("++++++++++++++++ \n%+v\n", outs)

	byCol, total := util.BalancesOfInputAddressByColor(par.OwnerSignatureScheme.Address(), outs)
	fmt.Printf("++++++++++++++++ \n%+v\n", byCol)

	txb := sctransaction.NewTransactionBuilder()

	txb.AddBalanceToOutput(par.Address, balance.New(balance.ColorNew, tokens))

	tokensRem := tokens
	reminderBalances := make([]*balance.Balance, 0, len(outs))
	reminderTotal := int64(0)
	for col, b := range byCol {
		if col == par.InputColor {
			if b > tokensRem {
				reminderBalances = append(reminderBalances, balance.New(par.InputColor, b-tokens))
				reminderTotal += b - tokens
				tokensRem = 0
			} else {
				tokensRem -= b
			}
		} else {
			reminderBalances = append(reminderBalances, balance.New(col, b))
			reminderTotal += b
		}
	}
	if reminderTotal+tokens != total {
		panic("reminderTotal + tokens != total")
	}

	// reminder outputs if any

	for _, remb := range reminderBalances {
		txb.AddBalanceToOutput(par.Address, remb)
	}
	oids := make([]valuetransaction.OutputID, 0, len(outs))
	for oid := range outs {
		oids = append(oids, oid)
	}
	if err := txb.AddInputs(oids...); err != nil {
		return nil, err
	}

	// adding 2 tokens: one for SC token, another for the init request

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
