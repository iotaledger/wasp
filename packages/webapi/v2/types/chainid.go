package types

import (
	"encoding/json"

	"github.com/iotaledger/wasp/packages/isc"
)

// ChainID is the string representation of isc.ChainID (bech32)
type ChainID string

func NewChainID(chainID *isc.ChainID) ChainID {
	return ChainID(chainID.String())
}

func (ch ChainID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ch))
}

func (ch *ChainID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := isc.ChainIDFromString(s)
	*ch = ChainID(s)
	return err
}

func (ch ChainID) ChainID() *isc.ChainID {
	chainID, err := isc.ChainIDFromString(string(ch))
	if err != nil {
		panic(err)
	}
	return chainID
}
