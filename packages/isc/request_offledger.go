package isc

import (
	"errors"
	"fmt"
	"io"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type offLedgerSignature struct {
	publicKey *cryptolib.PublicKey
	signature []byte
}

type offLedgerRequestData struct {
	allowance  *Assets
	chainID    ChainID
	contract   Hname
	entryPoint Hname
	gasBudget  uint64
	nonce      uint64
	params     dict.Dict
	signature  offLedgerSignature
}

var (
	_ Request                  = new(offLedgerRequestData)
	_ OffLedgerRequest         = new(offLedgerRequestData)
	_ UnsignedOffLedgerRequest = new(offLedgerRequestData)
	_ Calldata                 = new(offLedgerRequestData)
	_ Features                 = new(offLedgerRequestData)
)

func NewOffLedgerRequest(
	chainID ChainID,
	contract, entryPoint Hname,
	params dict.Dict,
	nonce uint64,
	gasBudget uint64,
) UnsignedOffLedgerRequest {
	return &offLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		nonce:      nonce,
		allowance:  NewEmptyAssets(),
		gasBudget:  gasBudget,
	}
}

func (req *offLedgerRequestData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	req.readEssence(rr)
	publicKey := rr.ReadBytes()
	if rr.Err != nil {
		return rr.Err
	}
	req.signature.publicKey, rr.Err = cryptolib.PublicKeyFromBytes(publicKey)
	req.signature.signature = rr.ReadBytes()
	return rr.Err
}

func (req *offLedgerRequestData) readEssence(rr *rwutil.Reader) {
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOffLedgerISC))
	rr.Read(&req.chainID)
	rr.Read(&req.contract)
	rr.Read(&req.entryPoint)
	req.params = dict.New()
	rr.Read(req.params)
	req.nonce = rr.ReadUint64()
	req.gasBudget = rr.ReadUint64()
	req.allowance = NewEmptyAssets()
	rr.Read(req.allowance)
}

func (req *offLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	req.writeEssence(ww)
	ww.WriteBytes(req.signature.publicKey.AsBytes())
	ww.WriteBytes(req.signature.signature)
	return ww.Err
}

func (req *offLedgerRequestData) writeEssence(ww *rwutil.Writer) {
	ww.WriteKind(rwutil.Kind(requestKindOffLedgerISC))
	ww.Write(&req.chainID)
	ww.Write(&req.contract)
	ww.Write(&req.entryPoint)
	ww.Write(req.params)
	ww.WriteUint64(req.nonce)
	ww.WriteUint64(req.gasBudget)
	ww.Write(req.allowance)
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (req *offLedgerRequestData) Allowance() *Assets {
	return req.allowance
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (req *offLedgerRequestData) Assets() *Assets {
	return nil
}

func (req *offLedgerRequestData) Bytes() []byte {
	return rwutil.WriterToBytes(req)
}

func (req *offLedgerRequestData) CallTarget() CallTarget {
	return CallTarget{
		Contract:   req.contract,
		EntryPoint: req.entryPoint,
	}
}

func (req *offLedgerRequestData) ChainID() ChainID {
	return req.chainID
}

func (req *offLedgerRequestData) essenceBytes() []byte {
	ww := rwutil.NewBytesWriter()
	req.writeEssence(ww)
	return ww.Bytes()
}

func (req *offLedgerRequestData) Expiry() (time.Time, iotago.Address) {
	return time.Time{}, nil
}

func (req *offLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	return req.gasBudget, false
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *offLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(req.Bytes())), 0)
}

func (req *offLedgerRequestData) IsOffLedger() bool {
	return true
}

func (req *offLedgerRequestData) NFT() *NFT {
	return nil
}

// Nonce incremental nonce used for replay protection
func (req *offLedgerRequestData) Nonce() uint64 {
	return req.nonce
}

func (req *offLedgerRequestData) Params() dict.Dict {
	return req.params
}

func (req *offLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

func (req *offLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(req.signature.publicKey.AsEd25519Address())
}

// Sign signs the essence
func (req *offLedgerRequestData) Sign(key *cryptolib.KeyPair) OffLedgerRequest {
	req.signature = offLedgerSignature{
		publicKey: key.GetPublicKey(),
		signature: key.GetPrivateKey().Sign(req.essenceBytes()),
	}
	return req
}

func (req *offLedgerRequestData) String() string {
	return fmt.Sprintf("offLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
		req.ID().String(),
		req.SenderAccount().String(),
		req.contract.String(),
		req.entryPoint.String(),
		req.Params().String(),
		req.nonce,
	)
}

func (req *offLedgerRequestData) TargetAddress() iotago.Address {
	return req.chainID.AsAddress()
}

func (req *offLedgerRequestData) TimeLock() time.Time {
	return time.Time{}
}

func (req *offLedgerRequestData) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

// VerifySignature verifies essence signature
func (req *offLedgerRequestData) VerifySignature() error {
	if !req.signature.publicKey.Verify(req.essenceBytes(), req.signature.signature) {
		return errors.New("invalid signature")
	}
	return nil
}

func (req *offLedgerRequestData) WithAllowance(allowance *Assets) UnsignedOffLedgerRequest {
	req.allowance = allowance.Clone()
	return req
}

func (req *offLedgerRequestData) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	req.gasBudget = gasBudget
	return req
}

func (req *offLedgerRequestData) WithNonce(nonce uint64) UnsignedOffLedgerRequest {
	req.nonce = nonce
	return req
}

// WithSender can be used to estimate gas, without a signature
func (req *offLedgerRequestData) WithSender(sender *cryptolib.PublicKey) UnsignedOffLedgerRequest {
	req.signature = offLedgerSignature{
		publicKey: sender,
		signature: []byte{},
	}
	return req
}
