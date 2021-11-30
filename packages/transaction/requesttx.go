package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type RequestParams struct {
	ChainID    *iscp.ChainID
	Contract   iscp.Hname
	EntryPoint iscp.Hname
	Transfer   *iscp.Assets
	Args       dict.Dict
}

type NewRequestTransactionParams struct {
	SenderKeyPair  *ed25519.KeyPair
	UnspentOutputs []iotago.Output
	Requests       []RequestParams
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
// To avoid empty transfer it defaults to 1 iota
func NewRequestTransaction(par NewRequestTransactionParams) (*iotago.Transaction, error) {
	panic("TODO use new txbuilder")
	// txb := utxoutil.NewBuilder(par.UnspentOutputs...)
	// for _, req := range par.Requests {
	// 	metadata := request.NewMetadata().
	// 		WithTarget(req.Contract).
	// 		WithEntryPoint(req.EntryPoint).
	// 		WithArgs(req.Args).
	// 		Bytes()
	// 	if req.Transfer.IsEmpty() {
	// 		req.Transfer = iscp.NewAssets(1, nil)
	// 	}
	// err := txb.AddExtendedOutputConsume(req.ChainID.AsAddress(), metadata, colored.ToL1Map(transfer))
	// if err != nil {
	// 	return nil, err
	// }
	// }

	// addr := ledgerstate.NewED25519Address(par.SenderKeyPair.PublicKey)
	// if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
	// 	return nil, err
	// }
	// tx, err := txb.BuildWithED25519(par.SenderKeyPair)
	// if err != nil {
	// 	return nil, err
	// }
	// return tx, nil
}
