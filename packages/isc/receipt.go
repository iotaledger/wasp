package isc

import (
	"encoding/hex"
	"fmt"

	"github.com/iotaledger/wasp/packages/vm/gas"
)

// Receipt represents a blocklog.RequestReceipt with a resolved error string
type Receipt struct {
	Request       []byte             `json:"request"`
	Error         *UnresolvedVMError `json:"error"`
	GasBudget     uint64             `json:"gasBudget"`
	GasBurned     uint64             `json:"gasBurned"`
	GasFeeCharged uint64             `json:"gasFeeCharged"`
	BlockIndex    uint32             `json:"blockIndex"`
	RequestIndex  uint16             `json:"requestIndex"`
	ResolvedError string             `json:"resolvedError"`
	GasBurnLog    *gas.BurnLog       `json:"-"`
}

func (r Receipt) DeserializedRequest() Request {
	req, err := NewRequestFromBytes(r.Request)
	if err != nil {
		panic(err)
	}
	return req
}

func (r Receipt) String() string {
	ret := fmt.Sprintf("ID: %s\n", r.DeserializedRequest().ID().String())
	ret += fmt.Sprintf("Err: %v\n", r.ResolvedError)
	ret += fmt.Sprintf("Block/Request index: %d / %d\n", r.BlockIndex, r.RequestIndex)
	ret += fmt.Sprintf("Gas budget / burned / fee charged: %d / %d /%d\n", r.GasBudget, r.GasBurned, r.GasFeeCharged)
	ret += fmt.Sprintf("Call data: %s\n", hex.EncodeToString(r.Request))
	return ret
}
