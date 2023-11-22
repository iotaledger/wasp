package iscclient

import (
	"math"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type OffLedgerRequest struct {
	data   isc.UnsignedOffLedgerRequest
	signed isc.OffLedgerRequest
}

func NewOffLedgerRequest(
	chainID wasmtypes.ScChainID,
	hContract, hFunction wasmtypes.ScHname,
	args []byte,
	allowance *wasmlib.ScAssets,
	nonce uint64,
) (*OffLedgerRequest, error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}

	req := isc.NewOffLedgerRequest(iscChainID, iscContract, iscFunction, params, nonce, math.MaxUint64)
	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	return &OffLedgerRequest{
		data: req,
	}, nil
}

func (req *OffLedgerRequest) Bytes() []byte {
	return req.signed.Bytes()
}

func (req *OffLedgerRequest) Sign(keyPair *Keypair) {
	privKey, err := cryptolib.PrivateKeyFromBytes(keyPair.GetPrivateKey())
	if err != nil {
		panic(err)
	}
	req.signed = req.data.Sign(cryptolib.KeyPairFromPrivateKey(privKey))
}

func (req *OffLedgerRequest) ID() wasmtypes.ScRequestID {
	return cvt.ScRequestID(req.signed.ID())
}
