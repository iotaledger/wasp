package coretypes

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

type RequestID ledgerstate.OutputID

func (rid RequestID) String() string {
	return fmt.Sprintf("[%d]%s", ledgerstate.OutputID(rid).OutputIndex(), ledgerstate.OutputID(rid).TransactionID().Base58())
}

func (rid RequestID) Short() string {
	txid := ledgerstate.OutputID(rid).TransactionID().Base58()
	return fmt.Sprintf("[%d]%s", ledgerstate.OutputID(rid).OutputIndex(), txid[:6]+"..")
}
