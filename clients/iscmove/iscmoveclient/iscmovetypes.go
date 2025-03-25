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

// intermediateMoveRequest is used to decode actual requests coming from move.
// The only difference between this and MoveRequest is the AssetsBag
// The Balances in AssetsBagWithBalance are unavailable in the bcs encoded Request coming from L1
// The type will get mapped into an actual MoveRequest after it has been enriched.
// It decouples the problem that other types which depend on AssetsBagWithBalances can't properly encode Balances.
type intermediateMoveRequest struct {
	ID     iotago.ObjectID
	Sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
	AssetsBag iscmove.Referent[iscmove.AssetsBag]
	Message   iscmove.Message
	Allowance []iscmove.CoinAllowance
	GasBudget uint64
}

type MoveRequest struct {
	ID     iotago.ObjectID
	Sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
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
