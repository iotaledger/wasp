package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// moveAnchor is the BCS equivalent for the move type Anchor
type moveAnchor struct {
	ID            iotago.ObjectID
	Assets        iscmove.Referent[iscmove.AssetsBag]
	StateMetadata []byte
	StateIndex    uint32
}

func (ma *moveAnchor) ToAnchor() *iscmove.Anchor {
	return &iscmove.Anchor{
		ID:            ma.ID,
		Assets:        ma.Assets,
		StateMetadata: ma.StateMetadata,
		StateIndex:    ma.StateIndex,
	}
}

type RequestResultEvent struct {
	RequestID iotago.ObjectID
	Error     string
}

type MoveRequest struct {
	ID        iotago.ObjectID
	Sender    *cryptolib.Address
	AssetsBag iscmove.Referent[iscmove.AssetsBagWithBalances]
	Message   iscmove.Message
	Allowance []iscmove.CoinAllowance
	GasBudget uint64
}

func (mr *MoveRequest) ToRequest() *iscmove.Request {
	assets := iscmove.NewAssets(0)
	for _, allowance := range mr.Allowance {
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
