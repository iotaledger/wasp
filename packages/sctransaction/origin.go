package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// NewChainOriginTransactionParams parameters of the origin transaction
type NewChainOriginTransactionParams struct {
	// ed25519 key pair used to spend inputs
	KeyPair *ed25519.KeyPair
	// state controller's address of the new chain.
	StateAddress ledgerstate.Address
	// all inputs to take tokens from. Compressed in the output
	AllInputs []ledgerstate.Output
	// Initial balance of the chain's state. Must be at least dust level
	Balance map[ledgerstate.Color]uint64
}

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(par NewChainOriginTransactionParams) (*ledgerstate.Transaction, coretypes.ChainID, error) {
	walletAddr := ledgerstate.NewED25519Address(par.KeyPair.PublicKey)
	txb := utxoutil.NewBuilder(par.AllInputs...)
	if err := txb.AddNewChainMint(par.Balance, par.StateAddress, hashing.NilHash[:]); err != nil {
		return nil, coretypes.ChainID{}, err
	}
	// adding reminder in compressing mode, i.e. all provided inputs will be consumed
	if err := txb.AddReminderOutputIfNeeded(walletAddr, nil, true); err != nil {
		return nil, coretypes.ChainID{}, err
	}
	tx, err := txb.BuildWithED25519(par.KeyPair)
	if err != nil {
		return nil, coretypes.ChainID{}, err
	}
	// determine aliasAddress of the newly minted chain
	chained, err := utxoutil.GetSingleChainedOutput(tx.Essence())
	if err != nil {
		return nil, coretypes.ChainID{}, err
	}
	chainID, err := coretypes.NewChainIDFromAddress(chained.Address())
	if err != nil {
		return nil, coretypes.ChainID{}, err
	}
	return tx, chainID, nil
}

type NewRootInitRequestTransactionParams struct {
	ChainID     coretypes.ChainID
	Description string
	KeyPair     *ed25519.KeyPair
	AllInputs   []ledgerstate.Output
}

// NewRootInitRequestTransaction is a first request to be sent to the uninitialized
// chain. At this moment it only is able to process this specific request
// the request contains minimum data needed to bootstrap the chain
// TransactionEssence must be signed by the same address which created origin transaction
func NewRootInitRequestTransaction(par NewRootInitRequestTransactionParams) (*ledgerstate.Transaction, error) {
	txb := utxoutil.NewBuilder(par.AllInputs...)

	args := requestargs.New(nil)

	const paramChainID = "$$chainid$$"
	const paramDescription = "$$description$$"

	args.AddEncodeSimple(paramChainID, codec.EncodeChainID(par.ChainID))
	args.AddEncodeSimple(paramDescription, codec.EncodeString(par.Description))

	data := RequestMetadata{
		TargetContractHname: coretypes.Hn("root"),
		EntryPoint:          coretypes.EntryPointInit,
		Args:                args,
	}
	err := txb.AddExtendedOutputSimple(par.ChainID.AsAddress(), data.Bytes(), map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 1,
	})
	if err != nil {
		return nil, err
	}
	addr := ledgerstate.NewED25519Address(par.KeyPair.PublicKey)
	if err := txb.AddReminderOutputIfNeeded(addr, nil, true); err != nil {
		return nil, err
	}
	tx, err := txb.BuildWithED25519(par.KeyPair)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
