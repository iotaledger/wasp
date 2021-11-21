package iscp

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/placeholders"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedger struct {
	chainID        ChainID
	contract       Hname
	entryPoint     Hname
	params         dict.Dict
	publicKey      ed25519.PublicKey
	sender         iotago.Address
	signature      ed25519.Signature
	nonce          uint64
	transferIotas  uint64
	transferTokens iotago.NativeTokens
	gasBudget      uint64
}

// implement RequestData interface
var _ RequestData = &OffLedger{}

func NewOffLedgerRequest() *OffLedger {
	return nil // TODO
}

func (r *OffLedger) Type() TypeCode {
	return TypeOffLedger
}

func (r *OffLedger) Request() Request {
	return r
}

func (r *OffLedger) TimeData() *TimeData {
	panic("implement me")
}

func (r *OffLedger) Unwrap() unwrap {
	return r
}

func (r *OffLedger) Features() Features {
	return r
}

// implements unwrap interface
var _ unwrap = &OffLedger{}

func (r *OffLedger) OffLedger() *OffLedger {
	return r
}

func (r *OffLedger) UTXO() unwrapUTXO {
	panic("not an UTXO RequestData")
}

// implements Features interface
var _ Features = &OffLedger{}

func (r *OffLedger) TimeLock() *TimeData {
	return nil
}

func (r *OffLedger) Expiry() (*TimeData, iotago.Address) {
	return nil, nil
}

func (r *OffLedger) ReturnAmount() (uint64, bool) {
	return 0, false
}

// implements iscp.Request interface
var _ Request = &OffLedger{}

// NewOffLedger creates a basic request
func NewOffLedger(contract, entryPoint Hname, params dict.Dict) *OffLedger {
	return &OffLedger{
		params:     params.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes with first type byte
func (r *OffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(byte(TypeOffLedger))
	r.writeToMarshalUtil(mu)
	return mu.Bytes()
}

// offLedgerFromMarshalUtil creates a request from previously serialized bytes. Does not expect type byte
func OffLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (req *OffLedger, err error) {
	req = &OffLedger{}
	if err := req.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *OffLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	r.writeEssenceToMarshalUtil(mu)
	mu.WriteBytes(r.signature[:])
}

func (r *OffLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := r.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	sig, err := mu.ReadBytes(len(r.signature))
	if err != nil {
		return err
	}
	copy(r.signature[:], sig)
	return nil
}

func (r *OffLedger) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(&r.chainID).
		Write(r.contract).
		Write(r.entryPoint).
		Write(r.params).
		WriteBytes(r.publicKey[:]).
		WriteUint64(r.nonce).
		WriteUint64(r.gasBudget).
		WriteUint64(r.transferIotas)
	// TODO write native Tokens
}

func (r *OffLedger) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if r.chainID, err = ChainIDFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := r.contract.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := r.entryPoint.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	_, err = dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	r.params = dict.New()
	pk, err := mu.ReadBytes(len(r.publicKey))
	if err != nil {
		return err
	}
	copy(r.publicKey[:], pk)
	if r.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.gasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.transferIotas, err = mu.ReadUint64(); err != nil {
		return err
	}

	// TODO read native Tokens
	//if r.transferTokens, err = colored.BalancesFromMarshalUtil(mu); err != nil {
	//	return err
	//}
	return nil
}

// only used for consensus
func (r *OffLedger) Hash() [32]byte {
	return hashing.HashData(r.Bytes())
}

// Sign signs essence
func (r *OffLedger) Sign(keyPair *ed25519.KeyPair) {
	r.publicKey = keyPair.PublicKey
	mu := marshalutil.New()
	r.writeEssenceToMarshalUtil(mu)
	r.signature = keyPair.PrivateKey.Sign(mu.Bytes())
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (r *OffLedger) Assets() *Assets {
	return nil
}

// Transfer transfer of assets from the sender's account to the target smart contract. Nil mean no transfer
func (r *OffLedger) Transfer() *Assets {
	return NewAssets(r.transferIotas, r.transferTokens)
}

func (r *OffLedger) WithGasBudget(gasBudget uint64) *OffLedger {
	r.gasBudget = gasBudget
	return r
}

// Tokens sets the transfers passed to the request
func (r *OffLedger) WithIotas(transferIotas uint64) *OffLedger {
	r.transferIotas = transferIotas
	return r
}

// Tokens sets the transfers passed to the request
func (r *OffLedger) WithTokens(tokens iotago.NativeTokens) *OffLedger {
	r.transferTokens = tokens // TODO clone
	return r
}

// VerifySignature verifies essence signature
func (r *OffLedger) VerifySignature() bool {
	mu := marshalutil.New()
	r.writeEssenceToMarshalUtil(mu)
	return r.publicKey.VerifySignature(mu.Bytes(), r.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (r *OffLedger) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

// Order number used for replay protection
func (r *OffLedger) Nonce() uint64 {
	return r.nonce
}

func (r *OffLedger) WithNonce(nonce uint64) Request {
	r.nonce = nonce
	return r
}

func (r *OffLedger) Params() dict.Dict {
	return r.params
}

func (r *OffLedger) SenderAccount() *AgentID {
	// TODO return iscp.NewAgentID(r.SenderAddress(), 0)
	return nil
}

func (r *OffLedger) SenderAddress() iotago.Address {
	if r.sender == nil {
		r.sender = placeholders.NewED25519Address(r.publicKey)
	}
	return r.sender
}

func (r *OffLedger) Target() RequestTarget {
	return RequestTarget{
		Contract:   r.contract,
		EntryPoint: r.entryPoint,
	}
}

func (r *OffLedger) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

func (r *OffLedger) GasBudget() uint64 {
	return r.gasBudget
}

func (r *OffLedger) String() string {
	return fmt.Sprintf("OffLedger::{ ID: %s, sender: %s, target: %s, entrypoint: %s, args: %s, nonce: %d }",
		r.ID().Base58(),
		"**not impl**", // TODO r.SenderAddress().Base58(),
		r.contract.String(),
		r.entryPoint.String(),
		r.Params().String(),
		r.nonce,
	)
}
