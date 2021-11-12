package requestdata

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"go.uber.org/atomic"
)

type OffLedger struct {
	args       requestargs.RequestArgs
	contract   iscp.Hname
	entryPoint iscp.Hname
	params     atomic.Value // mutable
	publicKey  ed25519.PublicKey
	sender     ledgerstate.Address
	signature  ed25519.Signature
	nonce      uint64
	transfer   colored.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &OffLedger{}

// NewOffLedger creates a basic request
func NewOffLedger(contract, entryPoint iscp.Hname, args requestargs.RequestArgs) *OffLedger {
	return &OffLedger{
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes with first type byte
func (req *OffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(byte(RequestDataTypeOffLedger))
	req.writeToMarshalUtil(mu)
	return mu.Bytes()
}

// offLedgerFromMarshalUtil creates a request from previously serialized bytes. Does not expect type byte
func offLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (req *OffLedger, err error) {
	req = &OffLedger{}
	if err := req.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return req, nil
}

func (req *OffLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	req.writeEssenceToMarshalUtil(mu)
	mu.WriteBytes(req.signature[:])
}

func (req *OffLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := req.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	sig, err := mu.ReadBytes(len(req.signature))
	if err != nil {
		return err
	}
	copy(req.signature[:], sig)
	return nil
}

func (req *OffLedger) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.contract).
		Write(req.entryPoint).
		Write(req.args).
		WriteBytes(req.publicKey[:]).
		WriteUint64(req.nonce).
		Write(req.transfer)
}

func (req *OffLedger) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := req.contract.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := req.entryPoint.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	a, err := dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	req.args = requestargs.New(a)
	pk, err := mu.ReadBytes(len(req.publicKey))
	if err != nil {
		return err
	}
	copy(req.publicKey[:], pk)
	if req.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if req.transfer, err = colored.BalancesFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// only used for consensus
func (req *OffLedger) Hash() [32]byte {
	return hashing.HashData(req.Bytes())
}

// Sign signs essence
func (req *OffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	mu := marshalutil.New()
	req.writeEssenceToMarshalUtil(mu)
	req.signature = keyPair.PrivateKey.Sign(mu.Bytes())
}

// Tokens returns the transfers passed to the request
func (req *OffLedger) Tokens() colored.Balances {
	return req.transfer
}

// Tokens sets the transfers passed to the request
func (req *OffLedger) WithTransfer(transfer colored.Balances) *OffLedger {
	req.transfer = transfer
	return req
}

// VerifySignature verifies essence signature
func (req *OffLedger) VerifySignature() bool {
	mu := marshalutil.New()
	req.writeEssenceToMarshalUtil(mu)
	return req.publicKey.VerifySignature(mu.Bytes(), req.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *OffLedger) ID() (requestID iscp.RequestID) {
	txid := ledgerstate.TransactionID(hashing.HashData(req.Bytes()))
	return iscp.RequestID(ledgerstate.NewOutputID(txid, 0))
}

// IsFeePrepaid always true for off-ledger
func (req *OffLedger) IsFeePrepaid() bool {
	return true
}

// Order number used for replay protection
func (req *OffLedger) Nonce() uint64 {
	return req.nonce
}

func (req *OffLedger) WithNonce(nonce uint64) iscp.Request {
	req.nonce = nonce
	return req
}

func (req *OffLedger) IsOffLedger() bool {
	return true
}

func (req *OffLedger) Params() (dict.Dict, bool) {
	par := req.params.Load()
	if par == nil {
		return nil, false
	}
	return par.(dict.Dict), true
}

func (req *OffLedger) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(req.SenderAddress(), 0)
}

func (req *OffLedger) SenderAddress() ledgerstate.Address {
	if req.sender == nil {
		req.sender = ledgerstate.NewED25519Address(req.publicKey)
	}
	return req.sender
}

func (req *OffLedger) Target() (iscp.Hname, iscp.Hname) {
	return req.contract, req.entryPoint
}

func (req *OffLedger) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

func (req *OffLedger) SetParams(params dict.Dict) {
	req.params.Store(params)
}

func (req *OffLedger) Args() requestargs.RequestArgs {
	return req.args
}

func (req *OffLedger) String() string {
	return fmt.Sprintf("OffLedger::{ ID: %s, sender: %s, target: %s, entrypoint: %s, args: %s, nonce: %d }",
		req.ID().Base58(),
		req.SenderAddress().Base58(),
		req.contract.String(),
		req.entryPoint.String(),
		req.Args().String(),
		req.nonce,
	)
}
