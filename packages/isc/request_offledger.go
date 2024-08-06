package isc

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/minio/blake2b-simd"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type OffLedgerRequestData struct {
	allowance *Assets
	chainID   ChainID
	msg       Message
	gasBudget uint64
	nonce     uint64
	signature *cryptolib.Signature
}

var (
	_ Request                  = new(OffLedgerRequestData)
	_ OffLedgerRequest         = new(OffLedgerRequestData)
	_ UnsignedOffLedgerRequest = new(OffLedgerRequestData)
	_ Calldata                 = new(OffLedgerRequestData)
)

type ImpersonatedOffLedgerRequestData struct {
	OffLedgerRequestData
	address *cryptolib.Address
}

func NewImpersonatedOffLedgerRequest(request *OffLedgerRequestData) ImpersonatedOffLedgerRequest {
	copyReq := *request
	copyReq.signature = nil
	return &ImpersonatedOffLedgerRequestData{
		OffLedgerRequestData: copyReq,
		address:              nil,
	}
}

func (r *ImpersonatedOffLedgerRequestData) WithSenderAddress(address *cryptolib.Address) OffLedgerRequest {
	addressBytes := address.Bytes()
	copy(r.address[:], addressBytes)
	return r
}

func (r *ImpersonatedOffLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(r.address)
}

func NewOffLedgerRequest(
	chainID ChainID,
	msg Message,
	nonce uint64,
	gasBudget uint64,
) UnsignedOffLedgerRequest {
	return &OffLedgerRequestData{
		chainID:   chainID,
		msg:       msg,
		nonce:     nonce,
		allowance: NewEmptyAssets(),
		gasBudget: gasBudget,
	}
}

func (req *OffLedgerRequestData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	req.readEssence(rr)
	req.signature = cryptolib.NewEmptySignature()
	rr.Read(req.signature)
	return rr.Err
}

func (req *OffLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}

func (req *OffLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	req.writeEssence(ww)
	ww.Write(req.signature)
	return ww.Err
}

func (req *OffLedgerRequestData) readEssence(rr *rwutil.Reader) {
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOffLedgerISC))
	rr.Read(&req.chainID)
	rr.Read(&req.msg.Target.Contract)
	rr.Read(&req.msg.Target.EntryPoint)
	req.msg.Params = CallArguments{}
	rr.Read(&req.msg.Params)
	req.nonce = rr.ReadAmount64()
	req.gasBudget = rr.ReadGas64()
	req.allowance = NewEmptyAssets()
	rr.Read(req.allowance)
}

func (req *OffLedgerRequestData) writeEssence(ww *rwutil.Writer) {
	ww.WriteKind(rwutil.Kind(requestKindOffLedgerISC))
	ww.Write(&req.chainID)
	ww.Write(&req.msg.Target.Contract)
	ww.Write(&req.msg.Target.EntryPoint)
	ww.Write(&req.msg.Params)
	ww.WriteAmount64(req.nonce)
	ww.WriteGas64(req.gasBudget)
	ww.Write(req.allowance)
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (req *OffLedgerRequestData) Allowance() *Assets {
	return req.allowance
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (req *OffLedgerRequestData) Assets() *Assets {
	return nil
}

func (req *OffLedgerRequestData) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *OffLedgerRequestData) Message() Message {
	return req.msg
}

func (req *OffLedgerRequestData) ChainID() ChainID {
	return req.chainID
}

func (req *OffLedgerRequestData) EssenceBytes() []byte {
	ww := rwutil.NewBytesWriter()
	req.writeEssence(ww)
	return ww.Bytes()
}

func (req *OffLedgerRequestData) messageToSign() []byte {
	ret := blake2b.Sum256(req.EssenceBytes())
	return ret[:]
}

func (req *OffLedgerRequestData) Expiry() (time.Time, *cryptolib.Address) {
	return time.Time{}, nil
}

func (req *OffLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	return req.gasBudget, false
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *OffLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(req.Bytes())), 0)
}

func (req *OffLedgerRequestData) IsOffLedger() bool {
	return true
}

// Nonce incremental nonce used for replay protection
func (req *OffLedgerRequestData) Nonce() uint64 {
	return req.nonce
}

func (req *OffLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

func (req *OffLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(req.signature.GetPublicKey().AsAddress())
}

// Sign signs the essence
func (req *OffLedgerRequestData) Sign(signer cryptolib.Signer) OffLedgerRequest {
	signature, _ := signer.Sign(req.messageToSign())
	req.signature = signature
	return req
}

func (req *OffLedgerRequestData) String() string {
	return fmt.Sprintf("offLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
		req.ID().String(),
		req.SenderAccount().String(),
		req.msg.Target.Contract.String(),
		req.msg.Target.EntryPoint.String(),
		req.msg.Params,
		req.nonce,
	)
}

func (req *OffLedgerRequestData) TargetAddress() *cryptolib.Address {
	return req.chainID.AsAddress()
}

func (req *OffLedgerRequestData) TimeLock() time.Time {
	return time.Time{}
}

func (req *OffLedgerRequestData) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

// VerifySignature verifies essence signature
func (req *OffLedgerRequestData) VerifySignature() error {
	if !req.signature.Validate(req.messageToSign()) {
		return errors.New("invalid signature")
	}
	return nil
}

func (req *OffLedgerRequestData) WithAllowance(allowance *Assets) UnsignedOffLedgerRequest {
	req.allowance = allowance.Clone()
	return req
}

func (req *OffLedgerRequestData) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	req.gasBudget = gasBudget
	return req
}

func (req *OffLedgerRequestData) WithNonce(nonce uint64) UnsignedOffLedgerRequest {
	req.nonce = nonce
	return req
}

// WithSender can be used to estimate gas, without a signature
func (req *OffLedgerRequestData) WithSender(sender *cryptolib.PublicKey) UnsignedOffLedgerRequest {
	req.signature = cryptolib.NewDummySignature(sender)
	return req
}

func (req *OffLedgerRequestData) GasPrice() *big.Int {
	return nil
}
