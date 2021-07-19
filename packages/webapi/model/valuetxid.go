package model

import (
	"encoding/json"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

// ValueTxID is the base58 representation of a transaction ID
type ValueTxID string

func NewValueTxID(id *ledgerstate.TransactionID) ValueTxID {
	return ValueTxID(id.String())
}

func (id ValueTxID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func (id *ValueTxID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := ledgerstate.TransactionIDFromBase58(s)
	*id = ValueTxID(s)
	return err
}

func (id ValueTxID) ID() ledgerstate.TransactionID {
	r, err := ledgerstate.TransactionIDFromBase58(string(id))
	if err != nil {
		panic(err)
	}
	return r
}
