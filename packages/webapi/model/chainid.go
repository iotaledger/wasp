package model

import (
	"encoding/json"

	"github.com/iotaledger/wasp/packages/isc"
)

// ChainIDBech32 is the string representation of isc.ChainIDBech32 (bech32)
type ChainIDBech32 string

func NewChainIDBech32(chainID isc.ChainID) ChainIDBech32 {
	return ChainIDBech32(chainID.String())
}

func (ch ChainIDBech32) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ch))
}

func (ch *ChainIDBech32) UnmarshalJSON(b []byte) error {
	if len(b) <= 2 { // empty string means len(b) == 2 // ("")
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := isc.ChainIDFromString(s)
	*ch = ChainIDBech32(s)
	return err
}

func (ch ChainIDBech32) ChainID() isc.ChainID {
	chainID, err := isc.ChainIDFromString(string(ch))
	if err != nil {
		panic(err)
	}
	return chainID
}
