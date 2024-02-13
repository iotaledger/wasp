package wallets

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp-wallet-sdk/types"
)

func SDKED25519SignatureToIOTAGo(responseSignature *types.Ed25519Signature) (iotago.Signature, error) {
	signatureBytes, err := iotago.DecodeHex(responseSignature.Signature)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := iotago.DecodeHex(responseSignature.PublicKey)
	if err != nil {
		return nil, err
	}

	signature := iotago.Ed25519Signature{}
	copy(signature.Signature[:], signatureBytes)
	copy(signature.PublicKey[:], publicKeyBytes)

	return &signature, nil
}
