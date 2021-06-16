package model

import (
	"encoding/json"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"
)

// ChainID is the base58 representation of chainid.ChainID
type ChainID string

func NewChainID(chainID *chainid.ChainID) ChainID {
	return ChainID(chainID.Base58())
}

func (ch ChainID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ch))
}

func (ch *ChainID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := chainid.ChainIDFromBase58(s)
	*ch = ChainID(s)
	return err
}

func (ch ChainID) ChainID() *chainid.ChainID {
	chainID, err := chainid.ChainIDFromBase58(string(ch))
	if err != nil {
		panic(err)
	}
	return chainID
}
