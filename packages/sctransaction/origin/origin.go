package origin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

type NewOriginTransactionParams struct {
	OriginAddress        address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	AllInputs            map[valuetransaction.OutputID][]*balance.Balance
}

// NewOriginTransaction
// - creates new origin transaction for the chain
// - origin tx approves empty state
// - origin tx mints chain token to the origin address. The address must be owned by the initial committee
func NewOriginTransaction(par NewOriginTransactionParams) (*sctransaction.Transaction, error) {
	txb, err := txbuilder.NewFromOutputBalances(par.AllInputs)
	if err != nil {
		return nil, err
	}

	// calculate origin state hash
	// - take empty state
	// - apply to it an empty batch
	// - take the hash. Note: hash of the state do not depend on the address
	originState := state.NewVirtualState(nil, &par.OriginAddress)
	if err := originState.ApplyBatch(state.MustNewOriginBlock(nil)); err != nil {
		return nil, err
	}
	originHash := originState.Hash()

	if err := txb.CreateOriginStateBlock(originHash, &par.OriginAddress); err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.OwnerSignatureScheme)
	return tx, nil
}

type NewBootupRequestTransactionParams struct {
	ChainID              coretypes.ChainID
	OwnerSignatureScheme signaturescheme.SignatureScheme
	AllInputs            map[valuetransaction.OutputID][]*balance.Balance
	CoreContractBinary   []byte
	VMType               string
}

// NewBootupRequestTransaction is a first request to be sent to the uninitialized
// chain. At this moment it only is able to process this specific request
// the request contains minimum data needed to bootstrap the chain
// Transaction must be signed by the same address which created origin transaction
func NewBootupRequestTransaction(par NewBootupRequestTransactionParams) (*sctransaction.Transaction, error) {
	txb, err := txbuilder.NewFromOutputBalances(par.AllInputs)
	if err != nil {
		return nil, err
	}
	bootupContractID := coretypes.NewContractID(par.ChainID, 0) // 0 is factory builtin contract
	bootupRequest := sctransaction.NewRequestBlock(bootupContractID, 0)
	args := dict.NewDict()
	args.Codec().SetChainID(vmconst.VarNameChainID, &par.ChainID)
	args.Codec().Set(vmconst.VarNameProgramData, par.CoreContractBinary)
	bootupRequest.SetArgs(args)

	if err := txb.AddRequestBlock(bootupRequest); err != nil {
		return nil, err
	}

	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.OwnerSignatureScheme)
	return tx, nil
}
