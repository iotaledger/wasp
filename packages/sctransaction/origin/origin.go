package origin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

type NewOriginTransactionParams struct {
	OriginAddress             address.Address
	OriginatorSignatureScheme signaturescheme.SignatureScheme
	AllInputs                 map[valuetransaction.OutputID][]*balance.Balance
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
	var dummyChainID coretypes.ChainID
	originState := state.NewVirtualState(nil, &dummyChainID)
	if err := originState.ApplyBlock(state.MustNewOriginBlock(nil)); err != nil {
		return nil, err
	}
	originHash := originState.Hash()

	if err := txb.CreateOriginStateSection(originHash, &par.OriginAddress); err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.OriginatorSignatureScheme)
	return tx, nil
}

type NewRootInitRequestTransactionParams struct {
	ChainID              coretypes.ChainID
	ChainColor           balance.Color
	ChainAddress         address.Address
	Description          string
	OwnerSignatureScheme signaturescheme.SignatureScheme
	AllInputs            map[valuetransaction.OutputID][]*balance.Balance
}

// NewRootInitRequestTransaction is a first request to be sent to the uninitialized
// chain. At this moment it only is able to process this specific request
// the request contains minimum data needed to bootstrap the chain
// Transaction must be signed by the same address which created origin transaction
func NewRootInitRequestTransaction(par NewRootInitRequestTransactionParams) (*sctransaction.Transaction, error) {
	txb, err := txbuilder.NewFromOutputBalances(par.AllInputs)
	if err != nil {
		return nil, err
	}
	rootContractID := coretypes.NewContractID(par.ChainID, root.Interface.Hname())
	initRequest := sctransaction.NewRequestSection(0, rootContractID, coretypes.EntryPointInit)
	args := dict.New()
	args.Set(root.ParamChainID, codec.EncodeChainID(par.ChainID))
	args.Set(root.ParamChainColor, codec.EncodeColor(par.ChainColor))
	args.Set(root.ParamChainAddress, codec.EncodeAddress(par.ChainAddress))
	args.Set(root.ParamDescription, codec.EncodeString(par.Description))
	initRequest.WithArgs(args)

	if err := txb.AddRequestSection(initRequest); err != nil {
		return nil, err
	}

	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.OwnerSignatureScheme)
	return tx, nil
}
