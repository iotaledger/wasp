package coretypes

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

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
	return fmt.Sprintf("[%d]%s", rid.OutputID().OutputIndex(), rid.OutputID().TransactionID().Base58())
}

func (rid RequestID) Short() string {
	txid := rid.OutputID().TransactionID().Base58()
	return fmt.Sprintf("[%d]%s", rid.OutputID().OutputIndex(), txid[:6]+"..")
}

// Request has two main implementation
// - sctransaction.RequestOnLedger
// - RequestOffLedger
type Request interface {
	// index == 0 for off ledger requests
	ID() RequestID
	// ledgerstate.Output interface for on-ledger reguests, nil for off-ledger requests
	Output() ledgerstate.Output
	// address of the sender for all requests,
	SenderAddress() ledgerstate.Address
	// account of the sander
	SenderAccount() *AgentID
	// returns contract/entry point pair
	Target() (Hname, Hname)
	// true or false for on-ledger requests, true for off-ledger
	IsFeePrepaid() bool
	// always nil for off-ledger
	Tokens() *ledgerstate.ColoredBalances
	// arguments of the call. Must be != nil (solidified). No arguments means empty dictionary
	Params() dict.Dict
}
