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
	args []byte,
	allowance *wasmlib.ScAssets,
	nonce uint64,
) (*OffLedgerRequest, error) {
	return &OffLedgerRequest{
		ChainID:    chainID,
		Contract:   hContract,
		EntryPoint: hFunction,
		Params:     args,
		Nonce:      nonce,
		Allowance:  allowance,
		GasBudget:  math.MaxUint64,
	}, nil
}

func (req *OffLedgerRequest) Bytes() []byte {
	enc := wasmtypes.NewWasmEncoder()
	enc.FixedBytes(req.Signature.PublicKey, 32)
	enc.Bytes(req.Signature.Signature)
	return enc.Buf()
}

func (req *OffLedgerRequest) Essence() []byte {
	enc := wasmtypes.NewWasmEncoder()
	enc.Byte(1) // requestKindOffLedgerISC
	wasmtypes.ChainIDEncode(enc, req.ChainID)
	wasmtypes.HnameEncode(enc, req.Contract)
	wasmtypes.HnameEncode(enc, req.EntryPoint)
	enc.FixedBytes(req.Params, uint32(len(req.Params)))
	enc.VluEncode(req.Nonce)
	if req.GasBudget < math.MaxUint64 {
		req.GasBudget++
	} else {
		req.GasBudget = 0
	}
	enc.VluEncode(req.GasBudget)
	enc.FixedBytes(req.Allowance.Bytes(), uint32(len(req.Allowance.Bytes())))
	return enc.Buf()
}

func (req *OffLedgerRequest) Sign(keyPair *Keypair) {
	req.Signature.PublicKey = keyPair.GetPublicKey()
	req.Signature.Signature = ed25519.Sign(keyPair.GetPrivateKey(), req.Essence())
}

func (req *OffLedgerRequest) ID() wasmtypes.ScRequestID {
	// req id is hash of req bytes with output index zero
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	h.Write(req.Bytes())
	h.Write([]byte{0, 0})
	hash := h.Sum(nil)
	return wasmtypes.RequestIDFromBytes(hash)
}
