package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"time"
)

// region RequestOffLedger  ///////////////////////////////////////////////////////

type RequestOffLedger struct {
	args       requestargs.RequestArgs
	contract   coretypes.Hname
	entryPoint coretypes.Hname
	params     dict.Dict
	publicKey  ed25519.PublicKey
	sender     ledgerstate.Address
	signature  ed25519.Signature
	timestamp  time.Time
	transfer   *ledgerstate.ColoredBalances
}

// implements coretypes.Request interface
var _ coretypes.Request = &RequestOffLedger{}

// NewRequestOffLedger creates a basic request
func NewRequestOffLedger(contract coretypes.Hname, entryPoint coretypes.Hname, args requestargs.RequestArgs) *RequestOffLedger {
	return &RequestOffLedger{
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		timestamp:  time.Now(),
	}
}

// NewRequestOffLedgerFromBytes creates a basic request from previously serialized bytes
func NewRequestOffLedgerFromBytes(data []byte) (request *RequestOffLedger, err error) {
	req := &RequestOffLedger{
		args: requestargs.New(nil),
	}
	buf := bytes.NewBuffer(data)
	if err = req.contract.Read(buf); err != nil {
		return
	}
	if err = req.entryPoint.Read(buf); err != nil {
		return
	}
	if err = req.args.Read(buf); err != nil {
		return
	}
	var n int
	n, err = buf.Read(req.publicKey[:])
	if err != nil || n != len(req.publicKey) {
		return nil, io.EOF
	}
	if err = util.ReadTime(buf, &req.timestamp); err != nil {
		return
	}
	var colors uint32
	if err = util.ReadUint32(buf, &colors); err != nil {
		return
	}
	if colors != 0 {
		balances := make(map[ledgerstate.Color]uint64)
		for i := uint32(0); i < colors; i++ {
			var color ledgerstate.Color
			n, err = buf.Read(color[:])
			if err != nil || n != len(color) {
				return nil, io.EOF
			}
			var balance uint64
			if err = util.ReadUint64(buf, &balance); err != nil {
				return
			}
			balances[color] = balance
		}
		req.transfer = ledgerstate.NewColoredBalances(balances)
	}
	n, err = buf.Read(req.signature[:])
	if err != nil || n != len(req.signature) {
		return nil, io.EOF
	}
	return req, nil
}

// Essence encodes request essence as bytes
func (req *RequestOffLedger) Essence() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	_ = req.contract.Write(buf)
	_ = req.entryPoint.Write(buf)
	_ = req.args.Write(buf)
	_, _ = buf.Write(req.publicKey[:])
	_ = util.WriteTime(buf, req.timestamp)
	if req.transfer == nil {
		_ = util.WriteUint32(buf, 0)
		return buf.Bytes()
	}
	_, _ = buf.Write(req.transfer.Bytes())
	return buf.Bytes()
}

// Bytes encodes request as bytes
func (req *RequestOffLedger) Bytes() []byte {
	return append(req.Essence(), req.signature[:]...)
}

// Sign signs essence
func (req *RequestOffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	req.signature = keyPair.PrivateKey.Sign(req.Essence())
}

// Transfer returns transfers passed to request
func (req *RequestOffLedger) Transfer(transfer *ledgerstate.ColoredBalances) {
	req.transfer = transfer
}

// VerifySignature verifies essence signature
func (req *RequestOffLedger) VerifySignature() bool {
	return req.publicKey.VerifySignature(req.Essence(), req.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *RequestOffLedger) ID() (requestId coretypes.RequestID) {
	txid := ledgerstate.TransactionID(hashing.HashData(req.Bytes()))
	return coretypes.RequestID(ledgerstate.NewOutputID(txid, 0))
}

// IsFeePrepaid always true for off-ledger
func (req *RequestOffLedger) IsFeePrepaid() bool {
	return true
}

// Order number used for ordering requests in the mempool. Priority order is a descending order
func (req *RequestOffLedger) Order() uint64 {
	return uint64(req.timestamp.UnixNano())
}

// Output nil for off-ledger requests
func (req *RequestOffLedger) Output() ledgerstate.Output {
	return nil
}

func (req *RequestOffLedger) Params() (dict.Dict, bool) {
	return req.params, req.params != nil
}

func (req *RequestOffLedger) SenderAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(req.SenderAddress(), 0)
}

func (req *RequestOffLedger) SenderAddress() ledgerstate.Address {
	if req.sender == nil {
		req.sender = ledgerstate.NewED25519Address(req.publicKey)
	}
	return req.sender
}

// SolidifyArgs return true if solidified successfully
func (req *RequestOffLedger) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
	if req.params != nil {
		return true, nil
	}
	solid, ok, err := req.args.SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	req.params = solid
	if req.params == nil {
		panic("req.solidArgs == nil")
	}
	return true, nil
}

func (req *RequestOffLedger) Target() (coretypes.Hname, coretypes.Hname) {
	return req.contract, req.entryPoint
}

// TimeLock returns time lock time or zero time if no time lock
func (req *RequestOffLedger) TimeLock() time.Time {
	// no time lock, return zero time
	return time.Time{}
}

func (req *RequestOffLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.transfer
}

// endregion /////////////////////////////////////////////////////////////////
