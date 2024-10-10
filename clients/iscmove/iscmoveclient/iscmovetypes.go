package iscmoveclient

import (
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
	id            iotago.ObjectID
	assets        referent[iscmove.AssetsBag]
	stateMetadata []byte
	stateIndex    uint32
}

func (ma *moveAnchor) ToAnchor() *iscmove.Anchor {
	return &iscmove.Anchor{
		ID:            ma.id,
		Assets:        *ma.assets.Value,
		StateMetadata: ma.stateMetadata,
		StateIndex:    ma.stateIndex,
	}
}

type moveRequest struct {
	id     iotago.ObjectID
	sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
	assetsBag referent[iscmove.AssetsBagWithBalances] // Need to decide if we want to use this Referent wrapper as well. Could probably be of *AssetsBag with `bcs:"optional`
	message   iscmove.Message
	allowance []iscmove.CoinAllowance
	gasBudget uint64
}

func (mr *moveRequest) ToRequest() *iscmove.Request {
	return &iscmove.Request{
		ID:        mr.id,
		Sender:    mr.sender,
		AssetsBag: *mr.assetsBag.Value,
		Message:   mr.message,
		Allowance: mr.allowance,
		GasBudget: mr.gasBudget,
	}
}
