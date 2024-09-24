package isc

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/minio/blake2b-simd"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type OffLedgerRequestData struct {
	allowance *Assets              `bcs:""`
	chainID   ChainID              `bcs:""`
	msg       Message              `bcs:""`
	gasBudget uint64               `bcs:""`
	nonce     uint64               `bcs:""`
	signature *cryptolib.Signature `bcs:""`
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

func (req *OffLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
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
	return bcs.MustMarshal(req)
}

func (req *OffLedgerRequestData) Equals(other Request) bool {
	otherR, ok := other.(*OffLedgerRequestData)
	if !ok {
		return false
	}
	return req.allowance.Equals(otherR.allowance) &&
		req.chainID.Equals(otherR.chainID) &&
		//req.msg.Equals(otherR.msg) &&
		req.gasBudget == otherR.gasBudget &&
		req.nonce == otherR.nonce &&
		req.signature.Equals(otherR.signature)
}

func (req *OffLedgerRequestData) Message() Message {
	return req.msg
}

func (req *OffLedgerRequestData) ChainID() ChainID {
	return req.chainID
}

func (req *OffLedgerRequestData) EssenceBytes() []byte {
	type offLedgerRequestDataEssence struct { // TODO: why is it not a separate embedded type?
		Allowance *Assets
		ChainID   ChainID
		Msg       Message
		GasBudget uint64
		Nonce     uint64
	}

	return bcs.MustMarshal(&offLedgerRequestDataEssence{
		Allowance: req.allowance,
		ChainID:   req.chainID,
		Msg:       req.msg,
		GasBudget: req.gasBudget,
		Nonce:     req.nonce,
	})
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
	return RequestID(hashing.HashData(req.Bytes()))
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
