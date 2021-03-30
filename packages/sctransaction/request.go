package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
)

type RequestParams struct {
	ChainID    coretypes.ChainID
	Contract   coretypes.Hname
	EntryPoint coretypes.Hname
	Args       requestargs.RequestArgs
}

type NewRequestTransactionParams struct {
	SenderKeyPair  *ed25519.KeyPair
	UnspentOutputs []ledgerstate.Output
	Requests       []RequestParams
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
func NewRequestTransaction(par NewRequestTransactionParams) (*ledgerstate.Transaction, error) {
	txb := utxoutil.NewBuilder(par.UnspentOutputs...)
	for _, req := range par.Requests {
		metadata := NewRequestMetadata().
			WithTarget(req.Contract).
			WithEntryPoint(req.EntryPoint).
			WithArgs(req.Args).
			Bytes()
		err := txb.AddExtendedOutputConsume(req.ChainID.AsAddress(), metadata, map[ledgerstate.Color]uint64{
			ledgerstate.ColorIOTA: 1,
		})
		if err != nil {
			return nil, err
		}
	}

	addr := ledgerstate.NewED25519Address(par.SenderKeyPair.PublicKey)
	if err := txb.AddReminderOutputIfNeeded(addr, nil, true); err != nil {
		return nil, err
	}
	tx, err := txb.BuildWithED25519(par.SenderKeyPair)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
