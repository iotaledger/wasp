package iscclient

import (
	"crypto/ed25519"
	"math"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type OffLedgerRequest struct {
	ChainID    wasmtypes.ScChainID
	Contract   wasmtypes.ScHname
	EntryPoint wasmtypes.ScHname
	Params     []byte
	Signature  OffLedgerSignature
	Nonce      uint64
	Allowance  *wasmlib.ScAssets
	GasBudget  uint64
}

type OffLedgerSignature struct {
	PublicKey ed25519.PublicKey
	Signature []byte
}

func NewOffLedgerRequest(
	chainID wasmtypes.ScChainID,
	hContract, hFunction wasmtypes.ScHname,
	params []byte,
	allowance *wasmlib.ScAssets,
	nonce uint64,
) (*OffLedgerRequest, error) {
	return &OffLedgerRequest{
		ChainID:    chainID,
		Contract:   hContract,
		EntryPoint: hFunction,
		Params:     params,
		Nonce:      nonce,
		Allowance:  allowance,
		GasBudget:  math.MaxUint64,
	}, nil
}

func (req *OffLedgerRequest) Bytes() []byte {
	enc := req.essenceEncode()
	enc.FixedBytes(req.Signature.PublicKey, 32)
	enc.Bytes(req.Signature.Signature)
	return enc.Buf()
}

func (req *OffLedgerRequest) Essence() []byte {
	return req.essenceEncode().Buf()
}

func (req *OffLedgerRequest) essenceEncode() *wasmtypes.WasmEncoder {
	enc := wasmtypes.NewWasmEncoder()
	enc.Byte(1) // requestKindOffLedgerISC
	wasmtypes.ChainIDEncode(enc, req.ChainID)
	wasmtypes.HnameEncode(enc, req.Contract)
	wasmtypes.HnameEncode(enc, req.EntryPoint)
	enc.FixedBytes(req.Params, uint32(len(req.Params)))
	enc.VluEncode(req.Nonce)
	gasBudget := req.GasBudget
	if gasBudget < math.MaxUint64 {
		gasBudget++
	} else {
		gasBudget = 0
	}
	enc.VluEncode(gasBudget)
	allowance := req.Allowance.Bytes()
	enc.FixedBytes(allowance, uint32(len(allowance)))
	return enc
}

func (req *OffLedgerRequest) ID() wasmtypes.ScRequestID {
	hash := blake2b.Sum256(req.Bytes())
	// req id is hash of req bytes with concatenated output index zero
	return wasmtypes.RequestIDFromBytes(append(hash[:], 0, 0))
}

func (req *OffLedgerRequest) Sign(keyPair *Keypair) {
	req.Signature.PublicKey = keyPair.GetPublicKey()
	hash := blake2b.Sum256(req.Essence())
	req.Signature.Signature = keyPair.Sign(hash[:])
}
