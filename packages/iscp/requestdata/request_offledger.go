package requestdata

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedger struct {
	contract       iscp.Hname
	entryPoint     iscp.Hname
	params         dict.Dict
	publicKey      ed25519.PublicKey
	sender         iotago.Address
	signature      ed25519.Signature
	nonce          uint64
	transferIotas  uint64
	transferTokens iotago.NativeTokens
	gasBudget      int64
}

// implements iscp.Request interface
var _ Request = &OffLedger{}

// NewOffLedger creates a basic request
func NewOffLedger(contract, entryPoint iscp.Hname, params dict.Dict) *OffLedger {
	return &OffLedger{
		params:     params.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes with first type byte
func (req *OffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(byte(TypeOffLedger))
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
		Write(req.params).
		WriteBytes(req.publicKey[:]).
		WriteUint64(req.nonce).
		WriteUint64(req.transferIotas)
	// TODO write native tokens
}

func (req *OffLedger) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := req.contract.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := req.entryPoint.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	_, err := dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	req.params = dict.New()
	pk, err := mu.ReadBytes(len(req.publicKey))
	if err != nil {
		return err
	}
	copy(req.publicKey[:], pk)
	if req.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	// TODO read native tokens
	//if req.transferTokens, err = colored.BalancesFromMarshalUtil(mu); err != nil {
	//	return err
	//}
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
func (req *OffLedger) Assets() (uint64, iotago.NativeTokens) {
	return req.transferIotas, req.transferTokens
}

func (req *OffLedger) WithGasBudget(gasBudget int64) *OffLedger {
	req.gasBudget = gasBudget
	return req
}

// Tokens sets the transfers passed to the request
func (req *OffLedger) WithIotas(transferIotas uint64) *OffLedger {
	req.transferIotas = transferIotas
	return req
}

// Tokens sets the transfers passed to the request
func (req *OffLedger) WithTokens(tokens iotago.NativeTokens) *OffLedger {
	req.transferTokens = tokens // TODO clone
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
func (req *OffLedger) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(req.Bytes())), 0)
}

// Order number used for replay protection
func (req *OffLedger) Nonce() uint64 {
	return req.nonce
}

func (req *OffLedger) WithNonce(nonce uint64) Request {
	req.nonce = nonce
	return req
}

func (req *OffLedger) Params() dict.Dict {
	return req.params
}

func (req *OffLedger) SenderAccount() *iscp.AgentID {
	// TODO return iscp.NewAgentID(req.SenderAddress(), 0)
	return nil
}

func (req *OffLedger) SenderAddress() iotago.Address {
	if req.sender == nil {
		req.sender = placeholders.NewED25519Address(req.publicKey)
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

func (req *OffLedger) GasBudget() int64 {
	return req.gasBudget
}

func (req *OffLedger) String() string {
	return fmt.Sprintf("OffLedger::{ ID: %s, sender: %s, target: %s, entrypoint: %s, args: %s, nonce: %d }",
		req.ID().Base58(),
		"**not impl**", // TODO req.SenderAddress().Base58(),
		req.contract.String(),
		req.entryPoint.String(),
		req.Params().String(),
		req.nonce,
	)
}
