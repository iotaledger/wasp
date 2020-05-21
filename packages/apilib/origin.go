package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type NewOriginParams struct {
	Address      address.Address
	OwnerAddress address.Address
	Description  string
	ProgramHash  hashing.HashValue
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
	ret := originState.Hash()
	return &ret
}

type NewOriginTransactionParams struct {
	NewOriginParams
	Input          valuetransaction.OutputID
	InputBalances  []*balance.Balance
	InputColor     balance.Color // default is ColorIOTA
	OwnerSigScheme signaturescheme.SignatureScheme
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
		return nil, fmt.Errorf("wrong inout color")
	}
	txb := sctransaction.NewTransactionBuilder()

	txb.AddInputs(par.Input)

	txb.AddBalanceToOutput(par.Address, balance.New(balance.ColorNew, 1))
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
	ret.Transaction.Sign(par.OwnerSigScheme)
	return ret, nil
}

func CreateOriginData(par NewOriginParams, nodeLocations []string) (*sctransaction.Transaction, *registry.SCMetaData) {
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

	originTx, err := NewOriginTransaction(NewOriginTransactionParams{
		NewOriginParams: par,
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
		Description:   par.Description,
		ProgramHash:   par.ProgramHash,
		NodeLocations: nodeLocations,
	}
	return originTx, scdata
}

func NewMetaData(par NewOriginParams, color *balance.Color, nodeLocations []string) *registry.SCMetaData {
	return &registry.SCMetaData{
		Address:       par.Address,
		Color:         *color,
		OwnerAddress:  par.OwnerAddress,
		Description:   par.Description,
		ProgramHash:   par.ProgramHash,
		NodeLocations: nodeLocations,
	}
}
