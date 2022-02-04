package model

import (
	"encoding/json"

	iotago "github.com/iotaledger/iota.go/v3"
)

// ValueTxID is the base58 representation of a transaction ID
type ValueTxID string

func NewValueTxID(id *iotago.TransactionID) ValueTxID {
	return ValueTxID(string(id[:]))
}

func (id ValueTxID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func (id *ValueTxID) UnmarshalJSON(b []byte) error {
	panic("TODO implement")
	// var s string
	// if err := json.Unmarshal(b, &s); err != nil {
	// 	return err
	// }
	// _, err := iotago.TransactionIDFromBase58(s)
	// *id = ValueTxID(s)
	// return err
}

func (id ValueTxID) ID() iotago.TransactionID {
	panic("TODO implement")
	// r, err := iotago.TransactionIDFromBase58(string(id))
	// if err != nil {
	// 	panic(err)
	// }
	// return r
}
