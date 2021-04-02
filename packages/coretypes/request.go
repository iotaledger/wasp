package coretypes

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// region RequestID ///////////////////////////////////////////////////////////////
type RequestID ledgerstate.OutputID

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	ret, err := ledgerstate.OutputIDFromMarshalUtil(mu)
	return RequestID(ret), err
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromBase58(b58 string) (ret RequestID, err error) {
	var oid ledgerstate.OutputID
	oid, err = ledgerstate.OutputIDFromBase58(b58)
	if err != nil {
		return
	}
	ret = RequestID(oid)
	return
}

func (rid RequestID) OutputID() ledgerstate.OutputID {
	return ledgerstate.OutputID(rid)
}

// Base58 returns a base58 encoded version of the request id.
func (rid RequestID) Base58() string {
	return ledgerstate.OutputID(rid).Base58()
}

func (rid RequestID) Bytes() []byte {
	return rid[:]
}

func (rid RequestID) String() string {
	return OID(rid.OutputID())
}

func (rid RequestID) Short() string {
	txid := rid.OutputID().TransactionID().Base58()
	return fmt.Sprintf("[%d]%s", rid.OutputID().OutputIndex(), txid[:6]+"..")
}

func OID(o ledgerstate.OutputID) string {
	return fmt.Sprintf("[%d]%s", o.OutputIndex(), o.TransactionID().Base58())
}

// endregion ////////////////////////////////////////////////////////////////////////////////////

// Request has two main implementation
// - sctransaction.RequestOnLedger
// - RequestOffLedger
type Request interface {
	// index == 0 for off ledger requests
	ID() RequestID
	// true or false for on-ledger requests, true for off-ledger
	IsFeePrepaid() bool
	// ledgerstate.Output interface for on-ledger reguests, nil for off-ledger requests
	Output() ledgerstate.Output
	// arguments of the call with the flag if they are ready. No arguments mean empty dictionary and true
	Params() (dict.Dict, bool)
	// account of the sender
	SenderAccount() *AgentID
	// address of the sender for all requests,
	SenderAddress() ledgerstate.Address
	// returns contract/entry point pair
	Target() (Hname, Hname)
	// always nil for off-ledger
	Tokens() *ledgerstate.ColoredBalances
	// number used for ordering requests in the mempool. Priority order is a descending order
	Order() uint64
}
