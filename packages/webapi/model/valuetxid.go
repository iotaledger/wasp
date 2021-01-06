package model

import (
	"encoding/json"

	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

// ValueTxID is the base58 representation of a transaction ID
type ValueTxID string

func NewValueTxID(id *valuetransaction.ID) ValueTxID {
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
	_, err := valuetransaction.IDFromBase58(s)
	*id = ValueTxID(s)
	return err
}

func (id ValueTxID) ID() valuetransaction.ID {
	r, err := valuetransaction.IDFromBase58(string(id))
	if err != nil {
		panic(err)
	}
	return r
}
