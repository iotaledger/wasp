package isc

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/samber/lo"
	"golang.org/x/crypto/blake2b"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

// OffLedgerRequestDataEssence implements UnsignedOffLedgerRequest
type OffLedgerRequestDataEssence struct {
	allowance *Assets `bcs:"export,optional"`
	chainID   ChainID `bcs:"export"`
	msg       Message `bcs:"export"`
	gasBudget uint64  `bcs:"export"`
	nonce     uint64  `bcs:"export"`
}

// OffLedgerRequestData implements OffLedgerRequest
type OffLedgerRequestData struct {
	OffLedgerRequestDataEssence

	signature *cryptolib.Signature `bcs:"export"`
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

func NewImpersonatedOffLedgerRequest(request *OffLedgerRequestDataEssence) ImpersonatedOffLedgerRequest {
	return &ImpersonatedOffLedgerRequestData{
		OffLedgerRequestData: OffLedgerRequestData{
			OffLedgerRequestDataEssence: OffLedgerRequestDataEssence{
				allowance: request.allowance,
				msg:       request.msg,
				gasBudget: request.gasBudget,
				nonce:     request.nonce,
			},
			signature: cryptolib.NewDummySignature(cryptolib.NewEmptyPublicKey()),
		},
		address: nil,
	}
}

func (r *ImpersonatedOffLedgerRequestData) WithSenderAddress(senderAddress *cryptolib.Address) OffLedgerRequest {
	r.address = senderAddress
	return r
}

func (r *ImpersonatedOffLedgerRequestData) SenderAccount() AgentID {
	return NewAddressAgentID(r.address)
}

func NewOffLedgerRequest(
	chainID ChainID,
	msg Message,
	nonce uint64,
	gasBudget uint64,
) UnsignedOffLedgerRequest {
	return &OffLedgerRequestDataEssence{
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
func (req *OffLedgerRequestData) Allowance() (*Assets, error) {
	return req.allowance, nil
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (req *OffLedgerRequestData) Assets() *Assets {
	return nil
}

func (req *OffLedgerRequestData) Bytes() []byte {
	var r Request = req
	return bcs.MustMarshal(&r)
}

func (req *OffLedgerRequestData) Message() Message {
	return req.msg
}

func (req *OffLedgerRequestDataEssence) Bytes() []byte {
	return bcs.MustMarshal(req)
}

func (req *OffLedgerRequestDataEssence) messageToSign() []byte {
	ret := blake2b.Sum256(req.Bytes())
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
	return NewAddressAgentID(req.signature.GetPublicKey().AsAddress())
}

func (req *OffLedgerRequestDataEssence) Sign(signer cryptolib.Signer) OffLedgerRequest {
	signature := lo.Must(signer.Sign(req.messageToSign()))
	return &OffLedgerRequestData{
		OffLedgerRequestDataEssence: *req,
		signature:                   signature,
	}
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

func (req *OffLedgerRequestDataEssence) WithAllowance(allowance *Assets) UnsignedOffLedgerRequest {
	req.allowance = allowance.Clone()
	return req
}

func (req *OffLedgerRequestDataEssence) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	req.gasBudget = gasBudget
	return req
}

func (req *OffLedgerRequestDataEssence) WithNonce(nonce uint64) UnsignedOffLedgerRequest {
	req.nonce = nonce
	return req
}

// WithSender can be used to estimate gas, without a signature
func (req *OffLedgerRequestDataEssence) WithSender(sender *cryptolib.PublicKey) OffLedgerRequest {
	return &OffLedgerRequestData{
		OffLedgerRequestDataEssence: *req,
		signature:                   cryptolib.NewDummySignature(sender),
	}
}

func (req *OffLedgerRequestData) GasPrice() *big.Int {
	return nil
}
