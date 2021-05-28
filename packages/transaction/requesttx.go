package transaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
)

type RequestParams struct {
	ChainID    coretypes.ChainID
	Contract   coretypes.Hname
	EntryPoint coretypes.Hname
	Transfer   *ledgerstate.ColoredBalances
	Args       requestargs.RequestArgs
}

type NewRequestTransactionParams struct {
	SenderKeyPair  *ed25519.KeyPair
	UnspentOutputs []ledgerstate.Output
	Requests       []RequestParams
}

var oneIota = map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
func NewRequestTransaction(par NewRequestTransactionParams) (*ledgerstate.Transaction, error) {
	txb := utxoutil.NewBuilder(par.UnspentOutputs...)
	for _, req := range par.Requests {
		metadata := request.NewRequestMetadata().
			WithTarget(req.Contract).
			WithEntryPoint(req.EntryPoint).
			WithArgs(req.Args).
			Bytes()
		transfer := oneIota
		if req.Transfer != nil {
			transfer = req.Transfer.Map()
		}
		err := txb.AddExtendedOutputConsume(req.ChainID.AsAddress(), metadata, transfer)
		if err != nil {
			return nil, err
		}
	}

	addr := ledgerstate.NewED25519Address(par.SenderKeyPair.PublicKey)
	if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
		return nil, err
	}
	tx, err := txb.BuildWithED25519(par.SenderKeyPair)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
