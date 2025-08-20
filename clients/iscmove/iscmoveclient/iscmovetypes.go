package iscmoveclient

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type MoveRequest struct {
	ID        iotago.ObjectID
	Sender    *cryptolib.Address
	AssetsBag iscmove.Referent[iscmove.AssetsBagWithBalances]
	Message   iscmove.Message
	Allowance []byte
	GasBudget uint64
}

func (mr *MoveRequest) ToRequest() *iscmove.Request {
	return &iscmove.Request{
		ID:           mr.ID,
		Sender:       mr.Sender,
		AssetsBag:    *mr.AssetsBag.Value,
		Message:      mr.Message,
		AllowanceBCS: mr.Allowance,
		GasBudget:    mr.GasBudget,
	}
}

type MoveCoin struct {
	ID      iotago.ObjectID
	Balance uint64
}
