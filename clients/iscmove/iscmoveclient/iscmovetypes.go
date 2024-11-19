package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type referent[T any] struct {
	ID    iotago.ObjectID
	Value *T `bcs:"optional"`
}

// moveAnchor is the BCS equivalent for the move type moveAnchor
type moveAnchor struct {
	ID            iotago.ObjectID
	Assets        referent[iscmove.AssetsBag]
	StateMetadata []byte
	StateIndex    uint32
	GasObjAddr    iotago.ObjectID
	TxFeePerReq   uint64
}

func (ma *moveAnchor) ToAnchor() *iscmove.Anchor {
	return &iscmove.Anchor{
		ID:            ma.ID,
		Assets:        *ma.Assets.Value,
		StateMetadata: ma.StateMetadata,
		StateIndex:    ma.StateIndex,
		GasObjAddr:    ma.GasObjAddr,
		TxFeePerReq:   ma.TxFeePerReq,
	}
}

type moveRequest struct {
	ID     iotago.ObjectID
	Sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
	AssetsBag referent[iscmove.AssetsBagWithBalances] // Need to decide if we want to use this Referent wrapper as well. Could probably be of *AssetsBag with `bcs:"optional`
	Message   iscmove.Message
	Allowance []iscmove.CoinAllowance
	GasBudget uint64
}

func (mr *moveRequest) ToRequest() *iscmove.Request {
	assets := iscmove.NewAssets(0)
	fmt.Println("mr.Allowance: ", mr.Allowance)
	for _, allowance := range mr.Allowance {
		fmt.Println("allowance: ", allowance)
		assets.AddCoin(allowance.CoinType, allowance.Balance)
	}
	return &iscmove.Request{
		ID:        mr.ID,
		Sender:    mr.Sender,
		AssetsBag: *mr.AssetsBag.Value,
		Message:   mr.Message,
		Allowance: *assets,
		GasBudget: mr.GasBudget,
	}
}

type MoveCoin struct {
	ID      iotago.ObjectID
	Balance uint64
}
