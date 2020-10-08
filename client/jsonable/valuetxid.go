package jsonable

import (
	"encoding/json"

	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

type ValueTxID struct {
	id valuetransaction.ID
}

func NewValueTxID(id *valuetransaction.ID) *ValueTxID {
	return &ValueTxID{id: *id}
}

func (id ValueTxID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.ID().String())
}

func (id *ValueTxID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r, err := valuetransaction.IDFromBase58(s)
	id.id = r
	return err
}

func (id ValueTxID) ID() *valuetransaction.ID {
	return &id.id
}
